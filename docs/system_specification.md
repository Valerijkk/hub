# Hub — Системная спецификация

> Единая база знаний. Обновляется SA при изменении требований.
> Детали: `docs/TECHNICAL.md`, API: `docs/wiki/API.md`, Swagger: `backend/docs/`

---

## Функциональные требования (сводка)

Полный список: `.planning/PROJECT.md`, `.planning/REQUIREMENTS.md`

| Модуль | ID prefix | Ключевые возможности |
|--------|-----------|---------------------|
| Auth | AUTH | Magic Link login, org/founder, invite-only, JWT, sessions, TG **notify bind only** |
| Project | PROJ | CRUD, RBAC, invites, teams, templates, archive |
| Tracker | TASK | 9 типов, workflow, links, comments, time tracking, automation rules |
| Kanban | KANB | Columns, WIP, swimlanes, filters |
| Scrum | SCRM | Sprints, burndown, velocity, planning poker |
| Messenger | MSG | Channels, DMs, threads, reactions, voice messages, issue links |
| TMS | TMS | Cases (classic/BDD), runs, environments, reports, traceability |
| Allure | ALLU | CI parser, matching, notifications, JUnit webhook |
| Calls | CALL | LiveKit voice + screen share, recording, scheduled |
| Dashboard | DASH | Personal/project/admin widgets |
| Notification | NOTF | In-app, Telegram bind, Web Push, quiet hours |
| Search | SRCH | Ctrl+K, Meilisearch indexer, JQL, saved filters |
| GitHub/GitLab | GHUB/GLAB | Webhooks, PR/MR linking, AI PR review |
| API | API | REST **~252** OpenAPI paths / ~286 route inventory (+1 WS); API keys, outbound webhooks |
| Import/Export | IMEX | CSV, JSON, PDF, ZIP backup, Jira import |
| Audit | AUDIT | Full audit log, trash 30d, GDPR, alerts |
| AI | AI | Multi-provider gateway, @hub-ai bot, JQL assist, usage audit |
| Intranet | INTR | News, directory, kudos (+leaderboard spec), random coffee (+auto-DM spec), poll-in-post spec, onboarding spec (FR-INTR) |
| Calendar | CAL | Unified events (calls, sprints, leave); spec: собственные события, ресурсы+booking, recurrence, iCal feeds, availability (FR-CAL) |
| CRM | CRM | Contacts, deals (lite) |
| HR | HR | Leave requests |
| Accounting | ACCT | Ledger, journal entries, invoices |
| Bases | BASE | No-code DB (HubBase/Table/Record) |
| Marketplace | MKT | Internal listings |
| Mail | MAIL | hub-mail admin, transactional settings |
| Status | STAT | Public status page |
| 1C | ONEC | Optional connector (`integrations/onec`) |
| Wiki | WIKI | **spec (Волны 5/8):** spaces + иерархия страниц, space-RBAC + restrict, версии/rollback, issue-links, user-scoped поиск (FR-WIKI) |

## Нефункциональные требования

| ID | Требование | Значение |
|----|------------|----------|
| NFR-01 | Response time API (p95) | < 500ms для CRUD |
| NFR-02 | Concurrent WS connections | 50+ per instance |
| NFR-03 | Call participants | до 50 |
| NFR-04 | Max file upload (backend) | 25 MB BodyLimit |
| NFR-05 | Auth rate limit | 3 попытки / 10 мин |
| NFR-06 | API rate limit | 100 req/sec |
| NFR-07 | Message rate limit | 60/min |
| NFR-08 | File upload rate limit | 10/min |
| NFR-09 | JWT access TTL | 15 мин |
| NFR-10 | Refresh token TTL | 30 дней, instant revoke |
| NFR-11 | Soft delete retention | 30 дней |
| NFR-12 | UTC in DB | Всегда, TZ на фронте |
| NFR-13 | Error format | Wire JSON: `{"error": "...", "request_id": "uuid"}` (`platform/errors`, `middleware.ErrorHandler`). FE `ApiError` maps status→`code`, body.error→`message` — не отдельный wire-контракт |
| NFR-14 | Pagination tasks (list-issues) | cursor + limit (max 100); `next_cursor` / `has_more` / `total` — `ADR-PAGINATION-CURSOR-ISSUES-2026` |
| NFR-15 | Pagination chat | cursor (infinite scroll) |
| NFR-16 | Accessibility | WCAG 2.1 AA |
| NFR-17 | Lighthouse score | ≥ 90 (цель v2.5) |
| NFR-18 | WebSocket room authz | Subscribe только при membership/RBAC (NF-02) |
| NFR-19 | Search user-scope | Фильтрация по project/channel membership (NF-03) |
| NFR-20 | Webhook signature | LiveKit/GitHub: body hash обязателен, skip запрещён (SEC-02) |
| NFR-21 | Outbound HTTP safety | `SafeHTTPClient` + `ValidateOutboundURL` для user URLs (SSRF) |
| NFR-22 | Security headers | CSP, X-Frame-Options, nosniff — ✅ `middleware.SecurityHeaders()`; **HSTS** — только Caddy edge (`deployments/Caddyfile.prod`, A2W4-1-001) |
| NFR-23 | Refresh token storage | Hash в Redis, не plaintext — ✅ |
| NFR-24 | Optimistic locking | `version` int на каждой изменяемой Ent |
| NFR-25 | Primary keys | UUID v4 only — string IDs запрещены |
| NFR-26 | Pub/sub transport | NATS JetStream only — Redis pub/sub запрещён |
| NFR-27 | File upload path | Presigned PUT в SeaweedFS; обязательный `size` + ContentLength bind ≤25 MB (ADR-STORAGE-NF09-SIZE-2026); backend buffer max 25 MB |
| NFR-28 | Tenancy | Single-tenant (1 instance = 1 team) |
| NFR-29 | Test coverage gate | ≥50% overall ✅ **83.0%** (2026-06-22 Docker baseline); ≥80% security-critical ✅ **6/6** |
| NFR-30 | Frontend test gate | Vitest **179** files; Playwright **20** E2E specs ✅ |

## Модель данных (ER — ключевые сущности)

### PostgreSQL (Ent)

```
User ──┬── Session
       ├── Organization (1 per instance)
       ├── ProjectMember ── Project ──┬── Issue ──┬── IssueLink
       │                              │           ├── Comment
       │                              │           ├── IssueMessageLink
       │                              │           └── CustomFieldValue
       │                              ├── WorkflowStatus / AutomationRule
       │                              ├── Board
       │                              ├── TestCase ── TestRun ── TestResult
       │                              └── Channel (ref)
       ├── AIProvider / AIUserPreference / AIUsageDaily
       ├── NewsPost / Kudo / CoffeeEnrollment
       ├── CRMContact / CRMDeal / LeaveRequest
       ├── HubBase ── HubTable ── HubColumn ── HubRecord
       ├── LedgerAccount / JournalEntry / Invoice
       └── NotificationPreference
```

**Ent schemas:** **82** сущности в `backend/ent/schema/` (2026-06-21).

**RBAC / прочие PG-сущности:** Role (founder/admin/member), ProjectRole, Team, InviteLink, AuditLog, TrashItem, Label, MarketplaceListing и др.

**Общие поля изменяемых Ent:**
- `id`: UUID v4 (не string)
- `created_at`, `updated_at`: UTC timestamp
- `version`: int, optimistic lock
- Soft delete: `deleted_at` или TrashItem

### MongoDB

```
messages: { channel_id, author_id, content, thread_id, reactions, ... }
allure_reports: { run_id, suites, tests, attachments, ... }
call_chat: { call_id, messages[] }  // ephemeral
```

### Redis

```
refresh:<token_hash> → session_id
magic_link:<email> → code (TTL)
rbac:<user_id>:<project_id> → permission bits
presence:<user_id> → status
rate_limit:<key> → counter
```

## API Контракты

**Base URL:** `/api`
**Auth:** `Authorization: Bearer <jwt>` или `X-API-Key: <key>`
**Content-Type:** `application/json`

### Примеры endpoints

| Method | Path | Описание |
|--------|------|----------|
| POST | `/auth/magic-link` | Отправить Magic Link код на email |
| POST | `/auth/magic-link/verify` | Verify code → tokens |
| POST | `/auth/refresh` | Refresh access token |
| GET | `/projects` | List projects |
| POST | `/projects/:id/issues` | Create issue |
| GET | `/projects/:id/issues?cursor=…&limit=50` | Cursor-paginated issues (`sprint_id=none` = backlog); NFR-14 / ADR-PAGINATION-CURSOR-ISSUES |
| GET | `/channels/:id/messages?cursor=...` | Cursor pagination |
| POST | `/tms/runs/:id/results` | Submit test result |
| POST | `/calls/:id/join` | Get LiveKit token |
| GET | `/search?q=...&types=issue,message` | Global search |
| WS | `/ws` | WebSocket upgrade |

Полная спецификация: `docs/wiki/API.md`, Swagger annotations в handlers.

### Endpoint inventory (2026-06-21)

| Категория | Count | Примечание |
|-----------|-------|------------|
| Documented in `docs/wiki/API.md` | **287+** | v4 + AI routes |
| Handler `RegisterRoutes` (code) | **283** | 21+ domain handlers |
| Infra (`/docs`, `/live`, `/ready`, `/health`, `/metrics`, `/status`, `/status/details`) | **+7** | probes + status; `/docs` — system admin (W7-A5-001); `/status` aggregate, `/status/details` auth (W7-A5-002) |
| **Total REST `/api/*`** | **~252** OpenAPI paths (`backend/docs/swagger.json`, measured 2026-07-28); handler `RegisterRoutes` inventory historically ~286 — методика: swagger paths vs route registrations | Wiki / Swagger |
| WebSocket `GET /ws` | **+1** | Вне префикса `/api` (`wiring_ws.go`) |
| Dev-only `/api/dev/*` | **3** | `DEV_MODE`, не в `-tags release` |

**По модулям (handler routes, выборочно):** auth 11 · project 24 · tracker 48 · messenger 44 · tms 52 · calls 24 · search 9 · notification 18 · ai 10+ · intranet 19 · bases 13 · accounting 13 · crm 10 · hr 8 · calendar 4 · marketplace 12 · github 13 · gitlab 11 · sprint 16 · dashboard 9 · audit 13 · apigateway 10 · backup 6 · imex 8 · storage 6 · rbac 8 · status 2 (`/status`, `/status/details`) · mail (admin) · onec 5.

### WebSocket Protocol

Handshake (ADR-WS-JWT-HANDSHAKE-2026): `POST /auth/ws-ticket` →
`GET /ws` with `Sec-WebSocket-Protocol: hub.ws.ticket.<ticket>`.
Query `?token=` is rejected (401). Room subscribe still requires membership/RBAC (NFR-18).

```json
{ "type": "subscribe", "room": "project:<uuid>" }
{ "type": "event", "room": "...", "payload": { ... } }
{ "type": "ping" } → { "type": "pong" }
```

Rooms: `project:<id>`, `channel:<id>`, `user:<id>`, `admin:<id>`

## DFD — создание задачи

```
[User] → POST /issues → [Auth Middleware] → [RBAC Check]
    → [IssueService.Create] → [IssueRepository] → [PostgreSQL]
    → [Outbox.Publish] → [NATS: hub.issues.created]
    → [WS Bridge] → [Subscribed Clients]
    → [NotificationService] → [Telegram Bot]
```

## BPMN — Magic Link auth

1. User enters email → rate limit check
2. Generate 6-digit code → Redis TTL 10 min
3. Resend sends email
4. User submits code → constant-time compare
5. Create session → JWT + refresh in Redis
6. Audit log: login event

## Маппинг User Stories → модули

| User Story | Модули | API prefix |
|------------|--------|------------|
| US-M01 | auth | `/auth/*` |
| US-M02 | project, tracker, messenger, dashboard | `/projects/*` |
| US-M03 | tracker, rbac | `/projects/:id/issues/*` |
| US-M04 | messenger, ws | `/channels/*`, `/ws` |
| US-M05 | tms, imex | `/tms/*` |
| US-M06 | calls | `/calls/*` |
| US-M07 | search | `/search` |

## Спецификации Волны 2 — v4 parity + Wiki (2026-07-03)

> Технические спецификации по верифицированным US Волны 1 (`docs/vision_and_scope.md`,
> разделы «Волна 1 аналитика»). Границы scope: `.planning/decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md`
> §2 (wiki RBAC) и §4 (intranet/calendar). Статус всех FR ниже: **spec** (реализация — Волны 4/5/8).
> Нумерация NFR этой волны: NFR-31…NFR-39 (INTR 31–32, CAL 33–36, WIKI 37–39).

### Intranet v2 (FR-INTR)

Доводка модуля `intranet` до паритета: опрос-в-посте, onboarding-чеклисты, постмодерация,
kudos-лидерборд, авто-DM для Random Coffee. Базовые сущности v4 (NewsPost, Kudo,
CoffeeEnrollment/CoffeeRound/CoffeePair) переиспользуются.

#### Функциональные требования

| ID | Source | Поведение | NFR link |
|----|--------|-----------|----------|
| FR-INTR-01 | US-INTR-01 | Автор поста прикрепляет к посту ровно один опрос: вопрос (1..500) + 2–10 вариантов. <2 вариантов → 400. Опрос существует только внутри поста (GZ-V4-01) | NFR-13 |
| FR-INTR-02 | US-INTR-01 | Голос: ровно один вариант на пользователя. Повторный голос → 422, счётчики не меняются. После голосования пользователь видит распределение в процентах | NFR-31 |
| FR-INTR-03 | US-INTR-01 | Редактирование вариантов после первого голоса → 422. Автор опроса видит только агрегаты (не список проголосовавших); голос анонимен для коллег | NFR-31 |
| FR-INTR-04 | US-INTR-02 | Admin назначает пользователю системный onboarding-шаблон (dev / non-dev, GZ-V4-02). Назначение — **копия** items шаблона; последующая правка шаблона не меняет назначенные чеклисты | NFR-13 |
| FR-INTR-05 | US-INTR-02 | Отмечать пункты может только владелец чеклиста (чужой → 403). Прогресс (N из M, %) виден владельцу и Admin (список по всем новичкам) | NFR-24 |
| FR-INTR-06 | US-INTR-02 | Повторное назначение шаблона тому же пользователю сбрасывает прогресс только при явном `confirm=true` (без него → 409). Деактивация пользователя замораживает чеклист (не удаляет) | NFR-24 |
| FR-INTR-07 | US-INTR-03 | Постмодерация (GZ-V4-03): Admin удаляет любой пост, автор — свой. Member на чужой → 403. Удаление → корзина 30 дней + audit log (actor, post_id); каскадно скрывает опрос и голоса | NFR-11 |
| FR-INTR-08 | US-INTR-03 | Восстановление поста из корзины возвращает пост вместе с опросом и счётчиками голосов | NFR-11 |
| FR-INTR-09 | US-INTR-04 | Лидерборд kudos по **полученным** за период `month`/`quarter`/`all`: имя, аватар, счёт. Пустой период → 200 + пустой список. Равенство очков — стабильная сортировка по дате первого kudos. Деактивированные получатели остаются в агрегате | NFR-32 |
| FR-INTR-10 | US-INTR-05 | Еженедельный cron-матчинг opt-in участников: пары, при нечётном N — одна тройка. Anti-repeat по истории CoffeePair (повтор только если иных комбинаций нет). 1 участник → без матчей + уведомление «пары нет» | NFR-32 |
| FR-INTR-11 | US-INTR-05 | После матчинга каждый участник пары/тройки получает in-app notification о напарнике(ах) (`CoffeeNotifier`, ADR-INTRANET-U05 notify-only). Auto-DM в messenger **не** создаётся до отдельного ADR. Выключивший участие до запуска не матчится | NFR-26 |

#### Новые NFR

| ID | Требование | Значение |
|----|------------|----------|
| NFR-31 | Уникальность голоса | Уникальный индекс `(poll_id, user_id)` в news_poll_vote; конкурентный повторный голос → 422, не дубль |
| NFR-32 | Агрегаты intranet | Лидерборд и матчинг: p95 < 500ms при 1k пользователей / 50k kudos (индекс по recipient+created_at); cron-матчинг идемпотентен (повторный запуск раунда не создаёт дублей пар) |

#### ERD-дельта (PostgreSQL / Ent)

Именование `news_poll*` — чтобы не конфликтовать с существующими messenger-сущностями
Poll/PollOption/PollVote (channel-scoped).

```
NewsPost 1──0..1 NewsPoll ──┬── NewsPollOption (2..10)
                            └── NewsPollVote (unique poll_id+user_id)
OnboardingTemplate 1──N OnboardingProgress N──1 User
Kudo (существует) → лидерборд = агрегатный запрос (view), без новой сущности
CoffeePair (существует) + dm_channel_id
```

| Сущность | Поля (помимо инвариантов id UUID v4 / version / created_at,updated_at UTC / soft delete) | Индексы |
|----------|--------------------------------------------------------------------------------------|---------|
| `news_poll` | `news_post_id` FK unique (0..1 на пост), `question` string(500), `closed` bool | `(news_post_id)` unique |
| `news_poll_option` | `poll_id` FK, `text` string(200), `position` int, `votes_count` int (денормализация) | `(poll_id, position)` |
| `news_poll_vote` | `poll_id` FK, `option_id` FK, `user_id` FK | `(poll_id, user_id)` **unique** (NFR-31) |
| `onboarding_template` | `name` string, `kind` enum{dev, non_dev}, `items` JSON `[{id, title}]`, `active` bool | `(kind)` |
| `onboarding_progress` | `user_id` FK, `template_id` FK, `items` JSON-копия `[{id, title, checked, checked_at}]`, `assigned_by` FK, `frozen` bool | `(user_id)` unique-active |
| `coffee_pair` (изм.) | + `dm_channel_id` UUID nullable (FK → Channel) | — |

Инварианты соблюдены: голоса удаляются каскадно с постом только через корзину (FR-INTR-07/08);
`onboarding_progress.items` — копия (FR-INTR-04); лидерборд не хранит state.

#### API-контракты

Все ошибки — `{"error","request_id"}` (NFR-13); rate limit — общий API 100/с (NFR-06) → 429.

```
### POST /api/intranet/news  (расширение существующего)
Auth: Bearer; RBAC: news.create (admin/founder)
Request: { "title", "body", "pinned"?, "poll"?: { "question": "1..500", "options": ["2..10 строк"] } }
Responses: 201 {news + poll}; 400 (<2 или >10 вариантов, пустой вопрос); 401; 403
Source FR: FR-INTR-01

### GET /api/intranet/news/:newsId/poll
Auth: Bearer
Responses: 200 { "id","question","closed","voted_option_id"|null,
                 "options":[{"id","text","votes_count","percent"}], "total_votes" }
          (percent виден только после собственного голоса или автору поста); 401; 404
          (unpublished draft: 404 для non-author/non-admin — как GetNews; TASK-189 / BE-V4-003)
Source FR: FR-INTR-02, FR-INTR-03

### POST /api/intranet/news/:newsId/poll/vote
Auth: Bearer
Request: { "option_id": "uuid" }
Responses: 200 {агрегаты как в GET}; 400 (option чужого poll); 401; 404
          (в т.ч. draft для non-author); 422 (повторный голос / poll closed); 429
Source FR: FR-INTR-02

### PUT /api/intranet/news/:newsId/poll
Auth: Bearer; RBAC: автор поста или admin
Request: { "question"?, "options"?, "closed"?, "version" }  // version ≥1 обязателен
Responses: 200; 400 (version omit/<1); 401; 403; 404; 409 (version conflict); 422 (правка options после первого голоса)
Source FR: FR-INTR-03

### GET /api/intranet/onboarding/templates
Auth: Bearer; RBAC: admin
Responses: 200 { "items":[{"id","name","kind","items":[...]}] }; 401; 403
Source FR: FR-INTR-04

### POST /api/intranet/onboarding/assignments
Auth: Bearer; RBAC: admin
Request: { "user_id", "template_id", "confirm"?: bool }
Responses: 201 {progress}; 400; 401; 403; 404 (user/template);
           409 (уже назначен, без confirm=true — FR-INTR-06)
Source FR: FR-INTR-04, FR-INTR-06

### GET /api/intranet/onboarding/assignments?offset=0&limit=50
Auth: Bearer; RBAC: admin
Responses: 200 { "items":[{"user":{...},"template_id","done_count","total_count","percent","frozen"}],
                 "total","offset","limit" }; 400 (limit>100); 401; 403
Source FR: FR-INTR-05  · Пагинация: offset/limit (не NFR-14; NFR-14 = cursor list-issues)

### GET /api/intranet/onboarding/me
Auth: Bearer
Responses: 200 {progress с items} | 404 (не назначен); 401
Source FR: FR-INTR-05

### PATCH /api/intranet/onboarding/me/items/:itemId
Auth: Bearer (только владелец — чужой чеклист недостижим по построению маршрута; прямой доступ к чужому → 403)
Request: { "checked": bool, "version" }
Responses: 200 {progress}; 401; 403 (чужой/frozen); 404; 409 (version)
Source FR: FR-INTR-05, FR-INTR-06

### GET /api/intranet/kudos/leaderboard?period=month|quarter|all&offset=0&limit=50
Auth: Bearer
Responses: 200 { "items":[{"user":{"id","name","avatar","deactivated"},"received_count"}],
                 "period","total","offset","limit" } (пустой период → items:[]);
           400 (period вне enum, limit>100); 401
Source FR: FR-INTR-09  · Пагинация: offset/limit (не NFR-14; NFR-14 = cursor list-issues)

### DELETE /api/intranet/news/:newsId  (расширение существующего)
Auth: Bearer; RBAC: admin ЛИБО автор поста (FR-INTR-07); Member на чужой → 403
Responses: 204 (→ корзина 30д + audit log); 401; 403; 404
Восстановление: через единый Trash API (`/admin/trash`), возвращает пост с опросом (FR-INTR-08)
Source FR: FR-INTR-07, FR-INTR-08
```

Random Coffee: клиентских эндпоинтов не добавляется (матчинг — cron; существующие
`/intranet/coffee/*` сохраняются). `POST /intranet/coffee/run` остаётся admin-запуском для отладки.

#### События NATS (`hub.intranet.*`, JetStream, outbox)

| Subject | Payload (ключевое) | Потребители |
|---------|--------------------|-------------|
| `hub.intranet.poll.voted` | poll_id, news_post_id, option_id, totals | dashboard (агрегаты), ws |
| `hub.intranet.news.deleted` | news_post_id, actor_id, had_poll | search (deindex), audit |
| `hub.intranet.news.restored` | news_post_id, actor_id | search (reindex) |
| `hub.intranet.onboarding.assigned` | user_id, template_id, assigned_by | notification (уведомление новичку) |
| `hub.intranet.onboarding.completed` | user_id, template_id | notification (админу) |
| `hub.intranet.coffee.matched` | round_id, pairs[{user_ids[], dm_channel_id}] | messenger (авто-DM), notification |
| `hub.intranet.coffee.skipped` | round_id, user_id, reason=not_enough_participants | notification |

DFD авто-DM: `[cron weekly] → CoffeeService.RunRound → Tx(CoffeeRound+CoffeePair) →
Outbox hub.intranet.coffee.matched → MessengerService (создать/переиспользовать DM, приветствие)
→ NotificationService`.

#### Трассируемость US → FR → API

| US | FR | API |
|----|----|-----|
| US-INTR-01 | FR-INTR-01..03 | POST /intranet/news (+poll), GET/PUT /intranet/news/:id/poll, POST .../poll/vote |
| US-INTR-02 | FR-INTR-04..06 | GET /intranet/onboarding/templates, POST/GET /intranet/onboarding/assignments, GET /intranet/onboarding/me, PATCH .../me/items/:itemId |
| US-INTR-03 | FR-INTR-07..08 | DELETE /intranet/news/:id (+ Trash API restore) |
| US-INTR-04 | FR-INTR-09 | GET /intranet/kudos/leaderboard |
| US-INTR-05 | FR-INTR-10..11 | cron + hub.intranet.coffee.matched (без новых клиентских endpoints) |

#### Edge-cases (сводно)

- Голос в закрытый (`closed`) опрос → 422; голос в опрос удалённого поста → 404.
- Удаление поста с опросом: опрос и голоса скрываются вместе с постом (одна корзинная запись).
- Onboarding: назначение деактивированному пользователю → 422; `frozen=true` блокирует PATCH items (403).
- Лидерборд: kudos, отправленные боту/самому себе, исключаются на уровне существующей валидации Kudo.
- Матчинг: нечётное N → ровно одна тройка; повтор прошлой пары допустим только при отсутствии иных комбинаций; деактивация между матчингом и встречей — DM остаётся, следующий раунд исключает.

### Calendar v2 (FR-CAL)

Доводка модуля `calendar` до паритета: ресурсы с бронированием, recurring-события,
iCal export/подписки, team availability. Календарь v4 — только агрегатор (calls/sprints/leave);
Волна 2 вводит **собственные события календаря** (`calendar_event`) как первокласную сущность —
без неё невозможны booking и recurrence.

#### Функциональные требования

| ID | Source | Поведение | NFR link |
|----|--------|-----------|----------|
| FR-CAL-01 | US-CAL-01 | Справочник ресурсов (GZ-V4-04): единая сущность, `kind` enum{room, equipment, other}. CRUD — только Admin (Member → 403); список — все. Деактивированный ресурс не предлагается для новых броней, существующие сохраняются | NFR-13 |
| FR-CAL-02 | US-CAL-01 | Бронирование ресурса при создании/правке события. Пересечение интервалов → 409 с указанием конфликтующего интервала; смежные интервалы (конец = начало) — не конфликт. Удаление события освобождает бронь; перенос — повторная проверка (409 при занятости) | NFR-33 |
| FR-CAL-03 | US-CAL-02 | Recurrence (GZ-V4-06): `freq` enum{daily, weekly, monthly} + ровно один из `until` (дата) / `count` (int ≥1). `until` в прошлом, `count`=0 или оба сразу → 400. Все вхождения ссылаются на одну серию | NFR-36 |
| FR-CAL-04 | US-CAL-02 | Правка серии применяется ко **всем** вхождениям (правка отдельного вхождения — вне scope), участникам — одно уведомление. Monthly 29/30/31 в коротком месяце → последний день месяца | NFR-36 |
| FR-CAL-05 | US-CAL-02 | Удаление серии удаляет все вхождения (корзина 30 дней). Ресурс на серии проверяется по каждому вхождению; частичный конфликт → 409 со списком конфликтных дат, серия не создаётся | NFR-11, NFR-33 |
| FR-CAL-06 | US-CAL-03 | Персональная iCal-ссылка с непереборным секретным токеном (≥128 бит, hash at rest); отзыв генерирует новый токен, старый немедленно невалиден (401) | NFR-34 |
| FR-CAL-07 | US-CAL-03 | Подписка на внешний iCal-фид по URL (GZ-V4-05): только https, `ValidateOutboundURL` (SSRF, NFR-21) + SafeHTTPClient + `io.LimitReader`. Невалидный/недоступный фид → ошибка валидации, подписка в статусе `error`, без дублей событий. Обновление ≥1 раз/час. События read-only (правка → 403). Лимит 10 подписок → 422. Удаление подписки убирает её события | NFR-34 |
| FR-CAL-08 | US-CAL-04 | Availability (GZ-V4-07): агрегация busy/free по выбранным пользователям = свои события календаря + вхождения серий + события iCal-подписок + **одобренные** отпуска HR (весь день; pending не показывается). Чужие события — только интервалы busy без деталей; прямой запрос чужого события → 403. Свои — с деталями. Пользователь без событий → full free (200) | NFR-35 |

#### Новые NFR

| ID | Требование | Значение |
|----|------------|----------|
| NFR-33 | Отсутствие двойных броней | Проверка пересечения на уровне БД (PostgreSQL exclusion constraint по `(resource_id, tstzrange(starts_at, ends_at))`); конкурентные брони → ровно одна успешна, вторая 409. Тест: 2 параллельных запроса |
| NFR-34 | Безопасность iCal-фидов | `ValidateOutboundURL` перед каждой загрузкой; SafeHTTPClient timeout 10s; `io.LimitReader` 1 MB на фид; refresh-джоба ≥1/час, ≤1 конкурентная загрузка на подписку; токен export-ссылки ≥128 бит энтропии, в БД — hash |
| NFR-35 | Availability p95 | < 700ms для 20 пользователей × окно 7 дней (индексы по `(user_id, starts_at)`; leave по `(user_id, status, start_date)`) |
| NFR-36 | Лимиты recurrence | `count` ≤ 365; `until` ≤ 2 года от `starts_at`; превышение → 400. Экспансия вхождений — на чтении (materialization не хранится), окно запроса ≤ 92 дня |

#### ERD-дельта (PostgreSQL / Ent)

```
User ──┬── CalendarEvent ──┬── EventRecurrence (0..1)
       │                   └── ResourceBooking N──1 Resource
       ├── ICalSubscription (≤10)
       └── ICalExportToken (1, rotate)
LeaveRequest (hr, существует) → участвует в availability как busy (status=approved)
```

| Сущность | Поля (помимо инвариантов id UUID v4 / version / UTC / soft delete) | Индексы/ограничения |
|----------|--------------------------------------------------------------------|---------------------|
| `calendar_event` | `owner_id` FK, `title` string(200), `description` text nullable, `starts_at`/`ends_at` timestamptz (UTC, NFR-12), `participants` JSON uuid[], `source` enum{hub, ical} | `(owner_id, starts_at)`; check `ends_at > starts_at` |
| `event_recurrence` | `event_id` FK unique, `freq` enum{daily, weekly, monthly}, `interval` int default 1, `until` timestamptz nullable, `count` int nullable | check: ровно один из until/count (FR-CAL-03) |
| `resource` | `name` string(120), `kind` enum{room, equipment, other}, `active` bool default true | `(kind, active)` |
| `resource_booking` | `resource_id` FK, `event_id` FK, `occurrence_date` date nullable (для серий), `starts_at`/`ends_at` timestamptz | **exclusion constraint** `(resource_id, tstzrange)` (NFR-33); `(event_id)` |
| `ical_subscription` | `user_id` FK, `url` string(2048), `name` string(120), `status` enum{active, error}, `last_synced_at` nullable, `last_error` string nullable, `etag` string nullable | `(user_id)`; ≤10/user (FR-CAL-07) |
| `ical_export_token` | `user_id` FK unique, `token_hash` string (SHA-256), `rotated_at` timestamptz | `(token_hash)` unique |

События внешних фидов хранятся как `calendar_event` с `source=ical` + FK
`subscription_id` (nullable), read-only на уровне сервиса; удаление подписки каскадно
удаляет её события (жёстко — это реплика внешних данных, не пользовательский контент).

#### API-контракты

```
### GET /api/calendar/resources        · 200 {items[]}; 401         (все роли)
### POST /api/calendar/resources       · Admin; 201; 400; 401; 403  (Member → 403, FR-CAL-01)
### PUT /api/calendar/resources/:id    · Admin; 200; 400; 401; 403; 404; 409 (version)
### DELETE /api/calendar/resources/:id · Admin; 204 (деактивация, брони сохраняются); 401; 403; 404
Source FR: FR-CAL-01

### POST /api/calendar/events
Auth: Bearer
Request: { "title", "starts_at", "ends_at", "description"?, "participants"?: [uuid],
           "resource_ids"?: [uuid],
           "recurrence"?: { "freq": "daily|weekly|monthly", "interval"?: 1, "until"? | "count"? } }
Responses: 201 {event};
           400 (until в прошлом / count=0 / оба until+count / лимиты NFR-36);
           401; 404 (ресурс);
           409 (конфликт брони; для серии — {"error","request_id","conflicts":[даты]});
           422 (деактивированный ресурс); 429
Source FR: FR-CAL-02, FR-CAL-03, FR-CAL-05

### PUT /api/calendar/events/:eventId          (правка серии = всех вхождений, FR-CAL-04)
Request: как POST + "version"
Responses: 200; 400; 401; 403 (чужое или source=ical); 404; 409 (version / бронь)
### DELETE /api/calendar/events/:eventId       · 204 (серия целиком → корзина 30д, брони освобождены); 401; 403; 404
Source FR: FR-CAL-04, FR-CAL-05

### GET /api/calendar/events?from=&to=          (существующий, расширяется)
Окно ≤ 92 дня (400 при превышении). Серии разворачиваются в вхождения на чтении (NFR-36).
Ответ дополняется полями "series_id"?, "source", "resources":[...]

### GET /api/calendar/availability?user_ids=uuid,uuid&from=&to=
Auth: Bearer
Responses: 200 { "users": [ { "user_id",
                  "intervals": [ { "starts_at","ends_at","busy":true,
                                   "detail": {title,...} | null } ] } ] }
           (detail только для событий, где запрашивающий — владелец/участник; чужие — null, FR-CAL-08);
           400 (>20 user_ids, окно >31 дня); 401
Source FR: FR-CAL-08  · NFR-35

### GET /api/calendar/subscriptions             · 200 {items[]}; 401
### POST /api/calendar/subscriptions
Request: { "url": "https://...", "name"? }
Responses: 201 {subscription, status};
           400 (не https / не прошёл ValidateOutboundURL / не-iCal контент);
           401; 422 (>10 подписок); 429
### DELETE /api/calendar/subscriptions/:id      · 204 (+ каскад событий фида); 401; 404
Source FR: FR-CAL-07  · NFR-34

### GET /api/calendar/export.ics?token=<secret>  (существующий переводится на токен)
Auth: секретный токен в query (без JWT — для внешних клиентов); 401 при невалидном/отозванном
### POST /api/calendar/export/rotate             · Bearer; 200 {"url": новая ссылка}; 401
Source FR: FR-CAL-06
```

#### События NATS (`hub.calendar.*`, JetStream, outbox)

| Subject | Payload (ключевое) | Потребители |
|---------|--------------------|-------------|
| `hub.calendar.event.created` / `.updated` / `.deleted` | event_id, series_id?, owner_id, participants | notification (одно уведомление на серию, FR-CAL-04), ws |
| `hub.calendar.booking.created` / `.released` | booking_id, resource_id, event_id, interval | dashboard (занятость ресурсов) |
| `hub.calendar.feed.sync_failed` | subscription_id, user_id, error | notification (владельцу подписки) |

DFD booking-конфликта: `[User] → POST /calendar/events → [Auth] → [CalendarService.Create]
→ [Tx: INSERT calendar_event + resource_booking (exclusion constraint)]
→ constraint violation → 409 {"error","request_id","conflicts"} | OK → Outbox hub.calendar.event.created`.

#### Трассируемость US → FR → API

| US | FR | API |
|----|----|-----|
| US-CAL-01 | FR-CAL-01..02 | GET/POST/PUT/DELETE /calendar/resources, POST/PUT /calendar/events (resource_ids) |
| US-CAL-02 | FR-CAL-03..05 | POST/PUT/DELETE /calendar/events (recurrence), GET /calendar/events (экспансия) |
| US-CAL-03 | FR-CAL-06..07 | GET /calendar/export.ics?token, POST /calendar/export/rotate, GET/POST/DELETE /calendar/subscriptions |
| US-CAL-04 | FR-CAL-08 | GET /calendar/availability |

#### Edge-cases (сводно)

- Смежные брони (15:00–16:00 после 14:00–15:00) — не конфликт: exclusion constraint по полуоткрытому интервалу `[)`.
- Перенос события с бронью: старая бронь удерживается до успешной проверки новой (Tx), при 409 ничего не меняется.
- Monthly 31-го: вхождение в феврале — 28/29 (clamp к последнему дню), серия не «теряет» месяц.
- Фид с дубликатами UID: upsert по `(subscription_id, external_uid)` — без дублей после повторной синхронизации.
- Отзыв export-токена: старая ссылка → 401 немедленно (проверка по hash в БД, не по кешу).
- Availability: события iCal-подписок участвуют в busy владельца; отпуск pending — не busy; TZ смотрящего — конвертация на фронте (NFR-12).

### Wiki — новый модуль (FR-WIKI)

Модуля в коде нет — спецификация с нуля (реализация: Волны 5/8; handbook = пространство общей
wiki, GZ-V4-12/13). Редактор markdown-first (WYSIWYG вне scope v1). RBAC — решение GZ-SEC-07/11:
права на space (read/write/admin), страницы наследуют, restrict-флаг сужает круг на отдельной
странице. Поиск — только user-scoped (инвариант NF-03 / NFR-19).

#### Функциональные требования

| ID | Source | Поведение | NFR link |
|----|--------|-----------|----------|
| FR-WIKI-01 | US-WIKI-01 | Space CRUD; страницы в иерархии `parent_id` внутри space; markdown-контент с предпросмотром; перемещение страницы по дереву (ссылки стабильны — UUID). Цикл (`A` под своего потомка) → 422; глубина >10 → 422 | NFR-37 |
| FR-WIKI-02 | US-WIKI-01 | Конкурентная правка — optimistic lock `version`: второй save → 409. Изображения — через storage presigned upload (NFR-27), не base64 | NFR-24 |
| FR-WIKI-03 | US-WIKI-01 | Удаление страницы с детьми — каскад **поддерева** в корзину 30 дней; восстановление возвращает поддерево. Space удаляется только пустым (иначе 422) | NFR-11 |
| FR-WIKI-04 | US-WIKI-02 | Роли space: `read` / `write` / `admin`; страницы наследуют права space. Не-участник space → 403 на любую страницу (в т.ч. по прямой ссылке). Управление правами — только space admin. У space всегда ≥1 admin: снятие последнего → 422 | NFR-38 |
| FR-WIKI-05 | US-WIKI-02 | Restrict-флаг страницы: явный allow-list (подмножество имеющих доступ к space — расширить доступ за пределы space нельзя, чужие id → 422). Вне списка → 403; страница скрыта из дерева и поиска. Снятие restrict возвращает наследование. System Admin — служебный доступ (восстановление/GDPR) с записью в audit log | NFR-38, NFR-39 |
| FR-WIKI-06 | US-WIKI-03 | Версия создаётся на каждый save (autosave-черновик не версионируется): автор, дата UTC, номер. Просмотр любой версии + diff с текущей. Restrict распространяется на историю (403) | NFR-38 |
| FR-WIKI-07 | US-WIKI-03 | Rollback: контент версии N-k становится **новой** версией N+1 (история не переписывается), фиксируется автор отката. Роль read → 403. История восстановленной из корзины страницы сохраняется | NFR-38 |
| FR-WIKI-08 | US-WIKI-04 | Issue-link M:N (страница ↔ задачи, двусторонне): из задачи видна ссылка на страницу (только при read-доступе к странице — иначе скрыта), на странице — задача со статусом. Привязка к задаче без доступа к её проекту → 403. Удаление задачи/страницы → tombstone в UI, без битых ссылок | NFR-13 |
| FR-WIKI-09 | US-WIKI-05 | Поиск через search-модуль (Ctrl+K, `types=wiki_page`): фильтрация по правам **на сервере на момент запроса** (space membership + restrict allow-list); страницы из корзины не выдаются; сниппет — только заголовок+фрагмент доступных страниц. Удаление из space → страницы исчезают из выдачи (0 результатов ≠ ошибка) | NFR-39 |

#### Новые NFR

| ID | Требование | Значение |
|----|------------|----------|
| NFR-37 | Лимиты wiki | Глубина дерева ≤ 10 (422); markdown ≤ 1 MB на страницу (400); title ≤ 200; страниц в space ≤ 5000; p95 чтения страницы < 300ms |
| NFR-38 | Целостность RBAC/версий | Проверка прав — в handler/service до данных; ≥1 admin на space (инвариант БД+сервис); каждая версия неизменяема после создания (append-only); rollback не удаляет версии |
| NFR-39 | User-scope поиска wiki | Фильтр по membership+restrict выполняется на сервере при каждом запросе (не на клиенте, не только при индексации — NF-03); негативный permission-тест: пользователь вне space получает 0 его страниц |

#### ERD (новые сущности, PostgreSQL / Ent)

```
WikiSpace ──┬── WikiSpaceMember (role: read|write|admin, ≥1 admin)
            └── WikiPage (parent_id self-FK, restricted flag)
                   ├── WikiPageRestriction (allow-list, только при restricted)
                   ├── WikiPageVersion (append-only)
                   └── WikiIssueLink N──N Issue (tracker)
```

| Сущность | Поля (помимо инвариантов id UUID v4 / version / UTC / soft delete) | Индексы/ограничения |
|----------|--------------------------------------------------------------------|---------------------|
| `wiki_space` | `name` string(120), `slug` string(60) unique, `description` nullable, `created_by` FK | `(slug)` unique |
| `wiki_space_member` | `space_id` FK, `user_id` FK, `role` enum{read, write, admin} | `(space_id, user_id)` unique; инвариант ≥1 admin (FR-WIKI-04) |
| `wiki_page` | `space_id` FK, `parent_id` self-FK nullable, `title` string(200), `content_md` text (≤1MB), `restricted` bool default false, `position` int, `created_by`/`updated_by` FK | `(space_id, parent_id, position)`; `(space_id, deleted_at)`; check: parent в том же space |
| `wiki_page_restriction` | `page_id` FK, `user_id` FK | `(page_id, user_id)` unique; только user'ы с доступом к space (FR-WIKI-05) |
| `wiki_page_version` | `page_id` FK, `number` int, `content_md` text, `title` string(200), `author_id` FK, `restored_from` int nullable | `(page_id, number)` unique; **append-only** (NFR-38) |
| `wiki_issue_link` | `page_id` FK, `issue_id` FK, `created_by` FK | `(page_id, issue_id)` unique |

Иерархия — adjacency list (`parent_id`) + проверка цикла/глубины в сервисе (рекурсивный CTE);
materialized path не требуется при глубине ≤10.

#### API-контракты

```
### GET  /api/wiki/spaces                        · 200 {items[]} (только мои space); 401
### POST /api/wiki/spaces                        · 201 (создатель = admin); 400; 401; 429
### GET  /api/wiki/spaces/:spaceId               · 200; 401; 403 (не участник); 404
### PUT  /api/wiki/spaces/:spaceId               · space admin; 200; 400; 401; 403; 404; 409 (version)
### DELETE /api/wiki/spaces/:spaceId             · space admin; 204; 401; 403; 404; 422 (space не пуст)
Source FR: FR-WIKI-01, FR-WIKI-03, FR-WIKI-04

### GET  /api/wiki/spaces/:spaceId/members       · участники space; 200; 401; 403; 404
### PUT  /api/wiki/spaces/:spaceId/members/:userId
Auth: Bearer; RBAC: space admin
Request: { "role": "read|write|admin" }
Responses: 200; 400; 401; 403; 404; 422 (снятие последнего admin)
### DELETE /api/wiki/spaces/:spaceId/members/:userId · space admin; 204; 401; 403; 404; 422 (последний admin)
Source FR: FR-WIKI-04

### GET  /api/wiki/spaces/:spaceId/tree
Дерево страниц space; restrict-страницы вне allow-list запрашивающего исключены (FR-WIKI-05).
Responses: 200 {tree}; 401; 403; 404

### POST /api/wiki/spaces/:spaceId/pages
Auth: Bearer; RBAC: write в space
Request: { "title": "1..200", "content_md"?: "≤1MB", "parent_id"?: uuid }
Responses: 201; 400 (лимиты NFR-37); 401; 403; 404; 422 (глубина >10); 429
### GET  /api/wiki/pages/:pageId                 · read + restrict-check; 200; 401; 403; 404
### PUT  /api/wiki/pages/:pageId                 · write + restrict-check
Request: { "title"?, "content_md"?, "version" }
Responses: 200 (+новая WikiPageVersion); 400; 401; 403; 404; 409 (version, FR-WIKI-02)
### DELETE /api/wiki/pages/:pageId               · write; 204 (поддерево → корзина 30д); 401; 403; 404
### POST /api/wiki/pages/:pageId/move
Request: { "parent_id": uuid|null, "position"?: int, "version" }
Responses: 200; 401; 403; 404; 409; 422 (цикл / глубина / другой space)
Source FR: FR-WIKI-01..03

### PUT /api/wiki/pages/:pageId/restrict
Auth: Bearer; RBAC: space admin
Request: { "restricted": bool, "allowed_user_ids": [uuid], "version" }
Responses: 200; 400; 401; 403; 404; 409;
           422 (allowed_user_ids содержит пользователей без доступа к space — FR-WIKI-05)
Source FR: FR-WIKI-05

### GET  /api/wiki/pages/:pageId/versions?offset=0&limit=50
· read + restrict-check; 200 {items:[{number, author, created_at}], total, offset, limit}; 400; 401; 403; 404
### GET  /api/wiki/pages/:pageId/versions/:number · содержимое версии + diff с текущей; 200; 401; 403; 404
### POST /api/wiki/pages/:pageId/versions/:number/restore
· RBAC: write (read → 403, FR-WIKI-07); 200 (новая версия N+1, restored_from=number); 401; 403; 404; 409
Source FR: FR-WIKI-06, FR-WIKI-07  · Пагинация: offset/limit (не NFR-14; NFR-14 = cursor list-issues)

### GET  /api/wiki/pages/:pageId/issue-links     · 200 {items:[{issue: {id,key,status,title} | tombstone}]}; 401; 403; 404
### POST /api/wiki/pages/:pageId/issue-links
Request: { "issue_id": uuid }
Responses: 201; 401; 403 (нет доступа к проекту задачи ИЛИ к странице); 404; 409 (дубль связи)
### DELETE /api/wiki/pages/:pageId/issue-links/:issueId · 204; 401; 403; 404
### GET  /api/projects/:projectId/issues/:issueId/wiki-links   (сторона tracker)
· 200 {items: только страницы, к которым у запрашивающего есть read — иначе скрыты (FR-WIKI-08)}; 401; 403; 404
Source FR: FR-WIKI-08

### Поиск: существующий GET /api/search?q=...&types=wiki_page
Фильтр прав на сервере при каждом запросе (FR-WIKI-09, NFR-39); отдельного endpoint не вводится.
```

#### События NATS (`hub.wiki.*`, JetStream, outbox)

| Subject | Payload (ключевое) | Потребители |
|---------|--------------------|-------------|
| `hub.wiki.page.created` / `.updated` | page_id, space_id, author_id, version_number | search (index/reindex), ws |
| `hub.wiki.page.deleted` / `.restored` | page_id, space_id, subtree_page_ids[] | search (deindex/reindex поддерева), audit |
| `hub.wiki.page.restricted_changed` | page_id, restricted, allowed_user_ids[] | search (пере-скоуп индекса) |
| `hub.wiki.space.member_changed` | space_id, user_id, role|removed | search (инвалидация скоупа), notification |
| `hub.wiki.issue_link.created` / `.removed` | page_id, issue_id, actor_id | tracker (обратная ссылка), ws |

DFD user-scoped поиска: `[User Ctrl+K] → GET /search?types=wiki_page → [Auth]
→ [SearchService: Meilisearch query] → [WikiScopeFilter: membership + restrict на момент запроса]
→ результат только доступных страниц` — фильтр обязателен на сервере (NF-03).

#### Трассируемость US → FR → API

| US | FR | API |
|----|----|-----|
| US-WIKI-01 | FR-WIKI-01..03 | /wiki/spaces CRUD, /wiki/spaces/:id/tree, /wiki/spaces/:id/pages, /wiki/pages/:id (GET/PUT/DELETE), /wiki/pages/:id/move |
| US-WIKI-02 | FR-WIKI-04..05 | /wiki/spaces/:id/members (GET/PUT/DELETE), /wiki/pages/:id/restrict |
| US-WIKI-03 | FR-WIKI-06..07 | /wiki/pages/:id/versions, /versions/:number, /versions/:number/restore |
| US-WIKI-04 | FR-WIKI-08 | /wiki/pages/:id/issue-links (GET/POST/DELETE), /projects/:pid/issues/:iid/wiki-links |
| US-WIKI-05 | FR-WIKI-09 | GET /search?types=wiki_page (существующий, + WikiScopeFilter) |

#### Edge-cases (сводно)

- Перемещение страницы в другой space — запрещено в v1 (422): права наследуются от space, перенос между скоупами — отдельное решение.
- Restrict на странице с детьми: флаг действует на саму страницу; дети наследуют space-права (restrict не каскадит) — иначе allow-list «протекает» вниз неявно.
- Snapshot-вопрос diff: diff считается на чтении между `content_md` версий, ничего не хранится дополнительно.
- Одновременный rollback двумя пользователями: второй получает 409 (version страницы).
- GDPR/удаление пользователя: авторство версий анонимизируется (tombstone), контент сохраняется (GZ-SEC-17/18 [default]).
- Поиск: результат не содержит restrict-страниц даже для участников space вне allow-list; membership проверяется на **момент запроса**, не индексации.

## Известные gaps (требование → код → тест)

Полная матрица: `.planning/COVERAGE.md`

### Актуальные пробелы (2026-06-23)

| Область | Статус | Gap |
|---------|--------|-----|
| Backend coverage overall | **83.0%** ✅ (2026-06-22 Docker baseline) | Gate ≥82% PASS |
| PG-03 secret rotation | ⏳ Manual | FA-06 Phase 1–3 — **единственный prod-блокер** |
| Lighthouse ≥90 formal | ⚠️ Partial | CI artifact job (PR #27, non-blocking) |
| NFR-16 WCAG 2.1 AA | ⚠️ Partial | Login axe PASS; full human audit pending |
| US-F03 slash `/ai test-cases` | ⚠️ Partial | Composer suggest вместо slash-команды |
| Webcam в звонках | ❌ Out of scope | TZ forbidden |
| OpenAI-compatible Hub API (US-F06) | ❌ Missing | v3.5 backlog |
| Test framework migration | ✅ Complete | 488 testframework / 0 internal |

### Закрытые gaps (post-audit)

| Область | Было | Сейчас |
|---------|------|--------|
| CRITICAL security (12) | Blockers | **FIXED** в коде |
| Overall coverage | 11.5% → 56.6% | **83.0%** (2026-06-22 Docker baseline) |
| Playwright E2E | not present | **20 specs** |
| Frontend Vitest | 0 | **179** files (`testframework/stages/unit/frontend-test/`, measured 2026-07-29; 61 `.ts` + 118 `.tsx`) |
| Test framework migration | in progress | **488 / 0 internal** ✅ |
| Lighthouse CI | missing | **artifact job** (PR #27) |
| Squash PR v2.5-trust | pending | **merged PR #1** ✅ |

---
*Последнее обновление: 2026-07-29 — TASK-183 inventory recount BE **812** / Vitest **179**; NFR-30 sync; COV 83.0% baseline 2026-06-22; FA-06 manual blocker*

## Messenger Phase 151 — bounded contract (2026-08-26)

This section records the implemented bounded slice, not completion of Phase 151.

- `ChannelResponse.capabilities` is the only frontend authority for Messenger actions. Missing capability data fails closed; the client does not infer rights from `role`, `is_readonly`, message ownership, or cached membership. Fresh projection is mandatory for Get/List and mutation/join responses, including Create/Update channel, new/existing DM, and AcceptInvite.
- `GET /messenger/channels/discover` returns a plain read-only `ChannelResponse[]` browse seed capped at 100 rows in `created_at DESC, id DESC` order. Candidates are live public same-tenant global channels or active-project channels where the actor is an active project member, and every candidate is re-authorized with fresh `ReadTimeline` before projection. There is no join/leave side effect, cursor, or completeness guarantee past row 100.
- The frontend Browse surface is lazy and identity-scoped. It renders the server order only for public rows with explicit `capabilities.can_read=true`; authorization errors remove the exact browse cache payload and no browse row navigates or initiates Join/Leave.
- Direct replies may set `also_send_to_channel`; the flag is valid only with a direct thread root and the same canonical message is projected into the thread and, when flagged, the root timeline.
- Forward authorization is asymmetric: source requires `CanRead && CanForward`; target requires `CanPost`. Both scoped proofs, or both legacy source/target membership checks, precede the first Mongo message lookup.
- Reactions require fresh `CanReact` before Mongo; legacy membership is also checked against the route channel before message lookup.
- `channel_removed` is a client revoke barrier. Composer sends/retries/uploads, reactions, message mutations, forward targets, typing, message queries, and route rendering must stop or fail closed after revoke. Only a server-authorized `channel_created` lifecycle event clears the marker before authoritative refetch; stale snapshots remain filtered while revoked.
- WebSocket room subscription, per-user channel lifecycle delivery, and channel audience projection require a fresh PostgreSQL `ChannelActionReceiveEvent` proof through `WSChannelAuthority`; denied, duplicate, and malformed candidates are omitted.
- A valid active user with no authorized unread targets receives `channels`, `dms`, `mentions`, and `threads` as empty maps plus `any_unread=false`, without entering the unread Mongo coordinator.
- Delete is two-stage: fresh membership/read eligibility permits one ownership lookup; an ordinary member may delete only their own message, while owner/admin/moderator may delete any message.
- Bookmarks are sensitive user-scoped data. Backend list enrichment groups distinct bookmark channels, completes one fresh read proof per unique target before Mongo, and never projects preview text when the message channel differs from the bookmark channel. Frontend query keys include the authenticated user; no-user cached rows are masked, and remove/navigation callbacks fail closed after an identity switch.
- `ListPins` proves fresh requested-channel authority before repository access and filters nil or mismatched-channel repository rows before projection. This is a response no-leak invariant, not proof of live database corruption.
- `TogglePin` treats only Ent NotFound as an absent pin. Any other `FindPin` error returns a redacted Internal error before `AddPin` or `RemovePin`; found rows retain deterministic unpin behavior.
- Pinned-message frontend cache keys include channel and authenticated user. Fetching is disabled without identity and disabled-query data is masked; unpin completion invalidates only the captured identity/channel key after rechecking identity.
- For an authenticated user, the pinned-list UI renders mutually exclusive loading, retryable error, successful-empty, and populated states. A query error never appears as an empty list; retry refetches the current identity/channel key. The logged-out disabled-query state renders no cached pin rows.
- Pinned-list access requires explicit `capabilities.can_list_pins=true`; unpin requires `can_pin=true`. Missing data fails closed. Unpin preflight and completion recheck the latest DTO snapshot, current store capabilities, identity, channel binding, and revoke barrier before mutation or exact query invalidation.
- A message-row pin toggle invalidates only `pinsQueryKey(channelId,userId)` after success. Missing/switched identity, DTO/store `can_pin` loss, or revoke prevents stale invalidation. The boolean toggle response is not treated as sufficient data to synthesize a `PinnedMessage` row.

- Root timeline state is identity-bound end to end: fetch, reconnect invalidation, WS deltas, generic invalidation, optimistic edit, and pin previews use messenger/messages/channelId/userId. Without identity or after channel revoke, room/query/event work fails closed and both flattened messages and raw hook data mask resident legacy or sentinel payloads. Delayed A results never render under B. Paged thread replies and production thread detail use distinct channel/root/user keys so their payload types cannot collide; both mask legacy/sentinel data and disable work without identity or after revoke. Identity cleanup preserves only current-user schemas, while channel revoke still matches both thread key shapes by channel.
- ThreadPanel uses the detail query only for its parent projection and renders replies exclusively from the WS-aware paged cache. Reply loading is independent of parent loading. A bounded pending map scoped to authenticated user, channel, and root preserves direct `message_sent` deltas that arrive before the initial reply snapshot exists, then merges them after fetch. Successful send fallback invalidates exact root plus optional snapshotted thread keys; live composer state cannot retarget completion.
- Direct replies are also journaled while a resident cache refetches; sent-message overflow schedules exact recovery. Edit/delete/reaction deltas use a bounded ordered journal whose overlay is applied to raw and flattened public thread data until canonical recovery completes without rearm. Mutation-journal overflow masks public thread data fail-closed. Pending sends capture the authenticated principal and cannot retry or complete under a later identity. Thread issue-link queries include user identity and render/mutate only for the current user on a non-revoked channel.
- `GET /api/projects/:projectId/message-links?channel_id=...&message_id=...` requires route `PermProjectView` plus a current actor. Before link lookup, the service executes exact non-deleted project-channel proof → fresh `IsProjectMember(projectID, actorID)` → current `IsChannelMember(channelID, actorID)`. Global, foreign, deleted, inactive/non-project, and channel-nonmember cases return the same 404 `канал не найден`; either membership-proof infrastructure failure returns one redacted 500. Authorized absence is `[]`; no Mongo root-existence proof is claimed. The proof is sequential, so a revoke-to-list TOCTOU remains. Create/unlink policy and client list/manage capabilities remain `UNBOUND`.
- Checked-in OpenAPI exposes only `POST /messenger/channels/{channelId}/messages/{messageId}/promote-to-issue`. `GET /messenger/unread` references `UnreadResponse` with `channels`, `dms`, `mentions`, and `threads` as `map[string]int64`, plus `any_unread: bool`. A Docker unit oracle semantically checks `swagger.json`, the embedded `docs.go` template, and `swagger.yaml`; no generator or live service is exercised, and annotation regeneration alignment remains pending.
- Promotion source binding is route-scoped: only after fresh `ChannelActionPromote(channelId)` may the backend execute one private `FindLiveMessageInChannel(channelId, messageId)` lookup. The Mongo predicate includes `_id`, `channel_id`, and `deleted_at:nil`; absent, deleted, and foreign-channel IDs are indistinguishable 404 responses. Missing capability, store errors, or malformed results are redacted 500 responses with no generic source-ID fallback and no downstream promotion effects.
- Promotion target authority is a separate fresh boundary after the scoped source. `RBACService.AuthorizePromotionTarget` bypasses PermissionCache and resolves an active actor plus an exact active, non-soft-deleted project in the actor organization before any founder/admin bypass. Unavailable targets return concealed 404, an ordinary actor without fresh `task.create` returns 403, and dependency/query/invariant failures return redacted 500 before downstream reads/effects. Custom `permission_names` override the legacy mask when nonempty; the legacy mask remains the compatibility fallback. Live Ent/PostgreSQL lifecycle execution remains `NOT_RUN`.
The pure post-create prerequisite accepts a full prepared message, deep-copies all mutable payload collections, freezes resolved mentions, assigns server pending/UTC state, rejects server-managed fields, and fails invalid JSON before repository access. Membership lifecycle coordination and canonical post-create/recovery production binding remain `UNBOUND`. Live PostgreSQL/Mongo/NATS/WS/browser behavior remains `NOT_RUN`; see the Phase 151 verification artifact.

## Phase 151 promotion parent/reply bounded contract (2026-08-26)

Following fresh route/source and target proofs, promotion resolves its parent through a channel-scoped live lookup and replies through a private `channel_id + thread_id(rootID) + deleted_at:nil` lookup before GetChannel, promoter, backlink, or linker work. Every row is validated; generic parent/reply fallback is prohibited. Replies are selected `created_at DESC, _id DESC`, capped at 100. This is unit evidence only: live Mongo/Ent-PostgreSQL/NATS/WS are `NOT_RUN`; delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`.

## Phase 151 selected-reply retention contract (2026-08-26)

After the private newest-100 reply lookup, the selected reply is retained exactly once when absent: a short page appends it, while a full page becomes first 99 page rows plus that reply. The parent is separate from the reply limit and an already-present selected reply is unchanged. Duplicate reply IDs and more than 100 returned rows are fixed/redacted 500 before GetChannel or any promotion effect. Assembly uses a fresh slice, so it does not mutate the repository-owned page. This is bounded unit evidence only: live Mongo/Ent-PostgreSQL/NATS/WS remain `NOT_RUN`; cross-store delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`.

## Phase 151 promotion source-ID defensive contract (2026-08-26)

After scoped source lookup, a `found=true` result is accepted only when source is nonnil, live, route-channel bound, and has nonempty ID exactly matching route `messageId`. Any malformed found result is a fixed/redacted 500 before target/thread/channel/effects; `found=false` remains concealed 404. Live Mongo/Ent-PostgreSQL/NATS/WS are `NOT_RUN`; cross-store delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`.

## Phase 151 mutation-adjacent target recheck contract (2026-08-26)

`PromoteToIssue` snapshots one target-authority checker instance. It performs an early fresh target proof, then performs a second fresh proof with the identical actor/project after scoped parent/replies/GetChannel/all description display-name reads and immediately before `CreateIssueFromPromotion`. At both proof points, nil/error/false-true maps to redacted 500, false-false to concealed 404, and true-false to 403. A second-proof rejection has no promoter/backlink/linker effects. The contract is bounded unit evidence: root/reply order and replacement of a checker instance are regression-covered; live Mongo/Ent-PostgreSQL/NATS/WS remain `NOT_RUN`. Source/channel/reply delete/revoke TOCTOU, canonical backlink/post-create lifecycle, and post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create) remain `OPEN/UNBOUND`.

## Phase 151 chronological promotion description contract (2026-08-26)

Promotion selection remains the newest 100 live replies ordered `created_at DESC, _id DESC`. After retaining an absent selected reply within that 100-row budget, every effective row must be nonnil with a nonempty ID and nonzero timestamp, and every adjacent tuple must remain strictly descending: the previous timestamp is later, or equal timestamps have `previous.ID > current.ID`. Validation occurs before GetChannel, display assembly, or effects and fails with a redacted internal error. The description then renders the parent first and the effective replies by `CreatedAt ASC`, breaking equal timestamps by `ID ASC`. Sorting occurs on a fresh slice and must not mutate repository-owned input. Bounded unit tests cover root and reply promotion, timestamp ties, full/short selected-reply retention, input-page immutability, timestamp inversion, and ID tie-break inversion. Live Mongo/Ent-PostgreSQL/NATS/WS remain `NOT_RUN`; parent/reply/channel delete/revoke TOCTOU, post-recheck source deletion before issue creation, canonical backlink/post-create lifecycle, and post-recheck target revoke/delete before non-atomic issue creation remain `OPEN/UNBOUND`.

## Phase 151 promotion dependency snapshot contract (2026-08-26)

At `PromoteToIssue` entry, the service must atomically read one `IssuePromoter`/`PromotePermissionChecker`/`IssueThreadLinker` configuration snapshot under a shared `RWMutex`, then release the lock before validation and all dependency calls. A setter invoked during either target proof changes only later request snapshots. Test-only checker/linker exports use the same lock. This guarantees request-stable dependencies and race-safe access, not atomic batching across three separate setter calls. Production config remains startup wiring before route registration. Live stores and promotion lifecycle transactionality remain outside this bounded unit contract.

## Phase 151 mutation-adjacent live-source recheck contract (2026-08-26)

Promotion must execute the same exact `(route channelID, route messageID)` live-source validator twice: once before target/thread reads and once after description assembly, immediately before the final target proof. Both proofs map lookup error or malformed found data to fixed/redacted 500 and missing/deleted to concealed 404. A second-proof failure returns no issue response and performs no issue promoter, backlink, or thread-link effect. The proof validates live identity, not content revision, and is not atomic with issue creation; later source deletion plus parent/reply/channel and target TOCTOU remain open.

## Phase 151 promotion live-parent recheck contract (2026-08-26)

For a reply source (`thread_id != ""`), `PromoteToIssue` must resolve an exact live root in the same route channel twice: initially before scoped reply reads, and after description/source recheck immediately before the final target-project proof. Missing or nested/non-root parent state is non-disclosive 404. Repository failure or nil, empty-ID, ID-mismatched, foreign-channel, or deleted found data is a fixed/redacted 500. Failure of the second proof returns no issue response and must not invoke the final target proof, issue promoter, backlink writer, or issue-thread linker. A root source performs no parent lookup. These are sequential proofs, not an atomic guarantee across Mongo, PostgreSQL, and tracker mutation.

## Promotion source-channel authority recheck (2026-08-27)

`PromoteToIssue` applies the same source-channel authority mapping twice: before reading the route-bound live source and after description plus live source/root rechecks, immediately before the final target-project proof. A configured channel-access resolver must still authorize `ChannelActionPromote`; resolver failure, malformed access, or missing capability uses the existing 403 contract. The legacy path still requires current channel membership, maps lookup failure to fixed/redacted 500, and maps non-membership to 403. A late rejection performs no second target proof, issue creation, backlink, or thread link. This sequential recheck does not make channel authorization and tracker mutation transactional; post-proof lifecycle races remain open.

## Phase 151 canonical create-result defensive contract (2026-08-27)

Before a pure post-create request may return a found canonical result or run effects from a persist result, the canonical message must be nonnil, have a nonempty ID, and match the authorized channel, actor/sender, and request `client_nonce` exactly. A lookup hit must report `Created=false`. Missing legacy document effect state normalizes to `complete`; otherwise only `pending` or `complete` is accepted, and `CreateMessageResult.PostCreateEffectsState` must be explicit and equal the normalized document state. Any mismatch returns the zero result with `ErrPostCreatePipelineInvalid` before derived effects. A valid existing lookup and a valid `Created=false` persist race return without repeating effects. The following v1 and legacy-reservation contracts close pure-seam nonce metadata classification; live invocation of the frozen typed nonce mapper, production writer binding, and live Mongo/NATS/WS behavior remain `UNBOUND/NOT_RUN`.

## Phase 151 durable client-nonce fingerprint v1 contract (2026-08-27)

Every pure-pipeline prepared document carries internal BSON `client_nonce_fingerprint_version=1` and a lowercase SHA-256 fingerprint of original create intent. The canonical input binds producer, channel, sender, nonce, message type, poll id, content JSON/text, thread id and root projection flag, forward source, voice fields, ordered attachments, and normalized mention/group sets. JSON map order/equivalent numeric encoding, mention/group ordering and duplicates, nil/empty attachments, and server timestamps do not change the fingerprint. Mutable/server-owned document state is excluded, so later edits do not redefine nonce ownership. Callers may not submit fingerprint metadata.

A canonical lookup or persist result is accepted only with exact version `1` and a fingerprint equal to the retry request before any derived effect. Missing, unknown, malformed, or different metadata fails closed. There is no implicit version-0 replay or fingerprint backfill because old documents cannot prove trusted producer or original pre-edit payload. Fully unversioned existing documents follow the fail-closed reservation contract below; partial metadata remains invalid. The pure seam exposes separate domain outcomes for a different original-create payload and a deleted canonical reservation. The pure service mapper below preserves their existing public conflict messages; live invocation, writers, and stores remain separately unbound.

## Phase 151 typed client-nonce outcome contract (2026-08-27)

Canonical result identity and pending/complete state agreement are validated before nonce classification. A lookup hit or a persist result with `Created=false` is an existing canonical record. If that record is deleted, nonce ownership is retained and the pure seam returns `ErrPostCreatePipelineNonceDeleted` before inspecting fingerprint metadata. Otherwise metadata must be version `1` with an exact lowercase 64-hex fingerprint; a well-formed fingerprint that differs from the retry request returns `ErrPostCreatePipelineNonceConflict`.

A persist result with `Created=true` is repository-owned new data. Deleted state or a fingerprint different from the prepared request is therefore `ErrPostCreatePipelineInvalid`, not a user conflict. Missing, unknown, malformed, or noncanonical fingerprint metadata also remains invalid. These sentinels are domain signals only. This outcome atom does not map them to public status/text, bind live send/forward writers, or classify version-0 data; the following adapter freezes mapping and the later legacy contract freezes no-replay classification, both without live invocation.

## Phase 151 post-create service error mapping contract (2026-08-27)

`ExportPhase151MapPostCreatePipelineError` accepts a pipeline error and returns the application error consumed by the centralized Fiber error handler. Nil returns nil. Direct or wrapped `ErrPostCreatePipelineNonceConflict` becomes AppError 409 with the same different-message `client_nonce` text already used by legacy `SendMessage`. Direct or wrapped `ErrPostCreatePipelineNonceDeleted` becomes AppError 409 with the distinct deleted-message reservation text.

Every other value, including `ErrPostCreatePipelineInvalid`, trusted-producer denial, wrapped infrastructure failures, and repository detail, becomes the same fixed AppError 500 save-message error. The original cause is not returned in the AppError message/details. This mapper is only a pure service adapter: no handler or user/forward/bot/poll/promotion writer invokes it yet, so no live HTTP behavior or cutover is claimed.

## Phase 151 legacy client-nonce reservation contract (2026-08-27)

After identity/state validation and deleted precedence, an existing canonical record is fully unversioned only when `client_nonce_fingerprint_version == 0` and the fingerprint is empty. Such a lookup hit or `Created=false` persist race returns `ErrPostCreatePipelineNonceLegacyReserved`, the zero public result, and no derived effect. It is not accepted as an idempotent replay because original producer and pre-edit create payload cannot be reconstructed.

A deleted fully unversioned record remains `ErrPostCreatePipelineNonceDeleted`. A `Created=true` fully unversioned result is an invalid repository invariant. Partial metadata, unknown version, malformed form, and uppercase fingerprint remain invalid. The service mapper translates direct or wrapped legacy reservation to the same generic different-message 409 used for an ordinary fingerprint conflict, avoiding disclosure of storage version. No migration or fingerprint backfill is performed.

## Phase 151 Mongo reply-count reconciliation contract (2026-08-27)

`ReconcileReplyCount(ctx, scope, message)` accepts only a nonnil persisted reply with nonempty distinct message/root IDs, `ChannelID == scope.ChannelID`, `SenderID == scope.ActorID`, no deletion timestamp, pending post-create state, and no completion timestamp. The scope must be a valid trusted producer scope. Any structural or state mismatch returns the typed pipeline validation/authority error before Mongo.

The count predicate is exactly the authorized channel, `thread_id == root ID`, and `deleted_at:nil`. The update predicate is exactly the root ID and same channel with `deleted_at:nil`, plus root shape expressed as missing or empty `thread_id`. The update is `$max: {reply_count: observedCount}` and requires `observedCount > 0`. Mongo error, impossible zero count, or unmatched root fails the effect so the pipeline cannot mark effects complete; a matched root with no modification is an idempotent success.

This contract is currently a concrete repository adapter only. It is not invoked by live producers or recovery, has no live Mongo/integration evidence, does not repair an already excessive counter, and does not make a concurrent create/delete sequence transactional.

## Phase 151 automatic thread-follow effect contract (2026-08-27)

`ExportPhase151BuildAutoThreadFollowPlan` accepts only a trusted-producer, live pending canonical message with exact scope actor/channel, nonempty ID and server `created_at`, no completion timestamp, and no self-root reply. It returns user IDs as author/replier followed by unique nonnil persisted resolved mentions, channel ID, root ID (`message.ID` for root; `thread_id` for reply), and an `auto` transition at the triggering UTC message position.

`ExportPhase151EnsureAutoThreadFollows` requires nonnil context, fresh scope resolver, and transition writer. For each planned user it calls `ResolveAuthorizedUnreadScope`. Direct/wrapped `ErrAuthorizedUnreadScopeDenied` and a valid scope without the exact channel target skip only that user. Any other resolver error, returned tenant/user mismatch, or writer error stops the effect. Only a valid same-tenant/same-user scope containing `{channel_id, kind:channel}` reaches `ApplyAuthorizedThreadFollowTransition`.

`CombinedRepository.EnsureAutoThreadFollows` binds the coordinator to the production Ent resolver and Mongo repository. The existing Mongo operation revalidates the live root and uses an insert-only auto upsert, so explicit manual intent and existing auto state are not advanced or overwritten. Earlier successful candidates may remain after a later error; retry is the convergence mechanism. Live writers/recovery do not invoke this adapter yet.

## Phase 151 post-create unread invalidation contract (2026-08-27)

`PublishUnreadInvalidation(ctx, scope, message)` accepts only a nonnil context, trusted producer scope, and an exact live pending canonical message with nonempty ID and server timestamp, matching authorized channel and actor/sender, no deletion, and no completion timestamp. A reply must name a distinct nonnil UUID root. Invalid input performs no audience lookup or publish.

After the post-create completion CAS succeeds, the effect resolves the current `unread_changed` audience through `WSChannelAuthority`. Audience failure publishes nothing and returns an error. Successful resolution publishes once per unique nonnil current recipient except the sender. Root payload is exactly `channel_id`, `target_kind=channel`, and `category=channel`. Reply payload is exactly `channel_id`, `target_kind=thread`, `category=followed_thread`, and `root_message_id`. Payloads contain no content or count.

The event publisher is best effort and exposes no publish error. A failure or crash after CAS may therefore lose the invalidation; a later idempotent create retry does not repeat completed effects. The client recovers through subscribe/reconnect and authoritative unread refetch. No transactional outbox, lossless delivery, or exactly-once guarantee is specified. The adapter is composed by the standalone complete post-create repository but is not bound to production writers.

## Phase 151 eligible mention resolution contract (2026-08-27)

`ResolveEligibleMentionUserIDs(ctx, scope, request)` first validates the trusted producer and prepared root/direct-reply request. With no individual or group mentions it returns an empty list and performs no authority, audience, presence, or role lookup. Otherwise it requires a fresh actor `ReceiveEvent` access result with the same tenant and channel as the trusted scope, plus a current recipient audience for that channel.

Individual mention strings must parse as UUIDs and belong to that audience. `all` selects the audience; `online` selects current Presence IDs intersected with it; and `owner`, `developer`, `tester`, or `viewer` select current project-role members intersected with it. Role groups do nothing for a non-project channel. Unknown groups, malformed/nil IDs, the sender, out-of-audience users, and duplicates are omitted. Output is sorted by UUID string before it is copied into `resolved_mention_user_ids`.

Authority/audience/presence/role errors return no resolved set before Mongo persistence. The current audience implementation may suppress a candidate-specific authorization error, which can omit a non-sender but cannot authorize an extra user. The resolver is composed by the full standalone post-create repository, but no live message producer invokes that repository yet.

## Phase 151 post-create repository composition contract (2026-08-27)

`NewPhase151PostCreateRepository(combined, mentions, invalidation)` requires a nonnil `CombinedRepository`, its nonnil Ent and Mongo members, and plain/typed-nonnil eligible-mention and unread-invalidation delegates. The resulting value must satisfy all seven `Phase151PostCreatePipelineRepository` ports: five promoted persistence/effect methods plus the two explicit delegated methods.

Explicit delegate errors are returned without reclassification. Successful mention results are copied before return so callers cannot mutate delegate-owned state. A nil composition receiver returns `ErrPostCreatePipelineInvalid` from the two explicit methods. The contract supplies no live writer routing or recovery claim/lease/fence behavior.

## Phase 151 post-create recovery claim/completion-fence contract (2026-08-27)

`post_create_recovery_owner_id`, `post_create_recovery_fence`, and `post_create_recovery_lease_until` are server-owned BSON fields. A prepared create request containing any nonzero value is invalid before repository access. Fence zero means never claimed; each successful claim increments the durable `int64` value and completion does not erase it.

The claim plan requires a valid trusted producer scope and an exact live pending canonical message with nonempty ID, matching channel/sender, nonzero creation time, no deletion, and no completion. Worker ID, time, and positive duration are mandatory. Time is normalized to UTC and cannot precede message creation. The predicate admits only an absent lease or a deadline `<= now`; the update sets owner/deadline and increments fence.

The completion plan additionally requires exact tenant/message/channel plus nonnil owner, positive fence, and nonzero lease. Completion is valid only at/after creation and strictly before lease expiry. Its predicate repeats exact owner/fence/deadline and requires the stored deadline both equal the claim and greater than completion time. It sets state/completion timestamp and unsets owner/deadline. `ErrPostCreateRecoveryLeaseLost` classifies local expiry and the concrete Mongo adapter also maps a completion no-match to it.

## Phase 151 Mongo post-create recovery adapter contract (2026-08-27)

`ClaimPendingPostCreateRecovery(ctx, scope, message, owner, now, duration)` builds the frozen claim plan, applies it through atomic return-after storage, and returns `(acquisition, true, nil)` only for a nonnil exact claimed document. No document is `(zero, false, nil)`. Store error remains discoverable through wrapping. Acquired data must preserve message ID, channel, sender, and creation timestamp; remain live pending; contain the requested owner, a positive fence, and a lease equal to the planned UTC deadline. Contradictory or malformed results are invalid.

`Phase151PostCreateRecoveryAcquisition` contains the claim token and post-update canonical message. The adapter copies `resolved_mention_user_ids` and the lease timestamp so recovery-critical effect input cannot be changed through the store result after return. `CompleteClaimedPostCreateRecovery` applies the frozen completion plan; a store miss is `ErrPostCreateRecoveryLeaseLost`, while a store error retains identity.

The concrete Mongo store uses return-after `FindOneAndUpdate`, treats `mongo.ErrNoDocuments` as a miss, and accepts completion only when `MatchedCount == 1`. Producer-scope reconstruction and claimed effect execution are specified by separate standalone contracts below; retry-loop binding and lease renewal remain unspecified.

## Phase 151 durable producer provenance contract (2026-08-27)

New pending canonical messages have `post_create_producer_provenance` with this strict v1 shape: `version=1`, `producer` in `{user, forward, system_bot, poll, promotion_backlink}`, `bot_kind`, and `integration_id`. For `system_bot`, `bot_kind=system` and `integration_id` is non-nil. For every other producer, bot kind is empty and integration ID is nil. The entire subdocument is server-owned; non-nil caller input is invalid.

Preparation copies the exact already-authorized producer proof into the document. Canonical lookup/persist results and concrete Mongo pending persistence require version 1 and exact equality with the request scope. Recovery reconstruction requires fresh valid target authority, a nonempty live pending message in the same channel, a nonnil canonical sender, and valid durable provenance. Its output preserves target tenant/channel, replaces the target actor with the canonical sender, and copies producer/bot/integration fields before rerunning closed producer validation. Claim and completion filters repeat exact subdocument equality; a successful return-after claim must still match the scope, and its returned provenance is defensively copied.

Legacy messages may lack the optional BSON subdocument. The reconstruction helper, provenance-fenced claim plans, and claimed runner reject them, so an outer recovery loop must skip them or route an explicit migration; it may not guess provenance. No content, message type, client nonce fingerprint, sender naming convention, or bot-looking payload may be used to infer provenance.

## Phase 151 claimed post-create recovery runner contract (2026-08-27)

`ExportPhase151ExecuteClaimedPostCreateRecovery(ctx, targetAuthority, scanned, owner, leaseDuration, clock, repository)` requires nonnil dependencies, a live context, fresh valid target tenant/channel authority, a pending canonical scanned message with durable v1 provenance, a nonnil owner UUID, and a positive lease. It reconstructs the original sender/producer scope and expected claim plan before calling `ClaimPendingPostCreateRecovery`.

A clean claim miss returns `(false, nil)` and performs no effects. An acquired result must exactly match the expected message identity/state/provenance, owner, positive fence, and UTC lease, and a second reconstruction must equal the first scope. Any error from that point returns `acquired=true`. The effect order is automatic follows, reply-count reconciliation only when `thread_id` is nonempty, completion-plan validation using a fresh clock value, fenced completion, and unread invalidation. Wrapped repository-stage errors retain identity.

Completion time must be nonzero, at/after message creation, and strictly before the acquired lease. A separate coordinator adapter exposes this runner behind the bounded scan interface; the later passive holder composes the accepted scanner graph and the once adapter exposes callable execution. No production installer or scheduler/process owner constructs a live worker. Lease renewal, outbox, live stores/runtime, and live-writer routing remain absent. Completion precedes best-effort unread publication, so delivery loss remains recoverable only through authoritative unread refetch.

## Phase 151 recovery coordinator adapter contract (2026-08-27)

Historical 2026-08-27 constructor snapshot (superseded): interval admission is no longer a scan-repository method. The current `NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, intervalGate, owner, leaseDuration, clock, repository)` returns nil unless scanner/interval gate/clock/repository are plain and typed-nonnil, owner is a nonnil UUID, and lease is positive. The value satisfies `Phase151PendingRecoveryRepository`. Every method also rejects a nil receiver/context or corrupted dependency set; a cancelled context is returned before delegation.

`TryAcquirePendingRecoveryInterval` validates its input and delegates only to the separate `Phase151PendingRecoveryIntervalGate`; gate false/error/invalid input admit no scan. `ListPendingRecoveryTargets`, `ReauthorizePendingRecoveryTarget`, and `ListPendingPostCreateDocuments` delegate to the narrowed `Phase151PendingRecoveryScanRepository`. Positive interval/limits and nonempty target tenant/channel/kind are mandatory. Reauthorization output must retain the exact target tenant/channel/kind and the closed service principal; no actor is part of this recovery authority.

`ReconcilePendingPostCreateEffects` passes the authorized scope and a value-owned scanned document to the claimed runner with the constructor-frozen worker/lease/clock/repository. Claim miss returns nil so the outer bounded coordinator continues. Runner-stage errors retain identity. The adapter defines no concrete interval store, PG pager, authorization source, Mongo scanner construction, scheduler, lease renewal, or live-writer cutover.

## Phase 151 recovery scanner authority decision (2026-08-27)

Historical 2026-08-27 snapshot (superseded): before authority was frozen, `Phase151PendingRecoveryTarget` carried tenant/channel/kind only and a scanner could not invent member authority. Current accepted boundaries include the closed principal, pure PG reauthorizer, Ent facts source, narrow pager, scanner, passive composition, and callable once adapter. Remaining work is production installation, scheduler/process-lifetime ownership, writer/server/runtime binding, and live-store execution. The standalone per-instance interval gate is not that scheduler or process owner.

## Calls credentials and Messenger completion CAS (2026-08-27)

The internal Calls service now exposes a complete typed `JoinCredential` for provider-neutral consumption; the existing token-only method remains backward compatible. The Jitsi frontend contract is an internal pure validation/projection boundary: it accepts only a self-hosted configured origin, provider `jitsi`, canonical room name, and JWT. It does not add a public join response, HTTP route, iframe, deployment, media session, recording, or provider cutover. LiveKit remains active.

Post-create Messenger completion is an internal Mongo CAS: exact trusted producer provenance, message/channel/sender identity, live pending state, absent completion, and compatible recovery fence are preconditions; the update changes only state to `complete` and UTC completion time. It does not bind production writers. Recovery principal, reauthorizer, scanner, passive composition, and callable once capability are accepted; production installer, scheduler/process owner, live stores/runtime, and client-nonce compatibility are still required before direct-writer cutover.
Production Messenger construction now requires the complete post-create composition through an additive constructor while retaining the legacy constructor for unchanged call sites and tests. The server builds one shared fresh audience authority, eligible-mention resolver, unread invalidator, and concrete repository. Construction performs no effects; user, forward, bot, poll, and promotion writers remain on their current paths.

## Calls room/expiry and Messenger recovery-scan contract (2026-08-28)

For Jitsi, `Call.RoomName` is the canonical provider room claim; blank, whitespace-only, and padded values are invalid, while `Call.ID` remains audit/correlation identity. The internal issued credential reports the exact UTC expiry encoded into its JWT. Commits: `343d1d0d`, `366bb825`.

Messenger pending recovery candidates must match the authorized target channel, `deleted_at:nil`, pending effects state, `completed_at:nil`, and `post_create_producer_provenance.version=1`; results are ordered `created_at ASC, _id ASC` and capped at 500. The scan has no lease predicate because atomic claim acquisition decides lease eligibility (`44b3a95e`). The recovery principal and bounded composition/callable capability exist; `SendMessage`, production scheduler/process owner, live recovery runtime/stores, Jitsi public API/runtime, and provider cutover remain `UNBOUND/NOT_RUN`; LiveKit remains active.

## Typed Jitsi join projection contract (2026-08-28)

`jitsiCredentialIssuer.IssueJoinCredential` (`0a332760`) accepts only the issuer's already validated configuration and server-loaded call/user inputs, invokes private issuance once, and returns `JoinCredential{Provider:"jitsi", RoomName:call.RoomName, ServiceURL:validated self-hosted public URL, Token:signed token}`. Any failure returns the zero result; legacy token issuance is unchanged. Provider factory binding, public route/DTO, frontend/deployment integration, and cutover remain `UNBOUND`; LiveKit remains active.

## Inert Jitsi host and Messenger recovery-page contract (2026-08-28)

`JitsiEmbedHost` (`ee0387e5`) accepts only a credential that passes the existing strict self-hosted contract and renders exactly one empty marked mount with a forwarded ref. Invalid/null credentials render nothing. Credential values are not stored or emitted, and the component performs no network, storage, logging, script, iframe, provider API, or runtime work. It is not bound to `CallPanel` or a route; LiveKit remains active.

`Phase151ClaimedRecoveryCoordinatorAdapter` (`ce46a0b2`) validates positive limits and owned input cursors before scanning, then requires an in-limit structurally valid page strictly ordered after the cursor by tenant UUID bytes, channel UUID bytes, and kind. Nonnil `Next` must be valid and equal the last target. Errors return a zero page; valid output is defensively copied. The closed principal and narrow Ent pager exist, and the current passive holder composes reauthorizer -> scanner -> process-local gate -> claimed coordinator; production scheduler/process ownership, runtime installation, and `SendMessage` remain `UNBOUND`.

## Messenger server-owned recovery authority contract (2026-08-28)

`Phase151RecoveryServicePrincipal` is closed and currently admits only `messenger.post_create_recovery`. `Phase151RecoveryTargetAuthority` consists exactly of `Principal`, `TenantID`, `ChannelID`, and `Kind`; it deliberately does not represent an actor, producer, bot, or integration. Validation accepts only that closed principal and a structurally valid exact target tuple; every invalid principal, malformed target, or identity drift maps to `ErrPostCreatePipelineInvalid`.

This is a prerequisite contract, not a production recovery API: scanner/runner still consume `AuthorizedMessageProducerScope`. Historical lack of cross-store composition is superseded by the passive holder, which composes the narrow Ent pager path and pure reauthorizer through scanner -> process-local gate -> claimed coordinator, plus the callable once adapter; it installs no production runtime. Scheduler/process ownership, live writer, and external/runtime binding remain `UNBOUND`.

## Calls truthful join-credential expiry contract (2026-08-28)

Internal `JoinCredential` includes `ExpiresAt`. For Jitsi it is the exact captured UTC instant signed into JWT `exp`; for LiveKit it is read only after verification of the just-issued JWT with the configured secret and normalized to UTC. LiveKit parse, signature, nil-claims, missing-exp, or zero-exp failure returns a zero credential and fixed non-secret error. No client-calculated or server-approximated expiry is authoritative.

`CallResponse`, handlers, routes, refresh semantics, guest policy, and frontend contracts are unchanged. LiveKit remains active; Jitsi public/API/runtime/deployment binding remains `UNBOUND`.

## JITSI-03A future authenticated join-credential API contract (2026-08-31)

The recommended additive API is `POST /calls/:callId/join-credential`. After global API-key/generic auth and `RequireInteractiveSession` succeed, any nonempty decoded body, nonzero `Content-Length`, or `Transfer-Encoding` is `400` before Calls service; the client supplies no room, provider, name, URL or token. A complete success has no extra top-level fields: canonical UUID `call_id` equal to the path; closed `provider=livekit|jitsi` equal to the immutable persisted Call provider; server-derived nonempty `room_name` and exact validated `service_url`; nonempty opaque `token`; and nonzero RFC3339 UTC `expires_at`. It sets `Cache-Control: no-store`. `CallResponse`, `/join`, `/guest-join`, `/invite/:code`, the LiveKit guest flow and LiveKit UI retain identical contracts.

Dedicated `RequireInteractiveSession` runs after global `APIKeyAuth` and generic Calls `authMw`, independently verifies Bearer, requires structurally valid non-nil UUID `uid` and `sid`, repeats the active-account check and never treats Locals as authority. No active persisted-session lookup/revocation is claimed. Invalid/revoked API keys retain global `401`; valid API-key-only and valid API-key-plus-Bearer requests follow current API-key precedence and are rejected `403` before Calls service. Without an API key, missing/malformed/expired Bearer or invalid `uid`/`sid` is `401`; account-status dependency failure is pre-service `503`. Malformed `callId` is `400`, Call authorization denial `403`, and other Calls projection/repository/config/issuer failures `500`; Calls-level `404`/`503` await a separately accepted non-enumerating taxonomy.

The handler invokes only `IssueJoinCredentialForAuthenticatedUser`. `CALL-DUR-08C` may allocate the required unknown connection epoch, but issuance creates no participant, joined/confirmed-presence or outbox effect and returns no credential field on error. Token, room and service URL are prohibited from errors, logs, audit and telemetry. `JITSI-03A` makes no guest-path change: accepted `CALL-DUR-08B/08D` must already deny a Jitsi-backed legacy guest request before provider credential issuance while retaining the unchanged LiveKit guest path.

This contract is `BLOCKED_NONDISPATCHABLE` until the sole-oracle queue and accepted dependencies complete in exact order `MPR-01 -> JITSI-02A -> CALL-DUR-01 -> CALL-DUR-03 -> CALL-DUR-02A -> CALL-DUR-02B -> CALL-DUR-02C -> CALL-DUR-02M -> ENT-CALL-GEN-01 -> CALL-DUR-08A -> CALL-DUR-08B -> CALL-DUR-08C -> CALL-DUR-08D`, followed by a separate exact `TaskContract` and sole discovered oracle handoff. Future ownership is `backend/internal/auth/middleware.go`; `backend/internal/calls/{handler.go,handler_calls.go,models.go}`; `backend/cmd/server/wiring_integrations.go`; `backend/tests/auth/phase150_interactive_session_middleware_test.go`; `backend/tests/calls/phase150_join_credential_handler_test.go`; `backend/docs/{docs.go,swagger.json,swagger.yaml}`; and `docs/wiki/API.md`. It consumes the existing `service.go` method without editing that file. Three-agent council alignment is decision input; formal verifier provenance is `UNVERIFIED`. Client/endpoint implementation is `ABSENT/UNBOUND`; OpenAPI generation, execution, Docker and runtime are `NOT_RUN`.

## CALL-DUR-08A provider-room resolver and Ent image gate (2026-08-31)

`ENT-CALL-GEN-01` is one required `BLOCKED_AUTHORITY_AND_RESOURCES` image refresh for accepted CALL-DUR-01 provider fields plus 02A/02B/02C durable schemas; dependent 08A is blocked too. Build-time `go generate ./ent` creates image-contained packages, while normal backend-unit overlays only schema sources; old-image evidence cannot prove generated APIs. No host generation or ignored generated-file Git manifest is accepted. A single founder-authorized rebuild receipt must prove host `backend/ent` contains only tracked `generate.go` and `schema/**`, then bind the complete accepted schema/source/generator set to successful generation sentinels, local image ID and subsequent canonical backend-unit evidence.

After that receipt, future `CALL-DUR-08A` owns exactly `models.go`, `repository.go` and `call_provider_repository_test.go`. Its persisted provider type accepts only `livekit|jitsi`; `models.go` also owns exact sentinel `ErrCallProviderLookupInvalid`. `GetCallByProviderAndRoomName` returns that sentinel before query for nil/typed-nil context, nil repository/client, invalid provider, empty room, or `strings.TrimSpace(roomName) != roomName`; then it uses exact generated provider AND room predicates with no fallback and `.Only(ctx)`. It resolves active or ended Calls, preserves globally unique room names, returns Ent not-found/not-singular and driver/dependency errors without wrapping, and adds no tenant/channel/deleted/status condition.

The `hubtest` oracle uses an in-process fake Ent Postgres dialect driver without real/SQLite database, filesystem, network, `enttest` or new dependency. It proves exact predicates/arguments, generated row mapping, no-row not-found, multi-row not-singular, unwrapped driver-error identity and exact invalid-input sentinel/no-query behavior. It does not prove migration/default/backfill, DB-trigger immutability, planner/index, callback authentication/routing, tenant isolation, provider selection or runtime. The provider-room tuple is resolver data after trusted ingress, never authn/authz or delivery identity.

## CALL-DUR-08B immutable registry contract (2026-08-31)

`ConferenceProviderRegistry` is constructed once from one or more `RegisteredConferenceProvider` values and has no mutating registration API. Valid IDs are exact `livekit|jitsi`. Every entry owns the existing `CallJoinCredentialIssuer`, ID and capabilities. Known capability bits are call-aware credential, room provision, room delete, guest credential and participant metadata; narrow Go interfaces make every advertised operational bit structurally verifiable.

Construction returns exact invalid or duplicate sentinels for empty input, nil/typed-nil entry, invalid ID, missing credential base, unknown bits, duplicate ID or capability/interface mismatch. When the receiver is nil or `r.providers == nil`, `Lookup` and `Require` return exact registry-invalid before ID or capability validation; the zero-value registry is invalid. On a valid constructed registry, `Require` validates provider ID before capability, so jointly invalid ID/capability returns exact unknown-ID. Known-not-registered remains distinct. `Require` otherwise accepts exactly one known bit and returns exact capability-unavailable when absent. Resolution is deterministic, makes no provider call, never falls back and emits no registered value, secret or credential in errors.

`LiveKitConferenceProvider` remains the full legacy provider and advertises all capabilities. `JitsiConferenceProvider` validates `JitsiConfig`, composes the existing issuer and implements only the trusted call-aware credential capability. It has no legacy provision/delete/room-string join/guest/participant-state method. `NewConferenceProvider`, config validation, configured service and server wiring are unchanged and continue rejecting active Jitsi. 08D owns persisted provider mapping, call-lifecycle routing and guest-denial/HTTP-mapping prerequisites; active-provider selection and cutover remain Phase 152. Credential allocation remains 08C.

The sole `hubtest` oracle is constructor/interface/lookup-only and performs no credential issuance, DB, filesystem, network or generated-provider access. Docker evidence remains `NOT_RUN` until the ordered dependency and resource gates accept a unique RED oracle.

## CALL-DUR-02 durable schema source contract (2026-08-31)

`CallParticipantConnection` owns immutable UUID `call_id`, `user_id`, and `epoch`; enum `state=unknown|connected|left` default unknown; optional connected/left timestamps; created/updated timestamps; existing positive default-one VersionMixin; unique `(call_id,user_id,epoch)`, partial left-state retention and user lookup indexes. `CallParticipantPresence` owns immutable UUID call/user; signed-int64 `connected_count` and `presence_sequence` default zero with nonnegative validators; created/updated timestamps; unique `(call_id,user_id)` and user lookup indexes. Each owns exact required/unique/immutable unidirectional `call` and `user` edges bound to scalar IDs: Call cascade and User no-action, with no reverse parent edit. They have no legacy participant relation, provider, delivery, token, outbox or event fields.

Connection timestamp truth is `unknown=(NULL,NULL)`, `connected=(NOT NULL,NULL)`, and `left=(connected_at nullable,left_at NOT NULL)`. Null-connected left is allowed only after reconciliation resolves a never-confirmed epoch; a direct unknown-left observation remains reconciliation-only and persists no transition. `left_at >= connected_at` is required when both exist. Presence has no mixin.

`CallProviderDelivery` owns immutable enum provider `livekit|jitsi`, immutable nonempty max-255 provider `delivery_id`, immutable kind `connected|disconnected`, immutable required UUID call/epoch, mutable optional nillable UUID user, result `pending|applied|ignored|reconcile` default pending, immutable received time and optional terminal time. Normal active-user admission requires a nonzero user; null is reserved for any terminal GDPR tombstone, while missing-user admission is specifically terminal ignored/null. Its only unique identity is `(provider,delivery_id)`; indexes are `(result,terminal_at)` and `(user_id)`. It stores no raw envelope/payload/digest, HMAC/key/certificate, room/raw identity, JWT/token, error/trace or outbox field, and it has no foreign key.

Delivery truth is `pending|reconcile => terminal_at IS NULL AND user_id IS NOT NULL/nonzero` and `applied|ignored => terminal_at IS NOT NULL AND (user_id IS NULL OR user_id is nonzero)`. Reconcile is unresolved and not retention-eligible; the schema has no mixin.

02M adds exact PostgreSQL tables for existing databases, enum/check/index shapes including all user lookup indexes, nonzero UUID and nonnegative/positive checks, state/timestamp and delivery result/terminal/user consistency, and connection/presence Call-cascade plus User-no-action FKs matching the child-owned Ent edges. SQL `now()` timestamp defaults are insert-only; no update trigger exists. Hard Call deletion may remove connection/presence before their normal 30-day cleanup; delivery persists independently as a replay barrier. No data is backfilled or inferred from `CallParticipant`. Every source atom has one external descriptor-only oracle and can reach only `SCHEMA_SOURCE_ACCEPTED`; one later authorized ENT-CALL-GEN-01 build must generate the entire accepted 01/02 set before any generated API claim.

The SQL checks implement those exact connection timestamp/order and delivery terminal/user truth tables. Delivery retention begins only for `applied|ignored`; unresolved `reconcile` rows remain durable.

After generation, a PostgreSQL parity gate compares clean-DB Ent DDL and 02M by columns, FK targets/actions and index semantics. Durable activation remains blocked until admission and erasure share User-first serialization: admission takes `FOR KEY SHARE` on the present User before a nonzero insert or, if deletion's first-operation `FOR UPDATE` wins and the User is absent, applies no presence effect and writes only terminal ignored/null; GDPR then erases linked rows and hard-deletes the User in that transaction. This prevents post-cleanup UUID reintroduction. Any failure rolls back all steps; concurrent integration must prove ingress-before-delete cleanup, ingress-after-lock no-effect tombstoning and tombstone replay no-op. Until then User no-action intentionally rejects hard-delete instead of silently losing linked history.

Both child schemas use literal immutable `edge.To` field bindings; `edge.From`, `Ref` and reverse parent edits are forbidden. 02M appends to CALL-DUR-01's one `BEGIN...COMMIT` and uses plain fail-closed DDL without existence/catalog guards or drops. Exact tables are `call_participant_connections`, `call_participant_presences`, and `call_provider_deliveries`; PostgreSQL UUID/VARCHAR/INTEGER/BIGINT/TIMESTAMPTZ shapes, checks and tuple/partial/user/dedupe/result indexes must match the Phase-150 plan, including delivery user lookup.

## Closed recovery reconstruction and call-aware issuer contract (2026-08-28)

`ExportPhase151ReconstructPostCreateRecoveryProducerScopeFromTargetAuthority` accepts only closed `Phase151RecoveryTargetAuthority` facts and a pending canonical message. It validates the authority first, derives tenant/channel only from authority, and derives sender plus producer metadata only from message provenance; invalid canonical data returns `ErrPostCreatePipelineInvalid` and invalid provenance retains `ErrTrustedMessageProducerDenied`. It does not perform fresh database authorization. The later scanner, passive composition, and callable once adapter are accepted, but production migration/installation, scheduler/process ownership, writer cutover, live stores, and recovery runtime remain `UNBOUND/NOT_RUN`.

`CallJoinCredentialIssuer` is internal-only: it structurally validates server-loaded active call/identity inputs and delegates to the existing LiveKit or Jitsi issuer. It does not load a call, authorize RBAC/membership/invite/guest policy, mutate state, or expose a service/factory/HTTP/client contract. LiveKit remains active; Jitsi self-host deployment, public route/DTO, client binding, cutover, and browser/media runtime remain `UNBOUND`.

## Recovery-authority migration and Calls projection contract (2026-08-28)

Recovery reconciliation consumes `Phase151RecoveryTargetAuthority`; the coordinator creates the closed service authority and the runner reconstructs the canonical effect scope both before and after claim. `AuthorizedMessageProducerScope` is restricted to validated Mongo-scan input. The accepted in-memory scanner adapter is not a fresh authorization or runtime binding; production cross-store scanner composition/binding of the narrow Ent pager and pure reauthorization, plus scheduling, writer cutover, and recovery runtime, remain `UNBOUND`.

`CallJoinCredentialProjector` has only an internal port, constructor, and method. Its required inputs are an active exact server-loaded `Call`, channel-member or instant-creator policy, server-known name, and one typed issuer invocation. Ent source implementation, service/`JoinCall`/HTTP, participant/factory/client, Jitsi cutover, and tenant claim are absent. LiveKit remains active; Docker evidence is `NOT_RUN` because Mail owns Docker and no earlier receipt survives the invalidated overlap.

## Fresh isolated Docker unit/race receipt (2026-08-28)

This entry supersedes the latest NOT_RUN execution status only for the named
source-test atoms below. After a confirmed 60-second empty Docker window, one
docker compose -f docker-compose.local.yml run --rm --no-deps backend-unit
run completed with target PASS, first exit 0, and build/vet/unit/race all PASS;
monitoring found no Mail overlap. The Calls issuer/projector and Messenger
authority-reconstruction, claimed-runner, and scan-proof tests each passed
under both unit and race.

The earlier run was stopped because of Mail overlap and remains invalid history;
no receipt is inherited from it. This is unit/race evidence only. All existing
UNBOUND, runtime, deployment, browser, and cutover statements remain unchanged.

## Calls membership argument-order contract (2026-08-28)

`CallJoinCredentialProjectionSource.IsChannelMember(ctx, channelID, authenticatedUserID)` uses channel-first repository ordering. For a channel call, `CallJoinCredentialProjector` must pass the server-loaded call channel as `channelID` and the authenticated caller as `authenticatedUserID`; denial or source error still short-circuits name lookup and credential issuance. This signature clarification does not bind an Ent/DB source, service, HTTP route, factory, client, tenant claim, Jitsi deployment, or provider cutover. LiveKit remains active.

The focused RED demonstrated the former reversal with exact swapped channel/actor UUID observations in unit and race while build/vet passed. The corrected atom has isolated Docker evidence: target/build/vet/unit/race `PASS`, first exit `0`, no Mail overlap. Browser, media, deployment, integration, and cutover execution remain `UNBOUND/NOT_RUN`.

## Calls projector structural validity contract (2026-08-28)

`CallJoinCredentialProjector.IssueForAuthenticatedUser` returns a zero credential with `ErrCallJoinCredentialInvalid` for a nil or uninitialized projector. A server-loaded call is also invalid when its room name is blank, whitespace-only, or padded; when its nonnil channel pointer contains `uuid.Nil`; or when an instant call has `CreatedBy == uuid.Nil`. These failures occur before membership proof, server-known-name lookup, and credential issuance.

RED-1 proved the previous nil-pointer panic in unit and race; RED-2 proved all five malformed call shapes while the receiver guard was already passing. The final isolated Docker receipt is target/build/vet/unit/race `PASS`, first exit `0`, with no Mail overlap. No public DTO/route/refresh, Ent/DB source, service/factory/client binding, Jitsi runtime, or cutover contract changes; LiveKit remains active.

## Pending post-create scan-plan authority contract (2026-08-28)

`ExportPhase151BuildPendingPostCreateScanPlanFromTargetAuthority(authority, limit)` requires the exact closed `messenger.post_create_recovery` principal, nonzero tenant/channel, nonempty kind, and a positive limit. It returns the legacy candidate filter, ascending `created_at` then `_id` order, and a hard maximum of 500. The API does not accept or derive an actor, producer, bot, or integration identity.

The valid RED failed only on the missing export after build passed. The first GREEN was invalidated by Mail overlap; the fresh isolated rerun passed target/build/vet/unit/race, first exit `0`, without overlap. Persistence, repository/port execution, reauthorization, scheduling, writer cutover, and live DB/RBAC behavior are not specified or proven by this pure builder.

## Calls Ent projection-source contract (2026-08-28)

`EntRepository.LoadActiveCall(ctx, callID)` accepts a nonnil repository/client/context and nonzero ID, then selects only the exact call with active status. `GetServerKnownCallUserName(ctx, userID)` applies the same structural guards, selects only the exact active user, and returns trimmed `DisplayName`, otherwise trimmed `Name`; an empty result is `ErrCallJoinCredentialInvalid`. Plain and typed-nil contexts fail before Ent access, while Ent query errors propagate unchanged.

The type satisfies `CallJoinCredentialProjectionSource` but is not production-wired. Positive Ent query/fallback semantics and live PostgreSQL remain `NOT_RUN`; no route, DTO, service, factory, provider/client, Jitsi runtime, deployment, or cutover contract changes. LiveKit remains active.

## Phase 151 pending recovery closed-authority contract (2026-08-28)

Phase151PendingRecoveryRepository.ReauthorizePendingRecoveryTarget(ctx, target) and the corresponding scan repository operation return Phase151RecoveryTargetAuthority.

ListPendingPostCreateDocuments(ctx, authority, limit) and ReconcilePendingPostCreateEffects(ctx, authority, document) accept no actor, producer, bot, integration identity, or mutable target argument.

The coordinator calls ExportPhase151ValidateRecoveryTargetAuthority(Phase151PostCreateRecoveryServicePrincipal, target, authority) before a scan. A reauthorization error or invalid/drifting principal, tenant, channel, or kind causes zero scan and zero reconcile.

The adapter repeats validation and forwards the same authority unchanged to its scanner and ExportPhase151ExecuteClaimedPostCreateRecovery.

The claimed runner is the sole recovery layer that reconstructs AuthorizedMessageProducerScope from canonical durable message provenance.

MongoRepository.ListPendingPostCreateDocuments builds from ExportPhase151BuildPendingPostCreateScanPlanFromTargetAuthority(authority, limit), preserving existing filter/order/cap behavior.

Kind must remain nonempty. A stricter concrete PostgreSQL channel-type projection is deferred. The passive holder now composes the narrow Ent pager path and pure PG reauthorizer into its scanner graph, but no scheduler/process owner, production installer, writer cutover, service/server wiring, or live integration is specified by this bounded contract.

## Calls service credential facade evidence (2026-08-28)

Commit `b9faba3c920371f182a7416e039a6f2708954fbb` adds `NewCallServiceWithProviderAndCredentialProjector` as an additive facade: a nil projector returns `ErrCallJoinCredentialInvalid`; the unchanged legacy provider constructor is called once; the supplied projector is stored; legacy constructors remain unchanged. `IssueJoinCredentialForAuthenticatedUser` rejects a nil receiver or absent projector, delegates exactly once to `CallJoinCredentialProjector` with exact context, call, and authenticated-user identity, preserves all five typed credential fields on success, and returns a zero credential with exact delegated error identity on failure. The legacy provider performs zero work on this facade path.

Valid RED container `worktree-master-backend-unit-run-e80469df6d84` finished `2026-08-28T18:40:12+03:00` with target `FAIL`, first exit `1`, build `PASS`, and vet/unit/race failure only because the constructor was absent. Final GREEN `worktree-master-backend-unit-run-866fc0b8b37b` finished `2026-08-28T18:49:00+03:00` with target/build/vet/unit/race `PASS`, first exit `0`, and four empty stderr logs. Independent review returned Spec Compliance `PASS` and Task quality `Approved`, with no Critical/Important/Minor findings; root independently read exact raw Docker receipts after the reviewer could not inspect them.

This is compile/unit/race evidence only. No server/runtime construction uses the facade; `JoinCall`, handlers/routes/DTO, participant lifecycle, provider factory/client selection, tenant authorization, self-hosted Jitsi deployment, provider cutover, integration, and browser/media runtime remain `UNBOUND/NOT_RUN`. LiveKit remains active; Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`.

## Phase 151 PG recovery-target reauthorizer contract (2026-08-28)

`PGPhase151PendingRecoveryTargetReauthorizer` consumes fresh singleton-tenant lookup and exact target facts. It fails closed unless tenant and channel match the request, the channel is live, `public|private` maps to `channel` or `dm|group` to `dm`, and project channels have an active tenant project while projectless channels have the tenant creator. It returns only `Phase151RecoveryTargetAuthority{Principal,TenantID,ChannelID,Kind}` with the closed recovery principal; actor, sender, bot, integration, and Mongo authority are excluded.

This reauthorizer source-port contract does not itself compose the separately implemented Ent recovery-target facts source or narrow Ent pager into a cross-store scan, scheduler/interval owner, production coordinator/repository composition, writer cutover, service/server wiring, or runtime/API surface. Those remain `UNBOUND/NOT_RUN`.

## Phase 151 Ent recovery-target facts contract (2026-08-28)

Commit 4b100ee6 implements EntRepository.ResolvePendingRecoveryTargetFacts for Phase151PendingRecoveryTargetFactSource. It requires a usable repository/client/context plus nonzero tenant and channel IDs before its Ent query; invalid or cancelled inputs fail closed. The source is not a pager, scanner, scheduler, writer, or production composition.

Missing-method-only RED worktree-master-backend-unit-run-e14c5e925814 is superseded by final GREEN worktree-master-backend-unit-run-546c91d7105f at 2026-08-28T20:26:28+03:00: target/build/vet/unit/race PASS, exit 0, four empty stderr logs. Independent review is PASS/Approved without findings; separate predicate review is PASS.

Evidence proves compile/interface and pre-query guards only. Successful query, NotFound, driver/generated SQL, projection, live PostgreSQL semantics, API/runtime wiring, and readiness remain NOT_RUN or UNBOUND as applicable.

## Phase 151 Ent pending-recovery target pager contract (2026-08-28)

Commit `748f4997` adds only `EntRepository.ListPendingRecoveryTargets(ctx, cursor, limit)` as the concrete pager for the narrow recovery-target port; it does not make `EntRepository` a full recovery scanner. Guard order is fixed: usable dependencies/context, `ctx.Err()`, positive and `MaxInt`-safe limit, structural cursor, singleton tenant resolution, then cursor-tenant equality before the single query. The query admits only live `public|private|dm|group` channels whose tenant is active through the channel project, or whose projectless channel creator is in the singleton tenant; it applies optional channel-ID-greater-than cursor, ascending channel ID, and a safe `limit + 1` fetch.

The pager maps returned channel kind, returns an owned page, and emits `Next` only from the last returned target when the extra row exists. Unit tests prove only compile/interface satisfaction and guards before singleton/query access. Cursor-tenant mismatch, singleton cardinality, generated Ent predicates/order/limit, query execution, mapping/pagination/Next, error-chain behavior, and live PostgreSQL behavior remain static-only or `NOT_RUN`. The passive holder composes the pager path through reauthorizer -> scanner -> process-local gate -> claimed coordinator, but scheduler/process ownership, production installation, writer cutover, service/server runtime, and integration remain `UNBOUND/NOT_RUN`.

## Phase 151 standalone pending-recovery interval-gate contract (2026-08-28)

`Phase151PendingRecoveryIntervalClock` and `Phase151ProcessLocalPendingRecoveryIntervalGate` (`cb995211`, cancellation correction `3632c37`) define only per-constructed-instance best-effort admission. The clock is distinct from `Phase151PostCreateRecoveryClock`; no scheduler, coordinator, scanner, singleton, distributed/global interval, exactly-once behavior, or recovery I/O ownership follows. Separate instances may both admit. The mutex protects only the gate check, `Now()` read, and timestamp mutation; it never covers recovery I/O.

The method checks cancellation before locking, after lock acquisition, and after `Now()` before mutation. Nil/typed-nil dependencies, invalid context, nonpositive interval, zero clock value, or backward time fail closed; below-minimum elapsed returns a denial without mutation, while equality/admission-or-later stores the new timestamp. Deterministic coverage includes every boundary, two channel-coordinated cancellation paths without sleeps, concurrent single admission, independent gates, and immediate admission consumption.

Valid missing-contract RED `9611a72a031b` preceded initial GREEN `323c19983456` at `2026-08-28T21:01:26+03:00`. Final-fix root-read receipts at `2026-08-28T21:55:49+03:00` record target/build/vet/unit/race PASS, exit 0, four stderr lengths 0, and **container ID not retained**; scoped re-review is ADDRESSED/PASS/APPROVED with 0/0/0 new breakage. Cross-store scanner/coordinator production migration, scheduler/process-lifetime owner/wiring, runtime/live stores/transports/browser, and integration remain UNBOUND/NOT_RUN.
## Calls configured credential-service composition (2026-08-28)

`NewCallServiceWithConfiguredProvider` is the server-composed internal constructor: `CallsConfig` selects the factory provider; that provider must implement the typed `CallJoinCredentialIssuer`; and `NewCallJoinCredentialProjector(source, issuer)` is passed to the additive Calls facade. Server wiring supplies the same `callsRepo` once as the service repository and once as the projection source. Factory or projector construction failure returns an error, so `wireIntegrations` logs and exits before `callsChatRepo.EnsureIndexes`.

This does not alter the public `JoinCall` contract, routes, DTOs, frontend, recording, or legacy token path. Server runtime is `NOT_RUN`; tenant authorization, Jitsi deployment, provider cutover, integration, and browser/media behavior remain `UNBOUND/NOT_RUN`. LiveKit remains active; Jitsi is rejected before Phase 152.


## Calls confirmed participant-presence contract (2026-08-28)

`ReduceConfirmedParticipantObservation(previous, observation)` is a pure future canonical lifecycle contract for an observation already authenticated and provider-normalized. `ParticipantPresenceUnknown`, `ParticipantPresenceConnected`, and `ParticipantPresenceLeft` are its only valid previous states; `CallID` and `UserID` are nonzero and observation kind is only `Connected` or `Left`. Any other previous/kind or zero ID returns exact `ErrConfirmedParticipantObservationInvalid` and the zero `ParticipantPresenceTransition`.

The transition table is fixed: `Unknown -> Connected` and `Left -> Connected` return next `Connected` with `Persist` and `PublishJoined`; duplicate connected is a no-op. `Connected -> Left` returns next `Left` with `Persist` and `PublishLeft`; `Unknown -> Left` and duplicate left are no-ops. Credential issuance is not presence and cannot emit a canonical joined effect. This is state-based idempotence only, never a durable dedupe or exactly-once guarantee.

Raw webhook verification, provider selection, persistence, NATS, inbox/outbox, transactional application, clocks, delivery IDs, multi-session, guest, refresh/revoke, API, service, repository, schema, server, and frontend bindings are absent. Legacy `/join`, DTO/token/frontend/recording remain unchanged. LiveKit remains active; Jitsi deployment/media/browser/cutover remains `UNBOUND/NOT_RUN`.

## Calls trusted-provider participant-presence mapper contract (2026-08-29)

`NormalizeTrustedProviderParticipantConnection` accepts only `TrustedProviderParticipantConnection{CallID, UserID, Kind}` after the source has authenticated and normalized the provider event, resolved its room to a call, and resolved participant identity to a Hub user. `TrustedProviderParticipantConnected` produces `ConfirmedParticipantObservation{CallID, UserID, Kind: Connected}`; `TrustedProviderParticipantDisconnected` produces the corresponding `Left` observation. A zero call/user ID or a zero/unknown kind returns the zero observation with `ErrConfirmedParticipantObservationInvalid`.

It is not a reducer invocation, webhook verifier, presence writer/publisher, or integration boundary. It accepts no raw provider event, room, identity, JWT, provider, token, or delivery ID and performs no context, clock, network, repository, NATS, membership, provider-selection, or session-identity operation. Runtime/wiring, reducer invocation, webhook application, persistence/NATS, durable identity/order/dedupe, Jitsi deployment/media/browser/cutover remain `UNBOUND/NOT_RUN`; LiveKit remains active.

## Phase 151 post-create recovery scanner adapter contract (2026-08-29)

`Phase151PostCreateRecoveryScannerAdapter` is an unbound in-memory sequential composition of exactly a pending-recovery target pager, reauthorizer, and pending-document lister. It satisfies only the existing scanner interface. A nil receiver/context/dependency shape returns `ErrPostCreatePipelineInvalid` and a zero result before consulting `ctx.Err()`; after structural validation, a cancelled context returns that exact error and zero result. Valid calls delegate exactly once to their matching port with original inputs, preserving the exact returned value, error, and aliases, including a nonzero value together with an error. It adds no domain validation, synthesized result, retry, copying, logging, network, concurrency, or I/O.

This scanner-only contract implies no Ent/Mongo execution, production owner/config/scheduler/server/runtime binding, transaction, dedupe, exactly-once semantics, or live-store behavior. The later passive holder now composes reauthorizer -> scanner -> process-local gate -> claimed coordinator, but it is not a production installer; server/scheduler/executor/runtime remain `UNBOUND`, and all live-store and integration evidence is `NOT_RUN`.

## Calls confirmed participant connection key (2026-08-29)

`ConfirmedParticipantConnectionKey` is a pure future internal Hub value containing `(CallID, UserID, Epoch)` for a connection fact already trusted upstream. Its only canonical rendering is `hub.call-connection.v1/<call>/<user>/<epoch>`. Every component must be nonzero; invalid construction yields `ErrConfirmedParticipantConnectionKeyInvalid` and the zero key, `IsValid` returns `false`, and `Canonical` is empty.

Neither the value nor pointer implements `fmt.Stringer`. Tests prevent accidental default stringification only: this is not redaction or a security guarantee, because exported UUID fields remain formattable; runtime operations must continue not to log connection keys. The value neither creates epochs nor parses/serializes data, selects a provider, accepts webhooks, persists/refcounts connections, or binds service/server/runtime/Jitsi code. Existing legacy behavior remains unchanged. LiveKit remains active; Jitsi is `UNBOUND/NOT_RUN`.

## Phase 151 pending-recovery runtime-config contract (2026-08-29)

`Phase151PendingRecoveryRuntimeConfig` is a private-field immutable value for future composition. Its constructor succeeds only for a nonzero owner UUID, positive lease duration, and an existing `Phase151PendingRecoveryConfig` with positive `MaxPages`, `MaxTargets`, `PendingScanLimit`, and `MinimumInterval`; all invalid shapes return nil. Nil receiver accessors return zero owner, zero lease, and zero execution config. The constructor copies its execution config and `ExecutionConfig()` returns a copy, so mutations on either side cannot alter stored state.

This config value does not generate an owner, establish singleton/distributed uniqueness, renew a lease, add caps/defaults, or parse environment/deployment input. The later passive holder composes reauthorizer -> scanner -> process-local gate -> claimed coordinator, but constructs no production factory/installer/server/scheduler/executor/process owner and binds no live stores, writers, runtime, or integration. All such production behavior remains `UNBOUND/NOT_RUN`.

## Calls confirmed participant connection-refcount contract (2026-08-29)

`ReduceConfirmedParticipantConnection` is pure future internal logic over a previously accepted `ConfirmedParticipantConnectionKey`, a trusted confirmed observation, and the current aggregate. It validates all structural inputs before any ended-state branch; invalid input returns its exact invalid sentinel and a zero result. For an active call, first connected moves the aggregate `0 -> 1` and requests joined; final left moves it `1 -> 0` and requests left. Duplicate connected/left observations are no-ops. A left for an unknown connection requests reconciliation only. An active same-epoch connection cannot transition `Left -> Connected`; a new epoch is required. A valid observation for an ended call is descriptive no-op and cannot create individual-left fan-out.

The contract is not persistence, CAS, transaction, event, reconciliation execution, provider/webhook, or runtime behavior. It performs no I/O or logging and does not alter the earlier simple presence reducer, legacy calls behavior, service/API/schema/server/frontend paths, provider cutover, or Jitsi binding. LiveKit remains active; Jitsi deployment, media, browser, and cutover remain `UNBOUND/NOT_RUN`.

## Phase 151 pending-recovery composition contract (2026-08-29)

`NewPhase151PendingRecoveryComposition(config, postCreateRepository, clock)` returns nil unless `config` is a valid immutable runtime-config value, the post-create repository contains a complete Combined Ent/Mongo graph, and `clock` is plain nonnil (typed-nil is invalid). It composes the existing PG target reauthorizer, scanner adapter, process-local interval gate, and claimed coordinator, forwarding the validated owner UUID and lease duration statically. A nonnil composition returns the same opaque `Phase151PendingRecoveryRepository` identity from `PendingRecoveryRepository()` and a value copy from `PendingRecoveryConfig()`; nil accessors return zero values.

Apart from exactly three config accessor reads (`OwnerID()`, `LeaseDuration()`, and `ExecutionConfig()`), the contract makes no clock, delegate, persistence, or transport method call and performs no I/O or recovery execution. It defines no executor, `RunOnce`, scheduler/goroutine, process owner, server/startup, or production installer. It therefore binds neither live stores nor runtime; production installation and live execution remain `UNBOUND/NOT_RUN`.

## Calls confirmed participant connection state-change contract (2026-08-29)

`BuildConfirmedParticipantConnectionStateChange(transition)` accepts a valid `ConfirmedParticipantConnectionTransition` only after its key, presence states, aggregate counts, persistence flags, reconcile flag, and publication boundary are mutually consistent. The only applied shapes are `Unknown -> Connected` with `nextCount = previousCount + 1`, and `Connected -> Left` with `nextCount = previousCount - 1`; joined is true exactly for `0 -> 1`, while left is true exactly for `1 -> 0`. The resulting private immutable payload supplies typed getters for expected and next connection state, expected and next aggregate count, and the permitted joined/left publication flag.

Descriptive no-ops and `Unknown -> Unknown` reconciliation return the zero payload, `shouldApply=false`, and nil. Any malformed, forged, partial, contradictory, overflow, or unsupported transition returns the exact `ErrConfirmedParticipantConnectionStateChangeInvalid`, zero payload, and `false`. The private `confirmedParticipantConnectionStateChangeStore` describes a future atomic comparison of both expected values and joint commit of both next values; `ConfirmedParticipantConnectionStateChangeCASApplied` is valid only after both updates commit. Conflict/non-application implies neither mutation nor publication.

This is not an adapter or persistence implementation: it adds no schema, database transaction, outbox/event, retry, reload, context semantics, provider/webhook, service/API/server/runtime, or Jitsi binding. Neither value nor pointer implements Stringer, text, or JSON marshalers; this prevents accidental output interfaces only, not formatting through the typed getters, so connection values remain subject to operational no-logging. LiveKit remains active; Jitsi remains `UNBOUND/NOT_RUN`.
## Phase 151 pending-recovery once adapter contract (2026-08-29)

`NewPhase151PendingRecoveryOnceAdapter(repository, config)` succeeds only for a plain or typed-nonnil existing `Phase151PendingRecoveryRepository` and positive bounded configuration, retaining the configuration as a value copy. `Execute` returns exact `ErrPostCreatePipelineInvalid` with zero result for invalid receiver/context/dependencies before consulting `ctx.Err()`. After structural acceptance, cancellation returns exact `ctx.Err()` before executor delegation. A valid request delegates exactly once directly to the existing executor, without an added wrapper or retry; gate denial ends before scan and existing wrapped gate errors retain `errors.Is` identity.

This contract adds no scheduler, ticker, goroutine, process owner, server startup/shutdown, production composition/wiring, live-store execution, or integration. It is a callable capability only; production installation remains `UNBOUND/NOT_RUN`.
