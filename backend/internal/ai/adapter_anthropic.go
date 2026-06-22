package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

const anthropicAPIVersion = "2023-06-01"

// AnthropicAdapter — нативный Messages API (AI-6).
type AnthropicAdapter struct {
	client *http.Client
}

// NewAnthropicAdapter — Anthropic adapter с SSRF-safe client.
func NewAnthropicAdapter() *AnthropicAdapter {
	return &AnthropicAdapter{client: safehttp.NewSafeHTTPClient(90 * time.Second)}
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// ChatCompletion POST https://api.anthropic.com/v1/messages
func (a *AnthropicAdapter) ChatCompletion(ctx context.Context, baseURL, apiKey, model string, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		base = "https://api.anthropic.com"
	}
	if err := safehttp.ValidateOutboundURL(base); err != nil {
		return nil, fmt.Errorf("invalid provider URL: %w", err)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("anthropic: api key required")
	}
	system, msgs := splitAnthropicMessages(req.Messages)
	body, err := json.Marshal(anthropicRequest{
		Model:     model,
		MaxTokens: 4096,
		System:    system,
		Messages:  msgs,
	})
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", anthropicAPIVersion)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("anthropic HTTP %d: %s", resp.StatusCode, truncate(string(raw), 512))
	}
	var ar anthropicResponse
	if err := json.Unmarshal(raw, &ar); err != nil {
		return nil, fmt.Errorf("anthropic decode: %w", err)
	}
	text := anthropicText(&ar)
	out := &ChatCompletionResponse{Model: model}
	out.Choices = []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	}{{Message: ChatMessage{Role: "assistant", Content: text}, FinishReason: "stop"}}
	out.Usage.PromptTokens = ar.Usage.InputTokens
	out.Usage.CompletionTokens = ar.Usage.OutputTokens
	out.Usage.TotalTokens = ar.Usage.InputTokens + ar.Usage.OutputTokens
	return out, nil
}

// ChatCompletionStream — полный ответ одним delta (Anthropic SSE отложен).
func (a *AnthropicAdapter) ChatCompletionStream(ctx context.Context, baseURL, apiKey, model string, req ChatCompletionRequest, onDelta func(StreamDelta) error) error {
	resp, err := a.ChatCompletion(ctx, baseURL, apiKey, model, req)
	if err != nil {
		return err
	}
	text := extractAssistantText(resp)
	if text != "" {
		if err := onDelta(StreamDelta{Content: text}); err != nil {
			return err
		}
	}
	return onDelta(StreamDelta{Done: true})
}

func splitAnthropicMessages(msgs []ChatMessage) (system string, out []anthropicMessage) {
	for _, m := range msgs {
		switch m.Role {
		case "system":
			if system == "" {
				system = m.Content
			} else {
				system += "\n" + m.Content
			}
		case "assistant":
			out = append(out, anthropicMessage{Role: "assistant", Content: m.Content})
		default:
			out = append(out, anthropicMessage{Role: "user", Content: m.Content})
		}
	}
	return system, out
}

func anthropicText(ar *anthropicResponse) string {
	var b strings.Builder
	for _, block := range ar.Content {
		if block.Type == "text" {
			b.WriteString(block.Text)
		}
	}
	return strings.TrimSpace(b.String())
}
