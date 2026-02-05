package service

import (
	"context"
	"time"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/event"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"go.uber.org/zap"
)

type AnalyticsService struct {
	producer *event.Producer
}

func NewAnalyticsService(producer *event.Producer) *AnalyticsService {
	return &AnalyticsService{producer: producer}
}

func (s *AnalyticsService) TrackPageView(path, userAgent string) {
	payload := map[string]interface{}{
		"event":      "page.viewed",
		"path":       path,
		"user_agent": userAgent,
		"source":     "storefront",
	}

	go func() {
		if s.producer == nil {
			// If producer failed to init, just return
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.producer.Emit(ctx, path, payload); err != nil {
			logger.Log.Error("Failed to publish page.viewed event",
				zap.Error(err),
				zap.String("path", path),
			)
		}
	}()
}
