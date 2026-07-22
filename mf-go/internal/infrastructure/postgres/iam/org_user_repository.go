package iam

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/model"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

// OrgUserRepo implements repository.OrgUserRepository with PostgreSQL.
type OrgUserRepo struct {
	db *pgxpool.Pool
}

// NewOrgUserRepo creates a new OrgUserRepo.
func NewOrgUserRepo(db *pgxpool.Pool) *OrgUserRepo {
	return &OrgUserRepo{db: db}
}

func (r *OrgUserRepo) Add(ctx context.Context, orgUser *model.OrganizationUser) error {
	now := time.Now().UTC()
	orgUser.CreatedAt = now
	orgUser.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO organization_users (organization_id, user_id, status, invited_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (organization_id, user_id) DO UPDATE SET status=$3, updated_at=$6`,
		orgUser.OrganizationID, orgUser.UserID, orgUser.Status, orgUser.InvitedBy, orgUser.CreatedAt, orgUser.UpdatedAt,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to add org user", err)
	}
	return nil
}

func (r *OrgUserRepo) Remove(ctx context.Context, orgID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM organization_users WHERE organization_id=$1 AND user_id=$2`, orgID, userID,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to remove org user", err)
	}
	return nil
}

func (r *OrgUserRepo) GetByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*model.OrganizationUser, error) {
	var ou model.OrganizationUser
	err := r.db.QueryRow(ctx,
		`SELECT organization_id, user_id, status, invited_by, created_at, updated_at
		 FROM organization_users WHERE organization_id=$1 AND user_id=$2`, orgID, userID,
	).Scan(&ou.OrganizationID, &ou.UserID, &ou.Status, &ou.InvitedBy, &ou.CreatedAt, &ou.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "organization user not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get org user", err)
	}
	return &ou, nil
}

func (r *OrgUserRepo) ListByOrg(ctx context.Context, orgID uuid.UUID, offset, limit int) ([]*model.OrganizationUser, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM organization_users WHERE organization_id=$1`, orgID,
	).Scan(&total)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to count org users", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT organization_id, user_id, status, invited_by, created_at, updated_at
		 FROM organization_users WHERE organization_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		orgID, limit, offset,
	)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to list org users", err)
	}
	defer rows.Close()

	var users []*model.OrganizationUser
	for rows.Next() {
		var ou model.OrganizationUser
		if err := rows.Scan(&ou.OrganizationID, &ou.UserID, &ou.Status, &ou.InvitedBy, &ou.CreatedAt, &ou.UpdatedAt); err != nil {
			return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to scan org user", err)
		}
		users = append(users, &ou)
	}
	return users, total, nil
}

func (r *OrgUserRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.OrganizationUser, error) {
	rows, err := r.db.Query(ctx,
		`SELECT organization_id, user_id, status, invited_by, created_at, updated_at
		 FROM organization_users WHERE user_id=$1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to list user orgs", err)
	}
	defer rows.Close()

	var orgs []*model.OrganizationUser
	for rows.Next() {
		var ou model.OrganizationUser
		if err := rows.Scan(&ou.OrganizationID, &ou.UserID, &ou.Status, &ou.InvitedBy, &ou.CreatedAt, &ou.UpdatedAt); err != nil {
			return nil, domainErr.New(domainErr.ErrInternal, "failed to scan user org", err)
		}
		orgs = append(orgs, &ou)
	}
	return orgs, nil
}
