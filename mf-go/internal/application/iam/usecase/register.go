package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/application/iam/dto"
	iamEvent "github.com/masterfabric-go/masterfabric/internal/domain/iam/event"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/repository"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/service"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
)

// RegisterUseCase handles user registration.
type RegisterUseCase struct {
	userRepo repository.UserRepository
	auth     service.AuthService
	eventBus events.EventBus
}

// NewRegisterUseCase creates a new RegisterUseCase.
func NewRegisterUseCase(userRepo repository.UserRepository, auth service.AuthService, eventBus events.EventBus) *RegisterUseCase {
	return &RegisterUseCase{userRepo: userRepo, auth: auth, eventBus: eventBus}
}

// Execute registers a new user.
func (uc *RegisterUseCase) Execute(ctx context.Context, req dto.RegisterRequest) (*dto.UserInfo, error) {
	// Check if user already exists
	existing, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, domainErr.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domainErr.New(domainErr.ErrAlreadyExists, "user with this email already exists", nil)
	}

	// Hash password
	hash, err := uc.auth.HashPassword(req.Password)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to hash password", err)
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: hash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Status:       model.UserStatusActive,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Publish domain event to Kafka
	_ = uc.eventBus.Publish(ctx, events.TopicIAM, iamEvent.UserRegistered{
		UserID:    user.ID,
		Email:     user.Email,
		Timestamp: time.Now().UTC(),
	})

	return &dto.UserInfo{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt,
	}, nil
}
