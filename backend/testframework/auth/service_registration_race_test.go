package auth_test

// Тесты сервиса аутентификации: интеграционные и unit-сценарии testframework.
import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/auth"
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

// TestCreateFirstOrMemberUser_ConcurrentOnlyOneFounder — A3W6-2-002 / SEC-06:
// concurrent first-user path даёт ровно одного Founder; удаление firstUserMu/tx должно краснеть.
func TestCreateFirstOrMemberUser_ConcurrentOnlyOneFounder(t *testing.T) {
	dsn := "file:regtest_founder_race_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	svc := auth.NewRegistrationModeService(auth.NewEntRepository(client), nil)
	ctx := context.Background()

	const n = 20
	start := make(chan struct{})
	var wg sync.WaitGroup
	results := make([]*ent.User, n)
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			email := fmt.Sprintf("founder-race-%d-%s@example.com", idx, uuid.New().String()[:8])
			u, err := svc.CreateFirstOrMemberUser(ctx, func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate {
				return b.SetEmail(email).SetSystemRole(role)
			})
			results[idx] = u
			errs[idx] = err
		}(i)
	}
	close(start)
	wg.Wait()

	founders := 0
	members := 0
	for i := 0; i < n; i++ {
		require.NoError(t, errs[i], "goroutine %d", i)
		require.NotNil(t, results[i], "goroutine %d", i)
		switch results[i].SystemRole {
		case user.SystemRoleFounder:
			founders++
		case user.SystemRoleMember:
			members++
		default:
			t.Fatalf("unexpected system_role %q from goroutine %d", results[i].SystemRole, i)
		}
	}
	assert.Equal(t, 1, founders)
	assert.Equal(t, n-1, members)

	dbFounders, err := client.User.Query().Where(user.SystemRoleEQ(user.SystemRoleFounder)).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, dbFounders)

	total, err := client.User.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, n, total)
}
