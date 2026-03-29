# Фаза 4: TMS и Allure -- Техническая документация

## 1. Архитектура модуля TMS

### 1.1. Общая схема

Модуль TMS следует стандартной трёхслойной архитектуре Hub:

```
HTTP-запрос
    |
    v
Handler (handler.go)         -- парсинг запросов, RBAC, маршрутизация
    |
    v
Service (service.go)          -- бизнес-логика, версионирование, стейт-машина, Allure импорт
    |
    v
Repository (repository.go)   -- интерфейс доступа к данным
    |
    +--- EntRepository        -- PostgreSQL через Ent ORM
    +--- MongoRepository      -- MongoDB для сырого Allure JSON
    +--- CombinedRepository   -- объединяет оба под единым интерфейсом
```

Кросс-модульное взаимодействие через NATS:

```
TMSService --> NATS --> Notification Module (Phase 6)
    |                        |
    |                        v
    |                   Telegram Bot
    |
    +--> Shared Ent Client --> Issue (tracker module)
```

### 1.2. Структура файлов

```
backend/internal/tms/
  handler.go              -- Fiber HTTP-обработчики (22+ эндпоинтов)
  service.go              -- TMSService с бизнес-логикой
  repository.go           -- Repository интерфейс + входные структуры
  ent_repository.go       -- Ent/PG реализация
  mongo_repository.go     -- MongoDB реализация (Allure JSON)
  combined_repository.go  -- CombinedRepository (делегирует Ent + Mongo)
  dto.go                  -- DTO запросов/ответов (40+ типов)
  events.go               -- NATS-события (4 типа)
  allure_types.go         -- Go-структуры Allure JSON формата
  allure_parser.go        -- Защитный парсер Allure
  allure_matcher.go       -- Многостратегийный матчер
  importer.go             -- Импорт CSV/JSON/XML/Excel
  exporter.go             -- Экспорт CSV/JSON/XML/Excel/PDF
  service_test.go         -- Unit-тесты (23+ тестов)
  allure_parser_test.go   -- Тесты парсера и матчера (14 тестов)
```

### 1.3. Ключевые зависимости

| Пакет | Назначение |
|-------|-----------|
| `github.com/gofiber/fiber/v3` | HTTP-фреймворк |
| `entgo.io/ent` | ORM для PostgreSQL |
| `go.mongodb.org/mongo-driver/v2` | MongoDB драйвер |
| `github.com/valerii/hub/backend/internal/platform/nats` | NATS-клиент |
| `github.com/valerii/hub/backend/internal/rbac` | RBAC middleware |
| `github.com/xuri/excelize/v2` | Excel чтение/запись |
| `github.com/johnfercher/maroto/v2` | PDF генерация |

---

## 2. Схемы данных (Ent)

### 2.1. TestCase

Основная сущность тест-кейса.

| Поле | Тип | Ограничения | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable, default: uuid.New | Идентификатор |
| case_number | int | positive, immutable | Порядковый номер в проекте |
| case_key | string | unique, immutable | Ключ: HUB-TC-42 |
| title | string | not empty, max 500 | Название |
| description | text | optional | Описание |
| precondition | text | optional | Предусловие |
| postcondition | text | optional | Постусловие |
| case_type | enum | "classic", "bdd", default: "classic" | Тип кейса |
| priority | enum | "critical", "high", "medium", "low", "trivial", default: "medium" | Приоритет |
| status | enum | "draft", "active", "deprecated", default: "draft" | Статус |
| steps | JSON ([]TestStep) | optional | Шаги в формате JSON |
| full_name | string | optional | Полное имя для Allure |
| class_name | string | optional | Имя класса для Allure |
| method_name | string | optional | Имя метода для Allure |
| external_id | string | optional | Внешний ID (@TestCase) |
| version | int | default: 1 | Текущая версия |
| created_at | time | immutable, default: now | Дата создания |
| updated_at | time | default: now, update: now | Дата обновления |

**Рёбра:**
- `project` (From Project) -- M:1, required
- `suite` (From TestSuite) -- M:1, optional
- `created_by` (From User) -- M:1, required
- `versions` (To TestCaseVersion) -- 1:M
- `results` (To TestResult) -- 1:M
- `labels` (To Label) -- M:M
- `linked_issues` (To Issue) -- M:M
- `components` (To Component) -- M:M

**Индексы:**
- `case_key` -- поиск по ключу
- `external_id` -- Allure матчинг по аннотации
- `full_name` -- Allure матчинг по полному имени
- `class_name, method_name` -- составной для Allure матчинга
- `status` -- фильтрация
- `priority` -- фильтрация

### 2.2. TestCaseVersion

Снимок версии тест-кейса для версионирования и отката.

| Поле | Тип | Ограничения | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| version_number | int | positive | Номер версии |
| snapshot | JSON (map[string]interface{}) | required | Полный снимок состояния |
| change_description | text | optional | Описание изменений |
| created_at | time | immutable | Дата создания |

**Рёбра:**
- `test_case` (From TestCase) -- M:1, required
- `created_by` (From User) -- M:1

**Индексы:** `version_number`

### 2.3. TestSuite

Папки/секции для организации тест-кейсов (древовидная структура).

| Поле | Тип | Ограничения | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| name | string | not empty, max 200 | Название |
| description | text | optional | Описание |
| sort_order | int | default: 0 | Порядок сортировки |
| created_at | time | immutable | Дата создания |
| updated_at | time | update: now | Дата обновления |

**Рёбра:**
- `project` (From Project) -- M:1, required
- `test_cases` (To TestCase) -- 1:M
- `children` / `parent` (self-ref) -- дерево папок

**Индексы:** `name`

### 2.4. TestRun

Прогон тестов (ручной, автоматический или смешанный).

| Поле | Тип | Ограничения | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| title | string | not empty, max 500 | Название |
| description | text | optional | Описание |
| run_type | enum | "manual", "automatic", "mixed", default: "manual" | Тип прогона |
| status | enum | "created", "in_progress", "completed", "aborted", default: "created" | Статус |
| total_count | int | default: 0 | Всего тестов |
| passed_count | int | default: 0 | Пройдено |
| failed_count | int | default: 0 | Провалено |
| skipped_count | int | default: 0 | Пропущено |
| broken_count | int | default: 0 | Сломано |
| unknown_count | int | default: 0 | Неизвестно |
| allure_run_id | string | optional | ID прогона Allure |
| started_at | time | optional, nillable | Время запуска |
| completed_at | time | optional, nillable | Время завершения |
| created_at | time | immutable | Дата создания |
| updated_at | time | update: now | Дата обновления |

**Рёбра:**
- `project` (From Project) -- M:1, required
- `created_by` (From User) -- M:1, required
- `results` (To TestResult) -- 1:M
- `assignees` (To User) -- M:M

**Индексы:** `status`, `run_type`, `allure_run_id`, `created_at`

### 2.5. TestResult

Результат выполнения отдельного теста в прогоне.

| Поле | Тип | Ограничения | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| status | enum | "passed", "failed", "broken", "skipped", "unknown", default: "unknown" | Статус |
| comment | text | optional | Комментарий |
| step_results | JSON ([]StepResult) | optional | Результаты шагов |
| duration_ms | int64 | optional, nillable | Длительность (мс) |
| allure_uuid | string | optional | UUID из Allure |
| error_message | text | optional | Сообщение об ошибке |
| error_trace | text | optional | Стектрейс ошибки |
| is_flaky | bool | default: false | Нестабильный тест |
| match_strategy | enum | "annotation", "fullname", "classmethod", "unmapped", default: "unmapped" | Стратегия матчинга |
| started_at | time | optional, nillable | Время начала |
| completed_at | time | optional, nillable | Время завершения |
| created_at | time | immutable | Дата создания |

**Рёбра:**
- `run` (From TestRun) -- M:1, required
- `test_case` (From TestCase) -- M:1, optional (unmapped результаты)
- `executed_by` (From User) -- M:1, optional
- `files` (To FileMetadata) -- M:M

**Индексы:** `status`, `allure_uuid`, `is_flaky`

### 2.6. Component

Компоненты для категоризации тест-кейсов.

| Поле | Тип | Ограничения | Описание |
|------|-----|-------------|----------|
| id | UUID | PK, immutable | Идентификатор |
| name | string | not empty, max 200 | Название |
| description | text | optional | Описание |
| created_at | time | immutable | Дата создания |

**Рёбра:**
- `project` (From Project) -- M:1, required
- `test_cases` (From TestCase) -- M:M (backref)

**Индексы:** `name`

### 2.7. Диаграмма связей

```
Project ─1:M─> TestCase ─M:M─> Issue (linked_issues / linked_test_cases)
   |              |
   |              +─1:M─> TestCaseVersion
   |              +─M:1─> TestSuite (suite)
   |              +─M:M─> Label
   |              +─M:M─> Component
   |              +─1:M─> TestResult
   |
   +──1:M─> TestRun ─1:M─> TestResult ─M:M─> FileMetadata
   |            |              |
   |            +─M:M─> User   +─M:1─> User (executed_by)
   |            (assignees)    +─M:1─> TestCase
   |
   +──1:M─> TestSuite ─self─> TestSuite (parent/children)
   |
   +──1:M─> Component
```

---

## 3. MongoDB коллекции

### 3.1. allure_results

Коллекция для хранения сырого JSON результатов Allure.

**Структура документа:**
```json
{
  "_id": ObjectId,
  "run_id": "uuid-string",
  "results": [
    {
      "uuid": "...",
      "fullName": "com.app.LoginTest.testLogin",
      "name": "testLogin",
      "status": "passed",
      "statusDetails": {"message": "...", "trace": "..."},
      "start": 1711234567890,
      "stop": 1711234567990,
      "steps": [...],
      "labels": [{"name": "suite", "value": "LoginTest"}],
      "links": [{"type": "tms", "name": "TC-42"}],
      "parameters": [...],
      "attachments": [...]
    }
  ],
  "created_at": ISODate
}
```

**Индексы:**
- `run_id` -- поиск результатов по прогону

**Назначение:** Сырой JSON сохраняется для аудита и отладки парсера. Структурированные данные извлекаются в PostgreSQL (TestResult). Дублирование обеспечивает возможность повторного парсинга при изменении логики матчинга.

---

## 4. API эндпоинты

### 4.1. Тест-кейсы

| Метод | Путь | Тело запроса | Ответ | RBAC |
|-------|------|-------------|-------|------|
| POST | /projects/:projectId/tms/cases | CreateTestCaseRequest | TestCaseResponse | PermTestCaseCreate |
| GET | /projects/:projectId/tms/cases | query: offset, limit, suite_id, status, priority, search | PaginatedResponse[TestCaseResponse] | PermProjectView |
| GET | /projects/:projectId/tms/cases/:caseKey | - | TestCaseResponse | PermProjectView |
| PUT | /projects/:projectId/tms/cases/:caseKey | UpdateTestCaseRequest | TestCaseResponse | PermTestCaseEdit |
| DELETE | /projects/:projectId/tms/cases/:caseKey | - | 204 | PermTestCaseEdit |

### 4.2. Версии тест-кейсов

| Метод | Путь | Ответ | RBAC |
|-------|------|-------|------|
| GET | /projects/:projectId/tms/cases/:caseKey/versions | []TestCaseVersionResponse | PermProjectView |
| POST | /projects/:projectId/tms/cases/:caseKey/rollback/:versionId | TestCaseResponse | PermTestCaseEdit |

### 4.3. Привязка к задачам

| Метод | Путь | Тело | Ответ | RBAC |
|-------|------|------|-------|------|
| POST | /projects/:projectId/tms/cases/:caseKey/links | LinkIssueRequest | 200 | PermTestCaseEdit |
| GET | /projects/:projectId/tms/cases/:caseKey/links | - | []IssueBriefResponse | PermProjectView |
| DELETE | /projects/:projectId/tms/cases/:caseKey/links/:issueId | - | 204 | PermTestCaseEdit |

### 4.4. Папки (Suites)

| Метод | Путь | Тело | Ответ | RBAC |
|-------|------|------|-------|------|
| POST | /projects/:projectId/tms/suites | CreateSuiteRequest | SuiteResponse | PermTestCaseCreate |
| GET | /projects/:projectId/tms/suites | - | []SuiteResponse (дерево) | PermProjectView |
| PUT | /projects/:projectId/tms/suites/:suiteId | UpdateSuiteRequest | SuiteResponse | PermTestCaseEdit |
| DELETE | /projects/:projectId/tms/suites/:suiteId | - | 204 | PermTestCaseEdit |
| POST | /projects/:projectId/tms/suites/:suiteId/move | MoveSuiteRequest | SuiteResponse | PermTestCaseEdit |

### 4.5. Компоненты

| Метод | Путь | Тело | Ответ | RBAC |
|-------|------|------|-------|------|
| POST | /projects/:projectId/tms/components | CreateComponentRequest | ComponentResponse | PermTestCaseCreate |
| GET | /projects/:projectId/tms/components | - | []ComponentResponse | PermProjectView |
| DELETE | /projects/:projectId/tms/components/:componentId | - | 204 | PermTestCaseEdit |

### 4.6. Прогоны (Runs)

| Метод | Путь | Тело | Ответ | RBAC |
|-------|------|------|-------|------|
| POST | /projects/:projectId/tms/runs | CreateTestRunRequest | TestRunResponse | PermTestRunCreate |
| GET | /projects/:projectId/tms/runs | query: offset, limit, status | PaginatedResponse[TestRunResponse] | PermProjectView |
| GET | /projects/:projectId/tms/runs/:runId | - | TestRunDetailResponse | PermProjectView |
| POST | /projects/:projectId/tms/runs/:runId/cases | AddTestCasesToRunRequest | 200 | PermTestRunCreate |
| POST | /projects/:projectId/tms/runs/:runId/assignees | AssignTestersRequest | 200 | PermTestRunCreate |
| POST | /projects/:projectId/tms/runs/:runId/start | - | TestRunResponse | PermTestRunExecute |
| POST | /projects/:projectId/tms/runs/:runId/complete | - | TestRunResponse | PermTestRunExecute |
| POST | /projects/:projectId/tms/runs/:runId/reopen | - | TestRunResponse | PermTestRunExecute |

### 4.7. Результаты (Results)

| Метод | Путь | Тело | Ответ | RBAC |
|-------|------|------|-------|------|
| PUT | /projects/:projectId/tms/runs/:runId/results/:resultId | UpdateTestResultRequest | TestResultResponse | PermTestRunExecute |
| POST | /projects/:projectId/tms/runs/:runId/results/:resultId/bug | CreateBugRequest | Issue | PermTestRunExecute |

### 4.8. Allure

| Метод | Путь | Тело | Ответ | RBAC |
|-------|------|------|-------|------|
| POST | /projects/:projectId/tms/allure/import | ImportAllureRequest (JSON) | TestRunResponse | PermTestCaseCreate |
| POST | /projects/:projectId/tms/runs/:runId/allure | ImportAllureRequest | TestRunResponse | PermTestRunExecute |

### 4.9. Отчёты

| Метод | Путь | Параметры | Ответ | RBAC |
|-------|------|-----------|-------|------|
| GET | /projects/:projectId/tms/reports/coverage | - | CoverageReport | PermProjectView |
| GET | /projects/:projectId/tms/reports/trends | query: limit | []TrendPoint | PermProjectView |
| GET | /projects/:projectId/tms/reports/flaky | - | []TestCaseResponse | PermProjectView |
| GET | /projects/:projectId/tms/reports/comparison | query: run1, run2 | RunComparison | PermProjectView |
| GET | /projects/:projectId/tms/reports/history | - | []TestRunSummary | PermProjectView |

### 4.10. Импорт/Экспорт

| Метод | Путь | Тело/Параметры | Ответ | RBAC |
|-------|------|----------------|-------|------|
| POST | /projects/:projectId/tms/import | multipart: file, format, mapping | ImportResult | PermTestCaseCreate |
| GET | /projects/:projectId/tms/export | query: format (csv/json/xml/excel) | файл | PermProjectView |
| GET | /projects/:projectId/tms/runs/:runId/export | query: format (csv/json/pdf) | файл | PermProjectView |

---

## 5. Allure интеграция

### 5.1. Формат allure-results JSON

Структуры описаны в `allure_types.go`. Все поля опциональны для защитного парсинга:

```go
type AllureResult struct {
    UUID          string               `json:"uuid,omitempty"`
    HistoryID     string               `json:"historyId,omitempty"`
    TestCaseID    string               `json:"testCaseId,omitempty"`
    FullName      string               `json:"fullName,omitempty"`
    Name          string               `json:"name,omitempty"`
    Status        string               `json:"status,omitempty"`
    StatusDetails *AllureStatusDetails  `json:"statusDetails,omitempty"`
    Start         *int64               `json:"start,omitempty"`
    Stop          *int64               `json:"stop,omitempty"`
    Steps         []AllureStep          `json:"steps,omitempty"`
    Labels        []AllureLabel         `json:"labels,omitempty"`
    Links         []AllureLink          `json:"links,omitempty"`
    Parameters    []AllureParameter     `json:"parameters,omitempty"`
    Attachments   []AllureAttachment    `json:"attachments,omitempty"`
    Description   string               `json:"description,omitempty"`
}
```

**Используемые поля при импорте:**
- `status` -- преобразуется в статус TMS (passed/failed/broken/skipped/unknown)
- `fullName` -- для матчинга (стратегия 3)
- `name` -- для матчинга (стратегия 4)
- `labels` -- suite для матчинга, AS_ID/tms для аннотации
- `links` -- type="tms" для матчинга (стратегия 1)
- `statusDetails.message` -- error_message результата
- `statusDetails.trace` -- error_trace результата
- `start`, `stop` -- время выполнения, вычисление duration_ms

### 5.2. Стратегии сопоставления

Матчер (`allure_matcher.go`) проверяет стратегии в порядке приоритета:

| Приоритет | Стратегия | Реализация | Метод репозитория |
|-----------|----------|------------|-------------------|
| 1 | Аннотация | Links type="tms" name -> external_id | FindTestCaseByExternalID |
| 2 | Аннотация (label) | Labels name="AS_ID"/"tms" value -> external_id | FindTestCaseByExternalID |
| 3 | FullName | result.FullName -> test_case.full_name | FindTestCaseByFullName |
| 4 | Class+Method | suite label + result.Name -> class_name + method_name | FindTestCaseByClassMethod |
| - | Unmapped | ни одна стратегия не сработала | - |

### 5.3. Процесс импорта

```
1. Получение JSON -> ParseAllureResults (allure_parser.go)
2. Сохранение сырого JSON в MongoDB (StoreAllureResults)
3. Создание TestRun (тип: automatic)
4. Для каждого результата:
   a. MatchTestCase -> определение стратегии и связанного тест-кейса
   b. Формирование CreateTestResultInput
5. BatchCreateTestResults -> массовая вставка в PostgreSQL
6. recomputeRunStats -> пересчёт счётчиков прогона
7. Завершение прогона (completed)
8. publishEvent(tms.testrun.completed) -> NATS
```

### 5.4. Обработка ошибок

- Невалидный JSON: `ParseAllureResults` возвращает ошибку с описанием.
- Результат без статуса: предупреждение в slog, статус -> "unknown".
- Ошибка матчинга: результат создаётся с `match_strategy: "unmapped"`, без привязки к тест-кейсу.
- Ошибка записи в MongoDB: логируется, не блокирует создание результатов в PostgreSQL.

---

## 6. Кросс-модульные связи

### 6.1. TestCase <-> Issue (Ent edges)

```go
// testcase.go
edge.To("linked_issues", Issue.Type)  // M:M

// issue.go (backref)
edge.From("linked_test_cases", TestCase.Type).Ref("linked_issues")
```

Двусторонняя связь позволяет:
- Из тест-кейса видеть привязанные задачи.
- Из задачи видеть привязанные тест-кейсы (для покрытия).

### 6.2. Создание бага из упавшего теста

```go
func (s *TMSService) CreateBugFromFailedResult(ctx, projectID, userID, resultID uuid.UUID) (*ent.Issue, error)
```

Использует `Repository.CreateBugIssue`, который:
1. Создаёт Issue через общий `ent.Client` (не импортирует tracker).
2. Устанавливает type = "bug", status = "open".
3. Привязывает Issue к TestCase через edge.

### 6.3. NATS события

| Событие | Когда публикуется | Payload | Подписчики |
|---------|-------------------|---------|------------|
| `tms.testrun.created` | Создание прогона | TestRunCreatedEvent | -- |
| `tms.testrun.completed` | Завершение прогона | TestRunCompletedEvent | Notification Module (Telegram) |
| `tms.coverage.updated` | Привязка/отвязка тест-кейса | CoverageUpdatedEvent | Dashboard |
| `tms.testcase.linked` | Привязка к задаче | TestCaseLinkedEvent | Tracker UI update |

Публикация nil-safe: если NATS-клиент nil (в тестах), событие не публикуется, ошибки нет.

### 6.4. Как трекер видит связанные тест-кейсы

Issue.Edges.LinkedTestCases -- обратный edge, добавленный в Issue schema. Позволяет:
- В карточке задачи показать список привязанных тест-кейсов.
- На дашборде проекта показать покрытие задач тестами.

---

## 7. Импорт/Экспорт

### 7.1. Поддерживаемые форматы и библиотеки

| Формат | Импорт | Экспорт | Библиотека |
|--------|--------|---------|------------|
| CSV | ImportFromCSV | ExportTestCasesToCSV | encoding/csv (stdlib) |
| JSON | ImportFromJSON | ExportTestCasesToJSON | encoding/json (stdlib) |
| XML | ImportFromXML | ExportTestCasesToXML | encoding/xml (stdlib) |
| Excel | ImportFromExcel | ExportTestCasesToExcel | excelize/v2 |
| PDF | - | ExportTestRunToPDF | maroto/v2 |

### 7.2. Маппинг полей

Автоопределение (`AutoDetectMapping`) поддерживает двуязычные заголовки:

| Заголовок (EN) | Заголовок (RU) | Поле TMS |
|----------------|----------------|----------|
| title, name, test name | название | title |
| description | описание | description |
| precondition | предусловие | precondition |
| steps | шаги | steps |
| priority | приоритет | priority |
| status | статус | status |
| type | тип | case_type |

Ручной маппинг передаётся через JSON в параметре `mapping` multipart-запроса.

### 7.3. Валидация данных при импорте

`ValidateImportData` проверяет:
- Обязательность `title` (пустое название -> ошибка).
- Допустимые значения `priority` (critical/high/medium/low/trivial).
- Ошибки с номером строки (`Row` = индекс + 2 для учёта заголовка).

Невалидные строки пропускаются, валидные создаются. Результат: `ImportResult{Created, Skipped, Errors, ValidationErrors}`.

### 7.4. Генерация PDF (maroto)

Структура PDF-отчёта прогона:

```
+-------------------------------------------------+
| Название прогона (16pt, bold, center)            |
| Создан: 2026-03-21 15:04:05 (9pt, gray, center) |
+-------------------------------------------------+
| Всего | Passed | Failed | Skipped | Broken | %  |
|  50   |   45   |   3    |    1    |   1    | 90 |
+-------------------------------------------------+
| Ключ  | Название | Статус | Время  | Коммент. |
| TC-1  | Login    | passed | 120ms  | OK       |
| TC-2  | Logout   | failed | 450ms  | Timeout  |
+-------------------------------------------------+
```

Цветовая индикация: passed -- зелёный, failed -- красный, broken -- оранжевый.

---

## 8. Безопасность и производительность

### 8.1. RBAC проверки

Разрешения проверяются на уровне маршрутов через `rbac.RequirePermission`:

```go
cases.Post("/", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.CreateTestCase)
cases.Put("/:caseKey", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.UpdateTestCase)
runs.Post("/", rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc), h.CreateTestRun)
```

| Разрешение | Bitmap | Операции |
|-----------|--------|----------|
| PermTestCaseCreate | bit 18 | Создание кейсов, импорт |
| PermTestCaseEdit | bit 19 | Редактирование, удаление, откат, привязка |
| PermTestRunCreate | bit 20 | Создание прогонов, добавление кейсов |
| PermTestRunExecute | bit 21 | Выполнение тестов, старт/стоп прогона |
| PermProjectView | bit 0 | Просмотр кейсов, прогонов, отчётов |

### 8.2. Предвычисленная статистика

Счётчики (total, passed, failed, skipped, broken, unknown) хранятся непосредственно на TestRun:
- При каждом обновлении результата вызывается `recomputeRunStats`.
- Метод подсчитывает все результаты прогона группировкой по статусу.
- Избавляет от агрегации при каждом GET-запросе прогона.

### 8.3. Пакетная вставка

`BatchCreateTestResults` вставляет все результаты Allure-импорта за одну транзакцию:
- Используется Ent bulk create builder.
- Минимизирует количество roundtrip к БД.

### 8.4. Индексы на критичных полях

| Таблица | Поля индекса | Назначение |
|---------|-------------|------------|
| test_cases | case_key | Поиск кейса по ключу (уникальный) |
| test_cases | external_id | Allure матчинг по аннотации |
| test_cases | full_name | Allure матчинг по полному имени |
| test_cases | class_name, method_name | Allure матчинг по классу+методу |
| test_runs | allure_run_id | Поиск прогона Allure |
| test_runs | created_at | Сортировка истории |
| test_results | allure_uuid | Предотвращение дублей |
| test_results | is_flaky | Фильтрация flaky-тестов |
| allure_results (MongoDB) | run_id | Поиск сырых данных по прогону |

---

## 9. Технический долг и известные ограничения

### 9.1. Дерево папок строится в памяти

Текущая реализация загружает все папки проекта и строит дерево в Go-коде:
- Для проектов с <1000 папками это приемлемо.
- CTE (Common Table Expression) для рекурсивных запросов не реализован.
- При масштабировании потребуется оптимизация.

### 9.2. Allure attachments не загружаются

Парсер обрабатывает метаданные вложений (имя, тип, источник), но не загружает сами файлы:
- JSON метаданные сохраняются в MongoDB.
- Загрузка бинарных вложений (скриншотов, логов) из CI требует отдельного API.
- Может быть реализовано в Phase 7.

### 9.3. Тестовые окружения отложены

TMS-07 (тестовые окружения: dev/staging/prod, конфигурации браузер/ОС) отложен до Phase 7:
- Требует добавления сущности TestEnvironment и связей.
- Не критично для первоначальной функциональности.

### 9.4. Пересчёт статистики при каждом обновлении

`recomputeRunStats` считает все результаты прогона при каждом обновлении:
- Для прогонов с <10 000 результатами это приемлемо.
- При масштабировании можно перейти на инкрементальное обновление счётчиков.

### 9.5. Отсутствует кеширование отчётов

Отчёты (trends, flaky, coverage) вычисляются на лету при каждом запросе:
- Для 50 пользователей и умеренного количества данных нагрузка приемлема.
- Redis-кеш с инвалидацией по событиям может быть добавлен при необходимости.
