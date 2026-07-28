//go:build hubtest

package messenger_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
	"github.com/valerii/hub/backend/testframework/mongotest"
)

// TestAddReaction_ConcurrentFirstReact_NoDuplicate_A2W3_3_004 — два параллельных
// first-react на один emoji: ровно один ReactionEntry, оба userID.
// Mutation: убрать $ne из push-фильтра → 2 entries → FAIL.
func TestAddReaction_ConcurrentFirstReact_NoDuplicate_A2W3_3_004(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_react_race_")
	repo := messenger.NewMongoRepository(db)
	ctx := context.Background()

	msgID, err := repo.InsertMessage(ctx, &messenger.MessageDocument{
		ChannelID:   uuid.New(),
		SenderID:    uuid.New(),
		ContentJSON: map[string]interface{}{"text": "race"},
		ContentText: "race",
		CreatedAt:   time.Now().UTC(),
	})
	require.NoError(t, err)

	userA := uuid.New().String()
	userB := uuid.New().String()
	emoji := "🔥"

	var start sync.WaitGroup
	start.Add(2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, uid := range []string{userA, userB} {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()
			start.Done()
			start.Wait()
			errs <- repo.AddReaction(ctx, msgID, userID, emoji)
		}(uid)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	found, err := repo.FindMessageByID(ctx, msgID)
	require.NoError(t, err)

	var fireEntries int
	var users []string
	for _, r := range found.Reactions {
		if r.Emoji == emoji {
			fireEntries++
			users = r.UserIDs
		}
	}
	assert.Equal(t, 1, fireEntries, "concurrent first-react must not duplicate ReactionEntry")
	assert.ElementsMatch(t, []string{userA, userB}, users)
}
