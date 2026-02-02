package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"go.uber.org/zap"
)

type Event struct {
	EventID    string                 `json:"event_id"`
	Timestamp  time.Time              `json:"timestamp"`
	Service    string                 `json:"service_name"`
	UserID     string                 `json:"user_id,omitempty"`
	SessionID  string                 `json:"session_id,omitempty"`
	CampaignID string                 `json:"campaign_id,omitempty"`
	OfferID    string                 `json:"offer_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	p, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{producer: p, topic: topic}, nil
}

func (p *Producer) Emit(ctx context.Context, key string, event interface{}) error {
	val, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(val),
		Timestamp: time.Now(),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		logger.Log.Error("Failed to send message", zap.Error(err), zap.String("topic", p.topic))
		return err
	}

	logger.Log.Debug("Message stored", zap.String("topic", p.topic), zap.Int32("partition", partition), zap.Int64("offset", offset))
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

