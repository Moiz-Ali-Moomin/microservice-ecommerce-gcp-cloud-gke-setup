package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/event"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/httpserver"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/tracing"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/redirect-service/internal/handler"
	"go.uber.org/zap"
)

func main() {
	logger.Init("redirect-service")
	
	tp, err := tracing.InitTracer("redirect-service")
	if err != nil {
		logger.Log.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(nil) }()

	// Kafka Producer Setup
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		// FIX: Fail fast if env var missing in production
		if os.Getenv("ENV") == "production" {
			logger.Log.Fatal("KAFKA_BROKERS env var required")
		}
		brokers = []string{"kafka-headless.kafka.svc:9092"}
	}

	// Producer for "cta.clicked"
	clickProducer, err := event.NewProducer(brokers, "cta.clicked")
	if err != nil {
		logger.Log.Error("Failed to create click producer", zap.Error(err))
	}
	defer clickProducer.Close()

	// Producer for "checkout.redirected"
	redirectProducer, err := event.NewProducer(brokers, "checkout.redirected")
	if err != nil {
		logger.Log.Error("Failed to create redirect producer", zap.Error(err))
	}
	defer redirectProducer.Close()

	h := handler.New(clickProducer, redirectProducer)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/click/", h.HandleClick)

	// FIX: Use shared httpserver for graceful shutdown & metrics
	server := httpserver.New(mux, "redirect-service", "8080")
	if err := server.Run(); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}

