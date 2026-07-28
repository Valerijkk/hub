//go:build hubtest

package ent_test

// TASK-068 / P3-BE-17 + P3-BE-18: Ent MaxLen/mentions hygiene + mojibake comments.
import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func schemaDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "ent", "schema")
}

// TestSchema_P3BE18_NoMojibake — grep-инварианты: нет «ибра» / битой «Р».
func TestSchema_P3BE18_NoMojibake(t *testing.T) {
	dir := schemaDir(t)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	brokenIbra := []byte("ибра")
	// Классический mojibake «И» → U+0420 + U+0098
	brokenI := []byte{0xd0, 0xa0, 0xc2, 0x98}

	var hits []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		require.NoError(t, err)
		if strings.Contains(string(data), string(brokenIbra)) {
			hits = append(hits, e.Name()+":ибра")
		}
		if strings.Contains(string(data), string(brokenI)) {
			hits = append(hits, e.Name()+":Р")
		}
	}
	assert.Empty(t, hits, "mojibake residual in ent/schema")
}

// TestSchema_P3BE17_HygieneMarkers — MaxLen / mentions UUID / один Unique source.
func TestSchema_P3BE17_HygieneMarkers(t *testing.T) {
	dir := schemaDir(t)

	inv, err := os.ReadFile(filepath.Join(dir, "invite.go"))
	require.NoError(t, err)
	assert.Contains(t, string(inv), `field.String("code")`)
	assert.Contains(t, string(inv), `MaxLen(64)`)

	news, err := os.ReadFile(filepath.Join(dir, "newspost.go"))
	require.NoError(t, err)
	assert.Contains(t, string(news), `field.Text("body")`)
	assert.Contains(t, string(news), `MaxLen(50000)`)

	cmt, err := os.ReadFile(filepath.Join(dir, "comment.go"))
	require.NoError(t, err)
	assert.Contains(t, string(cmt), `field.JSON("mentions", []uuid.UUID{})`)

	api, err := os.ReadFile(filepath.Join(dir, "apikey.go"))
	require.NoError(t, err)
	assert.NotContains(t, string(api), `index.Fields("key_hash")`)

	call, err := os.ReadFile(filepath.Join(dir, "call.go"))
	require.NoError(t, err)
	assert.NotContains(t, string(call), `index.Fields("room_name")`)
	assert.NotContains(t, string(call), `index.Fields("invite_code")`)

	chi, err := os.ReadFile(filepath.Join(dir, "channelinvite.go"))
	require.NoError(t, err)
	assert.NotContains(t, string(chi), `index.Fields("token").Unique()`)
}

// TestInvite_CodeMaxLen_P3BE17 — oversize code → Ent validation (без SQL migrate).
func TestInvite_CodeMaxLen_P3BE17(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:invite_maxlen_p3be17_"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("invite-maxlen@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Invite MaxLen").
		SetKey("IML1").
		Save(ctx)
	require.NoError(t, err)

	ok, err := client.Invite.Create().
		SetCode("ABCD1234").
		SetProjectID(proj.ID).
		SetCreatedByID(admin.ID).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, "ABCD1234", ok.Code)

	_, err = client.Invite.Create().
		SetCode(strings.Repeat("X", 65)).
		SetProjectID(proj.ID).
		SetCreatedByID(admin.ID).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsValidationError(err), "MaxLen(64) on code: %v", err)
}

// TestNewsPost_BodyMaxLen_P3BE17 — oversize body → Ent validation (TEXT, без varchar migrate).
func TestNewsPost_BodyMaxLen_P3BE17(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:news_maxlen_p3be17_"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("news-maxlen@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.NewsPost.Create().
		SetTitle("ok").
		SetBody(strings.Repeat("a", 50000)).
		SetAuthorID(admin.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.NewsPost.Create().
		SetTitle("over").
		SetBody(strings.Repeat("b", 50001)).
		SetAuthorID(admin.ID).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsValidationError(err), "MaxLen(50000) on body: %v", err)
}

// TestComment_MentionsUUID_P3BE17 — mentions хранятся как []uuid.UUID.
func TestComment_MentionsUUID_P3BE17(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:comment_mentions_p3be17_"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	author, err := client.User.Create().
		SetEmail("cmt-mentions@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	mentioned := uuid.New()
	proj, err := client.Project.Create().SetName("Mentions").SetKey("MNT1").Save(ctx)
	require.NoError(t, err)
	status, err := client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)
	iss, err := client.Issue.Create().
		SetTitle("t").
		SetIssueNumber(1).
		SetIssueKey("MNT1-1").
		SetIssueType(issue.IssueTypeTask).
		SetProjectID(proj.ID).
		SetStatusID(status.ID).
		SetReporterID(author.ID).
		Save(ctx)
	require.NoError(t, err)

	c, err := client.Comment.Create().
		SetBodyJSON(map[string]interface{}{"type": "doc"}).
		SetBodyText("hi @" + mentioned.String()).
		SetMentions([]uuid.UUID{mentioned}).
		SetIssueID(iss.ID).
		SetAuthorID(author.ID).
		Save(ctx)
	require.NoError(t, err)
	require.Len(t, c.Mentions, 1)
	assert.Equal(t, mentioned, c.Mentions[0])
}
