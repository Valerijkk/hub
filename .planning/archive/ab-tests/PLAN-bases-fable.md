# PLAN: Bases → паритет с Lark Base

> **A/B-тест оркестрации — вариант Fable 5 (max effort)**
> Дата: 2026-07-02 · Автор: team-lead+architect (Fable 5) · Read-only анализ кода
> Источники: `.planning/archive/july-2026/EXPANSION-ANALYSIS-2026-07-02.md` (эпик B5, XL), `backend/internal/bases/*`,
> `backend/ent/schema/hub{base,table,column,record}.go`, `frontend/src/modules/bases/bases-page.tsx`,
> `.cursor/playbooks/team-workflow.md`, `.cursor/rules/60-tz-constraints.mdc`, `20-backend.mdc`, `30-frontend.mdc`

---

## 1. Анализ текущего состояния

### 1.1 Что есть (по фактическому коду)

| Слой | Файлы | Объём | Содержание |
|---|---|---|---|
| Ent-схемы | `hubbase.go`, `hubtable.go`, `hubcolumn.go`, `hubrecord.go` | 4 схемы | UUID PK ✅, `version` ✅, `deleted_at` soft-delete ✅, UTC ✅. Record хранит cells как JSON `map[columnUUID]value` |
| Backend | `handler.go`, `service.go`, `repository.go`, `dto.go`, `export_hooks.go` | 855 стр | CRUD base/table/column/record; layering чистый (Repository interface, `*ent.Client` только в EntRepository); optimistic lock по `version`; error mapping 400/404/409/500 |
| Frontend | `modules/bases/bases-page.tsx` + `api/bases.ts` + `routes/_authenticated/bases.tsx` | 1 файл / 341 стр | Список баз, табы таблиц, plain-HTML `<table>` grid, inline-edit по blur, добавление колонок/записей. TanStack Query, invalidate-on-mutate |
| Типы колонок | enum в `hubcolumn.go` | 7 типов | text, number, date, select, checkbox, url, user_ref |
| Тесты | `export_hooks.go` (hubtest) | — | Только helpers-экспорт; cov bases 78.6% (ниже gate 82%) |

### 1.2 Чего нет (gap-матрица к Lark Base)

| Возможность Lark Base | Статус в Hub | Критичность |
|---|---|---|
| Formula-поля (вычисляемые, функции) | ❌ нет | ГЛАВНЫЙ gap |
| Relations (link между таблицами), lookup, rollup | ❌ нет | Критично |
| Views: grid / kanban / calendar / gallery | ❌ только один hardcoded grid | Критично |
| Фильтры / сортировки / группировки (персистентные, per-view) | ❌ нет (сервер отдаёт все записи `Order(updated_at desc)`) | Критично |
| Sharing / права (per-base роли, доступ по ссылке) | ❌ нет: `ListBases` отдаёт **все** базы всем аутентифицированным, любой может редактировать чужую базу | **Security-gap** (аналог NF-03) |
| Пагинация записей | ❌ `GetTableView` грузит всю таблицу целиком | Perf-gap |
| Импорт/экспорт (CSV/XLSX) | ❌ нет (модуль `imex` в проекте есть — переиспользовать) | Средняя |
| API для no-code автоматизаций | ❌ нет (модуль `apigateway` есть — точка интеграции) | Средняя |
| Типы полей: multi-select, attachment, phone/email, rating, created_at/by auto | ❌ 7 из ~15 | Средняя |
| Real-time коллаборация (WS-обновления grid) | ❌ нет | Средняя |
| История записи / audit trail | ❌ нет (модуль `audit` есть) | Низкая |

### 1.3 Что переиспользуем из платформы

`imex` (импорт/экспорт), `apigateway` (API-ключи для автоматизаций), `ws` + NATS JetStream (real-time),
`rbac` (паттерн permission-checks), `audit` (история), `storage`/SeaweedFS (attachment-поля),
`search`/Meilisearch (поиск по записям — отдельный backlog).

---

## 2. Целевая модель данных (Ent)

Все схемы: `field.UUID("id").Default(uuid.New).Immutable()` + `version int` + `created_at/updated_at` UTC + `deleted_at` soft-delete + индексы по `deleted_at` (инварианты `60-tz-constraints`).

### 2.1 Изменяемые существующие схемы

| Схема | Изменение |
|---|---|
| `HubColumn` | + enum-значения `col_type`: `multi_select`, `email`, `phone`, `rating`, `attachment`, `link` (relation), `lookup`, `rollup`, `formula`, `created_time`, `created_by_auto`. + поле `config JSON` (typed-настройки: formula text, link target, rollup agg) — отдельно от display-`options` |
| `HubRecord` | + composite-индекс `(table_id via edge, deleted_at)`; cells JSON остаётся (schemaless — правильный выбор для no-code), вычисленные значения formula/lookup **не** хранятся в `cells`, а в `computed` JSON (кэш, инвалидируется по `computed_rev`) |
| `HubBase` | + edge `members` → `HubBaseMember`; + `is_public bool` (read-only share-link, v2 фичи) |

### 2.2 Новые схемы

| Схема | Ключевые поля | Назначение |
|---|---|---|
| `HubBaseMember` | `base_id`, `user_id`, `role enum(owner/editor/commenter/viewer)`, unique(base,user) | Права per-base. Owner создаётся автоматически при создании базы |
| `HubView` | `table_id`, `name`, `view_type enum(grid/kanban/calendar/gallery)`, `config JSON` (kanban: group-колонка; calendar: date-колонка; gallery: cover), `filters JSON`, `sorts JSON`, `group_by JSON`, `column_order JSON`, `column_widths JSON`, `sort_order int` | Персистентные представления. Default grid-view создаётся с таблицей |
| `HubRelation` | `source_column_id`, `source_table_id`, `target_table_id`, `link_type enum(one_way/two_way)`, `symmetric_column_id nullable` | Метаданные link-колонок. Ссылки хранятся в cells как `[]recordUUID` |
| `HubRecordLink` | `relation_id`, `source_record_id`, `target_record_id`, unique-triple | Нормализованная таблица связей — для целостности, обратных lookup и каскадного отвязывания при удалении записи (JSON-массив в cells — только денормализованный кэш для чтения) |
| `HubFormulaDep` | `table_id`, `column_id`, `depends_on_column_id` | Рёбра графа зависимостей formula/lookup/rollup **на уровне колонок**. Materialized при сохранении формулы; используется для инкрементального пересчёта и обнаружения циклов |

**Решение:** cells остаются JSON (не EAV-таблица `cell`). Обоснование: масштаб 10–50 человек, таблицы до ~50k записей; EAV дал бы честные SQL-фильтры, но взорвал бы сложность миграций и кода. Фильтры/сортировки исполняются через PostgreSQL JSONB-операторы (`cells->>'colUUID'`) + GIN-индекс по `cells`; при деградации — уходим в вычисление на сервере постранично. Зафиксировать как ADR-BASES-01.

---

## 3. Архитектура

### 3.1 Formula engine (главное архитектурное решение)

**Собственный интерпретатор на Go. Без внешних eval-библиотек, без WASM/JS-рантайма.**

```
formula text ──lexer──► tokens ──Pratt/recursive-descent parser──► typed AST
                                        │ (валидация: имена колонок → UUID, типы, циклы)
                                        ▼
                              сохранение в HubColumn.config + рёбра в HubFormulaDep
запись изменилась ──► граф колонок (топологическая сортировка) ──► eval AST per record
                                        ▼
                              HubRecord.computed JSON (кэш вычисленных значений)
```

| Компонент | Решение | Обоснование |
|---|---|---|
| Парсер | Рекурсивный спуск + Pratt для приоритетов операторов, пакет `internal/bases/formula` (`lexer.go`, `parser.go`, `ast.go`) | ~400–600 строк, полный контроль, нет зависимостей. Прецедент в кодовой базе: JQL-парсер `search/jql.go` (348 стр) |
| Система типов | `number / string / bool / date / array` + `error`-значение (`#DIV/0!`, `#REF!` — как в таблицах: ошибка это значение, не panic) | Ошибки формул не роняют запрос; `panic()` запрещён tz_constraints |
| Функции | Whitelist-реестр `map[string]FuncDef` (арность, типы, реализация). MVP ≈ 30 функций: IF, AND/OR/NOT, SUM/AVG/MIN/MAX/COUNT, CONCAT/LEFT/RIGHT/LEN/TRIM/UPPER/LOWER, TODAY/NOW/DATEADD/DATEDIF/YEAR/MONTH/DAY, ROUND/ABS/MOD, ISBLANK, TEXT | Никакого произвольного кода: только AST-walk по whitelist. Расширение реестра — атомарные PR |
| Безопасность | Лимиты на этапе parse: длина формулы ≤ 2000 симв., глубина AST ≤ 30, узлов ≤ 500. На этапе eval: budget 10k операций/формула, `context.WithTimeout` 100ms/record, батч 1s | Отсутствие user-code + бюджеты = нет DoS-вектора |
| Зависимости | Граф **на уровне колонок** (не ячеек): formula-колонка зависит от входных колонок. Цикл → отказ при сохранении формулы (`400 formula_cycle`). Топологическая сортировка при пересчёте | Cell-level graph (как в Excel) — overkill: формулы Lark-типа row-scoped, зависимости одинаковы для всех строк колонки |
| Пересчёт | **Синхронно** row-scoped формулы изменённой записи (в той же транзакции UpdateRecord — пользователь сразу видит результат). **Асинхронно** через NATS JetStream (`hub.bases.recompute`) — каскады: смена определения формулы (пересчёт всей колонки батчами по 500), rollup при изменении связанных записей | Синхронная часть дешёвая (одна запись); тяжёлое — в фон, идемпотентный consumer, `computed_rev` защищает от гонок |
| Хранение результата | `HubRecord.computed` JSON-кэш; ответ API мержит cells+computed. Formula-колонки read-only при записи (сервер отбрасывает — расширение `normalizeCells`) | Чтение grid не вычисляет ничего — только отдаёт кэш |

### 3.2 Relations / lookup / rollup

- **Link-колонка**: `col_type=link`, `config={target_table_id, two_way}`; запись значения = список recordUUID → сервис синхронизирует `HubRecordLink` (source of truth) и кэш в cells обеих сторон при two-way.
- **Lookup**: `col_type=lookup`, `config={link_column_id, target_column_id}` — read-only, вычисляется тем же движком пересчёта (узел графа зависимостей, как формула c внешним входом).
- **Rollup**: `col_type=rollup`, `config={link_column_id, target_column_id, agg: sum|avg|min|max|count|countall}` — агрегат по связанным записям; пересчёт всегда асинхронный (NATS) при изменении target-записей.
- **Целостность**: soft-delete записи → отвязка links (deleted_at в HubRecordLink); удаление link-колонки → soft-delete relation + links; `#REF!` в lookup при потере target.
- Запрет: cross-**base** relations в MVP (только внутри одной base) — упрощает права и граф.

### 3.3 Views, фильтры, сортировки, группировки

- **Server-side исполнение** (не клиентская фильтрация): `GET /records` принимает `view_id` либо ad-hoc `filters/sorts` — иначе пагинация бессмысленна.
- Filter DSL (JSON, версионируемый): `{combinator: and|or, conditions: [{column_id, op, value}]}`, ops по типам: eq/neq/contains/empty/not_empty/gt/lt/gte/lte/is_before/is_after/is_any_of. Максимум 20 условий, 1 уровень вложенности combinator (MVP).
- Компилятор фильтра → Ent predicate поверх JSONB (`sql.P(...)` c `cells->>'uuid'` + касты по col_type). Фильтры по formula/lookup — по `computed`.
- Kanban/calendar/gallery — **те же records + группировка**: kanban = group-by select-колонка (server отдаёт группы с per-group limit/offset), calendar = записи в date-диапазоне месяца, gallery = grid c cover-полем. Один endpoint записей, `view_type` меняет только форму запроса и рендер.

### 3.4 Права (устранение security-gap)

Middleware-уровень сервиса: каждый метод получает `userID` и проверяет `HubBaseMember` role. Матрица: viewer=read, commenter=read (+комментарии в v2), editor=CRUD записей/представлений, owner=+схема (таблицы/колонки/члены/удаление базы). `ListBases` — только базы, где user член. 403 `forbidden`. Обязательная проверка в **service**-слое (не только handler) — паттерн NF-02/NF-03.

### 3.5 Real-time (v2 внутри roadmap)

NATS `hub.bases.record.{created,updated,deleted}` → ws-hub → комната `base:{id}` c membership-проверкой при subscribe (инвариант NF-02). Frontend: TanStack Query invalidate по WS-событию.

---

## 4. API-контракты (REST)

Ошибки: `{"error": "message", "request_id": "uuid"}`. Пагинация записей — **offset/limit** (страницы grid; инвариант tz: cursor только для чата). Rate limit: общий API 100/сек; import — files 10/мин.

### 4.1 Существующие (меняются)

| Метод | Путь | Изменение |
|---|---|---|
| GET | `/api/bases` | Только базы пользователя (member-scope) |
| GET | `/api/bases/:baseId/tables/:tableId` | Deprecated → заменяется `/records` + `/views`; временно сохраняется |

### 4.2 Новые endpoints (группами)

| Группа | Endpoints | Коды |
|---|---|---|
| Records | `GET /tables/:tableId/records?view_id=&offset=&limit=&filters=&sorts=&group_by=` (limit ≤ 200, default 50; ответ `{records, total, groups?}`) · `POST /records:batch` (создание/обновление до 100 за запрос — нужно импорту и grid-paste) | 200/400 (`invalid_filter`)/403/404 |
| Views | `GET/POST /tables/:tableId/views` · `PATCH/DELETE /views/:viewId` (config/filters/sorts/group_by/column_order/widths, version-lock) | 200/201/400/403/404/409 |
| Columns | `PATCH /columns/:columnId` (переименование, config; **смена col_type** — только совместимые пары, иначе `400 incompatible_type_change`) · `DELETE` (soft) | 200/400/403/409 |
| Formula | `POST /tables/:tableId/formula/validate` `{expression}` → `{valid, ast_summary, deps: [column_id], error?, error_pos?}` — dry-run для редактора · создание = `POST /columns` c `col_type=formula` (валидация повторяется, цикл → `400 formula_cycle`) | 200/400 |
| Relations | создание = `POST /columns` c `col_type=link` · `GET /records/:recordId/links/:columnId?offset&limit` (связанные записи для picker) | 200/201/400/403 |
| Members | `GET/POST /bases/:baseId/members` · `PATCH/DELETE /members/:userId` (последнего owner удалить нельзя → `400 last_owner`) | 200/201/400/403 |
| Import/Export | `POST /tables/:tableId/import` (CSV/XLSX через `imex`; presigned upload, маппинг колонок, dry-run режим) · `GET /tables/:tableId/export?format=csv&view_id=` (применяет фильтры view) | 200/202/400/413 |
| Automation API | `GET/POST` records через `apigateway` API-key scope `bases:read`/`bases:write` (переиспользование существующего механизма ключей) | 200/401/403 |

Итого ≈ 22 новых/изменённых endpoint. Все — Swagger-аннотации + `docs/wiki/API.md` + `system_specification.md` (docs-sync rule 55).

---

## 5. Frontend-декомпозиция

Из 1 файла (341 стр) → полноценный модуль по `30-frontend.mdc` (components/hooks/stores внутри `modules/bases/`):

```
frontend/src/modules/bases/
├── bases-page.tsx                 # композиция: sidebar + table tabs + view area (<150 стр)
├── components/
│   ├── base-sidebar.tsx           # список баз, создание, member-badge
│   ├── table-tabs.tsx             # табы таблиц + добавление
│   ├── view-switcher.tsx          # grid/kanban/calendar/gallery + управление views
│   ├── view-toolbar.tsx           # фильтры/сортировки/группировки (popover-билдеры)
│   ├── grid/
│   │   ├── grid-view.tsx          # TanStack Virtual: виртуализация строк (>50k записей)
│   │   ├── grid-header.tsx        # resize/reorder колонок, меню колонки
│   │   ├── grid-row.tsx           # memo-строка
│   │   └── cells/                 # cell-renderer per col_type (text/number/date/select/
│   │                              #   checkbox/user/link/formula/lookup/rating/attachment)
│   ├── kanban-view.tsx / calendar-view.tsx / gallery-view.tsx
│   ├── column-editor.tsx          # создание/настройка колонки (config per type)
│   ├── formula-editor.tsx         # textarea + автодополнение колонок/функций,
│   │                              #   подсветка ошибки по error_pos из /validate
│   ├── record-detail.tsx          # side-panel записи (все поля + links-picker)
│   ├── link-record-picker.tsx     # поиск/выбор записей target-таблицы
│   ├── member-manage-dialog.tsx   # sharing/права
│   └── import-dialog.tsx / export-menu.tsx
├── hooks/
│   ├── use-records.ts             # paginated query + optimistic cell update (rollback по 409)
│   ├── use-views.ts / use-columns.ts / use-base-members.ts
│   └── use-formula-validate.ts    # debounced dry-run
└── stores/
    └── bases-store.ts             # Zustand: selected base/table/view, edit-state ячейки
```

**Требования:** ErrorBoundary вокруг модуля; lazy-load kanban/calendar/gallery и formula-editor; optimistic updates для ячеек (главный UX-паттерн grid; rollback + toast при `409 version conflict`); skeleton loaders и empty states с CTA (долг из ЭПИК D); никаких `!important`, Shadcn — className-only; типы API фиксируются первым коммитом в `api/bases.ts` (контракт be↔fe).

**A11y (WCAG AA):** grid = `role="grid"` c roving tabindex, стрелки/Tab/Enter/Esc для навигации и редактирования ячеек; view-switcher = `role="tablist"`; фильтр-билдеры keyboard-доступны; focus-ring на активной ячейке; aria-live для результатов пересчёта formula; контраст токенов AA. Axe в e2e.

---

## 6. Roadmap: фазы → атомарные планы → волны

Оценки: S ≈ 1–2 дн · M ≈ 3–5 дн · L ≈ 1–2 нед. Каждый план ≤ 200–300 строк = атомарный коммит. Волны ≤ 4 агентов, разведены по файлам, контракт первым коммитом. Каждая волна: регрессия → `*-ver` → `/qa-gate` → docs-sync. Общий размер эпика — **XL** (соответствует B5 из EXPANSION-ANALYSIS).

### Фаза BS-0 — Аналитика и контракты (M)

| # | План | Исполнитель | Артефакт | Depends |
|---|---|---|---|---|
| 0-1 | US/AC: формулы, relations, views, sharing, import (персоны из `user_needs_analysis`) | ba1 | `vision_and_scope.md` §Bases | — |
| 0-2 | US/AC: права-матрица, лимиты, edge-cases (циклы, #REF!, конфликты) | ba2 | там же | — |
| 0-3 | ERD (§2) + API-контракты (§4) + filter DSL | sa1 | `system_specification.md` | 0-1,0-2 |
| 0-4 | ADR-BASES-01 (JSONB vs EAV) + ADR-BASES-02 (formula engine §3.1) + C4-обновление | arch | `architecture.md`, `.planning/decisions/` | 0-3 |
| 0-5 | DESIGN-SPEC: grid, view-switcher, formula-editor, состояния, a11y | designer | DESIGN-SPEC-bases.md | 0-1 |

**Волны:** W1: ba1 ∥ ba2 ∥ designer (черновик) → ba-ver → W2: sa1 ∥ arch ∥ designer (финал) → sa-ver/arch-ver/designer-ver.
**Success criteria:** все AC тестируемы; контракты покрывают §4 полностью; ADR приняты; designer-spec 8 разделов.

### Фаза BS-1 — Фундамент: права + пагинация + схемы (M)

| # | План | Исполнитель | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 1-1 | Ent: `HubBaseMember`, миграция owner-бэкфилл; расширение enum col_type + `config` | be2 | `ent/schema/*` | BS-0 | S |
| 1-2 | Права: member-scope во всех service-методах, ListBases-фильтр, 403 | be1 | `bases/service.go`, `repository.go` | 1-1 | M |
| 1-3 | `GET /records` offset/limit + batch endpoint (пока без фильтров) | be1 | `bases/handler.go`, `dto.go` | 1-2 | S |
| 1-4 | Members API + tests | be2 | `bases/members_*.go` | 1-2 | S |
| 1-5 | Contract-типы `api/bases.ts` (первый коммит) + members-dialog UI | fe1 | `api/bases.ts`, `member-manage-dialog.tsx` | 1-4 | S |
| 1-6 | Тесты: права-матрица (viewer/editor/owner × операции), пагинация | qa-auto1 | `backend/testframework/bases/` | 1-2..1-4 | S |

**Волны:** W1: 1-1 → W2: 1-2 ∥ 1-4 (be1/be2, разные файлы) → W3: 1-3 ∥ 1-5 ∥ 1-6 → be-ver/fe-ver/qa-auto-ver.
**Success:** чужая база недоступна (403, TC-BS-SEC-01..08 ↔ автотесты); grid грузит страницы ≤200; cov bases ≥82%.

### Фаза BS-2 — Views + фильтры/сортировки/группировки (L)

| # | План | Исполнитель | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 2-1 | Ent `HubView` + default-view бэкфилл + Views CRUD API | be2 | `ent/schema/hubview.go`, `bases/views_*.go` | BS-1 | M |
| 2-2 | Filter DSL компилятор → JSONB predicates + sorts + group_by | be1 | `bases/query/*.go` | 2-1 | M |
| 2-3 | FE-декомпозиция: разбор bases-page на sidebar/tabs/grid (без новых фич — рефактор) | fe2 | `modules/bases/components/*` | 1-5 | M |
| 2-4 | Grid v2: TanStack Virtual, cell-renderers per type, optimistic updates, a11y roving tabindex | fe1 | `grid/*` | 2-3 | L |
| 2-5 | View-switcher + view-toolbar (фильтр/сорт/группировка билдеры) | fe2 | `view-switcher.tsx`, `view-toolbar.tsx` | 2-1, 2-3 | M |
| 2-6 | Kanban/calendar/gallery views (lazy) | fe1 | `kanban-view.tsx` и др. | 2-4, 2-5 | M |
| 2-7 | Тесты: filter-компилятор unit (все ops × типы), views API; e2e grid+filter | qa-auto1 ∥ qa-auto2 | testframework, `stages/e2e/domains/bases.spec` | 2-2, 2-5 | M |

**Волны:** W1: 2-1 ∥ 2-3 → W2: 2-2 ∥ 2-4 ∥ 2-5 → W3: 2-6 ∥ 2-7 (qa×2) → верификаторы → приёмка.
**Success:** 4 типа view работают и персистентны; фильтры server-side (страница 50 записей из 10k — p95 < 300ms); axe без critical на grid; e2e bases.spec green.

### Фаза BS-3 — Relations / lookup / rollup (L)

| # | План | Исполнитель | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 3-1 | Ent `HubRelation` + `HubRecordLink`; link-колонка create/delete; синхронизация two-way | be2 | `ent/schema/*`, `bases/relations_*.go` | BS-1 | M |
| 3-2 | Lookup/rollup вычисление + NATS consumer пересчёта (общий с BS-4 механизм `computed`) | be1 | `bases/compute/*.go` | 3-1 | M |
| 3-3 | Links API (`/records/:id/links/:col`) + целостность при delete (#REF!) | be2 | `bases/relations_handler.go` | 3-1 | S |
| 3-4 | UI: link-cell + link-record-picker + record-detail side-panel | fe1 | `link-record-picker.tsx`, `record-detail.tsx` | 3-3 | M |
| 3-5 | UI: lookup/rollup config в column-editor | fe2 | `column-editor.tsx` | 3-2 | S |
| 3-6 | Тесты: two-way sync, cascade delete, rollup-агрегации, гонки пересчёта (`-race`) | qa-auto1 | testframework | 3-2, 3-3 | M |

**Волны:** W1: 3-1 → W2: 3-2 ∥ 3-3 → W3: 3-4 ∥ 3-5 ∥ 3-6 → верификаторы.
**Success:** связь A↔B двусторонне консистентна (TC ↔ автотест); rollup обновляется ≤ 2s после изменения target; удаление записи не оставляет висячих links; race-тесты green.

### Фаза BS-4 — Formula engine (L, самая рискованная)

| # | План | Исполнитель | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 4-1 | Lexer + parser + AST + валидация (типы, имена, лимиты) | be1 | `bases/formula/{lexer,parser,ast}.go` | BS-0 (ADR) | M |
| 4-2 | Реестр функций (30 MVP) + evaluator + error-значения | be1 | `bases/formula/{funcs,eval}.go` | 4-1 | M |
| 4-3 | `HubFormulaDep` + граф/циклы/topo-sort + синхронный пересчёт записи + async batch (NATS) | be2 | `ent/schema/hubformuladep.go`, `bases/compute/graph.go` | 4-1, 3-2 | M |
| 4-4 | `/formula/validate` + formula col_type wiring + read-only в normalizeCells | be2 | `bases/handler.go`, `service.go` | 4-2, 4-3 | S |
| 4-5 | Formula-editor UI (автодополнение, error_pos подсветка, справка функций) | fe1 | `formula-editor.tsx`, `use-formula-validate.ts` | 4-4 | M |
| 4-6 | Тесты: table-driven parser/eval (≥150 кейсов: приоритеты, типы, ошибки, unicode, лимиты); fuzz-тест парсера; каскад пересчёта | qa-auto1 ∥ qa-auto2 | testframework, e2e formula.spec | 4-2, 4-4 | M |

**Волны:** W1: 4-1 → W2: 4-2 ∥ 4-3 → W3: 4-4 ∥ 4-5 → W4: 4-6 (qa×2) ∥ security-auditor (DoS-бюджеты) → верификаторы.
**Success:** 30 функций по спеке; цикл отклоняется при сохранении; формула 10k записей пересчитывается фоном ≤ 10s; fuzz 60s без паники; eval-budget срабатывает (негативный тест); cov formula-пакета ≥ 85%.

### Фаза BS-5 — Import/export, automation API, real-time, polish (M)

| # | План | Исполнитель | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 5-1 | CSV/XLSX import через `imex` (маппинг, dry-run, batch insert) + export с фильтрами view | be1 | `bases/imex_*.go` | BS-2 | M |
| 5-2 | Automation API через `apigateway` scopes + rate limits | be2 | `apigateway/*`, `bases/` | BS-1 | S |
| 5-3 | WS real-time: NATS события + комната `base:{id}` c membership-check (NF-02) | be2 | `ws/`, `bases/events.go` | BS-1 | S |
| 5-4 | Import-dialog + export-menu + WS-invalidate hook + skeleton/empty states | fe2 | `import-dialog.tsx` и др. | 5-1, 5-3 | M |
| 5-5 | Ручные TC полного паритет-сценария + e2e-полировка + Lighthouse | qa-manual ∥ qa-auto2 | `docs/test_plan.md`, e2e | всё | M |

**Волны:** W1: 5-1 ∥ 5-2 ∥ 5-3 → W2: 5-4 ∥ 5-5 → верификаторы → code-reviewer → **team-lead финальная приёмка эпика**.
**Success:** импорт 5k строк CSV ≤ 30s c отчётом об ошибках; automation API работает по ключу с scope; второй пользователь видит правку ячейки ≤ 2s; полный `/qa-gate` green; docs-sync (API.md, system_specification, QUICKSTART, COVERAGE).

### Сводный план фаз

| Фаза | Содержание | Оценка | Depends | Ключевой риск |
|---|---|---|---|---|
| BS-0 | Аналитика, ADR, дизайн | M | — | недоспецификация формул |
| BS-1 | Права + пагинация + схемы | M | BS-0 | миграция-бэкфилл owner |
| BS-2 | Views + фильтры + FE grid v2 | L | BS-1 | perf JSONB-фильтров |
| BS-3 | Relations/lookup/rollup | L | BS-1 | консистентность two-way |
| BS-4 | Formula engine | L | BS-0, BS-3(compute) | сложность/безопасность |
| BS-5 | Import/export, API, WS, приёмка | M | BS-1..4 | scope creep |

Параллелизм между фазами: BS-2 (FE-тяжёлая) и BS-3 (BE-тяжёлая) могут идти внахлёст после BS-1 — be-агенты на BS-3, fe-агенты на BS-2 (балансировка по правилу §5.2 team-workflow). Итоговая длительность ≈ 5–7 недель календарно при 2be+2fe+2qa.

### Тестовая стратегия (сквозная)

| Уровень | Где | Что |
|---|---|---|
| Unit (Go) | `backend/testframework/bases/` | formula parser/eval (table-driven+fuzz), filter-компилятор, права-матрица, graph/циклы |
| Integration | там же, `-tags hubtest`, Docker PG | CRUD+RBAC, relations sync, пересчёт NATS, import |
| Vitest | `testframework/stages/unit/frontend-test/domains/bases/` | cell-renderers, hooks (optimistic+rollback), formula-editor автодополнение |
| E2E | `testframework/stages/e2e/domains/bases.spec.ts`, `bases-formula.spec.ts` | сценарий: база → колонки → записи → фильтр → kanban → формула → share; axe |
| Traceability | `docs/test_plan.md` | TC-BS-xx ↔ автотест (id в аннотации), правило testing-playbook |

---

## 7. Риски и митигации

| # | Риск | Вер./Влияние | Митигация |
|---|---|---|---|
| R1 | **Formula engine сложнее оценки** (типы, edge-cases дат/чисел, локали) | Выс./Выс. | Своя реализация по образцу `jql.go`; MVP-набор 30 функций жёстко зафиксирован в BS-0; table-driven+fuzz тесты с W1 фазы; error-значения вместо отказов; расширение функций — отдельные S-планы после паритета |
| R2 | Каскадный пересчёт → DoS/деградация (формула на 50k записей, глубокие цепочки) | Ср./Выс. | Column-level graph + топологический батч-пересчёт через NATS (не в запросе); бюджеты ops/timeout; лимит глубины цепочки зависимостей ≤ 10; метрики Prometheus на длительность пересчёта |
| R3 | JSONB-фильтры медленные на больших таблицах | Ср./Ср. | GIN-индекс; лимит 50k записей/таблица (см. §8 Q3); бенчмарк в BS-2 W2 как success-критерий; fallback-план — materialized колонки для сортируемых типов (ADR-дополнение, не блокирует MVP) |
| R4 | Two-way relations рассинхронизация (JSON-кэш vs HubRecordLink) | Ср./Выс. | HubRecordLink = source of truth, кэш перестраиваем; инвариант-тесты с `-race`; ремонтная job консистентности (admin endpoint) |
| R5 | Права добавляются на живой модуль — регрессия для существующих баз | Низ./Выс. | Миграция-бэкфилл: owner каждой базы = её создатель; e2e-smoke до/после; feature-parity тест старых endpoints |
| R6 | FE grid-производительность (виртуализация + optimistic + WS) | Ср./Ср. | TanStack Virtual (уже в стеке), memo-строки, нормализованный store; Lighthouse-замер в BS-5; profil-бюджет: 10k строк без jank |
| R7 | Scope creep до «полного Lark» (автоматизации-конструктор, история версий, комментарии к записям) | Выс./Ср. | Границы зафиксированы в BS-0 (product + product-ver); всё сверх §6 — backlog `gsd-add-backlog`; team-lead блокирует расползание волн |
| R8 | Конфликты параллельных агентов в `service.go`/`main.go` | Ср./Низ. | Разнесение по файлам (§6); wiring routes — один агент на фазу; новые подпакеты `formula/`, `compute/`, `query/` изолируют работу be1/be2 |

---

## 8. Открытые вопросы к продукту

| # | Вопрос | Влияние на план |
|---|---|---|
| Q1 | **Глубина паритета формул:** 30 функций MVP достаточно, или нужен Lark-полный набор (~100+, включая REGEX, массивные)? | +M/L на BS-4-расширение |
| Q2 | Нужны ли **автоматизации-конструктор** (триггер «запись изменилась» → действие) в этом эпике, или достаточно REST automation API? | Конструктор = отдельная фаза L (вне текущего scope) |
| Q3 | **Лимиты:** максимум записей/таблица (предлагаю 50k), таблиц/база (100), баз/инстанс, колонок/таблица (200), formula-колонок/таблица (30)? Нужен ли UI-индикатор квот? | Фиксируется в BS-0; влияет на R2/R3 |
| Q4 | **Производительность больших таблиц:** целевой p95 для grid 10k записей с 3 фильтрами — 300ms ок? Тестовые данные какого масштаба генерировать? | Success criteria BS-2 |
| Q5 | Attachment-поля (через SeaweedFS) — в этот эпик или backlog? В §6 не включены | +S/M в BS-3 либо backlog |
| Q6 | Share-по-ссылке (public read-only view) — нужен для паритета или только member-sharing? | +S в BS-5 либо backlog |
| Q7 | История изменений записи (audit trail) и комментарии к записям — паритет-требование? | +M, отдельная фаза |
| Q8 | Целевое покрытие: держим общий gate ≥82% и ≥85% для `formula/` (security-critical по DoS-поверхности) — подтвердить | COVERAGE.md цели |

---

## Саммари

Bases сегодня — скелет (855 стр CRUD-бэкенда, 1 FE-файл на 341 стр, 7 типов колонок, **нет прав — любой видит и правит все базы**). План закрывает паритет с Lark Base за 6 фаз (BS-0…BS-5, суммарно XL ≈ 5–7 недель): сначала security/пагинация/членство, затем views+фильтры server-side c JSONB-предикатами, relations/lookup/rollup через нормализованную таблицу связей, formula engine как собственный Pratt-парсер → typed AST → whitelist-интерпретатор с column-level графом зависимостей и NATS-пересчётом, финально import/export через `imex`, automation API через `apigateway` и real-time через ws/NATS. Frontend разбирается из монофайла в полный модуль (grid c виртуализацией, view-switcher, formula-editor, a11y `role="grid"`). Все волны ≤ 4 агентов из ростера, планы атомарные ≤ 300 строк, каждый слой закрыт верификатором и `/qa-gate`. Главный риск — formula engine — купирован прецедентом `jql.go`, жёстким MVP-набором из 30 функций, fuzz-тестами и eval-бюджетами. Перед стартом требуются ответы на Q1–Q3 (глубина формул, автоматизации, лимиты) — они меняют оценку BS-4/BS-5.
