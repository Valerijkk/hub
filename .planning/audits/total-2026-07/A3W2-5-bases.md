# A3 W2 агент A5 — bases (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W2-5-bases.md](A2W2-5-bases.md) · A1: [A2-5-bases.md](A2-5-bases.md)  
> Assignment: [W2-ASSIGNMENTS.md](W2-ASSIGNMENTS.md) § A5 · Периметр: **16** · Прочитано: **16/16** · Пропущено: **0**  
> Remediation: TASK-111 (A2W2-5-001/002)

## Сверка с A2 (A2W2-5-*)

| A2 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| A2W2-5-001 | FIXED-OK | закрыт | `SoftDeleteBase` cascade tx records→columns→tables→base (`repository.go:146–212`); mutation `TestTask111_SoftDeleteBase_Cascade` (вне периметра, `testframework/tms/`) |
| A2W2-5-002 | FIXED-OK | закрыт | unique DSN: `handler_test.go:23` `bases_<uuid>`; `service_test.go:18` `bases_svc_<uuid>`; repo уже `t.Name` |

**REGRESS:** 0

### Сверка A1 / FIXED-OK cascade (подтверждение на HEAD)

| Stage A / A2 FIXED class | Tag | Evidence |
|--------------------------|-----|----------|
| W2-A5-001 CreateBase tx | FIXED-OK | `repository.go:60–94` base+Main+Name |
| W2-A5-002 UpdateRecord merge | FIXED-OK | `service.go:248–257` |
| W2-A5-004 VersionEQ | FIXED-OK | UpdateBase/Table/Column/Record WHERE VersionEQ |
| W2-A5-005 has_more/total | FIXED-OK | `dto.go` + `CountTableRecords` + `service.go:160–171` |
| W2-A5-007 CreateTable tx | FIXED-OK | `repository.go:215–240` |
| W2-A5-008 ColType allowlist | FIXED-OK | `isValidColumnType` → ErrInvalidInput |
| W2-A5-009 cross-base table | FIXED-OK | `HasBaseWith` + regression вне периметра |
| W2-A5-010 ErrForbidden→403 | FIXED-OK | `service_wave98_test.go:36–42` |
| W2-A5-011 version&lt;1 → 400 | FIXED-OK | UpdateBase/UpdateRecord |
| W2-A5-012 ListBases cap | FIXED-OK | `MaxOwnerBases` + Limit |
| authorizeBase / HasBaseWith / A13-002 | FIXED-OK | без регрессии |

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/bases/dto.go` | 122 | 0 | `has_more` + ColType OK |
| `backend/internal/bases/export_hooks.go` | 23 | 0 | чисто |
| `backend/internal/bases/handler.go` | 281 | 0 | чисто |
| `backend/internal/bases/repository.go` | 402 | 2 | cascade OK; SoftDeleteTable asym |
| `backend/internal/bases/service.go` | 372 | 2 | authorizeBase+OL OK; MaxLen/trash gaps |
| `backend/testframework/bases/fixtures_test.go` | 54 | 0 | чисто |
| `backend/testframework/bases/handler_flow_test.go` | 147 | 0 | чисто |
| `backend/testframework/bases/handler_test.go` | 172 | 0 | unique DSN FIXED-OK |
| `backend/testframework/bases/handler_wave107_test.go` | 72 | 0 | чисто |
| `backend/testframework/bases/handler_wave115_test.go` | 139 | 0 | чисто |
| `backend/testframework/bases/rbac_test.go` | 236 | 0 | owner-scope держит |
| `backend/testframework/bases/repository_table_cap_a13002_test.go` | 81 | 0 | чисто |
| `backend/testframework/bases/repository_wave98_test.go` | 83 | 1 | SoftDeleteTable без cascade children |
| `backend/testframework/bases/service_test.go` | 140 | 0 | unique DSN FIXED-OK |
| `backend/testframework/bases/service_wave115_test.go` | 76 | 0 | чисто |
| `backend/testframework/bases/service_wave98_test.go` | 92 | 0 | Forbidden branch FIXED |

Файлов в списке: 16. Прочитано: 16. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A3W2-A5-001 | RESIDUAL (хвост W2-A5-003 / A2W2-5-001) | P3 | `service.go:117–121`; audit `ListTrash`/`RestoreItem`/`PurgeExpired*` (вне модуля) | Cascade soft-delete есть; **нет** restore / admin-trash type `hubbase` / 30d hard-purge | ТЗ-корзина 30д: DELETE базы → вечный 404 без UI restore; GDPR-purge не чистит HubBase tree | Зарегистрировать `hubbase` в audit trash + Restore (cascade clear `deleted_at`) + PurgeExpired job |
| A3W2-A5-002 | NEW | P3 | `repository.go:315–321`; cascade `156–164` | `SoftDeleteTable` ставит `deleted_at` только на table; columns/records живые. `SoftDeleteBase` cascade берёт лишь `hubtable.DeletedAtIsNil()` → дети уже soft-deleted table **не** каскадятся | Асимметрия с SoftDeleteBase; orphans в БД; HTTP DeleteTable пока нет, но repo+тест `repository_wave98` закрепляют неполный путь | Cascade SoftDeleteTable как у base; либо cascade SoftDeleteBase по **всем** tableIDs базы (вкл. deleted) |
| A3W2-A5-003 | NEW | P3 | `service.go:72–76`, `124–128`, `175–178`; schema `MaxLen(200)` | Имя base/table/column &gt;200 не валидируется в service → Ent constraint → `mapServiceError` → **500**, не 400 | Тот же класс, что закрытый W2-A5-008 (ColType allowlist): клиентский ввод → 500 internal | `utf8.RuneCountInString` / `len` ≤200 → `ErrInvalidInput`; regression-тест |

## Чисто (подтверждено после remediation)

- **authorizeBase** owner-scope: foreign → 403, missing → 404; HTTP denial suite `rbac_test.go` жив.
- **CreateBase / CreateTable** атомарные Ent-tx (Main+Name / Name column).
- **UpdateRecord** merge cells + VersionEQ; conflict → 409.
- **GetTableView** `total`/`has_more` при cap `MaxTableRecords`; write-path `GetTableMeta`.
- **ListBases** owner filter + `Limit(MaxOwnerBases)`.
- **ColType** allowlist; `mapServiceError` ErrForbidden→403.
- **SoftDeleteBase cascade** records→columns→tables→base в одной tx (TASK-111).
- **SQLite DSN** уникальны в assignment-фикстурах (`uuid` / `t.Name`).
- Handler: только authMw + service ACL; нет `*ent.Client`; UUID parse; soft-delete record/base → 204/404.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 3 |
| **Total** | **3** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active) | 2 (`A2W2-5-001`, `A2W2-5-002`) |
| FIXED-OK (A1 cascade classes) | 11 |
| RESIDUAL | 1 (`A3W2-A5-001`) |
| NEW | 2 (`A3W2-A5-002`, `A3W2-A5-003`) |
| REGRESS | 0 |

A2 baseline: 2 активных → **FIXED-OK 2** · RESIDUAL 1 (хвост trash) · NEW 2 · REGRESS 0 · open **3** (все P3).

**Топ ID:** A3W2-A5-001 (trash restore/purge) · A3W2-A5-002 (SoftDeleteTable cascade asym) · A3W2-A5-003 (MaxLen→500).
