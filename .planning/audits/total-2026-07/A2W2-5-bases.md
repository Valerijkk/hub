# A2 · W2 агент A5 — bases (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: [A2-5-bases.md](A2-5-bases.md) · TASK-038/059/061/062/063

## Прочитанные файлы

| Файл | Строк | Open | Вердикт |
|------|------:|-----:|---------|
| `backend/internal/bases/dto.go` | 104 | 0 | `has_more` + ColType OK |
| `backend/internal/bases/export_hooks.go` | 18 | 0 | чисто |
| `backend/internal/bases/handler.go` | 260 | 0 | чисто |
| `backend/internal/bases/repository.go` | 312 | 1 | RESIDUAL SoftDelete cascade |
| `backend/internal/bases/service.go` | 349 | 0 | authorizeBase + merge + OL OK |
| `backend/testframework/bases/fixtures_test.go` | 47 | 0 | чисто |
| `backend/testframework/bases/handler_flow_test.go` | 129 | 0 | чисто |
| `backend/testframework/bases/handler_test.go` | 146 | 1 | RESIDUAL shared DSN |
| `backend/testframework/bases/handler_wave107_test.go` | 61 | 0 | чисто |
| `backend/testframework/bases/handler_wave115_test.go` | 120 | 0 | чисто |
| `backend/testframework/bases/rbac_test.go` | 199 | 0 | owner-scope держит |
| `backend/testframework/bases/repository_table_cap_a13002_test.go` | 66 | 0 | чисто |
| `backend/testframework/bases/repository_wave98_test.go` | 69 | 0 | unique DSN via t.Name |
| `backend/testframework/bases/service_test.go` | 112 | 1 | RESIDUAL shared DSN |
| `backend/testframework/bases/service_wave115_test.go` | 59 | 0 | чисто |
| `backend/testframework/bases/service_wave98_test.go` | 81 | 0 | Forbidden branch FIXED |

Файлов: **16/16/0**.

## FIXED-OK

| Stage A ID | Evidence |
|------------|----------|
| W2-A5-001 | `CreateBase` одна Ent tx base+Main+Name (`repository.go:60–…`) |
| W2-A5-002 | `UpdateRecord` merge cells (`service.go:248–257`) |
| W2-A5-004 | `VersionEQ` в WHERE UpdateBase/Table/Column/Record |
| W2-A5-005 | `TableViewResponse.HasMore` + `CountTableRecords` |
| W2-A5-007 | `CreateTable` tx table+default column |
| W2-A5-008 | `isValidColumnType` allowlist → 400 |
| W2-A5-009 | `cross_base_table_task059_test.go` GetTableView cross-base → 404 |
| W2-A5-010 | `TestMapServiceError_AllBranches` включает ErrForbidden→403 |
| W2-A5-011 | `req.Version < 1` → ErrInvalidInput |
| W2-A5-012 | `ListBases` `Limit(MaxOwnerBases)` + cap-тест |
| authorizeBase / HasBaseWith / A13-002 | **без регрессии** |

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | Почему | Фикс |
|----|-----|-----|-------------|----------|--------|------|
| A2W2-5-001 | RESIDUAL (W2-A5-003) | P2 | `repository.go:146–152` | `SoftDeleteBase` ставит `deleted_at` только на HubBase; tables/columns/records остаются живыми | ТЗ-корзина 30д неполна; нет cascade/restore/purge | Tx cascade soft-delete children + restore/purge job |
| A2W2-5-002 | RESIDUAL (W2-A5-006) | P2 | `service_test.go:18`; `handler_test.go:23` | Фиксированный SQLite DSN (`file:bases_svc`, `file:bases`) + `cache=shared` | Parallel flaky; другие файлы уже на uuid/t.Name | `file:bases_<uuid>?mode=memory&cache=shared` |

## Инварианты

| Инвариант | A2 статус |
|-----------|-----------|
| `authorizeBase` owner-scope | **держит** — нет REGRESS |
| Foreign base HTTP 403 | **держит** (`rbac_test.go`) |
| table/row `HasBaseWith` | **держит** + cross-base regression |
| Optimistic lock VersionEQ | **FIXED-OK** |
| Soft-delete cascade/trash | **RESIDUAL** |
| P0/P1 | **0** |

## Сводка

| | Count |
|--|------:|
| Files | 16/16/0 |
| FIXED-OK classes | 11 |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 2 |
| Open P0/P1/P2/P3 | 0 / 0 / 2 / 0 |
