package auth_test

// Тесты сервиса аутентификации: интеграционные и unit-сценарии testframework.
import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
)

// TestUpsertSystemSetting_ConcurrentCreate гарантирует отсутствие паники/гонок при upsert system setting конкурентный доступ создание.
func TestUpsertSystemSetting_ConcurrentCreate(t *testing.T) {
	svc, ctx := newRegistrationTestEnv(t)

	var wg sync.WaitGroup
	errs := make([]error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errs[idx] = svc.SetRegistrationMode(ctx, "open")
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

// TestCreateFirstOrMemberUser_SecondUserIsMember проверяет сценарий: создание first or участник user second user is участник.
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
