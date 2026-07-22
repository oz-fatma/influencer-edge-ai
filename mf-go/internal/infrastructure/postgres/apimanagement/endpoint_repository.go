package apimanagement

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/model"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

// EndpointRepo implements repository.EndpointRepository with PostgreSQL.
type EndpointRepo struct {
	db *pgxpool.Pool
}

// NewEndpointRepo creates a new EndpointRepo.
func NewEndpointRepo(db *pgxpool.Pool) *EndpointRepo {
	return &EndpointRepo{db: db}
}

func (r *EndpointRepo) Create(ctx context.Context, ep *model.Endpoint) error {
	if ep.ID == uuid.Nil {
		ep.ID = uuid.New()
	}
	now := time.Now().UTC()
	ep.CreatedAt = now
	ep.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO app_endpoints (id, app_id, method, path, version, backend_service, backend_action, schema, audit_level, pii_masking, event_after, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		ep.ID, ep.AppID, ep.Method, ep.Path, ep.Version, ep.BackendService, ep.BackendAction,
		ep.Schema, ep.AuditLevel, ep.PIIMasking, ep.EventAfter, ep.Status, ep.CreatedAt, ep.UpdatedAt,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to create endpoint", err)
	}
	return nil
}

func (r *EndpointRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Endpoint, error) {
	var ep model.Endpoint
	err := r.db.QueryRow(ctx,
		`SELECT id, app_id, method, path, version, backend_service, backend_action, schema, audit_level, pii_masking, event_after, status, created_at, updated_at
		 FROM app_endpoints WHERE id=$1`, id,
	).Scan(&ep.ID, &ep.AppID, &ep.Method, &ep.Path, &ep.Version, &ep.BackendService, &ep.BackendAction,
		&ep.Schema, &ep.AuditLevel, &ep.PIIMasking, &ep.EventAfter, &ep.Status, &ep.CreatedAt, &ep.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "endpoint not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get endpoint", err)
	}
	return &ep, nil
}

func (r *EndpointRepo) GetByMethodPath(ctx context.Context, appID uuid.UUID, method, path, version string) (*model.Endpoint, error) {
	var ep model.Endpoint
	err := r.db.QueryRow(ctx,
		`SELECT id, app_id, method, path, version, backend_service, backend_action, schema, audit_level, pii_masking, event_after, status, created_at, updated_at
		 FROM app_endpoints WHERE app_id=$1 AND method=$2 AND path=$3 AND version=$4`, appID, method, path, version,
	).Scan(&ep.ID, &ep.AppID, &ep.Method, &ep.Path, &ep.Version, &ep.BackendService, &ep.BackendAction,
		&ep.Schema, &ep.AuditLevel, &ep.PIIMasking, &ep.EventAfter, &ep.Status, &ep.CreatedAt, &ep.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "endpoint not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get endpoint by method/path", err)
	}
	return &ep, nil
}

func (r *EndpointRepo) Update(ctx context.Context, ep *model.Endpoint) error {
	ep.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`UPDATE app_endpoints SET method=$1, path=$2, version=$3, backend_service=$4, backend_action=$5,
		 schema=$6, audit_level=$7, pii_masking=$8, event_after=$9, status=$10, updated_at=$11 WHERE id=$12`,
		ep.Method, ep.Path, ep.Version, ep.BackendService, ep.BackendAction,
		ep.Schema, ep.AuditLevel, ep.PIIMasking, ep.EventAfter, ep.Status, ep.UpdatedAt, ep.ID,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to update endpoint", err)
	}
	return nil
}

func (r *EndpointRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM app_endpoints WHERE id=$1`, id)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to delete endpoint", err)
	}
	return nil
}

func (r *EndpointRepo) ListByApp(ctx context.Context, appID uuid.UUID, offset, limit int) ([]*model.Endpoint, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM app_endpoints WHERE app_id=$1`, appID).Scan(&total)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to count endpoints", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, app_id, method, path, version, backend_service, backend_action, schema, audit_level, pii_masking, event_after, status, created_at, updated_at
		 FROM app_endpoints WHERE app_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, appID, limit, offset,
	)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to list endpoints", err)
	}
	defer rows.Close()

	var endpoints []*model.Endpoint
	for rows.Next() {
		var ep model.Endpoint
		if err := rows.Scan(&ep.ID, &ep.AppID, &ep.Method, &ep.Path, &ep.Version, &ep.BackendService, &ep.BackendAction,
			&ep.Schema, &ep.AuditLevel, &ep.PIIMasking, &ep.EventAfter, &ep.Status, &ep.CreatedAt, &ep.UpdatedAt); err != nil {
			return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to scan endpoint", err)
		}
		endpoints = append(endpoints, &ep)
	}
	return endpoints, total, nil
}
