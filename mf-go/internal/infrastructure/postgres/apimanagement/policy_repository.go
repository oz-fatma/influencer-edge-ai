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

// PolicyRepo implements repository.PolicyRepository with PostgreSQL.
type PolicyRepo struct {
	db *pgxpool.Pool
}

// NewPolicyRepo creates a new PolicyRepo.
func NewPolicyRepo(db *pgxpool.Pool) *PolicyRepo {
	return &PolicyRepo{db: db}
}

func (r *PolicyRepo) Create(ctx context.Context, p *model.EndpointPolicy) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO app_endpoint_policies (id, endpoint_id, required_permission, rate_limit, auth_policy, validation_policy, extra_policies, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.EndpointID, p.RequiredPermission, p.RateLimit, p.AuthPolicy, p.ValidationPolicy, p.ExtraPolicies, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to create policy", err)
	}
	return nil
}

func (r *PolicyRepo) GetByEndpointID(ctx context.Context, endpointID uuid.UUID) (*model.EndpointPolicy, error) {
	var p model.EndpointPolicy
	err := r.db.QueryRow(ctx,
		`SELECT id, endpoint_id, required_permission, rate_limit, auth_policy, validation_policy, extra_policies, created_at, updated_at
		 FROM app_endpoint_policies WHERE endpoint_id=$1`, endpointID,
	).Scan(&p.ID, &p.EndpointID, &p.RequiredPermission, &p.RateLimit, &p.AuthPolicy, &p.ValidationPolicy, &p.ExtraPolicies, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "policy not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get policy", err)
	}
	return &p, nil
}

func (r *PolicyRepo) Update(ctx context.Context, p *model.EndpointPolicy) error {
	p.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`UPDATE app_endpoint_policies SET required_permission=$1, rate_limit=$2, auth_policy=$3, validation_policy=$4, extra_policies=$5, updated_at=$6 WHERE id=$7`,
		p.RequiredPermission, p.RateLimit, p.AuthPolicy, p.ValidationPolicy, p.ExtraPolicies, p.UpdatedAt, p.ID,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to update policy", err)
	}
	return nil
}

func (r *PolicyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM app_endpoint_policies WHERE id=$1`, id)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to delete policy", err)
	}
	return nil
}
