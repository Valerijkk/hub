package search_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/search"
)

// A2W5-A5-006 / A2W5-A4-003: missing entity tables → total outage error, не empty success.

func TestSearch_AllShardsWithMembership_RecoversFromMissingTables(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	channelID := uuid.New()

	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO channel_members (channel_id, user_id) VALUES (?, ?)`,
		channelID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	// searchAll включает message (nil mongo = OK) — исключаем, чтобы проверить total outage.
	_, _, err = svc.Search(context.Background(), "hub query", userID,
		[]string{"issue", "test_case", "user", "project", "file", "channel"}, 20)
	require.Error(t, err)
	assert.ErrorIs(t, err, search.ErrSearchUnavailable)
}

// A3W5-5-002: дубли closed-DB / nil-mongo сняты —
// каноны: TestGetUserChannelIDs_ClosedDB, TestSearchMessages_NilMongoReturnsEmpty.
