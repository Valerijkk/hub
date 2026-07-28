# W2 агент A5 — bases

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: `authorizeBase()` owner-scope (fix `98fb3df`), IDOR foreign base, table/row RBAC, soft delete, N+1, LimitReader, RBAC tests.  
> Сверка с `grok-2026-07`: A13-002 **закрыт** (`MaxTableRecords` + `GetTableMeta`); A4-013 CreateBase-tx — **ещё открыт**.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/bases/dto.go` | 120 | 1 | замечания |
| `backend/internal/bases/export_hooks.go` | 23 | 0 | чисто |
| `backend/internal/bases/handler.go` | 281 | 0 | чисто |
| `backend/internal/bases/repository.go` | 261 | 5 | замечания |
| `backend/internal/bases/service.go` | 354 | 8 | замечания |
| `backend/testframework/bases/fixtures_test.go` | 54 | 0 | чисто |
| `backend/testframework/bases/handler_flow_test.go` | 147 | 0 | чисто |
| `backend/testframework/bases/handler_test.go` | 172 | 1 | замечания |
| `backend/testframework/bases/handler_wave107_test.go` | 72 | 0 | чисто |
| `backend/testframework/bases/handler_wave115_test.go` | 139 | 0 | чисто |
| `backend/testframework/bases/rbac_test.go` | 236 | 1 | замечания |
| `backend/testframework/bases/repository_table_cap_a13002_test.go` | 81 | 0 | чисто |
| `backend/testframework/bases/repository_wave98_test.go` | 79 | 1 | замечания |
| `backend/testframework/bases/service_test.go` | 138 | 1 | замечания |
| `backend/testframework/bases/service_wave115_test.go` | 76 | 0 | чисто |
| `backend/testframework/bases/service_wave98_test.go` | 90 | 1 | замечания |

Файлов в списке: 16. Прочитано: 16. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W2-A5-001 | P2 | `service.go:72-97` | **A4-013 (residual):** `CreateBase` → `CreateTable` → `CreateColumn` без одной Ent-транзакции | При ошибке на column/table остаётся orphan base (или base+table без колонки Name). Частичная запись нарушает инвариант «новая база = Main+Name» | `WithTx`: base+table+column атомарно; rollback при любой ошибке |
| W2-A5-002 | P2 | `service.go:251-252`; `repository.go:248-249` | `UpdateRecord` делает `SetCells(normalizeCells(req))` — **полная замена** JSON, без merge с `current.Cells` | Partial PUT с одним ключом затирает остальные ячейки. FE (`bases-page.tsx`) шлёт `...rec.cells`, но любой другой клиент/скрипт теряет данные. `service_test.go:111-116` обновляет одну колонку после двух и не assert'ит сохранение второй | Merge: `out = copy(current); for k,v := range normalized { out[k]=v }`; либо явно документировать full-replace и требовать полный map. Тест: create 2 cells → update 1 → assert второй жив |
| W2-A5-003 | P2 | `repository.go:105-111`; `service.go:126-130` | `SoftDeleteBase` ставит `deleted_at` только на `HubBase`; tables/columns/records остаются с `deleted_at=NULL`. Нет restore / hard-purge | ТЗ: soft delete через корзину 30 дней. Дети-«живые» orphan-строки; restore/GDPR-purge базы не зачистит дерево. API сейчас недоступен (authorize → GetBase `DeletedAtIsNil`), но модель данных неполная | Каскадный soft-delete children в той же tx; endpoint restore; job hard-purge >30d |
| W2-A5-004 | P2 | `service.go:105-107` + `repository.go:93-95`; `service.go:248-249` + `repository.go:245-247` | Optimistic lock: сравнение `version` в service, в UPDATE только `AddVersion(1)` **без** `Where(VersionEQ)` | TOCTOU: два параллельных PUT с одним version оба проходят check → оба пишут → lost update, version уезжает на +2 без 409 | Repo: `Where(VersionEQ(req.Version))`; 0 rows → `ErrVersionConflict`. Service передаёт expected version в repo |
| W2-A5-005 | P2 | `repository.go:132-136`; `service.go:171-179` | `GetTable`/`GetTableView` режут records на `MaxTableRecords=500` **молча** — в DTO нет `has_more` / `total` / cursor | Пользователь не видит строки старше cap и не знает, что данные обрезаны. A13-002 закрыл DoS, но UX/контракт API дырявый | Добавить `truncated`/`has_more`/`total` + offset/cursor pagination в `TableViewResponse` |
| W2-A5-006 | P2 | `service_test.go:18`; `handler_test.go:23`; `repository_wave98_test.go:18` | SQLite DSN зафиксирован (`file:bases_svc`, `file:bases`, `file:bases_repo`) + `cache=shared` | При `go test -parallel` фикстуры шарят одну in-memory БД → flaky / ложные FAIL/PASS. `rbac_test.go:26` уже использует `uuid.NewString()` — эталон | Уникальный DSN на тест: `file:bases_<uuid>?mode=memory&cache=shared&_fk=1` |
| W2-A5-007 | P3 | `service.go:141-151` | `CreateTable` + default column `Name` без tx (тот же паттерн, что A4-013) | Ошибка `CreateColumn` → таблица без дефолтной колонки | Одна tx table+column |
| W2-A5-008 | P3 | `service.go:198-204`; `dto.go:98-103` | `ColType` не валидируется allowlist'ом констант; пустой → `text`, иначе as-is в Ent | Невалидный `col_type` → ошибка Ent → `500 internal`, не `400` | Switch/allowlist `ColumnText…ColumnUserRef`; иначе `ErrInvalidInput` |
| W2-A5-009 | P3 | `rbac_test.go:146-187` | Denial-suite бьёт foreign **base** (403), но нет кейса: owner baseA + `tableId` из baseB (тот же или чужой owner) → 404 через `HasBaseWith` | Регресс снятия `hubtable.HasBaseWith(baseID)` в `GetTable`/`GetTableMeta` не поймается, пока `authorizeBase` жив | Тест: две базы одного owner; `GetTableView(baseA, tableB)` → 404; mutation-proof на repo |
| W2-A5-010 | P3 | `service_wave98_test.go:36-41` | `TestMapServiceError_AllBranches` не покрывает `ErrForbidden` → 403 | Ветка `mapServiceError` для Forbidden тестируется только косвенно в `rbac_test.go:235`; имя «AllBranches» врёт | Добавить `assert.Equal(403, fiberStatus(mapServiceError(bases.ErrForbidden)))` |
| W2-A5-011 | P3 | `service.go:105`; `service.go:248` | `if req.Version > 0` — при `version=0`/omit конфликт не проверяется | Клиент без version получает last-write-wins вопреки полю `version` в схеме | Требовать version на PUT (400 если 0) или atomic `VersionEQ` всегда |
| W2-A5-012 | P3 | `repository.go:65-78` | `ListBases` без `Limit`/пагинации | Рост числа баз у пользователя → неограниченный ответ (слабее, чем records, но паттерн тот же) | Offset/limit или разумный cap + cursor |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `service.go:42-57` `authorizeBase` | Owner-scope: `Edges.Owner.ID != userID` → `ErrForbidden`; `ent.IsNotFound` → `ErrNotFound` (не 403 на missing — enumeration closed). Вызывается из Get/Update/DeleteBase, CreateTable, GetTableView, CreateColumn, Create/Update/DeleteRecord |
| `repository.go:65-78` `ListBases` | `HasOwnerWith(ownerID)` + `DeletedAtIsNil` — аналог NF-03 закрыт |
| `repository.go:121-138` `GetTable` / `142-154` `GetTableMeta` | `HasBaseWith(baseID, DeletedAtIsNil)` — IDOR table↔чужой base закрыт на repo |
| `repository.go:156-166` `GetRecord` | Scope `HasTableWith(tableID)` — row не из другой таблицы |
| Soft delete filters | Base/table/column/record queries и updates с `DeletedAtIsNil`; hard delete HTTP нет |
| A13-002 | `MaxTableRecords=500` на view; write-path через `GetTableMeta`/`GetRecord` — N+1/full-scan на mutate закрыт |
| N+1 | Eager `WithOwner`/`WithTables`/`WithColumns`/`WithRecords` — циклов query-per-row нет |
| LimitReader | В модуле нет upload/stream body кроме Fiber `Bind().JSON` — N/A (глобальный BodyLimit) |
| `handler.go` | Только authMw (owner ACL на service — осознанная модель, не project-RBAC); UUID parse; `mapServiceError` 404/403/400/409/500; `*ent.Client` нет |
| `export_hooks.go` | hubtest re-export `mapServiceError`/`normalizeCells` — без prod-логики |
| `dto.go` (кроме ColType) | UUID PK в ответах; version fields; cells `map[string]interface{}` |
| `rbac_test.go` (качество) | **Реальный EntRepository** (не мок) + Fiber HTTP denial: list empty, GET/PUT/DELETE/table/column/record → 403; service-level Forbidden vs NotFound; positive owner path. Guard на том слое, где живёт |
| `repository_table_cap_a13002_test.go` | Mutation-ready: cap + meta empty records; asserts есть |
| `handler_*` / `service_wave115` | Happy-path + 400/404/409; soft-delete base → GET 404 |
| `fixtures_test.go` | Type aliases / sentinel re-export — без assert-free Test* |
| `normalizeCells` | Allowlist по UUID колонок — unknown keys drop (покрыто `service_wave98`) |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | 16 / 16 / 0 |
| Находок | **12** |
| P0 | 0 |
| P1 | 0 |
| P2 | 6 |
| P3 | 6 |

### Топ-3

1. **W2-A5-001** — `CreateBase` без tx (residual A4-013): orphan base при сбое default table/column.  
2. **W2-A5-002** — `UpdateRecord` затирает ячейки при partial PUT (нет merge).  
3. **W2-A5-003** — `SoftDeleteBase` без cascade / restore / 30d purge (разрыв с ТЗ-корзиной).

### Вердикт по фокусу аудита

| Инвариант | Статус |
|-----------|--------|
| `authorizeBase` owner-scope (`98fb3df`) | **держит** — foreign base 403, missing 404 |
| IDOR foreign base (HTTP + service) | **закрыт**, покрытие `rbac_test.go` сильное |
| table/row scope (`HasBaseWith` / `HasTableWith`) | **закрыт** в repo; не хватает cross-base regression-теста (W2-A5-009) |
| Soft delete | Base/record soft; cascade/trash UI — gap (W2-A5-003) |
| N+1 / unbounded records | A13-002 **закрыт**; silent truncate — W2-A5-005 |
| LimitReader | N/A в модуле |
| RBAC tests quality | Высокое (real Ent + HTTP); gaps: shared DSN, cross-base table, Forbidden в mapServiceError |

P0/P1 по bases на текущем коде **не найдены**. Критичный owner-scope IDOR не регрессировал.
