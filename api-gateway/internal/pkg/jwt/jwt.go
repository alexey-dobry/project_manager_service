package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Verifier — проверяет access-токены, выпущенные auth-service.
type Verifier struct {
	secret []byte
}

func New(secret string) (*Verifier, error) {
	if len(secret) < 16 {
		return nil, errors.New("jwt secret must be at least 16 chars")
	}
	return &Verifier{secret: []byte(secret)}, nil
}

// ErrInvalidToken — единственная ошибка, которую отдаёт верификатор.
var ErrInvalidToken = errors.New("invalid or expired token")

type claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// ParseAccess — извлекает userID и роль из подписанного токена.
// Gateway не требует знать enum ролей — этим занимается backend.
func (v *Verifier) ParseAccess(tokenStr string) (userID uuid.UUID, role string, err error) {
	c := &claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return v.secret, nil
	})
	if err != nil || !tok.Valid {
		return uuid.Nil, "", ErrInvalidToken
	}
	id, err := uuid.Parse(c.Subject)
	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}
	return id, c.Role, nil
}
