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
| W6-A3-006 | FIXED | TASK-115 | unknown permission name → 400 (ErrUnknownPermission) |
| W6-A3-007 | FIXED-OK | закрыт | StructValidator + `validate:"required"` на CustomRoleRequest |
| W6-A3-008 | FIXED | TASK-115 | ClearMembersAndDeleteRole Ent Tx |
| W6-A3-009 | FIXED | TASK-115 | ClearCustomRole: cache-fail ≠ 500 после demotion |
| W6-A3-010 | FIXED-OK | закрыт | `authz_ent_negative_task037_test.go` CreateRole Forbidden |
| W6-A3-011 | FIXED-OK | закрыт | `cache_invalidate_p1be33_test.go` |
| W6-A3-012 | FIXED-OK | закрыт | UnassignRole парсит `:roleId` + match current |
| W6-A3-013 | FIXED | TASK-115 | `CaptureBearerToken` rename; `RequireAuth` deprecated alias |
| W6-A3-014 | FIXED-OK | закрыт | InvalidateProject/User через Lua SCAN+UNLINK |
| W6-A3-015 | FIXED | TASK-115 | extract-only тест CaptureBearerToken (не ValidBearer) |

REGRESS: **0**. Fiber v3 MW order: **OK**.

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|----------|---------------|------------|
| A2W6-A3-001 | FIXED | P2 | `handler.go` Create/Update | `unknown permission name` → 500 | Клиентская валидация должна быть 400 | TASK-115: `ErrUnknownPermission` → 400 |
| A2W6-A3-002 | FIXED | P2 | `repository.go` ClearMembersAndDeleteRole | Delete без Ent Tx | Partial clear + fail delete | TASK-115: Ent Tx clear+delete; invalidate после commit |
| A2W6-A3-003 | FIXED | P2 | `custom_roles.go` ClearCustomRole | Invalidate fail → 500 после demotion | Misleading 500 | TASK-115: best-effort Invalidate→InvalidateProject, не 500 |
| A2W6-A3-004 | FIXED | P3 | `middleware.go` | `RequireAuth` misnomer | Копипаст fail-open | TASK-115: `CaptureBearerToken` + deprecated alias |
| A2W6-A3-005 | FIXED | P3 | `middleware_test.go` | ValidBearer misleading green | Цементирует no-JWT | TASK-115: extract-only тест `CaptureBearerToken` |

## Counts

| | n |
|--|--:|
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 0 |
| FIXED-OK | 15 |
| P0 / P1 / P2 / P3 (open) | 0 / 0 / 0 / 0 |
| **Σ open** | **0** (TASK-115 2026-07-28) |
