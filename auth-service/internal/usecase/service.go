package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/student-pm/auth-service/internal/domain"
)

// AuthService реализует сценарии аутентификации, зависит только от портов.
type AuthService struct {
	users    UserRepository
	tokens   RefreshTokenRepository
	hasher   PasswordHasher
	jwt      TokenProvider
	now      func() time.Time
}

// NewAuthService — конструктор; при now=nil используется time.Now.
func NewAuthService(
	u UserRepository,
	t RefreshTokenRepository,
	h PasswordHasher,
	j TokenProvider,
	now func() time.Time,
) *AuthService {
	if now == nil {
		now = time.Now
	}
	return &AuthService{users: u, tokens: t, hasher: h, jwt: j, now: now}
}

// Register создаёт пользователя и сразу выдаёт пару токенов.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if !in.Role.IsValid() {
		return nil, domain.ErrInvalidRole
	}

	// Проверяем, что email свободен.
	existing, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        in.Email,
		PasswordHash: hash,
		FullName:     in.FullName,
		Role:         in.Role,
		CreatedAt:    s.now().UTC(),
		UpdatedAt:    s.now().UTC(),
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	tokens, err := s.issueTokens(ctx, user, "", "")
	if err != nil {
		return nil, err
	}
	return &AuthResult{User: user, Tokens: tokens}, nil
}

// Login проверяет пароль и выдаёт пару токенов.
func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))

	user, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// Не раскрываем, что именно неверно — email или пароль.
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if err := s.hasher.Compare(user.PasswordHash, in.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	tokens, err := s.issueTokens(ctx, user, in.UserAgent, in.IPAddress)
	if err != nil {
		return nil, err
	}
	return &AuthResult{User: user, Tokens: tokens}, nil
}

// Refresh — выдача новой пары по refresh-токену с ротацией.
// Старый токен ревокается, чтобы исключить повторное использование.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	hash := s.jwt.HashRefresh(refreshToken)
	rec, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}
	if !rec.IsActive(s.now()) {
		return nil, domain.ErrInvalidToken
	}

	user, err := s.users.GetByID(ctx, rec.UserID)
	if err != nil {
		return nil, err
	}

	// Ротация: ревокаем старый, выдаём новый.
	if err := s.tokens.Revoke(ctx, rec.ID, s.now().UTC()); err != nil {
		return nil, err
	}
	tokens, err := s.issueTokens(ctx, user, rec.UserAgent, rec.IPAddress)
	if err != nil {
		return nil, err
	}
	return &AuthResult{User: user, Tokens: tokens}, nil
}

// Logout — ревокация всех refresh-токенов пользователя.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.tokens.RevokeAllForUser(ctx, userID, s.now().UTC())
}

// GetByID возвращает пользователя по идентификатору.
func (s *AuthService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}

// Update — частичное обновление профиля.
func (s *AuthService) Update(ctx context.Context, id uuid.UUID, in UpdateUserInput) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.FullName != nil {
		user.FullName = *in.FullName
	}
	if in.GroupID != nil {
		user.GroupID = in.GroupID
	}
	if in.Role != nil {
		if !in.Role.IsValid() {
			return nil, domain.ErrInvalidRole
		}
		user.Role = *in.Role
	}
	user.UpdatedAt = s.now().UTC()
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// issueTokens — общий путь выдачи пары access+refresh с записью refresh в БД.
func (s *AuthService) issueTokens(
	ctx context.Context, user *domain.User, ua, ip string,
) (TokenPair, error) {
	access, accessExp, err := s.jwt.GenerateAccess(user.ID, user.Role)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, refreshHash, refreshExp, err := s.jwt.GenerateRefresh()
	if err != nil {
		return TokenPair{}, err
	}
	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: refreshExp,
		CreatedAt: s.now().UTC(),
		UserAgent: ua,
		IPAddress: ip,
	}
	if err := s.tokens.Save(ctx, rt); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:      access,
		AccessExpiresAt:  accessExp,
		RefreshToken:     refresh,
		RefreshExpiresAt: refreshExp,
	}, nil
}
