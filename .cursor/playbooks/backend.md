# Hub — Эталонный Playbook по бэкенд-разработке

> Единый справочник для AI-агентов бэкенда Hub: `be1`, `be2` (параллельная разработка фич) и `be-ver` (верификация).
> Стек: Go 1.24/1.25 · Fiber v3 (fasthttp) · Ent + Atlas · PostgreSQL 17 · MongoDB 8 · Redis 7 · NATS JetStream · WebSocket (Fiber) · LiveKit · SeaweedFS (S3 через aws-sdk-go-v2).
> Связанные правила: `.cursor/rules/20-backend.mdc`, `.cursor/rules/10-style.mdc`, `.cursor/rules/60-tz-constraints.mdc`.
> Связанные playbook'и: `.cursor/playbooks/observability.md`, `.cursor/playbooks/testing.md`, `.cursor/playbooks/code-comments.md`, `.cursor/playbooks/git-workflow.md`.

Этот документ — источник истины по паттернам бэкенда. Он не заменяет базу знаний `docs/*`, а конкретизирует её для написания Go-кода. Все примеры взяты из реального кода репозитория (модули `auth`, `tracker`, `platform/*`, `ws`, `outbox`) — имена типов, пакетов и функций аутентичны.

---

## Оглавление

1. [Кто использует playbook и когда](#1-кто-использует-playbook-и-когда)
2. [Слоистая архитектура](#2-слоистая-архитектура)
3. [Обработка ошибок](#3-обработка-ошибок)
4. [Context](#4-context)
5. [Ent ORM](#5-ent-orm)
6. [Fiber v3](#6-fiber-v3)
7. [Безопасность](#7-безопасность)
8. [NATS JetStream и outbox](#8-nats-jetstream-и-outbox)
9. [MongoDB](#9-mongodb)
10. [Конкурентность](#10-конкурентность)
11. [Производительность](#11-производительность)
12. [Логирование и наблюдаемость](#12-логирование-и-наблюдаемость)
13. [Тестирование бэкенда](#13-тестирование-бэкенда)
14. [Комментарии в коде](#14-комментарии-в-коде)
15. [Git и коммиты](#15-git-и-коммиты)
16. [Обязательная регрессия и docs-sync](#16-обязательная-регрессия-и-docs-sync)
17. [Definition of Done и что проверяет be-ver](#17-definition-of-done-и-что-проверяет-be-ver)
18. [Антипаттерны Hub](#18-антипаттерны-hub)
19. [Быстрый чек-лист перед передачей be-ver](#19-быстрый-чек-лист-перед-передачей-be-ver)

---

## 1. Кто использует playbook и когда

### 1.1. Роли

| Агент | Роль | Что делает | Skill |
|-------|------|------------|-------|
| `be1` | Backend Dev, поток A | Реализует фичи бэкенда (handler/service/repository + тесты), делает кросс-ревью PR `be2` | `.cursor/project_skills/BE_skill.md` |
| `be2` | Backend Dev, поток B | Реализует фичи бэкенда параллельно с `be1`, делает кросс-ревью PR `be1` | `.cursor/project_skills/BE_skill.md` |
| `be-ver` | Верификатор бэкенда | Readonly-проверка результата: layering, security-инварианты, обработка ошибок, context, тесты, отсутствие data race и регрессии | — |

`be1` и `be2` — это два параллельных разработчика. Они пишут код одновременно, поэтому главная опасность — конфликты и рассинхрон интерфейсов. `be-ver` включается **после** того, как разработчик считает задачу готовой, и выносит вердикт `verified: true/false` с приложенным логом регрессии.

### 1.2. Когда открывать этот playbook

- Перед началом любой бэкенд-задачи (`be1`/`be2`) — сверить слои, паттерны ошибок, security-инварианты.
- Во время верификации (`be-ver`) — как чек-лист инвариантов (§17, §18, §19).
- При сомнении «а как в Hub принято» — сначала сюда, затем в `docs/architecture.md` и `docs/system_specification.md`.

### 1.3. Как be1 и be2 делят работу без конфликтов

Правила разделения труда, чтобы два разработчика не наступали друг на друга:

1. **Разные модули — по умолчанию.** Идеальный сплит: `be1` берёт один `backend/internal/<module>/`, `be2` — другой. Модуль (`auth`, `tracker`, `messenger`, `tms`, `calls`, `dashboard`, `notification`, `search`, `rbac`, …) — это единица владения. Внутри модуля один разработчик — один хозяин.

2. **Один модуль — разные файлы.** Если оба вынуждены работать в одном модуле, делите по файлам роли: `be1` — `service_issues.go` + `repo_issues.go`, `be2` — `service_comments.go` + `repo_comments.go`. Модули Hub уже нарезаны по фичам именно для этого (см. `tracker/`: `service_issues.go`, `service_comments.go`, `service_boards.go`, `service_workflow.go`, `service_time.go`, `service_labels.go` — каждый самодостаточен).

3. **Общие интерфейсы — согласовать ДО кода.** Файлы, которые трогают оба (`repository.go` с интерфейсом `Repository`, `dto.go`, `service.go` с конструктором `NewTrackerService`), — зона повышенного риска. Договорённость: интерфейс правит только один, второй потребляет. Изменения интерфейса анонсируются в PR-описании и `.planning/STATE.md`.

4. **Порядок мержа.** Первым мержится PR, меняющий общий интерфейс/DTO. Второй разработчик делает rebase на свежий `master` перед своим мержем. Никаких параллельных правок одного и того же метода интерфейса.

5. **Кросс-ревью обязателен.** `be1` ревьюит PR `be2` и наоборот. Ревью проверяет: слои не протекли, ошибки обёрнуты, тесты есть, регрессия зелёная.

<good-example>

**Хороший сплит одной фичи (issue-linking):**

```
be1:  tracker/repo_links.go      (EntRepository методы CreateIssueLink/…)
      tracker/service_links.go   (бизнес-логика)
      tracker/handler_links.go   (HTTP)
be2:  tracker/repository.go      (добавил интерфейс IssueLinkRepository — смержил ПЕРВЫМ)
      затем работает над своим доменом (comments)
```

Общий `IssueLinkRepository` вынесен в отдельный интерфейс (ISP), `be2` смержил его первым, `be1` реализовал.
</good-example>

<bad-example>

**Плохо:** оба правят `service_issues.go` метод `UpdateIssue` в разных ветках → конфликт слияния + потерянная логика optimistic locking у одного из них.
</bad-example>

### 1.4. Чек-лист координации (be1/be2)

- [ ] Задача привязана к конкретному модулю/файлам; пересечений с параллельной задачей нет.
- [ ] Если трогается общий интерфейс/DTO — это анонсировано, PR мержится первым.
- [ ] Перед своим мержем сделан rebase на актуальный `master`.
- [ ] PR другого разработчика отревьюен (layering, errors, tests, регрессия).
- [ ] `.planning/STATE.md` обновлён session-заметкой.

---

## 2. Слоистая архитектура

### 2.1. Слои и правило зависимостей

Поток запроса строго однонаправленный:

```
HTTP → Handler → Service → Repository (interface) → EntRepository → Ent/DB
                    │
                    └─→ NATS / Redis / внешние API (через интерфейсы)
```

Каждый слой зависит **только** от слоя ниже и **только через интерфейс**. Верхний слой не знает про реализацию нижнего.

| Слой | Файл | Ответственность | Категорически нельзя |
|------|------|-----------------|----------------------|
| Handler | `handler*.go` | Парсинг HTTP, извлечение userID/params, вызов service, маппинг в статус-код | Бизнес-логика, `*ent.Client`, прямые запросы к БД |
| Service | `service*.go` | Бизнес-логика, оркестрация репозитория/workflow/событий, валидация инвариантов | `*ent.Client`, `fiber.Ctx`, парсинг HTTP-тела |
| Repository (interface) | `repository.go` | Контракт доступа к данным (доменные методы) | Реализация (только сигнатуры) |
| EntRepository | `ent_repository.go` / `repo_*.go` | Реализация на Ent, транзакции, outbox, eager loading | Знание про `fiber.Ctx`, HTTP-статусы |

**Инвариант №1 (жёсткий):** `*ent.Client` живёт ТОЛЬКО в `ent_repository.go` / `repo_*.go`. Ни в Service, ни в Handler его быть не должно. Это нарушение layering и один из ключевых грепов `be-ver`.

### 2.2. Структура модуля

```
backend/internal/<module>/
├── handler.go / handler_<feature>.go  # HTTP-слой
├── service.go / service_<feature>.go  # бизнес-логика + конструктор
├── repository.go                      # интерфейсы Repository (+ узкие доменные)
├── ent_repository.go / repo_<feature>.go  # реализация на Ent
├── dto.go                             # Request/Response структуры (json + validate)
├── events.go                          # доменные события (payload-структуры для NATS/outbox)
└── (тесты — в backend/testframework/<module>/, НЕ здесь)
```

Крупные модули (`tracker`, `auth`) нарезаны по фичам: `service_issues.go`, `service_comments.go`, `repo_issues.go`, `handler_issues.go` и т.д. Это осознанный приём — уменьшает конфликты между `be1`/`be2` и облегчает навигацию.

### 2.3. Handler-слой

Handler делает ровно четыре вещи: (1) извлекает и валидирует вход (params, userID, тело), (2) вызывает service, (3) прокидывает ошибку в централизованный `ErrorHandler`, (4) возвращает статус-код и JSON.

<good-example>

Реальный handler трекера (`tracker/handler_issues.go`):

```go
// CreateIssue — Создать задачу
// @Summary Создать задачу
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateIssueRequest true "Данные задачи"
// @Success 201 {object} IssueResponse "Задача создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues [post]
func (h *Handler) CreateIssue(c fiber.Ctx) error {
    projectID, err := httputil.ParseProjectID(c)
    if err != nil {
        return err
    }
    userID, err := httputil.UserIDFromCtx(c)
    if err != nil {
        return err
    }

    var req CreateIssueRequest
    if err := c.Bind().JSON(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
    }

    resp, err := h.service.CreateIssue(c.Context(), projectID, userID, req)
    if err != nil {
        // Возвращаем ошибку как есть: ErrorHandler развернёт AppError в JSON с request_id.
        return err
    }

    return c.Status(fiber.StatusCreated).JSON(resp)
}
```

Здесь: `c.Context()` пробрасывается в service; ошибка service возвращается напрямую (не логируется и не оборачивается второй раз); статус `201` для create.
</good-example>

<bad-example>

Handler лезет в БД и содержит бизнес-логику — двойное нарушение:

```go
func (h *Handler) CreateIssue(c fiber.Ctx) error {
    var req CreateIssueRequest
    _ = c.Bind().JSON(&req)
    // НЕЛЬЗЯ: *ent.Client в handler, вычисление issue_key — это бизнес-логика.
    p, _ := h.db.Project.UpdateOneID(projectID).AddTaskCounter(1).Save(c.Context())
    issue, _ := h.db.Issue.Create().
        SetIssueKey(fmt.Sprintf("%s-%d", p.Key, p.TaskCounter)).
        SetTitle(req.Title).
        Save(c.Context())
    return c.JSON(issue)
}
```
</bad-example>

### 2.4. Service-слой

Service содержит бизнес-логику и оркестрирует зависимости. Он принимает уже распарсенные значения (`uuid.UUID`, DTO), общается с миром через интерфейсы (`Repository`, `WorkflowEngine`, `WIPChecker`, `OutboxPublisher`).

<good-example>

Реальный конструктор сервиса (`tracker/service.go`) — зависимости через интерфейсы:

```go
// TrackerService реализует бизнес-логику модуля трекера задач.
type TrackerService struct {
    repo         Repository       // интерфейс, не *ent.Client
    workflow     *WorkflowEngine
    wipChecker   WIPChecker       // интерфейс, может быть nil
    threadMirror ThreadMirror
}

// NewTrackerService создаёт новый сервис трекера.
// wipChecker может быть nil (WIP-лимиты не будут проверяться).
func NewTrackerService(repo Repository, wipChecker WIPChecker) *TrackerService {
    wf := NewWorkflowEngine(repo)
    return &TrackerService{repo: repo, workflow: wf, wipChecker: wipChecker}
}
```

Реальный метод сервиса (`tracker/service_issues.go`) — валидация workflow-перехода и optimistic locking делегируется репозиторию:

```go
func (s *TrackerService) UpdateIssue(ctx context.Context, issueKey string, userID uuid.UUID, req UpdateIssueRequest) (*IssueResponse, error) {
    existing, err := s.repo.FindIssueByKey(ctx, issueKey)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, apperrors.NewNotFound("задача не найдена")
        }
        return nil, apperrors.NewInternal("ошибка получения задачи")
    }

    // Бизнес-правило: смена статуса требует валидного workflow-перехода.
    if req.StatusID != nil {
        newStatusID, err := uuid.Parse(*req.StatusID)
        if err != nil {
            return nil, apperrors.NewBadRequest("некорректный status_id")
        }
        if existing.Edges.Status == nil {
            return nil, apperrors.NewBadRequest("задача не имеет статуса")
        }
        currentStatusID := existing.Edges.Status.ID
        if newStatusID != currentStatusID {
            if err := s.workflow.ValidateTransition(ctx, currentStatusID, newStatusID); err != nil {
                return nil, err
            }
        }
    }

    input := UpdateIssueInput{ /* маппинг DTO → доменный input */ Version: req.Version }
    updated, err := s.repo.UpdateIssue(ctx, existing.ID, input)
    if err != nil {
        return nil, err
    }
    return toIssueResponse(updated), nil
}
```
</good-example>

**Инвариант:** Service возвращает доменные `*AppError` (`apperrors.NewNotFound`, `NewBadRequest`, `NewConflict`, …), а не `fiber.Error`. Fiber-специфику знает только handler.

### 2.5. Repository-слой (интерфейс + реализация)

Интерфейс определяет контракт. Реализация — на Ent. Используем **ISP** (узкие доменные интерфейсы) плюс композитный `Repository` для удобного мока.

<good-example>

Реальные интерфейсы (`tracker/repository.go`) — сегрегация + композит:

```go
// Узкие доменные интерфейсы (ISP).
type IssueRepository interface {
    CreateIssue(ctx context.Context, projectID uuid.UUID, input CreateIssueInput) (*ent.Issue, error)
    FindIssueByKey(ctx context.Context, issueKey string) (*ent.Issue, error)
    ListIssues(ctx context.Context, projectID uuid.UUID, params ListIssuesParams) ([]*ent.Issue, int, error)
    UpdateIssue(ctx context.Context, issueID uuid.UUID, input UpdateIssueInput) (*ent.Issue, error)
    DeleteIssue(ctx context.Context, issueID uuid.UUID) error
    MoveIssue(ctx context.Context, issueID uuid.UUID, statusID uuid.UUID, rank string) (*ent.Issue, error)
}

// Композитный интерфейс — для обратной совместимости и мока в тестах.
type Repository interface {
    IssueRepository
    IssueLinkRepository
    WorkflowRepository
    CommentRepository
    // …
    GetDB() *ent.Client // прямой доступ к Ent только для JQL/сложных запросов
}

// EntRepository реализует Repository на Ent с публикацией событий через outbox.
type EntRepository struct {
    client    *ent.Client   // *ent.Client живёт ТОЛЬКО здесь
    publisher OutboxPublisher
}

func NewEntRepository(client *ent.Client, publisher OutboxPublisher) *EntRepository {
    return &EntRepository{client: client, publisher: publisher}
}
```

Обратите внимание: доменные методы принимают `uuid.UUID`, а не `string` (парсинг строк — в сервисе), и возвращают `*ent.Issue`, а не HTTP-DTO.
</good-example>

**О методе `GetDB() *ent.Client`.** Это осознанный люк для сложных случаев (JQL-поиск), когда фиксированные методы репозитория не подходят. Использовать редко и только в handler'ах уровня «поиск», где нельзя заранее знать предикаты. Не превращать его в способ обойти слои для обычного CRUD.

### 2.6. Полный пример мини-модуля

Ниже — эталонный скелет модуля `example` (склеен из паттернов `tracker`/`auth`). Демонстрирует все четыре слоя.

```go
// ---- dto.go ----
package example

type CreateWidgetRequest struct {
    Name string `json:"name" validate:"required,min=1,max=100"`
}

type WidgetResponse struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    Version int    `json:"version"`
}
```

```go
// ---- repository.go ----
package example

import (
    "context"

    "github.com/google/uuid"
    "github.com/valerii/hub/backend/ent"
)

type WidgetRepository interface {
    CreateWidget(ctx context.Context, orgID uuid.UUID, name string) (*ent.Widget, error)
    FindWidgetByID(ctx context.Context, id uuid.UUID) (*ent.Widget, error)
    UpdateWidget(ctx context.Context, id uuid.UUID, name string, version int) (*ent.Widget, error)
}

type EntRepository struct {
    client *ent.Client
}

func NewEntRepository(client *ent.Client) *EntRepository {
    return &EntRepository{client: client}
}
```

```go
// ---- ent_repository.go ----
package example

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/valerii/hub/backend/ent"
    "github.com/valerii/hub/backend/ent/widget"
    apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

func (r *EntRepository) CreateWidget(ctx context.Context, orgID uuid.UUID, name string) (*ent.Widget, error) {
    w, err := r.client.Widget.Create().
        SetOrgID(orgID).
        SetName(name).
        Save(ctx)
    if err != nil {
        return nil, fmt.Errorf("create widget: %w", err)
    }
    return w, nil
}

func (r *EntRepository) FindWidgetByID(ctx context.Context, id uuid.UUID) (*ent.Widget, error) {
    return r.client.Widget.Query().
        Where(widget.IDEQ(id), widget.DeletedAtIsNil()).
        Only(ctx)
}

// UpdateWidget — оптимистичная блокировка через version.
func (r *EntRepository) UpdateWidget(ctx context.Context, id uuid.UUID, name string, version int) (*ent.Widget, error) {
    affected, err := r.client.Widget.Update().
        Where(widget.IDEQ(id), widget.VersionEQ(version)).
        SetName(name).
        AddVersion(1).
        Save(ctx)
    if err != nil {
        return nil, fmt.Errorf("update widget: %w", err)
    }
    if affected == 0 {
        return nil, apperrors.NewConflict("version mismatch: виджет был изменён другим пользователем")
    }
    return r.FindWidgetByID(ctx, id)
}
```

```go
// ---- service.go ----
package example

import (
    "context"

    "github.com/google/uuid"
    "github.com/valerii/hub/backend/ent"
    apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

type Service struct {
    repo WidgetRepository
}

func NewService(repo WidgetRepository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateWidget(ctx context.Context, orgID uuid.UUID, req CreateWidgetRequest) (*WidgetResponse, error) {
    w, err := s.repo.CreateWidget(ctx, orgID, req.Name)
    if err != nil {
        return nil, apperrors.NewInternal("не удалось создать виджет")
    }
    return toWidgetResponse(w), nil
}

func (s *Service) GetWidget(ctx context.Context, id uuid.UUID) (*WidgetResponse, error) {
    w, err := s.repo.FindWidgetByID(ctx, id)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, apperrors.NewNotFound("виджет не найден")
        }
        return nil, apperrors.NewInternal("ошибка получения виджета")
    }
    return toWidgetResponse(w), nil
}

func toWidgetResponse(w *ent.Widget) *WidgetResponse {
    return &WidgetResponse{ID: w.ID.String(), Name: w.Name, Version: w.Version}
}
```

```go
// ---- handler.go ----
package example

import (
    "github.com/gofiber/fiber/v3"
    apperr "github.com/valerii/hub/backend/internal/platform/errors"
    httputil "github.com/valerii/hub/backend/internal/platform/http"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) CreateWidget(c fiber.Ctx) error {
    orgID, err := httputil.ParseUUIDParam(c, "orgId")
    if err != nil {
        return err
    }
    var req CreateWidgetRequest
    if err := c.Bind().JSON(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
    }
    resp, err := h.service.CreateWidget(c.Context(), orgID, req)
    if err != nil {
        return err
    }
    return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) RegisterRoutes(api fiber.Router) {
    g := api.Group("/orgs/:orgId/widgets")
    g.Post("/", h.CreateWidget)
}
```

### 2.7. Чек-лист слоёв

- [ ] `*ent.Client` встречается только в `ent_repository.go` / `repo_*.go`.
- [ ] Handler не содержит бизнес-логики (только parse → call → status).
- [ ] Service не импортирует `fiber` и не парсит HTTP-тело.
- [ ] Service зависит от `Repository` (интерфейс), а не от `*ent.Client`.
- [ ] Репозиторные методы принимают `uuid.UUID`/доменные input-структуры, а не сырые строки.
- [ ] Конвертация `*ent.X` → `XResponse` — в `toXResponse`, не в handler.

---

## 3. Обработка ошибок

### 3.1. Философия

Ошибка в Hub несёт три вещи: (1) причину для логов (обёрнутая цепочка через `%w`), (2) HTTP-статус и безопасное сообщение для клиента (`AppError`), (3) `request_id` для корреляции. Голый `return err` уничтожает контекст и запрещён.

### 3.2. Пакет apperrors

Единый пакет — `apperrors "github.com/valerii/hub/backend/internal/platform/errors"`.

<good-example>

Реальная структура (`platform/errors/errors.go`):

```go
// AppError — структурированная ошибка с HTTP-кодом, сообщением и ID запроса.
type AppError struct {
    Code       int    `json:"code"`
    Message    string `json:"message"`
    RequestID  string `json:"request_id,omitempty"`
    Details    any    `json:"details,omitempty"`
    RetryAfter int    `json:"-"` // секунды для заголовка Retry-After (429)
}

func (e *AppError) Error() string {
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func NewBadRequest(msg string) *AppError   { return &AppError{Code: 400, Message: msg} }
func NewUnauthorized(msg string) *AppError  { return &AppError{Code: 401, Message: msg} }
func NewForbidden(msg string) *AppError     { return &AppError{Code: 403, Message: msg} }
func NewNotFound(msg string) *AppError      { return &AppError{Code: 404, Message: msg} }
func NewConflict(msg string) *AppError      { return &AppError{Code: 409, Message: msg} }
func NewRateLimited(msg string) *AppError   { return &AppError{Code: 429, Message: msg} }
func NewInternal(msg string) *AppError      { return &AppError{Code: 500, Message: msg} }
```

Готовые сообщения — в `platform/errors/messages.go` (`MsgBadRequest`, `MsgUnauthorized`, `MsgForbidden`, `MsgNotFound`, `MsgInternal`, `MsgConflict`).
</good-example>

### 3.3. Централизованный маппинг в HTTP

Fiber сконфигурирован с `ErrorHandler: middleware.ErrorHandler` (см. `cmd/server/wiring_fiber.go`). Handler просто возвращает ошибку — разворачивание в JSON `{"error"/"message", "request_id"}` происходит централизованно.

<good-example>

Реальный `platform/middleware/middleware.go`:

```go
func ErrorHandler(c fiber.Ctx, err error) error {
    requestID, _ := c.Locals("requestID").(string)

    var appErr *apperrors.AppError
    if errors.As(err, &appErr) {
        appErr.RequestID = requestID
        if appErr.RetryAfter > 0 {
            c.Set("Retry-After", strconv.Itoa(appErr.RetryAfter))
        }
        return c.Status(appErr.Code).JSON(appErr)
    }

    var fiberErr *fiber.Error
    if errors.As(err, &fiberErr) {
        return c.Status(fiberErr.Code).JSON(apperrors.AppError{
            Code:      fiberErr.Code,
            Message:   fiberErr.Message,
            RequestID: requestID,
        })
    }

    // Неизвестная ошибка — 500, детали не утекают клиенту.
    return c.Status(500).JSON(apperrors.AppError{
        Code:      500,
        Message:   "internal server error",
        RequestID: requestID,
    })
}
```

Следствие: **никогда не логируйте и не оборачивайте ошибку повторно в handler'е** — просто `return err`. Внутренние детали не должны утекать клиенту: для 500 клиент видит generic-сообщение, а причина уходит в лог.
</good-example>

### 3.4. Оборачивание: fmt.Errorf с %w

На границах, где ошибка полезна для диагностики (репозиторий, интеграции), всегда оборачиваем с `%w`, добавляя контекст операции.

<good-example>

Из `tracker/repo_issues.go` и `auth/service_session.go`:

```go
tx, err := r.client.Tx(ctx)
if err != nil {
    return nil, fmt.Errorf("начало транзакции: %w", err)
}
// …
if err := s.storeRefreshToken(ctx, refreshToken, payload); err != nil {
    return nil, fmt.Errorf("create session: store refresh: %w", err)
}
```
</good-example>

<bad-example>

```go
tx, err := r.client.Tx(ctx)
if err != nil {
    return nil, err // потерян контекст: где упало и что делали
}
```
</bad-example>

### 3.5. errors.Is / errors.As / ent.IsNotFound

- `ent.IsNotFound(err)` — «запись не найдена» → маппим в `apperrors.NewNotFound(...)`.
- `errors.Is(err, target)` — сравнение с sentinel (например `redis.Nil`, `context.DeadlineExceeded`).
- `errors.As(err, &target)` — извлечение конкретного типа (`*AppError`, `*fiber.Error`).

<good-example>

Реальный паттерн `ent.IsNotFound` (`tracker/service_issues.go`):

```go
i, err := s.repo.FindIssueByKey(ctx, issueKey)
if err != nil {
    if ent.IsNotFound(err) {
        return nil, apperrors.NewNotFound("задача не найдена")
    }
    return nil, apperrors.NewInternal("ошибка получения задачи")
}
```

Реальный `errors.Is(redis.Nil)` (`auth/service_session.go`):

```go
data, err := s.redis.Get(ctx, refreshRedisKey(token)).Bytes()
if errors.Is(err, redis.Nil) {
    return nil, fmt.Errorf("refresh token not found")
}
if err != nil {
    return nil, err
}
```
</good-example>

### 3.6. Sentinel-ошибки

Для доменных состояний, которые вызывающий код должен различать, объявляем sentinel-ошибки в пакете модуля.

<good-example>

Из `auth/magiclink.go`:

```go
var (
    ErrMagicLinkRateLimited   = fmt.Errorf("magic link rate limit exceeded")
    ErrMagicLinkInvalid       = fmt.Errorf("invalid or expired code")
    ErrMagicLinkVerifyLimited = fmt.Errorf("too many verification attempts, try again later")
)

// Возврат sentinel'а:
if count > magicLinkRateMax {
    return "", ErrMagicLinkRateLimited
}
```

Вызывающий сравнивает через `errors.Is(err, auth.ErrMagicLinkRateLimited)` и решает, вернуть `429` или `400`.
</good-example>

### 3.7. Запрет panic в production

`panic()` в продакшн-путях запрещён. Аудит подтверждает: `panic()` в production — 0 найдено (`security_guidelines.md`). Возвращайте ошибку.

Исключения, где `recover` уместен: воркеры, которые не должны падать целиком из-за одной записи. Тогда `recover` логирует и продолжает.

<good-example>

Из `outbox/worker.go` — локальный recover, чтобы паника на одном событии не уносила весь loop:

```go
func (w *Worker) processEvent(ctx context.Context, evt *ent.OutboxEvent) {
    defer func() {
        if r := recover(); r != nil {
            slog.Error("outbox: panic while processing event",
                "event_id", evt.ID,
                "event_type", evt.EventType,
                "panic", r,
                "stack", string(debug.Stack()),
            )
        }
    }()
    // … публикация в NATS …
}
```

И глобальный `fiberrecover` в `wiring_fiber.go` — чтобы паника в handler'е не роняла процесс, а превращалась в `500` с `request_id`.
</good-example>

<bad-example>

```go
func (s *Service) mustGet(ctx context.Context, id uuid.UUID) *ent.Widget {
    w, err := s.repo.FindWidgetByID(ctx, id)
    if err != nil {
        panic(err) // роняет процесс на пользовательском вводе
    }
    return w
}
```
</bad-example>

### 3.8. Что нельзя логировать

Никогда не пишите в лог секреты: JWT целиком, magic-link коды в проде, refresh-токены, webhook-секреты, содержимое `req.Body` с чувствительными полями. См. §7 и §12.

### 3.9. Чек-лист ошибок

- [ ] Нет голого `return err` на границах (везде `%w` с контекстом операции).
- [ ] `ent.IsNotFound` → `apperrors.NewNotFound`; не 500 на «не найдено».
- [ ] Service возвращает `*AppError`, не `fiber.Error`.
- [ ] Handler возвращает ошибку без повторного логирования/оборачивания.
- [ ] Нет `panic()` в prod-путях (кроме контролируемого `recover` в воркерах).
- [ ] Секреты не попадают в текст ошибки/лог.

---

## 4. Context

### 4.1. Правило первого параметра

`context.Context` — всегда первый параметр функции, имя `ctx`. Это соблюдается сквозь все слои.

<good-example>

```go
func (s *TrackerService) CreateIssue(ctx context.Context, projectID, userID uuid.UUID, req CreateIssueRequest) (*IssueResponse, error)
func (r *EntRepository) FindIssueByID(ctx context.Context, id uuid.UUID) (*ent.Issue, error)
```
</good-example>

<bad-example>

```go
func (s *Service) CreateIssue(req CreateIssueRequest, ctx context.Context) // ctx не первый
func (s *Service) CreateIssue(req CreateIssueRequest)                      // ctx потерян
```
</bad-example>

### 4.2. Propagation из Fiber

В handler'е берём контекст запроса через `c.Context()` и пробрасываем в service. Не создавайте `context.Background()` в handler'е — иначе теряется отмена и дедлайн запроса.

<good-example>

```go
resp, err := h.service.CreateIssue(c.Context(), projectID, userID, req)
```
</good-example>

<bad-example>

```go
resp, err := h.service.CreateIssue(context.Background(), projectID, userID, req) // потеряли cancel/deadline запроса
```
</bad-example>

### 4.3. Отмена и дедлайны

Контекст отменяется, когда клиент разорвал соединение или истёк таймаут. Все вызовы БД/Redis/NATS принимают `ctx` и уважают отмену. Для фоновых операций с ограничением по времени используйте `context.WithTimeout`.

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel() // всегда defer cancel — иначе утечка

result, err := externalClient.Fetch(ctx, url)
if errors.Is(err, context.DeadlineExceeded) {
    return nil, apperrors.NewInternal("upstream timeout")
}
```

### 4.4. context.DeadlineExceeded и Canceled

Различайте таймаут (`context.DeadlineExceeded`) и отмену клиентом (`context.Canceled`). При отмене клиентом обычно не нужно шуметь в логах на уровне error.

```go
if errors.Is(err, context.Canceled) {
    // клиент ушёл — это норма, лог на debug
    slog.DebugContext(ctx, "request canceled by client")
    return nil, err
}
if errors.Is(err, context.DeadlineExceeded) {
    slog.WarnContext(ctx, "operation timed out", "op", "fetch")
    return nil, apperrors.NewInternal("timeout")
}
```

### 4.5. rootCtx для фоновых горутин

Долгоживущие горутины (rate-limiter cleanup, outbox worker, ws Hub) привязываются к корневому контексту приложения (`rootCtx`), чтобы корректно завершаться при graceful shutdown.

<good-example>

Из `platform/middleware/ratelimit.go` — cleanup-горутина реагирует на `ctx.Done()`:

```go
func RateLimit(ctx context.Context, config RateLimitConfig) fiber.Handler {
    if ctx == nil {
        ctx = context.Background()
    }
    limiter := &rateLimiter{config: config}
    go limiter.cleanup(ctx) // ctx = rootCtx приложения
    // …
}

func (rl *rateLimiter) cleanup(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return // детерминированный shutdown, без утечки горутины
        case <-ticker.C:
            // … очистка …
        }
    }
}
```

Проброс `rootCtx` виден в `wiring_fiber.go`: `middleware.RateLimit(d.rootCtx, middleware.APIRateLimit)`.
</good-example>

### 4.6. Запрет хранить ctx в структурах

`context.Context` не хранится в полях структур — он передаётся как аргумент на каждый вызов. Хранение ctx в структуре ломает жизненный цикл (структура живёт дольше запроса) и маскирует утечки.

<bad-example>

```go
type Service struct {
    ctx  context.Context // НЕЛЬЗЯ
    repo Repository
}

func (s *Service) Do() error {
    return s.repo.Save(s.ctx, ...) // ctx протух ещё в прошлом запросе
}
```
</bad-example>

<good-example>

```go
type Service struct {
    repo Repository
}

func (s *Service) Do(ctx context.Context) error {
    return s.repo.Save(ctx, ...) // ctx на каждый вызов
}
```

Исключение — конструкторы фоновых компонентов, которым `rootCtx` передаётся для управления жизненным циклом горутины (как в `RateLimit`). Это не «хранение ctx запроса», а привязка фоновой задачи к жизни приложения.
</good-example>

### 4.7. Логирование с контекстом

Используйте `slog.InfoContext(ctx, ...)` / `WarnContext` / `ErrorContext` — они подхватывают trace/`request_id` из ctx (см. §12).

### 4.8. Чек-лист context

- [ ] `ctx context.Context` — первый параметр каждой доменной функции.
- [ ] В handler'е используется `c.Context()`, а не `context.Background()`.
- [ ] `context.WithTimeout`/`WithCancel` всегда сопровождается `defer cancel()`.
- [ ] `context.DeadlineExceeded`/`Canceled` обрабатываются осознанно.
- [ ] Нет `ctx` в полях структур (кроме привязки фоновой горутины к rootCtx).
- [ ] Фоновые горутины завершаются по `ctx.Done()`.

---

## 5. Ent ORM

### 5.1. Схемы: обязательные поля

Каждая изменяемая Ent-схема Hub включает: UUID v4 PK (`Immutable`), `version` (optimistic lock), `created_at`/`updated_at` (UTC), `deleted_at` (soft delete, корзина 30 дней для сущностей с корзиной).

<good-example>

Реальная схема (`backend/ent/schema/issue.go`):

```go
func (Issue) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            Immutable(),                         // UUID v4, строки запрещены
        field.String("issue_key").
            Unique().
            Immutable().
            Comment("Project-prefixed ID: HUB-1"),
        field.String("title").
            NotEmpty().
            MaxLen(500),
        field.JSON("description_json", map[string]interface{}{}).
            Optional().
            Comment("Plate editor JSON"),
        field.Enum("issue_type").
            Values("bug", "task", "story", "epic", "sub_task", "improvement", "tech_debt", "spike", "release"),
        field.Enum("priority").
            Values("critical", "high", "medium", "low", "trivial").
            Default("medium"),
        field.Int("version").
            Default(1).
            Comment("Optimistic locking"),       // обязательно на mutable
        field.Time("deleted_at").
            Optional().
            Nillable().
            Comment("Время мягкого удаления (корзина 30 дней)"),
        field.Time("created_at").
            Default(nowUTC).                      // UTC-хелпер, не time.Now локальный
            Immutable(),
        field.Time("updated_at").
            Default(nowUTC).
            UpdateDefault(nowUTC),
    }
}
```

`nowUTC` — общий хелпер, возвращающий `time.Now().UTC()`, чтобы БД всегда хранила UTC (TZ-инвариант).
</good-example>

### 5.2. Edges (связи)

Связи описываются в `Edges()`. Двунаправленные связи через `edge.From(...).Ref(...)`.

<good-example>

Реальные edges (`issue.go`):

```go
func (Issue) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("project", Project.Type).Ref("issues").Unique().Required(),
        edge.From("status", WorkflowStatus.Type).Ref("issues").Unique().Required(),
        edge.To("assignee", User.Type).Unique(),
        edge.To("reporter", User.Type).Unique(),
        edge.To("children", Issue.Type).From("parent").Unique(), // self-reference
        edge.To("comments", Comment.Type),
        edge.To("labels", Label.Type),
        // …
    }
}
```
</good-example>

### 5.3. Индексы

Индексы объявляются в `Indexes()` — по полям фильтрации/сортировки и soft-delete.

<good-example>

Реальные индексы (`issue.go`):

```go
func (Issue) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("issue_key"),
        index.Fields("issue_number"),
        index.Fields("rank"),
        index.Fields("created_at"),
        index.Fields("priority"),
        index.Fields("due_date"),
        index.Fields("deleted_at"), // ускоряет WHERE deleted_at IS NULL
    }
}
```
</good-example>

### 5.4. Миграции Atlas

Схемы генерируют код (`go generate ./ent`), миграции применяются через Atlas. Правила:

- Меняете схему → регенерируете Ent-код (`go generate ./ent`) → генерируете миграцию Atlas.
- Не редактируйте сгенерированный код в `backend/ent/*` вручную (кроме `backend/ent/schema/*`).
- Миграции идут в контроль версий; несовместимые изменения (drop колонки, смена типа) требуют плана и согласования (данные UTC, версия — критичны).
- Новое обязательное поле на существующей таблице — добавляйте с `Default(...)` или отдельной data-миграцией, чтобы не сломать существующие строки.

> **Drift и целевое состояние (2026).** Фактически в коде сейчас применяется авто-миграция `client.Schema.Create(ctx)` — каталога versioned-миграций в `backend/ent/migrate/` нет. Целевое состояние: **versioned migrations Atlas + `atlas migrate lint` в CI** (dev-database в services) — lint ловит destructive/data-dependent изменения до мержа и защищает историю миграций от конкурентных правок be1/be2. Для монолита с одной БД авто-миграция терпима; переход на versioned — обязательное предусловие schema-per-service (курс на микросервисы, `MICROSERVICES-ROADMAP`).

### 5.5. Транзакции

Атомарные многошаговые операции — через `client.Tx(ctx)`. Паттерн: rollback при любой ошибке, commit в конце, outbox-событие внутри той же транзакции.

<good-example>

Реальный `CreateIssue` (`tracker/repo_issues.go`) — счётчик + создание + событие атомарно:

```go
func (r *EntRepository) CreateIssue(ctx context.Context, projectID uuid.UUID, input CreateIssueInput) (*ent.Issue, error) {
    tx, err := r.client.Tx(ctx)
    if err != nil {
        return nil, fmt.Errorf("начало транзакции: %w", err)
    }

    // Атомарный инкремент счётчика задач проекта.
    p, err := tx.Project.UpdateOneID(projectID).AddTaskCounter(1).Save(ctx)
    if err != nil {
        _ = tx.Rollback()
        return nil, fmt.Errorf("инкремент счётчика задач: %w", err)
    }

    issueKey := fmt.Sprintf("%s-%d", p.Key, p.TaskCounter)

    created, err := tx.Issue.Create().
        SetIssueNumber(p.TaskCounter).
        SetIssueKey(issueKey).
        SetTitle(input.Title).
        SetProjectID(projectID).
        SetReporterID(input.ReporterID).
        Save(ctx)
    if err != nil {
        _ = tx.Rollback()
        return nil, fmt.Errorf("создание задачи: %w", err)
    }

    // Outbox-событие пишется в ТОЙ ЖЕ транзакции — атомарность данных и события.
    if r.publisher != nil {
        if err = r.publisher.PublishInTx(ctx, tx, fmt.Sprintf("project.%s.issue.created", projectID), IssueCreatedEvent{
            IssueID:  created.ID.String(),
            IssueKey: issueKey,
        }); err != nil {
            _ = tx.Rollback()
            return nil, fmt.Errorf("публикация события: %w", err)
        }
    }

    if err = tx.Commit(); err != nil {
        return nil, fmt.Errorf("фиксация транзакции: %w", err)
    }
    return r.FindIssueByID(ctx, created.ID)
}
```

Ключевые моменты: (1) операции внутри tx используют `tx.<Entity>`, а не `r.client.<Entity>`; (2) любой ранний выход → `tx.Rollback()`; (3) событие пишется через `PublishInTx` внутри tx; (4) commit возвращает ошибку — её тоже оборачиваем.
</good-example>

<bad-example>

Забыт rollback → соединение и блокировки утекают; событие вне транзакции → рассинхрон (событие есть, данных нет):

```go
tx, _ := r.client.Tx(ctx)
p, err := tx.Project.UpdateOneID(projectID).AddTaskCounter(1).Save(ctx)
if err != nil {
    return nil, err // НЕТ rollback — транзакция подвисла
}
created, _ := tx.Issue.Create()./* … */.Save(ctx)
tx.Commit()
// Событие ПОСЛЕ commit и вне tx — если тут упадём, данные есть, события нет.
r.publisher.Publish(ctx, r.client, "…", evt)
```
</bad-example>

### 5.6. Optimistic locking через version

Изменяемые сущности имеют поле `version`. Update идёт с условием `VersionEQ(expected)` и `AddVersion(1)`. Если `affected == 0` — версия не совпала, возвращаем `409 Conflict`.

<good-example>

Реальный `UpdateIssue` (`tracker/repo_issues.go`):

```go
func (r *EntRepository) UpdateIssue(ctx context.Context, issueID uuid.UUID, input UpdateIssueInput) (*ent.Issue, error) {
    update := r.client.Issue.Update().
        Where(
            issue.IDEQ(issueID),
            issue.VersionEQ(input.Version), // клиент прислал версию, которую редактировал
        ).
        AddVersion(1)

    if input.Title != nil {
        update = update.SetTitle(*input.Title)
    }
    // … остальные поля …

    affected, err := update.Save(ctx)
    if err != nil {
        return nil, fmt.Errorf("обновление задачи: %w", err)
    }
    if affected == 0 {
        // Никто не совпал по (id, version) → кто-то изменил запись раньше.
        return nil, apperrors.NewConflict("version mismatch: задача была изменена другим пользователем")
    }
    return r.FindIssueByID(ctx, issueID)
}
```

DTO обязывает клиента прислать версию: в `UpdateIssueRequest` поле `Version int json:"version" validate:"required"`.
</good-example>

### 5.7. N+1 и eager loading через .With...()

По умолчанию Ent не грузит связи. Читаете связи в цикле — получаете N+1. Решение: `.WithX()` eager loading в самом запросе.

<good-example>

Реальный `FindIssueByID` (`tracker/repo_issues.go`) — все нужные связи загружаются одним запросом:

```go
func (r *EntRepository) FindIssueByID(ctx context.Context, id uuid.UUID) (*ent.Issue, error) {
    return r.client.Issue.Query().
        Where(issue.IDEQ(id), issue.DeletedAtIsNil()).
        WithAssignee().
        WithReporter().
        WithStatus().
        WithLabels().
        WithParent().
        WithChildren().
        WithProject().
        Only(ctx)
}
```

После этого `issue.Edges.Assignee`, `issue.Edges.Status` и т.д. доступны без дополнительных запросов.
</good-example>

<bad-example>

N+1: для каждого issue отдельный запрос за статусом:

```go
issues, _ := r.client.Issue.Query().All(ctx)
for _, iss := range issues {
    // Каждая итерация — новый SELECT в БД.
    st, _ := iss.QueryStatus().Only(ctx)
    _ = st
}
```
</bad-example>

### 5.8. Soft delete

Удаление — установка `deleted_at`, а не физический DELETE. Все Query исключают удалённые через `DeletedAtIsNil()`.

<good-example>

```go
// DeleteIssue выполняет мягкое удаление (устанавливает deleted_at).
func (r *EntRepository) DeleteIssue(ctx context.Context, issueID uuid.UUID) error {
    return r.client.Issue.UpdateOneID(issueID).
        SetDeletedAt(time.Now().UTC()). // UTC!
        Exec(ctx)
}

// Все чтения фильтруют удалённые.
r.client.Issue.Query().Where(issue.IDEQ(id), issue.DeletedAtIsNil()).Only(ctx)
```

Физический hard delete — только по явному запросу Admin/GDPR (каскадное удаление с `errors.Join` частичных ошибок, см. `audit/service.go` `DeleteUserGDPR`). Очистка корзины (30 дней) — cron `PurgeExpired`.
</good-example>

### 5.9. Hooks и privacy

Ent-хуки (`Hooks()`) и privacy-политики (`Policy()`) применяются осознанно и централизованно (например, автозаполнение аудита, tenant-скоуп). Не размазывайте бизнес-логику по хукам — основная логика в service-слое. Хуки — для сквозных инвариантов (timestamps, аудит), privacy — для жёстких запретов доступа на уровне ORM.

### 5.10. Прямой доступ к Ent (JQL)

Для запросов с динамическими предикатами (JQL-поиск) допустим `repo.GetDB()`, но с обязательным скоупом по проекту/пользователю и soft-delete фильтром.

<good-example>

Из `tracker/handler_issues.go` (`SearchIssuesJQL`) — динамические предикаты, но базовый скоуп фиксирован:

```go
query := h.service.repo.GetDB().Issue.Query().
    Where(issue.HasProjectWith(project.ID(projectID))). // скоуп по проекту
    Where(issue.DeletedAtIsNil())                        // soft-delete
for _, pred := range predicates {                        // JQL-предикаты
    query = query.Where(pred)
}
issues, err := query.WithStatus().WithAssignee().Limit(100).All(c.Context())
```
</good-example>

### 5.11. Чек-лист Ent

- [ ] PK — `field.UUID(...).Default(uuid.New).Immutable()`; никаких `field.String("id")`.
- [ ] Изменяемая схема имеет `version` (`Int.Default(1)`).
- [ ] `created_at`/`updated_at` через UTC-хелпер; время в БД — UTC.
- [ ] Есть `deleted_at` + индекс, если сущность поддерживает корзину.
- [ ] Многошаговые изменения — в `client.Tx`; rollback на каждой ошибке; commit-ошибка обёрнута.
- [ ] Outbox-событие пишется через `PublishInTx` внутри той же транзакции.
- [ ] Update изменяемой сущности использует `VersionEQ` + `AddVersion(1)` + проверку `affected == 0`.
- [ ] Чтения со связями используют `.WithX()` (нет N+1).
- [ ] Все Query исключают soft-deleted (`DeletedAtIsNil`).
- [ ] После изменения схемы регенерирован Ent-код и создана миграция Atlas.

---

## 6. Fiber v3

### 6.1. Почему именно Fiber v3

Fiber v3 на fasthttp — единственный HTTP-фреймворк Hub. **Запрещено** заменять на `net/http` или gorilla (в том числе для WebSocket — используется встроенный Fiber-механизм). Сигнатура handler'а в v3 — `func(c fiber.Ctx) error` (значение, не указатель `*fiber.Ctx` как в v2).

### 6.2. Конфигурация приложения

<good-example>

Реальная конфигурация (`cmd/server/wiring_fiber.go`):

```go
app := fiber.New(fiber.Config{
    ReadTimeout:             cfg.Server.ReadTimeout,
    WriteTimeout:            cfg.Server.WriteTimeout,
    Immutable:               cfg.Server.Immutable,
    ErrorHandler:            middleware.ErrorHandler, // централизованный маппинг ошибок
    EnableTrustedProxyCheck: true,                    // NF-11: не доверяем X-Forwarded-For слепо
    TrustedProxies: []string{
        "127.0.0.1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
    },
    // ME-15: глобальный лимит тела запроса — 25 МБ (align c presigned-загрузкой).
    BodyLimit: 25 * 1024 * 1024,
})
```

`EnableTrustedProxyCheck` + приватные CIDR — обязательны: без них `c.IP()` подделывается через `X-Forwarded-For`, что ломает rate-limit по IP.
</good-example>

> **Предупреждение при апгрейде на Fiber v3 stable (2026).** Hub сидит на `v3.0.0-beta.3`, а v3 вышел в stable (текущая v3.3.0 требует Go 1.25). В stable переименован proxy-конфиг: `EnableTrustedProxyCheck`/`TrustedProxies` → `TrustProxy`/`TrustProxyConfig`; часть полей `fiber.Config` (`DisableStartupMessage`, `EnablePrefork` и др.) переехала в `ListenConfig`; middleware перешли на пакет `extractors`. Примеры этого раздела документируют beta-API — при апгрейде они перестанут компилироваться (есть CLI `fiber migrate --to v3`). Сам апгрейд желателен: в v3.0.0–v3.3.0 закрыты security-фиксы (host normalization, oversized bodies, cookie handling). Концепт trusted-proxy обязателен и в stable — меняется только имя поля.

### 6.3. Порядок middleware (не переставлять)

Порядок критичен для безопасности и корректности. Реальный порядок из `wiring_fiber.go`:

<good-example>

```go
app.Use(middleware.RequestID())                       // 1. request_id для всех логов/ошибок
app.Use(fiberrecover.New(fiberrecover.Config{         // 2. recover сразу после RequestID —
    EnableStackTrace: true,                            //    паника уже несёт request_id
}))
app.Use(observability.HTTPTraceMiddleware())          // 3. трейсинг
app.Use(middleware.SecurityHeaders())                 // 4. security headers рано (X-Frame-Options, CSP, nosniff)
app.Use(middleware.Logger())                          // 5. структурный лог
app.Use(middleware.CORS())                            // 6. CORS
app.Get("/metrics", observability.MetricsEndpoint(cfg.Server.MetricsAllowedCIDRs))
app.Use(observability.HTTPMetricsMiddleware())        // 7. метрики
app.Use(middleware.RateLimit(d.rootCtx, middleware.APIRateLimit)) // 8. глобальный лимит 100/сек

api := app.Group("/api")
api.Use(auditmod.HTTPAuditMiddleware(d.auditSvc.RecordHTTPRequest)) // аудит всех /api
api.Use("/files", middleware.RateLimitByUser(d.rootCtx, middleware.FileUploadRateLimit))     // 10/мин
api.Use("/messenger", middleware.RateLimitByUser(d.rootCtx, middleware.MessageSendRateLimit)) // 60/мин
```

Каноническая ментальная модель порядка (согласована с `.cursor/rules/20-backend.mdc`):

```
RequestID → Recover → (Trace) → SecurityHeaders → Logger → CORS → (Metrics) → RateLimit → APIKeyAuth → routes
```

APIKeyAuth должен стоять **до** регистрации защищённых маршрутов (SEC-01): если поставить после — эндпоинты открыты. Это критичный инвариант аудита.
</good-example>

<bad-example>

RateLimit до RequestID → в логах rate-limit нет `request_id`; APIKeyAuth после маршрутов → защита не применяется. Оба — регрессия безопасности.
</bad-example>

### 6.4. Парсинг тела и query (Fiber v3 Bind)

В Fiber v3 используется `c.Bind()`: `c.Bind().JSON(&req)` для тела, `c.Bind().Query(&req)` для query-параметров.

<good-example>

Из `tracker/handler_issues.go`:

```go
var req CreateIssueRequest
if err := c.Bind().JSON(&req); err != nil {
    return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
}

// query:
var lreq ListIssuesRequest
if err := c.Bind().Query(&lreq); err != nil {
    return fiber.NewError(fiber.StatusBadRequest, "некорректные параметры запроса")
}
```
</good-example>

### 6.5. Валидация DTO (validate-теги)

DTO несут `json`-теги и `validate`-теги (пакет go-playground/validator). Ограничения: `required`, `min`/`max`, `oneof`, `omitempty`.

<good-example>

Реальные DTO (`tracker/dto.go`):

```go
type CreateIssueRequest struct {
    Title     string `json:"title" validate:"required,min=1,max=500"`
    IssueType string `json:"issue_type" validate:"required,oneof=bug task story epic sub_task improvement tech_debt spike release"`
    Priority  string `json:"priority,omitempty" validate:"omitempty,oneof=critical high medium low trivial"`
    StoryPoints *int `json:"story_points,omitempty" validate:"omitempty,oneof=1 2 3 5 8 13 21"`
}

type UpdateIssueRequest struct {
    Title   *string `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
    Version int     `json:"version" validate:"required"` // optimistic lock — обязателен
}

type ListIssuesRequest struct {
    Limit   int    `json:"limit,omitempty" query:"limit" validate:"omitempty,min=1,max=100"`
    OrderBy string `json:"order_by,omitempty" query:"order_by" validate:"omitempty,oneof=created_at priority updated_at"`
}
```

Обратите внимание на комбинацию тегов `json` + `query` для параметров, приходящих и в теле, и в строке запроса.
</good-example>

### 6.6. Группы маршрутов и регистрация

Маршруты группируются по ресурсу. Каждый модуль реализует `RegisterRoutes(api fiber.Router)` и монтируется в общую группу `/api`.

<good-example>

```go
func (h *Handler) RegisterRoutes(api fiber.Router) {
    projects := api.Group("/projects/:projectId")
    issues := projects.Group("/issues")
    issues.Post("/", h.CreateIssue)
    issues.Get("/", h.ListIssues)
    issues.Get("/:issueKey", h.GetIssue)
    issues.Put("/:issueKey", h.UpdateIssue)
    issues.Delete("/:issueKey", h.DeleteIssue)
    issues.Put("/:issueKey/position", h.MoveIssue)
    issues.Post("/jql", h.SearchIssuesJQL)
}
```
</good-example>

### 6.7. Извлечение userID и UUID-параметров

Используйте общие хелперы `platform/http`, а не ручной парсинг. `UserIDFromCtx` читает `c.Locals("userID")`, `ParseUUIDParam` валидирует UUID и отклоняет `uuid.Nil`.

<good-example>

Реальные хелперы (`platform/http/helpers.go`):

```go
func UserIDFromCtx(c fiber.Ctx) (uuid.UUID, error) {
    idStr, ok := c.Locals("userID").(string)
    if !ok || idStr == "" {
        return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
    }
    id, err := uuid.Parse(idStr)
    if err != nil {
        return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
    }
    return id, nil
}

func ParseUUIDParam(c fiber.Ctx, name string) (uuid.UUID, error) {
    idStr := utils.CopyString(c.Params(name)) // ME-12: копируем — fasthttp переиспользует буфер
    if idStr == "" {
        return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, name+" is required")
    }
    id, err := uuid.Parse(idStr)
    if err != nil {
        return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
    }
    if id == uuid.Nil {
        // Нулевой UUID не должен доходить до БД (непредсказуемое поведение).
        return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
    }
    return id, nil
}
```

Важно про fasthttp: `c.Params(...)` возвращает строку, ссылающуюся на переиспользуемый буфер. Если строку нужно сохранить дольше обработчика — копируйте (`utils.CopyString`). Это специфика Fiber/fasthttp, которой нет в `net/http`.
</good-example>

### 6.8. Возврат ошибок и статусов

- Успех: `c.Status(201).JSON(resp)`, `c.JSON(resp)` (200), `c.SendStatus(fiber.StatusNoContent)` (204).
- Ошибка: `return err` (доменный `*AppError` из service) — развернёт `ErrorHandler`. Для ранней HTTP-ошибки в handler — `fiber.NewError(fiber.StatusBadRequest, msg)`.

### 6.9. c.Context() vs c.Locals()

- `c.Context()` — `context.Context` запроса (для проброса в service). В Fiber v3 это стандартный `context.Context`.
- `c.Locals(key)` — request-scoped хранилище (userID, requestID), заполняется middleware.

### 6.10. Чек-лист Fiber

- [ ] Сигнатура handler'а — `func(c fiber.Ctx) error` (v3, значение).
- [ ] Порядок middleware не нарушен; APIKeyAuth до защищённых маршрутов.
- [ ] `EnableTrustedProxyCheck` + `BodyLimit` заданы в `fiber.Config`.
- [ ] Тело/query парсятся через `c.Bind().JSON/Query`.
- [ ] DTO имеют `validate`-теги (`required`, `oneof`, `min/max`).
- [ ] userID/UUID извлекаются через `httputil.UserIDFromCtx`/`ParseUUIDParam`.
- [ ] Строки из `c.Params` копируются, если переживают обработчик.
- [ ] `c.Context()` пробрасывается в service.

---

## 7. Безопасность

Security-инварианты Hub — не рекомендации, а жёсткие требования TZ (`docs/tz_constraints.md`) и результат аудита (12/12 CRITICAL исправлены). Их нарушение — блокер мержа. `be-ver` и `security-auditor` грепают именно по этим паттернам.

### 7.1. SSRF: safehttp для user-controlled URL

Любой исходящий запрос на URL, который может задать пользователь (GitLab/GitHub интеграции, WebPush endpoint, OpenGraph link preview, webhook), идёт через `safehttp.NewSafeHTTPClient` + предварительную проверку `ValidateOutboundURL`. **Запрещён** `http.DefaultClient`.

<good-example>

Реальная защита (`platform/http/safedialer.go`):

```go
// ValidateOutboundURL: только http/https, без userinfo, непустой host.
func ValidateOutboundURL(rawURL string) error {
    u, err := url.Parse(rawURL)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }
    scheme := strings.ToLower(u.Scheme)
    if scheme != "http" && scheme != "https" {
        return fmt.Errorf("%w: %s", ErrBlockedScheme, u.Scheme)
    }
    if u.User != nil {
        return ErrUserInfoNotAllowed
    }
    if u.Host == "" {
        return fmt.Errorf("invalid URL: empty host")
    }
    return nil
}

// SafeDialContext резолвит хост и отклоняет приватные/loopback/link-local IP
// (закрывает DNS-rebinding — резолв прямо перед подключением).
func NewSafeHTTPClient(timeout time.Duration) *http.Client {
    return &http.Client{
        Timeout: timeout,
        Transport: &http.Transport{
            DialContext:           SafeDialContext(),
            TLSHandshakeTimeout:   5 * time.Second,
            ResponseHeaderTimeout: 10 * time.Second,
            IdleConnTimeout:       60 * time.Second,
        },
    }
}
```

Заблокированы диапазоны: `10.0.0.0/8`, `127.0.0.0/8`, `169.254.0.0/16` (метаданные AWS `169.254.169.254`!), `172.16.0.0/12`, `192.168.0.0/16`, IPv6 ULA/link-local и др.

Использование в модуле:

```go
client := safehttp.NewSafeHTTPClient(10 * time.Second)
if err := safehttp.ValidateOutboundURL(userURL); err != nil {
    return apperrors.NewBadRequest("недопустимый URL")
}
resp, err := client.Get(userURL) // и dialer повторно проверит IP после резолва
```
</good-example>

<bad-example>

```go
// SSRF: пользователь укажет http://169.254.169.254/latest/meta-data/ и утащит креды инстанса.
resp, err := http.DefaultClient.Get(userURL)
```
</bad-example>

### 7.2. DoS: io.LimitReader вместо io.ReadAll

Чтение тела/ответа без лимита — memory DoS. Оборачивайте reader в `io.LimitReader`. Максимум для тела запроса align с Fiber `BodyLimit` (25 МБ); для файлов/бэкапов — свой cap.

<good-example>

```go
const maxBody = 25 << 20 // 25 MiB
limited := io.LimitReader(r.Body, maxBody)
data, err := io.ReadAll(limited) // теперь безопасно
```

Реальные места применения (`security_guidelines.md`): `backup/service.go` (`LimitReader` 512 MiB cap), `apigateway/webhook.go` (`io.Copy(io.LimitReader)` вместо `ReadFrom`), `platform/http/opengraph.go` (`LimitReader` на превью).
</good-example>

<bad-example>

```go
data, err := io.ReadAll(r.Body) // атакующий шлёт 10 ГБ → OOM
```
</bad-example>

### 7.3. Timing attack: subtle.ConstantTimeCompare

Сравнение секретов (magic-link коды, refresh-хеши, webhook-токены) — только `subtle.ConstantTimeCompare` (или `hmac.Equal`). **Запрещён** `!=` / `==` для секретов.

<good-example>

Реально (`auth/magiclink.go`, SEC-07):

```go
stored, err := s.redis.Get(ctx, codeKey).Result()
// …
// SEC-07: constant-time compare — защита от timing side-channel при подборе кода.
if subtle.ConstantTimeCompare([]byte(stored), []byte(code)) != 1 {
    s.recordFailure(ctx, failsKey)
    return ErrMagicLinkInvalid
}
```

И для refresh-хеша (`auth/service_session.go`):

```go
oldHash := hashRefreshToken(oldRefreshToken)
if subtle.ConstantTimeCompare([]byte(sess.RefreshToken), []byte(oldHash)) != 1 {
    return fmt.Errorf("refresh token: hash mismatch")
}
```
</good-example>

<bad-example>

```go
if stored != code { // timing side-channel: время сравнения зависит от совпавшего префикса
    return ErrMagicLinkInvalid
}
```
</bad-example>

### 7.4. Webhook: подпись обязательна, skip запрещён

Проверка подписи вебхука (LiveKit, GitHub, GitLab) не может быть пропущена — даже при отсутствии claim реджектим (SEC-02). Fail-closed.

<good-example>

```go
// ПРАВИЛЬНО — reject if hash claim missing.
expectedHash, ok := claims["sha256"].(string)
if !ok || expectedHash == "" {
    return fiber.ErrUnauthorized
}
// затем hmac.Equal(computed, expected)
```
</good-example>

<bad-example>

```go
// ЗАПРЕЩЕНО — если подписи нет, пропускаем проверку (обход аутентичности).
if expectedHash != "" {
    // verify…
}
// иначе — принимаем как валидный
```
</bad-example>

### 7.5. Авторизация: RBAC/membership на каждом чувствительном действии

Каждый endpoint, отдающий или меняющий данные, проверяет право пользователя (system role / project role / membership). Особо: WS-подписки и поиск.

**WebSocket (NF-02 / SEC-H1):** подписка на комнату проходит `Authorizer.CanSubscribe` — неизвестный префикс fail-closed.

<good-example>

Реальный `ws/authorizer.go`:

```go
func (a *configAuthorizer) CanSubscribe(ctx context.Context, userIDStr, room string) (bool, error) {
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        return false, fmt.Errorf("ws authz: invalid user id %q: %w", userIDStr, err)
    }
    prefix, idStr, ok := splitRoom(room) // "channel:{id}" → ("channel", id)
    if !ok {
        return false, fmt.Errorf("ws authz: malformed room %q", room)
    }
    entityID, _ := uuid.Parse(idStr)
    switch prefix {
    case "channel":
        if a.cfg.IsChannelMember == nil {
            return false, nil // fail-closed
        }
        return a.cfg.IsChannelMember(ctx, entityID, userID)
    case "call":
        if a.cfg.IsCallParticipantOrChannelMember == nil {
            return false, nil
        }
        return a.cfg.IsCallParticipantOrChannelMember(ctx, entityID, userID)
    case "project":
        if a.cfg.IsProjectMember == nil {
            return false, nil
        }
        return a.cfg.IsProjectMember(ctx, entityID, userID)
    default:
        return false, nil // неизвестный префикс — запрет
    }
}
```
</good-example>

**Поиск (NF-03 / SEC-H2):** любой поиск фильтруется по данным, доступным пользователю. Никогда не отдавайте глобальную выборку.

<good-example>

Из `search/service.go`: `searchMessages` фильтрует `channel_id ∈ getUserChannelIDs(userID)`; issues/projects/files — по `projectIDs` пользователя.

```go
// Mongo-фильтр поиска сообщений — только каналы пользователя.
filter := bson.M{"channel_id": bson.M{"$in": channelIDs}}
```
</good-example>

<bad-example>

```go
// Утечка: поиск по всем сообщениям без скоупа пользователя.
filter := bson.M{"$text": bson.M{"$search": q}}
```
</bad-example>

### 7.6. Rate limiting

Лимиты (TZ mandatory) — предопределены в `platform/middleware/ratelimit.go`:

| Тип | Лимит | Конфиг |
|-----|-------|--------|
| Auth | 3 / 10 мин | `AuthRateLimit` |
| API general | 100 / сек | `APIRateLimit` |
| File upload | 10 / мин | `FileUploadRateLimit` |
| Messages | 60 / мин | `MessageSendRateLimit` |
| AI | 20 / мин | `AIRateLimit` |

<good-example>

Реально (`ratelimit.go`):

```go
var (
    AuthRateLimit        = RateLimitConfig{Max: 3, Window: 10 * time.Minute}
    APIRateLimit         = RateLimitConfig{Max: 100, Window: 1 * time.Second}
    FileUploadRateLimit  = RateLimitConfig{Max: 10, Window: 1 * time.Minute}
    MessageSendRateLimit = RateLimitConfig{Max: 60, Window: 1 * time.Minute}
)
```

Для чувствительных путей (auth, magic-link verify) — Redis-лимитер `RateLimitRedis` (устойчив к перезапуску и горизонтальному масштабированию, ключ по email через `ExtractEmailFromBody`, а не только по IP). При сбое Redis — fallback на in-memory (fail-safe, не fail-open).
</good-example>

### 7.7. Аутентификация: Magic Link + JWT + refresh

Модель Hub (ADR-AUTH-01, TZ):

- **Только Magic Link (email).** Ни паролей, ни SSO/OAuth/LDAP. Telegram — только notify bind, не login.
- **JWT access — 15 минут TTL**, HS256, claims `uid`/`sid`.
- **Refresh — 30 дней, в Redis**, мгновенный отзыв (удаление ключа), ротация при каждом refresh.
- **Refresh at rest хешируется SHA-256** (NF-12) — в Redis и в записи сессии хранится хеш, не сам токен.
- **Rate limit auth — 3/10мин.**

<good-example>

Хеширование refresh (`auth/service_session.go`, NF-12):

```go
func hashRefreshToken(token string) string {
    sum := sha256.Sum256([]byte(token))
    return hex.EncodeToString(sum[:])
}

func refreshRedisKey(plaintext string) string {
    return "refresh:" + hashRefreshToken(plaintext)
}
```

Генерация access-токена с TTL 15 минут (accessTTL передаётся из config):

```go
func (s *SessionService) generateAccessToken(userID, sessionID uuid.UUID) (string, error) {
    now := time.Now().UTC()
    claims := JWTClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)), // 15m
            IssuedAt:  jwt.NewNumericDate(now),
            Issuer:    "hub",
        },
        UserID:    userID.String(),
        SessionID: sessionID.String(),
    }
    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}
```

Валидация токена проверяет метод подписи (защита от `alg=none`):

```go
token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (any, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    return s.jwtSecret, nil
})
```

Ротация refresh (порядок ME-02 против TOCTOU): (1) генерируем новый → (2) сохраняем новый в Redis → (3) обновляем хеш в сессии → (4) удаляем старый:

```go
newRefresh := uuid.New().String()
newHash := hashRefreshToken(newRefresh)
accessToken, _ := s.generateAccessToken(userID, sessionID)
if err := s.storeRefreshToken(ctx, newRefresh, *payload); err != nil { /* … */ }
if err := s.repo.UpdateSessionRefreshHash(ctx, sessionID, newHash, time.Now().UTC()); err != nil { /* … */ }
s.deleteRefreshToken(ctx, oldRefreshToken) // мгновенный отзыв старого
```
</good-example>

### 7.8. Секреты и шифрование at rest

- Секреты — только из env (`.env` в `.gitignore`, `.env.example` без значений). Никаких hardcoded creds в репо (INF-03).
- Токены интеграций (GitHub/GitLab) шифруются AES-GCM (`platform/crypto/aesgcm.go`).
- `DEV_MODE=true` запрещён в CI/prod (INF-02).

### 7.9. Валидация и санитизация входа

| Вход | Правило |
|------|---------|
| UUID-параметры | `httputil.ParseUUIDParam` (+ отклонение `uuid.Nil`) |
| Тело запроса | `c.Bind().JSON` + validate-теги, лимит 25 МБ |
| User URLs | `safehttp.ValidateOutboundURL` перед fetch |
| Webhook body | обязательная подпись + хеш |
| Имена файлов | `sanitizeFilename` (path traversal, 200-char cap) |
| Mongo-поиск | фильтр по каналам/проектам пользователя |

### 7.10. Чек-лист security (грепается be-ver)

- [ ] Нет `http.DefaultClient` для user-URL → `safehttp.NewSafeHTTPClient` + `ValidateOutboundURL`.
- [ ] Нет `io.ReadAll` без `io.LimitReader`.
- [ ] Сравнение секретов — `subtle.ConstantTimeCompare` / `hmac.Equal`, не `!=`.
- [ ] Проверка подписи вебхука не пропускается (fail-closed).
- [ ] RBAC/membership на каждом чувствительном endpoint.
- [ ] WS-подписка проходит `CanSubscribe`; неизвестный префикс — запрет.
- [ ] Поиск/выборка скоупятся по данным пользователя.
- [ ] Rate-limit применён (auth 3/10м, api 100/сек, файлы 10/мин, сообщения 60/мин).
- [ ] Refresh хешируется (SHA-256); access TTL 15м; отзыв мгновенный.
- [ ] Нет hardcoded секретов; нет `DEV_MODE=true` в prod-путях.
- [ ] Секреты не логируются.

---

## 8. NATS JetStream и outbox

### 8.1. NATS — единственный pub/sub

NATS + JetStream — единственная шина событий Hub. **Redis pub/sub использовать запрещено.** Core NATS (`Publish`) — для эфемерных real-time сигналов (typing, presence); JetStream — для durable-событий, которые должны пережить рестарт (уведомления, webhook-диспатч, аудит).

### 8.2. Клиент и стрим

<good-example>

Реальный `platform/nats/client.go`:

```go
const StreamName = "EVENTS" // единый JetStream stream для всех модулей

type Client struct {
    Conn      *nats.Conn
    JetStream jetstream.JetStream
}

func (c *Client) EnsureStreams(ctx context.Context) error {
    _, err := c.JetStream.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
        Name: StreamName,
        Subjects: []string{
            "project.>", "tracker.>", "messenger.channel.>",
            "tms.>", "calls.channel.>", "calls.recording.>",
            "auth.session.>", "sprint.>", "notification.>",
        },
        Retention: jetstream.InterestPolicy,
        MaxAge:    7 * 24 * time.Hour,
    })
    return err
}
```

Core-публикация (fire-and-forget, nil-safe):

```go
func (c *Client) PublishJSON(subject string, payload interface{}) {
    if c == nil {
        slog.Debug("NATS не подключён, пропускаем публикацию", "subject", subject)
        return
    }
    data, err := json.Marshal(payload)
    if err != nil {
        slog.Error("не удалось сериализовать событие", "subject", subject, "error", err)
        return
    }
    if err := c.Conn.Publish(subject, data); err != nil {
        slog.Error("не удалось опубликовать событие", "subject", subject, "error", err)
    }
}
```
</good-example>

### 8.3. Соглашение об subjects

Формат subject: `hub.<domain>.<event>` в общей ментальной модели; в реальном коде используются доменно- specific иерархии, совместимые с фильтрами стрима:

- `project.<projectID>.issue.created` / `.updated` / `.moved` / `.assigned`
- `tracker.<...>.issue.>`
- `messenger.channel.<id>.message_sent`
- `auth.session.created`
- `notification.>`, `sprint.>`, `tms.run.completed.>`

Иерархия с точками позволяет подписчикам фильтровать по wildcard (`project.*.issue.>`). Именование должно быть предсказуемым: `<домен>.<сущность>.<действие в прошедшем времени>`.

<good-example>

Реальные фильтры webhook-консьюмера (`nats/client.go`):

```go
var WebhookFilterSubjects = []string{
    "project.*.issue.>",
    "tracker.*.issue.>",
    "project.created",
    "tms.run.completed.>",
    "sprint.>",
    "messenger.channel.*.message_sent",
}
```

Заметьте: `notification.*`, `auth.session.*`, `calls.*` намеренно исключены из webhook-фильтра — они не доставляются наружу (APIGW-03).
</good-example>

### 8.4. Outbox-паттерн (transactional outbox)

Прямая публикация в NATS внутри бизнес-операции опасна: если БД закоммитилась, а публикация упала (или наоборот) — рассинхрон. Решение — **transactional outbox**: событие пишется в таблицу `outbox_event` в ТОЙ ЖЕ транзакции, что и данные; отдельный worker publish'ит его в NATS с ретраями (at-least-once).

<good-example>

Публикатор (`platform/outbox/publisher.go`):

```go
type Publisher struct{}

// PublishInTx: событие в outbox внутри существующей Ent-транзакции — атомарно с данными.
func (p *Publisher) PublishInTx(ctx context.Context, tx *ent.Tx, eventType string, payload any) error {
    data, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    _, err = tx.OutboxEvent.Create().
        SetEventType(eventType).
        SetPayload(data).
        Save(ctx)
    return err
}

// Publish: без внешней транзакции (автономно).
func (p *Publisher) Publish(ctx context.Context, client *ent.Client, eventType string, payload any) error {
    data, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    _, err = client.OutboxEvent.Create().SetEventType(eventType).SetPayload(data).Save(ctx)
    return err
}
```

Репозиторий использует `PublishInTx` внутри `client.Tx` (см. §5.5 `CreateIssue`). Интерфейс `OutboxPublisher` в `tracker/repository.go`:

```go
type OutboxPublisher interface {
    PublishInTx(ctx context.Context, tx *ent.Tx, eventType string, payload any) error
    Publish(ctx context.Context, client *ent.Client, eventType string, payload any) error
}
```
</good-example>

### 8.5. Outbox worker: at-least-once + ретраи + изоляция сбоев

<good-example>

Worker (`platform/outbox/worker.go`) — polling каждые 100мс, до 50 событий за батч, до 10 попыток, локальный recover на каждое событие:

```go
func (w *Worker) processBatch(ctx context.Context) {
    events, err := w.client.OutboxEvent.Query().
        Where(
            outboxevent.StatusEQ(outboxevent.StatusPending),
            outboxevent.AttemptsLT(10),
        ).
        Order(ent.Asc(outboxevent.FieldCreatedAt)).
        Limit(50).
        All(ctx)
    if err != nil {
        slog.Error("outbox: failed to query events", "error", err)
        return
    }
    for _, evt := range events {
        w.processEvent(ctx, evt) // recover внутри — сбой одной записи не валит батч
    }
}

func (w *Worker) processEvent(ctx context.Context, evt *ent.OutboxEvent) {
    defer func() {
        if r := recover(); r != nil {
            slog.Error("outbox: panic while processing event",
                "event_id", evt.ID, "event_type", evt.EventType,
                "panic", r, "stack", string(debug.Stack()))
        }
    }()

    if err := w.nats.Publish(evt.EventType, evt.Payload); err != nil {
        // Инкремент попытки + запись last_error, событие попадёт в следующую итерацию.
        w.client.OutboxEvent.UpdateOneID(evt.ID).
            SetAttempts(evt.Attempts + 1).
            SetLastError(err.Error()).
            Save(ctx)
        return
    }
    // Успех — помечаем published.
    w.client.OutboxEvent.UpdateOneID(evt.ID).
        SetStatus(outboxevent.StatusPublished).
        SetPublishedAt(time.Now()).
        Save(ctx)
}
```

Семантика: **at-least-once**. Если publish прошёл, но пометить `published` не удалось — событие пере-опубликуется. Это осознанный выбор: «лучше доставить дважды, чем потерять».
</good-example>

**Масштабирование и трейсинг worker'а (2026):**

- **2+ экземпляра worker'а — только с `SELECT ... FOR UPDATE SKIP LOCKED`.** Текущий поллинг `StatusPending` без блокировок корректен ровно для одного экземпляра; при горизонтальном масштабировании relay-воркеров два инстанса заберут один батч и опубликуют дубли. В Ent — выборка батча внутри транзакции с `query.ForUpdate(sql.WithLockAction(sql.SkipLocked))`.
- **Publish-side дедуп через `Nats-Msg-Id`.** Worker при публикации ставит заголовок `Nats-Msg-Id` = ID outbox-события; JetStream с настроенным `Duplicates`-окном стрима сам отбрасывает повторную публикацию после ретрая worker'а. Практический «exactly-once» = Msg-Id дедуп + идемпотентный консьюмер (§8.6).
- **Trace context через событие.** Span context сериализуется в колонку outbox-таблицы; worker инжектит `traceparent`/`tracestate` в NATS-заголовки (`TextMapPropagator` + header carrier), консьюмер извлекает их при обработке. Без этого трейс рвётся на границе producer → outbox → consumer.

### 8.6. Идемпотентность консьюмеров

Раз доставка at-least-once, консьюмеры обязаны быть идемпотентны: одно и то же событие может прийти дважды. Способы:

- Дедупликация по идентификатору события/сущности (обработали `issue.ID` + `event_type` — пропускаем повтор).
- Upsert вместо insert там, где возможно.
- Естественная идемпотентность (установка статуса в конкретное значение).
- Publish-side дополнение (2026): заголовок `Nats-Msg-Id` + `Duplicates`-окно стрима (§8.5) отсеивает повторные публикации ещё на стороне JetStream — но consumer-side дедуп это не отменяет (redelivery после потерянного ack всё равно возможен).

<good-example>

```go
// Консьюмер уведомлений: дедуп по (event_id) через Redis SET NX с TTL.
ok, err := rdb.SetNX(ctx, "notif:seen:"+eventID, 1, 24*time.Hour).Result()
if err != nil {
    return err // временный сбой — пусть JetStream доставит снова
}
if !ok {
    return nil // уже обрабатывали — no-op, ack
}
// … доставить уведомление …
```
</good-example>

<bad-example>

```go
// Не идемпотентно: при повторной доставке пользователь получит два одинаковых пуша.
sendPush(ctx, userID, msg)
```
</bad-example>

### 8.7. Durable consumers

Каждый модуль-потребитель (audit, webhook-dispatcher, notification) создаёт свой durable consumer на стриме `EVENTS`. Durable = позиция чтения переживает рестарт. Эфемерные подписки (`calls.chat.>`) — для real-time без реплея.

### 8.8. Чек-лист NATS/outbox

- [ ] Используется NATS/JetStream, не Redis pub/sub.
- [ ] Subject именуется предсказуемо: `<домен>.<сущность>.<действие>` (глагол в прошедшем времени).
- [ ] Событие, критичное к консистентности, пишется через `PublishInTx` в той же транзакции, что и данные.
- [ ] Эфемерные real-time сигналы — через core `Publish`/`PublishJSON` (nil-safe).
- [ ] Консьюмер идемпотентен (дедуп/upsert) — доставка at-least-once.
- [ ] Долгоживущие консьюмеры — durable; позиция переживает рестарт.
- [ ] Payload сериализуется в JSON; структура события в `events.go` модуля.
- [ ] (2026) При 2+ экземплярах outbox-worker'а — `SKIP LOCKED`; публикация с `Nats-Msg-Id`; trace context пропагируется через NATS-заголовки.

---

## 9. MongoDB

### 9.1. Где используется MongoDB

PostgreSQL — основная реляционная БД (через Ent). MongoDB 8 используется для документо-ориентированных данных с гибкой схемой и высоким объёмом:

- **messenger** — сообщения чатов (высокая частота записи, гибкая структура, cursor-пагинация).
- **calls** — чат внутри звонков.
- **TMS** — Allure/JUnit отчёты (вложенные структуры результатов тестов).

Драйвер — официальный `go.mongodb.org/mongo-driver/v2`.

### 9.2. Подключение

<good-example>

Реальный `platform/database/mongodb.go`:

```go
func NewMongoClient(ctx context.Context, url, dbName string) (*mongo.Database, error) {
    client, err := mongo.Connect(options.Client().ApplyURI(url))
    if err != nil {
        return nil, err
    }
    if err := client.Ping(ctx, nil); err != nil {
        return nil, err
    }
    slog.Info("connected to MongoDB", "database", dbName)
    return client.Database(dbName), nil
}
```
</good-example>

### 9.3. Паттерн репозитория для Mongo

Тот же принцип слоёв: репозиторий инкапсулирует `*mongo.Collection`, service работает через интерфейс. Mongo-специфика (`bson.M`, курсоры) не протекает в service.

<good-example>

```go
type MessageRepository interface {
    InsertMessage(ctx context.Context, m *Message) (string, error)
    ListByChannel(ctx context.Context, channelID uuid.UUID, before string, limit int) ([]*Message, error)
}

type MongoMessageRepository struct {
    col *mongo.Collection
}

func NewMongoMessageRepository(db *mongo.Database) *MongoMessageRepository {
    return &MongoMessageRepository{col: db.Collection("messages")}
}

func (r *MongoMessageRepository) ListByChannel(ctx context.Context, channelID uuid.UUID, before string, limit int) ([]*Message, error) {
    filter := bson.M{"channel_id": channelID.String()}
    if before != "" {
        // Cursor-пагинация: сообщения старше курсора.
        filter["_id"] = bson.M{"$lt": before}
    }
    opts := options.Find().
        SetSort(bson.D{{Key: "_id", Value: -1}}). // новые сверху
        SetLimit(int64(limit))
    cur, err := r.col.Find(ctx, filter, opts)
    if err != nil {
        return nil, fmt.Errorf("find messages: %w", err)
    }
    defer cur.Close(ctx)

    var out []*Message
    if err := cur.All(ctx, &out); err != nil {
        return nil, fmt.Errorf("decode messages: %w", err)
    }
    return out, nil
}
```

Всегда `defer cur.Close(ctx)`; всегда `ctx` в каждой операции; ошибки оборачиваются `%w`.
</good-example>

### 9.4. Скоуп по пользователю (security)

Как и в §7.5 — поиск/выборка по Mongo фильтруется данными пользователя. Никогда не отдавайте коллекцию без фильтра принадлежности.

<good-example>

```go
// Поиск сообщений — только каналы, где пользователь состоит (NF-03).
filter := bson.M{
    "channel_id": bson.M{"$in": userChannelIDs},
    "$text":      bson.M{"$search": query},
}
```
</good-example>

### 9.5. Индексы

Индексы создаются программно при инициализации (idempotent через `CreateOne`/`CreateMany`). Индексируйте поля фильтра/сортировки: `channel_id`, `created_at`, текстовые индексы для поиска.

<good-example>

```go
_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
    {Keys: bson.D{{Key: "channel_id", Value: 1}, {Key: "_id", Value: -1}}}, // листинг по каналу + cursor
    {Keys: bson.D{{Key: "content", Value: "text"}}},                        // полнотекстовый поиск
})
if err != nil {
    return fmt.Errorf("create message indexes: %w", err)
}
```
</good-example>

### 9.6. UUID и время в Mongo

- UUID сущностей Postgres хранятся как строки (`channelID.String()`) для консистентного матчинга.
- Время — UTC (как и везде). Не полагайтесь на локальную зону клиента при записи.

### 9.7. Чек-лист MongoDB

- [ ] Доступ через репозиторий-интерфейс; `*mongo.Collection` не протекает в service.
- [ ] `ctx` в каждой операции; `defer cur.Close(ctx)` для курсоров.
- [ ] Выборка скоупится по данным пользователя (channel/project IDs).
- [ ] Индексы под фильтры/сортировки/поиск созданы.
- [ ] Ошибки оборачиваются `%w`.
- [ ] Cursor-пагинация для лент (чат), не offset.

---

## 10. Конкурентность

### 10.1. Общий принцип

Go-конкурентность мощна и опасна. Правило Hub: любой код с горутинами/каналами тестируется с `-race`, а разделяемое состояние защищается либо каналами (передача владения), либо `sync`-примитивами. Регрессия обязательно включает `-race`.

### 10.2. Горутины и жизненный цикл

Каждая запущенная горутина должна иметь понятный способ завершения (обычно `ctx.Done()`). Горутина без выхода — утечка.

<good-example>

Из `platform/middleware/ratelimit.go` — cleanup завершается по контексту:

```go
func (rl *rateLimiter) cleanup(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return // graceful shutdown, без утечки
        case <-ticker.C:
            // очистка
        }
    }
}
```
</good-example>

<bad-example>

```go
go func() {
    for {
        time.Sleep(5 * time.Minute)
        cleanup() // нет способа остановить — горутина живёт вечно
    }
}()
```
</bad-example>

### 10.3. Сериализация состояния через одну горутину

Мощный приём против гонок: все мутации разделяемой структуры проходят через одну горутину, принимающую команды по каналам. Так сделан WS Hub.

<good-example>

Из `ws/hub.go` — `register`/`unregister`/`subscribe`/`unsubscribe` сериализуются в одной `Run`-горутине:

```go
func (h *Hub) Run(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            h.beginShutdown()
            h.shutdownAllClients()
            return
        case client := <-h.register:
            h.registerClient(client)     // мутация карт — только здесь
        case client := <-h.unregister:
            h.removeClient(client)
        case req := <-h.subscribe:
            h.subscribeClient(req.client, req.room)
        case req := <-h.unsubscribe:
            h.unsubscribeClient(req.client, req.room)
        }
    }
}
```

Публичные методы просто шлют в канал (send-and-forget), не мутируя состояние напрямую:

```go
func (h *Hub) Subscribe(client *Client, room string) {
    select {
    case h.subscribe <- subscribeReq{client: client, room: room}:
    case <-h.stopping: // не блокируемся после shutdown
    }
}
```
</good-example>

### 10.4. Закрытие каналов: правило SEC-05

Самая частая паника в Go-WS — «send on closed channel». Правило Hub (SEC-05): **закрывает канал только одна сторона, и это не та сторона, что пишет.** В Hub канал `send` НЕ закрывается никогда; сигнал завершения идёт через отдельный канал `done`, а все продюсеры проверяют `<-done` перед записью.

<good-example>

Из `ws/hub.go` — модель клиента и продюсеры:

```go
type Client struct {
    send      chan []byte
    done      chan struct{}
    closeOnce sync.Once
    // …
}

// signalClientDone закрывает ТОЛЬКО done (идемпотентно через closeOnce).
func (h *Hub) signalClientDone(client *Client) {
    client.closeOnce.Do(func() {
        close(client.done)
    })
}

// Продюсер: перед записью в send проверяет done через select.
func (h *Hub) BroadcastToRoom(room string, msg ServerMessage) {
    data, _ := json.Marshal(msg)
    val, ok := h.rooms.Load(room)
    if !ok {
        return
    }
    members := val.(*sync.Map)
    members.Range(func(_, value any) bool {
        client := value.(*Client)
        select {
        case <-client.done:
            // Клиент завершается — не отправляем (no-op, не паника).
        case client.send <- data:
            // Отправлено.
        default:
            // Буфер переполнен — drop, чтобы не блокировать Hub (NF-13, fail-soft).
            slog.Warn("ws client send buffer full, dropping message",
                "userID", client.UserID, "connID", client.ConnectionID, "room", room)
        }
        return true
    })
}
```

При удалении клиента (`removeClient`) закрывается только `done`, а `send` намеренно не закрывается — его соберёт GC:

```go
// SEC-05: закрываем ТОЛЬКО done. Канал send не закрывается, т.к. параллельные
// producer'ы пишут в него и проверяют done через select — иначе "send on closed channel".
h.signalClientDone(client)
```
</good-example>

<bad-example>

Классическая паника: receiver закрывает `send`, а параллельный broadcaster пишет в него:

```go
func (h *Hub) removeClient(c *Client) {
    close(c.send) // НЕЛЬЗЯ: другой горутине, которая сейчас делает c.send <- data, прилетит паника
}
```
</bad-example>

### 10.5. sync-примитивы

- `sync.Mutex` — для короткой защиты небольшого состояния (например `Client.rooms` в `ws`).
- `sync.Map` — для конкурентных карт с высоким чтением (реестр клиентов/комнат в `ws`).
- `sync.Once` — идемпотентное однократное действие (`closeOnce`, `stopOnce`).
- Не защищайте мьютексом то, что уже сериализовано одной горутиной (двойная защита = запутанность).

<good-example>

Из `ws/hub.go`:

```go
type Client struct {
    rooms map[string]bool
    mu    sync.Mutex // защищает rooms при доступе из разных горутин
}

func (h *Hub) subscribeClient(client *Client, room string) {
    client.mu.Lock()
    client.rooms[room] = true
    client.mu.Unlock()
    // реестр комнат Hub — sync.Map, безопасен для конкурентного доступа
    actual, _ := h.rooms.LoadOrStore(room, &sync.Map{})
    actual.(*sync.Map).Store(client.ConnectionID, client)
}
```
</good-example>

### 10.6. Неблокирующая отправка (backpressure)

Отправка в канал с `default` (как в §10.4) — стратегия backpressure: при переполнении буфера сообщение дропается, а не блокирует систему (NF-13, осознанный fail-soft для real-time). Для критичных данных используйте durable-путь (outbox/JetStream), а не WS.

### 10.7. Graceful shutdown

При `ctx.Done()`: перестаём принимать новую работу (`stopping` канал), дренируем/закрываем существующие соединения, завершаем фоновые горутины.

<good-example>

Из `ws/hub.go`:

```go
func (h *Hub) beginShutdown() {
    h.stopOnce.Do(func() { close(h.stopping) })
}

func (h *Hub) Register(client *Client) {
    if h.isStopping() {
        h.signalClientDone(client) // после shutdown — сразу сигналим done
        return
    }
    select {
    case h.register <- client:
    case <-h.stopping:
        h.signalClientDone(client)
    }
}
```
</good-example>

### 10.8. Тестирование гонок

Регрессия по конкурентному коду ВСЕГДА с `-race`:

```bash
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/ws/...
```

`-count=1` отключает кеш тестов; `-race` детектирует гонки данных; `-timeout` защищает от дедлоков.

Для детерминированного тестирования кода с таймерами/тикерами/фоновыми горутинами (outbox-worker, rate-limiter cleanup, WS hub) используйте `testing/synctest` (stable в Go 1.25): тест выполняется в изолированном «пузыре» с виртуальными часами, `synctest.Wait()` дожидается стабилизации горутин. Это убирает `time.Sleep` и flaky из конкурентных тестов — `-race` при этом остаётся обязательным. (2026)

### 10.9. Чек-лист конкурентности

- [ ] Каждая горутина завершается по `ctx.Done()` (нет утечек).
- [ ] Разделяемое состояние сериализовано (одна горутина) ИЛИ защищено `sync`.
- [ ] Каналы закрываются одной стороной; продюсеры не пишут в закрытый канал.
- [ ] Паттерн `done` + `select` для «обречённых» получателей (SEC-05).
- [ ] `sync.Once` для идемпотентного закрытия/остановки.
- [ ] Неблокирующая отправка (`default`) там, где нельзя блокировать шину.
- [ ] Graceful shutdown реализован (`stopping` + drain).
- [ ] Тесты прогоняются с `-race -count=1`.

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

## 12. Логирование и наблюдаемость

> Полный playbook — `.cursor/playbooks/observability.md`. Здесь — минимум для бэкенда.

### 12.1. slog + контекст

Логирование только структурное, через `slog`, context-aware (`InfoContext`/`WarnContext`/`ErrorContext`). Контекст несёт `request_id`/trace.

<good-example>

```go
slog.InfoContext(ctx, "issue created",
    "id", issue.ID,
    "project_id", projectID,
    "request_id", middleware.RequestIDFromCtx(ctx),
)
```

Реальный `request_id` берётся из middleware `RequestID` (первый в цепочке) и подхватывается `ErrorHandler` (`c.Locals("requestID")`).
</good-example>

<bad-example>

```go
fmt.Println("issue created")            // не структурно, без контекста
log.Printf("issue created %v", issue)   // legacy, теряет корреляцию
```
</bad-example>

### 12.2. Уровни

- `Error` — сбой, требующий внимания (не удалось записать в БД, паника воркера).
- `Warn` — деградация/fallback (Redis pipeline упал → in-memory; буфер WS переполнен → drop).
- `Info` — значимые бизнес-события (сессия создана, стрим готов).
- `Debug` — детально для разработки (клиент WS зарегистрирован, отмена клиентом).

### 12.3. Что нельзя логировать

Никогда: полный JWT, magic-link коды (в проде), refresh-токены, webhook-секреты, содержимое тел с PII. В dev-режиме код magic-link логируется намеренно (`IsDevMode`), в проде — нет.

<good-example>

Из `auth/magiclink.go` — код в лог только в dev:

```go
if s.resend != nil {
    if err := s.resend.SendMagicLinkEmail(ctx, email, code); err != nil {
        slog.Error("не удалось отправить magic link email", "email", email, "error", err)
        return code, nil
    }
} else {
    slog.Info("magic link code (dev mode)", "email", email, "code", code) // только dev
}
```
</good-example>

### 12.4. Метрики и трейсинг

HTTP-метрики и трейсинг подключены middleware (`observability.HTTPMetricsMiddleware`, `HTTPTraceMiddleware`); `/metrics` доступен по allowlist CIDR. `observability.Timed` даёт latency доменных операций. Для прод-нагрузки трейсы сэмплируются — `ParentBased(TraceIDRatioBased(N))` вместо дефолтного AlwaysSample (100% трейсинг под нагрузкой дорог, 2026). Детали — в observability-playbook.

### 12.5. Чек-лист логирования

- [ ] Используется `slog` с `*Context`-вариантами, не `fmt.Println`/`log.Printf`.
- [ ] В логах есть `request_id`/контекст для корреляции.
- [ ] Секреты не логируются (JWT/refresh/коды/токены/PII).
- [ ] Правильные уровни (Error/Warn/Info/Debug).
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

## 14. Комментарии в коде

> Полный playbook — `.cursor/playbooks/code-comments.md`. Здесь — правило и минимум.

### 14.1. WHY, не WHAT

Комментарий объясняет намерение, компромисс или неочевидное ограничение — то, что код сам сказать не может. Комментарии-нарраторы («// increment counter») запрещены. Эмодзи в коде/комментариях запрещены. Язык комментариев в Go — английский; проза в docs — русский (в существующем коде встречаются и русские комментарии — следуйте стилю файла, но для нового кода предпочтителен English согласно `10-style.mdc`).

<good-example>

Комментарий объясняет ПОЧЕМУ (реально из `ws/hub.go`, `platform/http/helpers.go`):

```go
// SEC-05: закрываем ТОЛЬКО done. Канал send не закрывается, т.к. параллельные
// producer'ы пишут в него и проверяют done через select — иначе "send on closed channel".
h.signalClientDone(client)

// ME-12: отклоняем uuid.Nil — нулевой ID не должен попадать в запросы к БД,
// поведение движка для zero UUID может отличаться от 404.
if id == uuid.Nil {
    return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
}
```

Ссылки на ID находок аудита (SEC-05, ME-12, NF-13, HI-07) — хорошая практика Hub: связывают код с решением/аудитом.
</good-example>

<bad-example>

```go
i++            // increment i
// set title
create.SetTitle(input.Title)
// return the result
return resp, nil
```
</bad-example>

### 14.2. Godoc для экспортируемого

Каждый экспортируемый тип/функция/метод — с doc-комментарием, начинающимся с имени сущности.

<good-example>

```go
// SessionService отвечает за JWT, refresh-токены и жизненный цикл сессий (SC-14-1).
type SessionService struct { /* … */ }

// RefreshToken валидирует refresh-токен, выполняет ротацию и возвращает новую пару токенов.
func (s *SessionService) RefreshToken(ctx context.Context, oldRefreshToken string) (*TokenPairResponse, error)
```
</good-example>

### 14.3. Package doc

Каждый пакет — с package-комментарием в одном файле.

<good-example>

```go
// Package errors определяет структурированные ошибки приложения с HTTP-кодами и сообщениями.
package errors

// Package tracker реализует трекер задач: CRUD, Kanban-доски, workflow, комментарии, учёт времени.
package tracker
```
</good-example>

### 14.4. Чек-лист комментариев

- [ ] Комментарии объясняют WHY, нет нарраторов WHAT.
- [ ] Нет эмодзи в коде/комментариях.
- [ ] Экспортируемые сущности имеют godoc, начинающийся с их имени.
- [ ] У пакета есть package doc.
- [ ] Нетривиальные решения ссылаются на ID находки/ADR, где уместно.

---

## 15. Git и коммиты

> Полный playbook — `.cursor/playbooks/git-workflow.md`. Правила — `.cursor/rules/50-git.mdc`.

### 15.1. Формат коммита (Conventional Commits)

```
<type>(<scope>): <description>
```

- **type:** `feat` | `fix` | `refactor` | `test` | `docs` | `chore` | `perf` | `ci`
- **scope:** `auth` | `tracker` | `messenger` | `tms` | `calls` | `dashboard` | `notification` | `search` | `ws` | `platform` | `frontend` | `infra` | `ent`

<good-example>

```
feat(tracker): add WIP limits to Kanban columns
fix(ws): prevent send on closed channel panic
refactor(auth): extract AuthRepository interface
test(auth): add magic-link brute-force lockout test
```
</good-example>

### 15.2. Ветки

```
feat/<module>/<description>
fix/<module>/<description>
refactor/<module>/<description>
```

### 15.3. Что никогда не коммитить

`.env`, секреты, coverage dumps (`backend/cov*`, `*.out`), бинарники, `livekit-prod.yaml` с ключами. Перед коммитом — скан diff на секреты.

### 15.4. Pre-commit (бэкенд)

```bash
cd backend && go build -tags release ./... && go test -race -count=1 ./...
```

Скан diff'а на секреты — автоматизированный `gitleaks` (pre-commit + CI), а не только ручной `git diff --staged | grep`: grep даёт и пропуски, и шум. Секреты в git — подтверждённый риск Hub (PG-03, ротация). (2026)

### 15.5. Правила для be1/be2

PR другого разработчика ревьюится (кросс-ревью). PR ссылается на задачу плана. Squash-merge в `master`. PR с изменением общего интерфейса мержится первым (см. §1.3).

### 15.6. Чек-лист git

- [ ] Формат `type(scope): description`, scope из списка.
- [ ] Ветка по конвенции `feat|fix|refactor/<module>/...`.
- [ ] В diff нет секретов/dumps/бинарников.
- [ ] Pre-commit регрессия зелёная.
- [ ] PR отревьюен вторым backend-разработчиком.

---

## 16. Обязательная регрессия и docs-sync

> Правило: `.cursor/rules/56-regression-verify.mdc`, `.cursor/rules/55-docs-sync.mdc`.

### 16.1. Минимальная регрессия после изменений

После любого изменения бэкенда — до handoff/PR — прогоняется минимальная регрессия затронутого слоя:

```bash
cd backend && go build -tags release ./... && go vet ./... && go test -race -count=1 -timeout=600s ./...
```

Каноничный (Docker) прогон тестов фреймворка:

```bash
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default \
  -v "${PWD}/backend:/app" -w /app \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./testframework/..."
```

Три обязательных шага перед handoff:

1. `go build -tags release ./...` — сборка проходит с prod-тегом (тестовый код под `hubtest` не должен ломать release).
2. `go vet ./...` — статический анализ (подозрительные конструкции, форматные строки, копирование мьютексов).
3. `go test -race ...` — тесты с детектором гонок.

Supply-chain шаги — обязательные CI-инструменты (2026):

4. `go mod verify` + `go mod tidy -diff` — целостность module cache и отсутствие подмешанных/устаревших записей `go.sum`; ранние шаги перед build, стоят секунды.
5. `govulncheck ./...` — официальный сканер уязвимостей Go team; call-graph анализ репортит только реально достижимые CVE (почти нулевой шум). Дополнительно — scheduled-прогон по `master` (новые advisories появляются без изменения кода); закрывает окно «CVE опубликована → Dependabot ещё не поднял версию».

Полный gate — slash `/qa-gate` (останавливается на первой FAIL). `verified: true` — только при зелёной регрессии с приложенным логом.

### 16.2. Docs-sync

После функционального изменения обновляется затронутая документация в той же сессии (правило 55). Матрица для бэкенда:

| Изменил | Обязательные документы |
|---------|------------------------|
| Любая фича/фикс | `.planning/STATE.md` (session note) |
| Новый/изменённый API route/контракт | `docs/system_specification.md`, `docs/wiki/API.md` |
| Тесты/coverage | `.planning/COVERAGE.md`, `docs/test_plan.md` |
| Security | `docs/security_guidelines.md` |
| CI/deploy/Docker | `docs/stack_and_tools.md`, `deployments/.env.example` |

**Новый endpoint → обязательно обновить `docs/system_specification.md` и `docs/wiki/API.md`.** Плюс swag-аннотации на самом handler'е (`@Summary`, `@Router`, `@Security BearerAuth`) — см. §6 и реальный `handler_issues.go`.

### 16.3. Test-triage при падении

При FAIL сначала сверить корректность функционала по `docs/` (skill `.cursor/skills/test-triage/SKILL.md`), затем решать: чинить код или тест. Не подгонять тест под баг.

### 16.4. Артефакты

После coverage/race-прогонов удалить локальные dumps (`*.out`, `backend/cov*`, `tmp_*`) — правило 45. Они в `.gitignore`, но проверьте `git status`.

### 16.5. Чек-лист регрессии

- [ ] `go build -tags release ./...` — зелёный.
- [ ] `go vet ./...` — без замечаний.
- [ ] `go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...` — зелёный.
- [ ] (2026) `go mod verify` + `govulncheck ./...` — без findings (в CI обязательны).
- [ ] Новый endpoint отражён в `system_specification.md` + `wiki/API.md` + swag-аннотации.
- [ ] `.planning/STATE.md` обновлён.
- [ ] Артефакты покрытия удалены; `git status` чист от dumps.

---

## 17. Definition of Done и что проверяет be-ver

### 17.1. Definition of Done бэкенд-задачи

Задача считается готовой к верификации, когда выполнено ВСЁ:

- [ ] **Слои.** Handler/Service/Repository/EntRepository разделены; `*ent.Client` только в `ent_repository.go`/`repo_*.go`.
- [ ] **DTO + валидация.** Есть Request/Response с `json`+`validate` тегами; вход валидируется.
- [ ] **Ошибки.** Доменные `*AppError`; обёртка `%w`; `ent.IsNotFound`→404; конфликт версии→409; нет голого `return err`; нет `panic()`.
- [ ] **Context.** `ctx` первым параметром; проброшен `c.Context()`; фоновые горутины по `ctx.Done()`.
- [ ] **Ent.** UUID PK; `version` на mutable; UTC; soft delete; транзакции с rollback; outbox `PublishInTx` в tx; нет N+1; миграция создана.
- [ ] **Fiber.** Порядок middleware соблюдён; APIKeyAuth до защищённых маршрутов; `Bind()` для парсинга.
- [ ] **Security.** safehttp для user-URL; `LimitReader`; `ConstantTimeCompare` для секретов; webhook-подпись; RBAC/скоуп; rate-limit; секреты не логируются.
- [ ] **События.** NATS/JetStream (не Redis pub/sub); subject по конвенции; критичные — через outbox; консьюмеры идемпотентны.
- [ ] **Конкурентность.** Нет утечек горутин; каналы закрываются корректно (SEC-05); `-race` зелёный.
- [ ] **Производительность.** Лимиты пагинации; индексы; `observability.Timed` на значимых операциях.
- [ ] **Логи.** `slog` context-aware; уровни; без секретов.
- [ ] **Тесты.** В `testframework/{domain}/` под `hubtest`; покрыты happy/ошибки/404/409/authz; `-race`.
- [ ] **Комментарии.** WHY-not-WHAT; godoc на экспортируемом; без эмодзи.
- [ ] **Регрессия.** `build -tags release` + `go vet` + `go test -race` — зелёные, лог приложен.
- [ ] **Docs.** `system_specification.md` + `wiki/API.md` при новых эндпоинтах; `STATE.md` обновлён.
- [ ] **Git.** Коммит `type(scope): ...`; diff без секретов; кросс-ревью.

### 17.2. Что именно проверяет be-ver

`be-ver` — readonly-верификатор. Его работа — не переписать, а вынести вердикт `verified: true/false` с доказательствами. Проверяет:

1. **Layering (грепом).** `rg "ent\.Client" backend/internal/*/handler*.go backend/internal/*/service*.go` — должно быть пусто (кроме легитимного `GetDB()`-доступа в поиске). `fiber` не импортируется в service.
2. **Security-инварианты (грепом).** `http.DefaultClient`, `io.ReadAll` без `LimitReader`, `!=` для секретов, `panic(`, пропуск webhook-подписи, поиск без скоупа.
3. **Обработка ошибок.** Нет голого `return err` на границах; `ent.IsNotFound`→404; конфликт версии→409.
4. **Context.** `ctx` первым; нет `context.Background()` в handler; нет ctx в структурах.
5. **Ent-инварианты.** `version` + `AddVersion(1)` + проверка `affected==0`; транзакции с rollback; outbox в tx; soft delete фильтр.
6. **Тесты и регрессия.** Тесты существуют и релевантны; лог `-race` зелёный и приложен; `go vet` чист.
7. **Data race.** Прогон `-race` не показал гонок; конкурентный код следует SEC-05.
8. **Регрессия/drift.** Изменение не сломало соседние модули; docs синхронизированы; нет забытых dumps в diff.

Вердикт `verified: false`, если найдена хотя бы одна CRITICAL/HIGH проблема из §7 или §18, либо регрессия красная/отсутствует.

### 17.3. Грепы be-ver (быстрый набор)

```bash
# *ent.Client вне репозитория
rg -n "\*ent\.Client|ent\.Client\{" backend/internal --glob '!**/*ent_repository.go' --glob '!**/repo_*.go'
# SSRF
rg -n "http\.DefaultClient" backend/internal
# DoS
rg -n "io\.ReadAll" backend/internal
# timing
rg -n "== *\"|!= *\"" backend/internal/auth
# panic в prod
rg -n "panic\(" backend/internal
# fiber в service
rg -n "gofiber/fiber" backend/internal/*/service*.go
```

Дополнение (2026): `gosec` (SAST — taint-анализ SQLi/SSRF/path traversal, слабое крипто) как CI-gate с SARIF-отчётом в GitHub code scanning автоматизирует часть этих грепов, но **не заменяет** их — ручной набор be-ver остаётся обязательным.

---

## 18. Антипаттерны Hub

Таблица «нельзя → почему → как правильно». Это сводный чек-лист самых частых нарушений, каждое из которых — потенциальный `verified: false`.

| Нельзя | Почему | Как правильно |
|--------|--------|---------------|
| `*ent.Client` в Handler/Service | Ломает layering, невозможно мокать | Repository-интерфейс; `*ent.Client` только в `ent_repository.go` |
| `http.DefaultClient` для user-URL | SSRF (метаданные облака, внутренняя сеть), нет таймаута | `safehttp.NewSafeHTTPClient` + `ValidateOutboundURL` |
| `io.ReadAll` без лимита | Memory DoS | `io.LimitReader(r, max)` (align 25 МБ / свой cap) |
| `!=`/`==` для секретов | Timing side-channel | `subtle.ConstantTimeCompare` / `hmac.Equal` |
| Пропуск webhook-подписи при отсутствии claim | Обход аутентичности (SEC-02) | Fail-closed: нет подписи → `401` |
| `panic()` в prod-пути | Падение процесса на вводе | `return error`; `recover` только в воркерах |
| Голый `return err` | Потеря контекста диагностики | `fmt.Errorf("op: %w", err)` |
| `return err` из БД как 500 на «не найдено» | Неверный статус | `ent.IsNotFound(err)` → `NewNotFound` (404) |
| Update без `version` | Потерянные обновления (lost update) | `VersionEQ` + `AddVersion(1)` + `affected==0`→409 |
| Прямая публикация в NATS в бизнес-операции | Рассинхрон данных и событий | Outbox: `PublishInTx` в той же транзакции |
| Redis pub/sub для событий | Нарушение TZ (единственный pub/sub — NATS) | NATS JetStream |
| `close(send)` со стороны получателя | «send on closed channel» panic | Закрывать только `done`; продюсеры `select <-done` (SEC-05) |
| `context.Background()` в handler | Теряется cancel/deadline запроса | `c.Context()` |
| `ctx` в поле структуры | Протухший контекст, маскировка утечек | `ctx` аргументом каждого вызова |
| N+1 (запрос связи в цикле) | Деградация под нагрузкой | Eager `.WithX()`; агрегаты `Count`/`GroupBy` |
| Листинг без верхнего лимита | Дамп таблицы, OOM | Жёсткий `Limit` (≤100), default 50 |
| Поиск/выборка без скоупа пользователя | Утечка чужих данных (NF-03) | Фильтр по channel/project IDs пользователя |
| WS-подписка без authz | Подписка на чужие комнаты (NF-02) | `Authorizer.CanSubscribe`, fail-closed |
| Логирование JWT/refresh/кодов | Утечка секретов | Не логировать; коды — только dev-mode |
| `field.String("id")` для PK | Нарушение TZ (UUID v4) | `field.UUID(...).Default(uuid.New).Immutable()` |
| Локальное время в БД | Рассинхрон TZ | UTC везде (`nowUTC`, `time.Now().UTC()`) |
| Hard delete по умолчанию | Потеря данных, нет корзины | Soft delete (`deleted_at`), hard — только Admin/GDPR |
| `DEV_MODE=true` в prod-путях | Обход origin-check/секретов | Только dev; запрещено в CI |
| Комментарии-нарраторы, эмодзи в коде | Шум, нарушение стиля | WHY-комментарии, godoc |
| Тест рядом в `internal/` | Ломает изоляцию prod-сборки | `testframework/{domain}/` под `hubtest` |

---

## 19. Быстрый чек-лист перед передачей be-ver

Единый финальный чек-лист. Пройдите его перед тем, как пометить задачу готовой к верификации. Если хоть один пункт красный — не передавайте.

### Архитектура и слои

- [ ] `*ent.Client` только в `ent_repository.go` / `repo_*.go`.
- [ ] Handler = parse → call → status; без бизнес-логики и БД.
- [ ] Service без `fiber` и без парсинга HTTP; зависит от `Repository`-интерфейса.
- [ ] Репозиторные методы принимают `uuid.UUID`/input-структуры; конвертация в `toXResponse`.

### Ошибки и context

- [ ] Нет голого `return err`; везде `%w` с контекстом операции.
- [ ] `ent.IsNotFound`→404; конфликт версии→409; Service возвращает `*AppError`.
- [ ] Нет `panic()` в prod (кроме `recover` в воркерах).
- [ ] `ctx` первым параметром; `c.Context()` в handler; нет ctx в структурах; фон завершается по `ctx.Done()`.

### Ent и данные

- [ ] UUID PK; `version` на mutable; UTC; soft delete + фильтр `DeletedAtIsNil`.
- [ ] Транзакции с rollback на каждой ошибке; commit-ошибка обёрнута.
- [ ] Optimistic lock: `VersionEQ` + `AddVersion(1)` + `affected==0`→409.
- [ ] Нет N+1 (`.WithX()`); листинги с лимитом ≤100.
- [ ] Схема изменена → Ent-код регенерирован + миграция Atlas создана.

### Fiber и API

- [ ] Порядок middleware не нарушен; APIKeyAuth до защищённых маршрутов.
- [ ] `Bind().JSON/Query`; DTO с `validate`-тегами.
- [ ] Swag-аннотации на новых endpoint'ах (`@Router`, `@Security BearerAuth`).

### Безопасность (жёсткие инварианты)

- [ ] User-URL → `safehttp` + `ValidateOutboundURL`; нет `http.DefaultClient`.
- [ ] `io.LimitReader` (нет безлимитного `io.ReadAll`).
- [ ] Секреты — `subtle.ConstantTimeCompare`; webhook-подпись не пропускается.
- [ ] RBAC/membership + скоуп пользователя (WS `CanSubscribe`, поиск по своим данным).
- [ ] Rate-limit применён; refresh хеширован (SHA-256); access TTL 15м.
- [ ] Секреты не логируются; нет hardcoded creds; нет `DEV_MODE=true` в prod.

### События и конкурентность

- [ ] NATS/JetStream (не Redis pub/sub); subject по конвенции; критичные — через outbox `PublishInTx`.
- [ ] Консьюмеры идемпотентны (at-least-once).
- [ ] Каналы закрываются корректно (только `done`, SEC-05); нет утечек горутин.

### Качество, тесты, регрессия

- [ ] Тесты в `testframework/{domain}/` под `hubtest`; покрыты happy/валидация/404/409/authz.
- [ ] `go build -tags release ./...` — зелёный.
- [ ] `go vet ./...` — чист.
- [ ] `go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...` — зелёный, лог приложен.
- [ ] `slog` context-aware; уровни; без секретов; `observability.Timed` на значимых операциях.
- [ ] Комментарии WHY-not-WHAT; godoc; без эмодзи.

### Docs и git

- [ ] Новый endpoint → `system_specification.md` + `wiki/API.md`.
- [ ] `.planning/STATE.md` обновлён; артефакты покрытия удалены.
- [ ] Коммит `type(scope): ...`; diff без секретов; PR отревьюен вторым backend-разработчиком.
- [ ] Координация be1/be2: общий интерфейс мержится первым; rebase на актуальный `master`.

---

## Приложение: дельты 2026 на обсуждение

Из сверки с индустрией 2026 (`.planning/research/BEST-PRACTICES-2026-backend.md`). В правила playbook не внесены — каждая требует решения arch/product или ADR.

| Тема | Суть | Статус |
|------|------|--------|
| CSRF: Sec-Fetch-Site | Go 1.25 `http.CrossOriginProtection` — header-based CSRF-защита без токенов; для Fiber — аналогичная проверка в middleware. Зона риска: cookie-flow (`/ws?token=`) и будущие OIDC-редиректы | Требует решения при ADR-AUTH-SSO (обязательный пункт security-ревью SSO) |
| Iterators (`iter.Seq`/`iter.Seq2`) | Range-over-func для стриминга больших выборок (экспорт, backup) вместо загрузки слайса в память; stdlib уже перешёл | Требует решения (принять для новых пайплайн-API; существующее не рефакторить) |
| Ent Interceptors для soft-delete | Глобальный traversal-фильтр `deleted_at` + tenant-scope на уровне ORM (fail-closed) вместо ручного `DeletedAtIsNil()` в каждом Query | Требует решения (сильный выигрыш по безопасности; нужен обход для корзины/GDPR) |
| Contract testing (Pact) | Consumer-driven contracts для REST + async message contracts для NATS-событий (или schema registry) | Требует решения — предусловие распила на сервисы (`MICROSERVICES-ROADMAP`) |
| OTel semantic conventions v1 | Стандартные имена span'ов/метрик (`http.request.method`, `db.system.name`, `messaging.system`) — совместимость с готовыми Grafana-дашбордами | Требует решения (аудит имён; переименование — breaking для текущих дашбордов) |
| PostgreSQL 17 `JSON_TABLE` | Разворачивание JSONB в реляционные строки прямо в SQL — для аналитики/дашбордов вместо app-level парсинга | Требует решения (через Ent — только raw SQL/Modify) |

---

*Playbook бэкенда Hub. Источник истины по паттернам; при расхождении с `docs/*` и `.cursor/rules/*` приоритет — у TZ-инвариантов (`60-tz-constraints.mdc`) и результатов аудита. Обновляйте при изменении стандартов.*

