# TASK-009 — project routes: навесить RBAC на CRUD/members/invites/teams

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P0** · Статус: Исправлена  
> Источник аудита: W1-A5-001 · TASK-008 стадия B

## [OUTCOME]

Любой аутентифицированный пользователь без членства/права в проекте
получает 403 на mutate-операции project (update/archive/delete/members/
invites/teams), а не выполняет их по одному UUID.

## [ROLE]

- Role: be1
- Responsibility: wiring RBAC на `project.RegisterRoutes` + regression denial
- Handoff to: независимый агент стадии C (мутация)

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - `backend/internal/project/handler.go`
  - `backend/cmd/server/wiring_core.go` (порядок: RegisterRoutes после `rbacSvc`)
  - `backend/testframework/project/**`
  - `.planning/backlog/README.md`, этот файл (статус)
- Do not touch:
  - `POST /api/invites/join` — invite-код сам определяет проект (оставить auth-only)
  - `POST /api/projects` CreateProject — создатель становится owner (auth достаточно)
  - `GET /api/projects` ListProjects — уже scoped по membership в service
  - схемы Ent / миграции
- Constraints:
  - Fiber v3: `RequirePermission` **после** handler в цепочке
  - Паттерн копировать с `tracker.RegisterRoutes`
  - Docker + `POSTGRES_TEST_URL`, `-tags hubtest`
  - Один коммит; описание на русском
- Approval record: N/A (нет миграций/деплоя)

## [REFERENCES]

- Находка: `.planning/audits/total-2026-07/A1-5-project.md` W1-A5-001
- Пермишены: `backend/internal/rbac/permissions.go` (`PermProjectView/Edit/Delete/ManageMembers`)
- Образец: `backend/internal/tracker/handler.go` RegisterRoutes
- Wiring сегодня: `wiring_core.go:28` — `RegisterRoutes(api, requireAuth)` **до** создания `rbacSvc`
- Source precedence: код > отчёт аудита

## Суть дефекта

`RegisterRoutes` принимает только `authMw`. Комментарий упоминает `requirePerm`,
параметра нет. Любой JWT-пользователь с UUID чужого проекта может soft-delete,
archive, менять members/roles (в т.ч. на owner), создавать invites.

## [ACCEPTANCE]

- Done criteria:
  - `RegisterRoutes(api, authMw, rbacSvc *rbac.RBACService)`
  - На маршрутах с `:projectId`: View / Edit / Delete / ManageMembers по смыслу
  - Teams, затрагивающие project: ManageMembers или Edit
  - wiring: RegisterRoutes **после** `rbacSvc := ...`
  - Denial-тест: user без membership → 403 на `DELETE /projects/:id` и
    `POST /:id/members`
- New-behavior checks:
  - Owner/member с правом → 2xx как раньше
  - Stranger → 403 + `{"error","request_id"}`
- **Доказательство мутацией:** убрать middleware с `DeleteProject` → stranger
  получает не-403. Приложить красный/зелёный прогон (стадия C).
- Regression scope: `go test -tags hubtest -race -count=1 ./testframework/project/...`
- Evidence: SHA коммита, вывод теста denial
- Stop when: denial green, wiring с rbacSvc, статус задачи «Исправлена»

## [OUTPUT SCHEMA]

- Коммит `fix(project): ...`
- Тест denial в `testframework/project/`
- Статус в README + этот файл

## [OPERATING RULES]

- Не ослаблять List/Create без явной причины
- Не добавлять `*ent.Client` в handler/service
- Co-authored-by запрещён

## Stage C (W18-C1 mutation) — PASS

- Mutation: strip `RequirePermission` from `DeleteProject` route
- RED: `TestProjectRoutes_RBAC_OutsiderForbidden/delete` — expected 403, got 200
- GREEN: same suite after restore
- Mock-only?: no (Fiber + real RBACService / Ent sqlite)
