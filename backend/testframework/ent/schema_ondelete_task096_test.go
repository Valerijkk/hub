//go:build hubtest

package ent_test

// TASK-096 / ADR-ENT-ONDELETE-CASCADE-2026 — OnDelete schema guard + cascade evidence.
import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/poll"
	"github.com/valerii/hub/backend/ent/polloption"
	"github.com/valerii/hub/backend/ent/pollvote"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func task096SchemaDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "ent", "schema")
}

// TestSchema_TASK096_OnDeleteAnnotations — A2W8-1-003 / 4-002 / 4-006 markers.
func TestSchema_TASK096_OnDeleteAnnotations(t *testing.T) {
	dir := task096SchemaDir(t)
	mustContain := map[string][]string{
		"channel.go": {
			`entsql.OnDelete(entsql.Cascade)`,
			`edge.To("members"`,
			`edge.To("polls"`,
		},
		"call.go": {
			`entsql.OnDelete(entsql.Cascade)`,
			`edge.To("participants"`,
		},
		"poll.go": {
			`entsql.OnDelete(entsql.Cascade)`,
			`edge.To("options"`,
			`edge.To("votes"`,
		},
		"polloption.go": {
			`entsql.OnDelete(entsql.Cascade)`,
			`edge.To("votes"`,
		},
		"project.go": {
			`entsql.OnDelete(entsql.Cascade)`,
			`edge.To("issues"`,
			`edge.To("members"`,
		},
		"organization.go": {
			`entsql.OnDelete(entsql.NoAction)`,
			`edge.To("users"`,
			`edge.To("projects"`,
		},
	}
	for file, needles := range mustContain {
		src, err := os.ReadFile(filepath.Join(dir, file))
		require.NoError(t, err, file)
		body := string(src)
		for _, n := range needles {
			assert.Contains(t, body, n, file)
		}
	}
}

// TestSchema_TASK096_IssueLinkUniqueAndNoParentChild — A2W8-3-003.
func TestSchema_TASK096_IssueLinkUniqueAndNoParentChild(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(task096SchemaDir(t), "issuelink.go"))
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, `index.Edges("source", "target")`)
	assert.Contains(t, body, `Fields("link_type")`)
	assert.Contains(t, body, `Unique()`)
	assert.Contains(t, body, `"blocks"`)
	assert.Contains(t, body, `"relates_to"`)
	// Enum Values must not list parent/child (dual hierarchy closed).
	assert.NotContains(t, body, `"parent", "child"`)
	assert.NotContains(t, body, `"child", "parent"`)
}

// TestSchema_TASK096_MigrateSQLExists — SQL marker for live Postgres.
func TestSchema_TASK096_MigrateSQLExists(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	sqlPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "20260728_task096_ondelete_cascade.sql")
	src, err := os.ReadFile(sqlPath)
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, "needs-owner-apply")
	assert.Contains(t, body, "ADR-ENT-ONDELETE-CASCADE-2026")
	assert.Contains(t, body, "issuelink_link_type_issue_outgoing_links_issue_link_target")
	assert.Contains(t, body, "DELETE FROM issue_links WHERE link_type IN ('parent', 'child')")
}

// TestCascade_TASK096_PollHardDeleteRemovesOptionsVotes — A2W8-4-006 mutation evidence.
func TestCascade_TASK096_PollHardDeleteRemovesOptionsVotes(t *testing.T) {
	dsn := "file:task096_poll_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("p096-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		Save(ctx)
	require.NoError(t, err)

	ch, err := client.Channel.Create().
		SetChannelType(channel.ChannelTypePublic).
		SetName("poll-cascade").
		SetCreatedBy(u.ID).
		Save(ctx)
	require.NoError(t, err)

	p, err := client.Poll.Create().
		SetChannelID(ch.ID).
		SetCreatorID(u.ID).
		SetQuestion("Q?").
		Save(ctx)
	require.NoError(t, err)

	opt, err := client.PollOption.Create().
		SetPollID(p.ID).
		SetText("A").
		SetPosition(0).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.PollVote.Create().
		SetPollID(p.ID).
		SetOptionID(opt.ID).
		SetUserID(u.ID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.Poll.DeleteOneID(p.ID).Exec(ctx))

	nOpt, err := client.PollOption.Query().Where(polloption.HasPollWith(poll.IDEQ(p.ID))).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, nOpt, "options must CASCADE with poll")

	nVote, err := client.PollVote.Query().Where(pollvote.HasPollWith(poll.IDEQ(p.ID))).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, nVote, "votes must CASCADE with poll")
}

// TestCascade_TASK096_ChannelHardDeleteRemovesMembers — A2W8-1-003.
func TestCascade_TASK096_ChannelHardDeleteRemovesMembers(t *testing.T) {
	dsn := "file:task096_ch_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("c096-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		Save(ctx)
	require.NoError(t, err)

	ch, err := client.Channel.Create().
		SetChannelType(channel.ChannelTypePublic).
		SetName("hard-cascade").
		SetCreatedBy(u.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ChannelMember.Create().
		SetChannelID(ch.ID).
		SetUserID(u.ID).
		SetRole(channelmember.RoleOwner).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.Channel.DeleteOneID(ch.ID).Exec(ctx))

	n, err := client.ChannelMember.Query().Where(channelmember.ChannelID(ch.ID)).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "members must CASCADE on channel hard-delete")
}

// TestCascade_TASK096_ProjectPermanentDeleteWithMember — A2W8-4-002.
func TestCascade_TASK096_ProjectPermanentDeleteWithMember(t *testing.T) {
	dsn := "file:task096_proj_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("proj096-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		Save(ctx)
	require.NoError(t, err)

	p, err := client.Project.Create().
		SetName("Cascade").
		SetKey("C096").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetProjectID(p.ID).
		SetUserID(u.ID).
		SetProjectRole(projectmember.ProjectRoleOwner).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.Project.DeleteOneID(p.ID).Exec(ctx))

	_, err = client.Project.Get(ctx, p.ID)
	require.True(t, ent.IsNotFound(err))

	n, err := client.ProjectMember.Query().Where(projectmember.HasProjectWith(project.IDEQ(p.ID))).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

// TestCascade_TASK096_IssueLinkUniqueAllowsDifferentTypes — A2W8-3-003.
func TestCascade_TASK096_IssueLinkUniqueAllowsDifferentTypes(t *testing.T) {
	dsn := "file:task096_link_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("link096-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		Save(ctx)
	require.NoError(t, err)

	p, err := client.Project.Create().
		SetName("Links").
		SetKey("L096").
		Save(ctx)
	require.NoError(t, err)

	st, err := client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#000").
		SetProjectID(p.ID).
		Save(ctx)
	require.NoError(t, err)

	mkIssue := func(n int, key string) *ent.Issue {
		iss, err := client.Issue.Create().
			SetIssueNumber(n).
			SetIssueKey(key).
			SetTitle(key).
			SetIssueType(issue.IssueTypeTask).
			SetProjectID(p.ID).
			SetStatusID(st.ID).
			SetReporterID(u.ID).
			Save(ctx)
		require.NoError(t, err)
		return iss
	}
	a := mkIssue(1, "L096-1")
	b := mkIssue(2, "L096-2")

	_, err = client.IssueLink.Create().
		SetSourceID(a.ID).
		SetTargetID(b.ID).
		SetLinkType(issuelink.LinkTypeBlocks).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.IssueLink.Create().
		SetSourceID(a.ID).
		SetTargetID(b.ID).
		SetLinkType(issuelink.LinkTypeRelatesTo).
		Save(ctx)
	require.NoError(t, err, "blocks+relates_to on same pair must be allowed")

	_, err = client.IssueLink.Create().
		SetSourceID(a.ID).
		SetTargetID(b.ID).
		SetLinkType(issuelink.LinkTypeBlocks).
		Save(ctx)
	require.Error(t, err, "duplicate (source,target,type) must fail")

	// parent/child removed from enum
	_, err = client.IssueLink.Create().
		SetSourceID(a.ID).
		SetTargetID(b.ID).
		SetLinkType(issuelink.LinkType("parent")).
		Save(ctx)
	require.Error(t, err)
}
