package storage_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/internal/storage"
)

// TestGenerateUploadURL_RejectsOversized_P1_BE_38 — size > 25MB → reject.
// Mutation: снять size check → err==nil FAIL.
func TestGenerateUploadURL_RejectsOversized_P1_BE_38(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := newStorageServiceWithPresign(t, client)

	_, err := svc.GenerateUploadURL(ctx, "big.bin", "application/octet-stream", userID, projectID, storage.ExportMaxUploadBytes()+1)
	require.ErrorIs(t, err, storage.ErrFileTooLarge)
}

// TestConfirmUpload_RejectsForeignIssue_P1_BE_38 — confirm чужой issue → reject.
// Mutation: снять entity project check → err==nil FAIL.
func TestConfirmUpload_RejectsForeignIssue_P1_BE_38(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")

	otherProj, err := client.Project.Create().
		SetName("Other").
		SetKey("OTH").
		Save(ctx)
	require.NoError(t, err)

	status, err := client.WorkflowStatus.Create().
		SetName("Open").
		SetCategory(workflowstatus.CategoryTodo).
		SetProjectID(otherProj.ID).
		Save(ctx)
	require.NoError(t, err)

	foreignIssue, err := client.Issue.Create().
		SetIssueNumber(1).
		SetIssueKey("OTH-1").
		SetTitle("Foreign").
		SetIssueType(issue.IssueTypeTask).
		SetProjectID(otherProj.ID).
		SetReporterID(userID).
		SetStatusID(status.ID).
		Save(ctx)
	require.NoError(t, err)

	fileID := uuid.New()
	_, err = client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("a.pdf").
		SetContentType("application/pdf").
		SetSize(10).
		SetS3Key(projectID.String() + "/files/" + fileID.String() + "/a.pdf").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	err = svc.ConfirmUpload(ctx, fileID, "issue", foreignIssue.ID.String())
	require.ErrorIs(t, err, storage.ErrEntityProjectMismatch)
}

// TestValidateEntityScope_IssueSameProject_P1_BE_38 — свой issue → ok.
func TestValidateEntityScope_IssueSameProject_P1_BE_38(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")

	status, err := client.WorkflowStatus.Create().
		SetName("Open").
		SetCategory(workflowstatus.CategoryTodo).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	iss, err := client.Issue.Create().
		SetIssueNumber(1).
		SetIssueKey("STG-1").
		SetTitle("Local").
		SetIssueType(issue.IssueTypeTask).
		SetProjectID(projectID).
		SetReporterID(userID).
		SetStatusID(status.ID).
		Save(ctx)
	require.NoError(t, err)

	err = storage.ExportValidateEntityScope(svc, context.Background(), projectID, "issue", iss.ID.String())
	require.NoError(t, err)

	err = storage.ExportValidateEntityScope(svc, context.Background(), projectID, "issue", "not-a-uuid")
	assert.ErrorIs(t, err, storage.ErrInvalidEntity)
}
