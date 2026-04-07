package domain

import "errors"

var (
	ErrProjectNotFound    = errors.New("project not found")
	ErrTaskNotFound       = errors.New("task not found")
	ErrCommentNotFound    = errors.New("comment not found")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrInvalidPriority    = errors.New("invalid priority")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrForbidden          = errors.New("forbidden")
	ErrTaskNotInProject   = errors.New("task does not belong to this project")
	ErrCommentNotInTask   = errors.New("comment does not belong to this task")
	ErrEmptyContent       = errors.New("comment content cannot be empty")
)
