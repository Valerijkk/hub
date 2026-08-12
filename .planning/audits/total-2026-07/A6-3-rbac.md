# W6 агент A3 — rbac full

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: Fiber v3 middleware order, cache invalidation, custom roles, fail-closed.  
> Сверка: TZ (`docs/tz_constraints.md`), rule `20-backend.mdc` (Fiber v3 RBAC), prior FA-02-02 / grok A1-002 (IDOR — закрыт в коде).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/rbac/cache.go` | 62 | 1 | замечания |
| `backend/internal/rbac/custom_roles.go` | 236 | 4 | замечания |
| `backend/internal/rbac/export_hooks.go` | 35 | 0 | чисто |
| `backend/internal/rbac/handler.go` | 226 | 3 | замечания |
| `backend/internal/rbac/middleware.go` | 145 | 1 | замечания |
| `backend/internal/rbac/permissions.go` | 138 | 1 | замечания |
| `backend/internal/rbac/repository.go` | 175 | 0 | чисто |
| `backend/internal/rbac/service.go` | 77 | 2 | замечания |
| `backend/testframework/rbac/cache_test.go` | 212 | 0 | чисто |
| `backend/testframework/rbac/custom_roles_a1002_test.go` | 160 | 0 | чисто |
| `backend/testframework/rbac/handler_test.go` | 226 | 1 | замечания |
| `backend/testframework/rbac/middleware_integration_test.go` | 304 | 0 | чисто |
| `backend/testframework/rbac/middleware_test.go` | 95 | 1 | замечания |
| `backend/testframework/rbac/permissions_test.go` | 311 | 0 | чисто |
| `backend/testframework/rbac/service_test.go` | 183 | 1 | замечания |

Файлов в списке: 15. Прочитано: 15. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W6-A3-001 | P1 | `custom_roles.go:103-109` | После `UpdateCustomRole` ошибки `ListRoleMembersWithUser` и `cache.Invalidate` глотаются (`members, _ := …`, `_ = s.cache.Invalidate`) | Смена/сужение permissions роли уже в БД, а кэш участников до `defaultCacheTTL` (5 мин) держит старую **расширенную** маску → privilege retention | Не игнорировать err: при ошибке list/invalidate — fail update **или** retry + `InvalidateProject`; логировать; тест: member с кэшем → update perms ↓ → следующий `RequirePermission` сразу 403 |
| W6-A3-002 | P1 | `custom_roles.go:132-137` | `DeleteCustomRole`: `ListRoleMemberIDs` / `ClearMemberCustomRoleByID` / `InvalidateProject` — все ошибки через `_` | List fail → члены не сброшены, роль удаляется; Invalidate fail → кэш с elevated bitmask живёт до TTL; Clear fail → dangling `custom_role` edge | Транзакция: clear members + delete role; `InvalidateProject` с проверкой err (лог + retry); при ошибке clear — не удалять роль; тест на swallow-path |
| W6-A3-003 | P1 | `service.go:48-51`; `service_test.go:167-183` | `IsSystemAdmin` → кэш `^Permission(0)` на 5 мин; demote `SystemRole` **не** инвалидирует RBAC-кэш (в модуле нет хука). Тест `AdminCachedInRedis` **фиксирует** stale full-mask после demote | До TTL бывший admin/founder проходит любой `RequirePermission` по проекту | При смене `system_role` (audit/auth) — `Invalidate` всех ключей user или короткий TTL/не кэшировать admin-bypass; тест: demote → assert cache miss / 403 |
| W6-A3-004 | P2 | `service.go:63-66`; dual-write `custom_roles.go:34-36` | `GetEffectivePermissions` берёт только `customRole.Permissions` (uint64), **игнорирует** `permission_names` | Комментарии: имена — source of truth, Bit=0 невидимы bitmask-fast-path. Middleware никогда не увидит Bit=0 perm → fail-closed для новых прав **или** рассинхрон names↔uint64 | Effective: names → mask (+ Bit=0 path); `RequirePermission` по имени для Bit=0; тест dual-write mismatch |
| W6-A3-005 | P2 | `permissions.go:126-128` | `Has(0)`: `p&0 == 0` → **всегда true** | Будущий `RequirePermission(PermX)` с `Bit=0` = fail-open для всех, у кого дошли до `Has` | Guard: `if perm == 0 { deny/error }`; middleware не принимать Bit=0 через bitmask API |
| W6-A3-006 | P2 | `handler.go:56-59`; `handler.go:117-124` | `CreateCustomRole`/`UpdateCustomRole`: `unknown permission name` и прочие service-err → **500** | Клиентская ошибка валидации должна быть 400; 500 маскирует контракт и мешает клиентам | `errors.Is` / prefix `unknown permission` → 400; system/mismatch уже мапятся отдельно |
| W6-A3-007 | P2 | `handler.go:51-54`; `custom_roles.go:12-16` | Теги `validate:"required,…"` на `CustomRoleRequest` **не исполняются** (нет `validate.Struct` / StructValidator) | Пустой `permissions`/`name` доходят до service; DB `NotEmpty` ловит name, пустой perm-list — валидная «пустая роль» без HTTP-guard | После Bind: `validate.Struct(&req)` → 400; тест empty body fields |
| W6-A3-008 | P2 | `custom_roles.go:132-139` | Delete: clear members → invalidate → `DeleteRole` **без транзакции** | Частичный clear + fail delete / наоборот → несогласованность members↔role | Ent Tx: clear + delete; invalidate после commit |
| W6-A3-009 | P2 | `custom_roles.go:172-177` | `ClearCustomRole`: DB clear успешен, затем `return s.cache.Invalidate(...)` — при err Invalidate клиент получает 500, кэш **со старой elevated** маской | Demotion в БД есть, authz по кэшу — нет до TTL (то же семейство, что W6-A3-001) | Invalidate до clear с rollback, или best-effort InvalidateProject + не считать 500 успешной demotion без cache drop; метрика/алерт |
| W6-A3-010 | P2 | `handler_test.go:19-29`; `handler_test.go:32-47` | Handler-тесты: happy-path; `authMw` = no-op; **нет** denial (viewer → CreateRole 403) | Мутация «убрать `RequirePermission` с Post Create» не краснеет в `handler_test` | `TestHandler_CreateRole_Forbidden_Viewer` + остальные mutate routes |
| W6-A3-011 | P2 | `cache_test.go` / `service_test.go` (нет кейса) | Нет теста: UpdateCustomRole инвалидирует кэш членов; нет кейса «ListRoleMembers err → …» | Регресс W6-A3-001 не ловится | После Assign+GetEffective → Update perms ↓ → без ручного Invalidate assert !Has(elevated) |
| W6-A3-012 | P3 | `handler.go:212-225` | `UnassignRole`: path `:roleId` **не парсится**; всегда `ClearCustomRole(project, user)` | Контракт URL врёт; любой `roleId` при `manage_roles` снимает кастомную роль | Парсить `roleId`; сверять с текущим custom role member; mismatch → 404 |
| W6-A3-013 | P3 | `middleware.go:17-42` | `RequireAuth` только кладёт Bearer в Locals, **JWT не валидирует**, `userID` не ставит | Имя = auth; в prod не используется (только тесты), но копипаст = fail-open при «любом Bearer» | Deprecated/rename `CaptureBearerToken`; или делегировать в `auth.RequireAuth`; не экспортировать как auth |
| W6-A3-014 | P3 | `cache.go:52-61` | `InvalidateProject`: SCAN+DEL по одному ключу, не атомарно относительно concurrent `Set` | Окно: GetEffectivePermissions успевает `Set` после SCAN/до DEL → stale key до TTL | Lua/UNLINK pipeline; или version stamp в ключе; после mutate — retry SCAN |
| W6-A3-015 | P3 | `middleware_test.go:59-73` | `TestRequireAuth_ValidBearerNext` зелёный на `Bearer jwt-token-123` без подписи | Цементирует отсутствие проверки JWT у helper | Либо удалить helper из публичного API, либо тест «только extract» + явный комментарий; JWT — в `auth` fail-closed suite |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| Fiber v3 order (`handler.go:24-29`) | Канон `route.Method(path, handler, RequirePermission…)` — MW **перед** terminal; совпадает с tracker/tms и `middleware_integration_test.go:144-165` |
| `RequirePermission` fail-closed (`middleware.go:97-107`) | Нет userID → 401; not member → 403; прочий err GetEffective → **500** (не Next); !Has → 403 |
| `RequireAdmin` / founder (`middleware.go:113-136`) | `auth.IsSystemAdmin`; anonymous 401; покрыто integration-тестом |
| IDOR project↔role (custom_roles + a1002) | Update/Delete/Assign: `GetRoleProjectID` vs URL; system role update/delete blocked; stub-тесты A1-002/003 |
| Cache Get miss (`cache.go:29-37`) | Redis err / bad len → miss → DB (не fail-open с 0-perm как «разрешено» без membership check — membership всё равно из DB) |
| Assign invalidate (`custom_roles.go:162`) | Ошибка Invalidate **пробрасывается** (в отличие от Update/Delete) |
| Unknown `ProjectRole` (`service.go:61`) | `DefaultRolePermissions[unknown]` → 0 → fail-closed на Has |
| `repository.go` | Layering: Ent только в repo; IDOR helpers; no `*ent.Client` в handler/service |
| `export_hooks.go` | Только hubtest re-export |
| `permissions_test.go` | Registry unique bits/names; ValidPermissionName; validate names gate; round-trip bitmask |
| `custom_roles_a1002_test.go` | Foreign project reject; system update reject |
| `middleware_integration_test.go` | Anonymous/non-member/with/without perm; invalid UUID; missing projectId; Fiber v3 order documented |
| `cache_test.go` (базовое) | Get/Set/Invalidate; InvalidateProject isolation; Create/Assign/Delete custom role happy paths |

**Fiber v3 middleware order в этом списке файлов — OK** (регрессия FA-02-02 здесь не воспроизводится).

**Fail-closed на check-path:** в целом да (deny/500 вместо Next). Слабое место — **кэш после demote/update/delete** (W6-A3-001…003, 009): check fail-closed, **mutation invalidate — нет**.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **15 / 15 / 0** |
| Находок | **15** |
| P0 | 0 |
| P1 | 3 |
| P2 | 8 |
| P3 | 4 |

### Топ-3

1. **W6-A3-001 (P1)** — UpdateCustomRole глотает ошибки list/invalidate → до 5 мин stale elevated permissions.  
2. **W6-A3-002 (P1)** — DeleteCustomRole глотает clear/invalidate → dangling membership + stale cache после удаления роли.  
3. **W6-A3-003 (P1)** — кэш admin-bypass `^0` без инвалидации при demote SystemRole (тест явно сохраняет full mask).

**Residual вне списка:** инвалидация при Add/Remove/UpdateMemberRole живёт в `project/service.go` (логирует err, не fail); смена `system_role` — в auth/audit без RBAC-invalidate; схема `Role` без `version` (W8 ent/schema).
