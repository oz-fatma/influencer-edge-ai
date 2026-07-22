package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/apimanagement/dto"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/repository"
)

// UpdatePolicyUseCase handles endpoint policy updates.
type UpdatePolicyUseCase struct {
	policyRepo repository.PolicyRepository
}

// NewUpdatePolicyUseCase creates a new UpdatePolicyUseCase.
func NewUpdatePolicyUseCase(policyRepo repository.PolicyRepository) *UpdatePolicyUseCase {
	return &UpdatePolicyUseCase{policyRepo: policyRepo}
}

// Execute creates or updates a policy for an endpoint.
func (uc *UpdatePolicyUseCase) Execute(ctx context.Context, endpointID uuid.UUID, req dto.UpdatePolicyRequest) (*dto.PolicyInfo, error) {
	existing, _ := uc.policyRepo.GetByEndpointID(ctx, endpointID)

	authPolicy := req.AuthPolicy
	if authPolicy == "" {
		authPolicy = "jwt"
	}

	if existing != nil {
		// Update
		existing.RequiredPermission = req.RequiredPermission
		existing.RateLimit = req.RateLimit
		existing.AuthPolicy = authPolicy
		existing.ValidationPolicy = req.ValidationPolicy
		existing.ExtraPolicies = req.ExtraPolicies

		if err := uc.policyRepo.Update(ctx, existing); err != nil {
			return nil, err
		}

		return &dto.PolicyInfo{
			ID:                 existing.ID,
			EndpointID:         existing.EndpointID,
			RequiredPermission: existing.RequiredPermission,
			RateLimit:          existing.RateLimit,
			AuthPolicy:         existing.AuthPolicy,
			ValidationPolicy:   existing.ValidationPolicy,
			ExtraPolicies:      existing.ExtraPolicies,
		}, nil
	}

	// Create
	policy := &model.EndpointPolicy{
		EndpointID:         endpointID,
		RequiredPermission: req.RequiredPermission,
		RateLimit:          req.RateLimit,
		AuthPolicy:         authPolicy,
		ValidationPolicy:   req.ValidationPolicy,
		ExtraPolicies:      req.ExtraPolicies,
	}

	if err := uc.policyRepo.Create(ctx, policy); err != nil {
		return nil, err
	}

	return &dto.PolicyInfo{
		ID:                 policy.ID,
		EndpointID:         policy.EndpointID,
		RequiredPermission: policy.RequiredPermission,
		RateLimit:          policy.RateLimit,
		AuthPolicy:         policy.AuthPolicy,
		ValidationPolicy:   policy.ValidationPolicy,
		ExtraPolicies:      policy.ExtraPolicies,
	}, nil
}
