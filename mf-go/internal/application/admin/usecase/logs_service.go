package usecase

import (
	"context"

	"github.com/masterfabric-go/masterfabric/internal/application/admin/dto"
	adminRepo "github.com/masterfabric-go/masterfabric/internal/domain/admin/repository"
)

type LogsService struct {
	logs adminRepo.LLMRequestLogRepository
}

func NewLogsService(logs adminRepo.LLMRequestLogRepository) *LogsService {
	return &LogsService{logs: logs}
}

func (s *LogsService) ListRecent(ctx context.Context, limit int) (*dto.LLMLogsResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	items, err := s.logs.ListRecent(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]dto.LLMLogEntry, 0, len(items))
	for _, item := range items {
		out = append(out, dto.LLMLogEntry{
			ModelName:  item.ModelName,
			DurationMs: item.DurationMs,
			Success:    item.Success,
			CreatedAt:  item.CreatedAt,
		})
	}
	return &dto.LLMLogsResponse{Logs: out}, nil
}
