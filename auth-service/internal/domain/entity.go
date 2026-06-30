package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role — роль пользователя в системе.
type Role string

const (
	RoleStudent     Role = "student"
	RoleGroupLeader Role = "group_leader"
	RoleTeacher     Role = "teacher"
	RoleAdmin       Role = "admin"
)

// IsValid проверяет, что значение роли — одно из допустимых.
func (r Role) IsValid() bool {
	switch r {
	case RoleStudent, RoleGroupLeader, RoleTeacher, RoleAdmin:
		return true
	}
	return false
}

// User — основная сущность сервиса аутентификации.
type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // не сериализуется в JSON
	FullName     string     `json:"full_name"`
	Role         Role       `json:"role"`
	GroupID      *uuid.UUID `json:"group_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// RefreshToken — выданный refresh-токен. Храним хэш, а не сам токен.
type RefreshToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
	UserAgent  string
	IPAddress  string
}

// IsActive — токен ещё валиден и не отозван.
func (rt *RefreshToken) IsActive(now time.Time) bool {
	if rt.RevokedAt != nil {
		return false
	}
	return now.Before(rt.ExpiresAt)
}
