package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role — глобальная роль из JWT (issued by auth-service).
// Дублируем enum здесь, чтобы не создавать межсервисную зависимость пакетов.
type Role string

const (
	RoleStudent     Role = "student"
	RoleGroupLeader Role = "group_leader"
	RoleTeacher     Role = "teacher"
	RoleAdmin       Role = "admin"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleStudent, RoleGroupLeader, RoleTeacher, RoleAdmin:
		return true
	}
	return false
}

// MembershipRole — роль пользователя ВНУТРИ конкретной группы.
// Не путать с глобальной Role: в одной группе можно быть leader,
// в другой — обычным member.
type MembershipRole string

const (
	MembershipMember MembershipRole = "member"
	MembershipLeader MembershipRole = "leader"
)

func (m MembershipRole) IsValid() bool {
	switch m {
	case MembershipMember, MembershipLeader:
		return true
	}
	return false
}

// Group — студенческая группа.
type Group struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Course    int       `json:"course"`
	Faculty   string    `json:"faculty"`
	LeaderID  uuid.UUID `json:"leader_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Membership — связь пользователя и группы с ролью внутри группы.
type Membership struct {
	UserID      uuid.UUID      `json:"user_id"`
	GroupID     uuid.UUID      `json:"group_id"`
	RoleInGroup MembershipRole `json:"role_in_group"`
	JoinedAt    time.Time      `json:"joined_at"`
}
