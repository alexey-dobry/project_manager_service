package unit

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/student-pm/projects-service/internal/domain"
)

func TestProjectTransitions(t *testing.T) {
	type tc struct {
		from, to domain.ProjectStatus
		ok       bool
	}
	cases := []tc{
		{domain.ProjectDraft, domain.ProjectInProgress, true},
		{domain.ProjectDraft, domain.ProjectCompleted, false},
		{domain.ProjectInProgress, domain.ProjectReview, true},
		{domain.ProjectReview, domain.ProjectCompleted, true},
		{domain.ProjectReview, domain.ProjectInProgress, true}, // doc'd: revisions
		{domain.ProjectCompleted, domain.ProjectInProgress, false},
		{domain.ProjectCompleted, domain.ProjectArchived, true},
		{domain.ProjectArchived, domain.ProjectDraft, true}, // расконсервация
		// идемпотентность
		{domain.ProjectDraft, domain.ProjectDraft, true},
	}
	for _, c := range cases {
		got := c.from.CanTransitionTo(c.to)
		require.Equal(t, c.ok, got, "%s -> %s expected ok=%v", c.from, c.to, c.ok)
	}
}

func TestTaskTransitions(t *testing.T) {
	type tc struct {
		from, to domain.TaskStatus
		ok       bool
	}
	cases := []tc{
		{domain.TaskTodo, domain.TaskInProgress, true},
		{domain.TaskTodo, domain.TaskBlocked, true},
		{domain.TaskTodo, domain.TaskDone, false}, // нельзя сразу в done
		{domain.TaskInProgress, domain.TaskDone, true},
		{domain.TaskInProgress, domain.TaskBlocked, true},
		{domain.TaskBlocked, domain.TaskInProgress, true},
		{domain.TaskBlocked, domain.TaskDone, false}, // только через in_progress
		{domain.TaskDone, domain.TaskInProgress, true}, // переоткрытие
		{domain.TaskDone, domain.TaskTodo, false},
	}
	for _, c := range cases {
		require.Equal(t, c.ok, c.from.CanTransitionTo(c.to),
			"%s -> %s", c.from, c.to)
	}
}
