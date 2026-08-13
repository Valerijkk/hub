//go:build hubtest

package tracker_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tracker"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

func requireAppErrorCode(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "expected AppError, got %T: %v", err, err)
	require.Equal(t, code, appErr.Code, "unexpected AppError: %v", err)
}

// ---------------------------------------------------------------------------
// P1-BE-01 / P1-BE-02 — workflow + custom-field project-scope IDOR (Ent layer)
// Mutation: drop HasProjectWith / projectID from Where → FAIL.
// ---------------------------------------------------------------------------

func TestDeleteCustomField_RejectsForeignProject_P1BE02(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	field, err := env.repo.CreateCustomField(env.ctx, env.projectID, tracker.CreateCustomFieldInput{
		Name: "Severity", FieldType: "select", Options: []string{"S1"},
	})
	require.NoError(t, err)

	foreign, err := env.client.Project.Create().
		SetName("Foreign CF Proj").
		SetKey("FCF").
		Save(env.ctx)
	require.NoError(t, err)

	err = env.repo.DeleteCustomField(env.ctx, foreign.ID, field.ID)
	require.True(t, ent.IsNotFound(err), "cross-project custom field delete must 404, got %v", err)

	still, err := env.client.CustomFieldDefinition.Get(env.ctx, field.ID)
	require.NoError(t, err)
	require.Equal(t, field.ID, still.ID)

	require.NoError(t, env.repo.DeleteCustomField(env.ctx, env.projectID, field.ID))
}

func TestUpdateWorkflowStatus_RejectsForeignProject_P1BE01(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	status, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "Review", "in_progress", "#AAA", 10)
	require.NoError(t, err)

	foreign, err := env.client.Project.Create().
		SetName("Foreign WF Proj").
		SetKey("FWF").
		Save(env.ctx)
	require.NoError(t, err)

	newName := "Hacked"
	_, err = env.repo.UpdateWorkflowStatus(env.ctx, foreign.ID, status.ID, &newName, nil, nil)
	require.True(t, ent.IsNotFound(err), "cross-project status update must 404, got %v", err)

	got, err := env.client.WorkflowStatus.Get(env.ctx, status.ID)
	require.NoError(t, err)
	require.Equal(t, "Review", got.Name)
}

func TestDeleteWorkflowStatus_RejectsForeignProject_P1BE01(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	status, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "Temp", "todo", "#000", 99)
	require.NoError(t, err)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Del Status").
		SetKey("FDS").
		Save(env.ctx)
	require.NoError(t, err)

	err = env.repo.DeleteWorkflowStatus(env.ctx, foreign.ID, status.ID)
	require.True(t, ent.IsNotFound(err), "cross-project status delete must 404, got %v", err)

	_, err = env.client.WorkflowStatus.Get(env.ctx, status.ID)
	require.NoError(t, err)
}

func TestCreateWorkflowTransition_RejectsForeignStatus_P1BE01(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Trans Proj").
		SetKey("FTP").
		Save(env.ctx)
	require.NoError(t, err)

	foreignStatus, err := env.client.WorkflowStatus.Create().
		SetName("Foreign Todo").
		SetCategory("todo").
		SetColor("#111").
		SetSortOrder(0).
		SetProjectID(foreign.ID).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = env.repo.CreateWorkflowTransition(env.ctx, env.projectID, env.statusID, foreignStatus.ID)
	require.True(t, ent.IsNotFound(err), "cross-project transition create must 404, got %v", err)
}

func TestDeleteWorkflowTransition_RejectsForeignProject_P1BE01(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	toStatus, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "Next", "in_progress", "#222", 11)
	require.NoError(t, err)

	tr, err := env.repo.CreateWorkflowTransition(env.ctx, env.projectID, env.statusID, toStatus.ID)
	require.NoError(t, err)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Del Trans").
		SetKey("FDT").
		Save(env.ctx)
	require.NoError(t, err)

	err = env.repo.DeleteWorkflowTransition(env.ctx, foreign.ID, tr.ID)
	require.True(t, ent.IsNotFound(err), "cross-project transition delete must 404, got %v", err)

	_, err = env.client.WorkflowTransition.Get(env.ctx, tr.ID)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// P1-BE-03 — CreateIssue StatusID / ParentID / Labels ∈ project (service+Ent)
// ---------------------------------------------------------------------------

func TestCreateIssue_RejectsForeignStatus_P1BE03(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewTrackerService(env.repo, nil)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Status Proj").
		SetKey("FSP").
		Save(env.ctx)
	require.NoError(t, err)

	foreignStatus, err := env.client.WorkflowStatus.Create().
		SetName("Alien").
		SetCategory("todo").
		SetColor("#333").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(foreign.ID).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = svc.CreateIssue(env.ctx, env.projectID, env.userID, tracker.CreateIssueRequest{
		Title:     "Bad status",
		IssueType: "task",
		Priority:  "medium",
		StatusID:  foreignStatus.ID.String(),
	})
	require.Error(t, err)
	requireAppErrorCode(t, err, 404)
}

func TestCreateIssue_RejectsForeignParent_P1BE03(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewTrackerService(env.repo, nil)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Parent Proj").
		SetKey("FPP").
		Save(env.ctx)
	require.NoError(t, err)

	foreignStatus, err := env.client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory("todo").
		SetColor("#444").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(foreign.ID).
		Save(env.ctx)
	require.NoError(t, err)

	foreignRepo := tracker.NewEntRepository(env.client, nil)
	parent, err := foreignRepo.CreateIssue(env.ctx, foreign.ID, tracker.CreateIssueInput{
		Title: "Parent", IssueType: "task", Priority: "medium",
		ReporterID: env.userID, StatusID: &foreignStatus.ID,
	})
	require.NoError(t, err)

	_, err = svc.CreateIssue(env.ctx, env.projectID, env.userID, tracker.CreateIssueRequest{
		Title:     "Child of foreign",
		IssueType: "task",
		Priority:  "medium",
		ParentID:  parent.ID.String(),
	})
	require.Error(t, err)
	requireAppErrorCode(t, err, 404)
}

func TestCreateIssue_RejectsForeignLabel_P1BE03(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewTrackerService(env.repo, nil)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Label Proj").
		SetKey("FLP").
		Save(env.ctx)
	require.NoError(t, err)

	foreignLabel, err := env.client.Label.Create().
		SetName("alien").
		SetColor("#ff00ff").
		SetProjectID(foreign.ID).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = svc.CreateIssue(env.ctx, env.projectID, env.userID, tracker.CreateIssueRequest{
		Title:     "Bad label",
		IssueType: "task",
		Priority:  "medium",
		Labels:    []string{foreignLabel.ID.String()},
	})
	require.Error(t, err)
	requireAppErrorCode(t, err, 400)
}

// ---------------------------------------------------------------------------
// P1-BE-04 — Issue↔message link channel project + membership
// ---------------------------------------------------------------------------

func TestCreateIssueMessageLink_RejectsForeignChannel_P1BE04(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewTrackerService(env.repo, nil)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Host", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Chan Proj").
		SetKey("FCP").
		Save(env.ctx)
	require.NoError(t, err)

	foreignCh, err := env.client.Channel.Create().
		SetName("other-dev").
		SetChannelType("public").
		SetProjectID(foreign.ID).
		SetCreatedBy(env.userID).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = svc.CreateIssueMessageLink(env.ctx, env.projectID, issue.IssueKey, env.userID, tracker.CreateIssueMessageLinkRequest{
		ChannelID: foreignCh.ID.String(),
		MessageID: "msg-foreign",
	})
	require.Error(t, err)
	requireAppErrorCode(t, err, 404)
}

func TestCreateIssueMessageLink_RejectsNonMember_P1BE04(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewTrackerService(env.repo, nil)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Host2", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	ch, err := env.client.Channel.Create().
		SetName("members-only").
		SetChannelType("private").
		SetProjectID(env.projectID).
		SetCreatedBy(env.userID).
		Save(env.ctx)
	require.NoError(t, err)

	outsider, err := env.client.User.Create().
		SetEmail("outsider-" + uuid.New().String() + "@test.local").
		SetName("Outsider").
		Save(env.ctx)
	require.NoError(t, err)

	// Канал ∈ project, но outsider не member → 403.
	_, err = svc.CreateIssueMessageLink(env.ctx, env.projectID, issue.IssueKey, outsider.ID, tracker.CreateIssueMessageLinkRequest{
		ChannelID: ch.ID.String(),
		MessageID: "msg-no-member",
	})
	require.Error(t, err)
	requireAppErrorCode(t, err, 403)
}

func TestCreateIssueMessageLink_AllowsProjectMember_P1BE04(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewTrackerService(env.repo, nil)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Host3", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	ch, err := env.client.Channel.Create().
		SetName("ok-chan").
		SetChannelType("public").
		SetProjectID(env.projectID).
		SetCreatedBy(env.userID).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = env.client.ChannelMember.Create().
		SetChannelID(ch.ID).
		SetUserID(env.userID).
		SetRole("member").
		Save(env.ctx)
	require.NoError(t, err)

	resp, err := svc.CreateIssueMessageLink(env.ctx, env.projectID, issue.IssueKey, env.userID, tracker.CreateIssueMessageLinkRequest{
		ChannelID: ch.ID.String(),
		MessageID: "msg-ok",
	})
	require.NoError(t, err)
	require.Equal(t, ch.ID.String(), resp.ChannelID)
}
