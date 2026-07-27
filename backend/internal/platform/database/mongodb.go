package database

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// NewMongoClient создаёт клиент MongoDB и возвращает указанную базу данных.
func NewMongoClient(ctx context.Context, url, dbName string) (*mongo.Database, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	slog.Info("connected to MongoDB", "database", dbName)
	return client.Database(dbName), nil
}

// MongoDisconnectCleanup возвращает LIFO-cleanup для graceful shutdown (A8-007).
// Регистрировать через addCleanup сразу после успешного NewMongoClient.
func MongoDisconnectCleanup(db *mongo.Database) func() {
	return func() {
		if db == nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.Client().Disconnect(ctx); err != nil {
			slog.Warn("mongo disconnect error", "error", err)
		}
	}
}
