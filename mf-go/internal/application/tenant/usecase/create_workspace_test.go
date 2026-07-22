package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/tenant/dto"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/model"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
	"github.com/masterfabric-go/masterfabric/internal/shared/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockWorkspaceRepo is a mock implementation of WorkspaceRepository
type mockWorkspaceRepo struct {
	mock.Mock
}

func (m *mockWorkspaceRepo) Create(ctx context.Context, workspace *model.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockWorkspaceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Workspace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Workspace), args.Error(1)
}

func (m *mockWorkspaceRepo) GetBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*model.Workspace, error) {
	args := m.Called(ctx, orgID, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Workspace), args.Error(1)
}

func (m *mockWorkspaceRepo) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*model.Workspace, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Workspace), args.Error(1)
}

func (m *mockWorkspaceRepo) Update(ctx context.Context, workspace *model.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockWorkspaceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// mockOrgRepo is a mock implementation of OrgRepository
type mockOrgRepo struct {
	mock.Mock
}

func (m *mockOrgRepo) Create(ctx context.Context, org *model.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *mockOrgRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *mockOrgRepo) GetBySlug(ctx context.Context, slug string) (*model.Organization, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *mockOrgRepo) Update(ctx context.Context, org *model.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *mockOrgRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockOrgRepo) List(ctx context.Context, offset, limit int) ([]*model.Organization, int, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*model.Organization), args.Int(1), args.Error(2)
}

func TestCreateWorkspaceUseCase_Execute_Success(t *testing.T) {
	workspaceRepo := new(mockWorkspaceRepo)
	orgRepo := new(mockOrgRepo)
	eventBus := &mockEventBus{}

	orgID := uuid.New()
	userID := uuid.New()

	org := &model.Organization{
		ID:     orgID,
		Name:   "Test Org",
		Status: model.OrgStatusActive,
	}

	orgRepo.On("GetByID", mock.Anything, orgID).Return(org, nil)
	workspaceRepo.On("GetBySlug", mock.Anything, orgID, "test-workspace").Return(nil, domainErr.ErrNotFound)
	workspaceRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Workspace")).Return(nil)

	uc := NewCreateWorkspaceUseCase(workspaceRepo, orgRepo, eventBus)

	ctx := context.WithValue(context.Background(), middleware.ContextKeyOrganizationID, orgID)
	ctx = context.WithValue(ctx, middleware.ContextKeyUserID, userID)

	req := dto.CreateWorkspaceRequest{
		Name:        "Test Workspace",
		Slug:        "test-workspace",
		Description: "Test description",
	}

	result, err := uc.Execute(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Workspace", result.Name)
	assert.Equal(t, "test-workspace", result.Slug)
	assert.Equal(t, string(model.WorkspaceStatusActive), result.Status)
	assert.Equal(t, orgID, result.OrganizationID)

	workspaceRepo.AssertExpectations(t)
	orgRepo.AssertExpectations(t)
}

func TestCreateWorkspaceUseCase_Execute_SlugAlreadyTaken(t *testing.T) {
	workspaceRepo := new(mockWorkspaceRepo)
	orgRepo := new(mockOrgRepo)
	eventBus := &mockEventBus{}

	orgID := uuid.New()
	existingWorkspace := &model.Workspace{ID: uuid.New()}

	org := &model.Organization{
		ID:     orgID,
		Status: model.OrgStatusActive,
	}

	orgRepo.On("GetByID", mock.Anything, orgID).Return(org, nil)
	workspaceRepo.On("GetBySlug", mock.Anything, orgID, "test-workspace").Return(existingWorkspace, nil)

	uc := NewCreateWorkspaceUseCase(workspaceRepo, orgRepo, eventBus)

	ctx := context.WithValue(context.Background(), middleware.ContextKeyOrganizationID, orgID)

	req := dto.CreateWorkspaceRequest{
		Name: "Test Workspace",
		Slug: "test-workspace",
	}

	result, err := uc.Execute(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, domainErr.ErrAlreadyExists))
}

func TestCreateWorkspaceUseCase_Execute_OrgNotFound(t *testing.T) {
	workspaceRepo := new(mockWorkspaceRepo)
	orgRepo := new(mockOrgRepo)
	eventBus := &mockEventBus{}

	orgID := uuid.New()

	orgRepo.On("GetByID", mock.Anything, orgID).Return(nil, domainErr.ErrNotFound)

	uc := NewCreateWorkspaceUseCase(workspaceRepo, orgRepo, eventBus)

	ctx := context.WithValue(context.Background(), middleware.ContextKeyOrganizationID, orgID)

	req := dto.CreateWorkspaceRequest{
		Name: "Test Workspace",
		Slug: "test-workspace",
	}

	result, err := uc.Execute(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, domainErr.ErrNotFound))
}

// mockEventBus is a simple mock for EventBus
type mockEventBus struct {
	publishedEvents []interface{}
}

func (m *mockEventBus) Publish(ctx context.Context, topic string, event events.Event) error {
	if m.publishedEvents == nil {
		m.publishedEvents = []interface{}{}
	}
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *mockEventBus) Subscribe(topic string, handler events.Handler) {
	// No-op for tests
}

func (m *mockEventBus) Close() error {
	return nil
}
