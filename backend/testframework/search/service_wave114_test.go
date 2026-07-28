package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchProjects_EmptyScope(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	results, err := search.ExportSearchProjects(svc, context.Background(), "hub", nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchFiles_EmptyScope(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	results, err := search.ExportSearchFiles(svc, context.Background(), "doc", []uuid.UUID{})
	require.NoError(t, err)
	assert.Nil(t, results)
}

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

func TestSearchChannels_EmptyScope(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	results, err := search.ExportSearchChannels(svc, context.Background(), "general", nil, uuid.Nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchTestCases_EmptyScope(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	results, err := search.ExportSearchTestCases(svc, context.Background(), "case", nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}
