# PLAN: Bases → паритет с Lark Base (v4.5)

> **Автор:** `team-lead` + `arch` (A/B-тест оркестрации, вариант Opus 4.8, max effort)
> **Дата:** 2026-07-02 · **Режим:** декомпозиция (read-only по коду, без изменений)
> **Эпик:** B5 из `.planning/july-2026/EXPANSION-ANALYSIS-2026-07-02.md` §5 (ЭПИК B, оценка XL)
> **Формат:** `.cursor/playbooks/team-workflow.md` §6 (PLAN, волны ≤4 агента, атомарные планы ≤200–300 строк)
> **Инварианты:** `.cursor/rules/60-tz-constraints.mdc` (UUID v4 · `version` optimistic lock · UTC в БД · soft-delete 30д · `{error, request_id}` · Fiber v3 · Ent+Atlas · TanStack Router · Shadcn/Tailwind 4)
>
> **Ссылки на требования (не дублируем):** vision — «No-code DB (Lark Base)» (`docs/vision_and_scope.md`),
> текущее ограничение «нет формул/relations/views» зафиксировано в EXPANSION-ANALYSIS §2 (bases) и §3.1.
> Стек/паттерны — `docs/architecture.md`, `.cursor/rules/20-backend.mdc`, `.cursor/rules/30-frontend.mdc`.

---

## 0. Резюме плана

Bases сегодня — **скелет** (backend 855 стр · FE 1 файл 341 стр): `Base→Table→Column→Record` CRUD с
inline-редактором и 7 типами колонок. Для паритета с Lark Base нужны 4 больших подсистемы:
**(1) formula engine** (главный риск), **(2) relations/lookup/rollup**, **(3) движок представлений
(grid/kanban/calendar/gallery) + фильтры/сортировки/группировки**, **(4) sharing/права + импорт/экспорт**.

Декомпозировано на **15 фаз (B5.0–B5.14)**, ~58 атомарных задач, 4 команды параллелизма. Оценка эпика —
**XL (~7–9 недель)** при 2 be + 2 fe + qa-потоке. Критический путь: контракты → Ent-модель → formula core →
интеграция формул/relations → views-движок → FE-редизайн grid → представления → sharing/imex → hardening.

---

## 1. Анализ текущего состояния (по коду, 2026-07-02)

### 1.1 Что есть

| Слой | Файл(ы) | Что реализовано |
|---|---|---|
| Ent-схемы | `ent/schema/hubbase.go`, `hubtable.go`, `hubcolumn.go`, `hubrecord.go` | 4 сущности. UUID PK, `version`, `created_at/updated_at` (UTC через `nowUTC`), `deleted_at` (soft-delete), индексы по `deleted_at`/`sort_order`. |
| Модель ячеек | `hubrecord.go` | `cells` — `field.JSON(map[string]interface{})`, ключ = **UUID колонки** (строкой), значение — сырое. |
| Типы колонок | `hubcolumn.go` + `dto.go` | enum `col_type`: `text, number, date, select, checkbox, url, user_ref`; `options` JSON (для select: `{choices:[]}`). |
| Service | `internal/bases/service.go` (341 стр) | CRUD base/table/column/record; `normalizeCells` (отсекает ячейки без известной колонки); optimistic lock по `version` для base/record; авто-создание таблицы `Main` + колонки `Name` при создании базы. |
| Repository | `internal/bases/repository.go` | `Repository` interface + `EntRepository`. Eager-load owner/tables/columns/records. Records сортируются по `updated_at desc`, **грузятся все** (нет пагинации). |
| Handler/routes | `internal/bases/handler.go` | `/api/bases` (+ вложенные tables/columns/records), `authMw`, `{error}`-ошибки через `mapServiceError` (404/400/409/500). |
| Frontend | `src/modules/bases/bases-page.tsx` (341 стр) | Один компонент: список баз, вкладки таблиц, HTML-`<table>` grid с inline-`<input>` (всё как текст), добавление колонки/записи, удаление записи. Локальный `useState`, TanStack Query. |
| Test-хуки | `internal/bases/export_hooks.go` (`//go:build hubtest`) | Экспорт `mapServiceError`/`normalizeCells` для тестов. |

### 1.2 Чего нет (gap до Lark Base)

| Область | Пробел |
|---|---|
| **Формулы** | Нет вообще: ни типа `formula`, ни парсера, ни вычислений, ни зависимостей/пересчёта. |
| **Relations** | Нет link-полей между таблицами, lookup, rollup, join-таблицы связей. |
| **Представления** | Нет сущности view: только один «grid» захардкожен в FE. Нет kanban/calendar/gallery/form. |
| **Фильтры/сорт/группы** | Нет: записи всегда сортируются `updated_at desc`, фильтров/группировок нет. |
| **Типы полей** | Узкий набор (7). Нет multi_select, attachment, currency/percent, rating, email/phone, duration, created/modified time/by, autonumber. |
| **Ячейки-редакторы** | Все значения в FE — текстовый `<input>` (даже checkbox/date/select не типизированы). |
| **Пагинация записей** | Нет (грузятся все) — не масштабируется. Нарушает дух TZ (cursor для больших списков). |
| **Права/sharing** | Только `owner` edge + `authMw`. Нет членства базы, ролей (editor/viewer), публичных ссылок. Любой авторизованный видит все базы (`ListBases` без user-scope!). |
| **Импорт/экспорт** | Нет CSV/xlsx (в bases). `export_hooks.go` — это тест-хуки, не фича. |
| **Автоматизации/API** | Нет вебхуков/публичного API/триггеров. |
| **Тесты/coverage** | bases 78.6% (2026-06-22 Docker baseline; < gate 82%, EXPANSION §4.3). |

### 1.3 Замечания для арх./безопасности (учесть в фазах)

- **SEC:** `ListBases`/`GetBase` **не фильтруют по владельцу/членству** → любой пользователь читает чужие базы.
  Закрыть при вводе sharing (B5.11) и раньше — минимальный owner-scope в B5.1.
- **PERF:** `GetTableView` грузит все columns+records одним запросом → обязательна пагинация (B5.14) до роста таблиц.
- **Модель ячеек** (`map[column_uuid]any`) удобна, но затрудняет server-side фильтрацию/индексацию (JSONB) — учтено в B5.7/B5.14.
- **Именование:** «Column» в коде = «Field» в Lark. Вводим сущность `HubField` как эволюцию `HubColumn`
  (аддитивно, без ломающего rename — старый enum сохраняем, расширяем).

---

## 2. Целевая модель данных (Ent-схемы)

> Все схемы: UUID v4 PK (`Immutable`), `version int Default(1)` (optimistic lock), `created_at/updated_at`
> через `nowUTC` (UTC), `deleted_at *time.Time` (soft-delete 30д), индекс по `deleted_at`.

### 2.1 Диаграмма сущностей

```
Organization (1 на инстанс, существует)
   └─ HubBase ──< HubBaseMember >── User        (sharing/RBAC)
        │            (role: owner|editor|commenter|viewer)
        └─< HubTable
               ├─< HubField        (типы + config; formula/link/lookup/rollup)
               ├─< HubRecord        (cells JSON + computed JSON)
               │     └─< HubRecordLink >── HubRecord   (relations, симметрично)
               ├─< HubView          (grid|kanban|calendar|gallery|form; filters/sorts/group)
               └─  HubFieldDependency (DAG зависимостей формул/lookup/rollup)
   HubBaseShare (public link: token, view_id, read-only)
```

### 2.2 Схемы

| Сущность | Ключевые поля (сверх UUID/version/timestamps/deleted_at) | Edges | Примечание |
|---|---|---|---|
| **HubBase** (расшир.) | `name`, `description`, `icon?`, `color?` | `owner→User`, `tables→HubTable`, `members→HubBaseMember`, `shares→HubBaseShare` | Добавить `organization` scope (1 org). |
| **HubTable** (расшир.) | `name`, `sort_order`, `primary_field_id?` (UUID) | `base←`, `fields→`, `records→`, `views→` | `primary_field` — отображаемое поле в link-пикере. |
| **HubField** (эволюция HubColumn) | `name`, `field_type` (enum, расширенный), `config JSON`, `formula TEXT?`, `is_computed bool`, `sort_order` | `table←`, `deps→HubFieldDependency` | `config` типизируется по `field_type` (см. 2.3). Старые записи `col_type` мигрируют в `field_type`. |
| **HubRecord** (расшир.) | `cells JSON` (ввод пользователя), `computed JSON` (кэш вычисляемых), `display_order float` | `table←`, `created_by→User`, `links→HubRecordLink` | `computed` хранит закэшированные значения formula/lookup/rollup. `display_order` — ручная сортировка (float для вставки между). |
| **HubRecordLink** (нов.) | `field_id UUID` (link-поле), `sort_order` | `from_record→HubRecord`, `to_record→HubRecord` | Join relations. Симметричная запись создаётся, если у link есть `symmetric_field_id`. Уникальный индекс `(field_id, from_record, to_record)`. |
| **HubView** (нов.) | `name`, `view_type` (grid/kanban/calendar/gallery/form), `config JSON`, `is_default bool`, `sort_order` | `table←` | `config`: `visible_fields[]`, `field_order[]`, `filters` (дерево), `sorts[]`, `group_by?`, + type-specific (kanban `stack_field`, calendar `date_field`, gallery `cover_field`, form `field_settings`). |
| **HubFieldDependency** (нов.) | — | `field→HubField`, `depends_on→HubField` | Ребро DAG. Строится при сохранении формулы/lookup/rollup; используется для порядка пересчёта и детекции циклов. |
| **HubBaseMember** (нов.) | `role` enum (`owner`,`editor`,`commenter`,`viewer`) | `base←`, `user→User` | RBAC базы. `founder`/admin org — суперправа. |
| **HubBaseShare** (нов.) | `token` (unique, случайный), `permission` (`view`), `expires_at?` | `base←`, `view→HubView?` | Публичная read-only ссылка (без логина). Валидация как в notification WebPush (ValidateOutboundURL не нужен — только чтение). |

### 2.3 `field_type` (целевой набор) и `config`

| field_type | Хранение в `cells` | `config` | Вычисляемое |
|---|---|---|---|
| text, long_text | string | — | нет |
| number, currency, percent, rating, duration | float64 | `{precision, symbol?, max?}` | нет |
| checkbox | bool | — | нет |
| date, datetime | RFC3339 UTC string | `{include_time, tz_display}` | нет |
| select | choice-id | `{choices:[{id,label,color}]}` | нет |
| multi_select | []choice-id | `{choices:[...]}` | нет |
| user_ref | []user-uuid | `{multiple}` | нет |
| url, email, phone | string | — | нет |
| attachment | []{file_id,name,mime} | — (SeaweedFS presigned) | нет |
| **formula** | — (в `computed`) | `{expr, result_type}` | **да** |
| **link** | — (через HubRecordLink) | `{target_table_id, relationship, symmetric_field_id?}` | связь |
| **lookup** | — (в `computed`) | `{link_field_id, target_field_id}` | **да** |
| **rollup** | — (в `computed`) | `{link_field_id, target_field_id, agg}` | **да** |
| created_time, modified_time | — | — | авто (из записи) |
| created_by, autonumber | — | `{start}` | авто |

---

## 3. Архитектура ключевых подсистем

### 3.1 Formula engine (главный компонент) — `internal/bases/formula/`

**Принцип: собственный интерпретируемый DSL, НЕ исполнение произвольного Go/JS.** Никакого `eval`, рефлексии,
I/O, сети или доступа к диску. Только чистая функция `(AST, RecordContext) → TypedValue`.

Пайплайн:

```
expr(string)
  → Lexer (токены: number, string, ident, op, paren, comma)
  → Parser (recursive-descent / Pratt для приоритетов) → AST
  → Validator (типы, арность функций, существование field-ссылок, глубина ≤ N, детекция циклов)
  → Compiler (AST → замкнутый список инструкций; кэш по хэшу expr)
  → Evaluator (интерпретирует над значениями полей записи; typed: Number|Text|Bool|Date|Null|Array)
```

**Whitelist функций** (детерминированные, UTC-aware):
- Математика: `+ - * / %`, `SUM MIN MAX ROUND ABS CEIL FLOOR MOD POWER SQRT`
- Логика: `IF AND OR NOT SWITCH ISBLANK ISERROR`, сравнения `= != > < >= <=`
- Строки: `CONCAT LEFT RIGHT MID LEN LOWER UPPER TRIM REPLACE SUBSTITUTE FIND`
- Дата: `TODAY NOW DATEADD DATETIME_DIFF YEAR MONTH DAY HOUR WEEKDAY DATESTR` (всё в UTC, дисплей-TZ на FE)
- Поля: `{Field Name}` или `field(<uuid>)` — ссылка на другое поле той же записи

**Безопасность / DoS-защита (обязательно):**
- Лимит глубины AST (напр. 32), длины выражения (напр. 4096 симв.), числа узлов (напр. 512).
- Бюджет шагов интерпретатора + `context.WithTimeout` на вычисление (напр. 100мс/запись).
- Ошибки типов/деления на ноль → typed `#ERROR` (значение, а не паника). **`panic()` запрещён** (tz).
- Никаких пользовательских циклов/рекурсии в языке (нет loops) → сложность ограничена AST.

**Зависимости и пересчёт (dependency DAG):**
- **Уровень поля:** формула ссылается на поля → рёбра `HubFieldDependency`. При сохранении формулы —
  топологическая сортировка + **детекция цикла** (цикл → `422 formula_cycle`).
- **Пересчёт при изменении определения поля:** пересчитать поле по всем записям таблицы в топопорядке downstream.
- **Пересчёт при изменении ячейки записи:** пересчитать только formula-поля записи, чьи зависимости включают
  изменённое поле (в топопорядке). Кэш в `HubRecord.computed`.
- **Cross-record (lookup/rollup/relation):** изменение записи A инвалидацию dependent-записей связанной таблицы
  публикует через **NATS JetStream** (`hub.bases.recompute`), consumer пересчитывает батчем (async, идемпотентно).
- **Стратегия хранения:** **compute-on-write + cache** в `computed`; при чтении отдаём кэш. Фолбэк
  recompute-on-read для «stale» (версия схемы поля изменилась). Полный rebuild — фоновая задача.

### 3.2 Relations / Lookup / Rollup

- **Link:** поле типа `link` + строки в `HubRecordLink`. Создание/удаление связи — отдельные эндпоинты
  (`POST/DELETE /records/:id/links`). Симметрия: если `config.symmetric_field_id` задан — зеркальная запись.
- **Lookup:** читает `target_field` из связанных записей (по `link_field`), результат — массив/скаляр в `computed`.
- **Rollup:** агрегирует `target_field` связанных (`SUM/COUNT/AVG/MIN/MAX/AND/OR/CONCAT`) → скаляр в `computed`.
- Целостность: удаление записи/поля/таблицы каскадит на `HubRecordLink` (soft-delete) и триггерит recompute.

### 3.3 Движок представлений и фильтров — `internal/bases/view/`

- **View config (JSON, версионируется):** `filters` (дерево `{conjunction: and|or, conditions:[{field_id, op, value}], groups:[...]}`),
  `sorts` ([`{field_id, dir}`]), `group_by` (`field_id?`), `visible_fields`/`field_order`, type-specific.
- **Операторы фильтра** по типам: `is/is_not/contains/not_contains/is_empty/is_not_empty/>/</>=/<=/is_within (date)/is_any_of (multi)`.
- **Исполнение (server-side):** этап 1 — интерпретатор фильтров на Go над загруженной страницей (просто, корректно
  для ≤ неск. тыс. строк); этап 2 (B5.14) — push-down в Postgres JSONB (`cells ->> 'uuid'`) + GIN-индекс для
  частых операторов и масштаба. Сортировка/группировка — на том же уровне.
- **Курсорная пагинация** записей (TZ: большие списки = cursor) с учётом активных sorts.

### 3.4 Соответствие layering (rule 20-backend)

`handler → service → repository(interface) → Ent`. `formula/` и `view/` — доменные пакеты внутри `bases`,
вызываются из `service.go`; **`*ent.Client` не протекает** в service/handler (только через Repository).

---

## 4. API-контракты (REST `/api/bases`)

> Общее: `authMw`; ошибки `{ "error": "...", "request_id": "uuid" }`; коды `400/401/403/404/409/422/429`.
> Rate limiting API 100/сек (tz). Записи — **cursor-пагинация** (`?cursor=&limit=`), метаданные offset для UI.

| Метод | Путь | Назначение | Особые коды |
|---|---|---|---|
| GET | `/bases` | список баз **пользователя** (owner+member) | — |
| POST | `/bases` | создать базу | 201 |
| GET/PUT/DELETE | `/bases/:baseId` | базовые CRUD + soft-delete | 404, 409(version) |
| POST | `/bases/:baseId/tables` | создать таблицу | 201 |
| GET | `/bases/:baseId/tables/:tableId` | метаданные таблицы + fields + views | 404 |
| POST/PATCH/DELETE | `.../fields` `.../fields/:fieldId` | CRUD поля (вкл. formula/link/lookup/rollup) | 422(formula/cycle) |
| POST | `.../fields/validate-formula` | проверка выражения без сохранения | 422 |
| GET | `.../records?viewId=&cursor=&limit=` | записи с применением view (filter/sort/group) | 400(bad filter) |
| POST | `.../records` / `.../records/batch` | создать (одну/пакет) | 201 |
| PATCH/DELETE | `.../records/:recordId` | обновить ячейки / soft-delete | 409(version) |
| POST/DELETE | `.../records/:recordId/links` | добавить/удалить связь | 404, 409 |
| GET/POST/PATCH/DELETE | `.../views` `.../views/:viewId` | CRUD представлений | 404 |
| GET/POST/PATCH/DELETE | `/bases/:baseId/members` | sharing/RBAC | 403(нет прав) |
| POST/DELETE | `/bases/:baseId/shares` | публичная read-only ссылка | 201 |
| GET | `/public/bases/shares/:token` | чтение по публичной ссылке (без auth) | 404/410(expired) |
| POST | `/bases/:baseId/import` | импорт CSV/xlsx (presigned upload → job) | 202 |
| GET | `.../tables/:tableId/export?format=csv|xlsx` | экспорт | — |
| POST/GET/DELETE | `/bases/:baseId/webhooks` (опц., B5.13) | автоматизации | — |

Новый код ошибки уровня формул: `422` с `error` из набора `formula_parse_error | formula_type_error |
formula_cycle | formula_unknown_field | formula_too_complex`.

---

## 5. Фронтенд-декомпозиция (из 1 файла → модуль)

Правило `.cursor/rules/30-frontend.mdc`: `modules/bases/{components,hooks,stores}` + `api/bases.ts`.
Lazy-load тяжёлых частей (grid-виртуализация, formula-редактор, calendar).

```
frontend/src/
├── api/bases.ts                       # расширить: fields, views, records(cursor), links, share, imex
├── modules/bases/
│   ├── bases-page.tsx                 # тонкий layout: sidebar + рабочая область (из 341 стр → ~80)
│   ├── stores/bases-store.ts          # Zustand: selectedBase/table/view, черновики фильтров, выделение ячеек
│   ├── hooks/
│   │   ├── useBases.ts  useTableMeta.ts  useRecords.ts(infinite cursor)
│   │   ├── useViews.ts  useFieldMutations.ts  useRecordMutations.ts
│   │   ├── useFormulaValidation.ts  useCellSelection.ts (keyboard grid nav)
│   └── components/
│       ├── BaseSidebar.tsx  TableTabs.tsx
│       ├── ViewSwitcher.tsx           # grid/kanban/calendar/gallery/form
│       ├── grid/GridView.tsx          # виртуализированная сетка (aria-grid)
│       ├── grid/GridCell.tsx + editors/{Text,Number,Date,Select,MultiSelect,Checkbox,User,Link,Attachment}Cell.tsx
│       ├── grid/GridHeader.tsx  FieldConfigPanel.tsx
│       ├── formula/FormulaEditor.tsx  # автокомплит функций/полей + inline-валидация 422
│       ├── views/KanbanView.tsx  CalendarView.tsx  GalleryView.tsx  FormView.tsx
│       ├── toolbar/FilterBar.tsx  SortMenu.tsx  GroupMenu.tsx  FieldVisibilityMenu.tsx
│       ├── record/RecordDetailDrawer.tsx  LinkPicker.tsx
│       └── share/ShareDialog.tsx  ImportDialog.tsx
```

**a11y (обяз., проверяет fe-ver/designer-ver):** grid — `role="grid"`, roving `tabindex`, навигация стрелками,
`Enter` вход в редактирование / `Esc` отмена, `Tab` между ячейками; focus-trap в диалогах/drawer; formula
autocomplete — `role="listbox"` + `aria-activedescendant`; контраст AA; kanban drag доступен с клавиатуры;
skeleton/empty/error-состояния (сейчас «Загрузка...»). Токены `--hub-*`, Shadcn className-only.

---

## 6. Поэтапный roadmap (фазы → атомарные планы, волны ≤4)

> Оценки: **S** 1–2д · **M** 3–5д · **L** 1–2нед · **XL** >2нед. Каждый план — атомарный коммит (≤200–300 стр).
> Исполнители из ростера (§1.2 team-workflow). Каждый артефакт → персональный `*-ver` → `verified:true` при green.

### Фаза B5.0 — Discovery, контракты, дизайн-спека (M)

**Goal:** зафиксировать US/AC, FR/NFR+API+ERD, арх-решение formula/relations/views, DESIGN-SPEC — единый контракт.
**Depends on:** — (стартовая). **Success:** ba-ver/sa-ver/arch-ver/designer-ver verified; product-ver scope-lock.

| # | Задача | Исполн. | Файлы (docs) | Depends | Оценка |
|---|---|---|---|---|---|
| 0.1 | US/AC паритета (грид, формулы, relations, views, sharing, imex), MoSCoW | ba1 | `docs/vision_and_scope.md` | — | S |
| 0.2 | Персоны/edge-cases (большие таблицы, права, publiс-ссылка) | ba2 | vision_and_scope | — | S |
| 0.3 | FR/NFR + API-контракты (§4) + ERD (§2) | sa1 | `docs/system_specification.md`, `wiki/API.md` | 0.1 | M |
| 0.4 | Матрица типов полей ↔ операторы фильтра ↔ хранение | sa2 | system_specification | 0.1 | S |
| 0.5 | Арх-решение formula engine (DSL/AST/sandbox/DAG), relations, views | arch | `docs/architecture.md` (§bases) | 0.3 | M |
| 0.6 | DESIGN-SPEC: grid, view-switcher, formula-editor, состояния, a11y | designer | `DESIGN-SPEC-bases.md` | 0.3 | M |

**Волны:** W1: 0.1∥0.2 → W2: 0.3∥0.4 → W3: 0.5∥0.6.
**Тесты:** нет кода (спека). **Риски:** глубина паритета — вопросы §8 закрыть до старта.

---

### Фаза B5.1 — Data-model foundation (Ent + миграция + owner-scope) (M)

**Goal:** новые/расширенные Ent-схемы (§2) + Atlas-миграция + фикс owner-scope в list/get.
**Depends on:** B5.0 (ERD). **Success:** `go build` green; миграция применяется на чистой БД; список баз — только свои.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 1.1 | Эволюция `HubColumn→HubField` (новый enum field_type + config/formula/is_computed) | be2 | `ent/schema/hubcolumn.go`(→field), dto | 0.5 | M |
| 1.2 | Новые схемы: HubView, HubRecordLink, HubFieldDependency | be2 | `ent/schema/hub{view,recordlink,fielddependency}.go` | 0.5 | M |
| 1.3 | Sharing-схемы: HubBaseMember, HubBaseShare + edges к HubBase | be1 | `ent/schema/hub{basemember,baseshare}.go`, hubbase.go | 0.5 | S |
| 1.4 | `HubRecord`: `computed JSON`, `display_order`; `HubTable.primary_field_id` | be1 | hubrecord.go, hubtable.go | 0.5 | S |
| 1.5 | Atlas-миграция + backfill (col_type→field_type, default grid-view на таблицу) | be2 | `ent/migrate/*` | 1.1–1.4 | M |
| 1.6 | Owner/member-scope в `ListBases/GetBase` (SEC-фикс) | be1 | `bases/repository.go`, `service.go` | 1.3 | S |

**Волны:** W1: 1.1∥1.3∥1.4 → W2: 1.2 → W3: 1.5 → W4: 1.6.
**Тесты (qa-auto1):** миграция up/down, backfill корректность; repository owner-scope (unit+integration).
**Риски:** R-мигр — необратимый backfill → down-миграция + бэкап (Прил.F team-workflow); симметрия rename без потери данных.

---

### Фаза B5.2 — Расширение типов полей + типизированная валидация ячеек (M)

**Goal:** серверная типизация значений по `field_type` (multi_select, currency, rating, attachment, email/phone,
created/modified time/by, autonumber) + строгая валидация в `service`.
**Depends on:** B5.1. **Success:** невалидная ячейка → `422`; авто-поля (created_time и т.д.) считаются сервером.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 2.1 | Валидаторы значений по типу + `config`-валидация | be1 | `bases/fieldtype/validate.go` | 1.1 | M |
| 2.2 | Авто-поля (created/modified time/by, autonumber) в service | be2 | `bases/service.go` | 1.1 | S |
| 2.3 | Attachment через SeaweedFS presigned (aws-sdk-go-v2) | be2 | `bases/attachment.go` | storage-модуль | M |
| 2.4 | Расширить DTO/normalizeCells под новые типы | be1 | `bases/dto.go`, service.go | 2.1 | S |

**Волны:** W1: 2.1∥2.2 → W2: 2.3∥2.4.
**Тесты:** unit валидации каждого типа (happy+invalid+edge); attachment presigned mock. **Риск:** 25MB limit (tz NF-09).

---

### Фаза B5.3 — Formula engine core (parser/AST/eval/sandbox) (L) ★критический

**Goal:** самостоятельный пакет `formula/`: лексер→парсер→AST→валидатор→интерпретатор с whitelist-функциями и
DoS-лимитами. **Чистая функция, без БД.** **Depends on:** B5.0 (арх-решение). **Success:** ≥40 функций,
детерминизм, лимиты глубины/шагов/таймаута; `#ERROR` вместо паники; ≥90% coverage пакета.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 3.1 | Lexer (токены) + типы `TypedValue` | be1 | `bases/formula/lexer.go`, `value.go` | 0.5 | S |
| 3.2 | Parser (recursive-descent/Pratt) → AST + приоритеты | be1 | `bases/formula/parser.go`, `ast.go` | 3.1 | M |
| 3.3 | Библиотека функций (math/logic/string/date, UTC) | be2 | `bases/formula/functions.go` | 3.1 | M |
| 3.4 | Validator (арность, типы, глубина, field-refs) | be2 | `bases/formula/validate.go` | 3.2,3.3 | M |
| 3.5 | Evaluator + sandbox (step-budget, timeout, `#ERROR`) | be1 | `bases/formula/eval.go` | 3.2,3.3 | M |
| 3.6 | Компиляция+кэш по хэшу выражения | be2 | `bases/formula/compile.go` | 3.5 | S |

**Волны:** W1: 3.1 → W2: 3.2∥3.3 → W3: 3.4∥3.5 → W4: 3.6.
**Тесты (qa-auto1):** табличные тесты по каждой функции; property/fuzz на парсер (не паникует); лимиты (глубокая
вложенность, длинная строка → ошибка, не OOM); детерминизм NOW/TODAY (инъекция «текущего времени»).
**Риски:** R-formula (главный, см. §7) — sandbox и fuzz-тесты обязательны до интеграции.

---

### Фаза B5.4 — Интеграция формул: DAG зависимостей, пересчёт, кэш (L) ★критический

**Goal:** formula-поля в таблице: сохранение выражения строит DAG (детект цикла→422), пересчёт кэшируется в
`computed`, инкрементальный пересчёт при изменении поля/ячейки. **Depends on:** B5.1, B5.3.
**Success:** изменение ячейки → зависимые формулы пересчитаны; цикл отклоняется; `validate-formula` работает.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 4.1 | Построение DAG из AST + запись HubFieldDependency + детект цикла | be1 | `bases/formula/deps.go`, service.go | 3.2 | M |
| 4.2 | Пересчёт поля по всем записям (топопорядок) + запись `computed` | be2 | `bases/recompute.go` | 4.1 | M |
| 4.3 | Инкрементальный пересчёт при PATCH записи | be1 | `bases/service.go` | 4.2 | M |
| 4.4 | Эндпоинт `validate-formula` + маппинг 422-кодов | be2 | `bases/handler.go` | 4.1 | S |
| 4.5 | NATS-событие `hub.bases.recompute` (заготовка под cross-record) | be1 | `bases/events.go` | 4.2 | S |

**Волны:** W1: 4.1 → W2: 4.2∥4.4 → W3: 4.3∥4.5.
**Тесты:** цепочка A→B→C пересчёт; цикл→422; batch-пересчёт при смене формулы; конкурентный PATCH (version).
**Риск:** производительность пересчёта на больших таблицах → батч + async (B5.14 добьёт), лимит fan-out.

---

### Фаза B5.5 — Relations / Link-поля + join (L)

**Goal:** link-поле, HubRecordLink, симметрия, целостность при удалении. **Depends on:** B5.1.
**Success:** связать/отвязать записи между таблицами; симметричное поле; каскад soft-delete.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 5.1 | Link-поле config + репозиторий HubRecordLink | be2 | `bases/repository.go` | 1.2 | M |
| 5.2 | Service: link/unlink + симметрия + валидация target-таблицы | be1 | `bases/link_service.go` | 5.1 | M |
| 5.3 | Эндпоинты `/records/:id/links` (POST/DELETE) | be1 | `bases/handler.go` | 5.2 | S |
| 5.4 | Каскад целостности (удаление записи/поля/таблицы) + recompute-триггер | be2 | `bases/service.go` | 5.2,4.5 | M |

**Волны:** W1: 5.1 → W2: 5.2 → W3: 5.3∥5.4.
**Тесты:** one-to-many/many-to-many; симметрия; удаление → отвязка; RBAC (нельзя линковать в чужую базу).

---

### Фаза B5.6 — Lookup & Rollup (M)

**Goal:** lookup (значение из связанных) и rollup (агрегация) как computed-поля с cross-record пересчётом.
**Depends on:** B5.4, B5.5. **Success:** rollup SUM/COUNT/AVG…; lookup массив/скаляр; пересчёт при изменении связанной записи.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 6.1 | Lookup-резолвер (через link → target_field) | be1 | `bases/lookup.go` | 5.2,4.2 | M |
| 6.2 | Rollup-агрегаторы (SUM/COUNT/AVG/MIN/MAX/AND/OR/CONCAT) | be2 | `bases/rollup.go` | 5.2,4.2 | M |
| 6.3 | Cross-record consumer NATS `hub.bases.recompute` (идемпотентный) | be1 | `bases/recompute_consumer.go` | 4.5,6.1,6.2 | M |

**Волны:** W1: 6.1∥6.2 → W2: 6.3.
**Тесты:** rollup по 100+ связям; изменение связанной записи → пересчёт rollup/lookup; идемпотентность consumer.
**Риск:** каскад пересчёта (relation→rollup→formula) → ограничить глубину/циклы cross-table.

---

### Фаза B5.7 — Views-модель + фильтры/сортировки/группировки (server) (L)

**Goal:** CRUD представлений + серверная интерпретация filters/sorts/group_by + cursor-пагинация записей по view.
**Depends on:** B5.1. **Success:** `GET records?viewId=` возвращает отфильтрованные/сортированные/сгруппированные,
курсорная пагинация; операторы по всем типам.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 7.1 | HubView CRUD (service/handler/repo) + default grid-view | be1 | `bases/view_service.go`, handler.go | 1.2 | M |
| 7.2 | Интерпретатор фильтров (дерево and/or, операторы по типам) | be2 | `bases/view/filter.go` | 1.1 | M |
| 7.3 | Сортировка + группировка + видимость/порядок полей | be2 | `bases/view/sort_group.go` | 7.2 | M |
| 7.4 | Cursor-пагинация записей с учётом sorts | be1 | `bases/repository.go` | 7.3 | M |

**Волны:** W1: 7.1∥7.2 → W2: 7.3 → W3: 7.4.
**Тесты:** каждый оператор фильтра; вложенные and/or; сортировка стабильна; курсор без пропусков/дублей.
**Риск:** JSONB-фильтрация в Go не масштабируется → индексный push-down отложен в B5.14 (open Q §8).

---

### Фаза B5.8 — FE: модульная декомпозиция + grid-редактор (виртуализация, a11y) (L)

**Goal:** разложить `bases-page.tsx` на модуль (§5); виртуализированный grid с типизированными редакторами ячеек и
keyboard-навигацией. **Depends on:** B5.2, B5.7 (API), B5.0 (DESIGN-SPEC). **Success:** grid рендерит 5k строк
плавно; редакторы по типам; a11y-навигация; регрессия FE green.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 8.1 | `api/bases.ts` расширение (fields/views/records-cursor/links) + типы | fe1 | `src/api/bases.ts` | 0.3 | S |
| 8.2 | Каркас модуля: bases-page(layout), store, useBases/useTableMeta/useRecords | fe1 | `modules/bases/{bases-page,stores,hooks}` | 8.1 | M |
| 8.3 | GridView виртуализированный + GridHeader + aria-grid + keyboard nav | fe2 | `modules/bases/components/grid/*` | 8.1 | L |
| 8.4 | Типизированные редакторы ячеек (Text/Number/Date/Select/MultiSelect/Checkbox/User) | fe1 | `.../grid/editors/*` | 8.2 | M |
| 8.5 | FieldConfigPanel (создание/редакт. поля, выбор типа) | fe2 | `.../FieldConfigPanel.tsx` | 8.3 | M |

**Волны:** W1: 8.1 → W2: 8.2∥8.3 → W3: 8.4∥8.5.
**Тесты (qa-auto2):** Vitest на редакторы+store; keyboard-nav; axe без critical.
**Риск:** конфликт fe1/fe2 → 8.1 (api/types) первым коммитом, дальше разные файлы.

---

### Фаза B5.9 — FE: View-switcher + Kanban/Calendar/Gallery + toolbar фильтров (L)

**Goal:** переключение представлений и non-grid виды + панель фильтров/сортировок/группировок.
**Depends on:** B5.7, B5.8. **Success:** kanban drag (мышь+клавиатура), calendar по date-полю, gallery cover;
FilterBar строит дерево фильтров; lazy-load calendar.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 9.1 | ViewSwitcher + useViews + сохранение view-config | fe1 | `.../ViewSwitcher.tsx`, `hooks/useViews.ts` | 8.2 | M |
| 9.2 | KanbanView (stack по select) + drag a11y | fe2 | `.../views/KanbanView.tsx` | 9.1 | M |
| 9.3 | CalendarView (lazy) + GalleryView | fe1 | `.../views/{Calendar,Gallery}View.tsx` | 9.1 | M |
| 9.4 | Toolbar: FilterBar/SortMenu/GroupMenu/FieldVisibilityMenu | fe2 | `.../toolbar/*` | 9.1 | L |

**Волны:** W1: 9.1 → W2: 9.2∥9.3 → W3: 9.4.
**Тесты:** view-switch сохраняет config; фильтр меняет выборку; kanban drag клавиатурой; empty/error-состояния.

---

### Фаза B5.10 — FE: Formula-редактор + Link-пикер + RecordDetail (M)

**Goal:** редактор формул с автокомплитом функций/полей и inline-валидацией (422); пикер связей; drawer записи.
**Depends on:** B5.4, B5.6 (API validate/links). **Success:** ошибка формулы показывается до сохранения; link-пикер
ищет по primary_field; drawer редактирует все типы.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 10.1 | FormulaEditor + useFormulaValidation (debounced validate) | fe1 | `.../formula/FormulaEditor.tsx` | 4.4 | M |
| 10.2 | Автокомплит функций/полей (listbox a11y) | fe1 | `.../formula/autocomplete.ts` | 10.1 | M |
| 10.2b| LinkPicker + LinkCell + lookup/rollup read-only отображение | fe2 | `.../record/LinkPicker.tsx`, editors/LinkCell | 5.3,6.1 | M |
| 10.3 | RecordDetailDrawer (все типы, вкл. computed read-only) | fe2 | `.../record/RecordDetailDrawer.tsx` | 8.4 | M |

**Волны:** W1: 10.1∥10.2b → W2: 10.2∥10.3.
**Тесты:** валидная/невалидная формула UI; автокомплит клавиатурой; link CRUD из UI; computed нельзя редактировать.

---

### Фаза B5.11 — Sharing & права (members + публичная ссылка) (M)

**Goal:** роли базы (owner/editor/commenter/viewer) + RBAC во всех эндпоинтах + публичная read-only ссылка.
**Depends on:** B5.1 (schemas). **Success:** viewer не редактирует; публичная ссылка читается без логина;
403 при нарушении прав. **Security-auditor подключён.**

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 11.1 | Members CRUD + проверка прав (middleware/guard в service) | be1 | `bases/member_service.go`, handler.go | 1.3 | M |
| 11.2 | RBAC-инъекция во все bases-эндпоинты (роль→операция) | be2 | `bases/rbac.go`, service.go | 11.1 | M |
| 11.3 | Публичная share-ссылка (token, read-only, expiry, 410) | be1 | `bases/share_service.go`, handler.go | 1.3 | M |
| 11.4 | FE ShareDialog + управление доступом | fe1 | `.../share/ShareDialog.tsx` | 11.1 | M |

**Волны:** W1: 11.1 → W2: 11.2∥11.3 → W3: 11.4.
**Тесты:** матрица роль×операция; публичная ссылка (read-only, expired→410); security-auditor review.
**Риск:** SEC — закрыть исходный дефект «список чужих баз»; constant-time token compare.

---

### Фаза B5.12 — Импорт/Экспорт (CSV/xlsx) (M)

**Goal:** импорт CSV/xlsx (маппинг колонок→поля, presigned upload → async job) и экспорт таблицы.
**Depends on:** B5.2. **Success:** импорт 10k строк без OOM (`io.LimitReader`, стрим); экспорт с типами.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 12.1 | Экспорт CSV/xlsx (стрим) | be2 | `bases/export.go`, handler.go | 2.1 | M |
| 12.2 | Импорт: presigned upload + парсер + маппинг + job | be1 | `bases/import_service.go` | 2.1 | L |
| 12.3 | FE ImportDialog (маппинг колонок) + Export-кнопка | fe2 | `.../share/ImportDialog.tsx` | 12.1,12.2 | M |

**Волны:** W1: 12.1∥12.2 → W2: 12.3.
**Тесты:** round-trip export→import; битый CSV→422; лимит размера (25MB, tz); типизация при импорте.

---

### Фаза B5.13 — Automation API (опционально, по решению продукта) (M)

**Goal:** публичный API/вебхуки для no-code автоматизаций (record created/updated → webhook; API-токены).
**Depends on:** B5.11. **Success:** вебхук с подписью (нельзя skip — SEC-02); API-токен через apigateway.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 13.1 | Webhook-подписки + доставка (safehttp, ValidateOutboundURL, подпись) | be1 | `bases/webhook.go` | 11.2 | M |
| 13.2 | Публичный record API через apigateway API-key | be2 | `bases/public_api.go` | 11.2 | M |
| 13.3 | FE управление вебхуками/токенами | fe1 | `.../automation/*` | 13.1 | S |

**Волны:** W1: 13.1∥13.2 → W2: 13.3. **Риск:** SSRF (safehttp обяз.), подпись webhook не пропускать (tz FORBIDDEN).
**Примечание:** фаза — за флагом «нужно ли API» (open Q §8.5); может уйти в отдельный milestone.

---

### Фаза B5.14 — Hardening: perf, coverage ≥82%, e2e, docs (M)

**Goal:** масштаб (JSONB-индексы/push-down, батч-пересчёт async), coverage bases ≥82%, e2e-паритет-сценарии, docs-sync.
**Depends on:** B5.4–B5.12. **Success:** таблица 50k строк — фильтр/сорт < целевого; coverage ≥82%; e2e green; docs синхр.

| # | Задача | Исполн. | Файлы | Depends | Оценка |
|---|---|---|---|---|---|
| 14.1 | JSONB GIN-индекс + push-down частых фильтров в Postgres | be2 | `bases/repository.go`, migrate | 7.2 | M |
| 14.2 | Async-батч пересчёта formula/rollup (очередь NATS) | be1 | `bases/recompute_consumer.go` | 4.2,6.3 | M |
| 14.3 | Coverage bases ≥82% (unit+integration в `testframework/bases/`) | qa-auto1 | `backend/testframework/bases/*` | все | M |
| 14.4 | E2E Playwright: создать базу→формула→relation→view→share | qa-auto2 | `stages/e2e/domains/bases.spec` | FE-фазы | M |
| 14.5 | Ручные TC (traceability AC↔TC) + `test_plan.md` | qa-manual | `docs/test_plan.md` | все | M |
| 14.6 | Docs-sync: system_specification, wiki/API, COVERAGE, STATE | docs-writer | `docs/*`, `.planning/*` | все | S |

**Волны:** W1: 14.1∥14.2 → W2: 14.3∥14.4 → W3: 14.5∥14.6.
**Тесты:** нагрузочный (50k строк); coverage-gate; e2e-паритет; регрессия full `/qa-gate`.

---

### 6.99 Сводка волн верхнего уровня (порядок фаз)

```
B5.0 (контракты) → B5.1 (модель) → ┬─ B5.2 (типы полей)
                                    ├─ B5.3 (formula core) ─→ B5.4 (formula integ) ─┐
                                    └─ B5.5 (relations) ──────────────────────────── ┼─→ B5.6 (lookup/rollup)
B5.7 (views/filters, || после B5.1)                                                  │
FE: B5.8 (grid) → B5.9 (views UI) → B5.10 (formula/link UI)  (после соответств. API) │
B5.11 (sharing) ∥ B5.12 (imex) → [B5.13 automation, опц.] → B5.14 (hardening/gate) ──┘
```

Каждая фаза внутри — волны ≤4 агента (be1/be2/fe1/fe2/qa-auto1/qa-auto2/qa-manual/designer) + верификаторы
(`be-ver`/`fe-ver`/`qa-*-ver`); `code-reviewer` pre-merge; финальная приёмка — `team-lead`.

---

## 7. Риски и митигации

| # | Риск | Сев. | Митигация |
|---|---|---|---|
| R1 | **Formula engine — небезопасное вычисление** (произвольный код/DoS/паника) | ★CRIT | Собственный DSL без `eval`; whitelist-функции; лимиты глубины/шагов/таймаута; `#ERROR` вместо panic; fuzz-тесты парсера; изоляция пакета `formula/` (no I/O). |
| R2 | Циклические зависимости формул/rollup | HIGH | DAG + топосортировка; детект цикла при сохранении → `422 formula_cycle`; ограничение глубины cross-table. |
| R3 | Производительность на больших таблицах (фильтр/сорт/пересчёт) | HIGH | Cursor-пагинация; JSONB GIN + push-down (B5.14); async-батч пересчёта через NATS; лимит fan-out. |
| R4 | **Утечка чужих баз** (текущий `ListBases` без scope) | HIGH | Owner-scope в B5.1.6; полноценный RBAC B5.11; security-auditor review. |
| R5 | Модель `cells` (JSON по UUID) плохо индексируется | MED | Гибрид: computed-кэш + JSONB-индексы; денормализация горячих полей при необходимости. |
| R6 | Миграция HubColumn→HubField (потеря данных) | MED | Аддитивная эволюция enum; down-миграция; бэкап; backfill-тесты (Прил.F). |
| R7 | Конфликты параллельных FE/BE агентов | MED | Контракт/типы первым коммитом волны; разведение по файлам (§4 team-workflow). |
| R8 | Cross-record recompute-штормы (relation→rollup→formula) | MED | Идемпотентный NATS-consumer; батчинг; лимит глубины каскада; дедуп по record+field. |
| R9 | Attachment/import DoS (память) | MED | `io.LimitReader`, presigned PUT, 25MB BodyLimit (tz), стрим-парсинг. |
| R10 | Scope-разрастание (паритет «бесконечен») | MED | product-lock в B5.0; B5.13 за флагом; фичи сверх MoSCoW → backlog, не в волну. |

---

## 8. Открытые вопросы к продукту

| # | Вопрос | Влияние |
|---|---|---|
| 8.1 | **Глубина паритета:** полный Lark (formula+relation+lookup+rollup+4 вида+automation) или «достаточно для 10–50 чел» (grid+формулы+relation+фильтры)? | Определяет, входят ли B5.6/B5.9(calendar/gallery)/B5.13 в scope сейчас. |
| 8.2 | **Лимиты:** max строк/таблица, max полей, max формула-глубина, размер attachment? | Настройка sandbox-лимитов (B5.3) и стратегии индексации (B5.14). |
| 8.3 | **Производительность больших таблиц:** целевой размер (10k? 100k?) и SLA на фильтр/пересчёт? | Решает нужность JSONB push-down/материализации сразу vs позже. |
| 8.4 | **Sharing-модель:** только внутри org (члены) или публичные ссылки наружу? Комментарии (commenter-роль) нужны? | Объём B5.11 и security-периметр. |
| 8.5 | **Automation API (B5.13):** нужен ли публичный API/вебхуки сейчас, или это отдельный milestone? | Включение/исключение целой фазы. |
| 8.6 | **Realtime co-editing:** нужна ли совместная правка (WS-присутствие/live-updates) как в Lark, или optimistic + invalidate достаточно? | Потенциальная доп-фаза (WS), сейчас не заложена. |
| 8.7 | **Целевое покрытие:** держим ≥82% (как ядро) для новых фаз или ≥50% на первой итерации (как v4)? | Объём qa-фаз (B5.14). |
| 8.8 | **Параллелизм команды:** сколько потоков (2 be + 2 fe?) — от этого зависит длительность (7–9 нед при 4). | Тайминг roadmap. |

---

## Verification

Каждая фаза: исполнитель → регрессия слоя (green, лог) → docs-sync → персональный `*-ver` (`verified:true`) →
`/qa-gate` green → `code-reviewer` → **финальная приёмка `team-lead`**. Formula-фазы (B5.3/B5.4) и sharing (B5.11)
дополнительно проходят `security-auditor`. Ни один артефакт не входит в baseline без `verified:true` + green gate.

---

## Саммари

1. **Состояние:** bases — скелет (4 Ent-схемы, CRUD, 1 FE-файл 341 стр); нет формул, relations, views,
   фильтров, sharing-RBAC, пагинации; обнаружен SEC-дефект — `ListBases` отдаёт чужие базы.
2. **Модель:** эволюция `HubColumn→HubField` (+ formula/link/lookup/rollup и ~15 типов) и 5 новых сущностей
   (HubView, HubRecordLink, HubFieldDependency, HubBaseMember, HubBaseShare) с UUID/version/UTC/soft-delete.
3. **Ядро:** собственный **DSL formula engine** (lexer→parser→AST→sandbox-eval, whitelist-функции, лимиты),
   field-level DAG с детекцией циклов и кэшем `computed`; relations через join + lookup/rollup с cross-record NATS-пересчётом.
4. **Views:** серверный интерпретатор filters/sorts/groups + cursor-пагинация; FE — виртуализированный grid с
   типизированными редакторами и a11y-навигацией, view-switcher (kanban/calendar/gallery), formula-редактор с автокомплитом.
5. **Roadmap:** 15 фаз (B5.0–B5.14), ~58 атомарных планов, волны ≤4 агента; критический путь через formula core→integ.
6. **Оценка:** XL (~7–9 недель при 2 be + 2 fe + qa); главный риск — безопасность/производительность formula engine.
7. **Перед стартом** нужны ответы на 8 продуктовых вопросов (глубина паритета, лимиты, perf-SLA, automation, coverage).
8. **Ворота:** каждая фаза — verifier + `/qa-gate` + team-lead приёмка; formula/sharing — доп. security-auditor.
