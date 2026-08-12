# E5 — Locust hot-path scenarios (load layer)

**Дата:** 2026-07-27  
**Источник:** [`02-TEST-FRAMEWORK-MIGRATION.md`](02-TEST-FRAMEWORK-MIGRATION.md) §Э6; audit [`.planning/audits/grok-2026-07/13-perf.md`](../grok-2026-07/13-perf.md) A13-001…006  
**Статус:** design — **`tests/load/**` не создавать здесь**

---

## Цель

Зафиксировать первый набор Locust-сценариев на **горячих путях** продукта: login, лента/dashboard, search, issues/board — плюс пути, которые audit A13 пометил как unbounded / N+1.

Нагрузка **не** в default PR CI; отдельный schedule / manual (открытый вопрос в миграции §«Открытые вопросы» #3).

---

## A13 → load coverage

| Finding | Hot path | Severity | Locust scenario ID | Что меряем |
|---------|----------|----------|--------------------|------------|
| **A13-001** | `GET` board / Kanban `GetBoardData` unbounded `WithIssues` | P1 | `L-BOARD-01` | p95 latency, payload size, error rate при N users на один project |
| **A13-002** | Bases `GetTable` full load + write amplifies full table | P1 | `L-BASES-01` | GET table; POST/PATCH record (mutation) latency |
| **A13-003** | Dashboard project stats `.All` issues on cache miss | P1 | `L-DASH-01` | project dashboard / stats endpoints cold+warm |
| **A13-004** | Gantt N+1 `listIssueLinks` per row (FE) | P1 | `L-GANTT-01` | API: list issues with due + links batch/unbatched pattern |
| **A13-005** | Notification `Dispatch` sync fan-out × recipients | P1 | `L-NOTIF-01` | trigger fan-out event (mention / call-started) → observe API/queue lag *(indirect)* |
| **A13-006** | `ListSprints` + `WithIssues()` unbounded | P1 | `L-SPRINT-01` | list sprints for project under load |

Связанные усилители (не отдельные обязательные v1, но tags): A8-001 missing project FK index × A13-001; A4-014 pagination gap × A13-002.

---

## Product hot paths (must-have v1)

Из миграции: **логин, лента, поиск, создание задач**.

| ID | Name | Flow (HTTP) | Auth | Weight (suggested) | Pass criteria (draft) |
|----|------|-------------|------|--------------------:|------------------------|
| `L-AUTH-01` | Login / session | `POST /api/dev/login` *или* magic-link request+verify в load-env; refresh token | seed users | 10% | p95 < 300ms login API; 0% 5xx |
| `L-FEED-01` | Personal feed / dashboard | `GET` personal dashboard widgets (tasks/calls/runs/…) | bearer | 25% | p95 < 800ms warm; track cold separately |
| `L-SEARCH-01` | Global search | `GET /api/search?q=` (issues/messages as implemented) | bearer | 20% | p95 < 500ms; no 5xx; empty+hot queries |
| `L-ISSUE-01` | List + create issue | `GET` project issues (cursor); `POST` create issue | bearer | 25% | create p95 < 600ms; list p95 < 400ms @ limit≤100 |
| `L-BOARD-01` | Kanban board | board list → `getBoardData` | bearer | 10% | baseline before/after A13-001 fix |
| `L-SPRINT-01` | Sprint list | `GET` sprints | bearer | 5% | baseline A13-006 |
| `L-BASES-01` | Bases table | GET table + light mutation | bearer | 3% | baseline A13-002 |
| `L-DASH-01` | Project dashboard | project stats | bearer | 2% | baseline A13-003 |

**v1.1 (after scaffold):** `L-GANTT-01`, `L-NOTIF-01` — сложнее (N+1 pattern / async fan-out); нужны стабильные fixtures и, для notif, контролируемый trigger.

---

## Scenario catalogue (detailed)

### L-AUTH-01 — Login

- **Users:** pool of N pre-seeded emails (`load-user-{i}@hub.local`) — **не** долбить один admin.  
- **Steps:** seed once → login → store access/refresh → optional refresh.  
- **Avoid:** UI Playwright login under Locust; magic-link email side-effects в shared mailbox.  
- **Note:** в DEV load-стенде допустим `/api/dev/login`; prod-like стенд — отдельный ADR/seed.

### L-FEED-01 — Feed / personal dashboard

- **Maps A13-012/013** (serial FE + no Redis on personal) as *observability*, primary assert = API latency.  
- **Steps:** authenticated GET personal dashboard endpoints used by `dashboard.lazy`.  
- **Tags:** `feed`, `dashboard`.

### L-SEARCH-01 — Search

- **Steps:** mix queries: short (`a`), typical (`login`), empty, unicode.  
- **Assert:** status in {200}; track Meilisearch degraded mode if health says so.  
- **Tags:** `search`.

### L-ISSUE-01 — Issues CRUD-lite

- **Steps:** list issues (limit 50) → create issue → (optional) get by id.  
- **Data:** one shared project id from fixtures; titles unique per user+timestamp.  
- **Cleanup:** soft-delete / trash job optional nightly — не блокирует load run.  
- **Tags:** `tracker`, `issues`.

### L-BOARD-01 — Board (A13-001)

- **Steps:** list boards → get board data for active board.  
- **Hypothesis:** latency/payload grow with issue count; record `content-length`.  
- **Tags:** `tracker`, `a13-001`.

### L-SPRINT-01 — Sprints (A13-006)

- **Steps:** `GET` sprints for project.  
- **Tags:** `sprint`, `a13-006`.

### L-BASES-01 — Bases (A13-002)

- **Steps:** GET table (all records path) → create/update one record → observe second GET cost.  
- **Tags:** `bases`, `a13-002`.

### L-DASH-01 — Project dashboard (A13-003)

- **Steps:** bust/ignore cache if test header exists; else first call = cold, second = warm.  
- **Tags:** `dashboard`, `a13-003`.

### L-GANTT-01 — Links N+1 (A13-004) — v1.1

- **Steps:** for M issues with due dates, call `listIssueLinks` M times **vs** one batch endpoint (when exists).  
- **Purpose:** regression detector for N+1, not soak.

### L-NOTIF-01 — Fan-out (A13-005) — v1.1

- **Steps:** post message with @mention to channel with K members **or** trigger call-started.  
- **Observe:** time-to-list notifications for recipients; DB-level metrics if exposed.  
- **Caution:** дорого; только staging; cap K.

---

## Locust shape (plan)

```
tests/load/
  locustfile.py          # registers shapes / default user mix
  users.py               # HubApiUser(HttpUser)
  scenarios/
    auth_login.py
    feed_dashboard.py
    search.py
    issues.py
    board.py
    sprint.py
    bases.py
    dashboard_project.py
  data/
    users.csv            # email list for load env
```

**User class mix (v1):** weighted Tasks из таблицы выше.  
**Host:** staging / `docker-compose.local` with scaled data (seed script: ≥1k issues for board realism — отдельный seed task, не в PR).

**SLOs (черновик, утверждает product/QA):**

| Signal | Smoke load (10 users, 2m) | Soak (50 users, 15m) |
|--------|---------------------------|----------------------|
| HTTP 5xx | 0% | <0.1% |
| p95 login | <300ms | <500ms |
| p95 search | <500ms | <800ms |
| p95 board | record baseline | regress <20% after A13-001 fix target |

---

## Out of scope (v1)

- Browser/WebSocket Locust extensions  
- LiveKit media load  
- Breaking production  
- Replacing unit/integration coverage  

---

## Mapping → GSD

Историческая привязка Phase 135 не является исполнимым маршрутом. Актуальная
нагрузочная работа — [Phase 140 `tf-load-resilience`](../../phases/140-tf-load-resilience/PLAN.md),
её launcher и runtime остаются `NOT_READY` / `NOT_RUN`.
Зависит от **128** scaffold (+ желательно **134** integration auth clients для переиспользования httpx helpers).
