package auth

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
)

func TestUpsertSystemSetting_ConcurrentCreate(t *testing.T) {
	svc, ctx := newRegistrationTestEnv(t)
	const key = "registration_mode"

	var wg sync.WaitGroup
	errs := make([]error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errs[idx] = svc.upsertSystemSetting(ctx, key, "open", "registration mode")
		}(i)
	}
	wg.Wait()

	for _, err := range errs {
		require.NoError(t, err)
	}
	mode, err := svc.GetRegistrationMode(ctx)
	require.NoError(t, err)
	assert.Equal(t, "open", mode)
}

func TestCreateFirstOrMemberUser_SecondUserIsMember(t *testing.T) {
	svc, ctx := newRegistrationTestEnv(t)
	first, err := svc.CreateFirstOrMemberUser(ctx, func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate {
		return b.SetEmail("first-" + uuid.New().String()[:6] + "@example.com").SetSystemRole(role)
	})
	require.NoError(t, err)
	assert.Equal(t, user.SystemRoleFounder, first.SystemRole)

	second, err := svc.CreateFirstOrMemberUser(ctx, func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate {
		return b.SetEmail("second-" + uuid.New().String()[:6] + "@example.com").SetSystemRole(role)
	})
	require.NoError(t, err)
	assert.Equal(t, user.SystemRoleMember, second.SystemRole)
}
