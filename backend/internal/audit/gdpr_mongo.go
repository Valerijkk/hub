package audit

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MongoMessageEraser стирает messenger-сообщения пользователя по sender_id (P1-BE-34).
type MongoMessageEraser struct {
	coll *mongo.Collection
}

// NewMongoMessageEraser создаёт eraser для коллекции messages.
// db может быть nil — тогда eraser no-op (удобно в unit без Mongo).
func NewMongoMessageEraser(db *mongo.Database) *MongoMessageEraser {
	if db == nil {
		return &MongoMessageEraser{}
	}
	return &MongoMessageEraser{coll: db.Collection("messages")}
}

// DeleteMessagesBySender удаляет документы messages с данным sender_id.
func (e *MongoMessageEraser) DeleteMessagesBySender(ctx context.Context, senderID uuid.UUID) (int64, error) {
	if e == nil || e.coll == nil {
		return 0, nil
	}
	res, err := e.coll.DeleteMany(ctx, bson.M{"sender_id": senderID})
	if err != nil {
		return 0, fmt.Errorf("delete messages by sender: %w", err)
	}
	return res.DeletedCount, nil
}
