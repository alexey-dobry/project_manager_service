package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/student-pm/auth-service/internal/domain"
)

// UserRepository — порт доступа к пользователям.
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
}

// RefreshTokenRepository — хранение refresh-токенов.
type RefreshTokenRepository interface {
	Save(ctx context.Context, t *domain.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID, at time.Time) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID, at time.Time) error
}

// PasswordHasher — порт для bcrypt и подобных.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// TokenProvider — выпуск/проверка JWT.
type TokenProvider interface {
	GenerateAccess(userID uuid.UUID, role domain.Role) (token string, expiresAt time.Time, err error)
	GenerateRefresh() (token string, hash string, expiresAt time.Time, err error)
	HashRefresh(token string) string
	ParseAccess(token string) (userID uuid.UUID, role domain.Role, err error)
}
