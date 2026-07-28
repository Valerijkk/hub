//go:build hubtest

package audit_test

// TASK-059 / P3-BE-03 / W6-A5-006: после Restore* entity.DeletedAt == nil.
// Mutation: Restore без ClearDeletedAt → assert.Nil краснеет (list-empty мог бы дать ложный PASS).

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/testcase"
)

func TestRestoreIssue_ClearsDeletedAt_W6A5006(t *testing.T) {
	env := newAuditTrashEnv(t)
	iss := env.createIssue(t, nil)
	when := time.Now().UTC()

	require.NoError(t, env.repo.SoftDeleteIssue(env.ctx, iss.ID, when))
	require.NoError(t, env.repo.RestoreIssue(env.ctx, iss.ID))

	got, err := env.client.Issue.Get(env.ctx, iss.ID)
	require.NoError(t, err)
	assert.Nil(t, got.DeletedAt, "RestoreIssue must ClearDeletedAt")
}

func TestRestoreTestCase_ClearsDeletedAt_W6A5006(t *testing.T) {
	env := newAuditTrashEnv(t)
	when := time.Now().UTC()

	tc, err := env.client.TestCase.Create().
		SetProjectID(env.projectID).
		SetCaseNumber(99).
		SetCaseKey("TRH-TC-59").
		SetTitle("Restore assert").
		SetCaseType(testcase.CaseTypeClassic).
		SetCreatedByID(env.userA).
		Save(env.ctx)
	require.NoError(t, err)

	require.NoError(t, env.repo.SoftDeleteTestCase(env.ctx, tc.ID, when))
	require.NoError(t, env.repo.RestoreTestCase(env.ctx, tc.ID))

	got, err := env.client.TestCase.Get(env.ctx, tc.ID)
	require.NoError(t, err)
	assert.Nil(t, got.DeletedAt, "RestoreTestCase must ClearDeletedAt")
}

func TestRestoreFile_ClearsDeletedAt_W6A5006(t *testing.T) {
	env := newAuditTrashEnv(t)
	when := time.Now().UTC()

	file, err := env.client.FileMetadata.Create().
		SetFilename("restore.pdf").
		SetContentType("application/pdf").
		SetSize(42).
		SetS3Key("k-restore").
		SetS3Bucket("b1").
		SetProjectID(env.projectID).
		SetUploaderID(env.userA).
		Save(env.ctx)
	require.NoError(t, err)

	require.NoError(t, env.repo.SoftDeleteFile(env.ctx, file.ID, when))
	require.NoError(t, env.repo.RestoreFile(env.ctx, file.ID))

	got, err := env.client.FileMetadata.Get(env.ctx, file.ID)
	require.NoError(t, err)
	assert.Nil(t, got.DeletedAt, "RestoreFile must ClearDeletedAt")
}
