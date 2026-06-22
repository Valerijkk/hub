package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
)

func TestEnsureBotUser_CreateAndReuse(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:aibot_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	id1, err := ai.EnsureBotUser(ctx, client)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id1)

	id2, err := ai.EnsureBotUser(ctx, client)
	require.NoError(t, err)
	assert.Equal(t, id1, id2)
}

func TestEnsureBotUser_NilClient(t *testing.T) {
	_, err := ai.EnsureBotUser(context.Background(), nil)
	require.Error(t, err)
}
