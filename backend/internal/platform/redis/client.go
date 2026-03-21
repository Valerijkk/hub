package redis

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

// NewClient создаёт Redis-клиент из URL-строки (например, "redis://localhost:6379/0").
func NewClient(ctx context.Context, url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	slog.Info("connected to Redis")
	return client, nil
}
