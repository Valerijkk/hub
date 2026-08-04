# Аналитика: системный анализ

> Часть плейбука. Канон и оглавление — [`../analysis.md`](../analysis.md).

Функциональные и нефункциональные требования, модель данных, API.

---

## 4. Системный анализ (SA)

Владелец раздела: `sa1`, `sa2`. Основной артефакт: `docs/system_specification.md` +
`docs/wiki/API.md`. Верификатор: `sa-ver`. Вспомогательно: `docs/architecture.md`
(согласование контрактов), `docs/test_plan.md` (матрица покрытия), `.planning/COVERAGE.md`.

### 4.1 Что делает SA и чего не делает

**Делаю:** декомпозиция User Story в FR/NFR, проектирование ERD, API-контрактов, DFD,
матрицы трассируемости требование→API→код→тест, ведение `COVERAGE.md`, gap-анализ.

**НЕ делаю:** код, архитектурные решения о модульности/стеке (→ ARCH), бизнес-приоритеты
(→ BA).

### 4.2 Функциональные требования (FR)

FR — что система должна делать, на уровне поведения, а не реализации. В Hub FR группируются
по модулям с префиксом модуля (`AUTH`, `TASK`, `MSG`, `TMS`, …). Нумерация: `<PREFIX>-<NN>`.

Формат FR (из SA_skill):

```
REQ-ID: <MODULE>-<NN>
Source: US-M<NN>            # какая User Story порождает
Technical: endpoint / schema / constraint
NFR link: NFR-<NN>          # связанные нефункциональные требования
```

Требования тестируемости FR:
- Один FR — одно проверяемое поведение (атомарность).
- FR ссылается на порождающую US (`Source`) — иначе он «висит» и попадёт под вопрос sa-ver.
- FR формулируется в терминах наблюдаемого результата, а не внутренней реализации.

#### Good / Bad: FR

```
GOOD
REQ-ID: TMS-14
Source: US-S16
Technical: PATCH /api/tms/cases/:id { "flaky": true } — помечает кейс flaky;
           GET /api/tms/cases?flaky=true — фильтр; счётчик нестабильных прогонов за 30 дней
           хранится в TestCase.flaky_count (пересчёт при submit результата)
NFR link: NFR-14 (pagination offset/limit), NFR-24 (optimistic lock version)
Acceptance: соответствует AC US-S16 (флаг, фильтр, автоподсветка при 3+ исходах)
```

```
BAD
REQ-ID: TMS-14
Требование: сделать flaky-кейсы
- Нет Source (какая US?)
- Нет контракта (endpoint/schema)
- Нет связи с NFR
- «Сделать» — не проверяемое поведение
```

### 4.3 Нефункциональные требования (NFR)

NFR — ограничения качества: производительность, безопасность, надёжность, совместимость.
В Hub NFR имеют сквозную нумерацию `NFR-01…NFR-NN` и обязаны быть **числовыми/проверяемыми**.

Образцы стиля из проекта (реальные NFR):

```
| NFR-01 | Response time API (p95) | < 500ms для CRUD |
| NFR-05 | Auth rate limit | 3 попытки / 10 мин |
| NFR-13 | Error format | {"error": "...", "request_id": "uuid"} |
| NFR-24 | Optimistic locking | version int на каждой изменяемой Ent |
| NFR-25 | Primary keys | UUID v4 only — string IDs запрещены |
```

Требования тестируемости NFR:
- Числовой порог или однозначный факт (не «быстро», а «p95 < 500ms»).
- Указан способ проверки (нагрузочный тест, статический анализ, наличие поля в схеме).
- NFR не противоречит `tz_constraints.md` (mandatory-ограничения — источник многих NFR).

#### Good / Bad: NFR

```
GOOD
NFR-31 | Глубина вложенности тредов мессенджера | ≤ 1 уровень (тред от корневого сообщения);
         проверка: попытка ответить в ответ возвращает 422 {"error","request_id"}
```

```
BAD
NFR-31 | Мессенджер должен быть производительным и удобным
- Не измеримо, не проверяемо, нет порога, нет способа проверки → sa-ver вернёт FAIL
```

### 4.4 ERD (модель данных)

SA проектирует сущности, атрибуты, связи, ключи. Hub использует Ent + PostgreSQL для
реляционных данных, MongoDB для сообщений/отчётов, Redis для сессий/лимитов.

Обязательные инварианты изменяемой Ent-сущности (из TZ и NFR):

| Инвариант | Требование |
|-----------|-----------|
| Primary key | `id` UUID v4, `field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable()` — строки запрещены (NFR-25) |
| Оптимистичная блокировка | `version` int на каждой изменяемой схеме (NFR-24) |
| Время | `created_at`, `updated_at` в UTC; TZ конвертирует фронт (NFR-12) |
| Soft delete | `deleted_at` или запись в `TrashItem`, retention 30 дней (NFR-11) |

Шаблон описания сущности:

```
### Сущность: TestCase
Поля:
  - id: UUID v4, PK, immutable
  - project_id: UUID, FK → Project (cascade on project delete → soft)
  - title: string, not null
  - kind: enum {classic, bdd}
  - flaky: bool, default false
  - flaky_count: int, default 0            # NFR: пересчёт при submit результата
  - version: int                            # optimistic lock (NFR-24)
  - created_at, updated_at: UTC timestamp
  - deleted_at: UTC timestamp, nullable     # soft delete (NFR-11)
Связи:
  - TestCase 1—N TestRun
  - TestCase N—1 Issue (nullable, issue_id) # traceability тест→баг
Индексы:
  - (project_id, deleted_at) для листинга активных кейсов
  - (project_id, flaky) для фильтра flaky
```

Текстовая ER-диаграмма (образец стиля проекта — псевдографика допустима):

```
Project ──┬── Issue ──┬── Comment
          │           └── IssueLink
          ├── TestCase ── TestRun ── TestResult
          └── Channel
TestCase N──1 Issue   (issue_id, traceability)
```

#### Good / Bad: ERD

```
GOOD
- Issue.id UUID v4, version int, deleted_at nullable — все инварианты соблюдены
- Связь TestCase→Issue объявлена явно с nullable FK и described cascade
- Индекс под фильтр flaky задан
```

```
BAD
- TestCase.id как string "TC-42" → нарушает NFR-25 (UUID only)
- Нет version → нарушает NFR-24 (optimistic lock), гонки при конкурентной правке
- Hard delete без TrashItem → нарушает NFR-11 (soft delete 30 дней)
- Связь TestCase↔Issue «подразумевается», но не описана → sa-ver вернёт FAIL
```

### 4.5 API-контракты

SA описывает REST-контракты в `system_specification.md` (сводно) и `docs/wiki/API.md`
(детально). Обязательные соглашения Hub:

| Аспект | Правило |
|--------|---------|
| Base URL | `/api` |
| Auth | `Authorization: Bearer <jwt>` или `X-API-Key: <key>` |
| Content-Type | `application/json` |
| Формат ошибки | `{"error": "message", "request_id": "uuid"}` (NFR-13) |
| Пагинация задач | offset/limit: `?offset=0&limit=50` (NFR-14) |
| Пагинация чата | cursor (infinite scroll): `?cursor=...` (NFR-15) |
| Rate limits | API 100/сек (NFR-06), сообщения 60/мин (NFR-07), файлы 10/мин (NFR-08) |
| Коды ответов | 200/201 успех, 400 валидация, 401 auth, 403 RBAC, 404 not found, 409 конфликт версии, 422 бизнес-правило, 429 rate limit |

Шаблон контракта эндпоинта:

```
### POST /api/projects/:id/issues
Назначение: создать задачу в проекте
Auth: Bearer JWT; RBAC: право issue.create в проекте
Request body:
  {
    "title": "string, required, 1..255",
    "type": "enum: task|bug|story|... (9 типов)",
    "assignee_id": "uuid, optional"
  }
Responses:
  201 { "id": "uuid", "key": "HUB-123", "version": 1, ... }
  400 { "error": "title is required", "request_id": "uuid" }
  403 { "error": "forbidden", "request_id": "uuid" }
  409 { "error": "version conflict", "request_id": "uuid" }   # optimistic lock
Пагинация: N/A (create)
Идемпотентность: нет
Source FR: TASK-01 (Source US-M03)
```

Шаблон контракта листинга с пагинацией:

```
### GET /api/projects/:id/issues?offset=0&limit=50&status=open
Auth: Bearer JWT; RBAC: issue.read
Query: offset (int, ≥0), limit (int, 1..100, default 50), фильтры status/assignee/type
Response 200:
  {
    "items": [ { "id": "uuid", "key": "HUB-123", ... } ],
    "total": 1234,
    "offset": 0,
    "limit": 50
  }
Ошибки: 400 при limit>100 {"error","request_id"}
```

#### Good / Bad: API-контракт

```
GOOD
- Формат ошибки строго {"error","request_id"} на всех кодах
- Пагинация offset/limit для задач, cursor для сообщений — по NFR
- 409 при конфликте version (optimistic lock) описан явно
- Source FR указан → трассируемость
```

```
BAD
- Ошибка возвращается как {"message":"..."} → нарушает NFR-13
- Листинг задач без пагинации отдаёт весь список → нарушает NFR-14, DoS-риск
- Нет 409/422 → неясно поведение при конфликте версии/бизнес-правиле
- Нет Source FR → эндпоинт не привязан к требованию
```

### 4.6 DFD (Data Flow Diagram)

DFD показывает поток данных через слои. Образец стиля проекта (создание задачи):

```
[User] → POST /issues → [Auth Middleware] → [RBAC Check]
    → [IssueService.Create] → [IssueRepository] → [PostgreSQL]
    → [Outbox.Publish] → [NATS: hub.issues.created]
    → [WS Hub] → подписчики room project:<id>
    → [Notification] → in-app / Telegram / Web Push
```

DFD обязателен для сценариев с несколькими слоями/интеграциями (WS, NATS, внешние webhooks),
потому что именно там прячутся security-инварианты (RBAC-проверка до сервиса, подпись
webhook, user-scope фильтр). SA фиксирует, где проходят эти проверки.

### 4.7 Матрица трассируемости требование → API → код → тест

SA поддерживает сквозную матрицу в `.planning/COVERAGE.md`. Это ключевой артефакт, связывающий
все слои. Методология (из COVERAGE.md): Code = ✅/⚠️/❌/🔧, Tests = ✅/⚠️/❌.

Шаблон строки матрицы:

```
| US | FR | API | Code (файл) | Tests | Cov | Gap |
|----|----|----|-------------|-------|-----|-----|
| US-S16 | TMS-14 | PATCH /tms/cases/:id, GET /tms/cases?flaky | tms/handler_cases.go, tms/service.go | testframework/tms/*_test.go | 82.0% (2026-06-22) | — |
```

Правила ведения матрицы:
- Числа (endpoint count, coverage %) = проверяемые факты, сверенные с кодом/Docker-прогоном.
- Каждая US имеет строку; строки без Code/Tests подсвечиваются как Gap.
- Матрица — источник для `gap-анализа`: что обещано (US/FR) vs что реализовано/покрыто.

### 4.8 Полный шаблон вклада SA в system_specification.md

```markdown
## Функциональные требования (модуль TMS)
| ID | Source | Поведение | NFR link |
|----|--------|-----------|----------|
| TMS-14 | US-S16 | флаг flaky + история нестабильности | NFR-14, NFR-24 |
## Нефункциональные требования (новые)
| ID | Требование | Значение |
|----|------------|----------|
| NFR-31 | <ограничение> | <числовой порог> |
## Модель данных (ERD дополнение)
### Сущность: TestCase (изменения)
- +flaky bool, +flaky_count int
- индекс (project_id, flaky)
## API Контракты (дополнение wiki/API.md)
### PATCH /api/tms/cases/:id
<контракт по шаблону 4.5>
## DFD — <сценарий>
<поток по шаблону 4.6>
## Матрица трассируемости (COVERAGE.md)
| US | FR | API | Code | Tests | Gap |
```

### 4.9 Handoff SA → ARCH и SA → BE

После обновления спецификации SA:
1. Описание простым языком (2–5 предложений).
2. Обновляет `COVERAGE.md`, `.planning/STATE.md`, лог в `SA_skill.md`.
3. `verified: false` → `sa-ver`.
4. После `verified: true`:
   - если меняется модульность/стек/интеграция → handoff `arch`;
   - иначе → handoff `be`/`fe` с указанием FR/контрактов и ссылок на AC.

Пример handoff-заметки:

```
SA handoff → ARCH + BE (2026-07-02)
FR: TMS-14 (flaky-аналитика). Контракты: PATCH /tms/cases/:id, GET /tms/cases?flaky.
ERD: +flaky, +flaky_count в TestCase, индекс (project_id, flaky). Новых NFR: NFR-31.
Для ARCH: пересчёт flaky_count при submit — нужен ли outbox-event? граница модуля tms↔notification.
Для BE: соблюсти NFR-13 (ошибки), NFR-24 (version), NFR-14 (пагинация фильтра).
verified: false → sa-ver
```

---
