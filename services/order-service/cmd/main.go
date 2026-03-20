package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/apierror"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/event"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/middleware"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/tracing"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Order struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Items          []Item    `json:"items"`
	Total          float64   `json:"total"`
	Status         string    `json:"status"`
	IdempotencyKey string    `json:"-"` // stored internally, not exposed in response
	CreatedAt      time.Time `json:"created_at"`
}

type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// In production, replace with PostgreSQL/Redis persistence.
var (
	orders          = make(map[string]Order)
	idempotencyKeys = make(map[string]string) // idempotency-key → order ID
)

func main() {
	logger.Init("order-service")
	logger.Log.Info("Starting order-service...")

	tp, err := tracing.InitTracer("order-service")
	if err != nil {
		logger.Log.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"kafka-cluster-kafka-bootstrap.kafka:9092"}
	}

	producer, err := event.NewProducer(brokers, "order-service")
	if err != nil {
		logger.Log.Error("Failed to create kafka producer", zap.Error(err))
	} else {
		defer producer.Close()
	}

	// Rate limiter: 100 req/s per IP, burst of 20
	rateLimiter := middleware.NewRateLimiter(100, 20)

	mux := http.NewServeMux()

	// ── Observability endpoints ──────────────────────────────────────────────
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) //nolint:errcheck
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"}) //nolint:errcheck
	})

	// ── API v1 Routes ────────────────────────────────────────────────────────

	// POST /v1/orders — Create Order (with idempotency key)
	mux.HandleFunc("POST /v1/orders", func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", requestID)

		// ── Idempotency check ────────────────────────────────────────────────
		idempotencyKey := r.Header.Get("X-Idempotency-Key")
		if idempotencyKey != "" {
			if existingOrderID, ok := idempotencyKeys[idempotencyKey]; ok {
				// Return the existing order — same request, idempotent
				if order, exists := orders[existingOrderID]; exists {
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("X-Idempotent-Replayed", "true")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(order) //nolint:errcheck
					return
				}
			}
		}

		// ── Input parsing ────────────────────────────────────────────────────
		var order Order
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			apierror.BadRequest(w, "Request body must be valid JSON.", requestID)
			return
		}

		// ── Input validation ─────────────────────────────────────────────────
		if order.UserID == "" {
			apierror.ValidationError(w, "Field 'user_id' is required.", requestID)
			return
		}
		if len(order.Items) == 0 {
			apierror.ValidationError(w, "Field 'items' must contain at least one item.", requestID)
			return
		}
		for i, item := range order.Items {
			if item.ProductID == "" {
				apierror.ValidationError(w, fmt.Sprintf("Item at index %d is missing 'product_id'.", i), requestID)
				return
			}
			if item.Quantity <= 0 {
				apierror.ValidationError(w, fmt.Sprintf("Item at index %d must have 'quantity' > 0.", i), requestID)
				return
			}
			if item.Price < 0 {
				apierror.ValidationError(w, fmt.Sprintf("Item at index %d must have 'price' >= 0.", i), requestID)
				return
			}
		}

		// ── Create the order ─────────────────────────────────────────────────
		order.ID = uuid.New().String()
		order.Status = "PENDING"
		order.CreatedAt = time.Now().UTC()
		order.IdempotencyKey = idempotencyKey
		orders[order.ID] = order

		if idempotencyKey != "" {
			idempotencyKeys[idempotencyKey] = order.ID
		}

		// ── Emit Kafka event (async, non-blocking) ──────────────────────────
		if producer != nil {
			go func(o Order) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				evt := event.Event{
					EventID:   uuid.New().String(),
					Timestamp: time.Now(),
					Service:   "order-service",
					Metadata: map[string]interface{}{
						"order_id": o.ID,
						"user_id":  o.UserID,
						"total":    o.Total,
					},
				}
				if err := producer.Emit(ctx, "checkout.completed", o.ID, evt); err != nil {
					logger.Log.Error("Failed to emit order event",
						zap.Error(err),
						zap.String("order_id", o.ID),
					)
				}
			}(order)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(order) //nolint:errcheck
	})

	// GET /v1/orders — List Orders
	mux.HandleFunc("GET /v1/orders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		orderList := make([]Order, 0, len(orders))
		for _, o := range orders {
			orderList = append(orderList, o)
		}
		json.NewEncoder(w).Encode(orderList) //nolint:errcheck
	})

	// GET /v1/orders/{id} — Get Order by ID
	mux.HandleFunc("GET /v1/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		id := r.PathValue("id")

		order, ok := orders[id]
		if !ok {
			apierror.NotFound(w, "order", id, requestID)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order) //nolint:errcheck
	})

	// Apply rate limiting to all API routes
	rateLimitedMux := rateLimiter.RateLimit(mux)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           rateLimitedMux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	go func() {
		logger.Log.Info("order-service listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Listen failure", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down order-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown", zap.Error(err))
	}
	logger.Log.Info("order-service exited cleanly")
}
