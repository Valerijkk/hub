package ai_test

// TASK-044 / P2-BE-17: cancel mid-stream в ChatCompletionStream (adapter_stream ctx.Err).
import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/ai"
)

// TestChatCompletionStream_CancelMidRead_P2BE17 — после первого delta cancel;
// без ctx.Err() в цикле Scan блокируется на втором чанке ≥5s → RED.
func TestChatCompletionStream_CancelMidRead_P2BE17(t *testing.T) {
	blockSecond := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "no flush", 500)
			return
		}
		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hi\"}}]}\n\n")
		flusher.Flush()
		// Не выходим по r.Context().Done(): HTTP request отвязан от cancel ctx
		// (WithoutCancel), иначе mid-loop ctx.Err() не единственный путь остановки.
		select {
		case <-blockSecond:
			fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"late\"}}]}\n\n")
			flusher.Flush()
		case <-time.After(5 * time.Second):
			fmt.Fprintf(w, "data: [DONE]\n\n")
			flusher.Flush()
		}
	}))
	t.Cleanup(srv.Close)

	adapter := ai.ExportNewOpenAIAdapterWithClient(srv.Client())
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	start := time.Now()
	err := adapter.ChatCompletionStream(ctx, srv.URL, "key", "gpt-test", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "x"}},
	}, func(d ai.StreamDelta) error {
		if d.Content != "" {
			cancel()
		}
		return nil
	})
	elapsed := time.Since(start)

	require.ErrorIs(t, err, context.Canceled)
	require.Less(t, elapsed, 2*time.Second, "ctx.Err() must stop stream before hung second chunk")
	close(blockSecond)
}
