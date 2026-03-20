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

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/order-service/internal/db"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/apierror"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/event"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/middleware"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/tracing"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	logger.Init("order-service")
	logger.Log.Info("Starting order-service...")

	tp, err := tracing.InitTracer("order-service")
	if err != nil {
		logger.Log.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	// ── PostgreSQL (R1: replaces in-memory map) ────────────────────────────
	// POSTGRES_DSN injected via ExternalSecret → api-gateway-secrets.POSTGRES_DSN
	// or order-service-specific order-service-secrets.POSTGRES_DSN.
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		logger.Log.Fatal("POSTGRES_DSN env var is required")
	}

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer startupCancel()

	repo, err := db.NewRepository(startupCtx, dsn, logger.Log)
	if err != nil {
		logger.Log.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer repo.Close()

	// Run idempotent schema migration (safe on every deploy)
	if err := repo.Migrate(startupCtx); err != nil {
		logger.Log.Fatal("Schema migration failed", zap.Error(err))
	}

	// ── Kafka producer ─────────────────────────────────────────────────────
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

	// ── HTTP server ─────────────────────────────────────────────────────────
	rateLimiter := middleware.NewRateLimiter(100, 20)
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) //nolint:errcheck
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Readiness: DB must be reachable
		pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := repo.Ping(pingCtx); err != nil {
			http.Error(w, `{"status":"db_unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"}) //nolint:errcheck
	})

	// POST /v1/orders
	mux.HandleFunc("POST /v1/orders", func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", requestID)

		var order db.Order
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			apierror.BadRequest(w, "Request body must be valid JSON.", requestID)
			return
		}

		// Validation
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
				apierror.ValidationError(w, fmt.Sprintf("Item %d missing 'product_id'", i), requestID)
				return
			}
			if item.Quantity <= 0 {
				apierror.ValidationError(w, fmt.Sprintf("Item %d must have quantity > 0", i), requestID)
				return
			}
			if item.Price < 0 {
				apierror.ValidationError(w, fmt.Sprintf("Item %d must have price >= 0", i), requestID)
				return
			}
		}

		// Populate system fields
		order.ID = uuid.New().String()
		order.Status = "PENDING"
		order.IdempotencyKey = r.Header.Get("X-Idempotency-Key")

		// ── Persist to PostgreSQL (ACID, idempotent) ────────────────────────
		saved, err := repo.Create(r.Context(), order)
		if err != nil {
			logger.Log.Error("Failed to persist order", zap.Error(err), zap.String("order_id", order.ID))
			apierror.InternalServerError(w, requestID)
			return
		}

		// ── Emit typed Kafka event (async, non-blocking) ────────────────────
		if producer != nil {
			go func(o db.Order) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				evt := event.CheckoutCompleted{
					Metadata: event.Metadata{
						EventID:     uuid.New().String(),
						EventType:   event.TypeCheckoutCompleted,
						ServiceName: "order-service",
						UserID:      o.UserID,
						Timestamp:   time.Now().UTC(),
						Version:     "1",
					},
					OrderID:     o.ID,
					TotalAmount: o.Total,
					Currency:    o.Currency,
				}
				if err := producer.Emit(ctx, event.TypeCheckoutCompleted, o.ID, evt); err != nil {
					logger.Log.Error("Failed to emit checkout event",
						zap.Error(err), zap.String("order_id", o.ID))
				}
			}(saved)
		}

		w.Header().Set("Content-Type", "application/json")
		statusCode := http.StatusCreated
		// If idempotency key matched an existing order, return 200 not 201
		if saved.ID != order.ID {
			statusCode = http.StatusOK
			w.Header().Set("X-Idempotent-Replayed", "true")
		}
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(saved) //nolint:errcheck
	})

	// GET /v1/orders/{id}
	mux.HandleFunc("GET /v1/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		id := r.PathValue("id")

		order, err := repo.GetByID(r.Context(), id)
		if err != nil {
			if err == db.ErrNotFound {
				apierror.NotFound(w, "order", id, requestID)
				return
			}
			apierror.InternalServerError(w, requestID)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order) //nolint:errcheck
	})

	// GET /v1/orders?user_id=xxx
	mux.HandleFunc("GET /v1/orders", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			apierror.ValidationError(w, "Query param 'user_id' is required.", r.Header.Get("X-Request-ID"))
			return
		}
		orders, err := repo.ListByUser(r.Context(), userID)
		if err != nil {
			apierror.InternalServerError(w, r.Header.Get("X-Request-ID"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders) //nolint:errcheck
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           rateLimiter.RateLimit(mux),
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
