# W2 — project / dashboard (RBAC · layering · IDOR)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **be-ver** · readonly scan · 2026-07-23  
> Scope: `backend/internal/project/`, `backend/internal/dashboard/`  
> Код не менялся. Регрессия hubtest в этом прогоне не запускалась (scan-only волна).

## Верификация (BE): project + dashboard

- Воспроизведено (по коду): любой JWT-пользователь org → CRUD/members/invites чужого проекта по UUID; `GET /api/dashboard/project/:projectId` без membership.
- Layering: OK — `*ent.Client` только в `project/repository.go:96`, `dashboard/repository.go:48`; service/handler без прямого Ent-клиента. Частичный FAIL: `dashboard/handler.go:107` — handler читает `h.service.repo` напрямую.
- Security-инварианты: **FAIL** — нет `RequirePermission` / authorize-до-мутации на project CRUD/members/invites/teams; dashboard project stats без membership (класс NF-03 / IDOR).
- Concurrency: N/A (WS/NATS path не в фокусе); `JoinByInvite` + `TryConsumeInvite` — OK (HI-06).
- Регрессия: не прогонялась (repo-scan). Тесты домена не покрывают foreign-project 403 (project); dashboard — только admin 403 для member.
- Вердикт: **verified: false** — 2026-07-23

---

## Summary

| Severity | Count |
|----------|------:|
| CRITICAL | 6 |
| HIGH     | 4 |
| MEDIUM   | 4 |
| LOW      | 2 |

**OK (не findings):** `ListProjects` / `ListTrash` фильтруют по membership (`repository.go:117-143`); personal dashboard scoped по `userID`; layouts keyed по `userID`; `SaveProjectAsTemplate` проверяет `FindMember` (`templates.go:75-80`); admin dashboard отвергает non-admin (`handler.go:111-113` + тест).

---

## Findings

| Sev | Location | Evidence | Fix |
|-----|----------|----------|-----|
| **CRITICAL** | `project/handler.go:23-57` · `wiring_core.go:28` | `RegisterRoutes(api, authMw)` — только JWT. Комментарий про `requirePerm` не реализован; wiring: `projectHandler.RegisterRoutes(api, requireAuth)` без `rbacSvc`. Ни один маршрут не монтирует `rbac.RequirePermission` (в отличие от tracker/tms). | Прокинуть `*rbac.RBACService`; на маршруты с `:projectId` повесить `RequirePermission(PermProjectView/Edit/Delete/ManageMembers/…)` (Fiber v3: middleware **после** handler в сигнатуре). |
| **CRITICAL** | `project/service.go:86-102` · `handler.go:128-144` | `GetProject` → `FindByID` без membership. Не-member получает name/key/description/status/member_count; `myRole==""` не блокирует ответ. | До ответа: `FindMember` / `GetEffectivePermissions`; не-member → 403/404. |
| **CRITICAL** | `project/service.go:129-187,258-267` · `handler.go:160-276` | `UpdateProject` / `Archive*` / `DeleteProject` / `RestoreProject` не принимают `userID`, не вызывают authorize. Любой auth → soft-delete/archive чужого проекта. | Передавать actor; до мутации `PermProjectEdit` / `PermProjectDelete`; deny без membership. |
| **CRITICAL** | `project/service.go:306-397` · `handler.go:312-430` | Members API без authz. `ListMembers` утекает email/name; `AddMember`/`UpdateMemberRole` — произвольная роль вкл. `owner` (privilege escalation); `RemoveMember` — удаление чужих. | `PermProjectManageMembers` (+ view для list); запрет self-escalate в owner без права. |
| **CRITICAL** | `project/invite.go:28-41` · `handler.go:445-516` · `service.go:524-525` | `CreateInvite`/`ListInvites` без membership/`PermInviteCreate`. `RevokeInvite(inviteID)` **игнорирует** `:projectId` из URL → revoke любого invite по UUID. | RBAC на create/list; revoke: проверить `invite.project_id == path projectId` + manage permission. |
| **CRITICAL** | `dashboard/handler.go:70-86` · `service.go:175-309` | `Project`: `_, err := getUserID` — userID отброшен; `GetProjectStats(projectID)` без membership. Утечка tasks-by-status, bugs, overdue, activity, test trends. Redis cache `dashboard:project:{id}` усиливает чтение. | Membership/`PermProjectView` до stats; кэш не меняет authz. Негативный тест foreign project → 403. |
| **HIGH** | `project/handler.go:550-680` · `service.go:444-497` · `repository.go:558+` | Teams: только auth. `ListTeams` — все команды org; `AddTeamToProject` bulk-добавляет members в любой project без права. | Org-admin / `PermProjectManageMembers` на project; scoping list; deny без права. |
| **HIGH** | `project/templates.go:101-115` · `handler.go:283-294` | `DeleteTemplate` — любой auth удаляет любой **custom** template (нет ownership / creator check). System — OK forbid. | Удалять только creator или admin; иначе 403. |
| **HIGH** | `dashboard/handler.go:22-32` · `wiring_modules.go:26` | `rbacSvc` в `RegisterRoutes` **не используется**. Project/admin routes без `RequirePermission` / `RequireAdmin`. | Использовать `RequirePermission` на project path; admin → `rbac.RequireAdmin`. |
| **HIGH** | Тесты: `testframework/project/*`, `dashboard/handler_test.go` | Нет denial-suite foreign project (аналог `bases/rbac_test.go`). Dashboard: admin 403 есть; **нет** теста IDOR на `/dashboard/project/:id`. | Добавить RBAC denial suite: non-member GET/PUT/DELETE/members/invites/stats → 403. |
| **MEDIUM** | `dashboard/handler.go:106-107` | Admin: `h.service.repo.GetUser` — handler лезет в repository через service field (layering). | `service.IsAdmin(ctx, userID)` или middleware `RequireAdmin`. |
| **MEDIUM** | `dashboard/handler.go:111` · schema `user.go` SystemRole | Проверка только `admin`. Роль `founder` (TZ: один на org) → 403 на admin dashboard. | Разрешить `founder` **или** `admin` (как product-инвариант). |
| **MEDIUM** | `dashboard/service.go:458-469` · layout Save | `ApplyTemplate`/`SaveLayout` с чужим `project_id` без membership — пишет user-scoped layout, но создаёт связь на чужой UUID. | Перед save type=project — membership check. |
| **MEDIUM** | `project/service.go:253` | `nats.Publish("project.link_deleted", …)` напрямую, не outbox — риск потери при rollback/частичном fail. | Outbox `PublishInTx` (канон platform/outbox). |
| **LOW** | `project/handler.go:20-22` | Комментарий `requirePerm` без параметра в сигнатуре — drift docs/кода. | Синхронизировать сигнатуру/коммент с wiring. |
| **LOW** | `project/service.go:270-279` | `PermanentDelete` — комментарий «Только admin», в коде проверки нет (сейчас не в RegisterRoutes). | Не экспонировать без RequireAdmin; иначе добавить check. |

---

## Сценарии эксплуатации (минимальные)

1. User A создаёт проект. User B (JWT, не member) вызывает `GET/PUT/DELETE /api/projects/{A.id}`, `GET …/members`, `POST …/invites` → сейчас успех / данные; ожидается 403.
2. User B: `GET /api/dashboard/project/{A.id}` → stats чужого проекта; ожидается 403.
3. User B: `POST /api/teams/{tid}/projects/{A.id}` → bulk membership в чужой проект; ожидается 403.

---

## Рекомендуемый порядок фикса

1. Project routes + service authorize (CRITICAL CRUD/members/invites).  
2. Dashboard project membership (CRITICAL).  
3. Teams + template delete ownership (HIGH).  
4. Denial-suite tests + admin/founder + layering Admin (HIGH/MEDIUM).

---

## Verification block

- Goal: RBAC/IDOR audit project+dashboard — **FAIL** (дыры подтверждены по коду).  
- Layering: в основном OK; dashboard Admin handler→repo — MEDIUM.  
- TZ: membership до мутации/чтения — нарушено.  
- verified: **false**
