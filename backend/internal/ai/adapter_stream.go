package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

// StreamDelta — чанк OpenAI-совместимого SSE.
type StreamDelta struct {
	Content string
	Done    bool
}

// ChatStreamAdapter стримит chat completions от провайдера.
type ChatStreamAdapter interface {
	ChatCompletionStream(ctx context.Context, baseURL, apiKey, model string, req ChatCompletionRequest, onDelta func(StreamDelta) error) error
}

// ChatCompletionStream — POST {baseURL}/chat/completions со stream=true (SSE).
func (a *OpenAIAdapter) ChatCompletionStream(ctx context.Context, baseURL, apiKey, model string, req ChatCompletionRequest, onDelta func(StreamDelta) error) error {
	base := strings.TrimRight(baseURL, "/")
	if !a.skipOutboundURLValidate {
		if err := safehttp.ValidateOutboundURL(base); err != nil {
			return fmt.Errorf("invalid provider URL: %w", err)
		}
	}
	req.Model = model
	req.Stream = true
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}
	httpClient := a.client
	if a.streamClient != nil {
		httpClient = a.streamClient
	}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("provider request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("provider HTTP %d: %s", resp.StatusCode, truncate(string(raw), 512))
	}
	scanner := bufio.NewScanner(io.LimitReader(resp.Body, maxResponseBytes))
	scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)
	for scanner.Scan() {
		// W5-A2-004 / P2-BE-17: cancel mid-stream — не ждать следующего Read.
		if err := ctx.Err(); err != nil {
			return err
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return onDelta(StreamDelta{Done: true})
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		content := chunk.Choices[0].Delta.Content
		if content == "" {
			continue
		}
		if err := onDelta(StreamDelta{Content: content}); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stream: %w", err)
	}
	return onDelta(StreamDelta{Done: true})
}
