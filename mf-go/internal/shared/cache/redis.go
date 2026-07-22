package cache

import (
	"context"
	"fmt"
	"strings"

	"github.com/masterfabric-go/masterfabric/internal/shared/config"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new Redis client and verifies the connection.
func NewRedisClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	var client *redis.Client

	if cfg.URL != "" {
		opts, err := redis.ParseURL(normalizeRedisURL(cfg.URL))
		if err != nil {
			return nil, fmt.Errorf("parse redis url: %w", err)
		}
		client = redis.NewClient(opts)
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr(),
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	}

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

func normalizeRedisURL(raw string) string {
	if raw == "" {
		return "redis://localhost:6379"
	}
	if strings.HasPrefix(raw, "redis://") || strings.HasPrefix(raw, "rediss://") {
		return raw
	}
	return "redis://" + raw
}
