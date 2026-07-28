//go:build hubtest

package ent_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/githublink"
	"github.com/valerii/hub/backend/ent/hubcolumn"
	"github.com/valerii/hub/backend/ent/invite"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/notification"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// TestSchema_TASK107_Markers — schema-source evidence for A2W8-1/2/3 indexes/FK/polymorphic.
func TestSchema_TASK107_Markers(t *testing.T) {
	dir := schemaDir(t)

	board, err := os.ReadFile(filepath.Join(dir, "board.go"))
	require.NoError(t, err)
	assert.Contains(t, string(board), `field.Time("updated_at")`)
	assert.Contains(t, string(board), `field.Time("deleted_at")`)
	assert.Contains(t, string(board), `index.Edges("project").Fields("deleted_at")`)

	aip, err := os.ReadFile(filepath.Join(dir, "aiuserpreference.go"))
	require.NoError(t, err)
	assert.Contains(t, string(aip), `edge.To("default_provider", AIProvider.Type)`)
	assert.Contains(t, string(aip), `Field("default_provider_id")`)

	crmc, err := os.ReadFile(filepath.Join(dir, "crmcontact.go"))
	require.NoError(t, err)
	assert.Contains(t, string(crmc), `index.Fields("email")`)

	crmd, err := os.ReadFile(filepath.Join(dir, "crmdeal.go"))
	require.NoError(t, err)
	assert.Contains(t, string(crmd), `index.Edges("owner").Fields("deleted_at")`)

	gl, err := os.ReadFile(filepath.Join(dir, "gitlabintegration.go"))
	require.NoError(t, err)
	glStr := string(gl)
	idx := strings.Index(glStr, `field.String("webhook_secret")`)
	require.GreaterOrEqual(t, idx, 0)
	secretBlock := glStr[idx:]
	if len(secretBlock) > 240 {
		secretBlock = secretBlock[:240]
	}
	assert.Contains(t, secretBlock, `NotEmpty()`)
	assert.NotContains(t, secretBlock, `Optional()`)

	ghl, err := os.ReadFile(filepath.Join(dir, "githublink.go"))
	require.NoError(t, err)
	assert.Contains(t, string(ghl), `field.Int("version")`)
	assert.Contains(t, string(ghl), `index.Edges("integration").Fields("external_id", "link_type").Unique()`)

	gll, err := os.ReadFile(filepath.Join(dir, "gitlablink.go"))
	require.NoError(t, err)
	assert.Contains(t, string(gll), `index.Edges("integration").Fields("external_id", "link_type").Unique()`)

	hb, err := os.ReadFile(filepath.Join(dir, "hubbase.go"))
	require.NoError(t, err)
	assert.Contains(t, string(hb), `index.Edges("owner").Fields("deleted_at")`)

	hc, err := os.ReadFile(filepath.Join(dir, "hubcolumn.go"))
	require.NoError(t, err)
	assert.Contains(t, string(hc), `index.Edges("table").Fields("name").Unique()`)
	assert.Contains(t, string(hc), `index.Edges("table").Fields("deleted_at")`)

	hr, err := os.ReadFile(filepath.Join(dir, "hubrecord.go"))
	require.NoError(t, err)
	assert.Contains(t, string(hr), `index.Edges("table").Fields("deleted_at")`)

	inv, err := os.ReadFile(filepath.Join(dir, "invite.go"))
	require.NoError(t, err)
	assert.Contains(t, string(inv), `field.Int("version")`)
	assert.Contains(t, string(inv), `field.Time("revoked_at")`)

	n, err := os.ReadFile(filepath.Join(dir, "notification.go"))
	require.NoError(t, err)
	assert.Contains(t, string(n), `field.UUID("entity_id", uuid.UUID{})`)
	assert.Contains(t, string(n), `field.Int("version")`)

	mig := filepath.Join(filepath.Dir(dir), "..", "migrations", "20260728_task107_ent_indexes_fk_polymorphic.sql")
	migBody, err := os.ReadFile(mig)
	require.NoError(t, err)
	assert.Contains(t, string(migBody), "A2W8-1-004")
	assert.Contains(t, string(migBody), "A2W8-3-006")
}

// TestSchema_TASK107_RuntimeConstraints — enttest mutation proof for unique/FK/UUID.
func TestSchema_TASK107_RuntimeConstraints(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:task107_"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("task107-owner@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("T107").
		SetKey("T107").
		Save(ctx)
	require.NoError(t, err)

	// A2W8-1-004: Board soft-delete fields persist.
	b, err := client.Board.Create().
		SetName("Main").
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)
	require.Nil(t, b.DeletedAt)
	now := time.Now().UTC()
	b2, err := client.Board.UpdateOneID(b.ID).SetDeletedAt(now).Save(ctx)
	require.NoError(t, err)
	require.NotNil(t, b2.DeletedAt)

	// A2W8-1-005: default_provider_id FK — orphan UUID rejected.
	orphan := uuid.New()
	_, err = client.AIUserPreference.Create().
		SetUserID(owner.ID).
		SetDefaultProviderID(orphan).
		Save(ctx)
	require.Error(t, err, "orphan default_provider_id must fail FK")

	prov, err := client.AIProvider.Create().
		SetName("local").
		SetProviderType("ollama").
		SetBaseURL("http://127.0.0.1:11434").
		SetOwnerID(owner.ID).
		Save(ctx)
	require.NoError(t, err)
	pref, err := client.AIUserPreference.Create().
		SetUserID(owner.ID).
		SetDefaultProviderID(prov.ID).
		Save(ctx)
	require.NoError(t, err)
	require.NotNil(t, pref.DefaultProviderID)
	assert.Equal(t, prov.ID, *pref.DefaultProviderID)

	// A2W8-2-007: GitLab webhook_secret required.
	_, err = client.GitLabIntegration.Create().
		SetProjectPath("group/repo").
		SetAccessToken("token").
		SetProjectID(proj.ID).
		Save(ctx)
	require.Error(t, err, "empty webhook_secret must fail NotEmpty")

	gig, err := client.GitLabIntegration.Create().
		SetProjectPath("group/repo").
		SetAccessToken("token").
		SetWebhookSecret("secret").
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, "secret", gig.WebhookSecret)

	// A2W8-2-008: unique (integration, external_id, link_type)
	ws, err := client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)
	iss, err := client.Issue.Create().
		SetTitle("i").
		SetIssueKey("T107-1").
		SetIssueNumber(1).
		SetIssueType(issue.IssueTypeTask).
		SetProjectID(proj.ID).
		SetStatusID(ws.ID).
		SetReporterID(owner.ID).
		Save(ctx)
	require.NoError(t, err)
	ghInt, err := client.GitHubIntegration.Create().
		SetRepoOwner("o").
		SetRepoName("r").
		SetAccessToken("t").
		SetWebhookSecret("s").
		SetProjectID(proj.ID).
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
	_, err = client.GitHubLink.Create().
		SetLinkType(githublink.LinkTypePullRequest).
		SetExternalID("1").
		SetTitle("dup").
		SetURL("https://example.com/1").
		SetIssueID(iss.ID).
		SetIntegrationID(ghInt.ID).
		Save(ctx)
	require.Error(t, err, "duplicate github link must fail unique")

	// A2W8-2-009: unique (table, name) on HubColumn
	base, err := client.HubBase.Create().
		SetName("Base").
		SetOwnerID(owner.ID).
		Save(ctx)
	require.NoError(t, err)
	table, err := client.HubTable.Create().
		SetName("T").
		SetBaseID(base.ID).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.HubColumn.Create().
		SetName("col").
		SetColType(hubcolumn.ColTypeText).
		SetTableID(table.ID).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.HubColumn.Create().
		SetName("col").
		SetColType(hubcolumn.ColTypeNumber).
		SetTableID(table.ID).
		Save(ctx)
	require.Error(t, err, "duplicate column name must fail unique")

	// A2W8-3-002: Invite revoked_at + version
	inv, err := client.Invite.Create().
		SetCode("ABCD1234").
		SetProjectID(proj.ID).
		SetCreatedByID(owner.ID).
		SetProjectRole(invite.ProjectRoleDeveloper).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, inv.Version)
	require.Nil(t, inv.RevokedAt)
	revoked, err := client.Invite.UpdateOneID(inv.ID).
		SetRevokedAt(time.Now().UTC()).
		AddVersion(1).
		Save(ctx)
	require.NoError(t, err)
	require.NotNil(t, revoked.RevokedAt)
	assert.Equal(t, 2, revoked.Version)

	// A2W8-3-006: Notification entity_id UUID + version on mark-read path
	eid := uuid.New()
	n, err := client.Notification.Create().
		SetUserID(owner.ID).
		SetType("assigned").
		SetTitle("t").
		SetEntityType("issue").
		SetEntityID(eid).
		Save(ctx)
	require.NoError(t, err)
	require.NotNil(t, n.EntityID)
	assert.Equal(t, eid, *n.EntityID)
	assert.Equal(t, 1, n.Version)
	_, err = client.Notification.UpdateOneID(n.ID).
		SetIsRead(true).
		AddVersion(1).
		Save(ctx)
	require.NoError(t, err)
	got, err := client.Notification.Query().Where(notification.IDEQ(n.ID)).Only(ctx)
	require.NoError(t, err)
	assert.True(t, got.IsRead)
	assert.Equal(t, 2, got.Version)
}
