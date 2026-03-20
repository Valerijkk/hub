package database

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// NewMongoClient creates a MongoDB client and returns the specified database.
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
