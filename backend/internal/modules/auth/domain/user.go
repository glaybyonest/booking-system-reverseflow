package domain

import (
	"errors"
	"time"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrRefreshRevoked     = errors.New("refresh token revoked")
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

func (t RefreshToken) Active(now time.Time, hash string) bool {
	return t.RevokedAt == nil && t.ExpiresAt.After(now) && t.TokenHash == hash
}
