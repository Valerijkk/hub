package auth

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

// TestFirstUserRace проверяет фикс SEC-06: при 100 параллельных регистрациях
// ровно один пользователь должен получить SystemRoleAdmin. Без сериализации
// (mutex или advisory lock) count-then-create race назначил бы admin нескольким.
//
// Использует sqlite in-memory с shared cache, чтобы все горутины видели одну
// и ту же базу. Сериализация регистраций обеспечивается process-level
// `firstUserMu` внутри RegistrationModeService.CreateFirstOrMemberUser.
func TestFirstUserRace(t *testing.T) {
	dsn := "file:firstadminrace_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1&_journal=WAL&_busy_timeout=5000"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	regMode := NewRegistrationModeService(client)
	ctx := context.Background()

	const N = 100
	var wg sync.WaitGroup
	wg.Add(N)
	errs := make(chan error, N)

	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			email := fmt.Sprintf("u%d@example.com", i)
			_, err := regMode.CreateFirstOrMemberUser(ctx, func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate {
				return b.SetEmail(email).SetSystemRole(role)
			})
			if err != nil {
				errs <- fmt.Errorf("worker %d: %w", i, err)
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for e := range errs {
		t.Errorf("registration error: %v", e)
	}

	admins, err := client.User.Query().Where(user.SystemRoleEQ(user.SystemRoleAdmin)).Count(ctx)
	if err != nil {
		t.Fatalf("count admins: %v", err)
	}
	if admins != 1 {
		t.Fatalf("SEC-06: got %d admins, want exactly 1", admins)
	}

	total, err := client.User.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count users: %v", err)
	}
	if total != N {
		t.Fatalf("expected %d users created, got %d", N, total)
	}
}
