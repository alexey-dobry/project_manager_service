package usecase

import (
	"time"

	"github.com/google/uuid"

	"github.com/student-pm/auth-service/internal/domain"
)

// RegisterInput — входные данные регистрации.
type RegisterInput struct {
	Email    string
	Password string
	FullName string
	Role     domain.Role
}

// LoginInput — данные для входа.
type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

// TokenPair — пара access + refresh, возвращаемая после login/refresh.
type TokenPair struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}

// AuthResult объединяет токены и пользователя.
type AuthResult struct {
	User   *domain.User
	Tokens TokenPair
}

// UpdateUserInput — патч-обновление профиля.
type UpdateUserInput struct {
	FullName *string
	GroupID  *uuid.UUID
	Role     *domain.Role // менять может только admin — проверка в handler
}
