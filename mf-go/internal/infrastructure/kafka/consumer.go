package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/masterfabric-go/masterfabric/internal/shared/events"
)

// ConsumerHandler is called for each message consumed from Kafka.
// The raw Envelope is passed so handlers can inspect type/metadata and unmarshal Data.
type ConsumerHandler func(ctx context.Context, envelope *events.Envelope) error

// Consumer wraps kafka-go Readers to consume messages from one or more topics.
type Consumer struct {
	readers  []*kafkago.Reader
	logger   *slog.Logger
	handlers map[string][]ConsumerHandler // topic -> handlers
	mu       sync.RWMutex
	wg       sync.WaitGroup
	cancel   context.CancelFunc
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(brokers []string, groupID string, logger *slog.Logger) *Consumer {
	return &Consumer{
		logger:   logger,
		handlers: make(map[string][]ConsumerHandler),
	}
}

// Subscribe registers a handler for a Kafka topic.
// Must be called before Start.
func (c *Consumer) Subscribe(topic string, handler ConsumerHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[topic] = append(c.handlers[topic], handler)
}

// Start begins consuming from all subscribed topics.
// This is a non-blocking call; consumption runs in background goroutines.
func (c *Consumer) Start(ctx context.Context, brokers []string, groupID string) {
	ctx, c.cancel = context.WithCancel(ctx)

	c.mu.RLock()
	defer c.mu.RUnlock()

	for topic, handlers := range c.handlers {
		reader := kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 1,
			MaxBytes: 10e6, // 10 MB
		})
		c.readers = append(c.readers, reader)

		// Copy handlers for the goroutine closure
		topicHandlers := make([]ConsumerHandler, len(handlers))
		copy(topicHandlers, handlers)
		topicName := topic

		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.consumeLoop(ctx, reader, topicName, topicHandlers)
		}()

		c.logger.Info("kafka consumer started", "topic", topicName, "group", groupID)
	}
}

func (c *Consumer) consumeLoop(ctx context.Context, reader *kafkago.Reader, topic string, handlers []ConsumerHandler) {
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // Context cancelled, shutting down
			}
			c.logger.Error("kafka fetch error", "topic", topic, "error", err)
			continue
		}

		var envelope events.Envelope
		if err := json.Unmarshal(msg.Value, &envelope); err != nil {
			c.logger.Error("kafka unmarshal error", "topic", topic, "error", err, "offset", msg.Offset)
			// Commit to avoid reprocessing bad messages
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		// Dispatch to all handlers
		var handlerErr error
		for _, h := range handlers {
			if err := h(ctx, &envelope); err != nil {
				c.logger.Error("kafka handler error",
					"topic", topic,
					"event_type", envelope.Type,
					"event_id", envelope.ID,
					"error", err,
				)
				handlerErr = err
			}
		}

		// Only commit if all handlers succeeded (at-least-once delivery)
		if handlerErr == nil {
			if err := reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("kafka commit error", "topic", topic, "error", err)
			}
		}
	}
}

// Close gracefully shuts down all consumers.
func (c *Consumer) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()

	var errs []error
	for _, r := range c.readers {
		if err := r.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("kafka consumer close errors: %v", errs)
	}
	return nil
}
