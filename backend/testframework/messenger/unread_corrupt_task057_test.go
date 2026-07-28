//go:build hubtest

package messenger_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
)

// TestGetAll_CorruptCounter_LogsWarn_TASK057 — W3-A3-010 / P3-BE-01.
// Мутация: вернуть `n, _ := strconv.ParseInt` без slog → лог пуст / поле = 0, тест краснеет.
func TestGetAll_CorruptCounter_LogsWarn_TASK057(t *testing.T) {
	mr, rdb := newTestRedis(t)
	svc := messenger.NewUnreadService(rdb)
	ctx := context.Background()

	mr.HSet("unread:user-corrupt", "ch-ok", "3")
	mr.HSet("unread:user-corrupt", "ch-bad", "not-a-number")

	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	counts, err := svc.GetAll(ctx, "user-corrupt")
	require.NoError(t, err)
	assert.Equal(t, int64(3), counts["ch-ok"])
	_, hasBad := counts["ch-bad"]
	assert.False(t, hasBad, "corrupt field must be skipped, not coerced to 0")

	logs := buf.String()
	assert.True(t, strings.Contains(logs, "unread counter corrupt"),
		"expected slog warn on ParseInt err, got: %s", logs)
}
