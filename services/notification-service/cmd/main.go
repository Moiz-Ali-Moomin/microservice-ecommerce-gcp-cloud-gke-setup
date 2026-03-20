package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/event"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/tracing"
	"go.uber.org/zap"
)

// Use the DLQConsumer type directly — it exposes IsReady() and StartAsync()
var consumer *event.DLQConsumer

func main() {
	logger.Init("notification-service")

	tp, err := tracing.InitTracer("notification-service")
	if err != nil {
		logger.Log.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"kafka-cluster-kafka-bootstrap.kafka:9092"}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── DLQ producer (R2 fix) ─────────────────────────────────────────────
	// Routes poison-pill messages to checkout.completed.dlq after 3 retries,
	// then commits offset so healthy messages behind the poison pill are unblocked.
	dlqProducer, err := event.NewDLQProducer(brokers)
	if err != nil {
		logger.Log.Warn("DLQ producer unavailable, falling back to plain consumer", zap.Error(err))
	}

	// ── DLQConsumer: checkout.completed (was: conversion.completed — WRONG) ─
	consumer = event.NewDLQConsumer(
		brokers,
		"notification-group",
		[]string{event.TypeCheckoutCompleted},
		HandleNotification,
		dlqProducer,
		event.DLQOptions{MaxAttempts: 3, BackoffBase: 500 * time.Millisecond},
	)

	logger.Log.Info("Starting notification-service Kafka consumer (DLQ-aware)...")
	consumer.StartAsync(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", readyHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		<-sigterm
		logger.Log.Info("Shutting down notification-service...")
		cancel()
		_ = server.Close()
	}()

	logger.Log.Info("notification-service HTTP server listening on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal("HTTP server failed", zap.Error(err))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	if consumer != nil && consumer.IsReady() {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("kafka not ready"))
	}
}

// HandleNotification processes checkout.completed events and dispatches notifications.
// This is the BUSINESS LOGIC handler — DLQConsumer wraps it with retries.
func HandleNotification(ctx context.Context, key string, data []byte) error {
	var evt event.CheckoutCompleted
	if err := json.Unmarshal(data, &evt); err != nil {
		// Return error — DLQConsumer will retry up to MaxAttempts, then DLQ the message.
		logger.Log.Error("Failed to deserialize checkout event",
			zap.Error(err),
			zap.String("raw_length", string(rune(len(data)))),
		)
		return err
	}

	logger.Log.Info("Dispatching order confirmation notification",
		zap.String("order_id", evt.OrderID),
		zap.String("user_id", evt.UserID),
		zap.Float64("total_amount", evt.TotalAmount),
		zap.String("currency", evt.Currency),
		zap.String("event_id", evt.EventID),
	)

	// TODO: dispatch via notification provider (SendGrid, SNS, FCM)
	// notification.SendOrderConfirmation(ctx, evt.UserID, evt.OrderID, evt.TotalAmount)

	return nil
}
