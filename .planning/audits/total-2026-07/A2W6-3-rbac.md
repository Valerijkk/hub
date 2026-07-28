# A2 W6 агент A3 — rbac full

> TASK-008 · Stage A2 · read-only · 2026-07-28  
> Периметр: `W6-ASSIGNMENTS.md` §A3 (15 файлов) · сверка TASK-023 / TASK-045 / TASK-066 / TASK-037

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/rbac/cache.go | 85 | 0 | FIXED-OK (Lua InvalidateProject/User) |
| backend/internal/rbac/custom_roles.go | 273 | 2 | RESIDUAL (tx / Clear invalidate) |
| backend/internal/rbac/export_hooks.go | 35 | 0 | чисто |
| backend/internal/rbac/handler.go | 233 | 1 | RESIDUAL (unknown perm→500) |
| backend/internal/rbac/middleware.go | 145 | 1 | RESIDUAL (RequireAuth misnomer) |
| backend/internal/rbac/permissions.go | 142 | 0 | FIXED-OK (Has(0)) |
| backend/internal/rbac/repository.go | 175 | 0 | чисто |
| backend/internal/rbac/service.go | 115 | 0 | FIXED-OK (names SoT + demote recheck) |
| backend/testframework/rbac/cache_test.go | 212 | 0 | чисто |
| backend/testframework/rbac/custom_roles_a1002_test.go | 160 | 0 | чисто |
| backend/testframework/rbac/handler_test.go | 226 | 0 | FIXED-OK (denial вне списка: TASK-037) |
| backend/testframework/rbac/middleware_integration_test.go | 304 | 0 | чисто |
| backend/testframework/rbac/middleware_test.go | 101 | 1 | RESIDUAL |
| backend/testframework/rbac/permissions_test.go | 314 | 0 | FIXED-OK |
| backend/testframework/rbac/service_test.go | 187 | 0 | FIXED-OK (demote) |

Файлов: **15/15/0**.

## Перепроверка Stage A (W6-A3)

| Старый ID | Tag | Статус | Комментарий |
|-----------|-----|--------|-------------|
| W6-A3-001 | FIXED-OK | закрыт | Update: list/invalidate err пробрасываются (TASK-023) |
| W6-A3-002 | FIXED-OK | закрыт | Delete: clear/invalidate с err-check |
| W6-A3-003 | FIXED-OK | закрыт | full-mask cache → recheck system_role; demote test RED→GREEN |
| W6-A3-004 | FIXED-OK | закрыт | `permission_names` source of truth |
| W6-A3-005 | FIXED-OK | закрыт | `Has(0)` → false |
| W6-A3-006 | RESIDUAL | жив | unknown permission name → 500 (Create/Update) |
| W6-A3-007 | FIXED-OK | закрыт | StructValidator + `validate:"required"` на CustomRoleRequest |
| W6-A3-008 | RESIDUAL | частичный | clear+delete без одной Ent-tx (ошибок меньше, атомарности нет) |
| W6-A3-009 | RESIDUAL | жив | ClearCustomRole: DB clear → Invalidate; Invalidate fail → 500 при demotion в БД |
| W6-A3-010 | FIXED-OK | закрыт | `authz_ent_negative_task037_test.go` CreateRole Forbidden |
| W6-A3-011 | FIXED-OK | закрыт | `cache_invalidate_p1be33_test.go` |
| W6-A3-012 | FIXED-OK | закрыт | UnassignRole парсит `:roleId` + match current |
| W6-A3-013 | RESIDUAL | жив | `RequireAuth` не валидирует JWT (только Locals token) |
| W6-A3-014 | FIXED-OK | закрыт | InvalidateProject/User через Lua SCAN+UNLINK |
| W6-A3-015 | RESIDUAL | жив | `TestRequireAuth_ValidBearerNext` на fake Bearer |

REGRESS: **0**. Fiber v3 MW order: **OK**.

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|----------|---------------|------------|
| A2W6-A3-001 | RESIDUAL | P2 | `handler.go:56-59`, `117-124` | `unknown permission name` → 500 | Клиентская валидация должна быть 400 | `strings.Contains` / typed err → 400 |
| A2W6-A3-002 | RESIDUAL | P2 | `custom_roles.go:146-160` | Delete: clear members → invalidate → DeleteRole без Ent Tx | Partial clear + fail delete → inconsistency | Tx: clear+delete; invalidate после commit |
| A2W6-A3-003 | RESIDUAL | P2 | `custom_roles.go:209-214` | Clear: demotion в БД, затем Invalidate; err → 500 при уже cleared | Privilege retention до TTL при cache fail | Invalidate до clear / InvalidateProject + не 500 на cache-only fail |
| A2W6-A3-004 | RESIDUAL | P3 | `middleware.go:14-40` | `RequireAuth` misnomer — JWT не проверяет | Копипаст = fail-open «любой Bearer» | Rename `CaptureBearerToken` / deprecate |
| A2W6-A3-005 | RESIDUAL | P3 | `middleware_test.go` ValidBearer | Цементирует отсутствие JWT check | Misleading green | Удалить helper из публичного API или явный extract-only тест |

## Counts

| | n |
|--|--:|
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 5 |
| FIXED-OK | 10 |
| P0 / P1 / P2 / P3 (open) | 0 / 0 / 3 / 2 |
| **Σ open** | **5** |
