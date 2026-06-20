# Hub — Code Style Guide

> Единая база знаний. Обновляется ARCH при изменении стандартов.
> Cursor rules: `.cursor/rules/10-style.mdc`, `20-backend.mdc`, `30-frontend.mdc`

---

## Универсальные правила

- Комментарии объясняют **WHY**, не WHAT — никаких narrating comments
- Запрещены placeholder-комментарии: `// Handle error`, `// Return result`
- Без эмодзи в коде и комментариях
- Commit format: `feat(scope): description` | `fix(scope): description` | `refactor(scope): description`
- Scope: `auth`, `tracker`, `messenger`, `tms`, `calls`, `dashboard`, `notification`, `search`, `ws`, `platform`, `frontend`, `infra`, `ent`

## Go — именование

| Сущность | Конвенция | Пример |
|----------|-----------|--------|
| Пакеты | lowercase, одно слово | `tracker`, `platform` |
| Интерфейсы | Суффикс роли | `IssueRepository`, `EventPublisher` |
| Файлы | snake_case по роли | `handler.go`, `ent_repository.go` |
| Константы ошибок | `Err` prefix | `ErrNotFound`, `ErrConflict` |
| Context | Первый параметр | `func (s *Service) Create(ctx context.Context, ...)` |

## Go — структура модуля

```
backend/internal/<module>/
├── handler.go          # HTTP only
├── service.go          # Business logic
├── repository.go       # Interface
├── ent_repository.go   # Ent implementation
├── dto.go              # Request/Response
└── <module>_test.go    # Tests
```

## Go — обработка ошибок

```go
// Правильно
if ent.IsNotFound(err) { return nil, ErrNotFound }
if errors.Is(err, context.DeadlineExceeded) { ... }
return fmt.Errorf("create issue: %w", err)

// Запрещено
if err != nil { return nil, err }  // теряет контекст
panic(err)                         // в production paths
```

## Go — UUID и Ent

```go
// Правильно
field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable()
field.Int("version").Default(1)  // optimistic lock

// Запрещено
field.String("id")  // UUID как строка
```

## Go — логирование

```go
slog.InfoContext(ctx, "issue created", "id", issue.ID, "project_id", projectID)
// Запрещено: fmt.Println, log.Printf
```

## TypeScript/React — именование

| Сущность | Конвенция | Пример |
|----------|-----------|--------|
| Компоненты | PascalCase | `IssueCard.tsx` |
| Хуки | `use` prefix | `useIssues.ts` |
| Stores | camelCase + Store | `authStore.ts` |
| Routes | kebab-case файлы | `issue-detail.tsx` |
| API types | PascalCase | `CreateIssueRequest` |

## TypeScript/React — паттерны

```typescript
// Optional chaining
const name = user?.profile?.name ?? 'Unknown'

// Rules of Hooks — все хуки ДО любого return
export function Component({ id }: Props) {
  const data = useQuery({ queryKey: ['item', id], ... })
  if (!id) return null  // conditional AFTER hooks

// Stable keys
items.map(item => <Card key={item.id} ... />)

// Explicit button type
<button type="button" onClick={handleClose}>Cancel</button>
```

## Запрещённые паттерны

| Паттерн | Почему | Альтернатива |
|---------|--------|--------------|
| `http.DefaultClient` | SSRF, no timeout | `safehttp.NewSafeHTTPClient(10*time.Second)` |
| `io.ReadAll` без лимита | Memory DoS | `io.LimitReader(r, maxBytes)` |
| `!=` для секретов | Timing attack | `subtle.ConstantTimeCompare` |
| `*ent.Client` в Service/Handler | Нарушение layering | Repository interface |
| `panic()` в handlers | Production crash | return error |
| `any` без комментария | Потеря типов | Explicit type или `// reason` |
| `key={index}` в списках | Reconciliation bug | Stable ID |
| `!important` в CSS | Tech debt | Tailwind utilities |
| `user!.profile!.name` | Runtime crash | Optional chaining + guard |

## SOLID (обязательно)

- **S**: Handler не сохраняет в БД; Service не парсит HTTP
- **O**: Новые payment/webhook handlers через интерфейс, не правка существующего
- **L**: Mock Repository подставляется без изменения Service
- **I**: Раздельные интерфейсы `IssueReader`, `IssueWriter` при необходимости
- **D**: Service зависит от `IssueRepository`, не от `*ent.Client`

## DRY и KISS

- Бизнес-логика объявляется один раз в Service
- Избегать Factory-of-Factory; предпочитать прямые функции
- Общие HTTP-утилиты в `platform/http`, ошибки в `platform/errors`

## CSS / UI

- Tailwind CSS 4 + Shadcn/ui
- Shadcn: только `className` изменения, Radix поведение не трогать
- WCAG 2.1 AA: контраст, aria, keyboard navigation

## Git

См. `.cursor/rules/50-git.mdc`:
- Branch: `feat/<module>/<desc>`, `fix/<module>/<desc>`
- Conventional Commits
- Pre-commit: `go build ./...`, `go test -race`, `tsc --noEmit`, `npm run lint`

---
*Последнее обновление: 2026-06-17*
