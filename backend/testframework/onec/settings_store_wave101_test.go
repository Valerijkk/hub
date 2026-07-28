package onec_test
import (
	"github.com/valerii/hub/backend/internal/integrations/onec"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
)

func TestEntSettingsStore_GetUpsert(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:onec_set_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	store := onec.NewEntSettingsStore(client)

	val, err := store.Get(ctx, "onec.webhook_url")
	require.NoError(t, err)
	assert.Empty(t, val)

	require.NoError(t, store.Upsert(ctx, "onec.webhook_url", "https://1c.local/hook", "1C webhook"))
	val, err = store.Get(ctx, "onec.webhook_url")
	require.NoError(t, err)
	assert.Equal(t, "https://1c.local/hook", val)

	require.NoError(t, store.Upsert(ctx, "onec.webhook_url", "https://1c.local/v2", "updated"))
	val, err = store.Get(ctx, "onec.webhook_url")
	require.NoError(t, err)
	assert.Equal(t, "https://1c.local/v2", val)
}

func TestEntUserLookup_GetUserEmailPhone(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:onec_user_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	user, err := client.User.Create().
		SetEmail("onec-user@test.com").
		SetName("OneC User").
		SetPhone("+71234567890").
		Save(ctx)
	require.NoError(t, err)

	lookup := onec.NewEntUserLookup(client)
	email, phone, name, err := lookup.GetUserEmailPhone(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "onec-user@test.com", email)
	assert.Equal(t, "+71234567890", phone)
	assert.Equal(t, "OneC User", name)
}

func TestEntUserLookup_NotFound(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:onec_nf_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	lookup := onec.NewEntUserLookup(client)
	_, _, _, err := lookup.GetUserEmailPhone(context.Background(), uuid.New())
	require.Error(t, err)
}

func TestNewEntSettingsStoreAndUserLookup(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:onec_new_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	assert.NotNil(t, onec.NewEntSettingsStore(client))
	assert.NotNil(t, onec.NewEntUserLookup(client))
}
