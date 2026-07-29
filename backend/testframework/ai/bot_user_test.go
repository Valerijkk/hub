package ai_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/ai"
)

func TestEnsureBotUser_CreateAndReuse(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:aibot_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	repo := ai.NewEntBotUserRepository(client)
	id1, err := ai.EnsureBotUser(ctx, repo)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id1)

	id2, err := ai.EnsureBotUser(ctx, repo)
	require.NoError(t, err)
	assert.Equal(t, id1, id2)
}

func TestEnsureBotUser_NilRepo(t *testing.T) {
	_, err := ai.EnsureBotUser(context.Background(), nil)
	require.Error(t, err)
}

// A3W5-3-008: EnsureBotUser не принимает *ent.Client — только BotUserRepository.
func TestEnsureBotUser_LayeringNoEntClient(t *testing.T) {
	var _ ai.BotUserRepository = (*ai.EntBotUserRepository)(nil)
	_, err := ai.EnsureBotUser(context.Background(), ai.NewEntBotUserRepository(nil))
	require.Error(t, err)
}
