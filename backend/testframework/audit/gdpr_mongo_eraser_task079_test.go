package audit_test

// TASK-079 / W18-C5-001/002: MongoMessageEraser fail-closed + DeleteMany по sender_id.
import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/valerii/hub/backend/internal/audit"
	"github.com/valerii/hub/backend/testframework/mongotest"
)

// TestMongoMessageEraser_NilColl_FailClosed — W18-C5-001: (0,nil) запрещён.
func TestMongoMessageEraser_NilColl_FailClosed(t *testing.T) {
	e := audit.NewMongoMessageEraser(nil)
	n, err := e.DeleteMessagesBySender(context.Background(), uuid.New())
	require.ErrorIs(t, err, audit.ErrMongoMessageEraserUnavailable)
	assert.Equal(t, int64(0), n)

	var nilEraser *audit.MongoMessageEraser
	n, err = nilEraser.DeleteMessagesBySender(context.Background(), uuid.New())
	require.ErrorIs(t, err, audit.ErrMongoMessageEraserUnavailable)
	assert.Equal(t, int64(0), n)
}

// TestMongoMessageEraser_DeleteBySender_EntMongo — W18-C5-002: реальный DeleteMany.
// Мутация фильтра sender_id→user_id → DeletedCount=0 при живых docs → RED.
func TestMongoMessageEraser_DeleteBySender_EntMongo(t *testing.T) {
	db := mongotest.Connect(t, "gdpr_eraser_")
	coll := db.Collection("messages")
	ctx := context.Background()

	senderA := uuid.New()
	senderB := uuid.New()
	docs := []any{
		bson.M{"sender_id": senderA, "body": "a1"},
		bson.M{"sender_id": senderA, "body": "a2"},
		bson.M{"sender_id": senderB, "body": "b1"},
	}
	_, err := coll.InsertMany(ctx, docs)
	require.NoError(t, err)

	eraser := audit.NewMongoMessageEraser(db)
	deleted, err := eraser.DeleteMessagesBySender(ctx, senderA)
	require.NoError(t, err)
	assert.Equal(t, int64(2), deleted)

	remainA, err := coll.CountDocuments(ctx, bson.M{"sender_id": senderA})
	require.NoError(t, err)
	assert.Equal(t, int64(0), remainA)

	remainB, err := coll.CountDocuments(ctx, bson.M{"sender_id": senderB})
	require.NoError(t, err)
	assert.Equal(t, int64(1), remainB, "чужой sender не должен быть стёрт")
}
