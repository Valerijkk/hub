# Backend: Fiber v3 и HTTP-слой

> Часть плейбука. Канон и оглавление — [`../backend.md`](../backend.md).

Роутинг, middleware, обработка ошибок, контракты ответов.

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
