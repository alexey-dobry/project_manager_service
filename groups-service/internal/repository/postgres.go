package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/student-pm/groups-service/internal/domain"
	"github.com/student-pm/groups-service/internal/usecase"
)

const pgUniqueViolation = "23505"

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

// ===== groups =====

func (r *PostgresRepo) Create(ctx context.Context, g *domain.Group) error {
	const q = `
		INSERT INTO groups (id, name, course, faculty, leader_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, q, g.ID, g.Name, g.Course, g.Faculty, g.LeaderID, g.CreatedAt, g.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ErrGroupAlreadyExists
		}
		return err
	}
	return nil
}

func (r *PostgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Group, error) {
	const q = `
		SELECT id, name, course, faculty, leader_id, created_at, updated_at
		FROM groups WHERE id = $1
	`
	g := &domain.Group{}
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&g.ID, &g.Name, &g.Course, &g.Faculty, &g.LeaderID, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, err
	}
	return g, nil
}

// List — пагинированный список с фильтрами. Возвращает (rows, total, err).
func (r *PostgresRepo) List(ctx context.Context, f usecase.ListFilter) ([]domain.Group, int, error) {
	conds := []string{"1=1"}
	args := []any{}
	if f.Faculty != "" {
		args = append(args, f.Faculty)
		conds = append(conds, "faculty = $"+strconv.Itoa(len(args)))
	}
	if f.Course > 0 {
		args = append(args, f.Course)
		conds = append(conds, "course = $"+strconv.Itoa(len(args)))
	}
	where := strings.Join(conds, " AND ")

	// total
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM groups WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	q := "SELECT id, name, course, faculty, leader_id, created_at, updated_at FROM groups WHERE " +
		where +
		" ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)-1) + " OFFSET $" + strconv.Itoa(len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]domain.Group, 0)
	for rows.Next() {
		var g domain.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Course, &g.Faculty, &g.LeaderID, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, g)
	}
	return out, total, rows.Err()
}

func (r *PostgresRepo) Update(ctx context.Context, g *domain.Group) error {
	const q = `
		UPDATE groups
		SET name = $1, course = $2, faculty = $3, leader_id = $4, updated_at = $5
		WHERE id = $6
	`
	tag, err := r.pool.Exec(ctx, q, g.Name, g.Course, g.Faculty, g.LeaderID, g.UpdatedAt, g.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ErrGroupAlreadyExists
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrGroupNotFound
	}
	return nil
}

func (r *PostgresRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM groups WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrGroupNotFound
	}
	return nil
}

// ===== memberships =====

func (r *PostgresRepo) AddMember(ctx context.Context, m *domain.Membership) error {
	const q = `
		INSERT INTO group_memberships (group_id, user_id, role_in_group, joined_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.pool.Exec(ctx, q, m.GroupID, m.UserID, m.RoleInGroup, m.JoinedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ErrMemberAlreadyInGroup
		}
		return err
	}
	return nil
}

func (r *PostgresRepo) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	const q = `DELETE FROM group_memberships WHERE group_id = $1 AND user_id = $2`
	tag, err := r.pool.Exec(ctx, q, groupID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrMemberNotFound
	}
	return nil
}

func (r *PostgresRepo) ListMembers(ctx context.Context, groupID uuid.UUID) ([]domain.Membership, error) {
	const q = `
		SELECT user_id, group_id, role_in_group, joined_at
		FROM group_memberships
		WHERE group_id = $1
		ORDER BY joined_at ASC
	`
	rows, err := r.pool.Query(ctx, q, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Membership, 0)
	for rows.Next() {
		var m domain.Membership
		if err := rows.Scan(&m.UserID, &m.GroupID, &m.RoleInGroup, &m.JoinedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *PostgresRepo) GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.Membership, error) {
	const q = `
		SELECT user_id, group_id, role_in_group, joined_at
		FROM group_memberships WHERE group_id = $1 AND user_id = $2
	`
	m := &domain.Membership{}
	err := r.pool.QueryRow(ctx, q, groupID, userID).Scan(&m.UserID, &m.GroupID, &m.RoleInGroup, &m.JoinedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMemberNotFound
		}
		return nil, err
	}
	return m, nil
}
