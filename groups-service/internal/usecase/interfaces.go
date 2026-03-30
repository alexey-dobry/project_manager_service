package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/student-pm/groups-service/internal/domain"
)

// GroupRepository — порт доступа к группам и членству.
// Обе сущности живут в одном bounded context, поэтому один интерфейс.
type GroupRepository interface {
	Create(ctx context.Context, g *domain.Group) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Group, error)
	List(ctx context.Context, f ListFilter) ([]domain.Group, int, error)
	Update(ctx context.Context, g *domain.Group) error
	Delete(ctx context.Context, id uuid.UUID) error

	AddMember(ctx context.Context, m *domain.Membership) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]domain.Membership, error)
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.Membership, error)
}

// ListFilter — пагинация и фильтр для списка групп.
type ListFilter struct {
	Faculty string // точное совпадение, пусто = не фильтровать
	Course  int    // 0 = не фильтровать
	Limit   int    // <=0 → 20
	Offset  int
}
