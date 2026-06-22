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

const yandexCompletionURL = "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"

// YandexGPTAdapter — адаптер Yandex Foundation Models (AI-2).
type YandexGPTAdapter struct {
	client *http.Client
}

// NewYandexGPTAdapter — адаптер Yandex GPT.
func NewYandexGPTAdapter() *YandexGPTAdapter {
	return &YandexGPTAdapter{client: safehttp.NewSafeHTTPClient(90 * time.Second)}
}

func parseYandexCredentials(apiKey string) (folderID, key string, ok bool) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return "", "", false
	}
	parts := strings.SplitN(apiKey, ":", 2)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], true
	}
	return "", apiKey, true
}

type yandexRequest struct {
	ModelURI          string                 `json:"modelUri"`
	CompletionOptions map[string]interface{} `json:"completionOptions"`
	Messages          []yandexMessage        `json:"messages"`
}

type yandexMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type yandexResponse struct {
	Result struct {
		Alternatives []struct {
			Message struct {
				Role string `json:"role"`
				Text string `json:"text"`
			} `json:"message"`
		} `json:"alternatives"`
		Usage struct {
			InputTextTokens  string `json:"inputTextTokens"`
			OutputTextTokens string `json:"outputTextTokens"`
		} `json:"usage"`
	} `json:"result"`
}

// ChatCompletion — обращение к Yandex Foundation Models API.
func (a *YandexGPTAdapter) ChatCompletion(ctx context.Context, _, apiKey, model string, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	folderID, key, ok := parseYandexCredentials(apiKey)
	if !ok {
		return nil, fmt.Errorf("yandexgpt: api_key required (folder_id:api_key)")
	}
	if err := safehttp.ValidateOutboundURL(yandexCompletionURL); err != nil {
		return nil, err
	}
	modelURI := model
	if !strings.HasPrefix(modelURI, "gpt://") {
		if model == "" {
			model = "yandexgpt/latest"
		}
		modelURI = fmt.Sprintf("gpt://%s/%s", folderID, model)
	}
	msgs := make([]yandexMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m.Role == "system" {
			msgs = append(msgs, yandexMessage{Role: "system", Text: m.Content})
			continue
		}
		role := "user"
		if m.Role == "assistant" {
			role = "assistant"
		}
		msgs = append(msgs, yandexMessage{Role: role, Text: m.Content})
	}
	body, err := json.Marshal(yandexRequest{
		ModelURI: modelURI,
		CompletionOptions: map[string]interface{}{
			"stream":      false,
			"temperature": 0.6,
			"maxTokens":   4000,
		},
		Messages: msgs,
	})
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, yandexCompletionURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Api-Key "+key)
	httpReq.Header.Set("x-folder-id", folderID)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("yandexgpt request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("yandexgpt HTTP %d: %s", resp.StatusCode, truncate(string(raw), 512))
	}
	var yr yandexResponse
	if err := json.Unmarshal(raw, &yr); err != nil {
		return nil, fmt.Errorf("yandexgpt decode: %w", err)
	}
	text := ""
	if len(yr.Result.Alternatives) > 0 {
		text = strings.TrimSpace(yr.Result.Alternatives[0].Message.Text)
	}
	out := &ChatCompletionResponse{Model: modelURI}
	out.Choices = []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	}{{Message: ChatMessage{Role: "assistant", Content: text}, FinishReason: "stop"}}
	return out, nil
}

// ChatCompletionStream оборачивает non-streaming completion.
func (a *YandexGPTAdapter) ChatCompletionStream(ctx context.Context, baseURL, apiKey, model string, req ChatCompletionRequest, onDelta func(StreamDelta) error) error {
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

// IsYandexProvider определяет Yandex GPT endpoints.
func IsYandexProvider(baseURL string) bool {
	return strings.Contains(strings.ToLower(baseURL), "llm.api.cloud.yandex.net") ||
		strings.Contains(strings.ToLower(baseURL), "yandex")
}
