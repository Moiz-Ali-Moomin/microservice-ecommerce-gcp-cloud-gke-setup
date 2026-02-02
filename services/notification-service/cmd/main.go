package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/event"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/tracing"
	"go.uber.org/zap"
)

func main() {
	logger.Init("notification-service")
	
	tp, err := tracing.InitTracer("notification-service")
	if err != nil {
		logger.Log.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(nil) }()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"kafka-headless.kafka.svc:9092"}
	}

	// Consume 'conversion.completed' to send emails to admin/affiliate
	consumer := event.NewConsumer(
		brokers,
		"notification-group",
		[]string{"conversion.completed"},
		HandleNotification,
	)

	logger.Log.Info("Starting notification-service consumer...")
	if err := consumer.Start(context.Background()); err != nil {
		logger.Log.Fatal("Consumer failed", zap.Error(err))
	}
}

func HandleNotification(ctx context.Context, key string, data []byte) error {
	var evt event.Event
	if err := json.Unmarshal(data, &evt); err != nil {
		return nil
	}
	// Mock Email Sending
	logger.Log.Info("Sending notification email", 
		zap.String("event", "conversion"), 
		zap.String("session_id", evt.SessionID),
	)
	return nil
}

