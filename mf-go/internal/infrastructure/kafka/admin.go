package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/masterfabric-go/masterfabric/internal/shared/events"
)

// DefaultTopics returns the list of topics this application requires.
func DefaultTopics() []string {
	return []string{
		events.TopicIAM,
		events.TopicTenant,
		events.TopicAPIManagement,
		events.TopicAudit,
	}
}

// EnsureTopics creates Kafka topics if they don't already exist.
// This is meant to be called at application startup.
func EnsureTopics(ctx context.Context, brokerAddr string, topics []string, numPartitions int, replicationFactor int, logger *slog.Logger) error {
	conn, err := kafkago.Dial("tcp", brokerAddr)
	if err != nil {
		return fmt.Errorf("kafka dial: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("kafka controller: %w", err)
	}

	controllerConn, err := kafkago.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("kafka controller dial: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := make([]kafkago.TopicConfig, 0, len(topics))
	for _, t := range topics {
		topicConfigs = append(topicConfigs, kafkago.TopicConfig{
			Topic:             t,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		})
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		// kafka-go returns an error even if topic already exists; log and continue
		logger.Warn("kafka create topics (may already exist)", "error", err)
	}

	for _, t := range topics {
		logger.Info("kafka topic ensured", "topic", t)
	}
	return nil
}
