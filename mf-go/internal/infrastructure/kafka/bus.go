package kafka

import (
	"context"
	"log/slog"
	"reflect"
	"strings"
	"unicode"

	"github.com/masterfabric-go/masterfabric/internal/shared/events"
)

// Bus implements events.EventBus using Kafka as the transport.
// It uses a Producer for publishing and a Consumer for subscribing.
type Bus struct {
	producer *Producer
	consumer *Consumer
	brokers  []string
	groupID  string
	logger   *slog.Logger
}

// Verify interface compliance at compile time.
var _ events.EventBus = (*Bus)(nil)

// NewBus creates a new Kafka-backed event bus.
func NewBus(brokers []string, groupID string, logger *slog.Logger) *Bus {
	return &Bus{
		producer: NewProducer(brokers, logger),
		consumer: NewConsumer(brokers, groupID, logger),
		brokers:  brokers,
		groupID:  groupID,
		logger:   logger,
	}
}

// Publish sends an event to a Kafka topic.
// The event is serialized into an Envelope with the topic used as a routing key.
func (b *Bus) Publish(ctx context.Context, topic string, event events.Event) error {
	// Derive event type from the event struct for the envelope
	eventType := deriveEventType(event)
	return b.producer.Publish(ctx, topic, eventType, event)
}

// Subscribe registers a handler for a topic.
// The handler receives the raw Event. Under the hood we wrap it to unmarshal
// from the Kafka Envelope.
func (b *Bus) Subscribe(topic string, handler events.Handler) {
	b.consumer.Subscribe(topic, func(ctx context.Context, envelope *events.Envelope) error {
		// Re-wrap the raw JSON data as an event. Consumers that need the typed
		// event should type-switch on Envelope.Type and unmarshal Data.
		return handler(ctx, envelope)
	})
}

// Start begins consuming from all subscribed topics. Call after all Subscribe calls.
func (b *Bus) Start(ctx context.Context) {
	b.consumer.Start(ctx, b.brokers, b.groupID)
}

// Close shuts down both producer and consumer.
func (b *Bus) Close() error {
	var firstErr error
	if err := b.consumer.Close(); err != nil {
		firstErr = err
	}
	if err := b.producer.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

// deriveEventType returns a dotted event type string from the Go struct name.
// Examples: UserRegistered -> "user.registered", EndpointCreated -> "endpoint.created"
func deriveEventType(event events.Event) string {
	// If it's an envelope already, use its Type
	if env, ok := event.(*events.Envelope); ok {
		return env.Type
	}

	// Use reflection to get the struct name and convert to dot-notation.
	// e.g. "UserRegistered" -> "user.registered"
	t := reflect.TypeOf(event)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := t.Name()
	if name == "" {
		return "unknown"
	}
	return camelToDotCase(name)
}

// camelToDotCase converts "UserRegistered" to "user.registered".
func camelToDotCase(s string) string {
	var parts []string
	start := 0
	for i := 1; i < len(s); i++ {
		if unicode.IsUpper(rune(s[i])) {
			parts = append(parts, strings.ToLower(s[start:i]))
			start = i
		}
	}
	parts = append(parts, strings.ToLower(s[start:]))
	return strings.Join(parts, ".")
}
