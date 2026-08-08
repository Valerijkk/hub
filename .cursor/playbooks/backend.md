# Hub — Эталонный Playbook по бэкенд-разработке

> Единый справочник для AI-агентов бэкенда Hub: `be1`, `be2` (параллельная разработка фич) и `be-ver` (верификация).
> Стек: Go 1.24/1.25 · Fiber v3 (fasthttp) · Ent + Atlas · PostgreSQL 17 · MongoDB 8 · Redis 7 · NATS JetStream · WebSocket (Fiber) · LiveKit · SeaweedFS (S3 через aws-sdk-go-v2).
> Связанные правила: `.cursor/rules/20-backend.mdc`, `.cursor/rules/10-style.mdc`, `.cursor/rules/60-tz-constraints.mdc`.
> Связанные playbook'и: `.cursor/playbooks/observability.md`, `.cursor/playbooks/testing.md`, `.cursor/playbooks/code-comments.md`, `.cursor/playbooks/git-workflow.md`.

Этот документ — источник истины по паттернам бэкенда. Он не заменяет базу знаний `docs/*`, а конкретизирует её для написания Go-кода. Все примеры взяты из реального кода репозитория (модули `auth`, `tracker`, `platform/*`, `ws`, `outbox`) — имена типов, пакетов и функций аутентичны.

---

## Оглавление

Канон здесь. Технологические разделы — отдельными файлами: они нужны при
работе с конкретной технологией, а не на каждой задаче.

**В этом файле:** роль и объём, слои, контекст, тестирование и код-стайл,
Git, документация, Definition of Done, чек-лист перед `be-ver`.

| Файл | Когда открывать |
|------|-----------------|
| [`backend/architecture.md`](backend/architecture.md) | Проектируешь модуль: границы, зависимости, точки расширения |
| [`backend/security.md`](backend/security.md) | Аутентификация, авторизация, секреты, валидация входа |
| [`backend/observability.md`](backend/observability.md) | Логи, метрики, трассировка |
| [`backend/ent-orm.md`](backend/ent-orm.md) | Работаешь с Ent: схемы, миграции, запросы |
| [`backend/fiber-http.md`](backend/fiber-http.md) | Пишешь HTTP-слой на Fiber |
| [`backend/messaging.md`](backend/messaging.md) | NATS JetStream, outbox, MongoDB |
| [`backend/perf-testing.md`](backend/perf-testing.md) | Производительность и тестирование бэкенда |

Нумерация разделов сохранена — ссылки вида «§5» остаются рабочими.
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
| `be1` | Backend Dev, поток A | Реализует фичи бэкенда (handler/service/repository + тесты), делает кросс-ревью PR `be2` | `.cursor/agents/be1.md` / `be2.md` |
| `be2` | Backend Dev, поток B | Реализует фичи бэкенда параллельно с `be1`, делает кросс-ревью PR `be1` | `.cursor/agents/be1.md` / `be2.md` |
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

Перед коммитом выполняется task-owned limited Docker job `backend-unit` из
`AGENTS.md`: release build, vet и race-прогон `backend/tests/**`. Host-run
запрещён; evidence фиксирует source SHA, image, inner command, CPU/RAM,
collection, duration и cleanup.

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

После любого изменения бэкенда — до handoff/PR — запускается каноничный
task-owned limited Docker job `backend-unit` из `AGENTS.md`. Он проверяет
release build, vet и genuine isolated Go units только в `backend/tests/**`.
БД, broker/storage, HTTP/component/API и другие non-unit boundaries не
подмешиваются в этот job: их greenfield Python suites принадлежат
`testframework/{contract,api,integration,e2e,load}/**`.

Три обязательных шага перед handoff:

1. Release build внутри Docker — production graph собирается без тестовых зависимостей.
2. Vet внутри Docker — статический анализ чист.
3. Race-прогон `backend/tests/**` внутри Docker — non-zero collection и 0 data race.

Supply-chain шаги — обязательные CI-инструменты (2026):

4. Module integrity (`go mod verify` + tidy diff) выполняется внутри supply-chain Docker job.
5. `govulncheck` выполняется внутри закреплённого Docker/CI toolchain; scheduled-прогон по `master` закрывает окно новых advisories.

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

- [ ] Docker job `backend-unit` из `AGENTS.md` — release build + vet + `backend/tests/**` race PASS.
- [ ] Collection non-zero; source/image/resources/duration/cleanup приложены.
- [ ] Supply-chain Docker/CI job — module integrity + vulnerability scan без blocking findings.
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
- [ ] **Тесты.** Genuine isolated units в `backend/tests/{domain}/`; применимые happy/negative/authz/boundary cases; non-unit risks маршрутизированы в Python `testframework/**`.
- [ ] **Комментарии.** WHY-not-WHAT; godoc на экспортируемом; без эмодзи.
- [ ] **Регрессия.** Docker job `backend-unit` — release build/vet/race PASS, exact evidence и cleanup приложены.
- [ ] **Docs.** `system_specification.md` + `wiki/API.md` при новых эндпоинтах; `STATE.md` обновлён.
- [ ] **Git.** Коммит `type(scope): ...`; diff без секретов; кросс-ревью.

### 17.2. Что именно проверяет be-ver

`be-ver` — readonly-верификатор. Его работа — не переписать, а вынести вердикт `verified: true/false` с доказательствами. Проверяет:

1. **Layering (грепом).** `rg "ent\.Client" backend/internal/*/handler*.go backend/internal/*/service*.go` — должно быть пусто (кроме легитимного `GetDB()`-доступа в поиске). `fiber` не импортируется в service.
2. **Security-инварианты (грепом).** `http.DefaultClient`, `io.ReadAll` без `LimitReader`, `!=` для секретов, `panic(`, пропуск webhook-подписи, поиск без скоупа.
3. **Обработка ошибок.** Нет голого `return err` на границах; `ent.IsNotFound`→404; конфликт версии→409.
4. **Context.** `ctx` первым; нет `context.Background()` в handler; нет ctx в структурах.
5. **Ent-инварианты.** `version` + `AddVersion(1)` + проверка `affected==0`; транзакции с rollback; outbox в tx; soft delete фильтр.
6. **Тесты и регрессия.** Genuine units существуют и релевантны; Docker job `backend-unit` дал race/build/vet PASS с exact evidence и cleanup.
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
| Тест рядом в `internal/` | Смешивает production и test ownership | genuine unit в `backend/tests/{domain}/`; non-unit — Python `testframework/**` |

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

- [ ] Genuine isolated units в `backend/tests/{domain}/`; non-unit scenarios не маскируются под Go unit.
- [ ] Docker job `backend-unit` — release build/vet/race PASS; collection, duration и cleanup приложены.
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
