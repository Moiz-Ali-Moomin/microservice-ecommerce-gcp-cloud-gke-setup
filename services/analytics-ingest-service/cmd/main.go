package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/event"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/tracing"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type IngestRequest struct {
	EventName string                 `json:"event_name"`
	Payload   map[string]interface{} `json:"payload"`
}

func main() {
	logger.Init("analytics-ingest-service")

	tp, err := tracing.InitTracer("analytics-ingest-service")
	if err != nil {
		logger.Log.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(nil) }()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"kafka-headless.kafka.svc:9092"}
	}

	// We create producers for allowed topics on demand or pre-init map
	// For simplicity, let's assume one producer for 'raw.analytics' or dynamic
	// Master Prompt says "page.viewed", etc. We'll map dynamic here.

	// Pre-create producers
	producers := make(map[string]*event.Producer)
	topics := []string{"page.viewed", "cta.clicked", "user.signup", "checkout.completed"}

	for _, t := range topics {
		p, err := event.NewProducer(brokers, t)
		if err != nil {
			logger.Log.Error("Failed to create producer", zap.String("topic", t), zap.Error(err))
		} else {
			producers[t] = p
			defer p.Close()
		}
	}

	http.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		var req IngestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad JSON", http.StatusBadRequest)
			return
		}

		p, ok := producers[req.EventName]
		if !ok {
			http.Error(w, "Unknown Event Type", http.StatusBadRequest)
			return
		}

		// Wrap in standard event
		evt := event.Event{
			EventID:   uuid.New().String(),
			Timestamp: time.Now(),
			Service:   "analytics-ingest",
			Metadata:  req.Payload,
		}

		go func() {
			if err := p.Emit(context.Background(), req.EventName, uuid.New().String(), evt); err != nil {
				logger.Log.Error("Ingest failed", zap.Error(err))
			}
		}()

		w.WriteHeader(http.StatusAccepted)
	})

	logger.Log.Info("Starting analytics-ingest-service on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
