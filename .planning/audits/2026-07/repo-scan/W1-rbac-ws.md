# W1 — Security Audit: `backend/internal/rbac/` + `backend/internal/ws/`

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** security-auditor · **Дата:** 2026-07-23 · **Режим:** readonly  
> **Фокус:** NF-02 (WS membership/RBAC before subscribe), Fiber v3 middleware order, CheckOrigin  
> **OWASP/ASVS:** A01 Broken Access Control · A05 Misconfiguration · A07 Auth Failures · ASVS L2 V4/V13  
> **Код не менялся.**

## Чек-лист (кратко)

| Контроль | Результат |
|----------|-----------|
| NF-02: `CanSubscribe` до `Hub.Subscribe` | ✅ PASS — `handler.go:232-246` |
| Unknown room prefix fail-closed | ✅ PASS — `authorizer.go:77-79` |
| nil MembershipChecker fail-closed | ✅ PASS — `authorizer.go:63-75` |
| Production wiring передаёт реальный Authorizer | ✅ PASS — `wiring_ws.go:37-65` |
| CheckOrigin + `CORS_ORIGINS` | ✅ PASS — `handler.go:85-111` |
| Prod без origins → reject (не DEV) | ✅ PASS — `handler.go:75-78,102-106` |
| Fiber v3 RBAC arg order в `rbac/handler.go` | ✅ PASS — handler → `RequirePermission` |
| Group `authMw` на `/projects/:projectId/roles` | ✅ PASS — `handler.go:20` (`auth.RequireAuth` из wiring) |

## Findings

| severity | file:line | issue | evidence | fix |
|----------|-----------|-------|----------|-----|
| HIGH | `rbac/handler.go:100-114` + `rbac/custom_roles.go:73-87` + `rbac/repository.go:85-91` | Cross-project IDOR: UPDATE роли по `:roleId` без проверки принадлежности к `:projectId` (A01). Достаточно `PermProjectManageRoles` на любом проекте + UUID чужой роли. | `UpdateRole` парсит только `roleId`; `UpdateCustomRole`/`UpdateOneID(roleID)` не сверяют `GetRoleProjectID` с path `projectId`. Middleware проверяет права лишь на path-проект. | Перед update: `proj := GetRoleProjectID(roleID)`; если `proj != pathProjectID` → 404/403. Либо `Update…Where(role.IDEQ, role.HasProjectWith(project.IDEQ))`. Regression: manage на A + role из B → deny. |
| HIGH | `rbac/handler.go:130-138` + `rbac/custom_roles.go:103-122` + `rbac/repository.go:140-141` | Cross-project IDOR: DELETE роли по `:roleId` без привязки к `:projectId` (A01). | `DeleteRole`/`DeleteCustomRole`/`DeleteOneID(roleID)` не сравнивают project edge роли с path. | То же: assert `role.project_id == path projectId` до delete; иначе 404. Тест denial cross-project. |
| HIGH | `rbac/custom_roles.go:126-137` + `rbac/repository.go:153-156` + `ent/schema/projectmember.go:42-44` | Assign: кастомная роль другого проекта может быть повешена на member текущего (privilege injection). В схеме нет constraint «role.project == member.project». | `AssignCustomRole` проверяет member∈project, но не `GetRoleProjectID(roleID)==projectID`; `SetCustomRoleID` принимает любой UUID роли. | Перед assign: роль принадлежит `projectID` и `!IsSystem` (если нужно). Опционально DB/app invariant. Тест: role из B → assign в A → 403. |
| MEDIUM | `ws/handler.go:146-151` | Footgun: `authz == nil` → `allowAllAuthorizer` (полный bypass NF-02). Wiring сейчас безопасен, но любой будущий `Handler(hub, auth, nil)` молча открывает broadcast-IDOR. | `if authz == nil { authz = allowAllAuthorizer{} }`; prod: `wiring_ws.go:65` передаёт `wsAuthz`. | Fail-closed: `nil` → panic на старте / return 500 handler / запретить nil в prod build tag. `allowAll` только через явный test export. |
| MEDIUM | `ws/handler.go:139-157` | Access JWT в query `?token=` — утечка в access/proxy/CDN/browser history logs (A07/A09). | Комментарий «WebSocket не поддерживает кастомные заголовки»; токен читается из `c.Query("token")` до upgrade. | Предпочтительно short-lived WS ticket (одноразовый, TTL секунды) вместо access JWT; не логировать query; Sec-WebSocket-Protocol / post-connect auth если клиент позволяет. |
| MEDIUM | `rbac/middleware.go:16-41` | `rbac.RequireAuth()` не валидирует JWT: только кладёт raw token в Locals. Имя API вводит в заблуждение; при ошибочном использовании вместо `auth.RequireAuth` — auth bypass. | Комментарий «JWT валидация в auth.RequireAuth»; prod wiring использует `auth.RequireAuth` (`wiring_auth.go:73`, `wiring_core.go:39`). | Переименовать/удалить stub или сделать thin-wrapper над `auth.RequireAuth`; запретить standalone. |
| MEDIUM | `rbac/service.go:12,37-40,47-50` | Кеш effective permissions TTL 5 мин: после revoke/demote возможен stale allow до TTL, если invalidate пропущен/упал (A01 timing). | `defaultCacheTTL = 5 * time.Minute`; admin `^Permission(0)` тоже кешируется; invalidate есть на update/assign/delete, но ошибки глотаются (`_ = s.cache.Invalidate…`). | Не игнорировать err invalidate; при demote system-role — явная инвалидация; рассмотреть shorter TTL для admin mask / version stamp. |
| LOW | `ws/handler.go:88-91` | Пустой `Origin` разрешён (небраузерные клиенты). Accepted для CSWSH-модели, но ослабляет defense-in-depth если появится non-browser attacker с украденным token. | `if origin == "" { return true }` + тесты `origin_test.go` EmptyOriginAllowed. | Оставить accepted-risk с документом; опционально deny empty Origin в prod, если все клиенты — браузер. |
| LOW | `rbac/handler.go:187-200` | `UnassignRole`: path `:roleId` не используется — снимается любая custom role пользователя в проекте. Не IDOR, но путает ACL-аудит. | Handler парсит `projectId`/`userId`, вызывает `ClearCustomRole` без `roleId`. | Либо проверять, что текущая custom role == path roleId, либо убрать roleId из маршрута. |
| INFO | `rbac/handler.go:22-27` | Fiber v3 order корректен на всех 6 routes (нет FA-RBAC-FIBER3 regression в модуле). | Сигнатура `roles.Post("/", h.CreateRole, RequirePermission(...))` — middleware после handler. | Не менять порядок; lint/grep-guard на `RequirePermission(...), h.` в новых routes. |
| INFO | `ws/authorizer.go:46-80` + `ws/handler.go:232-246` + `cmd/server/wiring_ws.go:37-65` | NF-02 закрыт: membership checks для `channel:`/`call:`/`project:`; unknown prefix deny; instant call без ChannelID fail-closed. | `CanSubscribe` до `hub.Subscribe`; wiring: `IsChannelMember` / call+channel / `FindMember`. | Сохранить; regression-тесты authorizer уже в `testframework/ws/`. |
| INFO | `ws/handler.go:33-111` | CheckOrigin синхронизирован с `CORS_ORIGINS`; `DEV_MODE` allow-all только с warn; prod без списка — reject. | `checkOrigin` + `upgrader.CheckOrigin`; `origin_test.go` AllowedList / DevMode / ProductionRejectsWhenUnset. | Следить, чтобы prod compose/K3s не ставил `DEV_MODE=true` (вне scope модуля, devops). |

## Вердикт

- **CRITICAL открытых по NF-02 / CheckOrigin / Fiber v3 order в этих пакетах:** 0  
- **HIGH:** 3 (cross-project role IDOR update/delete/assign) — **блокер** для claim «RBAC roles API безопасен»  
- **NF-02 / CheckOrigin:** verified controls **PASS** (с MEDIUM footgun nil-authz и JWT-in-query)  
- **verified:** `false` — из-за HIGH IDOR на custom roles API  

`verified: false` — 2026-07-23 00:56 UTC+3

## Handoff

| Owner | Действие |
|-------|----------|
| `be1`/`be2` | Фикс project-scope на Update/Delete/Assign custom roles + denial tests |
| `be-ver` | Re-audit после фикса HIGH |
| `qa-auto1` | TC/hubtest: cross-project roleId на A с manage только на B |
| Не трогать | WS Authorizer/CheckOrigin — оставлять fail-closed паттерн |
