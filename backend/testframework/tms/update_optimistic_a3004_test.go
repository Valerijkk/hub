package tms_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestUpdateTestCase_OptimisticLock_A3_004 — сервисный контракт: stale version → 409.
// Проверяет mockRepository + TMSService (не Ent VersionEQ).
// Prod VersionEQ: TestEntRepository_UpdateTestCase_OptimisticLock_A3_004.
func TestUpdateTestCase_OptimisticLock_A3_004(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, uuid.New(), tms.CreateTestCaseRequest{
		Title: "Lock me", CaseType: "classic", Priority: "medium", Status: "draft",
	})
	require.NoError(t, err)

	title := "v2"
	_, err = svc.UpdateTestCase(ctx, projectID, created.CaseKey, uuid.New(), tms.UpdateTestCaseRequest{
		Title: &title, Version: created.Version,
	})
	require.NoError(t, err)

	stale := "stale"
	_, err = svc.UpdateTestCase(ctx, projectID, created.CaseKey, uuid.New(), tms.UpdateTestCaseRequest{
		Title: &stale, Version: created.Version, // stale
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 409, appErr.Code)
}
