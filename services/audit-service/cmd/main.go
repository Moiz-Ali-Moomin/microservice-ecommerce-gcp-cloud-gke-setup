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
	logger.Init("audit-service")
	
	tp, err := tracing.InitTracer("audit-service")
	if err != nil {
		logger.Log.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(nil) }()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"kafka-headless.kafka.svc:9092"}
	}

	// Audit all topics for compliance
	topics := []string{"page.viewed", "cta.clicked", "checkout.redirected", "conversion.completed", "campaign.attributed", "user.signup", "checkout.completed"}

	consumer := event.NewConsumer(
		brokers,
		"audit-group",
		topics,
		HandleAudit,
	)

	logger.Log.Info("Starting audit-service consumer...")
	if err := consumer.Start(context.Background()); err != nil {
		logger.Log.Fatal("Consumer failed", zap.Error(err))
	}
}

func HandleAudit(ctx context.Context, key string, data []byte) error {
	var evt event.Event
	if err := json.Unmarshal(data, &evt); err != nil {
		return nil
	}
	// Log to stdout, picked up by Loki
	logger.Log.Info("AUDIT_EVENT", 
		zap.String("event_id", evt.EventID),
		zap.String("type", evt.Service), // Naive mapping
		zap.Any("payload", evt),
	)
	return nil
}

