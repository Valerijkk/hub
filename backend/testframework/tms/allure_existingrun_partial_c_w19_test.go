//go:build hubtest

package tms_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/tms"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// C-W19-TT-002 / TASK-185: Allure ExistingRun — fail-closed на Mongo Store
// и match infra err (не slog-only / не silent continue).
//
// Mutation: вернуть slog-only Store + continue в buildAllureResultInputs → оба теста RED.

// TestTask185_ImportAllureExistingRun_MongoStoreFail_FailClosed —
// StoreAllureResults err → 500, PG results не пишутся.
func TestTask185_ImportAllureExistingRun_MongoStoreFail_FailClosed(t *testing.T) {
	repo := newMockRepo()
	repo.storeAllureErr = errors.New("mongo unavailable")
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "existing"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)
	before := len(repo.results)

	payload, err := json.Marshal([]tms.AllureResult{{
		UUID: "u-mongo", Name: "t", Status: "passed",
	}})
	require.NoError(t, err)

	_, err = svc.ImportAllureToExistingRun(ctx, projectID, runID, userID, tms.ImportAllureRequest{Results: payload})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 500, appErr.Code)
	assert.Contains(t, appErr.Error(), "сохранения allure")
	assert.Equal(t, before, len(repo.results), "PG results must not be written after Mongo fail")
	assert.Empty(t, repo.runResults[runID])
}

// TestTask185_ImportAllureExistingRun_MatchDBErr_FailClosed —
// MatchTestCase infra err (≠ NotFound) → 500, без silent skip / Partial-тихого успеха.
func TestTask185_ImportAllureExistingRun_MatchDBErr_FailClosed(t *testing.T) {
	repo := newMockRepo()
	repo.findMatchErr = errors.New("db timeout")
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "existing-match"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)
	before := len(repo.results)

	payload, err := json.Marshal([]tms.AllureResult{{
		UUID:   "u-match",
		Name:   "LoginTest",
		Status: "passed",
		Labels: []tms.AllureLabel{{Name: "automation_id", Value: "auto-1"}},
	}})
	require.NoError(t, err)

	_, err = svc.ImportAllureToExistingRun(ctx, projectID, runID, userID, tms.ImportAllureRequest{Results: payload})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 500, appErr.Code)
	assert.Contains(t, appErr.Error(), "сопоставления allure")
	assert.Equal(t, before, len(repo.results), "match DB err must not write PG results")
	assert.Empty(t, repo.runResults[runID])
}
