package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/masterfabric-go/masterfabric/internal/domain/admin/model"
	adminRepo "github.com/masterfabric-go/masterfabric/internal/domain/admin/repository"
	"github.com/masterfabric-go/masterfabric/internal/shared/database"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

type LLMConfigRepo struct {
	db             *pgxpool.Pool
	llmConfigTable string
}

func NewLLMConfigRepo(db *pgxpool.Pool, schema string) *LLMConfigRepo {
	return &LLMConfigRepo{
		db:             db,
		llmConfigTable: database.QualifyTable(schema, "llm_config"),
	}
}

var _ adminRepo.LLMConfigRepository = (*LLMConfigRepo)(nil)

func (r *LLMConfigRepo) Get(ctx context.Context) (*model.LLMConfig, error) {
	singletonID, err := uuid.Parse(model.SingletonConfigID)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "invalid singleton config id", err)
	}
	return r.getByID(ctx, singletonID)
}

func (r *LLMConfigRepo) Update(ctx context.Context, cfg *model.LLMConfig) (*model.LLMConfig, error) {
	singletonID, err := uuid.Parse(model.SingletonConfigID)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "invalid singleton config id", err)
	}
	now := time.Now().UTC()
	tag, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET
			system_prompt = $1,
			temperature = $2,
			max_tokens = $3,
			model = $4,
			updated_at = $5,
			updated_by = $6
		WHERE id = $7`, r.llmConfigTable),
		cfg.SystemPrompt, cfg.Temperature, cfg.MaxTokens, cfg.Model, now, cfg.UpdatedBy, singletonID,
	)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to update llm config", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domainErr.New(domainErr.ErrNotFound, "llm config not found", nil)
	}
	return r.getByID(ctx, singletonID)
}

func (r *LLMConfigRepo) getByID(ctx context.Context, id uuid.UUID) (*model.LLMConfig, error) {
	var cfg model.LLMConfig
	var updatedBy *uuid.UUID
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, system_prompt, temperature, max_tokens, model, updated_at, updated_by
		FROM %s WHERE id = $1`, r.llmConfigTable), id,
	).Scan(&cfg.ID, &cfg.SystemPrompt, &cfg.Temperature, &cfg.MaxTokens, &cfg.Model, &cfg.UpdatedAt, &updatedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "llm config not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get llm config", err)
	}
	cfg.UpdatedBy = updatedBy
	return &cfg, nil
}

type AdminRepo struct {
	db              *pgxpool.Pool
	userRolesTable  string
	rolesTable      string
}

func NewAdminRepo(db *pgxpool.Pool, schema string) *AdminRepo {
	return &AdminRepo{
		db:             db,
		userRolesTable: database.QualifyTable(schema, "user_roles"),
		rolesTable:     database.QualifyTable(schema, "roles"),
	}
}

var _ adminRepo.AdminRepository = (*AdminRepo)(nil)

func (r *AdminRepo) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1
			FROM %s ur
			INNER JOIN %s ro ON ro.id = ur.role_id
			WHERE ur.user_id = $1 AND ro.name = 'admin'
		)`, r.userRolesTable, r.rolesTable), userID,
	).Scan(&exists)
	if err != nil {
		return false, domainErr.New(domainErr.ErrInternal, "failed to check admin role", err)
	}
	return exists, nil
}
