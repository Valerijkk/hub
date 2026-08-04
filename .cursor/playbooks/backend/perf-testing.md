# Backend: производительность и тестирование

> Часть плейбука. Канон и оглавление — [`../backend.md`](../backend.md).

Замеры, профилирование, тестовые практики бэкенда.

---

## 11. Производительность

### 11.1. Пагинация: offset/limit для задач, cursor для лент

TZ-инвариант: **задачи — offset/limit страницами**, **чат/ленты — cursor (infinite scroll)**. В трекере реально используется курсор по `issue_key` со стабильным tie-breaker.

<good-example>

Cursor-пагинация задач (`tracker/repo_issues.go`) — курсор = `issue_key` последней строки, фильтр согласован с сортировкой:

```go
// Общее количество (до пагинации) — для UI.
total, err := query.Clone().Count(ctx)
if err != nil {
    return nil, 0, fmt.Errorf("подсчёт задач: %w", err)
}

// Курсор: строки после последнего выданного issue_key.
if params.Cursor != "" {
    cursorIssue, err := r.client.Issue.Query().
        Where(issue.IssueKeyEQ(params.Cursor), issue.HasProjectWith(project.IDEQ(projectID)), issue.DeletedAtIsNil()).
        Only(ctx)
    if err != nil {
        if ent.IsNotFound(err) {
            return []*ent.Issue{}, total, nil
        }
        return nil, 0, fmt.Errorf("курсорная задача: %w", err)
    }
    query = applyListIssuesCursorFilter(query, cursorIssue, params.OrderBy, params.Cursor)
}

// Жёсткий предел лимита — защита от «отдай всё».
limit := params.Limit
if limit <= 0 || limit > 100 {
    limit = 50
}
// Стабильный tie-breaker issue_key ASC — иначе курсор «скачет» на равных ключах сортировки.
query = query.Order(issue.ByCreatedAt(sql.OrderDesc()), issue.ByIssueKey())
items, err := query.Limit(limit).WithAssignee().WithStatus().WithLabels().All(ctx)
```

Фильтр курсора согласован с полем сортировки (иначе пропуски/дубли на границах страниц):

```go
func applyListIssuesCursorFilter(query *ent.IssueQuery, cursorIssue *ent.Issue, orderBy, cursor string) *ent.IssueQuery {
    keyAfter := issue.IssueKeyGT(cursor)
    switch orderBy {
    case "updated_at":
        return query.Where(issue.Or(
            issue.UpdatedAtLT(cursorIssue.UpdatedAt),
            issue.And(issue.UpdatedAtEQ(cursorIssue.UpdatedAt), keyAfter),
        ))
    default:
        return query.Where(issue.Or(
            issue.CreatedAtLT(cursorIssue.CreatedAt),
            issue.And(issue.CreatedAtEQ(cursorIssue.CreatedAt), keyAfter),
        ))
    }
}
```
</good-example>

<bad-example>

```go
// Нет верхнего предела лимита — клиент запросит limit=1000000 и уложит БД/память.
items, _ := query.Limit(params.Limit).All(ctx)
// Нестабильная сортировка без tie-breaker — курсор пропускает/дублирует строки.
query.Order(issue.ByCreatedAt())
```
</bad-example>

### 11.2. Всегда ограничивайте лимит

Любой листинг имеет жёсткий максимум (в трекере — 100, default 50). Даже для JQL: `Limit(100)`. Никаких безлимитных `.All(ctx)` по большим таблицам.

### 11.3. Индексы под запросы

Индексируйте то, по чему фильтруете/сортируете. См. §5.3 (Ent `Indexes()`) и §9.5 (Mongo). Отсутствие индекса на поле фильтра — seq scan на каждом запросе. `deleted_at` индексируется, т.к. каждый Query содержит `DeletedAtIsNil()`.

### 11.4. Избегайте N+1

Eager loading `.WithX()` (§5.7). Для агрегатов — `Count`/`GroupBy` на стороне БД, не в Go-цикле.

<bad-example>

```go
// Подсчёт в Go: тянем все строки в память ради счётчика.
all, _ := r.client.Issue.Query().All(ctx)
count := len(all)
```
</bad-example>

<good-example>

```go
count, err := r.client.Issue.Query().Where(issue.DeletedAtIsNil()).Count(ctx) // COUNT в БД
```
</good-example>

### 11.5. Пулы соединений

- PostgreSQL: пул настраивается на уровне драйвера/Ent (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`). Единый пул переиспользуется (`/health` использует shared `rawDB` — BE-H5).
- Redis: один `*redis.Client` на приложение, потокобезопасен, переиспользуется.
- HTTP-клиенты для интеграций: переиспользуйте `*http.Client` (safehttp), не создавайте на каждый запрос — иначе утечка соединений и рост latency.
- GOMAXPROCS (2026): рантайм Go 1.25 сам читает cgroup CPU-лимиты (v1/v2), выставляет GOMAXPROCS и перечитывает при изменении — библиотека `automaxprocs` не нужна. Условие: у backend-контейнера должны быть явные CPU limits в compose/K3s (чек devops), иначе поведение прежнее.

### 11.6. Замер времени операций

Оборачивайте значимые операции в `observability.Timed` — это даёт метрики latency без ручного таймера и без утечки секретов.

<good-example>

Реально (`tracker/service_issues.go`, `auth/service_session.go`):

```go
func (s *TrackerService) CreateIssue(ctx context.Context, projectID, userID uuid.UUID, req CreateIssueRequest) (*IssueResponse, error) {
    return observability.Timed(ctx, "tracker.CreateIssue", func(ctx context.Context) (*IssueResponse, error) {
        // … бизнес-логика …
    })
}

func (s *SessionService) RefreshToken(ctx context.Context, oldRefreshToken string) (*TokenPairResponse, error) {
    return observability.Timed(ctx, "auth.Refresh", func(ctx context.Context) (*TokenPairResponse, error) {
        // …
    })
}
```
</good-example>

### 11.7. Чек-лист производительности

- [ ] Пагинация: cursor для лент/чата, offset/limit-страницы для задач (в трекере — cursor по issue_key).
- [ ] У каждого листинга жёсткий верхний лимит (≤100).
- [ ] Сортировка со стабильным tie-breaker (для cursor).
- [ ] Поля фильтра/сортировки проиндексированы (PG + Mongo).
- [ ] Нет N+1: связи через `.WithX()`, агрегаты через `Count`/`GroupBy`.
- [ ] HTTP-клиенты и пулы переиспользуются.
- [ ] Значимые операции обёрнуты в `observability.Timed`.

---
## 13. Тестирование бэкенда

> Полный playbook — `.cursor/playbooks/testing.md`. Здесь — бэкенд-специфика.

### 13.1. Где живут тесты и почему

Бэкенд-тесты живут в `backend/testframework/{domain}/` с build-тегом `//go:build hubtest`, а **не** рядом с кодом в `internal/`. Запуск:

```bash
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...
```

**Почему в том же модуле, а не отдельным модулем?** Go запрещает импорт `internal/`-пакетов из другого модуля (правило видимости `internal`). Тесты обращаются к `github.com/valerii/hub/backend/internal/tracker` и т.п. — значит, они обязаны находиться в том же Go-модуле (`backend/`). Поэтому `testframework/` — это подкаталог `backend/`, а не отдельный модуль. Build-тег `hubtest` изолирует тестовый код от prod-сборки (`-tags release` его не видит).

### 13.2. Именование и структура

- Пакет теста: `<domain>_test` (например `tracker_test`) — black-box тестирование публичного API.
- Функция: `Test<Function>_<Scenario>` (например `TestUpdateIssue_VersionMismatch`).
- Фикстуры/фабрики: `backend/testframework/env/`, `factory/`.
- Инструмент: `testify` — `require` (останавливает тест при провале) и `assert` (продолжает).
- Контекст в тесте: `t.Context()` (Go 1.24) вместо `context.Background()` — отменяется по завершении теста и корректно гасит горутины/подключения, созданные тестом. В существующих тестах `context.Background()` не рефакторим массово; в новых — `t.Context()`. (2026)

<good-example>

Реальная шапка (`backend/testframework/tracker/service_test.go`):

```go
//go:build hubtest

// Package tracker содержит юнит-тесты для tracker.TrackerService.
package tracker_test

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/valerii/hub/backend/ent"
    "github.com/valerii/hub/backend/internal/tracker"
    apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)
```
</good-example>

### 13.3. Мок репозитория

Service тестируется с моком `Repository` (интерфейс существует именно для этого — принцип D из SOLID). Мок реализует все методы композитного интерфейса.

<good-example>

Реальный мок (`backend/testframework/tracker/service_test.go`):

```go
// mockRepo реализует tracker.Repository для тестирования TrackerService.
type mockRepo struct {
    issues          map[uuid.UUID]*ent.Issue
    issuesByKey     map[string]*ent.Issue
    statuses        map[uuid.UUID]*ent.WorkflowStatus
    projectCounters map[uuid.UUID]int
    updateVersionError bool // флаг для сценария optimistic-lock конфликта
}

func (m *mockRepo) CreateIssue(_ context.Context, projectID uuid.UUID, input tracker.CreateIssueInput) (*ent.Issue, error) {
    m.projectCounters[projectID]++
    num := m.projectCounters[projectID]
    key := fmt.Sprintf("TEST-%d", num)
    // … собираем *ent.Issue вручную …
}
```

Тест собирает service поверх мока и проверяет бизнес-поведение:

```go
func TestUpdateIssue_VersionMismatch(t *testing.T) {
    repo := newMockRepo()
    repo.updateVersionError = true // репозиторий вернёт конфликт версии
    svc := tracker.NewTrackerService(repo, nil)

    _, err := svc.UpdateIssue(context.Background(), "TEST-1", uuid.New(), tracker.UpdateIssueRequest{Version: 1})
    require.Error(t, err)

    var appErr *apperrors.AppError
    require.ErrorAs(t, err, &appErr)
    assert.Equal(t, 409, appErr.Code) // optimistic lock → 409 Conflict
}
```
</good-example>

### 13.4. Table-driven тесты

Для множества похожих кейсов — табличные тесты (например все 9 типов issue, все приоритеты).

<good-example>

```go
func TestCreateIssue_AllTypes(t *testing.T) {
    types := []string{"bug", "task", "story", "epic", "sub_task", "improvement", "tech_debt", "spike", "release"}
    for _, it := range types {
        t.Run(it, func(t *testing.T) {
            repo := newMockRepo()
            svc := tracker.NewTrackerService(repo, nil)
            resp, err := svc.CreateIssue(context.Background(), uuid.New(), uuid.New(),
                tracker.CreateIssueRequest{Title: "t", IssueType: it})
            require.NoError(t, err)
            assert.NotEmpty(t, resp.ID)
        })
    }
}
```
</good-example>

### 13.5. Интеграционные тесты

Для проверки реальных SQL/транзакций/optimistic-lock используются реальные БД (Docker/testcontainers) — canonical-прогон в Docker (`deployments/docker-compose.test.yml`, образ `golang:1.25-bookworm`, `-tags hubtest`). Схема применяется `client.Schema.Create(ctx)`.

### 13.6. Что покрывать в первую очередь

Приоритет покрытия (security-critical пакеты — цель ≥80%): `auth`, `ws`, `rbac`, `audit`, `calls`. Для новой фичи покрывайте: happy path, ошибки валидации, `ent.IsNotFound` → 404, optimistic-lock конфликт → 409, авторизацию/скоуп.

### 13.7. Test-triage при падении

> Обязательный skill: `.cursor/skills/test-triage/SKILL.md`.

При падении теста сначала определить, что верно — код или тест — сверяясь с `docs/`. Только потом чинить. Не «подгоняйте» тест под баг и не меняйте прод-код, не поняв, соответствует ли ожидание спецификации.

### 13.8. Новое в stdlib для тестов (2026)

- **`testing/synctest`** (stable в Go 1.25) — детерминированные тесты конкурентного кода: изолированный «пузырь» с виртуальными часами + `synctest.Wait()`. Применять для воркеров/тикеров/таймаутов вместо `time.Sleep` (см. §10.8).
- **Бенчмарки — только `for b.Loop()`** (Go 1.24): авто-исключение setup/cleanup из таймера, защита от dead-code elimination, один прогон функции на итерацию замера. `for range b.N` в новых бенчмарках не использовать. Учтите: `testing.AllocsPerRun` в Go 1.25 паникует при параллельных тестах.
- **`t.Context()`** — вместо `context.Background()` в новых тестах (см. §13.2).

### 13.9. Чек-лист тестов

- [ ] Тест в `backend/testframework/{domain}/` с `//go:build hubtest`.
- [ ] Пакет `<domain>_test`; имя `Test<Func>_<Scenario>`.
- [ ] Service покрыт через мок `Repository`.
- [ ] Покрыты: happy path, валидация, not-found (404), conflict (409), авторизация.
- [ ] Конкурентный код — с `-race`.
- [ ] Прогон: `go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...`.
- [ ] Артефакты покрытия удалены (`.cursor/rules/45-artifacts-cleanup.mdc`).

---
