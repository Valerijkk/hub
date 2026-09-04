//go:build !release

package dev

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	autoSeedProbeTimeout    = 10 * time.Second
	autoSeedSeedTimeout     = 3 * time.Minute
	autoSeedResponseMaxRead = int64(4096)
)

var autoSeedRetryDelays = [...]time.Duration{
	time.Second,
	2 * time.Second,
	4 * time.Second,
	8 * time.Second,
	15 * time.Second,
	15 * time.Second,
}

// StartAutoSeed schedules the existing dev seed endpoint after the local HTTP
// listener becomes reachable. It is intentionally best-effort: a seed failure
// never changes server readiness or health.
func (h *Handler) StartAutoSeed(ctx context.Context, port int, devMode bool) {
	if !devMode {
		return
	}
	if ctx == nil {
		slog.Warn("dev auto seed skipped", "reason", "missing_root_context")
		return
	}
	if port < 1 || port > 65535 {
		slog.Warn("dev auto seed skipped", "reason", "invalid_port")
		return
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d/api/dev", port)
	go h.runAutoSeed(ctx, baseURL)
}

func (h *Handler) runAutoSeed(ctx context.Context, baseURL string) {
	startedAt := time.Now()
	client := &http.Client{
		Timeout: autoSeedSeedTimeout,
		Transport: &http.Transport{
			Proxy: nil,
		},
	}
	defer client.CloseIdleConnections()

	slog.Info("dev auto seed started")
	for attempt, delay := range autoSeedRetryDelays {
		if !waitAutoSeedDelay(ctx, delay) {
			return
		}

		status, available := autoSeedRequest(ctx, client, http.MethodGet, baseURL+"/status", autoSeedProbeTimeout)
		if !available {
			if logAutoSeedFailure(ctx, attempt+1, "status", status, startedAt) {
				return
			}
			continue
		}

		status, seeded := autoSeedRequest(ctx, client, http.MethodPost, baseURL+"/seed", autoSeedSeedTimeout)
		if seeded {
			slog.Info("dev auto seed ready",
				"attempt", attempt+1,
				"elapsed", time.Since(startedAt).Round(time.Millisecond),
				"status", status,
			)
			return
		}
		if logAutoSeedFailure(ctx, attempt+1, "seed", status, startedAt) {
			return
		}
	}
}

func logAutoSeedFailure(ctx context.Context, attempt int, stage string, status int, startedAt time.Time) bool {
	if ctx.Err() != nil {
		return true
	}
	fields := []any{
		"attempt", attempt,
		"elapsed", time.Since(startedAt).Round(time.Millisecond),
		"stage", stage,
		"status", status,
	}
	if attempt == len(autoSeedRetryDelays) {
		slog.Error("dev auto seed final failed", fields...)
		return false
	}
	slog.Warn("dev auto seed retry", fields...)
	return false
}

func waitAutoSeedDelay(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func autoSeedRequest(ctx context.Context, client *http.Client, method, url string, timeout time.Duration) (int, bool) {
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, method, url, nil)
	if err != nil {
		return 0, false
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, autoSeedResponseMaxRead))

	return resp.StatusCode, resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
}
