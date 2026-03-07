package domain

import "errors"

// Доменные ошибки. На границе HTTP-слоя они мапятся в коды ответа.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidRole        = errors.New("invalid role")
)
