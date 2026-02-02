package main

import (
	"context"
	"os"
	"strings"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/attribution-service/internal/handler"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/event"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/tracing"
	"go.uber.org/zap"
)

func main() {
	logger.Init("attribution-service")
	
	tp, err := tracing.InitTracer("attribution-service")
	if err != nil {
		logger.Log.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(nil) }()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"kafka-headless.kafka.svc:9092"}
	}

	h := handler.New()

	// Consumer for "conversion.completed"
	consumer := event.NewConsumer(
		brokers,
		"attribution-group",
		[]string{"conversion.completed"},
		h.HandleConversionEvent,
	)

	logger.Log.Info("Starting attribution-service consumer...")
	if err := consumer.Start(context.Background()); err != nil {
		logger.Log.Fatal("Consumer failed", zap.Error(err))
	}
}

