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
	// streamClient — SSE без whole-request Timeout (W5-A2-008 / TASK-085).
	// Если nil (hubtest custom client) — используется client.
	streamClient *http.Client
	// skipOutboundURLValidate — только hubtest (ExportNewOpenAIAdapterWithClient + httptest).
	// Production NewOpenAIAdapter всегда валидирует BaseURL (SSRF).
	skipOutboundURLValidate bool
}

// NewOpenAIAdapter — adapter с SSRF-safe HTTP client.
func NewOpenAIAdapter() *OpenAIAdapter {
	return &OpenAIAdapter{
		client:       safehttp.NewSafeHTTPClient(90 * time.Second),
		streamClient: newSafeStreamHTTPClient(),
	}
}

// ChatCompletion POST {baseURL}/chat/completions
func (a *OpenAIAdapter) ChatCompletion(ctx context.Context, baseURL, apiKey, model string, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req.Stream {
		return nil, errors.New("use stream_id with stream=true for WebSocket delivery")
	}
	base := strings.TrimRight(baseURL, "/")
	if !a.skipOutboundURLValidate {
		if err := safehttp.ValidateOutboundURL(base); err != nil {
			return nil, fmt.Errorf("invalid provider URL: %w", err)
		}
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
	resp, err := a.client.Do(httpReq)
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
