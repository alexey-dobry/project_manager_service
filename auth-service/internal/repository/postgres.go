package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/student-pm/auth-service/internal/domain"
)

// pgUniqueViolation — SQLSTATE для нарушения unique-ограничения.
const pgUniqueViolation = "23505"

// PostgresRepo реализует UserRepository и RefreshTokenRepository.
type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

// ===== USERS =====

func (r *PostgresRepo) Create(ctx context.Context, u *domain.User) error {
	const q = `
		INSERT INTO users (id, email, password_hash, full_name, role, group_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, q,
		u.ID, u.Email, u.PasswordHash, u.FullName, u.Role, u.GroupID, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func (r *PostgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `
		SELECT id, email, password_hash, full_name, role, group_id, created_at, updated_at
		FROM users WHERE id = $1
	`
	return r.scanUser(r.pool.QueryRow(ctx, q, id))
}

func (r *PostgresRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, email, password_hash, full_name, role, group_id, created_at, updated_at
		FROM users WHERE email = $1
	`
	return r.scanUser(r.pool.QueryRow(ctx, q, email))
}

func (r *PostgresRepo) Update(ctx context.Context, u *domain.User) error {
	const q = `
		UPDATE users
		SET full_name = $1, role = $2, group_id = $3, updated_at = $4
		WHERE id = $5
	`
	tag, err := r.pool.Exec(ctx, q, u.FullName, u.Role, u.GroupID, u.UpdatedAt, u.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *PostgresRepo) scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.GroupID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// ===== REFRESH TOKENS =====

func (r *PostgresRepo) Save(ctx context.Context, t *domain.RefreshToken) error {
	const q = `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked_at, created_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, q,
		t.ID, t.UserID, t.TokenHash, t.ExpiresAt, t.RevokedAt, t.CreatedAt, t.UserAgent, t.IPAddress,
	)
	return err
}

func (r *PostgresRepo) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at, user_agent, ip_address
		FROM refresh_tokens WHERE token_hash = $1
	`
	t := &domain.RefreshToken{}
	err := r.pool.QueryRow(ctx, q, hash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt, &t.UserAgent, &t.IPAddress,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}
	return t, nil
}

func (r *PostgresRepo) Revoke(ctx context.Context, id uuid.UUID, at time.Time) error {
	const q = `UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2 AND revoked_at IS NULL`
	_, err := r.pool.Exec(ctx, q, at, id)
	return err
}

func (r *PostgresRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID, at time.Time) error {
	const q = `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	_, err := r.pool.Exec(ctx, q, at, userID)
	return err
}
