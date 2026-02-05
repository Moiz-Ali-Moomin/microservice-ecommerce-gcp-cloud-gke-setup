package main

import (
	"log"
	"net/http"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/tracing"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/httpserver"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/offer-service/internal/handler"
	"go.uber.org/zap"
)

func main() {
	logger.Init("offer-service")
	
	tp, err := tracing.InitTracer("offer-service")
	if err != nil {
		logger.Log.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(nil) }()

	h := handler.New()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/offers", h.GetOffers)
	mux.HandleFunc("/offers/", h.GetOfferByID)

	server := httpserver.New(mux, "offer-service", "8080")
	if err := server.Run(); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}

