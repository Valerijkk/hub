//go:build hubtest

package intranet_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/intranet"
)

// TestUpdateNewsPoll_VersionEQ_Task163 — A3W7-2-002: version <1 → 400; stale → 409; VersionEQ.
// Mutation: убрать VersionEQ / снова принимать omit version → omit или stale проходит FAIL.
func TestUpdateNewsPoll_VersionEQ_Task163(t *testing.T) {
	svc, adminID, _ := setupIntranetService(t)
	ctx := context.Background()

	created, err := svc.CreateNews(ctx, adminID, intranet.CreateNewsRequest{
		Title:     "Poll OL",
		Body:      "body",
		Published: true,
		Poll: &intranet.CreatePollPayload{
			Question: "Q?",
			Options:  []string{"A", "B"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, created.Poll)
	require.Equal(t, 1, created.Poll.Version)

	closed := true
	_, err = svc.UpdateNewsPoll(ctx, adminID, created.ID, intranet.UpdateNewsPollRequest{
		Closed:  &closed,
		Version: 0,
	})
	assert.ErrorIs(t, err, intranet.ErrInvalidInput)

	_, err = svc.UpdateNewsPoll(ctx, adminID, created.ID, intranet.UpdateNewsPollRequest{
		Closed: &closed,
		// omit version → 0
	})
	assert.ErrorIs(t, err, intranet.ErrInvalidInput)

	q := "Q2?"
	updated, err := svc.UpdateNewsPoll(ctx, adminID, created.ID, intranet.UpdateNewsPollRequest{
		Question: &q,
		Version:  created.Poll.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Version)
	assert.Equal(t, "Q2?", updated.Question)

	_, err = svc.UpdateNewsPoll(ctx, adminID, created.ID, intranet.UpdateNewsPollRequest{
		Question: &q,
		Version:  created.Poll.Version, // stale
	})
	assert.ErrorIs(t, err, intranet.ErrVersionConflict)
}

// TestUpdateNewsPoll_Concurrent_VersionEQ_Task163 — два writer с одной snapshot-version:
// первый Save побеждает, второй с той же version → 409 (VersionEQ).
// На SQLite не гоняем параллельные tx (table locked); TOCTOU доказывается последовательностью.
// Mutation: убрать VersionEQ → второй writer тоже OK FAIL.
func TestUpdateNewsPoll_Concurrent_VersionEQ_Task163(t *testing.T) {
	svc, adminID, _ := setupIntranetService(t)
	ctx := context.Background()

	created, err := svc.CreateNews(ctx, adminID, intranet.CreateNewsRequest{
		Title:     "Poll race",
		Body:      "body",
		Published: true,
		Poll: &intranet.CreatePollPayload{
			Question: "Race?",
			Options:  []string{"Yes", "No"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, created.Poll)
	ver := created.Poll.Version

	q1 := "winner"
	closed := true
	first, err := svc.UpdateNewsPoll(ctx, adminID, created.ID, intranet.UpdateNewsPollRequest{
		Question: &q1,
		Closed:   &closed,
		Version:  ver,
	})
	require.NoError(t, err)
	require.Equal(t, 2, first.Version)

	q2 := "loser"
	_, err = svc.UpdateNewsPoll(ctx, adminID, created.ID, intranet.UpdateNewsPollRequest{
		Question: &q2,
		Version:  ver, // stale snapshot — как второй concurrent writer
	})
	assert.ErrorIs(t, err, intranet.ErrVersionConflict)

	poll, err := svc.GetNewsPoll(ctx, adminID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, poll.Version)
	assert.Equal(t, "winner", poll.Question)
}
