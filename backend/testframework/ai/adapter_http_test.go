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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func mockHTTPClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func TestOpenAIAdapter_ChatCompletionStream_SSE(t *testing.T) {
	sseBody := "data: {\"choices\":[{\"delta\":{\"content\":\"Hi\"}}]}\n\ndata: [DONE]\n\n"
	client := mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.True(t, strings.HasSuffix(req.URL.Path, "/chat/completions"))
		assert.Equal(t, "Bearer sk-test", req.Header.Get("Authorization"))

		var body ai.ChatCompletionRequest
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		assert.True(t, body.Stream)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(sseBody)),
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		}, nil
	})

	a := ai.ExportNewOpenAIAdapterWithClient(client)
	var chunks []string
	err := a.ChatCompletionStream(context.Background(), "https://api.example.com/v1", "sk-test", "m", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	}, func(d ai.StreamDelta) error {
		if d.Content != "" {
			chunks = append(chunks, d.Content)
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"Hi"}, chunks)
}

func TestOpenAIAdapter_ChatCompletionStream_HTTPError(t *testing.T) {
	client := mockHTTPClient(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(strings.NewReader("upstream down")),
		}, nil
	})
	a := ai.ExportNewOpenAIAdapterWithClient(client)
	err := a.ChatCompletionStream(context.Background(), "https://api.example.com/v1", "", "m", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	}, func(ai.StreamDelta) error { return nil })
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 502")
}

func TestYandexGPTAdapter_ChatCompletion_Success(t *testing.T) {
	payload := `{"result":{"alternatives":[{"message":{"role":"assistant","text":"ответ"}}]}}`
	client := mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, ai.ExportYandexCompletionURL(), req.URL.String())
		assert.Equal(t, "Api-Key secret", req.Header.Get("Authorization"))
		assert.Equal(t, "folder1", req.Header.Get("x-folder-id"))
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(payload)),
		}, nil
	})

	a := ai.ExportNewYandexGPTAdapterWithClient(client)
	resp, err := a.ChatCompletion(context.Background(), "", "folder1:secret", "yandexgpt/latest", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{
			{Role: "system", Content: "sys"},
			{Role: "assistant", Content: "prev"},
			{Role: "user", Content: "q"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "ответ", ai.ExportExtractAssistantText(resp))
}

func TestYandexGPTAdapter_ChatCompletionStream(t *testing.T) {
	payload := `{"result":{"alternatives":[{"message":{"role":"assistant","text":"streamed"}}]}}`
	client := mockHTTPClient(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(payload)),
		}, nil
	})
	a := ai.ExportNewYandexGPTAdapterWithClient(client)
	var got []string
	err := a.ChatCompletionStream(context.Background(), "", "f:key", "yandexgpt/latest", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	}, func(d ai.StreamDelta) error {
		if d.Content != "" {
			got = append(got, d.Content)
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"streamed"}, got)
}

func TestGigaChatAdapter_ChatCompletion_Success(t *testing.T) {
	client := mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case strings.Contains(req.URL.String(), "oauth"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"access_token":"tok","expires_at":0}`)),
			}, nil
		default:
			out, _ := json.Marshal(ai.ChatCompletionResponse{
				Choices: []struct {
					Index        int         `json:"index"`
					Message      ai.ChatMessage `json:"message"`
					FinishReason string      `json:"finish_reason"`
				}{{Message: ai.ChatMessage{Role: "assistant", Content: "giga ok"}}},
			})
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(out)),
			}, nil
		}
	})

	a := ai.ExportNewGigaChatAdapterWithClient(client)
	resp, err := a.ChatCompletion(context.Background(), "https://gigachat.devices.sberbank.ru/api/v1", "id:secret", "GigaChat", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "ping"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "giga ok", ai.ExportExtractAssistantText(resp))
}

func TestGigaChatAdapter_ChatCompletionStream(t *testing.T) {
	client := mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.String(), "oauth") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"access_token":"tok"}`)),
			}, nil
		}
		out, _ := json.Marshal(ai.ChatCompletionResponse{
			Choices: []struct {
				Index        int         `json:"index"`
				Message      ai.ChatMessage `json:"message"`
				FinishReason string      `json:"finish_reason"`
			}{{Message: ai.ChatMessage{Content: "chunk"}}},
		})
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(out))}, nil
	})
	a := ai.ExportNewGigaChatAdapterWithClient(client)
	var chunks []string
	err := a.ChatCompletionStream(context.Background(), "https://gigachat.devices.sberbank.ru/api/v1", "id:secret", "GigaChat", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	}, func(d ai.StreamDelta) error {
		if d.Content != "" {
			chunks = append(chunks, d.Content)
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"chunk"}, chunks)
}

func TestGigaChatAdapter_accessToken_Cached(t *testing.T) {
	calls := 0
	client := mockHTTPClient(func(*http.Request) (*http.Response, error) {
		calls++
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"access_token":"cached"}`)),
		}, nil
	})
	a := ai.ExportNewGigaChatAdapterWithClient(client)
	tok1, err := ai.ExportGigaChatAccessToken(a, context.Background(), "client", "secret")
	require.NoError(t, err)
	tok2, err := ai.ExportGigaChatAccessToken(a, context.Background(), "client", "secret")
	require.NoError(t, err)
	assert.Equal(t, "cached", tok1)
	assert.Equal(t, tok1, tok2)
	assert.Equal(t, 1, calls)
}

func TestParseGigaCredentials(t *testing.T) {
	id, secret, ok := ai.ExportParseGigaCredentials("client:secret")
	assert.True(t, ok)
	assert.Equal(t, "client", id)
	assert.Equal(t, "secret", secret)
	_, _, ok = ai.ExportParseGigaCredentials("bad")
	assert.False(t, ok)
}

func TestAnthropicAdapter_ChatCompletion_Success(t *testing.T) {
	payload := `{"content":[{"type":"text","text":"claude says hi"}],"usage":{"input_tokens":3,"output_tokens":5}}`
	client := mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "sk-ant", req.Header.Get("x-api-key"))
		assert.True(t, strings.HasSuffix(req.URL.Path, "/v1/messages"))
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(payload)),
		}, nil
	})
	a := ai.ExportNewAnthropicAdapterWithClient(client)
	resp, err := a.ChatCompletion(context.Background(), "https://api.anthropic.com", "sk-ant", "claude-3-5-haiku-20241022", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{
			{Role: "system", Content: "be brief"},
			{Role: "user", Content: "hello"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "claude says hi", ai.ExportExtractAssistantText(resp))
	assert.Equal(t, 8, resp.Usage.TotalTokens)
}

func TestAnthropicAdapter_ChatCompletionStream(t *testing.T) {
	payload := `{"content":[{"type":"text","text":"stream text"}],"usage":{"input_tokens":1,"output_tokens":2}}`
	client := mockHTTPClient(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(payload)),
		}, nil
	})
	a := ai.ExportNewAnthropicAdapterWithClient(client)
	var chunks []string
	err := a.ChatCompletionStream(context.Background(), "https://api.anthropic.com", "key", "claude", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	}, func(d ai.StreamDelta) error {
		if d.Content != "" {
			chunks = append(chunks, d.Content)
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"stream text"}, chunks)
}

func TestService_TestProvider_AdapterRouting(t *testing.T) {
	okClient := mockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.String(), "oauth") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"access_token":"tok"}`)),
			}, nil
		}
		if strings.Contains(req.URL.Path, "messages") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"content":[{"type":"text","text":"pong"}]}`)),
			}, nil
		}
		if strings.Contains(req.URL.String(), "yandex") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"result":{"alternatives":[{"message":{"text":"pong"}}]}}`)),
			}, nil
		}
		out, _ := json.Marshal(ai.ChatCompletionResponse{
			Choices: []struct {
				Index        int         `json:"index"`
				Message      ai.ChatMessage `json:"message"`
				FinishReason string      `json:"finish_reason"`
			}{{Message: ai.ChatMessage{Content: "pong"}}},
		})
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(out))}, nil
	})

	cases := []struct {
		name string
		prov *ent.AIProvider
	}{
		{
			name: "gigachat",
			prov: &ent.AIProvider{
				ID: uuid.New(), Name: "Giga", ProviderType: aiprovider.ProviderTypeGigachat,
				BaseURL: "https://gigachat.devices.sberbank.ru/api/v1", APIKeyEnc: "id:secret",
				Models: []string{"GigaChat"}, Enabled: true,
			},
		},
		{
			name: "anthropic",
			prov: &ent.AIProvider{
				ID: uuid.New(), Name: "Claude", ProviderType: aiprovider.ProviderTypeAnthropic,
				BaseURL: "https://api.anthropic.com", APIKeyEnc: "sk-ant",
				Models: []string{"claude-3-5-haiku-20241022"}, Enabled: true,
			},
		},
		{
			name: "yandex",
			prov: &ent.AIProvider{
				ID: uuid.New(), Name: "Yandex", ProviderType: aiprovider.ProviderTypeOpenaiCompatible,
				BaseURL: "https://llm.api.cloud.yandex.net", APIKeyEnc: "folder:key",
				Models: []string{"yandexgpt/latest"}, Enabled: true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockRepo{providers: []*ent.AIProvider{tc.prov}}
			svc := ai.NewService(repo, &mockAdapter{}, nil)
			ai.ExportServiceSetGigaAdapter(svc, ai.ExportNewGigaChatAdapterWithClient(okClient))
			ai.ExportServiceSetAnthropicAdapter(svc, ai.ExportNewAnthropicAdapterWithClient(okClient))
			ai.ExportServiceSetYandexAdapter(svc, ai.ExportNewYandexGPTAdapterWithClient(okClient))

			res, err := svc.TestProvider(context.Background(), tc.prov.ID)
			require.NoError(t, err)
			assert.True(t, res.OK, res.Error)
		})
	}
}
