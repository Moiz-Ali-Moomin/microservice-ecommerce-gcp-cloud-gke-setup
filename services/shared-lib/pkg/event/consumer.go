package event

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"go.uber.org/zap"
)

type Handler func(ctx context.Context, key string, event []byte) error

type Consumer struct {
	brokers []string
	topics  []string
	groupID string
	handler Handler
}

func NewConsumer(brokers []string, groupID string, topics []string, handler Handler) *Consumer {
	return &Consumer{
		brokers: brokers,
		topics:  topics,
		groupID: groupID,
		handler: handler,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(c.brokers, c.groupID, config)
	if err != nil {
		return err
	}

	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle graceful shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigterm
		logger.Log.Info("Terminating consumer via signal")
		cancel()
	}()

	handler := &groupHandler{
		handler: c.handler,
	}

	for {
		if err := consumer.Consume(consumerCtx, c.topics, handler); err != nil {
			logger.Log.Error("Error from consumer", zap.Error(err))
			return err
		}
		if consumerCtx.Err() != nil {
			return nil
		}
	}
}

type groupHandler struct {
	handler Handler
}

func (h *groupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *groupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *groupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logger.Log.Debug("Message claimed", zap.String("topic", msg.Topic), zap.Int64("offset", msg.Offset))
		if err := h.handler(sess.Context(), string(msg.Key), msg.Value); err != nil {
			logger.Log.Error("Handler failed", zap.Error(err))
			// Decide whether to mark offset or not based on delivery guarantees
			// For now, we continue
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
