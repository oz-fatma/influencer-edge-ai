package events

import (
	"context"
	"log/slog"
	"sync"
)

// InProcessBus is an in-memory event bus using Go channels.
// Suitable for Phase 1 / testing / single-instance deployments.
type InProcessBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	logger   *slog.Logger
	ch       chan inProcessEnvelope
	done     chan struct{}
}

type inProcessEnvelope struct {
	ctx   context.Context
	topic string
	event Event
}

// Verify interface compliance at compile time.
var _ EventBus = (*InProcessBus)(nil)

// NewInProcessBus creates a new in-process event bus.
func NewInProcessBus(logger *slog.Logger, bufferSize int) *InProcessBus {
	if bufferSize <= 0 {
		bufferSize = 256
	}
	b := &InProcessBus{
		handlers: make(map[string][]Handler),
		logger:   logger,
		ch:       make(chan inProcessEnvelope, bufferSize),
		done:     make(chan struct{}),
	}
	go b.dispatch()
	return b
}

// Publish sends an event to all registered handlers asynchronously.
func (b *InProcessBus) Publish(ctx context.Context, topic string, event Event) error {
	select {
	case b.ch <- inProcessEnvelope{ctx: ctx, topic: topic, event: event}:
		return nil
	default:
		b.logger.Warn("in-process event bus buffer full, dropping event", "topic", topic)
		return nil
	}
}

// Subscribe registers a handler for a topic.
func (b *InProcessBus) Subscribe(topic string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[topic] = append(b.handlers[topic], handler)
	b.logger.Debug("in-process handler registered", "topic", topic)
}

// Close shuts down the event bus.
func (b *InProcessBus) Close() error {
	close(b.ch)
	<-b.done
	return nil
}

func (b *InProcessBus) dispatch() {
	defer close(b.done)
	for env := range b.ch {
		b.mu.RLock()
		handlers := b.handlers[env.topic]
		b.mu.RUnlock()

		for _, h := range handlers {
			if err := h(env.ctx, env.event); err != nil {
				b.logger.Error("in-process handler error", "topic", env.topic, "error", err)
			}
		}
	}
}
