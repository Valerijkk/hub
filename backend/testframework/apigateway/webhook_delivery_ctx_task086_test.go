package apigateway_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/apigateway"
)

// TestSendWebhook_ContextCancel_AbortsInFlight — mutation-ready (TASK-086 / A2W4-5-002):
// NewRequest без ctx → Do игнорирует cancel и висит до client Timeout;
// NewRequestWithContext → cancel прерывает Do с context.Canceled.
func TestSendWebhook_ContextCancel_AbortsInFlight(t *testing.T) {
	useUnsafeClient(t)

	entered := make(chan struct{})
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(func() {
		close(release)
		srv.Close()
	})

	d := &apigateway.WebhookDispatcher{}
	ctx, cancel := context.WithCancel(context.Background())
	target := localServerURL(srv.URL)

	errCh := make(chan error, 1)
	go func() {
		_, _, err := apigateway.ExportWebhookSendWebhook(d, ctx, target, []byte(`{"ping":true}`), "sig", "test.event")
		errCh <- err
	}()

	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not receive request")
	}

	cancel()

	select {
	case err := <-errCh:
		require.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) ||
			stringsContainsCanceled(err.Error()),
			"ожидали отмену ctx, получили: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("sendWebhook не прервался по cancel — вероятно NewRequest без ctx (мутация TASK-086)")
	}
}

func stringsContainsCanceled(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "canceled") ||
		strings.Contains(lower, "cancelled") ||
		strings.Contains(lower, "context canceled")
}

// TestSendWebhook_XHubDeliveryHeader — каждая попытка несёт валидный UUID
// в X-Hub-Delivery (TASK-086 / A2W4-5-003).
func TestSendWebhook_XHubDeliveryHeader(t *testing.T) {
	useUnsafeClient(t)

	wantID := uuid.New().String()
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Hub-Delivery")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	d := &apigateway.WebhookDispatcher{}
	_, _, err := apigateway.ExportWebhookSendWebhookWithDelivery(
		d, context.Background(), localServerURL(srv.URL), []byte(`{}`), "sig", "issue.created", wantID)
	require.NoError(t, err)
	assert.Equal(t, wantID, got)
	_, parseErr := uuid.Parse(got)
	require.NoError(t, parseErr, "X-Hub-Delivery must be UUID")
}

// TestDeliverWithRetry_XHubDeliveryUniquePerAttempt — retry получает новый
// delivery id на каждую попытку (не один id на весь job).
func TestDeliverWithRetry_XHubDeliveryUniquePerAttempt(t *testing.T) {
	useUnsafeClient(t)

	var (
		mu   sync.Mutex
		ids  []string
		hits int
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		ids = append(ids, r.Header.Get("X-Hub-Delivery"))
		hits++
		n := hits
		mu.Unlock()
		if n < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	epID := uuid.New()
	repo := &stubWebhookRepo{
		endpoints: []*ent.WebhookEndpoint{{
			ID: epID, URL: localServerURL(srv.URL), Secret: "secret", Events: []string{"*"}, IsActive: true,
		}},
	}
	d := apigateway.NewWebhookDispatcher(repo, nil)

	done := make(chan struct{})
	go func() {
		apigateway.ExportWebhookDispatcherDeliverWithRetry(
			d, context.Background(), repo.endpoints[0], "issue.created",
			[]byte(`{"x":1}`), map[string]interface{}{"x": 1})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("deliverWithRetry timed out")
	}

	mu.Lock()
	defer mu.Unlock()
	require.GreaterOrEqual(t, len(ids), 2, "нужны ≥2 попытки")
	require.NotEmpty(t, ids[0])
	require.NotEmpty(t, ids[1])
	assert.NotEqual(t, ids[0], ids[1], "каждая попытка — свой X-Hub-Delivery")
	_, err0 := uuid.Parse(ids[0])
	_, err1 := uuid.Parse(ids[1])
	require.NoError(t, err0)
	require.NoError(t, err1)
}
