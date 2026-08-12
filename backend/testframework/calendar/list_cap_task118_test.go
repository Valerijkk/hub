package calendar_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/calendar"
)

// TestListScheduledCalls_CapAndWithProject_W7A2010 — Limit + Project edge.
// Mutation: снять Limit → len > MaxCalendarList; снять WithProject → Project nil.
func TestListScheduledCalls_CapAndWithProject_W7A2010(t *testing.T) {
	dsn := "file:cal_cap_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	repo := calendar.NewEntRepository(client)

	owner, err := client.User.Create().
		SetEmail("cal-cap-" + uuid.New().String()[:8] + "@t.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().SetName("CalCap").SetKey("CCP").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).SetProjectID(proj.ID).SetProjectRole("developer").Save(ctx)
	require.NoError(t, err)
	ch, err := client.Channel.Create().
		SetChannelType(channel.ChannelTypePublic).
		SetName("cal-cap-ch").
		SetCreatedBy(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	n := calendar.MaxCalendarList + 3
	for i := 0; i < n; i++ {
		_, err = client.ScheduledCall.Create().
			SetTitle(fmt.Sprintf("Call-%d", i)).
			SetScheduledAt(time.Now().UTC().Add(time.Duration(i) * time.Hour)).
			SetChannelID(ch.ID).
			SetProjectID(proj.ID).
			SetCreatedBy(owner.ID).
			SetIsActive(true).
			Save(ctx)
		require.NoError(t, err)
	}

	got, err := repo.ListScheduledCalls(ctx, []uuid.UUID{proj.ID})
	require.NoError(t, err)
	assert.Equal(t, calendar.MaxCalendarList, len(got))
	require.NotNil(t, got[0].Edges.Project, "W7-A2-010: WithProject must load ProjectKey edge")
	assert.Equal(t, "CCP", got[0].Edges.Project.Key)
}

// TestListIssuesDueInRange_Cap_W7A2010 — Limit на дедлайны.
func TestListIssuesDueInRange_Cap_W7A2010(t *testing.T) {
	dsn := "file:cal_iss_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	repo := calendar.NewEntRepository(client)

	proj, err := client.Project.Create().SetName("IssCap").SetKey("ICP").Save(ctx)
	require.NoError(t, err)
	status, err := client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory("todo").
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	n := calendar.MaxCalendarList + 2
	for i := 0; i < n; i++ {
		due := from.AddDate(0, 0, 1+(i%20))
		_, err = client.Issue.Create().
			SetTitle(fmt.Sprintf("I-%d", i)).
			SetIssueKey(fmt.Sprintf("ICP-%d", i+1)).
			SetIssueNumber(i + 1).
			SetIssueType(issue.IssueTypeTask).
			SetProjectID(proj.ID).
			SetStatusID(status.ID).
			SetDueDate(due).
			Save(ctx)
		require.NoError(t, err)
	}

	got, err := repo.ListIssuesDueInRange(ctx, []uuid.UUID{proj.ID}, from, to)
	require.NoError(t, err)
	assert.Equal(t, calendar.MaxCalendarList, len(got))
	require.NotNil(t, got[0].Edges.Project)
}
