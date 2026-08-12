# Hub — Руководство по стилю кода

> Единая база знаний. Обновляется ARCH при изменении стандартов.  
> Cursor rules: `.cursor/rules/10-style.mdc`, `20-backend.mdc`, `30-frontend.mdc`

---

## Универсальные правила

| Правило | Детали |
|---------|--------|
| Комментарии | Объясняют **WHY**, не WHAT — без narrating comments |
| Placeholder | Запрещены: `// Handle error`, `// Return result` |
| Эмодзи | Запрещены в коде и комментариях |
| Commits | `feat(scope): desc` \| `fix(scope): desc` \| `refactor(scope): desc` |
| Scope | `auth`, `tracker`, `messenger`, `tms`, `calls`, `dashboard`, `notification`, `search`, `ws`, `platform`, `frontend`, `infra`, `ent` |
| Язык комментариев | **Русский** в Go/TS (канон после RU-COMMENTS); docs — русский; английский допустим в godoc-экспортах/публичных API-именах |

---

## Go — именование

| Сущность | Конвенция | Пример |
|----------|-----------|--------|
| Пакеты | lowercase, одно слово | `tracker`, `platform` |
| Интерфейсы | Суффикс роли | `IssueRepository`, `EventPublisher` |
| Файлы | snake_case по роли | `handler.go`, `ent_repository.go` |
| Ошибки | `Err` prefix | `ErrNotFound`, `ErrConflict` |
| Context | Первый параметр | `func (s *Service) Create(ctx context.Context, ...)` |
| Тесты | `Test<Function>_<Scenario>` | `TestCreateIssue_DuplicateKey` |

---

## Go — структура модуля (слои)

```
backend/internal/<module>/
├── handler.go          # HTTP only — parse, validate, status codes
├── service.go          # Business logic — без *ent.Client
├── repository.go       # Interface
├── ent_repository.go   # Ent implementation
├── dto.go              # Request/Response structs
└── export_hooks.go     # legacy inventory; delete after exact protected-unit cutover
```

Target genuine unit tests: `backend/tests/<module>/**`. Старый
`backend/testframework/<module>/**` — legacy inventory, delete-not-port и не
является authoring path.

**Запрещено:** `*ent.Client` в Handler или Service — только через Repository interface.

### Поток запроса

```
HTTP → Handler → Service → Repository → Ent/DB
                    ↓
              NATS / Redis / external APIs
```

---

## Go — обработка ошибок

```go
// ✅ Правильно
if ent.IsNotFound(err) {
    return nil, ErrNotFound
}
if errors.Is(err, context.DeadlineExceeded) {
    return nil, fmt.Errorf("timeout: %w", err)
}
return fmt.Errorf("create issue: %w", err)

// ❌ Запрещено
if err != nil { return nil, err }  // теряет контекст
panic(err)                         // в production paths
```

Публичные API-ошибки (wire JSON, NFR-13): `platform/errors` → `{"error","request_id"}`.  
FE `ApiError` (`code`/`message`/`request_id`) — клиентский маппинг HTTP status + тела, не отдельный контракт.

---

## Go — UUID и Ent

```go
field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable()
field.Int("version").Default(1)  // optimistic locking — обязательно на mutable schemas
```

- Primary keys: **UUID v4**, не string
- Soft delete: `deleted_at` + корзина 30 дней
- Время в БД: **UTC**

---

## Go — логирование

```go
slog.InfoContext(ctx, "issue created",
    "id", issue.ID,
    "project_id", projectID,
    "request_id", middleware.RequestIDFromCtx(ctx),
)
```

- Structured logging (`slog`), context-aware
- **Не логировать:** JWT, magic link codes, refresh tokens, webhook secrets

---

## Go — HTTP handlers и Swagger

Каждый **публичный** endpoint — аннотации swag (English):

```go
// CreateIssue создаёт задачу в проекте
// @Summary      Create issue
// @Description  Create a new issue in the project
// @Tags         tracker
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID" format(uuid)
// @Param        request body CreateIssueRequest true "Issue data"
// @Success      201 {object} IssueResponse
// @Failure      400 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /projects/{project_id}/issues [post]
```

Регенерация: `go generate ./ent` + swag в CI/dev.

---

## Go — паттерны безопасности

| Паттерн | Использование |
|---------|---------------|
| `safehttp.NewSafeHTTPClient` | User-controlled URLs (GitLab, OpenGraph, WebPush) |
| `ValidateOutboundURL` | Перед любым outbound fetch |
| `io.LimitReader` | Чтение body (max 25MB align с Fiber BodyLimit) |
| `subtle.ConstantTimeCompare` | Magic link codes, webhook tokens |
| `select` на `client.done` | WS send — не писать в закрытый channel |

---

## TypeScript/React — структура

```
frontend/src/
├── api/<module>.ts       # API client (один файл на backend module)
├── components/ui/        # Shadcn — только className
├── components/<feature>/ # Feature components
├── hooks/use<Name>.ts
├── modules/<module>/     # Self-contained features
├── routes/               # TanStack Router file-based (**77** `.tsx`, 2026-07-28)
├── stores/<module>-store.ts  # Zustand (**12** stores)
```

Target FE unit tests: `frontend/tests/<domain>/**`; общий setup/fixtures —
`frontend/tests/support/**`. `testframework/stages/unit/frontend-test/**` —
legacy inventory, delete-not-port и не является authoring path.

---

## TypeScript/React — именование

| Сущность | Конвенция | Пример |
|----------|-----------|--------|
| Компоненты | PascalCase | `IssueCard.tsx` |
| Хуки | `use` prefix | `useIssues.ts` |
| Stores | `useXStore` | `useAuthStore` |
| Routes | kebab в path | `$issueKey.tsx` |
| API types | PascalCase | `CreateIssueRequest` |
| Тесты | `*.test.ts(x)` | `auth-store.test.ts` |

---

## TypeScript/React — паттерны

```typescript
// Optional chaining
const name = user?.profile?.name ?? 'Unknown'

// Rules of Hooks — ВСЕ хуки ДО conditional return
export function Component({ id }: Props) {
  const data = useQuery({ queryKey: ['item', id], ... })
  const store = useAuthStore()
  if (!id) return null

// Stable list keys
items.map(item => <Card key={item.id} />)

// Explicit button type (не submit по умолчанию)
<button type="button" onClick={handleClose}>Cancel</button>

// TanStack Query keys
queryKey: ['issues', projectId, filters]
```

### API-клиент

```typescript
// api/client.ts — single-flight refresh на 401
// api/tracker.ts — typed get/post helpers
```

### Ленивая загрузка (обязательно для тяжёлых deps)

```typescript
const CallsRoom = lazy(() => import('./modules/calls/CallsRoom'))
```

### Границы ошибок (Error Boundaries)

Обязательны для: calls, dashboard, tracker, TMS, messenger (route wrappers).

---

## Запрещённые паттерны

| Паттерн | Почему | Альтернатива |
|---------|--------|--------------|
| `http.DefaultClient` | SSRF, no timeout | `safehttp` |
| `io.ReadAll` без лимита | Memory DoS | `LimitReader` |
| `!=` для секретов | Timing attack | `ConstantTimeCompare` |
| `*ent.Client` в Service/Handler | Layering | Repository |
| `panic()` в handlers | Crash | return error |
| `any` без комментария | Type safety | explicit type |
| `key={index}` | Reconciliation | stable `id` |
| `!important` в CSS | Tech debt | Tailwind utilities |
| `user!.profile!` | Runtime crash | optional chaining |
| Inline `style={{}}` over DS | Inconsistency | CSS variables `--hub-*` |

---

## SOLID (обязательно)

| Принцип | В Hub |
|---------|-------|
| **S** | Handler не пишет в БД; Service не парсит HTTP |
| **O** | Новые webhook handlers через interface |
| **L** | Mock Repository без изменения Service |
| **I** | `IssueReader` / `IssueWriter` при необходимости |
| **D** | Service → `IssueRepository`, не `*ent.Client` |

---

## DRY и KISS

- Бизнес-логика — **один раз** в Service
- HTTP utilities → `platform/http`
- Стандартные ошибки → `platform/errors`
- Middleware → `platform/middleware`
- Избегать излишних абстракций (Factory-of-Factory)

---

## CSS / UI

- **Tailwind CSS 4** + **Shadcn/ui** + design tokens `--hub-*`
- Shadcn: только `className`; Radix behavior не менять
- **WCAG 2.1 AA**: контраст, `aria-*`, keyboard nav, focus ring
- **Без `!important`** (rule 60-tz-constraints)

---

## Тесты — стиль

| Слой | Расположение | Теги |
|------|--------------|------|
| Backend unit | `backend/tests/{domain}/**` | только genuine isolated Go units |
| Frontend unit | `frontend/tests/{domain}/**`, `frontend/tests/support/**` | genuine Vitest/RTL units |
| Contract/API/integration | `testframework/tests/{contract,api,integration}/**` | future greenfield pytest suites; `NOT_READY` |
| E2E | `testframework/tests/e2e/**` | future greenfield pytest + Playwright; `NOT_READY` |
| Framework selftests | `testframework/selftests/**` | framework contracts, не SUT suites |

Именование Go: table-driven где уместно; testify `require` / `assert`.

---

## Git и pre-commit

См. `.cursor/rules/50-git.mdc`. Применимые проверки запрашиваются только как
fixed targets через `/qa-gate`; прямые host-команды, direct runner и inline
Docker/Compose не являются допустимым pre-commit fallback. Пока target
`NOT_READY`, фиксируется `NOT_RUN`, а не выдуманный green.

**Не коммитить:** `.env`, coverage dumps (`backend/cov*`), secrets.

---
*Последнее обновление: 2026-06-22 — RU-FIX-01: RU section headers*
