package tms_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestEntRepository_UpdateTestCase_OptimisticLock_A3_004 — stale Version на Postgres
// бьёт в EntRepository.UpdateTestCase (VersionEQ), не в мок.
// Mutation: снять testcase.VersionEQ(input.Version) в ent_repository.go → stale update
// проходит (affected>0), тест краснеет.
func TestEntRepository_UpdateTestCase_OptimisticLock_A3_004(t *testing.T) {
	repo, ctx, projectID, userID := newTMSPostgresRepoEnv(t)

	created, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Lock me", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)
	require.Equal(t, 1, created.Version)

	title := "v2"
	updated, err := repo.UpdateTestCase(ctx, created.ID, tms.UpdateTestCaseInput{
		Title: &title, Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "v2", updated.Title)
	assert.Equal(t, 2, updated.Version)

	stale := "should-not-apply"
	_, err = repo.UpdateTestCase(ctx, created.ID, tms.UpdateTestCaseInput{
		Title: &stale, Version: created.Version, // stale: still 1, DB has 2
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 409, appErr.Code)

	got, err := repo.GetTestCase(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "v2", got.Title, "stale write must not change title")
	assert.Equal(t, 2, got.Version)
}
