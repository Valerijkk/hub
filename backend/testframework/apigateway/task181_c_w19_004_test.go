//go:build hubtest

package apigateway_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
	entproject "github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/webhookendpoint"
	"github.com/valerii/hub/backend/internal/apigateway"
	"github.com/valerii/hub/backend/internal/project"
)

// TestSoftDeleteProject_DeactivatesWebhookEndpoints_TASK181 — C-W19-004.
// SoftDeleteWithCrossLinkCleanup в той же tx гасит endpoints; ListActive пуст;
// list-by-project для trash пуст.
// Mutation: убрать SetIsActive(false) из SoftDelete tx → IsActive остаётся true → FAIL.
// Mutation: убрать project.DeletedAtIsNil из ListWebhookEndpointsByProject → list непуст → FAIL.
func TestSoftDeleteProject_DeactivatesWebhookEndpoints_TASK181(t *testing.T) {
	dsn := "file:wh_task181_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	projRepo := project.NewEntRepository(client, nil)
	gwRepo := apigateway.NewEntRepository(client, apigatewayTestCipher(t))

	alive, err := client.Project.Create().SetName("Alive181").SetKey("A181").Save(ctx)
	require.NoError(t, err)
	dead, err := client.Project.Create().SetName("Dead181").SetKey("D181").Save(ctx)
	require.NoError(t, err)

	aliveEp, err := gwRepo.CreateWebhookEndpoint(ctx, alive.ID, "https://example.com/alive181", "s-alive", []string{"issue.created"})
	require.NoError(t, err)
	deadEp, err := gwRepo.CreateWebhookEndpoint(ctx, dead.ID, "https://example.com/dead181", "s-dead", []string{"issue.created"})
	require.NoError(t, err)
	require.True(t, deadEp.IsActive)

	_, _, err = projRepo.SoftDeleteWithCrossLinkCleanup(ctx, dead.ID)
	require.NoError(t, err)

	got, err := client.WebhookEndpoint.Get(ctx, deadEp.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive, "soft-delete project must deactivate endpoints in same tx")
	assert.Nil(t, got.DeletedAt, "endpoint row stays (IsActive=false), not soft-deleted")

	aliveStill, err := client.WebhookEndpoint.Get(ctx, aliveEp.ID)
	require.NoError(t, err)
	assert.True(t, aliveStill.IsActive)

	active, err := gwRepo.ListActiveWebhookEndpoints(ctx)
	require.NoError(t, err)
	require.Len(t, active, 1)
	require.NotNil(t, active[0].Edges.Project)
	assert.Equal(t, alive.ID, active[0].Edges.Project.ID)

	byDead, err := gwRepo.ListWebhookEndpointsByProject(ctx, dead.ID)
	require.NoError(t, err)
	assert.Empty(t, byDead, "list-by-project must be empty for trash project")

	byAlive, err := gwRepo.ListWebhookEndpointsByProject(ctx, alive.ID)
	require.NoError(t, err)
	require.Len(t, byAlive, 1)
	assert.Equal(t, aliveEp.ID, byAlive[0].ID)
}

// TestListWebhookEndpointsByProject_SoftDeletedProject_Empty_TASK181 —
// даже без SoftDeleteWithCrossLinkCleanup (только DeletedAt на project)
// list-by-project скрывает endpoints trash.
// Mutation: убрать project.DeletedAtIsNil из Where → list непуст → FAIL.
func TestListWebhookEndpointsByProject_SoftDeletedProject_Empty_TASK181(t *testing.T) {
	dsn := "file:wh_list181_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	gwRepo := apigateway.NewEntRepository(client, apigatewayTestCipher(t))

	proj, err := client.Project.Create().SetName("TrashList").SetKey("TL81").Save(ctx)
	require.NoError(t, err)
	_, err = gwRepo.CreateWebhookEndpoint(ctx, proj.ID, "https://example.com/tl81", "sec", []string{"*"})
	require.NoError(t, err)

	now := time.Now().UTC()
	_, err = client.Project.UpdateOneID(proj.ID).SetDeletedAt(now).Save(ctx)
	require.NoError(t, err)

	list, err := gwRepo.ListWebhookEndpointsByProject(ctx, proj.ID)
	require.NoError(t, err)
	assert.Empty(t, list)

	// Endpoint row still active in DB — filter is on project, not IsActive.
	n, err := client.WebhookEndpoint.Query().
		Where(
			webhookendpoint.HasProjectWith(entproject.IDEQ(proj.ID)),
			webhookendpoint.DeletedAtIsNil(),
			webhookendpoint.IsActive(true),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

// TestSoftDeleteProject_WebhookDeactivate_SourceGuard_TASK181 — mutation source.
// Mutation: убрать SetIsActive / WebhookEndpoint.Update из SoftDeleteWithCrossLinkCleanup → FAIL.
// Mutation: убрать DeletedAtIsNil из ListWebhookEndpointsByProject → FAIL.
// Mutation: ослабить ListActive HasProjectWith(DeletedAtIsNil) → FAIL (TASK-152 lock).
func TestSoftDeleteProject_WebhookDeactivate_SourceGuard_TASK181(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "internal")

	projSrc, err := os.ReadFile(filepath.Join(root, "project", "repository.go"))
	require.NoError(t, err)
	projBody := string(projSrc)
	fnIdx := strings.Index(projBody, "func (r *EntRepository) SoftDeleteWithCrossLinkCleanup")
	require.GreaterOrEqual(t, fnIdx, 0)
	next := strings.Index(projBody[fnIdx+1:], "\nfunc (r *EntRepository)")
	require.Greater(t, next, 0)
	fnBody := projBody[fnIdx : fnIdx+1+next]
	assert.Contains(t, fnBody, "WebhookEndpoint.Update")
	assert.Contains(t, fnBody, "webhookendpoint.HasProjectWith")
	assert.Regexp(t, `(?m)^\t\tSetIsActive\(false\)\.`, fnBody,
		"SetIsActive(false) must be a live call in SoftDeleteWithCrossLinkCleanup (not commented)")

	gwSrc, err := os.ReadFile(filepath.Join(root, "apigateway", "repository.go"))
	require.NoError(t, err)
	gwBody := string(gwSrc)

	listIdx := strings.Index(gwBody, "func (r *EntRepository) ListWebhookEndpointsByProject")
	require.GreaterOrEqual(t, listIdx, 0)
	listNext := strings.Index(gwBody[listIdx+1:], "\nfunc (r *EntRepository)")
	require.Greater(t, listNext, 0)
	listBody := gwBody[listIdx : listIdx+1+listNext]
	assert.Contains(t, listBody, "project.DeletedAtIsNil()")
	assert.Contains(t, listBody, "HasProjectWith")

	activeIdx := strings.Index(gwBody, "func (r *EntRepository) ListActiveWebhookEndpoints")
	require.GreaterOrEqual(t, activeIdx, 0)
	activeNext := strings.Index(gwBody[activeIdx+1:], "\nfunc (r *EntRepository)")
	require.Greater(t, activeNext, 0)
	activeBody := gwBody[activeIdx : activeIdx+1+activeNext]
	assert.Contains(t, activeBody, "HasProjectWith(project.DeletedAtIsNil())",
		"TASK-152 lock: не ослаблять ListActive soft-project filter")
}
