//go:build hubtest

package search_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/search"
)

// TestSearch_PartialFlagOnShardError — W5-A4-007 / P2-BE-18.
// Без membership: issues shard no-op OK; user shard падает → partial.
func TestSearch_PartialFlagOnShardError(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()

	repo := &mockSearchRepo{adminErr: errors.New("admin lookup failed")}
	svc := search.NewSearchService(repo, db, nil)

	results, partial, err := svc.Search(context.Background(), "hub", userID, []string{"issue", "user"}, 20)
	require.NoError(t, err)
	assert.True(t, partial)
	assert.Empty(t, results)
}
