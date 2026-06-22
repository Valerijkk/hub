package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRedisProviderRPMLimiter_NilSafe(t *testing.T) {
	var l *ai.RedisProviderRPMLimiter
	require.NoError(t, l.Allow(context.Background(), uuid.New(), 10))

	l = ai.NewRedisProviderRPMLimiter(nil)
	require.NoError(t, l.Allow(context.Background(), uuid.New(), 10))
	require.NoError(t, l.Allow(context.Background(), uuid.New(), 0))

	l = ai.NewRedisProviderRPMLimiter(redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"}))
	require.NoError(t, l.Allow(context.Background(), uuid.New(), 0))
}

func TestNoopProviderRPMLimiter(t *testing.T) {
	var l ai.NoopProviderRPMLimiter
	require.NoError(t, l.Allow(context.Background(), uuid.New(), 100))
}

func TestService_SetProviderRPMLimiter(t *testing.T) {
	svc := ai.NewService(&mockRepo{}, &mockAdapter{}, nil)
	svc.SetProviderRPMLimiter(ai.NoopProviderRPMLimiter{})
}
