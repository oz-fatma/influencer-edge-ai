package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/service"
	"github.com/masterfabric-go/masterfabric/internal/shared/config"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	"golang.org/x/crypto/bcrypt"
)

// JWTService implements service.AuthService using JWT and bcrypt.
type JWTService struct {
	secret     []byte
	expiration time.Duration
	issuer     string
}

// NewJWTService creates a new JWTService from config.
func NewJWTService(cfg config.JWTConfig) *JWTService {
	return &JWTService{
		secret:     []byte(cfg.Secret),
		expiration: time.Duration(cfg.ExpirationHours) * time.Hour,
		issuer:     cfg.Issuer,
	}
}

// customClaims extends JWT standard claims with our domain data.
type customClaims struct {
	jwt.RegisteredClaims
	UserID         string   `json:"user_id"`
	Email          string   `json:"email"`
	OrganizationID string   `json:"organization_id,omitempty"`
	Roles          []string `json:"roles,omitempty"`
	Permissions    []string `json:"permissions,omitempty"`
}

func (s *JWTService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func (s *JWTService) VerifyPassword(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return domainErr.New(domainErr.ErrUnauthorized, "invalid credentials", nil)
	}
	return nil
}

func (s *JWTService) GenerateToken(_ context.Context, claims service.TokenClaims) (string, error) {
	now := time.Now()
	c := customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   claims.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		UserID:         claims.UserID.String(),
		Email:          claims.Email,
		OrganizationID: claims.OrganizationID.String(),
		Roles:          claims.Roles,
		Permissions:    claims.Permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (s *JWTService) ValidateToken(_ context.Context, tokenStr string) (*service.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &customClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, domainErr.New(domainErr.ErrUnauthorized, "invalid token", err)
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return nil, domainErr.New(domainErr.ErrUnauthorized, "invalid token claims", nil)
	}

	userID, _ := uuid.Parse(claims.UserID)
	orgID, _ := uuid.Parse(claims.OrganizationID)

	return &service.TokenClaims{
		UserID:         userID,
		Email:          claims.Email,
		OrganizationID: orgID,
		Roles:          claims.Roles,
		Permissions:    claims.Permissions,
	}, nil
}
