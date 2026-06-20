package auth

import (
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

func newRegistrationTestEnv(t *testing.T) (*RegistrationModeService, context.Context) {
	t.Helper()
	dsn := "file:regtest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return NewRegistrationModeService(client, nil), context.Background()
}

func TestGetRegistrationMode_DefaultOpen(t *testing.T) {
	svc, ctx := newRegistrationTestEnv(t)
	mode, err := svc.GetRegistrationMode(ctx)
	require.NoError(t, err)
	assert.Equal(t, "open", mode)
}

func TestSetRegistrationMode_Upsert(t *testing.T) {
	svc, ctx := newRegistrationTestEnv(t)
	require.NoError(t, svc.SetRegistrationMode(ctx, "invite_only"))

	mode, err := svc.GetRegistrationMode(ctx)
	require.NoError(t, err)
	assert.Equal(t, "invite_only", mode)
}

func TestCheckRegistrationAllowed(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		wantErr error
	}{
		{"open allows", "open", nil},
		{"invite_only blocks", "invite_only", ErrRegistrationInviteOnly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, ctx := newRegistrationTestEnv(t)
			if tt.mode != "open" {
				require.NoError(t, svc.SetRegistrationMode(ctx, tt.mode))
			}
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

func TestIsFirstUser(t *testing.T) {
	svc, ctx := newRegistrationTestEnv(t)
	assert.True(t, svc.IsFirstUser(ctx))
}

func TestIsFirstUser_AfterUserExists(t *testing.T) {
	svc, ctx := newRegistrationTestEnv(t)
	_, err := svc.CreateFirstOrMemberUser(ctx, func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate {
		return b.SetEmail("first-" + uuid.New().String()[:6] + "@example.com").SetSystemRole(role)
	})
	require.NoError(t, err)
	assert.False(t, svc.IsFirstUser(ctx))
}
