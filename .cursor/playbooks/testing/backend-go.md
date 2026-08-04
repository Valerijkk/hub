# Backend: unit и integration на Go

> Часть плейбука тестирования. Канон и оглавление — [`../testing.md`](../testing.md).

Канон Hub для Go-тестов: hubtest, фикстуры, race, слои.

---

## 4. Backend unit/integration (Go)

Backend-тесты — фундамент пирамиды Hub (499 `_test.go` файлов, filesystem count
2026-07-26). Из-за ограничения Go на импорт `internal/*` из другого модуля они
физически лежат в `backend/testframework/{domain}/`, но принадлежат модулю
`github.com/valerii/hub/backend`.

### 4.1. Обязательный каркас теста

Каждый backend-тестовый файл начинается со стандартного заголовка:

```go
// Package tracker содержит юнит-тесты для tracker.TrackerService.
package tracker_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/tracker"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)
```

Правила каркаса:

1. **Пакет с суффиксом `_test`** (`tracker_test`, `tms_test`) — black-box тестирование через публичный API пакета. White-box доступ к приватным функциям — только через `export_hooks.go` (см. 4.10).
2. **testify:** `require` для критичных проверок (останавливают тест), `assert` для не критичных (продолжают).
3. **Контекст всегда первым** аргументом сервисов: `ctx := context.Background()`.
4. **UUID через `uuid.New()`**, не строки.

### 4.2. Тег сборки и запуск

Тесты, использующие `export_hooks.go`, требуют тег `hubtest`. Канонический прогон (как в CI job `backend-test`):

```bash
cd backend
go generate ./ent
go vet -tags hubtest ./...
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...
```

| Флаг | Зачем |
|------|-------|
| `-tags hubtest` | активирует `export_hooks.go` (`//go:build hubtest`) |
| `-race` | детектор гонок — обязателен всегда (rule 40) |
| `-count=1` | отключает кэш результатов тестов |
| `-timeout=600s` | верхняя граница на весь прогон |

Один домен: `go test -tags hubtest ./testframework/tracker/...`. Один тест: `go test -tags hubtest -run TestCreateIssue ./testframework/tracker/...`.

### 4.3. Table-driven tests

Основной паттерн Go. Таблица случаев + цикл с `t.Run`. Реальный пример `TestCreateIssue_InvalidUUIDs`:

```go
func TestCreateIssue_InvalidUUIDs(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name string
		req  tracker.CreateIssueRequest
		err  string // подстрока ожидаемой ошибки
	}{
		{
			name: "invalid assignee_id",
			req:  tracker.CreateIssueRequest{Title: "T", IssueType: "task", AssigneeID: "not-uuid"},
			err:  "assignee_id",
		},
		{
			name: "invalid parent_id",
			req:  tracker.CreateIssueRequest{Title: "T", IssueType: "task", ParentID: "bad"},
			err:  "parent_id",
		},
		{
			name: "invalid status_id",
			req:  tracker.CreateIssueRequest{Title: "T", IssueType: "task", StatusID: "bad"},
			err:  "status_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateIssue(ctx, projectID, userID, tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}
```

**Good:** каждый случай — отдельный подтест с осмысленным `name`; при падении видно, какая строка упала. **Bad:** один тест с 10 последовательными `assert` без `t.Run` — при падении первого остальные не проверяются, а имя не говорит, что сломалось.

### 4.4. `require` vs `assert` — когда что

| Ситуация | Использовать | Почему |
|----------|--------------|--------|
| Проверка `err` перед использованием результата | `require.NoError(t, err)` | если ошибка есть, дальше `resp` — nil, будет паника |
| Проверка `resp != nil` перед разыменованием | `require.NotNil(t, resp)` | защита от nil-паники |
| Type assertion (`err.(*apperrors.AppError)`) | `require.True(t, ok)` | без валидного каста дальше некорректно |
| Проверка значения поля | `assert.Equal(t, want, got)` | не критично для продолжения, показать все расхождения |

**Good — require перед разыменованием:**

```go
resp, err := svc.CreateIssue(ctx, projectID, userID, req)
require.NoError(t, err)      // стоп, если ошибка
require.NotNil(t, resp)      // стоп, если nil
assert.Equal(t, "TEST-1", resp.IssueKey)  // дальше безопасно
assert.Equal(t, 1, resp.Version)
```

**Bad — assert перед разыменованием:**

```go
resp, err := svc.CreateIssue(ctx, projectID, userID, req)
assert.NoError(t, err)              // тест продолжится даже при ошибке
assert.Equal(t, "TEST-1", resp.IssueKey)  // паника: resp == nil
```

### 4.5. Проверка доменных ошибок (`apperrors.AppError`)

Hub возвращает типизированные ошибки с HTTP-кодом. Проверяем не текст, а код:

```go
func TestGetIssue_NotFound(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	_, err := svc.GetIssue(ctx, "MISSING-1")
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Equal(t, 404, appErr.Code)
}
```

Матрица кодов, которую нужно покрывать (из `docs/TESTING.md`): 400 (bad request), 401 (unauth), 403 (RBAC), 404 (not found), 409 (version conflict), 429 (rate limit).

### 4.6. Subtests, `t.Parallel`, `t.Helper`, `t.Cleanup`

**`t.Run` — группировка и изоляция подслучаев** (см. 4.3).

**`t.Parallel` — параллельный прогон** независимых тестов. Ускоряет прогон, но требует полной изоляции состояния. В Hub env `Builder` даёт каждому тесту свой in-memory sqlite (уникальный DSN), поэтому параллелизм безопасен:

```go
func TestSomethingIndependent(t *testing.T) {
	t.Parallel() // безопасно: свой sqlite на тест
	e := env.New(t, "trk").Build()
	// ...
}
```

**Bad — `t.Parallel` с общим состоянием:**

```go
var sharedRepo = newMockRepo() // ПЛОХО: общий на все тесты

func TestA(t *testing.T) {
	t.Parallel()
	sharedRepo.CreateIssue(...) // гонка при -race
}
```

**`t.Helper` — помечает вспомогательную функцию**, чтобы стек ошибки указывал на вызывающий тест, а не на хелпер. Реальный пример из env `Builder`:

```go
func New(t *testing.T, prefix string) *Builder {
	t.Helper() // ошибки Build() укажут на тест, а не на эту строку
	if prefix == "" {
		prefix = "tf"
	}
	return &Builder{t: t, prefix: prefix}
}
```

**`t.Cleanup` — освобождение ресурсов** после теста (вместо `defer`, работает и в подтестах). Реальный пример — закрытие sqlite-клиента:

```go
func (b *Builder) Build() *Env {
	b.t.Helper()
	dsn := "file:" + b.prefix + "_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(b.t, "sqlite3", dsn)
	b.t.Cleanup(func() { _ = client.Close() }) // закрытие после теста
	// ...
}
```

### 4.7. Мок репозитория (паттерн `mockRepo`)

Основной способ изолировать сервис от БД в unit-тестах — ручной мок, реализующий интерфейс `Repository`. Пример из `tracker/service_test.go`: `mockRepo` хранит данные в `map` и реализует все методы интерфейса.

```go
// mockRepo реализует tracker.Repository для тестирования TrackerService.
type mockRepo struct {
	issues          map[uuid.UUID]*ent.Issue
	issuesByKey     map[string]*ent.Issue
	statuses        map[uuid.UUID]*ent.WorkflowStatus
	transitions     map[string]*ent.WorkflowTransition // ключ "fromID:toID"
	projectCounters map[uuid.UUID]int
	updateVersionError bool // флаг для эмуляции конфликта версий
}

func newMockRepo() *mockRepo {
	defaultStatusID := uuid.New()
	return &mockRepo{
		issues:      make(map[uuid.UUID]*ent.Issue),
		issuesByKey: make(map[string]*ent.Issue),
		statuses: map[uuid.UUID]*ent.WorkflowStatus{
			defaultStatusID: {ID: defaultStatusID, Name: "To Do", IsDefault: true},
		},
		transitions:     make(map[string]*ent.WorkflowTransition),
		projectCounters: make(map[uuid.UUID]int),
	}
}
```

Метод мока эмулирует поведение реального репозитория, включая ошибки:

```go
func (m *mockRepo) UpdateIssue(_ context.Context, issueID uuid.UUID, input tracker.UpdateIssueInput) (*ent.Issue, error) {
	i, ok := m.issues[issueID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	// Эмуляция оптимистичной блокировки: версия из запроса должна совпасть.
	if m.updateVersionError || i.Version != input.Version {
		return nil, apperrors.NewConflict("version mismatch")
	}
	if input.Title != nil {
		i.Title = *input.Title
	}
	i.Version++
	return i, nil
}
```

**Хелпер сборки сервиса** держим рядом:

```go
func newTestService() (*tracker.TrackerService, *mockRepo) {
	repo := newMockRepo()
	svc := tracker.NewTrackerService(repo, nil)
	return svc, repo
}
```

**Частичный мок через embedded interface** (для больших интерфейсов) — паттерн из `tms/allure_parser_test.go`. Встраиваем интерфейс и переопределяем только нужные методы; остальные дают nil-заглушки:

```go
type mockMatcherRepo struct {
	tms.Repository // embedded: неиспользуемые методы — заглушки
	externalIDMap  map[string]*ent.TestCase
	fullNameMap    map[string]*ent.TestCase
}

func (m *mockMatcherRepo) FindTestCaseByExternalID(_ context.Context, _ uuid.UUID, id string) (*ent.TestCase, error) {
	if tc, ok := m.externalIDMap[id]; ok {
		return tc, nil
	}
	return nil, &ent.NotFoundError{}
}
```

**Good:** мок эмулирует контракт (not found → `&ent.NotFoundError{}`, конфликт → `apperrors.NewConflict`). **Bad:** мок всегда возвращает `nil, nil` — тест «зелёный», но не проверяет реальное поведение сервиса на ошибках.

### 4.8. Env `Builder` + `Factory` (integration на реальной схеме)

Когда нужна реальная схема Ent (каскады, индексы, реальные запросы) — используем in-memory sqlite через env `Builder`. Он готовит user + project + дефолтный статус:

```go
func TestBoardData_RealSchema(t *testing.T) {
	e := env.New(t, "board").Build() // sqlite + user + project + status
	repo := tracker.NewRepository(e.Client)
	svc := tracker.NewTrackerService(repo, nil)

	_, err := svc.CreateIssue(e.Ctx, e.ProjectID, e.UserID, tracker.CreateIssueRequest{
		Title: "Real", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)
}
```

**Factory — создание дополнительных сущностей** (паттерн из `factory/user.go`):

```go
func TestMultipleUsers(t *testing.T) {
	e := env.New(t, "usr").Build()
	uf := factory.NewUserFactory(t, e.Client, e.Ctx)

	alice := uf.Create("alice") // уникальный email через uuid
	bob := uf.Create("bob")
	require.NotEqual(t, alice.ID, bob.ID)
}
```

Factory гарантирует уникальность (`email = prefix + "-" + uuid[:8] + "@test.com"`), поэтому тесты не конфликтуют по unique-констрейнтам.

### 4.9. Мок vs env.Builder (testcontainers) — выбор

| Критерий | Мок (`mockRepo`) | env.Builder (sqlite) | testcontainers (Postgres/Mongo) |
|----------|------------------|----------------------|---------------------------------|
| Скорость | мкс | мс | секунды (старт контейнера) |
| Что проверяет | логику сервиса | логику + реальную схему Ent | + реальные особенности PG/Mongo |
| Когда | большинство unit | integration слоёв | только специфика PG/Mongo (JSONB, полнотекст, транзакции) |
| CI-стоимость | нулевая | низкая | высокая (тянет образы) |

**Правило Hub:** по умолчанию — мок. Нужна реальная схема — `env.Builder` (sqlite). testcontainers/реальный Mongo — только в CI job (Mongo 8 сервис) и только там, где sqlite неадекватен (реальные Mongo-репозитории `*_mongo_test.go`). Не тащить Docker в каждый unit-тест — это ломает быстрый локальный цикл и жрёт лимит GitHub Actions.

### 4.10. White-box доступ через `export_hooks.go`

Приватные функции пакета нельзя вызвать из `tracker_test`. Для их тестирования пакет `internal/tracker` содержит `export_hooks.go` с тегом `//go:build hubtest`, экспортирующий приватные символы:

```go
//go:build hubtest

package tracker

// ExportBuildFieldChanges строит список изменений полей задачи для hubtest.
func ExportBuildFieldChanges(existing *ent.Issue, req UpdateIssueRequest) []FieldChange {
	return buildFieldChanges(existing, req)
}

// ExportTrackerServiceWorkflow возвращает workflow-движок для hubtest.
func ExportTrackerServiceWorkflow(s *TrackerService) *WorkflowEngine {
	return s.workflow
}
```

Тест вызывает экспортированную обёртку:

```go
func TestWorkflowValidation(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	defaultID := repo.getDefaultStatusID()
	inProgressID := repo.addStatusWithTransition("In Progress", workflowstatus.CategoryInProgress, defaultID)

	// Доступ к приватному workflow-движку через export hook.
	err := tracker.ExportTrackerServiceWorkflow(svc).ValidateTransition(ctx, defaultID, inProgressID)
	assert.NoError(t, err)

	err = tracker.ExportTrackerServiceWorkflow(svc).ValidateTransition(ctx, inProgressID, defaultID)
	assert.Error(t, err) // обратный переход не определён
}
```

**Правила `export_hooks.go`:**
- Обязательно `//go:build hubtest` — в prod-сборке (`-tags release`) этот файл не компилируется.
- Именование `Export<PrivateName>` — единый префикс.
- Комментарий на русском с суффиксом «для hubtest».
- Экспортировать минимально необходимое, не публичный API «на всякий».

### 4.11. HTTP handler-тесты (httptest)

Handler-слой тестируется через `httptest`/Fiber test app: проверяем статус-коды и тело. Матрица ожидаемых кодов — в `docs/TESTING.md` (раздел «API-тесты по модулям»). Пример логики:

```go
func TestCreateIssueHandler_Unauthorized(t *testing.T) {
	app := setupTestApp(t) // Fiber app с роутами tracker
	req := httptest.NewRequest("POST", "/api/projects/"+projID+"/issues",
		strings.NewReader(`{"title":"X","issue_type":"task"}`))
	req.Header.Set("Content-Type", "application/json")
	// без Authorization → ожидаем 401

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}
```

### 4.12. Покрытие и `-race`

```bash
# Профиль ТОЛЬКО в TMP (rule 45), никогда в backend/
go test -tags hubtest -race -coverprofile="$TMPDIR/hub-cov.out" ./testframework/...
go tool cover -func="$TMPDIR/hub-cov.out" | tail -1  # total
```

`-race` обязателен: Hub активно использует горутины (ws, nats, outbox). Гонки ловятся только с ним. Coverage-детали — раздел 14.

### 4.13. Чек-лист backend-теста

- [ ] Пакет `<domain>_test`, импорты сгруппированы, testify подключён.
- [ ] Прогон зелёный с `-tags hubtest -race -count=1`.
- [ ] `require` перед разыменованием, `assert` для значений.
- [ ] Доменные ошибки проверяются по `AppError.Code`, не по тексту.
- [ ] Table-driven для однотипных случаев, `t.Run` с осмысленными именами.
- [ ] Мок эмулирует контракт (ошибки, not found), а не `nil, nil`.
- [ ] Состояние изолировано (свой мок/sqlite на тест); `t.Parallel` только при изоляции.
- [ ] `t.Helper` в хелперах, `t.Cleanup` для ресурсов.
- [ ] White-box только через `export_hooks.go` с `//go:build hubtest`.
- [ ] Coverage-профиль в TMP, не закоммичен.

---
