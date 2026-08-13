package tracker_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/timelog"
	"github.com/valerii/hub/backend/ent/user"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestStartTimer_ConcurrentSingleOpen_A3_005 — два параллельных StartTimer на Postgres:
// ровно один открытый таймер. Приёмка: убрать tx/row-lock в StartTimer → 2 open → FAIL.
// SQLite сериализует запись и маскирует TOCTOU — concurrent только здесь (Postgres).
func TestStartTimer_ConcurrentSingleOpen_A3_005(t *testing.T) {
	env := newTrackerPostgresRepoEnv(t)

	created, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Concurrent timer", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	var start sync.WaitGroup
	start.Add(2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start.Done()
			start.Wait()
			_, err := env.repo.StartTimer(env.ctx, created.ID, env.userID)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	var okCount, conflictCount int
	for err := range errs {
		if err == nil {
			okCount++
			continue
		}
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) && appErr.Code == 409 {
			conflictCount++
			continue
		}
		t.Fatalf("unexpected StartTimer error: %v", err)
	}
	assert.Equal(t, 1, okCount, "ровно один StartTimer должен пройти")
	assert.Equal(t, 1, conflictCount, "второй StartTimer → 409 conflict")

	openCount, err := env.client.TimeLog.Query().
		Where(
			timelog.HasIssueWith(issue.IDEQ(created.ID)),
			timelog.HasUserWith(user.IDEQ(env.userID)),
			timelog.EndedAtIsNil(),
			timelog.StartedAtNotNil(),
		).
		Count(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, openCount, "exactly one open timer must remain after race")
}
