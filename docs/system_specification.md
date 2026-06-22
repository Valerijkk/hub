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
| Intranet | INTR | News, directory, kudos, random coffee |
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
