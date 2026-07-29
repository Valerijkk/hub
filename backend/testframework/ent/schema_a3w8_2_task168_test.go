//go:build hubtest

package ent_test

// TASK-168 / A3W8-2 — FileMetadata validators + owned_child OnDelete Cascade.
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
	"github.com/valerii/hub/backend/ent/coffeepair"
	"github.com/valerii/hub/backend/ent/coffeeround"
	"github.com/valerii/hub/backend/ent/comment"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/githublink"
	"github.com/valerii/hub/backend/ent/gitlablink"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/reaction"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func task168SchemaDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "ent", "schema")
}

func task168Open(t *testing.T, name string) *ent.Client {
	t.Helper()
	dsn := "file:task168_" + name + "_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func task168SeedProject(t *testing.T, client *ent.Client) (*ent.User, *ent.Project, *ent.WorkflowStatus) {
	t.Helper()
	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("t168-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		Save(ctx)
	require.NoError(t, err)
	p, err := client.Project.Create().
		SetKey("T168").
		SetName("Task168").
		Save(ctx)
	require.NoError(t, err)
	ws, err := client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetProjectID(p.ID).
		Save(ctx)
	require.NoError(t, err)
	return u, p, ws
}

// TestSchema_TASK168_OnDeleteCascades — A3W8-2-002…005.
// Mutation: убрать entsql.OnDelete(Cascade) с указанного edge → FAIL.
func TestSchema_TASK168_OnDeleteCascades(t *testing.T) {
	dir := task168SchemaDir(t)
	mustContain := map[string][]string{
		"coffeeround.go": {
			`edge.To("pairs"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
		"hubbase.go": {
			`edge.To("tables"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
		"githubintegration.go": {
			`edge.To("github_links"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
		"gitlabintegration.go": {
			`edge.To("gitlab_links"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
		"comment.go": {
			`edge.To("reactions"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
	}
	for file, needles := range mustContain {
		src, err := os.ReadFile(filepath.Join(dir, file))
		require.NoError(t, err, file)
		body := string(src)
		for _, n := range needles {
			assert.Contains(t, body, n, "%s missing %q", file, n)
		}
	}
}

// TestSchema_TASK168_FileMetadataValidators — A3W8-2-001.
// Mutation: убрать NonNegative/Required/Validate → FAIL.
func TestSchema_TASK168_FileMetadataValidators(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(task168SchemaDir(t), "filemetadata.go"))
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, `NonNegative()`)
	assert.Contains(t, body, `Validate(validateFileEntityID)`)
	assert.Contains(t, body, `Required()`)
	assert.Contains(t, body, `case "me", "temp":`)
}

// TestSchema_TASK168_AccessTokenNotEmpty — A3W8-2-006.
// Mutation: убрать NotEmpty с access_token → FAIL.
func TestSchema_TASK168_AccessTokenNotEmpty(t *testing.T) {
	dir := task168SchemaDir(t)
	for _, file := range []string{"githubintegration.go", "gitlabintegration.go"} {
		src, err := os.ReadFile(filepath.Join(dir, file))
		require.NoError(t, err, file)
		body := string(src)
		assert.Contains(t, body, `field.String("access_token")`, file)
		assert.Contains(t, body, `NotEmpty()`, file)
		assert.Contains(t, body, `Sensitive()`, file)
	}
}

// TestSchema_TASK168_MigrateSQLExists — migrate marker.
func TestSchema_TASK168_MigrateSQLExists(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	sqlPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "20260729_task168_filemetadata_ondelete_cascade.sql")
	src, err := os.ReadFile(sqlPath)
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, "needs-owner-apply")
	assert.Contains(t, body, "A3W8-2-001")
	assert.Contains(t, body, "coffee_pairs_coffee_rounds_pairs")
	assert.Contains(t, body, "hub_tables_hub_bases_tables")
	assert.Contains(t, body, "git_hub_links_git_hub_integrations_github_links")
	assert.Contains(t, body, "git_lab_links_git_lab_integrations_gitlab_links")
	assert.Contains(t, body, "reactions_comments_reactions")
	assert.Contains(t, body, "ON DELETE CASCADE")
}

// TestCascade_TASK168_CoffeeRoundHardDeleteRemovesPairs — A3W8-2-002.
// Mutation: убрать Cascade с pairs → FAIL (FK / orphan).
func TestCascade_TASK168_CoffeeRoundHardDeleteRemovesPairs(t *testing.T) {
	client := task168Open(t, "coffee")
	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("c168-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		Save(ctx)
	require.NoError(t, err)
	u2, err := client.User.Create().
		SetEmail("c168b-" + uuid.New().String()[:8] + "@t.com").
		SetName("U2").
		Save(ctx)
	require.NoError(t, err)

	round, err := client.CoffeeRound.Create().
		SetTriggeredByID(u.ID).
		SetPairCount(1).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.CoffeePair.Create().
		SetRoundID(round.ID).
		SetParticipantAID(u.ID).
		SetParticipantBID(u2.ID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.CoffeeRound.DeleteOneID(round.ID).Exec(ctx))

	n, err := client.CoffeePair.Query().Where(coffeepair.HasRoundWith(coffeeround.IDEQ(round.ID))).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "pairs must CASCADE with coffee round")
}

// TestCascade_TASK168_HubBaseHardDeleteRemovesTables — A3W8-2-003.
// Mutation: убрать Cascade с tables → FAIL (FK trap).
func TestCascade_TASK168_HubBaseHardDeleteRemovesTables(t *testing.T) {
	client := task168Open(t, "hubbase")
	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("hb168-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		Save(ctx)
	require.NoError(t, err)
	base, err := client.HubBase.Create().SetName("B").SetOwnerID(u.ID).Save(ctx)
	require.NoError(t, err)
	_, err = client.HubTable.Create().SetName("T").SetBaseID(base.ID).Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.HubBase.DeleteOneID(base.ID).Exec(ctx))

	n, err := client.HubTable.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "tables must CASCADE with hub base")
}

// TestCascade_TASK168_GitHubIntegrationDeleteRemovesLinks — A3W8-2-004.
// Mutation: убрать Cascade с github_links → FAIL (FK trap).
func TestCascade_TASK168_GitHubIntegrationDeleteRemovesLinks(t *testing.T) {
	client := task168Open(t, "ghub")
	ctx := context.Background()
	u, p, ws := task168SeedProject(t, client)
	iss, err := client.Issue.Create().
		SetTitle("i").
		SetIssueKey("T168-1").
		SetIssueNumber(1).
		SetIssueType(issue.IssueTypeTask).
		SetProjectID(p.ID).
		SetStatusID(ws.ID).
		SetReporterID(u.ID).
		Save(ctx)
	require.NoError(t, err)
	ghInt, err := client.GitHubIntegration.Create().
		SetRepoOwner("o").
		SetRepoName("r").
		SetAccessToken("tok").
		SetWebhookSecret("sec").
		SetProjectID(p.ID).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.GitHubLink.Create().
		SetLinkType(githublink.LinkTypePullRequest).
		SetExternalID("1").
		SetTitle("PR").
		SetURL("https://example.com/1").
		SetIssueID(iss.ID).
		SetIntegrationID(ghInt.ID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.GitHubIntegration.DeleteOneID(ghInt.ID).Exec(ctx))

	n, err := client.GitHubLink.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "github_links must CASCADE with integration")
}

// TestCascade_TASK168_GitLabIntegrationDeleteRemovesLinks — A3W8-2-004.
func TestCascade_TASK168_GitLabIntegrationDeleteRemovesLinks(t *testing.T) {
	client := task168Open(t, "glab")
	ctx := context.Background()
	u, p, ws := task168SeedProject(t, client)
	iss, err := client.Issue.Create().
		SetTitle("i").
		SetIssueKey("T168-2").
		SetIssueNumber(2).
		SetIssueType(issue.IssueTypeTask).
		SetProjectID(p.ID).
		SetStatusID(ws.ID).
		SetReporterID(u.ID).
		Save(ctx)
	require.NoError(t, err)
	glInt, err := client.GitLabIntegration.Create().
		SetProjectPath("g/r").
		SetAccessToken("tok").
		SetWebhookSecret("sec").
		SetProjectID(p.ID).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.GitLabLink.Create().
		SetLinkType(gitlablink.LinkTypeMergeRequest).
		SetExternalID("1").
		SetTitle("MR").
		SetURL("https://example.com/1").
		SetIssueID(iss.ID).
		SetIntegrationID(glInt.ID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.GitLabIntegration.DeleteOneID(glInt.ID).Exec(ctx))

	n, err := client.GitLabLink.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "gitlab_links must CASCADE with integration")
}

// TestCascade_TASK168_CommentHardDeleteRemovesReactions — A3W8-2-005.
func TestCascade_TASK168_CommentHardDeleteRemovesReactions(t *testing.T) {
	client := task168Open(t, "react")
	ctx := context.Background()
	u, p, ws := task168SeedProject(t, client)
	iss, err := client.Issue.Create().
		SetTitle("i").
		SetIssueKey("T168-3").
		SetIssueNumber(3).
		SetIssueType(issue.IssueTypeTask).
		SetProjectID(p.ID).
		SetStatusID(ws.ID).
		SetReporterID(u.ID).
		Save(ctx)
	require.NoError(t, err)
	cmt, err := client.Comment.Create().
		SetBodyJSON(map[string]interface{}{"t": "x"}).
		SetBodyText("x").
		SetIssueID(iss.ID).
		SetAuthorID(u.ID).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.Reaction.Create().
		SetEmoji("👍").
		SetCommentID(cmt.ID).
		SetUserID(u.ID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.Comment.DeleteOneID(cmt.ID).Exec(ctx))

	n, err := client.Reaction.Query().Where(reaction.HasCommentWith(comment.IDEQ(cmt.ID))).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "reactions must CASCADE with comment")
}

// TestValidator_TASK168_FileMetadataNegativeSize — A3W8-2-001.
// Mutation: убрать NonNegative → negative size проходит → FAIL.
func TestValidator_TASK168_FileMetadataNegativeSize(t *testing.T) {
	client := task168Open(t, "fmsize")
	ctx := context.Background()
	u, p, _ := task168SeedProject(t, client)
	_, err := client.FileMetadata.Create().
		SetFilename("a.txt").
		SetContentType("text/plain").
		SetSize(-1).
		SetS3Key("k").
		SetS3Bucket("b").
		SetUploaderID(u.ID).
		SetProjectID(p.ID).
		Save(ctx)
	require.Error(t, err, "negative size must fail NonNegative")
}

// TestValidator_TASK168_FileMetadataUploaderRequired — A3W8-2-001.
// Mutation: снова Optional uploader → create без uploader проходит → FAIL.
func TestValidator_TASK168_FileMetadataUploaderRequired(t *testing.T) {
	client := task168Open(t, "fmup")
	ctx := context.Background()
	_, p, _ := task168SeedProject(t, client)
	_, err := client.FileMetadata.Create().
		SetFilename("a.txt").
		SetContentType("text/plain").
		SetSize(1).
		SetS3Key("k").
		SetS3Bucket("b").
		SetProjectID(p.ID).
		Save(ctx)
	require.Error(t, err, "missing uploader must fail Required")
}

// TestValidator_TASK168_FileMetadataEntityID — A3W8-2-001.
// Mutation: убрать Validate → garbage entity_id проходит → FAIL.
func TestValidator_TASK168_FileMetadataEntityID(t *testing.T) {
	client := task168Open(t, "fmeid")
	ctx := context.Background()
	u, p, _ := task168SeedProject(t, client)

	_, err := client.FileMetadata.Create().
		SetFilename("bad.txt").
		SetContentType("text/plain").
		SetSize(1).
		SetS3Key("k-bad").
		SetS3Bucket("b").
		SetEntityID("not-a-uuid-or-sentinel").
		SetUploaderID(u.ID).
		SetProjectID(p.ID).
		Save(ctx)
	require.Error(t, err, "garbage entity_id must fail Validate")

	ok, err := client.FileMetadata.Create().
		SetFilename("ok.txt").
		SetContentType("text/plain").
		SetSize(1).
		SetS3Key("k-me").
		SetS3Bucket("b").
		SetEntityID("me").
		SetUploaderID(u.ID).
		SetProjectID(p.ID).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, "me", ok.EntityID)

	uid := uuid.New().String()
	ok2, err := client.FileMetadata.Create().
		SetFilename("ok2.txt").
		SetContentType("text/plain").
		SetSize(1).
		SetS3Key("k-uid").
		SetS3Bucket("b").
		SetEntityID(uid).
		SetUploaderID(u.ID).
		SetProjectID(p.ID).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, uid, ok2.EntityID)
}

// TestValidator_TASK168_EmptyAccessToken — A3W8-2-006.
// Mutation: убрать NotEmpty → empty PAT проходит → FAIL.
func TestValidator_TASK168_EmptyAccessToken(t *testing.T) {
	client := task168Open(t, "pat")
	ctx := context.Background()
	_, p, _ := task168SeedProject(t, client)

	_, err := client.GitHubIntegration.Create().
		SetRepoOwner("o").
		SetRepoName("r").
		SetAccessToken("").
		SetWebhookSecret("sec").
		SetProjectID(p.ID).
		Save(ctx)
	require.Error(t, err, "empty GitHub access_token must fail NotEmpty")

	_, err = client.GitLabIntegration.Create().
		SetProjectPath("g/r").
		SetAccessToken("").
		SetWebhookSecret("sec").
		SetProjectID(p.ID).
		Save(ctx)
	require.Error(t, err, "empty GitLab access_token must fail NotEmpty")
}
