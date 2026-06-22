package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIAdapter_ChatCompletion_Success(t *testing.T) {
	out, _ := json.Marshal(ai.ChatCompletionResponse{
		Choices: []struct {
			Index        int         `json:"index"`
			Message      ai.ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{{Message: ai.ChatMessage{Role: "assistant", Content: "hello from openai"}}},
	})
	client := mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.True(t, strings.HasSuffix(req.URL.Path, "/chat/completions"))
		assert.Equal(t, "Bearer sk-test", req.Header.Get("Authorization"))

		var body ai.ChatCompletionRequest
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		assert.False(t, body.Stream)
		assert.Equal(t, "gpt-4o-mini", body.Model)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(out)),
		}, nil
	})

	a := ai.ExportNewOpenAIAdapterWithClient(client)
	resp, err := a.ChatCompletion(context.Background(), "https://api.example.com/v1", "sk-test", "gpt-4o-mini", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "hello from openai", ai.ExportExtractAssistantText(resp))
}

func TestOpenAIAdapter_ChatCompletion_HTTPError(t *testing.T) {
	client := mockHTTPClient(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       io.NopCloser(strings.NewReader("rate limited")),
		}, nil
	})
	a := ai.ExportNewOpenAIAdapterWithClient(client)
	_, err := a.ChatCompletion(context.Background(), "https://api.example.com/v1", "sk-test", "m", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 429")
}

func TestOpenAIAdapter_ChatCompletion_InvalidJSONResponse(t *testing.T) {
	client := mockHTTPClient(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("not-json")),
		}, nil
	})
	a := ai.ExportNewOpenAIAdapterWithClient(client)
	_, err := a.ChatCompletion(context.Background(), "https://api.example.com/v1", "", "m", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode response")
}
