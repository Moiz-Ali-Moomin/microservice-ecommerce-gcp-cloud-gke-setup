package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Server wraps http.Server with graceful shutdown and observability built in.
type Server struct {
	httpServer *http.Server
	ready      bool
}

// New creates a production HTTP server with:
//   - /metrics — Prometheus scrape endpoint (no tracing overhead)
//   - /health  — liveness probe (always 200 if process is running)
//   - /ready   — readiness probe (200 only when server is fully initialised)
//   - All other requests wrapped with OTel tracing + Prometheus metrics
//   - Hardened timeouts to prevent slowloris & resource exhaustion
func New(handler http.Handler, serviceName string, port string) *Server {
	s := &Server{}

	mux := http.NewServeMux()

	// Observability endpoints — bypass application middleware (no trace noise)
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) //nolint:errcheck
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !s.ready {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "not_ready"}) //nolint:errcheck
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"}) //nolint:errcheck
	})

	// Application handler wrapped with OTel tracing + Prometheus metrics
	instrumentedHandler := middleware.Chain(
		handler,
		middleware.Tracing(serviceName),
		middleware.Metrics,
	)
	mux.Handle("/", instrumentedHandler)

	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
		// Timeout hierarchy (prevents slowloris + resource leaks)
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	return s
}

// SetReady marks the server as ready to serve traffic.
// Call this after all initialisation (DB ping, cache warmup) is complete.
func (s *Server) SetReady(ready bool) {
	s.ready = ready
}

// Run starts the HTTP server and blocks until SIGINT/SIGTERM, then
// gracefully shuts down with a 30-second timeout.
func (s *Server) Run() error {
	srvErr := make(chan error, 1)

	go func() {
		logger.Log.Info("Starting HTTP server", zap.String("addr", s.httpServer.Addr))
		s.ready = true
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srvErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-srvErr:
		return err
	case sig := <-quit:
		logger.Log.Info("Received signal, shutting down...", zap.String("signal", sig.String()))
	}

	// Mark not-ready immediately so load balancers drain connections
	s.ready = false

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Log.Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	logger.Log.Info("Server exited cleanly")
	return nil
}
