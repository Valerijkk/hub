# Техническая документация -- Трекер задач

## Архитектура модуля

Модуль трекера задач реализован по паттерну **Handler -> Service -> Repository** с разделением ответственности:

```
HTTP Request
     |
     v
+------------+    +------------------+    +----------------+
|  Handler   | -> | TrackerService   | -> |  Repository    |
| (Fiber)    |    | (бизнес-логика)  |    |  (интерфейс)   |
+------------+    +------------------+    +----------------+
                        |                        |
                        v                        v
                  +-------------+         +---------------+
                  | Workflow    |         | EntRepository  |
                  | Engine      |         | (Ent ORM)      |
                  +-------------+         +---------------+
                                                |
                                                v
                                         +------------+
                                         | PostgreSQL |
                                         +------------+
```

### Handler (`handler.go`)

- Принимает HTTP-запросы через Fiber v3
- Парсит параметры пути, query-параметры и тело запроса
- Извлекает userID из JWT (через `c.Locals("userID")`)
- Делегирует бизнес-логику в TrackerService
- Возвращает JSON-ответы с корректными HTTP-кодами
- Регистрирует маршруты с RBAC-middleware

### Service (`service.go`)

- Содержит всю бизнес-логику модуля
- Валидирует входные данные (UUID-парсинг, проверки)
- Вызывает WorkflowEngine для валидации переходов
- Записывает события в журнал активности
- Конвертирует Ent-сущности в DTO через converter-функции
- Не зависит от HTTP-фреймворка

### Repository (`repository.go`)

- Определяет интерфейс `Repository` с 30+ методами
- Реализация `EntRepository` использует Ent ORM
- Поддерживает `OutboxPublisher` для транзакционных событий
- Позволяет unit-тестирование через mock-реализацию

### WorkflowEngine (`workflow.go`)

- Отдельная структура (Single Responsibility Principle)
- Единственный метод `ValidateTransition(ctx, fromID, toID) error`
- Проверяет наличие записи в таблице workflow_transitions
- Возвращает 400 Bad Request при запрещенном переходе

---

## Ent-схемы

Модуль трекера определяет 12 Ent-схем. Все находятся в `backend/ent/schema/`.

### Issue (Задача)

Основная сущность модуля.

| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK, автогенерация |
| issue_number | int | Порядковый номер в проекте, immutable |
| issue_key | string | Уникальный ключ вида "HUB-1", immutable |
| title | string | Заголовок, max 500 |
| description_json | JSON | Plate editor JSON |
| description_text | text | Plain text для поиска |
| issue_type | enum | bug, task, story, epic, sub_task, improvement, tech_debt, spike, release |
| priority | enum | critical, high, medium (default), low, trivial |
| rank | string | LexoRank для позиционирования на Kanban |
| version | int | Оптимистичная блокировка, default 1 |
| estimate_minutes | int? | Оценка времени |
| time_spent_minutes | int | Фактическое время, default 0 |
| custom_fields | JSON | JSONB: {field_def_id -> value} |
| due_date | time? | Дедлайн |
| created_at | time | Immutable |
| updated_at | time | Auto-update |

**Edges:**
- `project` (M:1, required) -- принадлежит проекту
- `status` (M:1, required) -- текущий workflow-статус
- `assignee` (M:1) -- исполнитель
- `reporter` (M:1) -- автор
- `parent` / `children` (self-referencing) -- иерархия
- `comments` (1:M) -- комментарии
- `labels` (M:M) -- метки
- `time_logs` (1:M) -- записи времени
- `activities` (1:M) -- журнал активности
- `files` (1:M) -- вложения (FileMetadata из Phase 1)
- `outgoing_links` (1:M) -- исходящие связи
- `incoming_links` (M:1) -- входящие связи

**Индексы:** issue_key, issue_number, rank, created_at, priority, due_date

### IssueLink (Связь задач)

| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK |
| link_type | enum | blocks, blocked_by, relates_to, duplicates, duplicated_by, clones, cloned_from, causes, caused_by, parent, child |
| custom_type | string? | Пользовательский тип связи |
| created_at | time | Immutable |

**Edges:**
- `source` (M:1, required) -- исходная задача
- `target` (M:1, required) -- целевая задача

**Индексы:** unique(source, target)

### WorkflowStatus (Статус workflow)

| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK |
| name | string | Имя статуса, max 100 |
| category | enum | todo, in_progress, done |
| color | string | HEX-цвет, default "#6B7280" |
| sort_order | int | Порядок отображения, default 0 |
| is_default | bool | Статус по умолчанию, default false |
| created_at | time | Immutable |

**Edges:**
- `project` (M:1, required) -- принадлежит проекту
- `issues` (1:M) -- задачи в этом статусе
- `from_transitions` (1:M) -- переходы из этого статуса
- `to_transitions` (M:1) -- переходы в этот статус

**Индексы:** unique(project, name), (project, sort_order)

### WorkflowTransition (Переход workflow)

| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK |
| created_at | time | Immutable |

**Edges:**
- `from_status` (M:1, required) -- исходный статус
- `to_status` (M:1, required) -- целевой статус

**Индексы:** unique(from_status, to_status)

### Comment (Комментарий)

| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK |
| body_json | JSON | Plate editor JSON |
| body_text | text | Plain text для поиска/уведомлений |
| is_internal | bool | Внутренний комментарий, default false |
| mentions | JSON([]string) | UUID упомянутых пользователей |
| created_at | time | Immutable |
| updated_at | time | Auto-update |

**Edges:**
- `issue` (M:1, required) -- принадлежит задаче
- `author` (M:1, required) -- автор (User)
- `reactions` (1:M) -- реакции

### Reaction (Реакция)

| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK |
| emoji | string | Эмодзи, max 32 |
| created_at | time | Immutable |

**Edges:**
- `comment` (M:1, required) -- принадлежит комментарию
- `user` (M:1, required) -- пользователь

**Индексы:** unique(comment, user, emoji)

### ActivityLog (Журнал активности)

| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK |
| action | string | Тип действия (created, updated, commented, status_changed, linked, time_logged) |
| changes | JSON([]FieldChange) | Список изменений полей |
| created_at | time | Immutable |

**FieldChange:**
```go
type FieldChange struct {
    Field    string `json:"field"`
    OldValue any    `json:"old_value"`
    NewValue any    `json:"new_value"`
}
```

**Edges:**
- `issue` (M:1, required) -- принадлежит задаче
- `user` (M:1, required) -- кто совершил действие

**Индексы:** (issue, created_at)

### TimeLog (Запись времени)

| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK |
| minutes | int | Количество минут (>0) |
| description | string? | Описание, max 500 |
| started_at | time? | Начало таймера |
| ended_at | time? | Конец таймера |
| created_at | time | Immutable |

**Edges:**
- `issue` (M:1, required) -- принадлежит задаче
- `user` (M:1, required) -- кто записал время

### Label (Метка)

| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK |
| name | string | Имя метки, max 50 |
| color | string | HEX-цвет, default "#3B82F6" |
| created_at | time | Immutable |

**Edges:**
- `project` (M:1, required) -- принадлежит проекту
- `group` (M:1) -- группа меток
- `issues` (M:M) -- задачи с этой меткой

**Индексы:** unique(project, name)

### LabelGroup (Группа меток)

| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK |
| name | string | Имя группы, max 50 |
| created_at | time | Immutable |

**Edges:**
- `project` (M:1, required) -- принадлежит проекту
- `labels` (1:M) -- метки в группе

**Индексы:** unique(project, name)

### Board (Доска)

| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK |
| name | string | Имя доски, max 100 |
| filters | JSON | Сохраненные фильтры (status, assignee, priority, type) |
| is_default | bool | Доска по умолчанию, default false |
| created_at | time | Immutable |

**Edges:**
- `project` (M:1, required) -- принадлежит проекту

### CustomFieldDefinition (Определение кастомного поля)

| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK |
| name | string | Имя поля, max 100 |
| field_type | enum | text, number, date, checkbox, select, multi_select, user, url |
| options | JSON([]string) | Варианты для select/multi_select |
| required | bool | Обязательность, default false |
| sort_order | int | Порядок отображения, default 0 |
| created_at | time | Immutable |

**Edges:**
- `project` (M:1, required) -- принадлежит проекту

**Индексы:** unique(project, name)

---

## Движок рабочих процессов (WorkflowEngine)

### Структура

```go
type WorkflowEngine struct {
    repo Repository
}
```

### Метод ValidateTransition

```go
func (w *WorkflowEngine) ValidateTransition(ctx context.Context, fromStatusID, toStatusID uuid.UUID) error
```

Алгоритм:
1. Ищет запись `WorkflowTransition` с `from_status = fromStatusID` и `to_status = toStatusID`
2. Если найдена -- переход разрешен, возвращает `nil`
3. Если не найдена -- возвращает `400 Bad Request: "переход из текущего статуса не разрешен"`

### SeedDefaultWorkflow

Метод `Repository.SeedDefaultWorkflow(ctx, projectID)` создает стандартный набор статусов и переходов при создании нового проекта:
- 4 статуса: To Do (todo, default), In Progress (in_progress), In Review (in_progress), Done (done)
- Переходы: To Do -> In Progress -> In Review -> Done

---

## LexoRank позиционирование

### Назначение

LexoRank используется для позиционирования задач на Kanban-доске внутри колонки. Каждая задача имеет строковое поле `rank`, которое определяет порядок отображения.

### Принцип работы

При перетаскивании карточки:
1. Клиент вычисляет новый rank как среднее между rank предыдущей и следующей карточки
2. Новый rank передается в `MoveIssueRequest.Rank`
3. Сервер сохраняет новый rank в БД

### Порог ребалансировки

Если длина rank превышает 50 символов, требуется ребалансировка (пересчет rank для всех задач в колонке). Это происходит редко и выполняется фоновым процессом.

### Индексирование

Поле `rank` индексировано для быстрой сортировки:
```go
index.Fields("rank")
```

---

## Оптимистичная блокировка

### Механизм

1. Каждая задача имеет поле `version` (int, default 1)
2. При обновлении клиент передает `version` в запросе
3. Repository выполняет: `UPDATE issues SET ... WHERE id = ? AND version = ?`
4. Если `version` не совпадает -- возвращается `409 Conflict`
5. При успешном обновлении `version` увеличивается на 1

### Реализация в сервисе

```go
func (s *TrackerService) UpdateIssue(ctx, issueKey, userID, req) {
    existing := s.repo.FindIssueByKey(ctx, issueKey)
    // ... валидация ...
    input.Version = req.Version  // передаем ожидаемую версию
    updated := s.repo.UpdateIssue(ctx, existing.ID, input)
    // repo проверяет WHERE version = input.Version
}
```

### Mock-реализация для тестов

```go
func (m *mockRepo) UpdateIssue(_, issueID, input) {
    if i.Version != input.Version {
        return nil, apperrors.NewConflict("version mismatch")
    }
    i.Version++
}
```

---

## Атомарная нумерация

### Проблема

При конкурентном создании задач несколькими пользователями необходимо гарантировать уникальность номеров.

### Решение

Счетчик задач хранится в таблице `projects` (поле `task_counter`). При создании задачи выполняется атомарный инкремент в транзакции:

```go
project.Update().AddTaskCounter(1).Save(ctx)
```

Это использует PostgreSQL `UPDATE ... SET task_counter = task_counter + 1 RETURNING task_counter`, что гарантирует атомарность без блокировок.

### Формат ключа

```go
issueKey = fmt.Sprintf("%s-%d", project.Key, newCounter)
// Пример: "HUB-42"
```

---

## Паттерн Outbox

### Назначение

Outbox-паттерн обеспечивает надежную публикацию событий в NATS в рамках транзакции БД.

### Механизм

1. При создании/обновлении задачи в той же транзакции создается запись в таблице `outbox_events`
2. Фоновый воркер (100мс polling) читает непубликованные события
3. Публикует в NATS и помечает как обработанные
4. При сбое NATS события остаются в outbox и будут переопубликованы

### Интерфейс

```go
type OutboxPublisher interface {
    PublishInTx(ctx context.Context, tx *ent.Tx, eventType string, payload any) error
}
```

### События трекера

| Событие | Когда публикуется |
|---------|------------------|
| `tracker.issue.created` | Создание задачи |
| `tracker.issue.updated` | Обновление задачи |
| `tracker.issue.deleted` | Удаление задачи |
| `tracker.issue.moved` | Перемещение на доске |
| `tracker.comment.created` | Новый комментарий |
| `tracker.time.logged` | Запись времени |

---

## REST API справочник

### Базовый путь

Все эндпоинты трекера начинаются с `/api/projects/:projectId`. Требуется JWT-аутентификация (middleware `authMw`).

### Формат ошибок

```json
{
  "error": "описание ошибки",
  "code": 400
}
```

Коды ошибок:
- `400` -- некорректный запрос (невалидные данные, запрещенный переход)
- `401` -- не аутентифицирован
- `403` -- нет прав доступа (RBAC)
- `404` -- ресурс не найден
- `409` -- конфликт (оптимистичная блокировка, дубликат)

### Issues API

#### POST /api/projects/:projectId/issues

Создание задачи.

**Тело запроса:**
```json
{
  "title": "string (required, 1-500)",
  "issue_type": "bug|task|story|epic|sub_task|improvement|tech_debt|spike|release",
  "description_json": {"children": [...]},
  "priority": "critical|high|medium|low|trivial",
  "assignee_id": "uuid",
  "parent_id": "uuid",
  "status_id": "uuid",
  "labels": ["uuid"],
  "custom_fields": {"field_id": "value"},
  "estimate_minutes": 60,
  "due_date": "2026-04-01T00:00:00Z"
}
```

**Ответ (201):**
```json
{
  "id": "uuid",
  "issue_number": 1,
  "issue_key": "HUB-1",
  "title": "...",
  "issue_type": "task",
  "priority": "medium",
  "version": 1,
  "status": {"id": "uuid", "name": "To Do", "category": "todo", ...},
  "reporter": {"id": "uuid", "name": "...", "email": "..."},
  "created_at": "...",
  "updated_at": "..."
}
```

#### GET /api/projects/:projectId/issues

Список задач с курсорной пагинацией.

**Query-параметры:**
- `cursor` -- issue_key для смещения
- `limit` -- от 1 до 100 (default 50)
- `status` -- фильтр по статусам (массив)
- `priority` -- фильтр по приоритетам (массив)
- `assignee` -- UUID исполнителя
- `issue_type` -- тип задачи
- `search` -- текстовый поиск
- `order_by` -- сортировка: created_at, priority, updated_at

**Ответ (200):**
```json
{
  "items": [...],
  "next_cursor": "HUB-50",
  "has_more": true,
  "total": 142
}
```

#### PUT /api/projects/:projectId/issues/:issueKey

Обновление задачи с оптимистичной блокировкой.

**Тело запроса (все поля опциональны кроме version):**
```json
{
  "title": "New title",
  "priority": "high",
  "status_id": "uuid",
  "assignee_id": "uuid",
  "labels": ["uuid"],
  "custom_fields": {"field_id": "value"},
  "version": 1
}
```

**Ответ (200):** обновленная задача (IssueResponse)
**Ответ (409):** version mismatch

#### PUT /api/projects/:projectId/issues/:issueKey/position

Перемещение задачи на Kanban-доске.

**Тело запроса:**
```json
{
  "status_id": "uuid (required)",
  "rank": "aaa|bbb (required)"
}
```

Выполняет валидацию workflow-перехода перед перемещением.

---

## RBAC-права

### Права модуля трекера

| Право | Бит | Описание |
|-------|-----|----------|
| PermProjectView | 1 | Просмотр проекта и задач |
| PermProjectManageWorkflow | 32 | Управление workflow (статусы, переходы, кастомные поля) |
| PermTaskCreate | 64 | Создание задач |
| PermTaskEdit | 128 | Редактирование задач, меток, связей, время |
| PermTaskEditAny | 256 | Редактирование любых задач (не только своих) |
| PermTaskDelete | 512 | Удаление задач |
| PermTaskAssign | 1024 | Назначение исполнителя |
| PermTaskComment | 2048 | Добавление/редактирование/удаление комментариев |

### Роли по умолчанию

| Роль | Права трекера |
|------|--------------|
| Owner | Все права (^0) |
| Developer | View, Create, Edit, Assign, Comment |
| Tester | View, Create, Edit, Comment |
| Viewer | View только |

### Middleware

RBAC проверяется через middleware `rbac.RequirePermission(perm, rbacSvc)`, который:
1. Извлекает userID из контекста
2. Получает projectID из параметров пути
3. Проверяет `permission.Has(requiredPerm)` через кеш
4. Возвращает 403 если нет права

---

## NATS-события

| Событие | Payload | Когда |
|---------|---------|-------|
| `tracker.issue.created` | `{issue_id, project_id, issue_key, title, type, reporter_id}` | Создание задачи |
| `tracker.issue.updated` | `{issue_id, changes: [{field, old, new}]}` | Обновление полей |
| `tracker.issue.deleted` | `{issue_id, issue_key}` | Удаление задачи |
| `tracker.issue.moved` | `{issue_id, from_status_id, to_status_id}` | Смена статуса на доске |
| `tracker.comment.created` | `{comment_id, issue_id, author_id, mentions}` | Новый комментарий |
| `tracker.time.logged` | `{issue_id, user_id, minutes}` | Запись времени |

События публикуются через outbox-паттерн для гарантированной доставки.

---

## Фронтенд архитектура

### Компоненты

| Компонент | Путь | Назначение |
|-----------|------|-----------|
| KanbanBoard | `frontend/src/features/tracker/KanbanBoard.tsx` | Kanban-доска с колонками и карточками |
| IssueCard | `frontend/src/features/tracker/IssueCard.tsx` | Карточка задачи на доске |
| IssueForm | `frontend/src/features/tracker/IssueForm.tsx` | Форма создания/редактирования задачи |
| RichTextEditor | `frontend/src/features/tracker/RichTextEditor.tsx` | Plate-редактор для описаний/комментариев |
| WorkflowSettings | `frontend/src/features/tracker/WorkflowSettings.tsx` | Настройка статусов и переходов |

### Stores (TanStack Query)

- `useIssues(projectId, filters)` -- список задач с пагинацией
- `useIssue(issueKey)` -- детали задачи
- `useBoardData(boardId)` -- данные доски с колонками
- `useWorkflowStatuses(projectId)` -- статусы workflow
- Мутации: `useCreateIssue`, `useUpdateIssue`, `useMoveIssue`

### API Client

Типизированный HTTP-клиент на базе fetch + TanStack Query:
```typescript
const api = {
  issues: {
    list: (projectId, params) => get(`/projects/${projectId}/issues`, params),
    create: (projectId, data) => post(`/projects/${projectId}/issues`, data),
    update: (projectId, key, data) => put(`/projects/${projectId}/issues/${key}`, data),
  }
}
```

### Маршруты (TanStack Router)

```
/projects/$projectId/board        -- Kanban-доска
/projects/$projectId/issues       -- Список задач
/projects/$projectId/issues/$key  -- Детали задачи
/projects/$projectId/settings/workflow -- Настройки workflow
```

Все маршруты защищены `_authenticated` layout route.

---

## Решения и обоснования

### LexoRank vs Float для позиционирования

| Критерий | LexoRank (string) | Float |
|----------|-------------------|-------|
| Точность | Без потери точности | Теряет точность при частых вставках |
| Сортировка | Лексикографическая (нативная PG) | Числовая |
| Ребалансировка | При 50+ символах | При потере точности |
| Используется в | Jira, Linear, Notion | Trello (старые версии) |

**Выбрано:** LexoRank -- стандарт индустрии для Kanban-досок.

### Plate vs Tiptap для rich-text

| Критерий | Plate | Tiptap |
|----------|-------|--------|
| Стоимость | Бесплатно | Таблицы платные ($149/мес) |
| UI-фреймворк | Shadcn-native | Headless |
| @Упоминания | Из коробки | Плагин |
| Размер бандла | Средний | Средний |

**Выбрано:** Plate -- бесплатный, нативная интеграция с Shadcn/ui.

**Особенность:** Plate v49 использует `usePlateEditor` + `editor` prop вместо `value` prop на компоненте `Plate`. Тип `Value` импортируется из `@udecode/plate`, компоненты из `@udecode/plate/react`.

### Cursor vs Offset пагинация

| Критерий | Cursor | Offset |
|----------|--------|--------|
| Производительность | O(1) | O(offset) |
| Консистентность | Стабильная при вставках | "Перепрыгивание" при вставках |
| Произвольный доступ | Нет | Да (перейти на страницу N) |

**Выбрано:** Cursor -- производительнее для списков с частыми вставками.

### JSONB vs EAV для кастомных полей

| Критерий | JSONB | EAV (Entity-Attribute-Value) |
|----------|-------|-----|
| Производительность чтения | Одна строка | JOIN на N атрибутов |
| Гибкость | Полная | Полная |
| Индексирование | GIN-индекс | B-tree на value |
| Миграции | Не нужны | Не нужны |

**Выбрано:** JSONB -- проще, быстрее, нет JOIN-оверхеда.

### Mock Repository vs Testcontainers

| Критерий | Mock | Testcontainers |
|----------|------|---------------|
| Скорость тестов | Мгновенные | Запуск контейнера (5-10 сек) |
| Изоляция | Полная | Полная |
| Покрытие SQL | Нет | Да |
| CI-зависимости | Нет | Docker |

**Выбрано:** Mock repository для unit-тестов сервиса. Testcontainers будут использоваться для интеграционных тестов в фазе деплоя.

---

## Известные ограничения

### LexoRank: исчерпание пространства

При интенсивном перетаскивании карточек длина rank может расти. Когда длина превышает 50 символов, требуется ребалансировка. В текущей реализации ребалансировка не автоматизирована -- требуется ручной вызов.

**Митигация:** Мониторинг длины rank, предупреждение при приближении к порогу.

### Plate: привязка к версии

Plate v49 имеет ломающие изменения API по сравнению с предыдущими версиями. Обновление требует ревью миграционного гайда.

**Митигация:** Фиксация версии в package.json, обновление только при необходимости.

### Гранулярность diff в описании

Журнал активности фиксирует факт изменения описания, но не конкретные diff-ы (вставка/удаление текста). Полная история diff-ов описания не реализована.

**Митигация:** При необходимости можно добавить сравнение plain-text версий.

### Фильтрация по кастомным полям

Серверная фильтрация по значениям кастомных полей (JSONB) не реализована в текущей версии. Фильтрация доступна только по стандартным полям.

**Митигация:** Можно добавить GIN-индекс на `custom_fields` и JSONB-операторы в запросы.

### Уникальность реакций

Один пользователь может поставить только одну реакцию каждого типа на комментарий. Это обеспечивается уникальным индексом (comment, user, emoji).

---

*Документация составлена на основе реализации Phase 2 проекта Hub. Последнее обновление: 2026-03-21.*
