package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/influencer/dto"
	"github.com/redis/go-redis/v9"
)

const llmCallsKeyPrefix = "llm_calls:"

func userLLMCallsKey(userID uuid.UUID) string {
	return fmt.Sprintf("%s%s", llmCallsKeyPrefix, userID.String())
}

type LLMMetricsStore struct {
	client *redis.Client
}

func NewLLMMetricsStore(client *redis.Client) *LLMMetricsStore {
	return &LLMMetricsStore{client: client}
}

func (s *LLMMetricsStore) Record(ctx context.Context, userID uuid.UUID, metric *dto.LLMCallMetric) error {
	if s.client == nil {
		return fmt.Errorf("redis unavailable")
	}
	if metric.Timestamp == 0 {
		metric.Timestamp = time.Now().UnixMilli()
	}
	metric.ID = strconv.FormatInt(time.Now().UnixNano(), 10)

	payload, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("marshal metric: %w", err)
	}

	return s.client.ZAdd(ctx, userLLMCallsKey(userID), redis.Z{
		Score:  float64(metric.Timestamp),
		Member: string(payload),
	}).Err()
}

func (s *LLMMetricsStore) GetStats(ctx context.Context, userID uuid.UUID) (*dto.MonitoringStatsResponse, error) {
	if s.client == nil {
		return &dto.MonitoringStatsResponse{RecentCalls: []dto.LLMCallMetric{}}, nil
	}

	key := userLLMCallsKey(userID)
	pipe := s.client.Pipeline()
	cardCmd := pipe.ZCard(ctx, key)
	recentCmd := pipe.ZRevRange(ctx, key, 0, 19)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("pipeline exec: %w", err)
	}

	total, err := cardCmd.Result()
	if err != nil {
		return nil, fmt.Errorf("zcard: %w", err)
	}

	stats := &dto.MonitoringStatsResponse{TotalCalls: total}
	recentMembers, err := recentCmd.Result()
	if err != nil {
		return nil, fmt.Errorf("zrevrange: %w", err)
	}
	if len(recentMembers) == 0 {
		stats.RecentCalls = []dto.LLMCallMetric{}
		return stats, nil
	}

	stats.RecentCalls = make([]dto.LLMCallMetric, 0, len(recentMembers))
	var totalLatency int64
	var errorCount int64

	for _, member := range recentMembers {
		var m dto.LLMCallMetric
		if err := json.Unmarshal([]byte(member), &m); err != nil {
			continue
		}
		stats.RecentCalls = append(stats.RecentCalls, m)
		totalLatency += m.LatencyMs
		if m.Status == "error" {
			errorCount++
		}
	}

	n := int64(len(stats.RecentCalls))
	if n > 0 {
		stats.AvgLatencyMs = float64(totalLatency) / float64(n)
		stats.ErrorRate = (float64(errorCount) / float64(n)) * 100
	}
	return stats, nil
}
