package usecase

import (
	"context"

	"github.com/masterfabric-go/masterfabric/internal/application/iam/dto"
	adminRepo "github.com/masterfabric-go/masterfabric/internal/domain/admin/repository"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/repository"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/service"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

// LoginUseCase handles user authentication.
type LoginUseCase struct {
	userRepo  repository.UserRepository
	auth      service.AuthService
	adminRepo adminRepo.AdminRepository
}

// NewLoginUseCase creates a new LoginUseCase.
func NewLoginUseCase(userRepo repository.UserRepository, auth service.AuthService, adminRepo adminRepo.AdminRepository) *LoginUseCase {
	return &LoginUseCase{userRepo: userRepo, auth: auth, adminRepo: adminRepo}
}

// Execute authenticates a user and returns a JWT token.
func (uc *LoginUseCase) Execute(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrUnauthorized, "invalid credentials", nil)
	}

	if !user.IsActive() {
		return nil, domainErr.New(domainErr.ErrForbidden, "account is not active", nil)
	}

	if err := uc.auth.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, err
	}

	token, err := uc.auth.GenerateToken(ctx, service.TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
	})
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to generate token", err)
	}

	isAdmin := false
	if uc.adminRepo != nil {
		isAdmin, _ = uc.adminRepo.IsAdmin(ctx, user.ID)
	}

	return &dto.LoginResponse{
		Token: token,
		User: dto.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Status:    string(user.Status),
			CreatedAt: user.CreatedAt,
		},
		IsAdmin: isAdmin,
	}, nil
}
