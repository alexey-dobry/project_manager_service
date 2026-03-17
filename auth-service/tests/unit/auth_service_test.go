package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/student-pm/auth-service/internal/domain"
	"github.com/student-pm/auth-service/internal/pkg/hasher"
	jwtpkg "github.com/student-pm/auth-service/internal/pkg/jwt"
	"github.com/student-pm/auth-service/internal/usecase"
)

// newSUT — фабрика system-under-test с реальным хешером (мин. cost) и JWT.
func newSUT(t *testing.T) (*usecase.AuthService, *MockUserRepo, *MockTokenRepo) {
	t.Helper()
	users := NewMockUserRepo()
	tokens := NewMockTokenRepo()
	h := hasher.New(4) // bcrypt MinCost — быстрее в тестах
	tp, err := jwtpkg.New(jwtpkg.Config{
		Secret:     "test-secret-test-secret-1234567890",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
		Issuer:     "test",
	})
	require.NoError(t, err)
	return usecase.NewAuthService(users, tokens, h, tp, time.Now), users, tokens
}

func TestRegister_Success(t *testing.T) {
	svc, users, tokens := newSUT(t)
	ctx := context.Background()

	res, err := svc.Register(ctx, usecase.RegisterInput{
		Email:    "Ivan@uni.edu",
		Password: "secretPassword1",
		FullName: "Ivan Ivanov",
		Role:     domain.RoleStudent,
	})
	require.NoError(t, err)
	require.NotNil(t, res.User)
	require.Equal(t, "ivan@uni.edu", res.User.Email, "email must be lowered")
	require.NotEmpty(t, res.Tokens.AccessToken)
	require.NotEmpty(t, res.Tokens.RefreshToken)
	require.Equal(t, 1, tokens.Count())

	// Юзер реально лежит в репо.
	got, err := users.GetByEmail(ctx, "ivan@uni.edu")
	require.NoError(t, err)
	require.Equal(t, "Ivan Ivanov", got.FullName)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _, _ := newSUT(t)
	ctx := context.Background()
	in := usecase.RegisterInput{
		Email: "dup@uni.edu", Password: "passpass", FullName: "X Y", Role: domain.RoleStudent,
	}
	_, err := svc.Register(ctx, in)
	require.NoError(t, err)
	_, err = svc.Register(ctx, in)
	require.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}

func TestRegister_InvalidRole(t *testing.T) {
	svc, _, _ := newSUT(t)
	_, err := svc.Register(context.Background(), usecase.RegisterInput{
		Email: "x@y.z", Password: "passpass", FullName: "x", Role: "hacker",
	})
	require.ErrorIs(t, err, domain.ErrInvalidRole)
}

func TestLogin_Success(t *testing.T) {
	svc, _, _ := newSUT(t)
	ctx := context.Background()
	_, err := svc.Register(ctx, usecase.RegisterInput{
		Email: "a@a.a", Password: "rightPass", FullName: "A", Role: domain.RoleStudent,
	})
	require.NoError(t, err)

	res, err := svc.Login(ctx, usecase.LoginInput{Email: "a@a.a", Password: "rightPass"})
	require.NoError(t, err)
	require.NotEmpty(t, res.Tokens.AccessToken)
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, _, _ := newSUT(t)
	ctx := context.Background()
	_, _ = svc.Register(ctx, usecase.RegisterInput{
		Email: "b@b.b", Password: "rightPass", FullName: "B", Role: domain.RoleStudent,
	})
	_, err := svc.Login(ctx, usecase.LoginInput{Email: "b@b.b", Password: "wrong"})
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestLogin_UnknownEmail_LooksLikeWrongPassword(t *testing.T) {
	svc, _, _ := newSUT(t)
	_, err := svc.Login(context.Background(), usecase.LoginInput{Email: "no@one.x", Password: "x"})
	require.ErrorIs(t, err, domain.ErrInvalidCredentials,
		"должно маскировать отсутствие пользователя как InvalidCredentials")
}

func TestRefresh_RotatesAndRevokesOld(t *testing.T) {
	svc, _, tokens := newSUT(t)
	ctx := context.Background()
	res, err := svc.Register(ctx, usecase.RegisterInput{
		Email: "c@c.c", Password: "passpass", FullName: "C", Role: domain.RoleStudent,
	})
	require.NoError(t, err)

	old := res.Tokens.RefreshToken
	res2, err := svc.Refresh(ctx, old)
	require.NoError(t, err)
	require.NotEqual(t, old, res2.Tokens.RefreshToken, "refresh должен ротироваться")

	// Старый теперь невалиден.
	_, err = svc.Refresh(ctx, old)
	require.ErrorIs(t, err, domain.ErrInvalidToken)

	require.Equal(t, 2, tokens.Count(), "оба токена существуют, старый — revoked")
}

func TestRefresh_InvalidToken(t *testing.T) {
	svc, _, _ := newSUT(t)
	_, err := svc.Refresh(context.Background(), "garbage-token")
	require.ErrorIs(t, err, domain.ErrInvalidToken)
}

func TestLogout_RevokesAll(t *testing.T) {
	svc, _, _ := newSUT(t)
	ctx := context.Background()
	res, err := svc.Register(ctx, usecase.RegisterInput{
		Email: "d@d.d", Password: "passpass", FullName: "D", Role: domain.RoleStudent,
	})
	require.NoError(t, err)
	require.NoError(t, svc.Logout(ctx, res.User.ID))

	_, err = svc.Refresh(ctx, res.Tokens.RefreshToken)
	require.ErrorIs(t, err, domain.ErrInvalidToken)
}

func TestUpdate_FullNameAndRole(t *testing.T) {
	svc, _, _ := newSUT(t)
	ctx := context.Background()
	res, err := svc.Register(ctx, usecase.RegisterInput{
		Email: "e@e.e", Password: "passpass", FullName: "Old", Role: domain.RoleStudent,
	})
	require.NoError(t, err)

	newName := "New Name"
	newRole := domain.RoleGroupLeader
	updated, err := svc.Update(ctx, res.User.ID, usecase.UpdateUserInput{
		FullName: &newName,
		Role:     &newRole,
	})
	require.NoError(t, err)
	require.Equal(t, "New Name", updated.FullName)
	require.Equal(t, domain.RoleGroupLeader, updated.Role)
}
