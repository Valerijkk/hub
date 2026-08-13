package tracker_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_FindIssueInProject_RejectsForeignProject_A3_011 — issue проекта B
// не находится через project A (sqlite Ent). Postgres twin:
// TestEntRepository_FindIssueInProject_RejectsForeignProject_Postgres_A3_011.
// Mutation: снять issueProjectIDEQ → PASS на чужом проекте.
func TestEntRepository_FindIssueInProject_RejectsForeignProject_A3_011(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Project A3011 Target").
		SetKey("FRT").
		Save(env.ctx)
	require.NoError(t, err)

	status, err := env.client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory("todo").
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(foreign.ID).
		Save(env.ctx)
	require.NoError(t, err)

	foreignRepo := tracker.NewEntRepository(env.client, nil)
	created, err := foreignRepo.CreateIssue(env.ctx, foreign.ID, tracker.CreateIssueInput{
		Title:      "Secret target",
		IssueType:  "task",
		Priority:   "medium",
		ReporterID: env.userID,
		StatusID:   &status.ID,
	})
	require.NoError(t, err)

	_, err = env.repo.FindIssueInProject(env.ctx, env.projectID, created.ID)
	require.True(t, ent.IsNotFound(err), "foreign issue must not be visible via project A, got %v", err)

	got, err := foreignRepo.FindIssueInProject(env.ctx, foreign.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

// TestCreateIssueLink_RejectsCrossProjectTarget_A3_011 — CreateIssueLink отклоняет
// target из чужого проекта с NotFound (без утечки существования).
func TestCreateIssueLink_RejectsCrossProjectTarget_A3_011(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectA := uuid.New()
	projectB := uuid.New()
	userID := uuid.New()

	// Target first: mock issuesByKey is keyed only by issue_key, so create source
	// last so FindIssueByKey(projectA, source.Key) resolves correctly.
	target, err := repo.CreateIssue(ctx, projectB, tracker.CreateIssueInput{
		Title: "Target B", IssueType: "task", Priority: "medium", ReporterID: userID,
	})
	require.NoError(t, err)

	source, err := svc.CreateIssue(ctx, projectA, userID, tracker.CreateIssueRequest{
		Title: "Source A", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	_, err = svc.CreateIssueLink(ctx, projectA, source.IssueKey, userID, tracker.CreateIssueLinkRequest{
		TargetIssueID: target.ID.String(),
		LinkType:      "blocks",
	})
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Equal(t, 404, appErr.Code)
	assert.Equal(t, "целевая задача не найдена", appErr.Message)

	sourceID, err := uuid.Parse(source.ID)
	require.NoError(t, err)
	links, err := repo.ListIssueLinks(ctx, sourceID)
	require.NoError(t, err)
	assert.Empty(t, links, "cross-project link must not be created")
}
