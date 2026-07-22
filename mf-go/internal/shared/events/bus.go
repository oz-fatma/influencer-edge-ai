package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event is a marker interface for domain events.
type Event interface{}

// Handler is a function that handles an event.
type Handler func(ctx context.Context, event Event) error

// Envelope wraps an event with metadata for transport (Kafka, etc).
type Envelope struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	Source         string          `json:"source"`
	OrganizationID string          `json:"organization_id,omitempty"`
	AppID          string          `json:"app_id,omitempty"`
	Timestamp      time.Time       `json:"timestamp"`
	Data           json.RawMessage `json:"data"`
}

// NewEnvelope creates an envelope wrapping a domain event.
func NewEnvelope(eventType, source string, data interface{}) (*Envelope, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &Envelope{
		ID:        uuid.New().String(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now().UTC(),
		Data:      payload,
	}, nil
}

// EventBus is the interface both in-process and Kafka buses implement.
// Use this interface in use cases and domain services so the transport
// can be swapped without changing business logic.
type EventBus interface {
	// Publish sends an event asynchronously. Fire-and-forget.
	Publish(ctx context.Context, topic string, event Event) error
	// Subscribe registers a handler for a topic. Called at startup.
	Subscribe(topic string, handler Handler)
	// Close shuts down the bus gracefully.
	Close() error
}
