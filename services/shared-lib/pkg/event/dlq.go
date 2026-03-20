// services/shared-lib/pkg/event/dlq.go
//
// Dead-Letter Queue (DLQ) Consumer Wrapper
// ─────────────────────────────────────────────────────────────────────────────
//
// Problem this solves (R2):
//   The base Consumer marks offsets even when handler returns an error (at-least-once).
//   For transient errors (network hiccup) that's correct — retry will succeed.
//   For permanent failures (schema mismatch, corrupt message — "poison pills") this
//   causes an infinite loop: the error is logged, the message is re-consumed, error
//   logged again, forever. The consumer is stuck and all messages behind the poison
//   pill are blocked (per-partition).
//
// Solution:
//   DLQConsumer wraps handler execution with a configurable retry budget.
//   After MaxAttempts failures, the raw message bytes are published to a
//   dead-letter topic (`<original-topic>.dlq`) for human inspection,
//   and the offset is committed so the consumer makes forward progress.
//
// Usage:
//
//	dlqProducer := event.NewDLQProducer(brokers)
//	dlqConsumer := event.NewDLQConsumer(brokers, "my-group", []string{"checkout.completed"},
//	    handler, dlqProducer, event.DLQOptions{MaxAttempts: 3, BackoffBase: 500*time.Millisecond})
//	dlqConsumer.StartAsync(ctx)
package event

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/IBM/sarama"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// ── Prometheus metrics ────────────────────────────────────────────────────────

var (
	dlqMessagesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "consumer_dlq_messages_total",
		Help: "Total messages sent to dead-letter topics",
	}, []string{"original_topic", "consumer_group", "reason"})

	consumerHandlerRetries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "consumer_handler_retries_total",
		Help: "Total handler retries before DLQ or success",
	}, []string{"topic", "consumer_group"})
)

// DLQOptions configures retry behaviour for the DLQConsumer.
type DLQOptions struct {
	// MaxAttempts is the total number of times to try a message before DLQ.
	// Default: 3.
	MaxAttempts int

	// BackoffBase is the initial retry wait time. Doubles on each attempt (exponential).
	// Default: 500ms.
	BackoffBase time.Duration
}

func (o *DLQOptions) defaults() {
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = 3
	}
	if o.BackoffBase <= 0 {
		o.BackoffBase = 500 * time.Millisecond
	}
}

// DLQProducer produces failed messages to dead-letter topics.
type DLQProducer struct {
	producer sarama.SyncProducer
}

// NewDLQProducer creates a simple sync producer for writing to DLQ topics.
func NewDLQProducer(brokers []string) (*DLQProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 3

	p, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("dlq producer: %w", err)
	}
	return &DLQProducer{producer: p}, nil
}

// Send publishes a raw message to `<topic>.dlq` with the original key and
// a header recording the reason for DLQ routing.
func (d *DLQProducer) Send(topic string, key string, value []byte, reason string) error {
	dlqTopic := topic + ".dlq"
	msg := &sarama.ProducerMessage{
		Topic: dlqTopic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
		Headers: []sarama.RecordHeader{
			{Key: []byte("x-dlq-reason"), Value: []byte(reason)},
			{Key: []byte("x-dlq-original-topic"), Value: []byte(topic)},
			{Key: []byte("x-dlq-timestamp"), Value: []byte(time.Now().UTC().Format(time.RFC3339))},
		},
	}
	_, _, err := d.producer.SendMessage(msg)
	return err
}

func (d *DLQProducer) Close() error { return d.producer.Close() }

// ── DLQConsumer ───────────────────────────────────────────────────────────────

// DLQConsumer wraps Consumer with retry logic and dead-letter routing.
type DLQConsumer struct {
	inner   *Consumer
	dlq     *DLQProducer
	opts    DLQOptions
	topics  []string
	groupID string
}

// NewDLQConsumer creates a consumer that retries failed messages up to
// opts.MaxAttempts times before routing them to `<topic>.dlq`.
func NewDLQConsumer(
	brokers []string,
	groupID string,
	topics []string,
	handler Handler,
	dlqProducer *DLQProducer,
	opts DLQOptions,
) *DLQConsumer {
	opts.defaults()
	dlqHandler := makeDLQHandler(handler, dlqProducer, groupID, topics, opts)
	return &DLQConsumer{
		inner:   NewConsumer(brokers, groupID, topics, dlqHandler),
		dlq:     dlqProducer,
		opts:    opts,
		topics:  topics,
		groupID: groupID,
	}
}

// StartAsync starts the DLQ-aware consumer in the background.
func (d *DLQConsumer) StartAsync(ctx context.Context) {
	d.inner.StartAsync(ctx)
}

// IsReady delegates to the inner consumer's readiness state.
func (d *DLQConsumer) IsReady() bool { return d.inner.IsReady() }

// makeDLQHandler wraps the user-provided handler with retry + DLQ routing.
func makeDLQHandler(
	handler Handler,
	dlqProducer *DLQProducer,
	groupID string,
	topics []string,
	opts DLQOptions,
) Handler {
	primaryTopic := ""
	if len(topics) > 0 {
		primaryTopic = topics[0]
	}

	return func(ctx context.Context, key string, data []byte) error {
		var lastErr error

		for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
			err := handler(ctx, key, data)
			if err == nil {
				return nil // success — offset will be committed normally
			}

			lastErr = err
			consumerHandlerRetries.WithLabelValues(primaryTopic, groupID).Inc()

			if attempt < opts.MaxAttempts {
				// Exponential backoff: 500ms, 1s, 2s, ...
				wait := time.Duration(float64(opts.BackoffBase) * math.Pow(2, float64(attempt-1)))
				if wait > 10*time.Second {
					wait = 10 * time.Second
				}
				logger.Log.Warn("Handler failed, retrying",
					zap.Int("attempt", attempt),
					zap.Int("max_attempts", opts.MaxAttempts),
					zap.Duration("backoff", wait),
					zap.String("key", key),
					zap.Error(err),
				)
				time.Sleep(wait)
			}
		}

		// Exhausted retries — route to DLQ and commit offset (forward progress).
		reason := fmt.Sprintf("exhausted %d retries: %v", opts.MaxAttempts, lastErr)
		logger.Log.Error("Routing message to DLQ",
			zap.String("topic", primaryTopic+".dlq"),
			zap.String("key", key),
			zap.String("reason", reason),
		)

		if dlqProducer != nil {
			if dlqErr := dlqProducer.Send(primaryTopic, key, data, reason); dlqErr != nil {
				logger.Log.Error("Failed to send to DLQ — message will be lost",
					zap.Error(dlqErr),
					zap.String("key", key),
				)
				dlqMessagesTotal.WithLabelValues(primaryTopic, groupID, "dlq_send_failed").Inc()
			} else {
				dlqMessagesTotal.WithLabelValues(primaryTopic, groupID, "routed").Inc()
			}
		}

		// Return nil so the consumer marks the offset and makes forward progress.
		// The message is preserved in the DLQ topic for replay/investigation.
		return nil
	}
}
