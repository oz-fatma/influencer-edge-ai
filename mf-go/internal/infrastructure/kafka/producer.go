package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"

	"github.com/masterfabric-go/masterfabric/internal/shared/events"
)

// Producer wraps a kafka-go Writer to publish domain events.
type Producer struct {
	writer *kafkago.Writer
	logger *slog.Logger
}

// NewProducer creates a new Kafka producer.
func NewProducer(brokers []string, logger *slog.Logger) *Producer {
	w := &kafkago.Writer{
		Addr:         kafkago.TCP(brokers...),
		Balancer:     &kafkago.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond, // Low latency for event publishing
		RequiredAcks: kafkago.RequireOne,
		Async:        false, // Synchronous writes for reliability
	}

	return &Producer{
		writer: w,
		logger: logger,
	}
}

// Publish serializes the event into an Envelope and writes it to the given Kafka topic.
func (p *Producer) Publish(ctx context.Context, topic string, eventType string, event interface{}) error {
	envelope, err := events.NewEnvelope(eventType, "masterfabric-go", event)
	if err != nil {
		return fmt.Errorf("create envelope: %w", err)
	}

	value, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	msg := kafkago.Message{
		Topic: topic,
		Key:   []byte(uuid.New().String()),
		Value: value,
		Headers: []kafkago.Header{
			{Key: "event-type", Value: []byte(eventType)},
			{Key: "event-id", Value: []byte(envelope.ID)},
		},
		Time: time.Now().UTC(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("kafka publish failed", "topic", topic, "event_type", eventType, "error", err)
		return fmt.Errorf("kafka write: %w", err)
	}

	p.logger.Debug("kafka event published", "topic", topic, "event_type", eventType, "event_id", envelope.ID)
	return nil
}

// Close closes the Kafka writer.
func (p *Producer) Close() error {
	return p.writer.Close()
}
