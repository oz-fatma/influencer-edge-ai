package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email is required")
	}
	if len(email) > 255 {
		return errors.New("email must be at most 255 characters")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 128 {
		return errors.New("password must be at most 128 characters")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must contain uppercase, lowercase, and digit")
	}
	return nil
}

func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name is required")
	}
	if len(name) > 120 {
		return errors.New("name must be at most 120 characters")
	}
	return nil
}

func ValidatePlatform(platform string) error {
	platform = strings.TrimSpace(platform)
	if platform == "" {
		return errors.New("platform is required")
	}
	allowed := map[string]bool{
		"instagram": true, "tiktok": true, "youtube": true,
		"twitter": true, "linkedin": true, "other": true,
	}
	if !allowed[strings.ToLower(platform)] {
		return errors.New("platform must be one of: instagram, tiktok, youtube, twitter, linkedin, other")
	}
	return nil
}

func ValidateInfluencerName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("influencer_name is required")
	}
	if len(name) > 255 {
		return errors.New("influencer_name must be at most 255 characters")
	}
	return nil
}

func ValidateScore(score float64) error {
	if score < 0 || score > 100 {
		return errors.New("score must be between 0 and 100")
	}
	return nil
}
