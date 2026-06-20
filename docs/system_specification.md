# Hub — System Specification

> Единая база знаний. Обновляется SA при изменении требований.
> Детали: `docs/TECHNICAL.md`, API: `docs/wiki/API.md`, Swagger: `backend/docs/`

---

## Функциональные требования (сводка)

Полный список: `.planning/PROJECT.md`, `.planning/REQUIREMENTS.md`

| Модуль | ID prefix | Ключевые возможности |
|--------|-----------|---------------------|
| Auth | AUTH | Magic Link, Telegram, JWT, sessions, onboarding |
| Project | PROJ | CRUD, RBAC, invites, teams, templates, archive |
| Tracker | TASK | 9 типов, workflow, links, comments, time tracking |
| Kanban | KANB | Columns, WIP, swimlanes, filters |
| Scrum | SCRM | Sprints, burndown, velocity, planning poker |
| Messenger | MSG | Channels, DMs, threads, reactions, voice messages |
| TMS | TMS | Cases (classic/BDD), runs, environments, reports |
| Allure | ALLU | CI parser, matching, notifications |
| Calls | CALL | LiveKit voice + screen share, recording |
| Dashboard | DASH | Personal/project/admin widgets |
| Notification | NOTF | In-app, Telegram, Web Push, quiet hours |
| Search | SRCH | Ctrl+K, JQL, saved filters |
| GitHub/GitLab | GHUB/GLAB | Webhooks, PR/MR linking |
| API | API | REST **285** endpoints (+1 WS), API keys, webhooks |
| Import/Export | IMEX | CSV, JSON, PDF, ZIP backup |
| Audit | AUDIT | Full audit log, trash, GDPR |

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
| NFR-22 | Security headers | CSP, HSTS, X-Frame-Options — Phase 16 (v2.5) |
| NFR-23 | Refresh token storage | Hash в Redis, не plaintext — Phase 16 (v2.5) |
| NFR-24 | Optimistic locking | `version` int на каждой изменяемой Ent |
| NFR-25 | Primary keys | UUID v4 only — string IDs запрещены |
| NFR-26 | Pub/sub transport | NATS JetStream only — Redis pub/sub запрещён |
| NFR-27 | File upload path | Presigned PUT в SeaweedFS; backend buffer max 25 MB |
| NFR-28 | Tenancy | Single-tenant (1 instance = 1 team) |
| NFR-29 | Test coverage gate | ≥50% overall, ≥80% security-critical paths (Phase 18) |
| NFR-30 | Frontend test gate | Vitest unit + Playwright E2E на критических путях (Phase 18) |

## Модель данных (ER — ключевые сущности)

### PostgreSQL (Ent)

```
User ──┬── Session
       ├── ProjectMember ── Project ──┬── Issue ──┬── IssueLink
       │                              │           ├── Comment
       │                              │           ├── Label
       │                              │           └── CustomFieldValue
       │                              ├── WorkflowStatus
       │                              ├── Board
       │                              ├── TestCase ── TestRun ── TestResult
       │                              └── Channel (ref)
       └── NotificationPreference

Role (system: Admin/Member)
ProjectRole (Owner/Developer/Tester/Viewer + custom RBAC)
Team ── TeamMember
InviteLink
AuditLogEntry
TrashItem
```

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
| POST | `/auth/magic-link` | Отправить Magic Link |
| POST | `/auth/verify` | Verify code → tokens |
| POST | `/auth/telegram` | Telegram auth |
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

### Endpoint inventory (2026-06-17)

| Категория | Count | Примечание |
|-----------|-------|------------|
| Documented in `docs/wiki/API.md` | **280** | Production routes (без `/api/dev/*`) |
| Registered in handlers (`RegisterRoutes`) | **283** | 280 production + 3 dev-only |
| Infra (`/docs`, `/docs/doc.json`, `/health`) | **3** | `cmd/server/main.go` |
| **Total REST `/api/*`** | **286** | incl. ban endpoint (audit +1) |
| WebSocket `GET /ws` | **1** | Вне префикса `/api` |
| Dev-only `/api/dev/*` | **3** | `DEV_MODE` — не в release binary |

**По модулям (handler routes):** auth 11 · project 22 · templates 5 · rbac 6 · storage 4 · sprint 14 · tracker 38 · messenger 42 · tms 50 · search 5 · notification 10 · audit/admin 11 · dashboard 7 · github 9 · gitlab 7 · calls 21 · apigateway 8 · backup 4 · imex 6 · dev 3.

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

### Аудит SUMMARY.md (2026-05-15) — 57 findings

| Severity | Count | Production impact |
|----------|-------|-------------------|
| CRITICAL | 12 | Blockers — см. reconciliation в COVERAGE.md |
| HIGH | 16 | SSRF, IDOR, coverage, frontend auth/bundle |
| MEDIUM | 21 | Layering, a11y, Repository pattern gaps |
| LOW/INFO | 8 | Docs drift, unpinned images |

**Post-audit code status (2026-06-17):** ~9 из 12 CRITICAL исправлены в коде (SEC-01..07, INF-01/02/04, FE-01). Требуется formal re-audit (SA_VER + security-auditor). Остаются: **INF-03** (secret rotation), **COV-01** (11.5% coverage), **NFR-16/17** (a11y/Lighthouse).

### Критические пробелы покрытия тестами

| Область | Текущее | Цель | Gap |
|---------|---------|------|-----|
| Backend statement coverage | 11.5% | 50% | Phase 18 |
| Security-critical packages | <<80% | 80% | auth refresh, ws authz, nats, storage |
| Frontend Vitest | 0 files | module unit tests | Phase 18 |
| Playwright E2E | not present | 5+ critical paths | Phase 18 |

### Wiki drift

- `docs/wiki/API.md` синхронизирован с кодом: **280** production routes (+3 dev-only не документированы) = **283** handlers; **286** REST `/api/*` с infra (Stage 4.4, 2026-06-17).

---
*Последнее обновление: 2026-06-17 — SA gap matrix + endpoint recount*
