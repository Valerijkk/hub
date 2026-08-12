//go:build hubtest

package tms_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestEntRepository_UpdateTestCase_SoftDeleted_W2A3004 — Update soft-deleted TC → 404.
// Mutation: снять DeletedAtIsNil из UpdateTestCase → conflict/200 → FAIL.
func TestEntRepository_UpdateTestCase_SoftDeleted_W2A3004(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Soft", CaseType: "classic", Priority: "medium", Status: "draft",
		CreatedByID: userID,
	})
	require.NoError(t, err)
	require.NoError(t, repo.DeleteTestCase(ctx, projectID, tc.ID))

	title := "nope"
	_, err = repo.UpdateTestCase(ctx, tc.ID, tms.UpdateTestCaseInput{
		Title: &title, Version: tc.Version,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 404, appErr.Code)
}

// TestEntRepository_FindTestCase_SkipsSoftDeleted_W2A3005 — Find* не видит trash.
// Mutation: снять DeletedAtIsNil из FindTestCaseByExternalID → soft-deleted находится → FAIL.
func TestEntRepository_FindTestCase_SkipsSoftDeleted_W2A3005(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	ext := "ext-" + uuid.NewString()[:8]
	auto := "auto-" + uuid.NewString()[:8]
	full := "pkg.Class#method"
	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Match", CaseType: "classic", Priority: "medium", Status: "draft",
		CreatedByID: userID, ExternalID: ext, AutomationID: auto,
		FullName: full, ClassName: "Class", MethodName: "method",
	})
	require.NoError(t, err)
	require.NoError(t, repo.DeleteTestCase(ctx, projectID, tc.ID))

	_, err = repo.FindTestCaseByExternalID(ctx, projectID, ext)
	require.True(t, ent.IsNotFound(err), "soft-deleted by external_id must be hidden")

	_, err = repo.FindTestCaseByAutomationID(ctx, projectID, auto)
	require.True(t, ent.IsNotFound(err), "soft-deleted by automation_id must be hidden")

	_, err = repo.FindTestCaseByFullName(ctx, projectID, full)
	require.True(t, ent.IsNotFound(err), "soft-deleted by full_name must be hidden")

	_, err = repo.FindTestCaseByClassMethod(ctx, projectID, "Class", "method")
	require.True(t, ent.IsNotFound(err), "soft-deleted by class/method must be hidden")
}
