//go:build hubtest

package tracker_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/tracker"
)

// ---------------------------------------------------------------------------
// A3W1-2-001 — UpdateIssue rejects foreign labels (mirror CreateIssue P1-BE-03)
// Mutation: убрать CountLabelsInProject перед UpdateIssue Labels → FAIL.
// ---------------------------------------------------------------------------

func TestUpdateIssue_RejectsForeignLabel_A3W1_2_001(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewTrackerService(env.repo, nil)

	created, err := svc.CreateIssue(env.ctx, env.projectID, env.userID, tracker.CreateIssueRequest{
		Title: "Host", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Label Upd").
		SetKey("FLU").
		Save(env.ctx)
	require.NoError(t, err)

	foreignLabel, err := env.client.Label.Create().
		SetName("alien-upd").
		SetColor("#ff00ff").
		SetProjectID(foreign.ID).
		Save(env.ctx)
	require.NoError(t, err)

	labels := []string{foreignLabel.ID.String()}
	_, err = svc.UpdateIssue(env.ctx, env.projectID, created.IssueKey, env.userID, tracker.UpdateIssueRequest{
		Labels:  &labels,
		Version: created.Version,
	})
	require.Error(t, err)
	requireAppErrorCode(t, err, 400)

	own, err := env.repo.CreateLabel(env.ctx, env.projectID, "ok", "#00ff00", nil)
	require.NoError(t, err)
	okLabels := []string{own.ID.String()}
	updated, err := svc.UpdateIssue(env.ctx, env.projectID, created.IssueKey, env.userID, tracker.UpdateIssueRequest{
		Labels:  &okLabels,
		Version: created.Version,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Len(t, updated.Labels, 1)
	assert.Equal(t, own.ID.String(), updated.Labels[0].ID)
}

// ---------------------------------------------------------------------------
// A3W1-2-002 — CreateLabel rejects foreign LabelGroup
// Mutation: принять любой groupID без FindLabelGroupInProject → FAIL.
// ---------------------------------------------------------------------------

func TestCreateLabel_RejectsForeignGroup_A3W1_2_002(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewTrackerService(env.repo, nil)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Group Proj").
		SetKey("FGP").
		Save(env.ctx)
	require.NoError(t, err)

	foreignGroup, err := env.client.LabelGroup.Create().
		SetName("Alien Group").
		SetProjectID(foreign.ID).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = svc.CreateLabel(env.ctx, env.projectID, tracker.CreateLabelRequest{
		Name: "bad", Color: "#111111", GroupID: foreignGroup.ID.String(),
	})
	require.Error(t, err)
	requireAppErrorCode(t, err, 404)

	ownGroup, err := svc.CreateLabelGroup(env.ctx, env.projectID, tracker.CreateLabelGroupRequest{Name: "Local"})
	require.NoError(t, err)
	lbl, err := svc.CreateLabel(env.ctx, env.projectID, tracker.CreateLabelRequest{
		Name: "good", Color: "#222222", GroupID: ownGroup.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, lbl)
	assert.Equal(t, "good", lbl.Name)
}

// ---------------------------------------------------------------------------
// A3W1-2-003 — CreateRule/UpdateRule reject foreign channel_id
// Mutation: только non-empty check без FindChannelInProject → FAIL.
// ---------------------------------------------------------------------------

func TestCreateUpdateRule_RejectsForeignChannel_A3W1_2_003(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewAutomationService(env.repo)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Chan Auto").
		SetKey("FCA").
		Save(env.ctx)
	require.NoError(t, err)

	foreignCh, err := env.client.Channel.Create().
		SetName("other").
		SetChannelType("public").
		SetProjectID(foreign.ID).
		SetCreatedBy(env.userID).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = svc.CreateRule(env.ctx, env.projectID, env.userID, tracker.CreateAutomationRuleRequest{
		Name:        "post foreign",
		TriggerType: tracker.TriggerIssueCreated,
		ActionType:  tracker.ActionPostToChannel,
		ActionConfig: map[string]interface{}{
			"channel_id": foreignCh.ID.String(),
		},
	})
	require.Error(t, err)
	requireAppErrorCode(t, err, 404)

	localCh, err := env.client.Channel.Create().
		SetName("local").
		SetChannelType("public").
		SetProjectID(env.projectID).
		SetCreatedBy(env.userID).
		Save(env.ctx)
	require.NoError(t, err)

	created, err := svc.CreateRule(env.ctx, env.projectID, env.userID, tracker.CreateAutomationRuleRequest{
		Name:        "post local",
		TriggerType: tracker.TriggerIssueCreated,
		ActionType:  tracker.ActionPostToChannel,
		ActionConfig: map[string]interface{}{
			"channel_id": localCh.ID.String(),
		},
	})
	require.NoError(t, err)

	ruleID, err := uuid.Parse(created.ID)
	require.NoError(t, err)
	cfg := map[string]interface{}{"channel_id": foreignCh.ID.String()}
	_, err = svc.UpdateRule(env.ctx, env.projectID, ruleID, tracker.UpdateAutomationRuleRequest{
		ActionConfig: cfg,
	})
	require.Error(t, err)
	requireAppErrorCode(t, err, 404)
}

// ---------------------------------------------------------------------------
// A3W1-2-004 — ListThreadIssueLinks без N+1 FindIssueByID
// Mutation: вернуть per-link FindIssueByID loop → findIssueByIDCalls > 0 FAIL.
// ---------------------------------------------------------------------------

func TestListThreadIssueLinks_NoNPlusOne_A3W1_2_004(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	channelID := uuid.New()

	var created []*tracker.IssueResponse
	for i := 0; i < 3; i++ {
		iss, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
			Title: fmt.Sprintf("L%d", i), IssueType: "task", Priority: "medium",
		})
		require.NoError(t, err)
		created = append(created, iss)
	}

	for _, iss := range created {
		issueID := uuid.MustParse(iss.ID)
		repo.issues[issueID].Edges.Project = &ent.Project{ID: projectID}
		_, err := svc.CreateIssueMessageLink(ctx, projectID, iss.IssueKey, userID, tracker.CreateIssueMessageLinkRequest{
			ChannelID: channelID.String(), MessageID: "thread-root",
		})
		require.NoError(t, err)
	}

	before := repo.findIssueByIDCalls
	links, err := svc.ListThreadIssueLinks(ctx, projectID, channelID, "thread-root")
	require.NoError(t, err)
	require.Len(t, links, 3)
	assert.Equal(t, before, repo.findIssueByIDCalls,
		"A3W1-2-004: ListThreadIssueLinks must not call FindIssueByID (eager batch)")
}

// ---------------------------------------------------------------------------
// A3W1-2-005 — CreateIssueLink maps raw repo err → AppError
// Mutation: return nil, err без wrap → FAIL (не AppError).
// ---------------------------------------------------------------------------

type failCreateIssueLinkRepo struct {
	mockRepo
}

func (r *failCreateIssueLinkRepo) CreateIssueLink(context.Context, uuid.UUID, uuid.UUID, string, string) (*ent.IssueLink, error) {
	return nil, fmt.Errorf("sql: boom constraint")
}

func TestCreateIssueLink_MapsRepoErrToAppError_A3W1_2_005(t *testing.T) {
	base := newMockRepo()
	repo := &failCreateIssueLinkRepo{mockRepo: *base}
	svc := tracker.NewTrackerService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	src, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "S", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)
	tgt, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "T", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	_, err = svc.CreateIssueLink(ctx, projectID, src.IssueKey, userID, tracker.CreateIssueLinkRequest{
		TargetIssueID: tgt.ID, LinkType: "blocks",
	})
	require.Error(t, err)
	var ae *apperrors.AppError
	require.True(t, errors.As(err, &ae), "expected AppError, got %T: %v", err, err)
	assert.Equal(t, 500, ae.Code)
}

// ---------------------------------------------------------------------------
// A3W1-2-006 — UpdateComment: missing → 404; version mismatch → 409
// Mutation: любой IsNotFound → 409 → missing-comment 404 тест FAIL.
// ---------------------------------------------------------------------------

func TestUpdateComment_MissingVsVersionConflict_A3W1_2_006(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewTrackerService(env.repo, nil)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Cmt", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	cmt, err := env.repo.CreateComment(env.ctx, issue.ID, env.userID, tracker.CreateCommentInput{
		BodyJSON: map[string]interface{}{"t": "1"},
		BodyText: "one",
	})
	require.NoError(t, err)

	_, err = svc.UpdateComment(env.ctx, env.projectID, issue.IssueKey, env.userID, uuid.New(), tracker.UpdateCommentRequest{
		BodyJSON: map[string]interface{}{"t": "x"},
		BodyText: "x",
		Version:  1,
	})
	require.Error(t, err)
	requireAppErrorCode(t, err, 404)

	_, err = svc.UpdateComment(env.ctx, env.projectID, issue.IssueKey, env.userID, cmt.ID, tracker.UpdateCommentRequest{
		BodyJSON: map[string]interface{}{"t": "stale"},
		BodyText: "stale",
		Version:  99,
	})
	require.Error(t, err)
	requireAppErrorCode(t, err, 409)

	updated, err := svc.UpdateComment(env.ctx, env.projectID, issue.IssueKey, env.userID, cmt.ID, tracker.UpdateCommentRequest{
		BodyJSON: map[string]interface{}{"t": "ok"},
		BodyText: "ok",
		Version:  cmt.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", updated.BodyText)
}
