package domain

import "errors"

var (
	ErrGroupNotFound       = errors.New("group not found")
	ErrGroupAlreadyExists  = errors.New("group with this name already exists")
	ErrMemberAlreadyInGroup = errors.New("user is already a member of this group")
	ErrMemberNotFound      = errors.New("user is not a member of this group")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidRole         = errors.New("invalid role")
	ErrInvalidMembership   = errors.New("invalid membership role")
)
