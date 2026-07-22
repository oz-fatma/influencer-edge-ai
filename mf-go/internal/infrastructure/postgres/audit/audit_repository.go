package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/masterfabric-go/masterfabric/internal/domain/audit/model"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

// AuditRepo implements repository.AuditRepository with PostgreSQL.
type AuditRepo struct {
	db *pgxpool.Pool
}

// NewAuditRepo creates a new AuditRepo.
func NewAuditRepo(db *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) Create(ctx context.Context, log *model.AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	log.CreatedAt = time.Now().UTC()

	_, err := r.db.Exec(ctx,
		`INSERT INTO audit_logs (id, organization_id, app_id, endpoint_id, user_id, request_id, action, resource_type, resource_id, metadata, ip_address, user_agent, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		log.ID, log.OrganizationID, log.AppID, log.EndpointID, log.UserID,
		log.RequestID, log.Action, log.ResourceType, log.ResourceID,
		log.Metadata, log.IPAddress, log.UserAgent, log.CreatedAt,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to create audit log", err)
	}
	return nil
}

func (r *AuditRepo) ListByOrg(ctx context.Context, orgID uuid.UUID, offset, limit int) ([]*model.AuditLog, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE organization_id=$1`, orgID).Scan(&total); err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to count audit logs", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, organization_id, app_id, endpoint_id, user_id, request_id, action, resource_type, resource_id, metadata, ip_address, user_agent, created_at
		 FROM audit_logs WHERE organization_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, orgID, limit, offset,
	)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to list audit logs", err)
	}
	defer rows.Close()

	return r.scanLogs(rows, total)
}

func (r *AuditRepo) ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*model.AuditLog, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE user_id=$1`, userID).Scan(&total); err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to count audit logs", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, organization_id, app_id, endpoint_id, user_id, request_id, action, resource_type, resource_id, metadata, ip_address, user_agent, created_at
		 FROM audit_logs WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, limit, offset,
	)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to list audit logs", err)
	}
	defer rows.Close()

	return r.scanLogs(rows, total)
}

func (r *AuditRepo) ListByResource(ctx context.Context, resourceType, resourceID string, offset, limit int) ([]*model.AuditLog, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE resource_type=$1 AND resource_id=$2`, resourceType, resourceID).Scan(&total); err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to count audit logs", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, organization_id, app_id, endpoint_id, user_id, request_id, action, resource_type, resource_id, metadata, ip_address, user_agent, created_at
		 FROM audit_logs WHERE resource_type=$1 AND resource_id=$2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`, resourceType, resourceID, limit, offset,
	)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to list audit logs", err)
	}
	defer rows.Close()

	return r.scanLogs(rows, total)
}

func (r *AuditRepo) scanLogs(rows interface{ Next() bool; Scan(dest ...interface{}) error }, total int) ([]*model.AuditLog, int, error) {
	var logs []*model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		if err := rows.Scan(&l.ID, &l.OrganizationID, &l.AppID, &l.EndpointID, &l.UserID,
			&l.RequestID, &l.Action, &l.ResourceType, &l.ResourceID,
			&l.Metadata, &l.IPAddress, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to scan audit log", err)
		}
		logs = append(logs, &l)
	}
	return logs, total, nil
}
