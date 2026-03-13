package httpdelivery

import "time"

// RegisterRequest — запрос на регистрацию.
type RegisterRequest struct {
	Email    string `json:"email"     validate:"required,email,max=255"   example:"ivanov@uni.edu"`
	Password string `json:"password"  validate:"required,min=8,max=72"    example:"Pa$$w0rd123"`
	FullName string `json:"full_name" validate:"required,min=2,max=120"   example:"Иван Иванов"`
	Role     string `json:"role"      validate:"required,oneof=student group_leader teacher admin" example:"student"`
}

// LoginRequest — вход.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"  example:"ivanov@uni.edu"`
	Password string `json:"password" validate:"required,min=1"  example:"Pa$$w0rd123"`
}

// RefreshRequest — обмен refresh-токена.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"AbCdEf..."`
}

// UpdateUserRequest — частичное обновление профиля.
type UpdateUserRequest struct {
	FullName *string `json:"full_name,omitempty" validate:"omitempty,min=2,max=120"`
	GroupID  *string `json:"group_id,omitempty"  validate:"omitempty,uuid"`
	Role     *string `json:"role,omitempty"      validate:"omitempty,oneof=student group_leader teacher admin"`
}

// UserResponse — публичная проекция пользователя.
type UserResponse struct {
	ID        string    `json:"id"        example:"7c0a..."`
	Email     string    `json:"email"     example:"ivanov@uni.edu"`
	FullName  string    `json:"full_name" example:"Иван Иванов"`
	Role      string    `json:"role"      example:"student"`
	GroupID   *string   `json:"group_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthResponse — ответ на login/register/refresh.
type AuthResponse struct {
	User             UserResponse `json:"user"`
	AccessToken      string       `json:"access_token"`
	AccessExpiresAt  time.Time    `json:"access_expires_at"`
	RefreshToken     string       `json:"refresh_token"`
	RefreshExpiresAt time.Time    `json:"refresh_expires_at"`
}

// MessageResponse — простой ответ "ok".
type MessageResponse struct {
	Message string `json:"message" example:"ok"`
}
