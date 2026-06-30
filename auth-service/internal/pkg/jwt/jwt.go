package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/student-pm/auth-service/internal/domain"
)

// Provider — реализация usecase.TokenProvider на HS256.
// Refresh-токен — это случайные 32 байта (URL-safe base64); в БД хранится sha256-хэш.
type Provider struct {
	secret      []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
	issuer      string
}

// Config — параметры провайдера.
type Config struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
}

func New(cfg Config) (*Provider, error) {
	if len(cfg.Secret) < 16 {
		return nil, errors.New("jwt secret must be at least 16 chars")
	}
	if cfg.AccessTTL <= 0 {
		cfg.AccessTTL = 15 * time.Minute
	}
	if cfg.RefreshTTL <= 0 {
		cfg.RefreshTTL = 30 * 24 * time.Hour
	}
	if cfg.Issuer == "" {
		cfg.Issuer = "auth-service"
	}
	return &Provider{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		issuer:     cfg.Issuer,
	}, nil
}

// claims — содержимое access-токена.
type claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (p *Provider) GenerateAccess(userID uuid.UUID, role domain.Role) (string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(p.accessTTL)
	c := claims{
		Role: string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    p.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := tok.SignedString(p.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

func (p *Provider) ParseAccess(tokenStr string) (uuid.UUID, domain.Role, error) {
	c := &claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return p.secret, nil
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

// GenerateRefresh — случайный URL-safe токен. Возвращает (token, hash, exp).
func (p *Provider) GenerateRefresh() (string, string, time.Time, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", time.Time{}, err
	}
	tok := base64.RawURLEncoding.EncodeToString(buf)
	return tok, p.HashRefresh(tok), time.Now().UTC().Add(p.refreshTTL), nil
}

// HashRefresh — детерминированный sha256-хэш для поиска по БД.
func (p *Provider) HashRefresh(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
