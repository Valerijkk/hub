//go:build hubtest

package messenger_test

// TASK-190 / AUDIT-TM-003 — FindMessageByID / UpdateMessage не обслуживают trash.
import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
	"github.com/valerii/hub/backend/testframework/mongotest"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// TestFindMessageByID_SkipsSoftDeleted_AUDIT_TM_003
// Mutation: убрать deleted_at:nil из FindMessageByID → soft-deleted находится → FAIL.
func TestFindMessageByID_SkipsSoftDeleted_AUDIT_TM_003(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_tm003_")
	repo := messenger.NewMongoRepository(db)
	ctx := context.Background()
	require.NoError(t, repo.EnsureIndexes(ctx))

	msgID, err := repo.InsertMessage(ctx, &messenger.MessageDocument{
		ChannelID:   uuid.New(),
		SenderID:    uuid.New(),
		ContentJSON: map[string]interface{}{"text": "trash-me"},
		ContentText: "trash-me",
		CreatedAt:   time.Now().UTC(),
	})
	require.NoError(t, err)
	require.NoError(t, repo.SoftDeleteMessage(ctx, msgID))

	_, err = repo.FindMessageByID(ctx, msgID)
	require.ErrorIs(t, err, mongo.ErrNoDocuments)

	_, err = repo.UpdateMessage(ctx, msgID, map[string]interface{}{"text": "nope"}, "nope")
	require.ErrorIs(t, err, mongo.ErrNoDocuments)
}
