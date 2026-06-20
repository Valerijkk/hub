package calls

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type mongoChatStore struct {
	collection *mongo.Collection
}

func newMongoChatStore(coll *mongo.Collection) *mongoChatStore {
	return &mongoChatStore{collection: coll}
}

func (s *mongoChatStore) ensureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "call_id", Value: 1},
				{Key: "created_at", Value: 1},
			},
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("не удалось создать индексы чата звонков", "error", err)
		return err
	}

	slog.Info("MongoDB индексы для чата звонков созданы")
	return nil
}

func (s *mongoChatStore) insert(ctx context.Context, msg *CallChatMessage) error {
	_, err := s.collection.InsertOne(ctx, msg)
	if err != nil {
		return fmt.Errorf("не удалось сохранить сообщение чата: %w", err)
	}
	return nil
}

func (s *mongoChatStore) listByCallID(ctx context.Context, callID string, limit int) ([]CallChatMessage, error) {
	filter := bson.M{"call_id": callID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить сообщения чата: %w", err)
	}
	defer cursor.Close(ctx)

	var results []CallChatMessage
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("не удалось декодировать сообщения чата: %w", err)
	}
	return results, nil
}

func (s *mongoChatStore) deleteByCallID(ctx context.Context, callID string) error {
	_, err := s.collection.DeleteMany(ctx, bson.M{"call_id": callID})
	if err != nil {
		return fmt.Errorf("не удалось удалить чат звонка: %w", err)
	}
	return nil
}
