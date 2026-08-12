//go:build hubtest

package ai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/internal/ai"
)

// TestRedisProviderRPMLimiter_FailClosedOnRedisErr — мутация: return nil на Redis err → RED.
// A2W5-A2-001 / TASK-085.
func TestRedisProviderRPMLimiter_FailClosedOnRedisErr(t *testing.T) {
	l := ai.NewRedisProviderRPMLimiter(redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"}))
	err := l.Allow(context.Background(), uuid.New(), 10)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ai.ErrProviderRPMUnavailable), "got %v", err)
}

// TestRedisProviderRPMLimiter_MiniredisEnforceAndNoInflate — A2W5-A2-002 / W5-A2-007.
func TestRedisProviderRPMLimiter_MiniredisEnforceAndNoInflate(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	l := ai.NewRedisProviderRPMLimiter(rdb)
	pid := uuid.New()

	require.NoError(t, l.Allow(context.Background(), pid, 2))
	require.NoError(t, l.Allow(context.Background(), pid, 2))
	err = l.Allow(context.Background(), pid, 2)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ai.ErrProviderRateLimited))

	key := "ai:provider:rpm:" + pid.String()
	card, err := rdb.ZCard(context.Background(), key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(2), card, "rejected request must not inflate ZSET")
}

func TestOpenAIAdapter_StreamClientHasNoWholeRequestTimeout(t *testing.T) {
	a := ai.NewOpenAIAdapter()
	sc := ai.ExportOpenAIAdapterStreamClient(a)
	require.NotNil(t, sc)
	assert.Equal(t, time.Duration(0), sc.Timeout, "SSE client must not use Client.Timeout")
	c := ai.ExportOpenAIAdapterClient(a)
	require.NotNil(t, c)
	assert.Equal(t, 90*time.Second, c.Timeout)
}

func TestUnifiedCodeReviewProvider_NilAccessDenies(t *testing.T) {
	// A2W5-A3-005 / TASK-085: access==nil → deny.
	p := ai.NewUnifiedCodeReviewProvider(nil, nil)
	_, err := p.FetchReview(context.Background(), uuid.New(), "https://github.com/acme/app/pull/1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "access checker not configured")
}

func TestHandler_AssistJQL_DailyQuotaExceeded(t *testing.T) {
	// A2W5-A1-001: quota → 429, не 502.
	repo := &mockRepo{
		requestCount: ai.DefaultDailyRequestLimit,
		providers: []*ent.AIProvider{{
			ID: uuid.New(), Name: "System", BaseURL: "http://127.0.0.1:11434/v1",
			Models: []string{"llama3"}, Enabled: true, Scope: aiprovider.ScopeSystem,
		}},
	}
	svc := ai.NewService(repo, &mockAdapter{}, nil)
	h := ai.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", uuid.New().String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	body, err := json.Marshal(ai.JQLAssistRequest{Query: "my open bugs"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/jql/assist", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("Retry-After"))
}

type streamFallbackAdapter struct {
	calls atomic.Int32
	fail  bool
}

func (a *streamFallbackAdapter) ChatCompletion(context.Context, string, string, string, ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
	return &ai.ChatCompletionResponse{}, nil
}

func (a *streamFallbackAdapter) ChatCompletionStream(_ context.Context, _, _, _ string, _ ai.ChatCompletionRequest, onDelta func(ai.StreamDelta) error) error {
	n := a.calls.Add(1)
	if n == 1 {
		if err := onDelta(ai.StreamDelta{Content: "partial"}); err != nil {
			return err
		}
		return errors.New("http 503 upstream")
	}
	a.fail = true
	return onDelta(ai.StreamDelta{Content: "from-fallback", Done: true})
}

func TestStreamCompletionWithFallback_NoSwitchAfterDelta(t *testing.T) {
	// W5-A2-005 / TASK-085: после 1-го delta — no provider switch.
	userID := uuid.New()
	p1 := &ent.AIProvider{
		ID: uuid.New(), Name: "P1", BaseURL: "http://127.0.0.1:11434/v1",
		Models: []string{"m"}, Enabled: true, Scope: aiprovider.ScopeSystem, MaxRpm: 0,
	}
	p2 := &ent.AIProvider{
		ID: uuid.New(), Name: "P2", BaseURL: "http://127.0.0.1:11435/v1",
		Models: []string{"m"}, Enabled: true, Scope: aiprovider.ScopeSystem, MaxRpm: 0,
	}
	repo := &mockRepo{
		providers:     []*ent.AIProvider{p1, p2},
		allowExternal: map[uuid.UUID]bool{userID: true},
	}
	adapter := &streamFallbackAdapter{}
	svc := ai.NewService(repo, adapter, nil)

	var got []string
	_, _, err := ai.ExportServiceStreamCompletionWithFallback(
		svc, context.Background(), userID, "m",
		ai.ChatCompletionRequest{Model: "m", Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}}},
		func(d ai.StreamDelta) error {
			if d.Content != "" {
				got = append(got, d.Content)
			}
			return nil
		},
	)
	require.Error(t, err)
	assert.Equal(t, []string{"partial"}, got)
	assert.Equal(t, int32(1), adapter.calls.Load(), "must not call second provider after delta")
	assert.False(t, adapter.fail)
}
