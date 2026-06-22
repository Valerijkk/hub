package ai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
)

// ErrAllProvidersFailed — все провайдеры fallback chain упали (M3-008).
var ErrAllProvidersFailed = fmt.Errorf("all AI providers failed")

func isRetryableProviderError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrProviderRateLimited) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout"),
		strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "connection reset"),
		strings.Contains(msg, "no such host"),
		strings.Contains(msg, "tls handshake"),
		strings.Contains(msg, "http 429"),
		strings.Contains(msg, "http 500"),
		strings.Contains(msg, "http 502"),
		strings.Contains(msg, "http 503"),
		strings.Contains(msg, "http 504"),
		strings.Contains(msg, "http 520"),
		strings.Contains(msg, "http 522"),
		strings.Contains(msg, "http 524"):
		return true
	}
	return false
}

func isOfflineProvider(p *ent.AIProvider) bool {
	if p == nil {
		return false
	}
	if p.ProviderType == aiprovider.ProviderTypeOllama {
		return true
	}
	base := strings.ToLower(p.BaseURL)
	return strings.Contains(base, "127.0.0.1") ||
		strings.Contains(base, "localhost") ||
		strings.Contains(base, "ollama")
}

func (s *Service) resolveProviders(ctx context.Context, userID uuid.UUID, model string) ([]*ent.AIProvider, error) {
	providers, err := s.repo.ListProvidersForModel(ctx, userID, model)
	if ent.IsNotFound(err) {
		return nil, ErrNoProvider
	}
	if err != nil {
		return nil, err
	}
	allowExternal, err := s.externalAIAllowed(ctx, userID)
	if err != nil {
		return nil, err
	}
	providers = filterProvidersByConsent(providers, allowExternal)
	out := make([]*ent.AIProvider, 0, len(providers))
	for _, p := range providers {
		if !p.Enabled {
			continue
		}
		if s.offlineOnly && !isOfflineProvider(p) {
			continue
		}
		if err := s.rpm.Allow(ctx, p.ID, p.MaxRpm); err != nil {
			if errors.Is(err, ErrProviderRateLimited) {
				slog.WarnContext(ctx, "ai provider rpm limited, skipping", "provider", p.Name)
				continue
			}
			return nil, err
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		if err := consentErrorForModel(model, allowExternal, nil); err != nil {
			return nil, err
		}
		if s.offlineOnly {
			return nil, fmt.Errorf("no offline AI provider for model %q (AI_OFFLINE_ONLY)", model)
		}
		return nil, ErrNoProvider
	}
	return out, nil
}

func (s *Service) chatCompletionWithFallback(
	ctx context.Context,
	userID uuid.UUID,
	model string,
	req ChatCompletionRequest,
) (*ChatCompletionResponse, *ent.AIProvider, time.Duration, error) {
	providers, err := s.resolveProviders(ctx, userID, model)
	if err != nil {
		return nil, nil, 0, err
	}
	var lastErr error
	for i, p := range providers {
		start := time.Now()
		resp, err := s.chatAdapterFor(p).ChatCompletion(ctx, p.BaseURL, p.APIKeyEnc, model, req)
		latency := time.Since(start)
		if err == nil {
			if i > 0 {
				slog.InfoContext(ctx, "ai fallback succeeded", "provider", p.Name, "model", model, "attempt", i+1)
			}
			return resp, p, latency, nil
		}
		lastErr = err
		if !isRetryableProviderError(err) {
			return nil, nil, latency, err
		}
		slog.WarnContext(ctx, "ai provider failed, trying fallback", "provider", p.Name, "model", model, "error", err)
	}
	if lastErr != nil {
		return nil, nil, 0, fmt.Errorf("%w: %v", ErrAllProvidersFailed, lastErr)
	}
	return nil, nil, 0, ErrNoProvider
}

func (s *Service) streamCompletionWithFallback(
	ctx context.Context,
	userID uuid.UUID,
	model string,
	req ChatCompletionRequest,
	onDelta func(StreamDelta) error,
) (*ent.AIProvider, time.Duration, error) {
	providers, err := s.resolveProviders(ctx, userID, model)
	if err != nil {
		return nil, 0, err
	}
	var lastErr error
	for i, p := range providers {
		start := time.Now()
		err := s.chatAdapterFor(p).ChatCompletionStream(ctx, p.BaseURL, p.APIKeyEnc, model, req, onDelta)
		latency := time.Since(start)
		if err == nil {
			if i > 0 {
				slog.InfoContext(ctx, "ai stream fallback succeeded", "provider", p.Name, "model", model, "attempt", i+1)
			}
			return p, latency, nil
		}
		lastErr = err
		if !isRetryableProviderError(err) {
			return nil, latency, err
		}
		slog.WarnContext(ctx, "ai stream provider failed, trying fallback", "provider", p.Name, "model", model, "error", err)
	}
	if lastErr != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrAllProvidersFailed, lastErr)
	}
	return nil, 0, ErrNoProvider
}
