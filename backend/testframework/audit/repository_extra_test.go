package audit_test

// Тесты репозитория аудита: интеграционные и unit-сценарии testframework.
import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/internal/audit"
)

// TestRepository_SoftDeleteAndRestoreTestCase проверяет слой репозитория: репозиторий soft удаление and restore test case.
func TestRepository_SoftDeleteAndRestoreTestCase(t *testing.T) {
	env := newAuditTrashEnv(t)
	when := time.Now().UTC()

	tc, err := env.client.TestCase.Create().
		SetProjectID(env.projectID).
		SetCaseNumber(1).
		SetCaseKey("TRH-TC-1").
		SetTitle("Deleted case").
		SetCaseType(testcase.CaseTypeClassic).
		SetCreatedByID(env.userA).
		Save(env.ctx)
	require.NoError(t, err)

	require.NoError(t, env.repo.SoftDeleteTestCase(env.ctx, tc.ID, when))

	items, total, err := env.repo.ListDeletedTestCases(env.ctx, &env.projectID, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)

	require.NoError(t, env.repo.RestoreTestCase(env.ctx, tc.ID))

	items, total, err = env.repo.ListDeletedTestCases(env.ctx, &env.projectID, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)

	got, err := env.client.TestCase.Get(env.ctx, tc.ID)
	require.NoError(t, err)
	assert.Nil(t, got.DeletedAt, "A2W6-A5-007")
}

// TestRepository_TestCase_RetentionHidesOlderThan30Days — A2W6-A5-003.
func TestRepository_TestCase_RetentionHidesOlderThan30Days(t *testing.T) {
	env := newAuditTrashEnv(t)
	tc, err := env.client.TestCase.Create().
		SetProjectID(env.projectID).
		SetCaseNumber(7).
		SetCaseKey("TRH-TC-7").
		SetTitle("Old case").
		SetCaseType(testcase.CaseTypeClassic).
		SetCreatedByID(env.userA).
		Save(env.ctx)
	require.NoError(t, err)

	old := time.Now().UTC().Add(-40 * 24 * time.Hour)
	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)
	require.NoError(t, env.repo.SoftDeleteTestCase(env.ctx, tc.ID, old))

	items, total, err := env.repo.ListDeletedTestCases(env.ctx, &env.projectID, cutoff, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

// TestRepository_File_RetentionHidesOlderThan30Days — A2W6-A5-003.
func TestRepository_File_RetentionHidesOlderThan30Days(t *testing.T) {
	env := newAuditTrashEnv(t)
	file, err := env.client.FileMetadata.Create().
		SetFilename("old.pdf").
		SetContentType("application/pdf").
		SetSize(10).
		SetS3Key("k-old").
		SetS3Bucket("b1").
		SetProjectID(env.projectID).
		SetUploaderID(env.userA).
		Save(env.ctx)
	require.NoError(t, err)

	old := time.Now().UTC().Add(-40 * 24 * time.Hour)
	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)
	require.NoError(t, env.repo.SoftDeleteFile(env.ctx, file.ID, old))

	items, total, err := env.repo.ListDeletedFiles(env.ctx, &env.projectID, cutoff, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

// TestRepository_PurgeExpiredTestCases проверяет отклонение при репозиторий очистка просроченный test cases.
func TestRepository_PurgeExpiredTestCases(t *testing.T) {
	env := newAuditTrashEnv(t)
	tc, err := env.client.TestCase.Create().
		SetProjectID(env.projectID).
		SetCaseNumber(2).
		SetCaseKey("TRH-TC-2").
		SetTitle("Purge me").
		SetCaseType(testcase.CaseTypeClassic).
		SetCreatedByID(env.userA).
		Save(env.ctx)
	require.NoError(t, err)

	old := time.Now().UTC().Add(-40 * 24 * time.Hour)
	require.NoError(t, env.repo.SoftDeleteTestCase(env.ctx, tc.ID, old))

	n, err := env.repo.PurgeExpiredTestCases(env.ctx, time.Now().UTC().Add(-30*24*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	_, err = env.client.TestCase.Get(env.ctx, tc.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err), "A2W6-A5-005: hard delete must remove row")
}

// TestRepository_ClearUserAssignments проверяет слой репозитория: репозиторий clear user assignments.
func TestRepository_ClearUserAssignments(t *testing.T) {
	env := newAuditTrashEnv(t)
	env.createIssue(t, &env.userA)

	n, err := env.repo.ClearUserAssignments(env.ctx, env.userA)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

// TestRepository_PurgeExpiredFiles проверяет отклонение при репозиторий очистка просроченный files.
func TestRepository_PurgeExpiredFiles(t *testing.T) {
	env := newAuditTrashEnv(t)
	when := time.Now().UTC()

	file, err := env.client.FileMetadata.Create().
		SetFilename("purge.pdf").
		SetContentType("application/pdf").
		SetSize(100).
		SetS3Key("k2").
		SetS3Bucket("b2").
		SetProjectID(env.projectID).
		SetUploaderID(env.userA).
		Save(env.ctx)
	require.NoError(t, err)

	old := when.Add(-40 * 24 * time.Hour)
	require.NoError(t, env.repo.SoftDeleteFile(env.ctx, file.ID, old))

	n, err := env.repo.PurgeExpiredFiles(env.ctx, when.Add(-30*24*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	_, err = env.client.FileMetadata.Get(env.ctx, file.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err), "A2W6-A5-005: hard delete must remove file")
}

// TestRepository_PurgeExpiredProjects проверяет отклонение при репозиторий очистка просроченный projects.
func TestRepository_PurgeExpiredProjects(t *testing.T) {
	env := newAuditTrashEnv(t)
	when := time.Now().UTC()
	old := when.Add(-40 * 24 * time.Hour)

	emptyProj, err := env.client.Project.Create().
		SetName("Purge Only").
		SetKey("PUR").
		Save(env.ctx)
	require.NoError(t, err)

	require.NoError(t, env.repo.SoftDeleteProject(env.ctx, emptyProj.ID, old))

	n, err := env.repo.PurgeExpiredProjects(env.ctx, when.Add(-30*24*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	_, err = env.client.Project.Get(env.ctx, emptyProj.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err), "A2W6-A5-005: hard delete must remove project")
}

// TestRepository_DeleteUserRelatedData проверяет слой репозитория: репозиторий удаление user related data.
func TestRepository_DeleteUserRelatedData(t *testing.T) {
	env := newAuditTrashEnv(t)
	iss := env.createIssue(t, nil)

	_, err := env.client.Comment.Create().
		SetIssueID(iss.ID).
		SetAuthorID(env.userA).
		SetBodyText("note").
		SetBodyJSON(map[string]any{"type": "doc"}).
		Save(env.ctx)
	require.NoError(t, err)

	cmt, err := env.client.Comment.Query().Only(env.ctx)
	require.NoError(t, err)
	_, err = env.client.Reaction.Create().
		SetCommentID(cmt.ID).
		SetUserID(env.userA).
		SetEmoji("👍").
		Save(env.ctx)
	require.NoError(t, err)

	_, err = env.client.ActivityLog.Create().
		SetIssueID(iss.ID).
		SetUserID(env.userA).
		SetAction("updated").
		Save(env.ctx)
	require.NoError(t, err)

	_, err = env.client.TimeLog.Create().
		SetIssueID(iss.ID).
		SetUserID(env.userA).
		SetMinutes(30).
		Save(env.ctx)
	require.NoError(t, err)

	n, err := env.repo.DeleteUserReactions(env.ctx, env.userA)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	n, err = env.repo.DeleteUserComments(env.ctx, env.userA)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	n, err = env.repo.DeleteUserActivityLogs(env.ctx, env.userA)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	n, err = env.repo.DeleteUserTimeLogs(env.ctx, env.userA)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

// TestAuditService_DeleteUserGDPR_WithReassign проверяет бизнес-логику: audit сервис удаление user gdpr with reassign.
func TestAuditService_DeleteUserGDPR_WithReassign(t *testing.T) {
	env := newAuditTrashEnv(t)
	target := env.userA
	reassign := env.userB
	svc := audit.NewAuditService(env.repo, nil)

	require.NoError(t, svc.DeleteUserGDPR(env.ctx, uuid.New(), target, &reassign))

	_, err := env.client.User.Get(env.ctx, target)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}
