package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/event"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/tracing"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/landing-service/internal/handler"
	"go.uber.org/zap"
)

func main() {
	logger.Init("landing-service")
	
	tp, err := tracing.InitTracer("landing-service")
	if err != nil {
		logger.Log.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(nil) }()

	// Kafka Producer Setup
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"kafka-headless.kafka.svc:9092"}
	}

	producer, err := event.NewProducer(brokers, "page.viewed")
	if err != nil {
		logger.Log.Error("Failed to create producer", zap.Error(err))
	}
	defer producer.Close()

	h := handler.New(producer)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/", h.ServeLandingPage)

	logger.Log.Info("Starting landing-service on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

