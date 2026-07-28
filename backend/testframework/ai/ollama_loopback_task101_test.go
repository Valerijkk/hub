//go:build hubtest

package ai_test

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/ai"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

// TestOpenAIAdapter_OfflineLoopback_ConsentDial_P1BEU01 — Ollama на 127.0.0.1
// проходит ValidateOfflineAIURL + offline dialer (не SafeDial block).
// Мутация: убрать offlineClient / всегда ValidateOutboundURL → RED.
func TestOpenAIAdapter_OfflineLoopback_ConsentDial_P1BEU01(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "pong"}},
			},
		})
	})
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()

	base := "http://" + ln.Addr().String() + "/v1"
	require.NoError(t, safehttp.ValidateOfflineAIURL(base))
	require.ErrorIs(t, safehttp.ValidateOutboundURL(base), safehttp.ErrBlockedAddress,
		"глобальный SSRF должен по-прежнему блокировать loopback литерал")

	a := ai.NewOpenAIAdapter()
	resp, err := a.ChatCompletion(context.Background(), base, "", "llama3", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Choices)
	assert.Equal(t, "pong", resp.Choices[0].Message.Content)
}

// TestOpenAIAdapter_CloudURL_StillUsesSafeDial_P1BEU01 — metadata IP остаётся на SafeDial path.
func TestOpenAIAdapter_CloudURL_StillUsesSafeDial_P1BEU01(t *testing.T) {
	a := ai.NewOpenAIAdapter()
	_, err := a.ChatCompletion(context.Background(), "http://169.254.169.254/", "", "m", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid provider URL")
}

func TestOpenAIAdapter_LoopbackRejectedByOutboundOnlyPath(t *testing.T) {
	require.ErrorIs(t, safehttp.ValidateOutboundURL("http://127.0.0.1:11434/v1"), safehttp.ErrBlockedAddress)
}

func TestOpenAIAdapter_OfflineStream_ConsentDial_P1BEU01(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n")
	})
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()

	base := "http://" + ln.Addr().String() + "/v1"
	a := ai.NewOpenAIAdapter()
	var got string
	err = a.ChatCompletionStream(context.Background(), base, "", "llama3", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	}, func(d ai.StreamDelta) error {
		got += d.Content
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", got)
}
