package search_test

import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Empty-scope duplicates removed (W5-A5-005) — канон: TestSearchScopedEntities_EmptyProjectIDsReturnNothing.

func TestSearchMessages_NoMongo(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	channelID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO channel_members (channel_id, user_id) VALUES (?, ?)`,
		channelID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.ExportNewSearchServiceFull(nil, db, nil)
	results, err := search.ExportSearchMessages(svc, context.Background(), "hello", userID)
	require.NoError(t, err)
	assert.Nil(t, results)
}
