package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

const maxResponseBytes = 4 << 20 // 4 MiB

// ChatAdapter вызывает OpenAI-compatible chat completions.
type ChatAdapter interface {
	ChatCompletion(ctx context.Context, baseURL, apiKey, model string, req ChatCompletionRequest) (*ChatCompletionResponse, error)
	ChatStreamAdapter
}

// OpenAIAdapter — DeepSeek, Ollama, OpenRouter и др.
type OpenAIAdapter struct {
	client *http.Client
	// streamClient — SSE без Client.Timeout на весь запрос (TASK-085 / W5-A2-003).
	// nil → fallback на client.
	streamClient *http.Client
	// offlineClient — loopback-only dialer (ADR-AI-OLLAMA-LOOPBACK / P1-BE-U01).
	// nil в hubtest с ExportNewOpenAIAdapterWithClient (custom transport).
	offlineClient *http.Client
	// skipOutboundURLValidate — только hubtest (ExportNewOpenAIAdapterWithClient + httptest).
	// Production NewOpenAIAdapter всегда валидирует BaseURL (SSRF).
	skipOutboundURLValidate bool
}

// NewOpenAIAdapter — adapter с SSRF-safe HTTP client + offline loopback client.
func NewOpenAIAdapter() *OpenAIAdapter {
	return &OpenAIAdapter{
		client:        safehttp.NewSafeHTTPClient(90 * time.Second),
		streamClient:  newSafeStreamHTTPClient(),
		offlineClient: safehttp.NewOfflineAIHTTPClient(90 * time.Second),
	}
}

// resolveHTTPClient — cloud BYOK → SafeDial; Ollama loopback → AllowLoopback
// только после ValidateOfflineAIURL (явный consent-path, не ослабляет SSRF).
func (a *OpenAIAdapter) resolveHTTPClient(base string) (*http.Client, error) {
	if a.skipOutboundURLValidate {
		return a.client, nil
	}
	if err := safehttp.ValidateOfflineAIURL(base); err == nil {
		if a.offlineClient == nil {
			return nil, errors.New("offline AI HTTP client not configured")
		}
		return a.offlineClient, nil
	}
	if err := safehttp.ValidateOutboundURL(base); err != nil {
		return nil, fmt.Errorf("invalid provider URL: %w", err)
	}
	return a.client, nil
}

// resolveStreamHTTPClient — как resolveHTTPClient, но предпочитает streamClient (без whole-request Timeout).
func (a *OpenAIAdapter) resolveStreamHTTPClient(base string) (*http.Client, error) {
	c, err := a.resolveHTTPClient(base)
	if err != nil {
		return nil, err
	}
	if a.skipOutboundURLValidate {
		if a.streamClient != nil {
			return a.streamClient, nil
		}
		return c, nil
	}
	// offline / SafeDial path: streamClient только для cloud SafeDial client
	if c == a.client && a.streamClient != nil {
		return a.streamClient, nil
	}
	return c, nil
}

// ChatCompletion POST {baseURL}/chat/completions
func (a *OpenAIAdapter) ChatCompletion(ctx context.Context, baseURL, apiKey, model string, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req.Stream {
		return nil, errors.New("use stream_id with stream=true for WebSocket delivery")
	}
	base := strings.TrimRight(baseURL, "/")
	httpClient, err := a.resolveHTTPClient(base)
	if err != nil {
		return nil, err
	}
	req.Model = model
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	url := base + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("provider request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("provider HTTP %d: %s", resp.StatusCode, truncate(string(raw), 512))
	}
	var out ChatCompletionResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
