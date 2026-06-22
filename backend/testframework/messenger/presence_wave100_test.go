package messenger_test

// Тесты модуля мессенджера (волна 100): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseStatusDuration_AllVariants проверяет сценарий: разбор status duration all variants.
func TestParseStatusDuration_AllVariants(t *testing.T) {
	assert.Equal(t, 30*time.Minute, messenger.ParseStatusDuration("30m"))
	assert.Equal(t, time.Hour, messenger.ParseStatusDuration("1h"))
	assert.Equal(t, 4*time.Hour, messenger.ParseStatusDuration("4h"))
	assert.Equal(t, 24*time.Hour, messenger.ParseStatusDuration("1d"))
	assert.Equal(t, 7*24*time.Hour, messenger.ParseStatusDuration("1w"))
	assert.Equal(t, 365*24*time.Hour, messenger.ParseStatusDuration("0"))
	assert.Equal(t, 24*time.Hour, messenger.ParseStatusDuration("unknown"))
	assert.GreaterOrEqual(t, messenger.ParseStatusDuration("today"), time.Minute)
}

// TestGetOnlineUserIDs_ReturnsRecentHeartbeats проверяет сценарий: получение online user i ds returns recent heartbeats.
func TestGetOnlineUserIDs_ReturnsRecentHeartbeats(t *testing.T) {
	mr, rdb := newTestRedis(t)
	svc := messenger.NewPresenceService(rdb)
	ctx := context.Background()

	require.NoError(t, svc.Heartbeat(ctx, "online-1"))

	old := time.Now().Add(-10 * time.Minute).Unix()
	mr.ZAdd(messenger.ExportPresenceOnlineKey(), float64(old), "away-user")

	ids, err := svc.GetOnlineUserIDs(ctx)
	require.NoError(t, err)
	assert.Contains(t, ids, "online-1")
	assert.NotContains(t, ids, "away-user")
}

// TestStartCleanupLoop_StopsOnCancel проверяет сценарий: start cleanup loop stops on cancel.
func TestStartCleanupLoop_StopsOnCancel(t *testing.T) {
	_, rdb := newTestRedis(t)
	svc := messenger.NewPresenceService(rdb)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		svc.StartCleanupLoop(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup loop did not stop")
	}
}

// TestSetCustomStatus_WithDurationTTL проверяет сценарий: set custom status with duration ttl.
func TestSetCustomStatus_WithDurationTTL(t *testing.T) {
	_, rdb := newTestRedis(t)
	svc := messenger.NewPresenceService(rdb)
	ctx := context.Background()

	require.NoError(t, svc.SetCustomStatus(ctx, "user-ttl", "Focus mode", "30m"))
	text, err := svc.GetCustomStatus(ctx, "user-ttl")
	require.NoError(t, err)
	assert.Equal(t, "Focus mode", text)
}

// TestHookTestRepo_SetupAndInsertMessage проверяет сценарий: hook test repo setup and insert message.
func TestHookTestRepo_SetupAndInsertMessage(t *testing.T) {
	fixture := messenger.NewHookTestRepo()
	userID := uuid.New()
	botID := uuid.New()
	chID := fixture.SetupDM(userID, botID)
	msgID := fixture.InsertMessage(chID, userID, "hello bot", "")
	assert.NotEmpty(t, msgID)

	pubID := fixture.SetupPublicChannel(userID, botID)
	assert.NotEqual(t, uuid.Nil, pubID)
}

// TestMessengerService_SetAndGetCustomStatus проверяет бизнес-логику: мессенджер сервис set and получение custom status.
func TestMessengerService_SetAndGetCustomStatus(t *testing.T) {
	_, rdb := newTestRedis(t)
	presence := messenger.NewPresenceService(rdb)
	svc := messenger.NewMessengerService(nil, nil, presence, nil)
	ctx := context.Background()

	require.NoError(t, svc.SetCustomStatus(ctx, "svc-user", "Deep work", "30m"))
	text, err := svc.GetCustomStatus(ctx, "svc-user")
	require.NoError(t, err)
	assert.Equal(t, "Deep work", text)
}
