package env

// Тесты модуля тестового окружения: интеграционные и unit-сценарии testframework.

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// Env — immutable snapshot после Build (Value Object для тестов).
type Env struct {
	T         *testing.T
	Client    *ent.Client
	Ctx       context.Context
	UserID    uuid.UUID
	ProjectID uuid.UUID
	StatusID  uuid.UUID
}

// Builder — fluent builder для Ent sqlite env (Builder pattern, DRY).
type Builder struct {
	t       *testing.T
	prefix  string
	user    *ent.User
	project *ent.Project
	status  *ent.WorkflowStatus
}

// New создаёт Builder, привязанный к t (Helper() вызывается в Build).
func New(t *testing.T, prefix string) *Builder {
	t.Helper()
	if prefix == "" {
		prefix = "tf"
	}
	return &Builder{t: t, prefix: prefix}
}

// WithUser переопределяет поля пользователя по умолчанию до Build.
func (b *Builder) WithUser(email, name string) *Builder {
	b.t.Helper()
	if b.user != nil {
		require.Fail(b.t, "WithUser: user already created; call before Build or use new Builder")
	}
	_ = email
	_ = name
	return b
}

// Build открывает in-memory sqlite Ent-клиент и наполняет user + project + статус по умолчанию.
func (b *Builder) Build() *Env {
	b.t.Helper()
	dsn := "file:" + b.prefix + "_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(b.t, "sqlite3", dsn)
	b.t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()

	user, err := client.User.Create().
		SetEmail(b.prefix + "-" + uuid.New().String()[:8] + "@test.com").
		SetName("Test User").
		Save(ctx)
	require.NoError(b.t, err)

	proj, err := client.Project.Create().
		SetName("Test Project").
		SetKey("TST").
		Save(ctx)
	require.NoError(b.t, err)

	status, err := client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(b.t, err)

	return &Env{
		T:         b.t,
		Client:    client,
		Ctx:       ctx,
		UserID:    user.ID,
		ProjectID: proj.ID,
		StatusID:  status.ID,
	}
}
