# A3 W6 агент A3 — rbac full (FINAL re-audit)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · ветка `cursor-grok-4.5`  
> Mirror: [A2W6-3-rbac.md](A2W6-3-rbac.md) · Baseline Stage A: [A6-3-rbac.md](A6-3-rbac.md)  
> Периметр: 15 файлов (8 prod + 7 tests) · Прочитано: **15/15** · Пропущено: **0**  
> Plan: [A3-PLAN.md](A3-PLAN.md) · Assignment: [W6-ASSIGNMENTS.md](W6-ASSIGNMENTS.md) § A3  
> Фокус: Fiber v3 middleware order · fail-closed · project membership · founder

> **Вне списка assignment (evidence FIXED-OK, без отдельных находок слота):**  
> `task115_permission_hygiene_test.go`, `authz_ent_negative_task037_test.go`,  
> `cache_invalidate_p1be33_test.go`, `unassign_invalidate_task066_test.go` ·  
> wiring: `wiring_core.go` (`auth.RequireAuth` + `RegisterRoutes`, `RequireAdmin` на `/docs`).

## Сверка с A2 (A2W6-A3-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W6-A3-001 | FIXED-OK | закрыт | TASK-115: `ErrUnknownPermission` → 400 (`handler.go:58–60`, `121–123`); `task115_permission_hygiene_test.go` |
| A2W6-A3-002 | FIXED-OK | закрыт | Ent Tx `ClearMembersAndDeleteRole` (`repository.go:144–162`); invalidate после commit |
| A2W6-A3-003 | FIXED-OK | закрыт | Clear: best-effort Invalidate→InvalidateProject, не 500 (`custom_roles.go:211–217`) |
| A2W6-A3-004 | FIXED-OK | закрыт | `CaptureBearerToken` + deprecated `RequireAuth` (`middleware.go:14–43`) |
| A2W6-A3-005 | FIXED-OK | закрыт | extract-only тесты (`middleware_test.go:60–98`) |

### Cascade FIXED-OK (Stage A W6-A3-* / A2)

| Stage A / A2 | Evidence (файл:строка) |
|--------------|------------------------|
| W6-A3-001 | Update: list/invalidate err пробрасываются / InvalidateProject fallback — `custom_roles.go:103–121` |
| W6-A3-002 / A2-002 | Tx clear+delete — `repository.go:144–162`; `DeleteCustomRole` — `custom_roles.go:144–157` |
| W6-A3-003 | full-mask cache → recheck `IsSystemAdmin` — `service.go:40–55`; demote test RED→GREEN |
| W6-A3-004 | `permission_names` SoT — `service.go:77–84` |
| W6-A3-005 | `Has(0)` → false — `permissions.go:127–131` |
| W6-A3-006 / A2-001 | unknown perm → 400 |
| W6-A3-007 | StructValidator в `wiring_fiber.go` + `validate` на `CustomRoleRequest` |
| W6-A3-008 | = A2-002 Tx |
| W6-A3-009 / A2-003 | Clear cache hygiene |
| W6-A3-010 | CreateRole Forbidden viewer — `authz_ent_negative_task037_test.go` (вне слота) |
| W6-A3-011 | Update invalidate покрыт service/cache + P1-BE-33 suite (вне слота) |
| W6-A3-012 | Unassign `:roleId` match — `custom_roles.go:183–204`; TASK-066 test |
| W6-A3-013 / A2-004 | CaptureBearerToken |
| W6-A3-014 | Lua SCAN+DEL — `cache.go:12–27`, `83–85` |
| W6-A3-015 / A2-005 | extract-only tests |

**REGRESS:** 0 · Fiber v3 MW order: **OK**

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/rbac/cache.go` | 85 | 1 | FIXED-OK Lua; NEW comment DEL≠UNLINK |
| `backend/internal/rbac/custom_roles.go` | 275 | 1 | FIXED-OK Clear/Update/IDOR; NEW Delete cache→500 |
| `backend/internal/rbac/export_hooks.go` | 35 | 0 | чисто |
| `backend/internal/rbac/handler.go` | 239 | 0 | FIXED-OK Fiber v3 order + 400 unknown |
| `backend/internal/rbac/middleware.go` | 145 | 1 | FIXED-OK Capture/RequireAdmin; NEW RequireSystemRole |
| `backend/internal/rbac/permissions.go` | 142 | 0 | FIXED-OK Has(0) |
| `backend/internal/rbac/repository.go` | 203 | 0 | FIXED-OK Tx ClearMembersAndDeleteRole |
| `backend/internal/rbac/service.go` | 115 | 0 | FIXED-OK demote recheck + founder bypass |
| `backend/testframework/rbac/cache_test.go` | 212 | 0 | чисто |
| `backend/testframework/rbac/custom_roles_a1002_test.go` | 164 | 0 | чисто IDOR/system |
| `backend/testframework/rbac/handler_test.go` | 226 | 0 | happy-path; denial вне слота TASK-037 |
| `backend/testframework/rbac/middleware_integration_test.go` | 304 | 0 | membership/founder/Fiber v3 |
| `backend/testframework/rbac/middleware_test.go` | 120 | 0 | FIXED-OK extract-only |
| `backend/testframework/rbac/permissions_test.go` | 314 | 0 | FIXED-OK |
| `backend/testframework/rbac/service_test.go` | 187 | 0 | FIXED-OK demote + founder |

Файлов в списке: **15**. Прочитано: **15**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A3W6-3-001 | NEW | P2 | `custom_roles.go:154–156` (vs Clear `211–217`) | `DeleteCustomRole`: после успешного Tx clear+delete `InvalidateProject` fail → `return err` → handler **500** | Роль уже удалена / edges сняты; клиент видит провал; elevated bitmask в Redis живёт до TTL. TASK-115 закрыл тот же класс только для `ClearCustomRole` | Как Clear: best-effort InvalidateProject + slog/metric, `return nil`; тест: close Redis → Delete HTTP 200 + role gone (mutation: вернуть hard-fail → 500) |
| A3W6-3-002 | NEW | P3 | `middleware.go:45–68` | `RequireSystemRole(role)` — exact string match; founder ≠ `"admin"` | Prod сейчас использует `RequireAdmin` (`IsSystemAdmin`). Экспорт exact-match — копипаст-footgun: `RequireSystemRole("admin")` режет founder (ORG-04) | Deprecate в пользу `RequireAdmin`; или `IsSystemAdmin` при `role=="admin"`; тест founder + RequireSystemRole("admin") |
| A3W6-3-003 | NEW | P3 | `cache.go:12–26` vs комментарий `:69` | Комментарий/имя: «Lua SCAN+**UNLINK**»; скрипт зовёт `DEL` | Не security-баг (Lua-атомарность есть), но drift документации после W6-A3-014 | Либо `UNLINK` в скрипте, либо поправить комментарий на DEL |

## FIXED-OK (сводка)

| ID | Evidence |
|----|----------|
| A2W6-A3-001..005 | TASK-115 в HEAD; typed err→400; Tx delete; Clear cache hygiene; CaptureBearerToken |
| W6-A3-003..005, 012..014 | demote recheck; names SoT; Has(0); Unassign roleId; Lua invalidate |
| Fiber v3 | `handler.go:24–29` = `Method(path, handler, RequirePermission…)`; group `authMw` = `auth.RequireAuth` (`wiring_core.go:41`) |
| Fail-closed check | no userID→401; not member→403; !Has→403; GetEffective err→500 (не Next) |
| Founder | `RequireAdmin` + `GetEffectivePermissions` bypass без membership (`service.go:57–66`; integration/service tests) |
| Membership | `GetProjectMemberWithCustomRole` NotFound → 403 «not a project member» |

## Чисто

- **Fiber v3 order:** канон соблюдён в rbac RegisterRoutes; FA-02-02 не регрессирует.
- **Fail-closed:** `Has(0)`; unknown project_role → mask 0; non-member 403; CaptureBearerToken не в prod path.
- **Founder/admin:** `auth.IsSystemAdmin` в RequireAdmin и GetEffectivePermissions.
- **IDOR:** projectId↔role на Update/Delete/Assign/Clear; system roles blocked; a1002 stubs.
- **Layering:** Ent только в repository; handler → service.
- **Тесты слота:** cache Get/Set/InvalidateProject; permissions registry; middleware anonymous/member/admin/founder; service demote/custom role; handler happy-path (+ TASK-115/037 вне слота).

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 2 |
| **Total** | **3** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active → closed) | 5 (`A2W6-A3-001`…`005`) |
| FIXED-OK (Stage A cascade) | 15 |
| RESIDUAL | 0 |
| NEW | 3 (`A3W6-3-001`…`003`) |
| REGRESS | 0 |

**Топ IDs:** `A3W6-3-001` (Delete cache-fail → 500 / stale elevated) · `A3W6-3-002` (RequireSystemRole founder footgun) · `A3W6-3-003` (DEL≠UNLINK doc).
