package auth_test

// Тесты сервиса аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func newRegistrationTestEnv(t *testing.T) (*auth.RegistrationModeService, context.Context) {
	t.Helper()
	dsn := "file:regtest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return auth.NewRegistrationModeService(auth.NewEntRepository(client), nil), context.Background()
}

// TestGetRegistrationMode_DefaultInviteOnly проверяет сценарий: получение регистрация mode default приглашение only.
func TestGetRegistrationMode_DefaultInviteOnly(t *testing.T) {
	svc, ctx := newRegistrationTestEnv(t)
	mode, err := svc.GetRegistrationMode(ctx)
	require.NoError(t, err)
	assert.Equal(t, "invite_only", mode)
}

// TestSetRegistrationMode_Upsert проверяет сценарий: set регистрация mode upsert.
func TestSetRegistrationMode_Upsert(t *testing.T) {
	svc, ctx := newRegistrationTestEnv(t)
	require.NoError(t, svc.SetRegistrationMode(ctx, "invite_only"))

	mode, err := svc.GetRegistrationMode(ctx)
	require.NoError(t, err)
	assert.Equal(t, "invite_only", mode)
}

// TestCheckRegistrationAllowed проверяет сценарий: проверка регистрация allowed.
func TestCheckRegistrationAllowed(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		wantErr error
	}{
		{"open allows", "open", nil},
		{"invite_only blocks", "invite_only", auth.ErrRegistrationInviteOnly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, ctx := newRegistrationTestEnv(t)
			require.NoError(t, svc.SetRegistrationMode(ctx, tt.mode))
			err := svc.CheckRegistrationAllowed(ctx)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestCheckRegistrationAllowed_SettingsFail_InviteOnly — W6-A1-003 / P2-BE-10:
// сбой чтения registration_mode → fail-closed invite_only (не open).
func TestCheckRegistrationAllowed_SettingsFail_InviteOnly(t *testing.T) {
	dsn := "file:regfail_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	svc := auth.NewRegistrationModeService(auth.NewEntRepository(client), nil)
	require.NoError(t, client.Close())

	err := svc.CheckRegistrationAllowed(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, auth.ErrRegistrationInviteOnly))
}

// TestIsFirstUser проверяет сценарий: is first user.
func TestIsFirstUser(t *testing.T) {
	svc, ctx := newRegistrationTestEnv(t)
	assert.True(t, svc.IsFirstUser(ctx))
}

// TestIsFirstUser_AfterUserExists проверяет сценарий: is first user after user exists.
func TestIsFirstUser_AfterUserExists(t *testing.T) {
	svc, ctx := newRegistrationTestEnv(t)
	_, err := svc.CreateFirstOrMemberUser(ctx, func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate {
		return b.SetEmail("first-" + uuid.New().String()[:6] + "@example.com").SetSystemRole(role)
	})
	require.NoError(t, err)
	assert.False(t, svc.IsFirstUser(ctx))
}
