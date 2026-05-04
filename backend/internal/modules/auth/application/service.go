package application

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"reserveflow/backend/internal/infrastructure/auth"
	apperrors "reserveflow/backend/internal/infrastructure/errors"
	"reserveflow/backend/internal/modules/auth/domain"
)

type Repository interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	CreateRefreshToken(ctx context.Context, token domain.RefreshToken) error
	GetRefreshToken(ctx context.Context, id string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id string, revokedAt time.Time) error
}

type Service struct {
	repo Repository
	jwt  *auth.Service
}

type TokenPair struct {
	AccessToken           string    `json:"accessToken"`
	RefreshToken          string    `json:"refreshToken"`
	AccessTokenExpiresAt  time.Time `json:"accessTokenExpiresAt"`
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
}

type AuthResult struct {
	User   UserDTO   `json:"user"`
	Tokens TokenPair `json:"tokens"`
}

type UserDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewService(repo Repository, jwt *auth.Service) *Service {
	return &Service{repo: repo, jwt: jwt}
}

func (s *Service) Register(ctx context.Context, email, password, name string) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	if email == "" || !strings.Contains(email, "@") || len(password) < 8 || name == "" {
		return nil, apperrors.Validation("Invalid registration payload")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	now := time.Now().UTC()
	user := domain.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return nil, apperrors.Conflict(apperrors.CodeValidationError, "User with this email already exists")
		}
		return nil, apperrors.Internal(err)
	}
	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, err
	}
	return &AuthResult{User: toDTO(user), Tokens: *tokens}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.Unauthorized("Invalid email or password")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, apperrors.Unauthorized("Invalid email or password")
	}
	tokens, err := s.issueTokens(ctx, *user)
	if err != nil {
		return nil, err
	}
	return &AuthResult{User: toDTO(*user), Tokens: *tokens}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, apperrors.Unauthorized("Invalid refresh token")
	}
	stored, err := s.repo.GetRefreshToken(ctx, claims.ID)
	if err != nil {
		return nil, apperrors.Unauthorized("Invalid refresh token")
	}
	if !stored.Active(time.Now().UTC(), auth.HashToken(refreshToken)) {
		return nil, apperrors.Unauthorized("Refresh token is expired or revoked")
	}
	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, apperrors.Unauthorized("Invalid refresh token")
	}
	if err := s.repo.RevokeRefreshToken(ctx, stored.ID, time.Now().UTC()); err != nil {
		return nil, apperrors.Internal(err)
	}
	return s.issueTokens(ctx, *user)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return apperrors.Unauthorized("Invalid refresh token")
	}
	if err := s.repo.RevokeRefreshToken(ctx, claims.ID, time.Now().UTC()); err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

func (s *Service) Me(ctx context.Context, userID string) (*UserDTO, error) {
	if userID == "" {
		return nil, apperrors.Unauthorized("Missing user context")
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "User not found", http.StatusNotFound)
	}
	dto := toDTO(*user)
	return &dto, nil
}

func (s *Service) issueTokens(ctx context.Context, user domain.User) (*TokenPair, error) {
	access, accessExp, err := s.jwt.NewAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	refresh, refreshID, refreshExp, err := s.jwt.NewRefreshToken(user.ID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if err := s.repo.CreateRefreshToken(ctx, domain.RefreshToken{
		ID:        refreshID,
		UserID:    user.ID,
		TokenHash: auth.HashToken(refresh),
		ExpiresAt: refreshExp,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		return nil, apperrors.Internal(err)
	}
	return &TokenPair{
		AccessToken:           access,
		RefreshToken:          refresh,
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
	}, nil
}

func toDTO(user domain.User) UserDTO {
	return UserDTO{ID: user.ID, Email: user.Email, Name: user.Name, Role: user.Role, CreatedAt: user.CreatedAt}
}
