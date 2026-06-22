package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
)

func TestEntSettingsStore_GetUpsert(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:aistore_"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	store := ai.NewEntSettingsStore(client)
	ctx := context.Background()

	val, err := store.Get(ctx, ai.ExportSettingBotSystemPrompt())
	require.NoError(t, err)
	assert.Empty(t, val)

	require.NoError(t, store.Upsert(ctx, ai.ExportSettingBotSystemPrompt(), "Be helpful", "bot prompt"))
	val, err = store.Get(ctx, ai.ExportSettingBotSystemPrompt())
	require.NoError(t, err)
	assert.Equal(t, "Be helpful", val)

	require.NoError(t, store.Upsert(ctx, ai.ExportSettingBotSystemPrompt(), "Be concise", "bot prompt"))
	val, err = store.Get(ctx, ai.ExportSettingBotSystemPrompt())
	require.NoError(t, err)
	assert.Equal(t, "Be concise", val)
}
