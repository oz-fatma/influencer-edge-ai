package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const llmCallsKey = "llm_calls"

type LLMCallMetric struct {
	ID             string `json:"id"`
	InfluencerName string `json:"influencer_name"`
	LatencyMs      int64  `json:"latency_ms"`
	Status         string `json:"status"`
	Model          string `json:"model"`
	Timestamp      int64  `json:"timestamp"`
}

type LLMMonitoringStats struct {
	TotalCalls    int64           `json:"total_calls"`
	AvgLatencyMs  float64         `json:"avg_latency_ms"`
	ErrorRate     float64         `json:"error_rate"`
	RecentCalls   []LLMCallMetric `json:"recent_calls"`
}

type LLMMetricsStore struct {
	redis *redis.Client
}

func NewLLMMetricsStore(rdb *redis.Client) *LLMMetricsStore {
	return &LLMMetricsStore{redis: rdb}
}

func (s *LLMMetricsStore) Record(ctx context.Context, metric *LLMCallMetric) error {
	if metric.Timestamp == 0 {
		metric.Timestamp = time.Now().UnixMilli()
	}
	metric.ID = strconv.FormatInt(time.Now().UnixNano(), 10)

	payload, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("marshal metric: %w", err)
	}

	return s.redis.ZAdd(ctx, llmCallsKey, redis.Z{
		Score:  float64(metric.Timestamp),
		Member: string(payload),
	}).Err()
}

func (s *LLMMetricsStore) GetStats(ctx context.Context) (*LLMMonitoringStats, error) {
	pipe := s.redis.Pipeline()
	cardCmd := pipe.ZCard(ctx, llmCallsKey)
	recentCmd := pipe.ZRevRange(ctx, llmCallsKey, 0, 19)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("pipeline exec: %w", err)
	}

	total, err := cardCmd.Result()
	if err != nil {
		return nil, fmt.Errorf("zcard: %w", err)
	}

	stats := &LLMMonitoringStats{TotalCalls: total}

	recentMembers, err := recentCmd.Result()
	if err != nil {
		return nil, fmt.Errorf("zrevrange recent: %w", err)
	}

	if len(recentMembers) == 0 {
		stats.RecentCalls = []LLMCallMetric{}
		return stats, nil
	}

	stats.RecentCalls = make([]LLMCallMetric, 0, len(recentMembers))
	var totalLatency int64
	var errorCount int64

	for _, member := range recentMembers {
		var m LLMCallMetric
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
