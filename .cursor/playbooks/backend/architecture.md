# Backend: архитектура сервиса

> Часть плейбука. Канон и оглавление — [`../backend.md`](../backend.md).

Модули, границы, зависимости, точки расширения.

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
└── (genuine isolated units — в backend/tests/<module>/, НЕ здесь)
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
