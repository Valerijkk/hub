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
| NFR-30 | Frontend test gate | Vitest **99+** tests; Playwright **20** E2E specs ✅ |

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
Source FR: FR-INTR-02, FR-INTR-03

### POST /api/intranet/news/:newsId/poll/vote
Auth: Bearer
Request: { "option_id": "uuid" }
Responses: 200 {агрегаты как в GET}; 400 (option чужого poll); 401; 404; 422 (повторный голос / poll closed); 429
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
| Frontend Vitest | 0 | **94** files (`testframework/stages/unit/frontend-test/`, measured 2026-07-28) |
| Test framework migration | in progress | **488 / 0 internal** ✅ |
| Lighthouse CI | missing | **artifact job** (PR #27) |
| Squash PR v2.5-trust | pending | **merged PR #1** ✅ |

---
*Последнее обновление: 2026-06-23 — Wave C complete, COV 83.0% (2026-06-22 Docker baseline), FA-06 manual blocker, Lighthouse CI*
