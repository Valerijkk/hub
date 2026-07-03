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
| API | API | REST **286** endpoints (+1 WS), API keys, outbound webhooks |
| Import/Export | IMEX | CSV, JSON, PDF, ZIP backup, Jira import |
| Audit | AUDIT | Full audit log, trash 30d, GDPR, alerts |
| AI | AI | Multi-provider gateway, @hub-ai bot, JQL assist, usage audit |
| Intranet | INTR | News, directory, kudos (+leaderboard spec), random coffee (+auto-DM spec), poll-in-post spec, onboarding spec (FR-INTR) |
| Calendar | CAL | Unified events (calls, sprints, leave) |
| CRM | CRM | Contacts, deals (lite) |
| HR | HR | Leave requests |
| Accounting | ACCT | Ledger, journal entries, invoices |
| Bases | BASE | No-code DB (HubBase/Table/Record) |
| Marketplace | MKT | Internal listings |
| Mail | MAIL | hub-mail admin, transactional settings |
| Status | STAT | Public status page |
| 1C | ONEC | Optional connector (`integrations/onec`) |

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
| NFR-13 | Error format | `{"error": "...", "request_id": "uuid"}` |
| NFR-14 | Pagination tasks | offset/limit |
| NFR-15 | Pagination chat | cursor (infinite scroll) |
| NFR-16 | Accessibility | WCAG 2.1 AA |
| NFR-17 | Lighthouse score | ≥ 90 (цель v2.5) |
| NFR-18 | WebSocket room authz | Subscribe только при membership/RBAC (NF-02) |
| NFR-19 | Search user-scope | Фильтрация по project/channel membership (NF-03) |
| NFR-20 | Webhook signature | LiveKit/GitHub: body hash обязателен, skip запрещён (SEC-02) |
| NFR-21 | Outbound HTTP safety | `SafeHTTPClient` + `ValidateOutboundURL` для user URLs (SSRF) |
| NFR-22 | Security headers | CSP, HSTS, X-Frame-Options — ✅ `middleware.SecurityHeaders()` |
| NFR-23 | Refresh token storage | Hash в Redis, не plaintext — ✅ |
| NFR-24 | Optimistic locking | `version` int на каждой изменяемой Ent |
| NFR-25 | Primary keys | UUID v4 only — string IDs запрещены |
| NFR-26 | Pub/sub transport | NATS JetStream only — Redis pub/sub запрещён |
| NFR-27 | File upload path | Presigned PUT в SeaweedFS; backend buffer max 25 MB |
| NFR-28 | Tenancy | Single-tenant (1 instance = 1 team) |
| NFR-29 | Test coverage gate | ≥50% overall ✅ **83.0%**; ≥80% security-critical ✅ **6/6** |
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
| GET | `/projects/:id/issues?offset=0&limit=50` | Paginated issues |
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
| Infra (`/docs`, `/health`, `/metrics`) | **+3** | `cmd/server/main.go` |
| **Total REST `/api/*`** | **286** | Wiki synced |
| WebSocket `GET /ws` | **+1** | Вне префикса `/api` |
| Dev-only `/api/dev/*` | **3** | `DEV_MODE`, не в `-tags release` |

**По модулям (handler routes, выборочно):** auth 11 · project 24 · tracker 48 · messenger 44 · tms 52 · calls 24 · search 9 · notification 18 · ai 10+ · intranet 19 · bases 13 · accounting 13 · crm 10 · hr 8 · calendar 4 · marketplace 12 · github 13 · gitlab 11 · sprint 16 · dashboard 9 · audit 13 · apigateway 10 · backup 6 · imex 8 · storage 6 · rbac 8 · status 3 · mail (admin) · onec 5.

### WebSocket Protocol

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
> разделы «Волна 1 аналитика»). Границы scope: `.planning/GRAY-ZONES-DECISIONS-2026-07-03.md`
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
| FR-INTR-11 | US-INTR-05 | Для каждой пары авто-DM в messenger с приветственным сообщением; существующий DM переиспользуется (новое сообщение, не новый канал). Участники получают notification. Выключивший участие до запуска не матчится | NFR-26 |

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
Request: { "question"?, "options"?, "closed"?, "version" }
Responses: 200; 401; 403; 404; 409 (version); 422 (правка options после первого голоса)
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
Source FR: FR-INTR-05  · Пагинация: offset/limit (NFR-14)

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
Source FR: FR-INTR-09  · Пагинация: offset/limit (NFR-14)

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

## Известные gaps (требование → код → тест)

Полная матрица: `.planning/COVERAGE.md`

### Актуальные пробелы (2026-06-23)

| Область | Статус | Gap |
|---------|--------|-----|
| Backend coverage overall | **83.0%** ✅ | Gate ≥82% PASS |
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
| Overall coverage | 11.5% → 56.6% | **83.0%** |
| Playwright E2E | not present | **20 specs** |
| Frontend Vitest | 0 | **99/99 tests** |
| Test framework migration | in progress | **488 / 0 internal** ✅ |
| Lighthouse CI | missing | **artifact job** (PR #27) |
| Squash PR v2.5-trust | pending | **merged PR #1** ✅ |

---
*Последнее обновление: 2026-06-23 — Wave C complete, COV 83.0%, FA-06 manual blocker, Lighthouse CI*
