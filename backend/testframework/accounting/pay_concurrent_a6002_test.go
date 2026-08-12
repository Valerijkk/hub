package accounting_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/journalentry"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/accounting"
)

// postgresTestURL возвращает DSN для интеграционных тестов (локальный compose или CI).
func postgresTestURL(t *testing.T) string {
	t.Helper()
	uri := os.Getenv("POSTGRES_TEST_URL")
	if uri == "" {
		uri = "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := sql.Open("postgres", uri)
	if err == nil {
		err = db.PingContext(ctx)
		_ = db.Close()
	}
	if err != nil {
		fallback := "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
		if uri != fallback {
			db2, err2 := sql.Open("postgres", fallback)
			if err2 == nil && db2.PingContext(ctx) == nil {
				_ = db2.Close()
				return fallback
			}
			if db2 != nil {
				_ = db2.Close()
			}
		}
		t.Fatalf("PostgreSQL недоступен для A6-002 (POSTGRES_TEST_URL=%q): %v", uri, err)
	}
	return uri
}

// openAccountingPostgresEnt создаёт изолированную БД и Ent-клиент на Postgres.
// SQLite сериализует запись и маскирует TOCTOU — concurrent pay проверяем только здесь.
func openAccountingPostgresEnt(t *testing.T) *ent.Client {
	t.Helper()
	adminURI := postgresTestURL(t)

	dbName := "hub_a6002_" + strings.ReplaceAll(uuid.New().String(), "-", "")
	admin, err := sql.Open("postgres", adminURI)
	require.NoError(t, err)
	_, err = admin.Exec("CREATE DATABASE " + dbName)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = admin.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()", dbName)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + dbName)
		_ = admin.Close()
	})

	u, err := url.Parse(adminURI)
	require.NoError(t, err)
	u.Path = "/" + dbName
	dsn := u.String()

	client := enttest.Open(t, "postgres", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// TestMarkInvoicePaid_ConcurrentSingleJournal_A6_002 — два параллельных pay на Postgres:
// ровно одна posted-проводка. Приёмка: снять Where(StatusEQ(Issued)) в PayInvoiceAtomic → 2 проводки → FAIL.
func TestMarkInvoicePaid_ConcurrentSingleJournal_A6_002(t *testing.T) {
	client := openAccountingPostgresEnt(t)
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("cfo-race-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	repo := accounting.NewEntRepository(client)
	svc := accounting.NewService(repo)
	require.NoError(t, repo.EnsureDefaultChart(ctx))

	inv, err := svc.CreateInvoice(ctx, admin.ID, accounting.CreateInvoiceRequest{
		CustomerName: "Acme",
		Description:  "Race pay",
		AmountCents:  10000,
		Currency:     "RUB",
	})
	require.NoError(t, err)
	_, err = svc.IssueInvoice(ctx, admin.ID, inv.ID)
	require.NoError(t, err)

	// Барьер: оба goroutine доходят до MarkInvoicePaid до того, как любой завершит tx.
	var start sync.WaitGroup
	start.Add(2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start.Done()
			start.Wait()
			_, err := svc.MarkInvoicePaid(ctx, admin.ID, inv.ID)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	var okCount, errCount int
	for err := range errs {
		if err == nil {
			okCount++
		} else {
			errCount++
			t.Logf("pay err: %v", err)
		}
	}
	assert.Equal(t, 1, okCount, "ровно один pay должен пройти")
	assert.Equal(t, 1, errCount, "второй pay должен получить ошибку CAS/status")

	posted, err := client.JournalEntry.Query().
		Where(
			journalentry.StatusEQ(journalentry.StatusPosted),
			journalentry.ReferenceEQ(inv.Number),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, posted, fmt.Sprintf("ожидали 1 posted journal для %s, получили %d", inv.Number, posted))

	paid, err := client.Invoice.Get(ctx, inv.ID)
	require.NoError(t, err)
	assert.Equal(t, "paid", string(paid.Status))
}
