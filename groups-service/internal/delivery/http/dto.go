package httpdelivery

import "time"

// CreateGroupRequest — создание группы.
type CreateGroupRequest struct {
	Name     string `json:"name"      validate:"required,min=2,max=120"      example:"БПИ-211"`
	Course   int    `json:"course"    validate:"required,gt=0,max=10"        example:"2"`
	Faculty  string `json:"faculty"   validate:"required,min=2,max=120"      example:"ФИТ"`
	LeaderID string `json:"leader_id" validate:"required,uuid"               example:"7c0a..."`
}

// UpdateGroupRequest — частичное обновление.
type UpdateGroupRequest struct {
	Name     *string `json:"name,omitempty"      validate:"omitempty,min=2,max=120"`
	Course   *int    `json:"course,omitempty"    validate:"omitempty,gt=0,max=10"`
	Faculty  *string `json:"faculty,omitempty"   validate:"omitempty,min=2,max=120"`
	LeaderID *string `json:"leader_id,omitempty" validate:"omitempty,uuid"`
}

// AddMemberRequest — добавление участника.
type AddMemberRequest struct {
	UserID      string `json:"user_id"       validate:"required,uuid"`
	RoleInGroup string `json:"role_in_group" validate:"required,oneof=member leader" example:"member"`
}

// GroupResponse — публичная проекция группы.
type GroupResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Course    int       `json:"course"`
	Faculty   string    `json:"faculty"`
	LeaderID  string    `json:"leader_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MembershipResponse — проекция membership.
type MembershipResponse struct {
	UserID      string    `json:"user_id"`
	GroupID     string    `json:"group_id"`
	RoleInGroup string    `json:"role_in_group"`
	JoinedAt    time.Time `json:"joined_at"`
}

// PaginatedGroupsResponse — список групп с пагинацией.
type PaginatedGroupsResponse struct {
	Items  []GroupResponse `json:"items"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// MessageResponse — простой ответ.
type MessageResponse struct {
	Message string `json:"message" example:"ok"`
}
