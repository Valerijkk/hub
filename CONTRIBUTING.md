# Контрибуция в Hub

Руководство по участию в разработке проекта.

## Git Workflow

### Ветвление

Основная ветка: `master`. Вся разработка ведётся через feature-ветки.

```
master              (стабильная ветка)
  └── feat/TASK-42-add-poker-timer     (фича)
  └── fix/BUG-15-auth-redirect         (багфикс)
  └── docs/update-api-reference        (документация)
  └── chore/upgrade-dependencies       (инфраструктура)
```

### Правила именования веток

```
{тип}/{ключ-задачи}-{краткое-описание}
```

| Тип | Назначение |
|---|---|
| `feat` | Новая функциональность |
| `fix` | Исправление бага |
| `docs` | Документация |
| `refactor` | Рефакторинг без изменения поведения |
| `test` | Тесты |
| `chore` | Инфраструктура, зависимости, CI |

### Процесс

1. Создать ветку от `master`
2. Разрабатывать, коммитить
3. Создать Pull Request
4. Пройти code review
5. Merge в `master` (squash merge)

## Формат коммитов

Используем [Conventional Commits](https://www.conventionalcommits.org/):

```
{тип}({скоуп}): {описание}

[тело]
```

### Примеры

```bash
feat(tracker): add WIP limits to kanban boards
fix(auth): fix token refresh race condition
docs(api): update TMS endpoints documentation
test(sprint): add burndown calculation tests
refactor(messenger): extract presence service
chore(deps): upgrade React to v19.1
```

### Типы коммитов

| Тип | Назначение | Пример |
|---|---|---|
| `feat` | Новая фича | `feat(tms): add Allure import` |
| `fix` | Багфикс | `fix(calls): fix recording upload` |
| `docs` | Документация | `docs: update README` |
| `test` | Тесты | `test(auth): add magic link tests` |
| `refactor` | Рефакторинг | `refactor(ws): simplify hub logic` |
| `chore` | Инфраструктура | `chore(docker): update Compose config` |
| `perf` | Оптимизация | `perf(search): add DB indexes` |

### Скоупы (по модулям)

`auth`, `project`, `tracker`, `sprint`, `messenger`, `tms`, `calls`, `dashboard`, `notification`, `search`, `github`, `gitlab`, `rbac`, `storage`, `imex`, `backup`, `audit`, `api`, `ws`, `deps`, `docker`, `ci`

## Code Style

### Go (Backend)

- Форматирование: `gofmt` (автоматическое)
- Линтер: `golangci-lint`
- Именование: стандартные Go-конвенции (CamelCase для экспортируемых, camelCase для приватных)
- Обработка ошибок: всегда проверять `err != nil`, использовать `fiber.NewError()` для HTTP-ошибок
- Комментарии: GoDoc-формат для всех экспортируемых функций и типов
- Контекст: передавать `context.Context` первым аргументом

```go
// CreateIssue создаёт новую задачу в проекте.
func (s *TrackerService) CreateIssue(ctx context.Context, req CreateIssueRequest) (*Issue, error) {
    if req.Title == "" {
        return nil, fiber.NewError(fiber.StatusBadRequest, "название обязательно")
    }
    // ...
}
```

### TypeScript (Frontend)

- Линтер: ESLint
- Форматирование: Prettier
- Строгий TypeScript: `strict: true`
- Импорты: абсолютные пути через `@/`
- Компоненты: функциональные компоненты с именованным экспортом
- Стилизация: Tailwind CSS утилиты, без inline-стилей

```tsx
export function IssueCard({ issue }: { issue: Issue }) {
  return (
    <div className="rounded-lg border p-4">
      <h3 className="font-medium">{issue.title}</h3>
    </div>
  );
}
```

### Общие правила

- Без `any` в TypeScript (использовать `unknown` + type guards)
- Без `console.log` в продакшн-коде
- Без хардкоженных строк (использовать i18n)
- Максимальная длина строки: 120 символов (Go), без жёсткого лимита (TS, Prettier форматирует)

## Pull Request

### Создание PR

1. Заголовок: `{тип}({скоуп}): краткое описание` (как коммит)
2. Описание: что сделано, зачем, как тестировать
3. Привязать задачу (если есть)

### Чеклист PR

- [ ] Код соответствует code style
- [ ] Добавлены/обновлены тесты (если применимо)
- [ ] Нет `console.log`, `TODO` без задачи, хардкоженных значений
- [ ] API endpoints задокументированы
- [ ] Миграции БД (если менялась схема)
- [ ] README/docs обновлены (если менялся API/архитектура)

### Code Review

- Минимум 1 approve для merge
- Автор отвечает за разрешение всех комментариев
- Squash merge в `master`

## Шаблоны Issue

### Bug Report

```markdown
## Описание бага

[Что происходит]

## Шаги для воспроизведения

1. ...
2. ...
3. ...

## Ожидаемое поведение

[Что должно происходить]

## Фактическое поведение

[Что происходит на самом деле]

## Окружение

- Браузер:
- ОС:
- Версия Hub:
```

### Feature Request

```markdown
## Описание фичи

[Что нужно реализовать]

## Мотивация

[Зачем это нужно, какую проблему решает]

## Предлагаемая реализация

[Как это может работать]

## Альтернативы

[Какие ещё варианты рассматривались]
```

## Локальная разработка

### Backend

```bash
cd backend
go run cmd/server/main.go
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### Тесты

```bash
# Backend
cd backend
go test ./internal/...

# Frontend
cd frontend
npm test
```

## Вопросы

Если что-то непонятно -- создайте issue или напишите в мессенджер (канал #dev).
