package auth_test

// Тесты репозитория аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func newAuthRepoEnv(t *testing.T) (*auth.EntRepository, *ent.Client, context.Context) {
	t.Helper()
	dsn := "file:authrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return auth.NewEntRepository(client), client, context.Background()
}

// TestEntRepository_UserExistsByEmail проверяет слой репозитория: ent репозиторий user exists by email.
func TestEntRepository_UserExistsByEmail(t *testing.T) {
	repo, client, ctx := newAuthRepoEnv(t)
	email := "exists-" + uuid.New().String()[:8] + "@test.com"

	ok, err := repo.UserExistsByEmail(ctx, email)
	require.NoError(t, err)
	assert.False(t, ok)

	_, err = client.User.Create().SetEmail(email).Save(ctx)
	require.NoError(t, err)

	ok, err = repo.UserExistsByEmail(ctx, email)
	require.NoError(t, err)
	assert.True(t, ok)
}

// TestEntRepository_UpsertSystemSetting проверяет слой репозитория: ent репозиторий upsert system setting.
func TestEntRepository_UpsertSystemSetting(t *testing.T) {
	repo, _, ctx := newAuthRepoEnv(t)
	require.NoError(t, repo.UpsertSystemSetting(ctx, "registration_mode", "open", "mode"))
	val, err := repo.GetSystemSettingValue(ctx, "registration_mode")
	require.NoError(t, err)
	assert.Equal(t, "open", val)
	require.NoError(t, repo.UpsertSystemSetting(ctx, "registration_mode", "invite_only", "mode"))
	val, err = repo.GetSystemSettingValue(ctx, "registration_mode")
	require.NoError(t, err)
	assert.Equal(t, "invite_only", val)
}

// TestEntRepository_CreateFirstOrMemberUser проверяет слой репозитория: ent репозиторий создание first or участник user.
func TestEntRepository_CreateFirstOrMemberUser(t *testing.T) {
	repo, _, ctx := newAuthRepoEnv(t)
	email := "founder-" + uuid.New().String()[:8] + "@test.com"
	u, err := repo.CreateFirstOrMemberUser(ctx, func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate {
		return b.SetEmail(email).SetSystemRole(role)
	})
	require.NoError(t, err)
	assert.Equal(t, user.SystemRoleFounder, u.SystemRole)
}

// TestIsNotFound_Auth проверяет сценарий: is not found auth.
func TestIsNotFound_Auth(t *testing.T) {
	repo, _, ctx := newAuthRepoEnv(t)
	_, err := repo.GetUser(ctx, uuid.New())
	assert.True(t, auth.IsNotFound(err))
}
