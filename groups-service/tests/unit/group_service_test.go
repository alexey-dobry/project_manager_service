package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/student-pm/groups-service/internal/domain"
	"github.com/student-pm/groups-service/internal/usecase"
)

func newSUT() (*usecase.GroupService, *MockRepo) {
	r := NewMockRepo()
	return usecase.NewGroupService(r, time.Now), r
}

func actor(role domain.Role) usecase.Actor {
	return usecase.Actor{ID: uuid.New(), Role: role}
}

// ===== Create =====

func TestCreate_AllowedForTeacherAndAdmin(t *testing.T) {
	for _, role := range []domain.Role{domain.RoleTeacher, domain.RoleAdmin} {
		t.Run(string(role), func(t *testing.T) {
			svc, _ := newSUT()
			g, err := svc.Create(context.Background(), actor(role), usecase.CreateGroupInput{
				Name: "БПИ-211", Course: 2, Faculty: "ФИТ", LeaderID: uuid.New(),
			})
			require.NoError(t, err)
			require.Equal(t, "БПИ-211", g.Name)
		})
	}
}

func TestCreate_ForbiddenForStudentAndLeader(t *testing.T) {
	for _, role := range []domain.Role{domain.RoleStudent, domain.RoleGroupLeader} {
		t.Run(string(role), func(t *testing.T) {
			svc, _ := newSUT()
			_, err := svc.Create(context.Background(), actor(role), usecase.CreateGroupInput{
				Name: "x", Course: 1, Faculty: "y", LeaderID: uuid.New(),
			})
			require.ErrorIs(t, err, domain.ErrForbidden)
		})
	}
}

func TestCreate_LeaderBecomesMember(t *testing.T) {
	svc, repo := newSUT()
	leaderID := uuid.New()
	g, err := svc.Create(context.Background(), actor(domain.RoleAdmin), usecase.CreateGroupInput{
		Name: "БПИ-211", Course: 2, Faculty: "ФИТ", LeaderID: leaderID,
	})
	require.NoError(t, err)

	members, err := repo.ListMembers(context.Background(), g.ID)
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, leaderID, members[0].UserID)
	require.Equal(t, domain.MembershipLeader, members[0].RoleInGroup)
}

func TestCreate_DuplicateName(t *testing.T) {
	svc, _ := newSUT()
	in := usecase.CreateGroupInput{Name: "БПИ-211", Course: 1, Faculty: "ФИТ", LeaderID: uuid.New()}
	_, err := svc.Create(context.Background(), actor(domain.RoleAdmin), in)
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), actor(domain.RoleAdmin), in)
	require.ErrorIs(t, err, domain.ErrGroupAlreadyExists)
}

// ===== Update =====

func TestUpdate_LeaderCanEditOwnGroup(t *testing.T) {
	svc, _ := newSUT()
	leader := actor(domain.RoleStudent) // глобально не teacher/admin
	g, err := svc.Create(context.Background(), actor(domain.RoleAdmin), usecase.CreateGroupInput{
		Name: "БПИ-211", Course: 2, Faculty: "ФИТ", LeaderID: leader.ID,
	})
	require.NoError(t, err)

	newName := "БПИ-211 обновл."
	_, err = svc.Update(context.Background(), leader, g.ID, usecase.UpdateGroupInput{Name: &newName})
	require.NoError(t, err)
}

func TestUpdate_StudentCannotEditNotOwnGroup(t *testing.T) {
	svc, _ := newSUT()
	g, _ := svc.Create(context.Background(), actor(domain.RoleAdmin), usecase.CreateGroupInput{
		Name: "x", Course: 1, Faculty: "y", LeaderID: uuid.New(),
	})

	random := actor(domain.RoleStudent)
	newName := "hacked"
	_, err := svc.Update(context.Background(), random, g.ID, usecase.UpdateGroupInput{Name: &newName})
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdate_LeaderCannotChangeLeader(t *testing.T) {
	svc, _ := newSUT()
	leader := actor(domain.RoleStudent)
	g, _ := svc.Create(context.Background(), actor(domain.RoleAdmin), usecase.CreateGroupInput{
		Name: "x", Course: 1, Faculty: "y", LeaderID: leader.ID,
	})

	other := uuid.New()
	_, err := svc.Update(context.Background(), leader, g.ID, usecase.UpdateGroupInput{LeaderID: &other})
	require.ErrorIs(t, err, domain.ErrForbidden, "лидер не может менять лидера — только teacher/admin")
}

// ===== Delete =====

func TestDelete_OnlyTeacherAdmin(t *testing.T) {
	svc, _ := newSUT()
	g, _ := svc.Create(context.Background(), actor(domain.RoleAdmin), usecase.CreateGroupInput{
		Name: "x", Course: 1, Faculty: "y", LeaderID: uuid.New(),
	})

	require.ErrorIs(t,
		svc.Delete(context.Background(), actor(domain.RoleStudent), g.ID),
		domain.ErrForbidden,
	)
	require.NoError(t,
		svc.Delete(context.Background(), actor(domain.RoleTeacher), g.ID),
	)
}

// ===== Members =====

func TestAddMember_ByLeader(t *testing.T) {
	svc, _ := newSUT()
	leader := actor(domain.RoleStudent)
	g, _ := svc.Create(context.Background(), actor(domain.RoleAdmin), usecase.CreateGroupInput{
		Name: "x", Course: 1, Faculty: "y", LeaderID: leader.ID,
	})

	newMember := uuid.New()
	m, err := svc.AddMember(context.Background(), leader, g.ID, usecase.AddMemberInput{
		UserID: newMember, Role: domain.MembershipMember,
	})
	require.NoError(t, err)
	require.Equal(t, newMember, m.UserID)
}

func TestAddMember_ByRandomStudent_Forbidden(t *testing.T) {
	svc, _ := newSUT()
	g, _ := svc.Create(context.Background(), actor(domain.RoleAdmin), usecase.CreateGroupInput{
		Name: "x", Course: 1, Faculty: "y", LeaderID: uuid.New(),
	})

	_, err := svc.AddMember(context.Background(), actor(domain.RoleStudent), g.ID, usecase.AddMemberInput{
		UserID: uuid.New(), Role: domain.MembershipMember,
	})
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAddMember_Duplicate(t *testing.T) {
	svc, _ := newSUT()
	g, _ := svc.Create(context.Background(), actor(domain.RoleAdmin), usecase.CreateGroupInput{
		Name: "x", Course: 1, Faculty: "y", LeaderID: uuid.New(),
	})
	uid := uuid.New()
	_, err := svc.AddMember(context.Background(), actor(domain.RoleAdmin), g.ID, usecase.AddMemberInput{
		UserID: uid, Role: domain.MembershipMember,
	})
	require.NoError(t, err)
	_, err = svc.AddMember(context.Background(), actor(domain.RoleAdmin), g.ID, usecase.AddMemberInput{
		UserID: uid, Role: domain.MembershipMember,
	})
	require.ErrorIs(t, err, domain.ErrMemberAlreadyInGroup)
}

func TestRemoveMember_SelfLeave(t *testing.T) {
	svc, _ := newSUT()
	g, _ := svc.Create(context.Background(), actor(domain.RoleAdmin), usecase.CreateGroupInput{
		Name: "x", Course: 1, Faculty: "y", LeaderID: uuid.New(),
	})
	user := actor(domain.RoleStudent)
	_, _ = svc.AddMember(context.Background(), actor(domain.RoleAdmin), g.ID, usecase.AddMemberInput{
		UserID: user.ID, Role: domain.MembershipMember,
	})
	require.NoError(t, svc.RemoveMember(context.Background(), user, g.ID, user.ID))
}

func TestRemoveMember_CannotKickLeader(t *testing.T) {
	svc, _ := newSUT()
	leaderID := uuid.New()
	g, _ := svc.Create(context.Background(), actor(domain.RoleAdmin), usecase.CreateGroupInput{
		Name: "x", Course: 1, Faculty: "y", LeaderID: leaderID,
	})
	require.ErrorIs(t,
		svc.RemoveMember(context.Background(), actor(domain.RoleAdmin), g.ID, leaderID),
		domain.ErrForbidden, "сначала смените лидера, потом удаляйте",
	)
}

func TestList_PaginationDefaults(t *testing.T) {
	svc, _ := newSUT()
	for i := 0; i < 5; i++ {
		_, err := svc.Create(context.Background(), actor(domain.RoleAdmin), usecase.CreateGroupInput{
			Name: "g" + string(rune('a'+i)), Course: 1, Faculty: "ФИТ", LeaderID: uuid.New(),
		})
		require.NoError(t, err)
	}
	groups, total, err := svc.List(context.Background(), usecase.ListFilter{})
	require.NoError(t, err)
	require.Equal(t, 5, total)
	require.Len(t, groups, 5)
}
