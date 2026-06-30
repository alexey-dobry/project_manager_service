package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/student-pm/groups-service/internal/domain"
)

// GroupService реализует сценарии работы с группами.
type GroupService struct {
	repo GroupRepository
	now  func() time.Time
}

func NewGroupService(r GroupRepository, now func() time.Time) *GroupService {
	if now == nil {
		now = time.Now
	}
	return &GroupService{repo: r, now: now}
}

// ===== queries =====

func (s *GroupService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Group, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *GroupService) List(ctx context.Context, f ListFilter) ([]domain.Group, int, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return s.repo.List(ctx, f)
}

func (s *GroupService) ListMembers(ctx context.Context, groupID uuid.UUID) ([]domain.Membership, error) {
	// Убедимся, что группа существует — иначе вернём 404, а не пустой список.
	if _, err := s.repo.GetByID(ctx, groupID); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, groupID)
}

// ===== commands =====

// Create: только teacher/admin могут заводить группы.
func (s *GroupService) Create(ctx context.Context, actor Actor, in CreateGroupInput) (*domain.Group, error) {
	if !canManageGroups(actor.Role) {
		return nil, domain.ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, domain.ErrInvalidName
	}
	if !domain.IsValidCourse(in.Course) {
		return nil, domain.ErrInvalidCourse
	}

	g := &domain.Group{
		ID:        uuid.New(),
		Name:      in.Name,
		Course:    in.Course,
		Faculty:   strings.TrimSpace(in.Faculty),
		LeaderID:  in.LeaderID,
		CreatedAt: s.now().UTC(),
		UpdatedAt: s.now().UTC(),
	}
	if err := s.repo.Create(ctx, g); err != nil {
		return nil, err
	}

	// Лидер автоматически становится участником группы с ролью leader.
	// Операция не атомарна: при сбое AddMember группа уже создана.
	leaderMembership := &domain.Membership{
		UserID:      in.LeaderID,
		GroupID:     g.ID,
		RoleInGroup: domain.MembershipLeader,
		JoinedAt:    s.now().UTC(),
	}
	if err := s.repo.AddMember(ctx, leaderMembership); err != nil &&
		!errors.Is(err, domain.ErrMemberAlreadyInGroup) {
		return nil, err
	}
	return g, nil
}

// Update: teacher/admin или текущий лидер группы.
func (s *GroupService) Update(ctx context.Context, actor Actor, id uuid.UUID, in UpdateGroupInput) (*domain.Group, error) {
	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canManageGroups(actor.Role) && g.LeaderID != actor.ID {
		return nil, domain.ErrForbidden
	}

	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, domain.ErrInvalidName
		}
		g.Name = name
	}
	if in.Course != nil {
		if !domain.IsValidCourse(*in.Course) {
			return nil, domain.ErrInvalidCourse
		}
		g.Course = *in.Course
	}
	if in.Faculty != nil {
		g.Faculty = strings.TrimSpace(*in.Faculty)
	}
	if in.LeaderID != nil {
		// Менять лидера может только teacher/admin.
		if !canManageGroups(actor.Role) {
			return nil, domain.ErrForbidden
		}
		g.LeaderID = *in.LeaderID
	}
	g.UpdatedAt = s.now().UTC()

	if err := s.repo.Update(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

// Delete: только teacher/admin.
func (s *GroupService) Delete(ctx context.Context, actor Actor, id uuid.UUID) error {
	if !canManageGroups(actor.Role) {
		return domain.ErrForbidden
	}
	return s.repo.Delete(ctx, id)
}

// AddMember: лидер группы или teacher/admin.
func (s *GroupService) AddMember(ctx context.Context, actor Actor, groupID uuid.UUID, in AddMemberInput) (*domain.Membership, error) {
	g, err := s.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if !canManageMembers(actor, g) {
		return nil, domain.ErrForbidden
	}
	if !in.Role.IsValid() {
		return nil, domain.ErrInvalidMembership
	}

	m := &domain.Membership{
		UserID:      in.UserID,
		GroupID:     groupID,
		RoleInGroup: in.Role,
		JoinedAt:    s.now().UTC(),
	}
	if err := s.repo.AddMember(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// RemoveMember: лидер группы или teacher/admin.
// Пользователь может также сам выйти из группы — поэтому отдельная ветка.
func (s *GroupService) RemoveMember(ctx context.Context, actor Actor, groupID, userID uuid.UUID) error {
	g, err := s.repo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	// Запрещаем удалять текущего лидера группы — сначала смените лидера.
	if g.LeaderID == userID {
		return domain.ErrForbidden
	}
	// Сам себя — можно. Иначе — только менеджер группы.
	if actor.ID != userID && !canManageMembers(actor, g) {
		return domain.ErrForbidden
	}
	return s.repo.RemoveMember(ctx, groupID, userID)
}

// ===== RBAC helpers =====

// canManageGroups — может создавать/удалять группы и менять лидера.
func canManageGroups(r domain.Role) bool {
	return r == domain.RoleTeacher || r == domain.RoleAdmin
}

// canManageMembers — может добавлять/удалять участников группы.
// Это либо глобальный teacher/admin, либо текущий лидер этой конкретной группы.
func canManageMembers(a Actor, g *domain.Group) bool {
	if canManageGroups(a.Role) {
		return true
	}
	return g.LeaderID == a.ID
}
