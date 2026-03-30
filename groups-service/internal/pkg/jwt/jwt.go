package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/student-pm/groups-service/internal/domain"
)

// Verifier парсит access-токены, выпущенные auth-service.
// Использует тот же HS256-секрет (читается из общего JWT_SECRET).
//
// Этот сервис НЕ выпускает токены — только проверяет.
type Verifier struct {
	secret []byte
}

func New(secret string) (*Verifier, error) {
	if len(secret) < 16 {
		return nil, errors.New("jwt secret must be at least 16 chars")
	}
	return &Verifier{secret: []byte(secret)}, nil
}

// claims — формат, совместимый с auth-service/internal/pkg/jwt.
type claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// ParseAccess — извлекает userID и роль из подписанного токена.
func (v *Verifier) ParseAccess(tokenStr string) (uuid.UUID, domain.Role, error) {
	c := &claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return v.secret, nil
	})
	if err != nil || !tok.Valid {
		return uuid.Nil, "", domain.ErrInvalidToken
	}
	id, err := uuid.Parse(c.Subject)
	if err != nil {
		return uuid.Nil, "", domain.ErrInvalidToken
	}
	role := domain.Role(c.Role)
	if !role.IsValid() {
		return uuid.Nil, "", domain.ErrInvalidToken
	}
	return id, role, nil
}
