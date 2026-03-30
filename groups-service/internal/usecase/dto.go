package usecase

import (
	"github.com/google/uuid"

	"github.com/student-pm/groups-service/internal/domain"
)

// Actor — тот, кто инициировал действие. Заполняется handler-ом
// из JWT-claims и пробрасывается в usecase для RBAC-проверок.
type Actor struct {
	ID   uuid.UUID
	Role domain.Role
}

type CreateGroupInput struct {
	Name     string
	Course   int
	Faculty  string
	LeaderID uuid.UUID
}

type UpdateGroupInput struct {
	Name     *string
	Course   *int
	Faculty  *string
	LeaderID *uuid.UUID
}

type AddMemberInput struct {
	UserID uuid.UUID
	Role   domain.MembershipRole
}
