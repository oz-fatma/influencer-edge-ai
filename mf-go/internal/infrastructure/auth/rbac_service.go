package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/repository"
	"github.com/redis/go-redis/v9"
)

const permissionCacheTTL = 15 * time.Minute

// RBACServiceImpl implements service.RBACService with Redis caching.
type RBACServiceImpl struct {
	roleRepo repository.RoleRepository
	redis    *redis.Client
}

// NewRBACService creates a new RBACServiceImpl.
func NewRBACService(roleRepo repository.RoleRepository, redisClient *redis.Client) *RBACServiceImpl {
	return &RBACServiceImpl{roleRepo: roleRepo, redis: redisClient}
}

func (s *RBACServiceImpl) cacheKey(userID, orgID uuid.UUID) string {
	return fmt.Sprintf("user:%s:org:%s:permissions", userID, orgID)
}

func (s *RBACServiceImpl) GetUserPermissions(ctx context.Context, userID, orgID uuid.UUID) ([]string, error) {
	// Try cache first
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, s.cacheKey(userID, orgID)).Result()
		if err == nil {
			var perms []string
			if json.Unmarshal([]byte(cached), &perms) == nil {
				return perms, nil
			}
		}
	}

	// Fetch from DB
	perms, err := s.roleRepo.GetUserPermissions(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}

	// Cache result
	if s.redis != nil && len(perms) > 0 {
		data, _ := json.Marshal(perms)
		_ = s.redis.Set(ctx, s.cacheKey(userID, orgID), data, permissionCacheTTL).Err()
	}

	return perms, nil
}

func (s *RBACServiceImpl) HasPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) (bool, error) {
	perms, err := s.GetUserPermissions(ctx, userID, orgID)
	if err != nil {
		return false, err
	}
	for _, p := range perms {
		if matchesPermission(p, permission) {
			return true, nil
		}
	}
	return false, nil
}

func (s *RBACServiceImpl) HasAnyPermission(ctx context.Context, userID, orgID uuid.UUID, permissions []string) (bool, error) {
	perms, err := s.GetUserPermissions(ctx, userID, orgID)
	if err != nil {
		return false, err
	}
	for _, granted := range perms {
		for _, required := range permissions {
			if matchesPermission(granted, required) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *RBACServiceImpl) InvalidateCache(ctx context.Context, userID, orgID uuid.UUID) error {
	if s.redis != nil {
		return s.redis.Del(ctx, s.cacheKey(userID, orgID)).Err()
	}
	return nil
}

func matchesPermission(granted, required string) bool {
	if granted == required || granted == "*" {
		return true
	}

	if strings.HasSuffix(granted, ":*") {
		prefix := strings.TrimSuffix(granted, ":*")
		if required == prefix || strings.HasPrefix(required, prefix+":") {
			return true
		}
	}

	if granted == "*:read" && strings.HasSuffix(required, ":read") {
		return true
	}

	return false
}
