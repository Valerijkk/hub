package audit

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ErrMongoMessageEraserUnavailable — Mongo eraser без коллекции (fail-closed, W18-C5-001).
var ErrMongoMessageEraserUnavailable = errors.New("mongo message eraser unavailable")

// MongoMessageEraser стирает messenger-сообщения пользователя по sender_id (P1-BE-34).
type MongoMessageEraser struct {
	coll *mongo.Collection
}

// NewMongoMessageEraser создаёт eraser для коллекции messages.
// db == nil → eraser с nil coll; DeleteMessagesBySender вернёт error (не silent success).
func NewMongoMessageEraser(db *mongo.Database) *MongoMessageEraser {
	if db == nil {
		return &MongoMessageEraser{}
	}
	return &MongoMessageEraser{coll: db.Collection("messages")}
}

// DeleteMessagesBySender удаляет документы messages с данным sender_id.
// W18-C5-001: nil eraser/coll → error (GDPR cascade не считает шаг успешным).
// Идемпотентен: повторный вызов (sync + JetStream retry C-W19-002) безопасен.
func (e *MongoMessageEraser) DeleteMessagesBySender(ctx context.Context, senderID uuid.UUID) (int64, error) {
	if e == nil || e.coll == nil {
		return 0, ErrMongoMessageEraserUnavailable
	}
	res, err := e.coll.DeleteMany(ctx, bson.M{"sender_id": senderID})
	if err != nil {
		return 0, fmt.Errorf("delete messages by sender: %w", err)
	}
	return res.DeletedCount, nil
}
