# Red Team

**Дата:** 2026-07-26  
**Ветка:** `cursor-grok-4.5`  
**Зона:** A19 — insider / low-privilege (обычный `member`, не admin/founder)  
**Режим:** READ-ONLY (кроме этого отчёта)  
**База:** `.planning/audits/grok-2026-07/01–16.md` + сверка с реальным `.go`

## Assumptions (attacker model)

| Параметр | Значение |
|----------|----------|
| Роль | `SystemRoleMember`, **не** admin/founder |
| Доступ | Валидный Magic Link → JWT (access 15m + refresh) **или** pending-user после регистрации (см. RT-001) |
| Membership | ≥1 проект (для цепочек с RBAC на path); для auth-only дыр membership **не** нужен |
| Знания | UUID проектов/каналов/спринтов/ролей/звонков из UI, WS, утечек, Ctrl+K, `/api/docs`, догадки ключей `PROJ-N` |
| Цель | Чтение чужих данных, мутация чужих ресурсов, privilege escalation внутри org, DoS уведомлений/звонков |
| Вне scope | Admin-only (backup restore, GDPR, ban founder), server RCE, physical access к Redis/PG |
| Инстанс | Self-hosted 1 org; blast radius = вся корпоративная org |

---

## Attack chains

### RT-001 Pending→JWT: вход до approve (CRITICAL)

**Вход:** внешний email + invite token (invite_only) **или** open registration; код из Magic Link email.

**Шаги:**
1. `POST /api/auth/magic-link` `{email, invite_token?}` → код на почту.
2. `POST /api/auth/magic-link/verify` `{email, code}` → `CreateSession` без проверки `status`.
3. Получить `access_token` / `refresh_token` при `user.status = pending`.
4. `Authorization: Bearer …` на любой auth-protected API (CRM, search, imex, messenger…).

**Выход/выгода:** полный API-доступ **до** admin approve (ORG-06 bypass). Комбинируется со всеми RT ниже без ожидания очереди.

**Должно остановить:** после verify — отказ если `status != active`; `RequireAuth` / refresh fail-closed на pending.

**Почему не остановило:**
- `auth/service_login.go:64` — всегда `CreateSession` после verify.
- `auth/middleware.go:128` — `userIsDeactivated` считает только `StatusDeactivated`, pending = «active».
- `auth/repository.go:284-292` — `CreatePendingMember` ставит `StatusPending`, но сессия всё равно выдаётся.

---

### RT-002 Messenger search: чтение всех каналов (CRITICAL)

**Вход:** любой JWT (даже без membership в целевых каналах).

**Шаги:**
1. `GET /api/messenger/search?query=пароль` (или `secret`, `token`, `deploy`) без `channel_id`.
2. При известном чужом `channel_id`: `?query=…&channel_id=<victim-channel-uuid>` — membership **не** проверяется.
3. Пагинация `limit` до 100 → массовый дамп текстов.

**Выход/выгода:** тексты сообщений **всех** каналов org (секреты, PII, обсуждения HR/сделок). Класс NF-03.

**Должно остановить:** user-scope: без `channel_id` — только `IsMember` каналы; с `channel_id` — `IsMember` иначе 403.

**Почему не остановило:**
- `messenger/handler_search.go:24-32` — не передаёт `userID`, не вызывает membership.
- `messenger/service_search.go:11-21` — сразу `repo.SearchMessages`.
- `messenger/mongo_repository.go:282-288` — фильтр только `$text` (+ опционально `channel_id`); **нет** allowlist каналов.

*(Контраст: Send/List сообщений проверяют `IsMember` — search — нет.)*

---

### RT-003 CRM: org-wide PII + mutate + invoice side-effect (CRITICAL)

**Вход:** любой JWT member.

**Шаги:**
1. `GET /api/crm/contacts` → полный список контактов (email/phone/notes).
2. `GET /api/crm/pipeline` → все deals org.
3. `PUT /api/crm/contacts/:id` — правка чужого контакта (нет owner check).
4. `PUT /api/crm/deals/:id` `{stage:"won"}` — чужая сделка → `tryCreateInvoice` (draft invoice).
5. `DELETE /api/crm/contacts/:id` / `DELETE /api/crm/deals/:id` — soft-delete чужих сущностей.

**Выход/выгода:** кража/порча коммерческой PII; фин. шум (чужие invoices); sabotage pipeline.

**Должно остановить:** `authorizeContact` / `authorizeDeal` (owner \|\| admin); list scoped + pagination.

**Почему не остановило:**
- `crm/handler.go:22-24,59-68,75+` — только `authMw`.
- `crm/service.go:80-123` — `UpdateContact`/`DeleteContact` без `userID`.
- `crm/service.go:172-214` — `UpdateDeal`/`DeleteDeal` без owner; `stage=won` → invoice (`:200-201`).
- `crm/service.go:43-53,125-147` — List/pipeline = org dump.

---

### RT-004 Cross-project role IDOR → privilege escalation (CRITICAL)

**Вход:** member с `PermProjectManageRoles` на **своём** проекте A; UUID роли проекта B (из UI/сетевых логов/коллеги).

**Шаги:**
1. `PUT /api/projects/{A}/roles/{roleB}` — тело с полным набором permissions (в т.ч. manage_roles, task_edit…). Middleware проверяет manage_roles **только на A**.
2. `POST /api/projects/{A}/roles/{roleB}/assign/{victimUserId}` — если victim ∈ A: назначается роль **B** (чужие permission bits) на membership в A.
3. Альтернатива: `DELETE /api/projects/{A}/roles/{roleB}` — снос кастомной роли чужого проекта (system role на Delete защищён, на Update — нет).

**Выход/выгода:** эскалация прав в A через чужую роль; DoS ролевой модели проекта B; подготовка к дальнейшему IDOR.

**Должно остановить:** `role.project_id == :projectId`; Update — отказ для `IsSystem`; assign только роль того же проекта.

**Почему не остановило:**
- `rbac/handler.go:99-114,154-168` — `projectId` из path не сверяется с `roleId`.
- `rbac/custom_roles.go:73-84` — `UpdateCustomRole(roleID)` без project scope.
- `rbac/custom_roles.go:126-132` — `AssignCustomRole` не проверяет `role ∈ project`.

---

### RT-005 Sprint Complete: кража задач чужого спринта (CRITICAL)

**Вход:** member с правом sprint-complete на проекте A; UUID активного спринта проекта B.

**Шаги:**
1. Разведка: UUID спринта B (доска, URL, export PDF, утечка).
2. `POST /api/projects/{A}/sprints/{sprintB}/complete` (точный path — sprint handler с `:projectId` + `:sprintId`).
3. Service грузит sprint **только по ID** → находит incomplete issues спринта B → `FindOrCreateNextPlanningSprint(ctx, projectID=A, …)` → `MoveIssuesToSprint` в спринт **проекта A**.

**Выход/выгода:** массовый перенос чужих задач в свой проект; завершение чужого активного спринта (sabotage).

**Должно остановить:** `GetSprint` с `Where(ID, projectID)`; mismatch → 404.

**Почему не остановило:**
- `sprint/service.go:107-149` — `GetSprint(ctx, sprintID)` без сверки `sp.ProjectID` с URL `projectID`.
- `sprint/repository.go:78+` — lookup по ID only.
- Перенос явно в `projectID` из URL (`:128` `FindOrCreateNextPlanningSprint(ctx, projectID, sp)`).

---

### RT-006 Imex: экспорт чужого проекта без RBAC (HIGH)

**Вход:** любой JWT; UUID чужого проекта (ListProjects чужих не нужен — UUID из invite/URL/док/коллеги).

**Шаги:**
1. `GET /api/projects/{victimProjectId}/export/issues` → CSV/JSON задач.
2. `GET /api/projects/{victimProjectId}/export/full` → ZIP проекта.
3. `GET /api/projects/{victimProjectId}/sprints/{sprintId}/export/pdf` / test-runs PDF.
4. (Опционально) `POST …/import/issues` — мусорный import в чужой проект (CSV mappings баг A5-001 ломает import, но preview/authz gap остаётся).

**Выход/выгода:** полный дамп чужого backlog/тестов без membership.

**Должно остановить:** `rbac.RequirePermission(PermProjectView / PermTaskCreate)` как у tracker на том же prefix.

**Почему не остановило:**
- `imex/handler.go:26-35` — только `authMw`, **нет** `RequirePermission`.
- PDF path (`:34-35`) — тот же gap + `pdf.go` не сверяет run/sprint с `projectId` (A5-010).

---

### RT-007 Tracker issueKey IDOR: чтение/правка чужих задач (HIGH)

**Вход:** member проекта A с `PermTaskView`/`PermTaskEdit`; знание ключа `B-42` (Ctrl+K, уведомление, угадывание `KEY-N`).

**Шаги:**
1. `GET /api/projects/{A}/issues/B-42` — RBAC на A, lookup по ключу **без** project scope → задача проекта B.
2. `PUT /api/projects/{A}/issues/B-42` — правка title/assignee/status чужой задачи.
3. Comments/time/links через тот же `FindIssueByKey` — расширение.

**Выход/выгода:** чтение/изменение чужого backlog при membership в любом другом проекте.

**Должно остановить:** `FindIssueByKey(ctx, projectID, key)` или post-check `issue.ProjectID == path projectId` → 404.

**Почему не остановило:**
- `tracker/handler_issues.go:99-110,128-143` — `projectId` не передаётся в Get/Update.
- `tracker/service_issues.go:123-124,191+` — `FindIssueByKey(ctx, issueKey)`.
- `tracker/repo_issues.go:140-150` — `Where(IssueKeyEQ)` only.

---

### RT-008 Calls: CreateCall в чужой канал + chat inject/exfil (HIGH)

**Вход:** JWT; UUID чужого messenger-канала / активного `callId`.

**Шаги:**
1. `POST /api/calls` (create in channel) с чужим `channelId` — **без** `IsChannelMember` → LiveKit room + NATS `calls.channel.*.started` у жертв.
2. Зная `callId`: send/list call chat — нет participant check → читать/писать чат звонка; NATS → WS `call:{id}` у легитимных участников.
3. `POST …/recording/start` на чужой active call — egress в S3 (DoS/privacy).
4. (Bonus) `RaiseHand` / `GetActiveCall` — метаданные/LiveKit API без участия.

**Выход/выгода:** социальная инженерия («ложный звонок»), eavesdrop/chat spam, запись чужого разговора.

**Должно остановить:** `IsChannelMember` / participant до Create/Chat/Recording/RaiseHand.

**Почему не остановило:**
- `calls/service.go:49-65` — `CreateCall` без membership.
- `calls/service_chat.go:33-79` — только статус active; нет participant/channel authz.
- `calls/recording.go:89-97` — StartRecording без membership (только call exists + active).

---

### RT-009 Push subscription hijack / unsubscribe DoS (HIGH)

**Вход:** JWT; знание `endpoint` чужой WebPush-подписки (XSS, shared device, лог, физический доступ к браузеру жертвы).

**Шаги:**
1. `POST /api/notifications/push/subscribe` с чужим `endpoint` + своими ключами → `SetUserID(attacker)` на существующей строке.
2. Или `DELETE` unsubscribe по endpoint **без** фильтра `UserID` → удаление чужой подписки.

**Выход/выгода:** жертва перестаёт получать push (security alerts, mentions, magic-adjacent ops signals).

**Должно остановить:** Update/Delete `Where(Endpoint AND UserID=caller)`; чужой endpoint → 404 без rebind.

**Почему не остановило:**
- `notification/push.go:230-237` — existing → `SetUserID(userID)` (hijack).
- `notification/push.go:252-261` — `RemoveSubscription` только по `Endpoint`.

---

### RT-010 Calendar: утечка HR leave всей org (HIGH)

**Вход:** любой JWT.

**Шаги:**
1. `GET /api/calendar/events?from=…&to=…` (или iCal export того же сервиса).
2. В ответе — **все** approved leave: имя/email + тип (sick/vacation/personal).

**Выход/выгода:** HR-privacy breach (больничные/отгулы) всем сотрудникам, не только HR/admin.

**Должно остановить:** leave scoped к своим + members общих проектов / admin policy.

**Почему не остановило:**
- `calendar/service.go:77-83` — `ListApprovedLeaveInRange` без project/user filter.
- `calendar/service.go:208-238` — `leaveEvent` кладёт DisplayName/Email + тип в title.

---

### RT-011 Dashboard project stats IDOR (HIGH)

**Вход:** любой JWT; UUID чужого проекта.

**Шаги:**
1. `GET /api/dashboard/project/{victimProjectId}`.
2. `userID` читается и **отбрасывается** (`_`); stats без membership.

**Выход/выгода:** метрики чужого проекта (tasks/bugs/activity) — разведка перед RT-005/006/007.

**Должно остановить:** membership + `PermProjectView` до `GetProjectStats`.

**Почему не остановило:**
- `dashboard/handler.go:71-81` — `_, err := getUserID`; далее stats по голому `projectID`.

---

### RT-012 GitHub CreateBranch: чужой PAT без project RBAC (HIGH)

**Вход:** любой JWT; `issueKey` + `integrationId` (UUID интеграции с PAT).

**Шаги:**
1. `POST /api/issues/{issueKey}/github/create-branch` `{integration_id, branch_name?}` — маршрут только `authMw`.
2. Service: `FindIssueByKey` + `GetIntegration` **без** сверки project issue ↔ integration; `apiClient(integration.AccessToken)` → создаёт ветку в чужом GitHub repo.

**Выход/выгода:** запись в чужой GitHub через украденный/известный integration UUID; шум в CI; подготовка к supply-chain noise.

**Должно остановить:** RBAC на проект issue; `integration.ProjectID == issue.ProjectID`.

**Почему не остановило:**
- `github/handler.go:51` — `authMw` only.
- `github/service.go:310-334` — нет project binding check перед использованием PAT.

---

## Неудавшиеся векторы (что проверил и не взломал)

| Вектор | Почему не взломал (как member) |
|--------|--------------------------------|
| WS subscribe чужой `channel:` / `project:` | `Authorizer.CanSubscribe` + membership; unknown prefix fail-closed (`ws/authorizer.go`, `handler.go`) |
| LiveKit webhook forge | Mandatory `sha256` claim + HMAC JWT (`calls/handler.go`) |
| Messenger Send/List/Edit без membership | `IsMember` на message path |
| Bases CRUD чужой базы | `authorizeBase` owner-scope (Wave 1 fix) |
| Global `/api/search` messages shard | Scoped через `channel_members` (в отличие от messenger search!) |
| Storage download чужого файла по ID | RBAC через `FileProjectID` из метаданных |
| SSRF via OpenGraph / WebPush / GitLab URL | `ValidateOutboundURL` + `SafeHTTPClient` / SafeDial |
| Magic Link brute-force | Redis RL 3/10m (+ tighter verify) |
| Password / Telegram login | Роутов нет; ADR-AUTH-01 |
| Fiber v3 RBAC order bypass на tracker/tms | Middleware **после** handler в аргументах — order OK |
| `/admin/backups`, GDPR, BanUser, accounting pay | Admin/founder gate — вне модели атакующего |
| Dev routes в prod | `-tags release` no-op |
| CORS `*` / CSRF cookie | Bearer-only; CORS allowlist fail-start |

---

## Оценка эксплуатабельности: 8.5/10

**Обоснование.** Для **обычного member** внутри org цепочки RT-001…007 — практически «copy-paste HTTP»: нет нужды в admin, RCE или утечке серверных секретов. Самый короткий путь к максимальному урону:

`pending/JWT (RT-001)` → `messenger search (RT-002)` + `CRM dump (RT-003)` + `imex export (RT-006)`  

При membership + manage_roles / sprint rights добавляются **эскалация** (RT-004) и **кража задач** (RT-005). Контроли периметра (SSRF, webhook LiveKit, WS room authz, Magic Link RL, bases) держатся — поэтому не 10/10: это не unauthenticated internet RCE, а **катастрофический Broken Access Control (OWASP A01)** для insider / compromised low-priv account.

| Метрика | Значение |
|---------|----------|
| Цепочек (RT-*) | **12** |
| CRITICAL | 5 (001–005) |
| HIGH | 7 (006–012) |
| Score | **8.5/10** |

`verified: true` — 2026-07-26 (readonly red-team; product code не менялся).

### Handoff (fix owners)

| Priority | RT | Owner |
|----------|----|-------|
| P0 | 001 | be1 — auth pending gate |
| P0 | 002 | be1 — messenger search scope |
| P0 | 003 | be2 — CRM authorize* |
| P0 | 004 | be1 — rbac role↔project |
| P0 | 005 | be2 — sprint project scope |
| P1 | 006–012 | be1/be2 по модулям (imex RBAC, tracker key scope, calls authz, push user scope, calendar leave, dashboard membership, github CreateBranch RBAC) |

После фиксов — повторный red-team pass (цепочки RT-001…012 должны стать «Неудавшиеся»).
