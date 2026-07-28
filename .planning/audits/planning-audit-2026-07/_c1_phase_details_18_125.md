
> **Restored 2026-07-27 (planning audit C1):** Phase Details for **18–125** generated from each phase `PLAN.md` under `.planning/phases/`. Phases **01–13** live under `.planning/milestones/` (`v1.0-phases/` + `v2.0-phases/`) — their details above remain the historical ROADMAP copy. Do not treat this block as a substitute for PLAN/VERIFICATION artifacts.


### Phase 18: AI Privacy (v3.5)
**Goal**: M3-030 privacy consent · user settings UI · enforcement on gateway + bot · tests + docs
**Depends on**: Phase 17 AI Native
**Requirements**: M3-030
**Success Criteria** (what must be TRUE):
  1. W1-01 `allow_external_ai` on `ai_user_preference` + migration
  2. W1-02 Gateway enforcement (chat/stream/JQL/bot)
  3. W1-03 PATCH `/ai/preferences` + settings UI
  4. W1-04 Unit tests + API.md
  5. W2-01 Playwright `/settings/ai` consent toggle
  6. W2-02 API smoke: preferences + consent 403
**Plans**: 2 plan(s) · source `.planning/phases/18-v35-ai-privacy/PLAN.md`

Plans:
- [x] Wave 1 — M3-030 External AI consent
- [x] Wave 2 — E2E smoke + consent enforcement

### Phase 19: AI Polish
**Goal**: M3-024 composer Tab suggestion · dev seed AI provider · MILESTONE-3 reconciliation
**Depends on**: Phase 18 AI Privacy
**Requirements**: M3-024
**Success Criteria** (what must be TRUE):
  1. W1-01 POST `/ai/composer/suggest` + service
  2. W1-02 Tab handler + ghost hint in message-composer
  3. W1-03 Unit tests (Go + Vitest)
  4. W2-01 Dev seed Ollama provider (idempotent)
  5. W2-02 API.md + MILESTONE-3 checkboxes
**Plans**: 2 plan(s) · source `.planning/phases/19-v30-ai-polish/PLAN.md`

Plans:
- [x] Wave 1 — M3-024 Composer AI suggest (Tab)
- [x] Wave 2 — Dev seed + docs

### Phase 20: Milestone Gate
**Goal**: ROADMAP/STATE sync (15–16 SHIPPED) · E2E M3-024 composer · COVERAGE refresh · PG-04 closed
**Depends on**: Phases 15–19 complete (code)
**Requirements**: M3-024, PG-04
**Success Criteria** (what must be TRUE):
  1. W1-01 ROADMAP: Phase 15 Trust + Phase 16 Automate [x]
  2. W1-02 Phase 16 PLAN SHIPPED footer
  3. W1-03 STATE.md → v3.0 milestone complete
  4. W1-04 TASKS-CLOSE-GAPS PG-04 marked done
  5. W2-01 `e2e/helpers/ai.ts` → `composerSuggestRaw`
  6. W2-02 API smoke: prefix validation + suggest
**Plans**: 3 plan(s) · source `.planning/phases/20-milestone-gate/PLAN.md`

Plans:
- [x] Wave 1 — Milestone close-out
- [x] Wave 2 — E2E AI composer (M3-024)
- [x] Wave 3 — Docs + Vitest

### Phase 21: Architecture Completion
**Goal**: Repository pattern on remaining 7 packages (incremental) · cron boilerplate fully unified
**Depends on**: Phase 20 Milestone Gate
**Requirements**: ARCH-01
**Success Criteria** (what must be TRUE):
  1. W1-01 `backup.Repository` + `EntRepository`
  2. W1-02 `BackupService` → repo (no `*ent.Client`)
  3. W1-03 Tests + main.go wiring
  4. W2-01 Migrate pending-user purge to `cron.Schedule`
  5. W2-02 `platform/cron` unit tests
  6. W3-01 COVERAGE ARCH-01 progress + STATE/ROADMAP
**Plans**: 3 plan(s) · source `.planning/phases/21-arch-completion/PLAN.md`

Plans:
- [x] Wave 1 — backup Repository
- [x] Wave 2 — cron cleanup
- [x] Wave 3 — docs

### Phase 22: Architecture storage
**Goal**: `storage` module on Repository · orphan cleanup via repo · tests green
**Depends on**: Phase 21 backup Repository
**Requirements**: ARCH-01
**Success Criteria** (what must be TRUE):
  1. W1-01 `storage.Repository` + `EntRepository`
  2. W1-02 `StorageService` + `OrphanCleanupService` → repo
  3. W1-03 Tests + main.go wiring
**Plans**: 1 plan(s) · source `.planning/phases/22-arch-storage/PLAN.md`

Plans:
- [x] Wave 1 — storage Repository

### Phase 23: Architecture apigateway
**Goal**: `apigateway` APIKey + Webhook on Repository · tests green
**Depends on**: Phase 22 storage Repository
**Requirements**: ARCH-01
**Success Criteria** (what must be TRUE):
  1. W1-01 `APIKeyRepository` + `WebhookRepository` + Ent
  2. W1-02 Services + dispatcher → repo
  3. W1-03 Tests + main.go wiring
**Plans**: 1 plan(s) · source `.planning/phases/23-arch-apigateway/PLAN.md`

Plans:
- [x] Wave 1 — apigateway Repository

### Phase 24: Architecture imex
**Goal**: `imex` on Repository · Jira import uses atomic `AllocateIssueKey` · tests green
**Depends on**: Phase 23 apigateway Repository
**Requirements**: ARCH-01
**Success Criteria** (what must be TRUE):
  1. W1-01 `imex.Repository` + `EntRepository`
  2. W1-02 Service + jira + pdf → repo
  3. W1-03 Tests + main.go wiring
**Plans**: 1 plan(s) · source `.planning/phases/24-arch-imex/PLAN.md`

Plans:
- [x] Wave 1 — imex Repository

### Phase 25: Architecture dashboard
**Goal**: `dashboard` on Repository · tests green · 15/17 ARCH-01
**Depends on**: Phase 24 imex Repository
**Requirements**: ARCH-01
**Success Criteria** (what must be TRUE):
  1. W1-01 `dashboard.Repository` + `EntRepository`
  2. W1-02 Service → repo
  3. W1-03 Tests + main.go wiring
**Plans**: 1 plan(s) · source `.planning/phases/25-arch-dashboard/PLAN.md`

Plans:
- [x] Wave 1 — dashboard Repository

### Phase 26: Architecture rbac
**Goal**: `rbac` service on Repository · tests green · 16/17 ARCH-01
**Depends on**: Phase 25 dashboard Repository
**Requirements**: ARCH-01
**Success Criteria** (what must be TRUE):
  1. W1-01 `rbac.Repository` + `EntRepository`
  2. W1-02 Service + custom_roles → repo
  3. W1-03 Tests + main.go wiring
**Plans**: 1 plan(s) · source `.planning/phases/26-arch-rbac/PLAN.md`

Plans:
- [x] Wave 1 — rbac Repository

### Phase 27: Architecture auth
**Goal**: `auth` on Repository · middleware без `*ent.Client` · **17/17 ARCH-01** · tests green
**Depends on**: Phase 26 rbac Repository
**Requirements**: ARCH-01
**Success Criteria** (what must be TRUE):
  1. W1-01 `auth.Repository` + `EntRepository` + tx helpers
  2. W1-02 Sessions/Login/Profile/Reg/Org → repo
  3. W1-03 middleware + tests + main.go
**Plans**: 1 plan(s) · source `.planning/phases/27-arch-auth/PLAN.md`

Plans:
- [x] Wave 1 — auth Repository

### Phase 28: v2.6 Architecture Gate
**Goal**: ARCH-01 verified · no `*ent.Client` in rbac middleware · full `go test ./...` green · milestone v2.6 Architecture closed
**Depends on**: Phase 27 auth Repository (ARCH-01 17/17)
**Requirements**: ARCH-01, OBS-01
**Success Criteria** (what must be TRUE):
  1. W1-01 `rbac.RequireAdmin` → Repository
  2. W1-02 Full backend regression
  3. W1-03 STATE/ROADMAP/COVERAGE milestone close
**Plans**: 1 plan(s) · source `.planning/phases/28-v26-arch-gate/PLAN.md`

Plans:
- [x] Wave 1 — ARCH cleanup + gate

### Phase 29: OBS-01
**Goal**: `Timed()` records in-memory metrics · `GET /api/admin/observability/ops` · tests green
**Depends on**: Phase 28 v2.6 gate
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. W1-01 `observability` metrics registry
  2. W1-02 Wire Timed + admin endpoint
  3. W1-03 Tests
**Plans**: 1 plan(s) · source `.planning/phases/29-obs-01/PLAN.md`

Plans:
- [x] Wave 1 — Operation metrics

### Phase 30: OBS-01 FE
**Goal**: Admin UI for in-memory operation metrics from Phase 29.
**Depends on**: Phase 29
**Requirements**: OBS-01
**Success Criteria** (what must be TRUE):
  1. Admin sees operation metrics table with count/errors/avg/max/last seen
  2. Refresh reloads from `GET /api/admin/observability/ops`
  3. Vitest green; E2E smoke when backend+DEV_MODE available
**Plans**: 2 plan(s) · source `.planning/phases/30-obs-01-fe/PLAN.md`

Plans:
- [x] Wave 1 — API + component
- [x] Wave 2 — Tests + backend instrumentation

### Phase 31: OBS-01 Prometheus
**Goal**: Expose Timed() and HTTP metrics for Prometheus scrape (M2-053 partial).
**Depends on**: Phase 30
**Requirements**: OBS-01
**Success Criteria** (what must be TRUE):
  1. Scrape returns valid Prometheus text with hub_* metrics
  2. Timed() operations appear in both admin UI and `/metrics`
  3. Full `go test ./...` green
**Plans**: 8 plan(s) · source `.planning/phases/31-obs-prometheus/PLAN.md`

Plans:
- [x] `hub_operation_calls_total{operation,result}` counter
- [x] `hub_operation_duration_milliseconds{operation}` histogram
- [x] `hub_http_requests_total{method,route,status}` counter
- [x] `hub_http_request_duration_milliseconds{method,route}` histogram
- [x] `GET /metrics` — promhttp scrape endpoint (unauthenticated, infra-scoped)
- [x] `HTTPMetricsMiddleware` in `main.go`
- [x] E2E `metrics.spec.ts`
- [x] Unit tests `prometheus_test.go`

### Phase 32: OBS-01 OTel
**Goal**: OTLP trace export when `OTEL_EXPORTER_OTLP_ENDPOINT` is set (FB-21).
**Depends on**: Phase 31
**Requirements**: OBS-01, FB-21
**Success Criteria** (what must be TRUE):
  1. Without OTEL endpoint: zero overhead noop (TracingEnabled=false)
  2. With endpoint: HTTP spans + Timed() operation spans exported via OTLP HTTP
  3. OBS-01 **COMPLETE** (metrics + UI + Prometheus + traces)
**Plans**: 7 plan(s) · source `.planning/phases/32-obs-otel/PLAN.md`

Plans:
- [x] `observability/otel.go` — `InitTracing`, `HTTPTraceMiddleware`
- [x] `Timed()` creates child spans when tracing enabled
- [x] W3C trace context propagation on HTTP requests
- [x] `hub.request_id` span attribute from RequestID middleware
- [x] Wired in `main.go` (init + middleware after RequestID)
- [x] Unit tests `otel_test.go` (tracetest recorder)
- [x] `.env.local.example` OTEL vars

### Phase 33: Monitoring stack
**Goal**: Grafana dashboard provisioning + Prometheus scrape + `/metrics` hardening.
**Depends on**: Phase 32
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 2 plan(s) · source `.planning/phases/33-monitoring-stack/PLAN.md`

Plans:
- [x] Wave 1 — Deploy artifacts
- [x] Wave 2 — Security

### Phase 34: Loki + alerts
**Goal**: Log aggregation via Loki/Promtail + Prometheus alert rules for Hub SLOs.
**Depends on**: Phase 33
**Requirements**: PG-02
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 2 plan(s) · source `.planning/phases/34-loki-alerts/PLAN.md`

Plans:
- [x] Wave 1 — Loki stack
- [x] Wave 2 — Alerts + log correlation

### Phase 35: Alertmanager
**Goal**: Route Prometheus alerts through Alertmanager; Grafana AM datasource.
**Depends on**: Phase 34
**Requirements**: PG-02
**Success Criteria** (what must be TRUE):
  1. Prometheus → Alertmanager connectivity configured
  2. Firing alerts visible in AM UI
  3. M2-058 **COMPLETE** (rules + routing; notification channel = ops webhook)
**Plans**: 6 plan(s) · source `.planning/phases/35-alertmanager/PLAN.md`

Plans:
- [x] `deployments/monitoring/alertmanager/alertmanager.yml` — routes, inhibit rules, null receivers (dev)
- [x] `alertmanager.prod.example.yml` — webhook template for prod ops
- [x] Prometheus `alerting.alertmanagers` → `alertmanager:9093`
- [x] Alertmanager service in `docker-compose.monitoring.yml` (`:9093`)
- [x] Grafana Alertmanager datasource (`hub-alertmanager`)
- [x] New alert: `HubBackendMetricsDown` (critical, `up==0`)

### Phase 36: Alert webhook
**Goal**: Close OBS-01 production loop — Alertmanager fires/resolves alerts appear in admin Audit Log.
**Depends on**: Phase 35
**Requirements**: OBS-01
**Success Criteria** (what must be TRUE):
  1. `POST /api/webhooks/alertmanager` — Bearer `ALERT_WEBHOOK_SECRET`
  2. `RecordAlertmanagerWebhook` → audit entries `alert.firing` / `alert.resolved`, entity `alert`
  3. Skip HTTP audit middleware for webhook path
  4. Config: `ALERT_WEBHOOK_SECRET` in `ServerConfig`
  5. Wire in `main.go` via `RegisterAlertWebhook`
  6. Frontend: audit filters + badge for alert actions
**Plans**: 8 plan(s) · source `.planning/phases/36-alert-webhook/PLAN.md`

Plans:
- [x] `POST /api/webhooks/alertmanager` — Bearer `ALERT_WEBHOOK_SECRET`
- [x] `RecordAlertmanagerWebhook` → audit entries `alert.firing` / `alert.resolved`, entity `alert`
- [x] Skip HTTP audit middleware for webhook path
- [x] Config: `ALERT_WEBHOOK_SECRET` in `ServerConfig`
- [x] Wire in `main.go` via `RegisterAlertWebhook`
- [x] Frontend: audit filters + badge for alert actions
- [x] Deploy: `.env.example`, `docker-compose.prod.yml`, `alertmanager.prod.example.yml`
- [x] Tests: `handler_alerts_test.go` (auth, record, disabled when secret empty)

### Phase 37: Biome gate
**Goal**: `npm run lint` → 0 errors, 0 warnings (PG-02 close).
**Depends on**: Phase 36
**Requirements**: PG-02
**Success Criteria** (what must be TRUE):
  1. Biome `--write` safe auto-fixes (21 files formatted/organized)
  2. Fix 10 lint errors (noArrayIndexKey, useExhaustiveDependencies, noAssignInExpressions)
  3. Fix 72 warnings across a11y + suspicious + style rules
  4. Add `@/lib/keyboard` helper (`activateOnEnterOrSpace`)
  5. `buildQueryString` typed as `object` (no `any`, tsc clean)
**Plans**: 5 plan(s) · source `.planning/phases/37-biome-gate/PLAN.md`

Plans:
- [x] Biome `--write` safe auto-fixes (21 files formatted/organized)
- [x] Fix 10 lint errors (noArrayIndexKey, useExhaustiveDependencies, noAssignInExpressions)
- [x] Fix 72 warnings across a11y + suspicious + style rules
- [x] Add `@/lib/keyboard` helper (`activateOnEnterOrSpace`)
- [x] `buildQueryString` typed as `object` (no `any`, tsc clean)

### Phase 38: v2.6 Milestone Gate
**Goal**: Close v2.6 milestone — docs sync, E2E/Vitest for Phase 36 alerts, qa-gate green, CI Biome strict.
**Depends on**: Phase 37
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. `docs/wiki/API.md` — `/metrics`, `/admin/observability/ops`, `/webhooks/alertmanager`, audit alert filters
  2. Vitest: alert audit entry rendering (`audit-log.test.tsx`)
  3. E2E: admin audit alert filter options (`admin-audit.spec.ts`)
  4. CI: Biome job renamed to gate (no continue-on-error)
  5. Fix race: `ai/service_test.go` mockStreamPublisher mutex
  6. qa-gate: build + race + tsc + lint + vitest
**Plans**: 6 plan(s) · source `.planning/phases/38-v26-milestone-gate/PLAN.md`

Plans:
- [x] `docs/wiki/API.md` — `/metrics`, `/admin/observability/ops`, `/webhooks/alertmanager`, audit alert filters
- [x] Vitest: alert audit entry rendering (`audit-log.test.tsx`)
- [x] E2E: admin audit alert filter options (`admin-audit.spec.ts`)
- [x] CI: Biome job renamed to gate (no continue-on-error)
- [x] Fix race: `ai/service_test.go` mockStreamPublisher mutex
- [x] qa-gate: build + race + tsc + lint + vitest

### Phase 39: Intranet News
**Goal**: Hub One v4 kickoff — corporate news feed (MOD-INTRA first slice).
**Depends on**: Phase 38
**Requirements**: MOD-01, MOD-02, MOD-04
**Success Criteria** (what must be TRUE):
  1. Ent schema `NewsPost` (soft delete, pin, publish, version)
  2. `backend/internal/intranet` — Repository + Service + Handler
  3. Routes `/api/intranet/news*` wired in `main.go`
  4. FE: `/intranet` route, sidebar «Новости», admin compose
  5. Tests: handler (sqlite), Vitest smoke
  6. API.md Intranet section
**Plans**: 6 plan(s) · source `.planning/phases/39-intranet-news/PLAN.md`

Plans:
- [x] Ent schema `NewsPost` (soft delete, pin, publish, version)
- [x] `backend/internal/intranet` — Repository + Service + Handler
- [x] Routes `/api/intranet/news*` wired in `main.go`
- [x] FE: `/intranet` route, sidebar «Новости», admin compose
- [x] Tests: handler (sqlite), Vitest smoke
- [x] API.md Intranet section

### Phase 40: Org directory
**Goal**: Справочник сотрудников Hub One v4: группировка по отделам, карточки профиля, интеграция с глобальным поиском.
**Depends on**: Phase 39
**Requirements**: MOD-02, MOD-04
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 10 plan(s) · source `.planning/phases/40-org-directory/PLAN.md`

Plans:
- [x] Backend
- [x] `GET /api/intranet/directory` — active users grouped by `department`
- [x] `GET /api/intranet/directory/users/:userId` — profile card (404 for pending/deactivated)
- [x] Extend `internal/intranet/` repository/service/handler
- [x] Search user URLs → `/intranet/directory?user={id}`
- [x] Frontend
- [x] `/intranet/directory` — `OrgDirectory` component
- [x] Sidebar «Справочник»
- [x] Ctrl+K action «Справочник сотрудников»
- [x] Profile dialog via `?user=` search param

### Phase 41: Unified calendar
**Goal**: First-class `/calendar` route aggregating scheduled calls, sprint ranges, and issue due dates.
**Depends on**: Phase 40
**Requirements**: MOD-04, MOD-05, MOD-03
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 10 plan(s) · source `.planning/phases/41-unified-calendar/PLAN.md`

Plans:
- [x] Backend (`internal/calendar/`)
- [x] `GET /api/calendar/events?from=&to=` — user-scoped aggregation
- [x] `GET /api/calendar/export.ics` — RFC 5545 export
- [x] RRULE expansion for recurring calls via `rrule-go`
- [x] Frontend
- [x] `/calendar` — `UnifiedCalendar` (reuses calls month/week/day/agenda views)
- [x] Filter chips: calls / sprints / deadlines
- [x] iCal download button
- [x] Sidebar + Ctrl+K shortcut
- [x] Tests

### Phase 42: CRM-lite
**Goal**: CRM-lite для IT-команды: контакты, pipeline сделок, привязка к проектам Hub.
**Depends on**: Phase 41
**Requirements**: MOD-05, MOD-06, MOD-03
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 10 plan(s) · source `.planning/phases/42-crm-lite/PLAN.md`

Plans:
- [x] Ent
- [x] `CrmContact` — name, email, phone, company, optional `linked_user`
- [x] `CrmDeal` — title, stage enum, amount_cents, contact, project, owner
- [x] Backend (`internal/crm/`)
- [x] Contacts CRUD + soft delete
- [x] Pipeline GET + deals CRUD
- [x] Project link validates membership
- [x] Frontend
- [x] `/crm` — contacts sidebar + kanban pipeline (drag-drop stage move)
- [x] Sidebar + Ctrl+K

### Phase 43: HR-lite
**Goal**: HR-lite: заявки на отсутствие, очередь согласования admin/founder, одобренные отпуска в unified calendar, уведомления при submit/decision.
**Depends on**: Phase 42
**Requirements**: MOD-06, MOD-03
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 10 plan(s) · source `.planning/phases/43-hr-lite/PLAN.md`

Plans:
- [x] Ent
- [x] `LeaveRequest` — leave_type, status, start/end dates, reason, review_note, version, reviewed_at
- [x] User edges: `leave_requests`, `leave_reviews`
- [x] Backend (`internal/hr/`)
- [x] CRUD flow: create, list mine, list pending (admin), approve/reject/cancel
- [x] Overlap validation for pending/approved ranges
- [x] `hrNotifierAdapter` → NotificationService on submit/decision
- [x] Calendar integration
- [x] `EventTypeLeave` in `internal/calendar`
- [x] `ListApprovedLeaveInRange` in calendar repository

### Phase 44: Kudos
**Goal**: «Спасибо» коллеге: in-app уведомление получателю + публичная лента на intranet.
**Depends on**: Phase 43
**Requirements**: MOD-03, MOD-08, MOD-07
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 10 plan(s) · source `.planning/phases/44-kudos/PLAN.md`

Plans:
- [x] Ent
- [x] `Kudo` — message (max 500), is_public, sender/recipient edges
- [x] Backend (`internal/intranet/` extended)
- [x] `POST /intranet/kudos` — send (no self-kudos)
- [x] `GET /intranet/kudos/feed` — public feed
- [x] `GET /intranet/kudos/received` — my received
- [x] `intranetKudosNotifier` → NotificationService
- [x] Frontend
- [x] Intranet hub tabs: Новости | Спасибо
- [x] `kudos-feed.tsx` — send form + feed/received tabs

### Phase 45: Marketplace
**Goal**: Внутренняя барахолка: объявления с модерацией admin/founder, публичная лента, уведомления mod queue и автору.
**Depends on**: Phase 44
**Requirements**: MOD-08, V4-06, MOD-07
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 10 plan(s) · source `.planning/phases/45-marketplace/PLAN.md`

Plans:
- [x] Ent
- [x] `MarketplaceListing` — title, description, category, status, price_cents, review fields, version
- [x] Backend (`internal/marketplace/`)
- [x] CRUD + approve/reject/sold/withdraw
- [x] Re-moderation on edit of approved listing
- [x] `marketplaceNotifierAdapter` → admins on submit, author on decision
- [x] `marketplaceChannelBroadcaster` → `#marketplace` channel via SendBotMessage (optional)
- [x] Frontend
- [x] `/marketplace` — browse/mine tabs, create form, admin mod queue
- [x] Sidebar + Ctrl+K

### Phase 46: Status page
**Goal**: Публичная страница `/status` без авторизации: статус Postgres, Redis, Mongo, NATS, API.
**Depends on**: Phase 45
**Requirements**: V4-06, MOD-07
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 5 plan(s) · source `.planning/phases/46-status-page/PLAN.md`

Plans:
- [x] Backend (`internal/status/`)
- [x] `GET /api/status` — aggregated probes, 503 if degraded
- [x] Frontend
- [x] `/status` — public route, auto-refresh 30s
- [x] Link to login

### Phase 47: Random Coffee
**Goal**: Opt-in участие → admin запускает matching → пары коллег → in-app уведомления.
**Depends on**: Phase 46
**Requirements**: V4-05, MOD-07, V4-04
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 8 plan(s) · source `.planning/phases/47-random-coffee/PLAN.md`

Plans:
- [x] Ent
- [x] `CoffeeEnrollment`, `CoffeeRound`, `CoffeePair`
- [x] Backend (`internal/intranet/` coffee_*)
- [x] Enrollment, match query, admin run matching
- [x] Shuffle + anti-repeat vs previous round
- [x] `intranetNotifier.NotifyCoffeeMatch`
- [x] Frontend
- [x] Intranet tab «Coffee» — opt-in, partner card, admin run

### Phase 48: 1C Connector
**Goal**: When admin approves a leave request, Hub POSTs a JSON webhook to a configurable 1C endpoint. Admin configures URL, bearer token, and employee ID mapping field.
**Depends on**: Phase 47
**Requirements**: MOD-07, V4-04
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/48-onec-connector/PLAN.md`

Plans:
- [x] `.planning/phases/48-onec-connector/PLAN.md` (single plan)

### Phase 49: Lark Base
**Goal**: Org-wide no-code tables: workspaces (bases), typed columns, grid records. MVP slice — no kanban views, automations, or formulas.
**Depends on**: Phase 48
**Requirements**: V4-04
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/49-lark-base/PLAN.md`

Plans:
- [x] `.planning/phases/49-lark-base/PLAN.md` (single plan)

### Phase 50: v4 Milestone Gate
**Goal**: Close Hub One v4 block (Phases 40–49) — qa-gate green, Vitest smoke for new modules, docs sync.
**Depends on**: Phase 49
**Requirements**: V4-03
**Success Criteria** (what must be TRUE):
  1. Vitest: `bases-page.test.tsx`, `onec-settings.test.tsx`
  2. Biome lint 0 errors (v4 FE files formatted)
  3. qa-gate: `go build -tags release`, `go vet`, `go test -race ./...`
  4. Frontend: `tsc`, `lint`, `vitest run` (79 tests)
  5. API docs already synced in Phases 40–49
  6. `VERIFICATION.md` gate report
**Plans**: 6 plan(s) · source `.planning/phases/50-v4-milestone-gate/PLAN.md`

Plans:
- [x] Vitest: `bases-page.test.tsx`, `onec-settings.test.tsx`
- [x] Biome lint 0 errors (v4 FE files formatted)
- [x] qa-gate: `go build -tags release`, `go vet`, `go test -race ./...`
- [x] Frontend: `tsc`, `lint`, `vitest run` (79 tests)
- [x] API docs already synced in Phases 40–49
- [x] `VERIFICATION.md` gate report

### Phase 51: Accounting MVP
**Goal**: Double-entry ledger for admins: chart of accounts, journal entries (draft → posted), trial balance. Plus Playwright smoke for v4 routes.
**Depends on**: Phase 50
**Requirements**: V4-03
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/51-accounting-mvp/PLAN.md`

Plans:
- [x] `.planning/phases/51-accounting-mvp/PLAN.md` (single plan)

### Phase 52: Accounting v4.5
**Goal**: Extend Hub accounting with customer invoices (draft → issued → paid with auto journal), P&L report, and CRM won-deal hook.
**Depends on**: Phase 51
**Requirements**: V4-03
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/52-accounting-v45/PLAN.md`

Plans:
- [x] `.planning/phases/52-accounting-v45/PLAN.md` (single plan)

### Phase 53: v4 Module Gate
**Goal**: Gate phases 40–52: integration tests for CRM won→invoice, HR reject flow, E2E accounting v4.5 UI, qa-gate green.
**Depends on**: Phase 52
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/53-v4-module-gate/PLAN.md`

Plans:
- [x] `.planning/phases/53-v4-module-gate/PLAN.md` (single plan)

### Phase 54: v4 Coverage Uplift
**Goal**: Raise low-coverage v4 packages `bases/` and `marketplace/` to ≥60% via service + handler tests.
**Depends on**: Phase 53
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/54-v4-coverage-uplift/PLAN.md`

Plans:
- [x] `.planning/phases/54-v4-coverage-uplift/PLAN.md` (single plan)

### Phase 55: onec + intranet
**Goal**: Raise `onec/` (38%) and confirm `intranet/` (54%) ≥50% statement coverage.
**Depends on**: Phase 54
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/55-onec-intranet-coverage/PLAN.md`

Plans:
- [x] `.planning/phases/55-onec-intranet-coverage/PLAN.md` (single plan)

### Phase 56: v4 Final Gate
**Goal**: Close v4 super-app milestone (phases 40–55): CRM last gap, regression gate green.
**Depends on**: Phase 55
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/56-v4-final-gate/PLAN.md`

Plans:
- [x] `.planning/phases/56-v4-final-gate/PLAN.md` (single plan)

### Phase 57: Post-v4 CI Handoff
**Goal**: Close the loop after Phase 56: v4 Playwright in `e2e-full`, E2E fix for `/accounting` admin gate, AGENTS/STATE handoff.
**Depends on**: Phase 56
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/57-post-v4-ci-handoff/PLAN.md`

Plans:
- [x] `.planning/phases/57-post-v4-ci-handoff/PLAN.md` (single plan)

### Phase 58: v4 Milestone Archive
**Goal**: Formal close of **Hub One v4 Super-app** milestone: archive SUMMARY, full qa-gate with `-race`, coverage refresh, docs sync.
**Depends on**: Phase 57 Post-v4 CI Handoff
**Requirements**: PG-03
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 3 plan(s) · source `.planning/phases/58-v4-milestone-archive/PLAN.md`

Plans:
- [x] `.planning/milestones/v4-super-app/SUMMARY.md` — milestone archive
- [x] `.planning/phases/58-v4-milestone-archive/VERIFICATION.md` — gate report
- [x] Updated: `STATE.md`, `ROADMAP.md`, `AGENTS.md`, `COVERAGE.md`

### Phase 59: Coverage + UX Close-out
**Goal**: Close PLAN-CLOSE-GAPS residual: **tracker** and **storage** ≥50%, **FB-11** issue-form Vitest.
**Depends on**: Phase 58 v4 Milestone Archive
**Requirements**: UX-1, FB-11, PG-03
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/59-coverage-ux-closeout/PLAN.md`

Plans:
- [x] `.planning/phases/59-coverage-ux-closeout/PLAN.md` (single plan)

### Phase 60: UX-1 Gate
**Goal**: Close UX-1 backlog items: **FB-01** shortcuts E2E, **UX-P1-06** axe calls/TMS, **FB-12** onboarding project step Zod validation.
**Depends on**: Phase 59 Coverage + UX Close-out
**Requirements**: UX-1, FB-01, UX-P1-06, FB-12, PG-03
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/60-ux1-gate/PLAN.md`

Plans:
- [x] `.planning/phases/60-ux1-gate/PLAN.md` (single plan)

### Phase 61: Coverage + Context Tooltips
**Goal**: 1. Raise `project` and `search` packages to ≥50% coverage 2. UX-P1-04: contextual tooltips (JQL, WIP, Allure) 3. Fix corrupted UTF-8 in `jql-search-page.tsx`
**Depends on**: Phase 60
**Requirements**: UX-P1-04
**Success Criteria** (what must be TRUE):
  1. project ≥50% coverage
  2. search ≥50% coverage
  3. Vitest green (context-help-tip + existing suite)
  4. UX-P1-04 tooltips visible on target surfaces
**Plans**: 4 plan(s) · source `.planning/phases/61-coverage-tooltips/PLAN.md`

Plans:
- [x] project ≥50% coverage
- [x] search ≥50% coverage
- [x] Vitest green (context-help-tip + existing suite)
- [x] UX-P1-04 tooltips visible on target surfaces

### Phase 62: Onboarding Profile RHF
**Goal**: Close **UX-P1-07 / FB-12** remainder: profile step (step 1) with RHF + Zod validation, wired language select, E2E gate.
**Depends on**: Phase 61
**Requirements**: UX-P1-07, FB-12
**Success Criteria** (what must be TRUE):
  1. Profile step uses RHF+Zod (matches issue-form pattern)
  2. Next disabled until name ≥ 2 chars
  3. Language select persisted to completeOnboarding
  4. Vitest + E2E spec added
**Plans**: 4 plan(s) · source `.planning/phases/62-onboarding-profile-rhf/PLAN.md`

Plans:
- [x] Profile step uses RHF+Zod (matches issue-form pattern)
- [x] Next disabled until name ≥ 2 chars
- [x] Language select persisted to completeOnboarding
- [x] Vitest + E2E spec added

### Phase 63: Production Release Gate
**Goal**: Full `/qa-gate` on accumulated Phases 59–62; production handoff with PG-03 runbook reference; schema route hygiene.
**Depends on**: Phase 62
**Requirements**: PG-03, COV-01
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/63-production-gate/PLAN.md`

Plans:
- [x] `.planning/phases/63-production-gate/PLAN.md` (single plan)

### Phase 64: Auth Coverage Uplift
**Goal**: Close COV-01 security-critical gap: `auth` package **≥80%** statement coverage.
**Depends on**: Phase 63
**Requirements**: COV-01
**Success Criteria** (what must be TRUE):
  1. auth ≥80%
  2. all auth tests pass
  3. security-critical 6/6 ≥80% restored
**Plans**: 3 plan(s) · source `.planning/phases/64-auth-coverage/PLAN.md`

Plans:
- [x] auth ≥80%
- [x] all auth tests pass
- [x] security-critical 6/6 ≥80% restored

### Phase 65: UX-P1-05 Rate Limit Timer + Messenger Coverage
**Goal**: 1. **UX-P1-05 / FB-03:** Live countdown toast on 429 + actionable `request_id` in error toasts 2. **Coverage:** `messenger/` **48% → ≥50%** via `ent_repository` integration tests
**Depends on**: Phase 64
**Requirements**: UX-P1-05, FB-03
**Success Criteria** (what must be TRUE):
  1. 429 toast counts down each second until dismiss
  2. Non-429 errors show `request_id` when present
  3. messenger ≥50%
  4. Vitest green, `go test ./internal/messenger/...` green
**Plans**: 4 plan(s) · source `.planning/phases/65-ux-rate-limit-messenger-cov/PLAN.md`

Plans:
- [x] 429 toast counts down each second until dismiss
- [x] Non-429 errors show `request_id` when present
- [x] messenger ≥50%
- [x] Vitest green, `go test ./internal/messenger/...` green

### Phase 66: Notification Coverage Uplift (42.5% → 50%)
**Goal**: Close COV gap: `notification/` package **≥50%** + telegram handler route fix.
**Depends on**: Phase 65
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. notification ≥50%
  2. all notification tests pass
  3. telegram DELETE /notifications/telegram not shadowed by /:id
**Plans**: 3 plan(s) · source `.planning/phases/66-notification-coverage/PLAN.md`

Plans:
- [x] notification ≥50%
- [x] all notification tests pass
- [x] telegram DELETE /notifications/telegram not shadowed by /:id

### Phase 67: Production Release Gate (Phases 64–66)
**Goal**: Re-run full `/qa-gate` after Phases 64–66; publish `PRODUCTION-HANDOFF.md`; refresh coverage baseline.
**Depends on**: Phase 66
**Requirements**: COV-01, UX-P1-05, NOTIF, PG-03
**Success Criteria** (what must be TRUE):
  1. COMPLETE
**Plans**: 1 plan(s) · source `.planning/phases/67-production-gate/PLAN.md`

Plans:
- [x] `.planning/phases/67-production-gate/PLAN.md` (single plan)

### Phase 68: Messenger Fan-out & Coverage (52.3% → 60%)
**Goal**: PLAN-CLOSE-GAPS **2.5**: messenger fan-out integration + coverage **≥60%**.
**Depends on**: Phase 67
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. messenger ≥60%
  2. fan-out unread test (sender excluded)
  3. all messenger tests pass
**Plans**: 3 plan(s) · source `.planning/phases/68-messenger-fanout-cov/PLAN.md`

Plans:
- [x] messenger ≥60%
- [x] fan-out unread test (sender excluded)
- [x] all messenger tests pass

### Phase 69: Apigateway Coverage + COV Baseline Refresh
**Goal**: 1. **SEC-01 / API-01:** `apigateway/` **54% → ≥60%** 2. Refresh overall COV baseline post Phase 68
**Depends on**: Phase 68
**Requirements**: SEC-01
**Success Criteria** (what must be TRUE):
  1. apigateway ≥60%
  2. all apigateway tests pass
  3. overall COV refreshed
**Plans**: 3 plan(s) · source `.planning/phases/69-apigateway-cov-baseline/PLAN.md`

Plans:
- [x] apigateway ≥60%
- [x] all apigateway tests pass
- [x] overall COV refreshed

### Phase 70: GitLab/GitHub Coverage (SEC-04)
**Goal**: 1. **SEC-04 / GH-01:** `gitlab/` and `github/` **48% / 45% → ≥50%** 2. Cover `EntRepository` (0%) and `RegisterRoutes` (0%) for both modules 3. Refresh overall COV baseline
**Depends on**: Phase 69
**Requirements**: SEC-04
**Success Criteria** (what must be TRUE):
  1. gitlab ≥50%
  2. github ≥50%
  3. RegisterRoutes covered both modules
  4. EntRepository covered both modules
  5. all gitlab/github tests pass
**Plans**: 5 plan(s) · source `.planning/phases/70-gitlab-github-cov/PLAN.md`

Plans:
- [x] gitlab ≥50%
- [x] github ≥50%
- [x] RegisterRoutes covered both modules
- [x] EntRepository covered both modules
- [x] all gitlab/github tests pass

### Phase 71: Mail + Dashboard Coverage (MAIL-01 / BE-H3)
**Goal**: 1. **MAIL-01 / Phase 15.2:** `mail/` **44.6% → ≥50%** — hub-mail handlers + admin API 2. **BE-H3:** `dashboard/` project repo + `Project` handler **61.7% → ≥65%** 3. Refresh overall COV baseline
**Depends on**: Phase 70
**Requirements**: MAIL-01
**Success Criteria** (what must be TRUE):
  1. mail ≥50%
  2. dashboard ≥65%
  3. hub-mail RegisterRoutes covered
  4. admin/mail RegisterRoutes covered
  5. all mail/dashboard tests pass
**Plans**: 5 plan(s) · source `.planning/phases/71-mail-dashboard-cov/PLAN.md`

Plans:
- [x] mail ≥50%
- [x] dashboard ≥65%
- [x] hub-mail RegisterRoutes covered
- [x] admin/mail RegisterRoutes covered
- [x] all mail/dashboard tests pass

### Phase 72: Production QA Gate (Post Phase 71)
**Goal**: Full `/qa-gate` re-run after Phases 70–71 (gitlab/github/mail/dashboard coverage).
**Depends on**: Phase 71
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. All backend race tests green
  2. Frontend tsc + Biome + Vitest green
  3. No regressions from Phase 70–71
**Plans**: 3 plan(s) · source `.planning/phases/72-production-qa-gate/PLAN.md`

Plans:
- [x] All backend race tests green
- [x] Frontend tsc + Biome + Vitest green
- [x] No regressions from Phase 70–71

### Phase 73: Platform HTTP/Outbox Coverage (SEC-04 / NF-07)
**Goal**: 1. `platform/http` **20.4% → ≥50%** — SafeDialer + OpenGraph (SSRF / NF-07) 2. `platform/outbox` **28.6% → ≥50%** — Publisher + Worker happy path
**Depends on**: Phase 72
**Requirements**: SEC-04
**Success Criteria** (what must be TRUE):
  1. platform/http ≥50%
  2. platform/outbox ≥50%
  3. SafeDialer + ValidateOutboundURL covered
  4. Publisher.Publish/PublishInTx covered
**Plans**: 4 plan(s) · source `.planning/phases/73-platform-http-outbox-cov/PLAN.md`

Plans:
- [x] platform/http ≥50%
- [x] platform/outbox ≥50%
- [x] SafeDialer + ValidateOutboundURL covered
- [x] Publisher.Publish/PublishInTx covered

### Phase 74: AI Module Coverage (M3 / COV-01)
**Goal**: `internal/ai/` **24.8% → ≥50%** — handlers, repository, settings, privacy, adapters (unit paths without live LLM).
**Depends on**: Phase 73
**Requirements**: COV-01
**Success Criteria** (what must be TRUE):
  1. ai/ ≥50%
  2. `go test -race ./internal/ai/...` PASS
  3. Handler routes + admin/user BYOK covered
  4. EntRepository preferences/usage covered
**Plans**: 4 plan(s) · source `.planning/phases/74-ai-coverage/PLAN.md`

Plans:
- [x] ai/ ≥50%
- [x] `go test -race ./internal/ai/...` PASS
- [x] Handler routes + admin/user BYOK covered
- [x] EntRepository preferences/usage covered

### Phase 75: Post-AI QA Gate Refresh
**Goal**: Re-run full `/qa-gate` after Phase 74 AI coverage additions; refresh overall COV baseline.
**Depends on**: Phase 74
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. Full qa-gate PASS
  2. No regressions from Phase 74 ai tests
  3. COV baseline updated
**Plans**: 3 plan(s) · source `.planning/phases/75-post-ai-qa-gate/PLAN.md`

Plans:
- [x] Full qa-gate PASS
- [x] No regressions from Phase 74 ai tests
- [x] COV baseline updated

### Phase 76: Sprint Ent + Handler Routes Coverage
**Goal**: `internal/sprint/` **58.0% → ≥60%** — EntRepository + RegisterRoutes (RBAC)
**Depends on**: Phase 75
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. sprint ≥60%
  2. EntRepository fully covered
  3. RegisterRoutes + UpdateSprint covered
  4. race PASS
**Plans**: 4 plan(s) · source `.planning/phases/76-sprint-coverage/PLAN.md`

Plans:
- [x] sprint ≥60%
- [x] EntRepository fully covered
- [x] RegisterRoutes + UpdateSprint covered
- [x] race PASS

### Phase 77: Storage Handler Routes (CR-04 / Presign)
**Goal**: `internal/storage/` **55.7% → ≥60%** — NewHandler, RegisterRoutes, RBAC authorizeProject, NewStorageService
**Depends on**: Phase 76
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. storage ≥60%
  2. RegisterRoutes covered
  3. CR-04 project-scoped auth paths covered
  4. race PASS
**Plans**: 4 plan(s) · source `.planning/phases/77-storage-routes-cov/PLAN.md`

Plans:
- [x] storage ≥60%
- [x] RegisterRoutes covered
- [x] CR-04 project-scoped auth paths covered
- [x] race PASS

### Phase 78: Tracker + TMS Coverage Wave
**Goal**: - `internal/tracker/` **52.1% → ≥60%** - `internal/tms/` **55.5% → ≥60%**
**Depends on**: Phase 77
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. tracker ≥60%
  2. tms ≥60%
  3. RegisterRoutes covered both modules
  4. race PASS
**Plans**: 4 plan(s) · source `.planning/phases/78-tracker-tms-coverage/PLAN.md`

Plans:
- [x] tracker ≥60%
- [x] tms ≥60%
- [x] RegisterRoutes covered both modules
- [x] race PASS

### Phase 79: Notification + Project Coverage Wave
**Goal**: - `internal/notification/` **55.7% → ≥60%** - `internal/project/` **59.4% → ≥60%**
**Depends on**: Phase 78
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. notification ≥60%
  2. project ≥60%
  3. race PASS both packages
  4. no production code changes (tests only)
**Plans**: 4 plan(s) · source `.planning/phases/79-notification-project-cov/PLAN.md`

Plans:
- [x] notification ≥60%
- [x] project ≥60%
- [x] race PASS both packages
- [x] no production code changes (tests only)

### Phase 80: Backup + HR + IMEX Coverage Wave
**Goal**: - `internal/backup/` **53.4% → ≥60%** - `internal/hr/` **53.1% → ≥60%** - `internal/imex/` **57.8% → ≥60%**
**Depends on**: Phase 79
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. backup ≥60%
  2. hr ≥60%
  3. imex ≥60%
  4. race PASS all three packages
**Plans**: 4 plan(s) · source `.planning/phases/80-backup-hr-imex-cov/PLAN.md`

Plans:
- [x] backup ≥60%
- [x] hr ≥60%
- [x] imex ≥60%
- [x] race PASS all three packages

### Phase 81: AI Coverage + QA Gate
**Goal**: - `internal/ai/` **50.7% → ≥60%** - Full `/qa-gate` PASS on branch
**Depends on**: Phase 80
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. ai ≥60%
  2. go build -tags release ./...
  3. go vet ./...
  4. go test -race ./...
  5. frontend tsc + lint + vitest (99/99)
**Plans**: 5 plan(s) · source `.planning/phases/81-ai-coverage-qa-gate/PLAN.md`

Plans:
- [x] ai ≥60%
- [x] go build -tags release ./...
- [x] go vet ./...
- [x] go test -race ./...
- [x] frontend tsc + lint + vitest (99/99)

### Phase 82: AI Bot Hook + Doc Search RAG
**Goal**: - Close Phase 81 deferred: `@hub-ai` messenger bot hook + `SearchDocProvider` full path - `internal/ai/` **68.5% → ≥70%**
**Depends on**: Phase 81
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. OnMessageSent + streamReply covered
  2. SearchContext with results/empty/error
  3. race PASS ai + messenger
  4. release build PASS
**Plans**: 4 plan(s) · source `.planning/phases/82-ai-bot-docsearch/PLAN.md`

Plans:
- [x] OnMessageSent + streamReply covered
- [x] SearchContext with results/empty/error
- [x] race PASS ai + messenger
- [x] release build PASS

### Phase 83: AI Bot Commands + Search Query
**Goal**: - Complete `@hub-ai` slash commands (test-cases, review, explain, translate, sql) - `internal/search/` **60.5% → ≥60%** (query paths) - `internal/ai/` **75.0% → ≥76%**
**Depends on**: Phase 82
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. All `/ai` command branches covered (except live VCS fetch)
  2. Search Search() orchestration paths
  3. race PASS ai + search
**Plans**: 3 plan(s) · source `.planning/phases/83-ai-commands-search/PLAN.md`

Plans:
- [x] All `/ai` command branches covered (except live VCS fetch)
- [x] Search Search() orchestration paths
- [x] race PASS ai + search

### Phase 84: VCS Review + Bases + Intranet Coffee
**Goal**: - `UnifiedCodeReviewProvider.FetchReview` GitLab path via mock HTTP - `internal/bases/` **60.8% → ≥65%** - `internal/gitlab/` MR review coverage uplift
**Depends on**: Phase 83
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. GitLab FetchMRForReview httptest
  2. AI Unified FetchReview GitLab routing
  3. bases ≥65%
  4. race PASS touched packages
**Plans**: 4 plan(s) · source `.planning/phases/84-vcs-bases-intranet/PLAN.md`

Plans:
- [x] GitLab FetchMRForReview httptest
- [x] AI Unified FetchReview GitLab routing
- [x] bases ≥65%
- [x] race PASS touched packages

### Phase 85: GitHub PR Review + Coverage Uplift
**Goal**: - `UnifiedCodeReviewProvider.FetchReview` **GitHub** path via mock HTTP - `FetchPRForReview` httptest coverage - Quiet-hours cross-midnight branch
**Depends on**: Phase 84
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. GitHub FetchPRForReview httptest
  2. AI Unified FetchReview GitHub routing
  3. Bot `/ai review` GitHub E2E
  4. race PASS touched packages
  5. `/qa-gate` PASS
**Plans**: 5 plan(s) · source `.planning/phases/85-github-pr-review/PLAN.md`

Plans:
- [x] GitHub FetchPRForReview httptest
- [x] AI Unified FetchReview GitHub routing
- [x] Bot `/ai review` GitHub E2E
- [x] race PASS touched packages
- [x] `/qa-gate` PASS

### Phase 86: Search + Tracker + TMS + Notification Uplift
**Goal**: Поднять пакеты ниже 65%: `search`, `tracker`, `tms`, `notification`.
**Depends on**: Phase 85
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. search ≥65%
  2. tracker ≥65%
  3. race PASS all packages
  4. `/qa-gate` PASS
**Plans**: 4 plan(s) · source `.planning/phases/86-coverage-uplift/PLAN.md`

Plans:
- [x] search ≥65%
- [x] tracker ≥65%
- [x] race PASS all packages
- [x] `/qa-gate` PASS

### Phase 87: TMS + Notification + NATS Uplift
**Goal**: Поднять покрытие `tms`, `notification`, `platform/nats` — закрыть пробелы в consumer events, ent batch repo и NATS publish.
**Depends on**: Phase 86
**Requirements**: NOTIF
**Success Criteria** (what must be TRUE):
  1. nats ≥65%
  2. tms ≥62%
  3. notification stable (+0.1%)
  4. race PASS all packages
  5. `/qa-gate` PASS (backend)
**Plans**: 5 plan(s) · source `.planning/phases/87-tms-notif-nats/PLAN.md`

Plans:
- [x] nats ≥65%
- [x] tms ≥62%
- [x] notification stable (+0.1%)
- [x] race PASS all packages
- [x] `/qa-gate` PASS (backend)

### Phase 88: TMS + Notification ≥65%
**Goal**: Довести `tms` и `notification` до **≥65%** statement coverage.
**Depends on**: Phase 87
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. tms ≥65%
  2. notification ≥65%
  3. race PASS
  4. `/qa-gate` PASS
**Plans**: 4 plan(s) · source `.planning/phases/88-tms-notif-65/PLAN.md`

Plans:
- [x] tms ≥65%
- [x] notification ≥65%
- [x] race PASS
- [x] `/qa-gate` PASS

### Phase 89: E2E Wave 3 + Dashboard Uplift
**Goal**: Закрыть **E2E-01** (5 critical Playwright paths) и поднять `dashboard/` coverage.
**Depends on**: Phase 88
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. E2E-01 five paths documented and runnable
  2. Playwright smoke PASS (frontend-only)
  3. dashboard ≥75%
  4. `/qa-gate` backend PASS
**Plans**: 4 plan(s) · source `.planning/phases/89-e2e-wave3/PLAN.md`

Plans:
- [x] E2E-01 five paths documented and runnable
- [x] Playwright smoke PASS (frontend-only)
- [x] dashboard ≥75%
- [x] `/qa-gate` backend PASS

### Phase 90: IMEX Jira Import Uplift
**Goal**: Поднять `imex/` — ImportJira JSON/handler flows и edge cases.
**Depends on**: Phase 89
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. imex ≥80%
  2. ImportJira handler ≥75%
  3. race PASS imex
  4. `/qa-gate` backend PASS
**Plans**: 4 plan(s) · source `.planning/phases/90-imex-jira/PLAN.md`

Plans:
- [x] imex ≥80%
- [x] ImportJira handler ≥75%
- [x] race PASS imex
- [x] `/qa-gate` backend PASS

### Phase 91: Intranet + Observability Uplift
**Goal**: Поднять `intranet/` и `platform/observability/` до ≥80% — coffee/news handler gaps и metrics allowlist.
**Depends on**: Phase 90
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. intranet ≥80%
  2. observability ≥80%
  3. race PASS intranet + observability
  4. `/qa-gate` backend + frontend PASS
**Plans**: 4 plan(s) · source `.planning/phases/91-intranet-observability/PLAN.md`

Plans:
- [x] intranet ≥80%
- [x] observability ≥80%
- [x] race PASS intranet + observability
- [x] `/qa-gate` backend + frontend PASS

### Phase 92: Messenger Polls + Search Repo
**Goal**: Закрыть 0% gaps в messenger poll repo + handler DeleteChannel/AddMembers; GetSavedFilter в search.
**Depends on**: Phase 91
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. Poll repo functions no longer 0%
  2. DeleteChannel + AddMembers handler covered
  3. GetSavedFilter ent repo covered
  4. messenger/search ≥70% — deferred (needs PG tsvector fixtures)
**Plans**: 4 plan(s) · source `.planning/phases/92-messenger-search/PLAN.md`

Plans:
- [x] Poll repo functions no longer 0%
- [x] DeleteChannel + AddMembers handler covered
- [x] GetSavedFilter ent repo covered
- [x] messenger/search ≥70% — deferred (needs PG tsvector fixtures)

### Phase 93: Auto-Channels NATS + Handler Gaps
**Goal**: Поднять `messenger/` до ≥70% — auto_channels NATS lifecycle, thread replies, revoke invite, poll handlers; search indexer + admin shard.
**Depends on**: Phase 92
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. messenger ≥70%
  2. auto_channels Start/Stop NATS covered
  3. RevokeInvite + ListThreadReplies handler covered
  4. search ≥70% — deferred (PG tsvector)
  5. race PASS
**Plans**: 5 plan(s) · source `.planning/phases/93-autochannels-handlers/PLAN.md`

Plans:
- [x] messenger ≥70%
- [x] auto_channels Start/Stop NATS covered
- [x] RevokeInvite + ListThreadReplies handler covered
- [x] search ≥70% — deferred (PG tsvector)
- [x] race PASS

### Phase 94: Search Meili + Bot + Notification JetStream
**Goal**: Поднять `search/` до ≥70%, `service_bot` coverage, notification consumer Start с JetStream.
**Depends on**: Phase 93
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. search ≥70%
  2. meilisearch indexer 0% → covered
  3. service_bot covered
  4. notification Consumer Start JetStream path
  5. race PASS
**Plans**: 5 plan(s) · source `.planning/phases/94-search-meili-bot/PLAN.md`

Plans:
- [x] search ≥70%
- [x] meilisearch indexer 0% → covered
- [x] service_bot covered
- [x] notification Consumer Start JetStream path
- [x] race PASS

### Phase 95: Notification Push/TG + Tracker Automation + TMS Export
**Goal**: Поднять `notification/`, `tracker/`, `tms/` до ≥70% через push/telegram bind, automation consumer, issue message links, exporter.
**Depends on**: Phase 94
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. notification ≥70%
  2. tracker ≥70%
  3. tms ≥70%
  4. automation_consumer 0% → covered
  5. handler_issue_message_link 0% → covered
  6. exporter Excel/PDF/CSV covered
**Plans**: 7 plan(s) · source `.planning/phases/95-notification-tracker-tms/PLAN.md`

Plans:
- [x] notification ≥70%
- [x] tracker ≥70%
- [x] tms ≥70%
- [x] automation_consumer 0% → covered
- [x] handler_issue_message_link 0% → covered
- [x] exporter Excel/PDF/CSV covered
- [x] race PASS on touched packages

### Phase 96: TMS TestRail/Mongo + Excel Import
**Goal**: Закрыть deferred Phase 95: `mongo_repository`, `testrail_import`, `ImportFromExcel`; подготовить messenger mongo integration tests.
**Depends on**: Phase 95
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. testrail CSV/XML/importTestRailCases covered
  2. ImportTestRailCases handler covered
  3. ImportFromExcel covered
  4. mongo_repository tests (integration, skip-safe)
  5. messenger mongo tests (integration, skip-safe)
  6. race PASS on tms + messenger
**Plans**: 6 plan(s) · source `.planning/phases/96-tms-testrail-mongo/PLAN.md`

Plans:
- [x] testrail CSV/XML/importTestRailCases covered
- [x] ImportTestRailCases handler covered
- [x] ImportFromExcel covered
- [x] mongo_repository tests (integration, skip-safe)
- [x] messenger mongo tests (integration, skip-safe)
- [x] race PASS on tms + messenger

### Phase 97: Mongo CI + Webhook Dispatcher + Marketplace
**Goal**: Mongo integration tests в CI, закрыть 0% webhook dispatcher, marketplace handlers, telegram sendMessage.
**Depends on**: Phase 96
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. CI Mongo service configured
  2. WebhookDispatcher Start/handleMessage/deliverWithRetry covered
  3. marketplace Update/Delete handlers covered
  4. telegram sendMessage 100%
  5. race PASS on touched packages
**Plans**: 5 plan(s) · source `.planning/phases/97-mongo-ci-webhook-marketplace/PLAN.md`

Plans:
- [x] CI Mongo service configured
- [x] WebhookDispatcher Start/handleMessage/deliverWithRetry covered
- [x] marketplace Update/Delete handlers covered
- [x] telegram sendMessage 100%
- [x] race PASS on touched packages

### Phase 98: Accounting + Calendar + COV 75%
**Goal**: Закрыть gaps **accounting** (65%) и **calendar** (65.6%); довести overall `internal/` COV до **≥75%** (deferred из Phase 97).
**Depends on**: Phase 97
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. accounting ≥75%
  2. calendar ≥75%
  3. overall COV ≥75%
  4. race PASS on touched packages
**Plans**: 4 plan(s) · source `.planning/phases/98-accounting-calendar/PLAN.md`

Plans:
- [x] accounting ≥75%
- [x] calendar ≥75%
- [x] overall COV ≥75%
- [x] race PASS on touched packages

### Phase 99: Backup + Package 75% Push
**Goal**: Довести **backup** (65.7%) и пакеты на грани 75% (**notification**, **tms**, **marketplace**) до целевого порога; укрепить overall COV.
**Depends on**: Phase 98
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. backup ≥75%
  2. notification/tms/marketplace ≥75%
  3. overall COV ≥75%
  4. race PASS on touched packages
**Plans**: 4 plan(s) · source `.planning/phases/99-backup-package-75/PLAN.md`

Plans:
- [x] backup ≥75%
- [x] notification/tms/marketplace ≥75%
- [x] overall COV ≥75%
- [x] race PASS on touched packages

### Phase 100: RestoreBackup + Tracker/Messenger/Search 75%
**Goal**: Закрыть **RestoreBackup** S3 happy path; довести **tracker**, **messenger**, **search** до ≥75%; удержать overall COV ≥75%.
**Depends on**: Phase 99
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. RestoreBackup S3 + zip restore path covered
  2. tracker/messenger/search ≥75%
  3. overall COV ≥75%
  4. race PASS on touched packages
**Plans**: 4 plan(s) · source `.planning/phases/100-restore-search-tracker/PLAN.md`

Plans:
- [x] RestoreBackup S3 + zip restore path covered
- [x] tracker/messenger/search ≥75%
- [x] overall COV ≥75%
- [x] race PASS on touched packages

### Phase 101: Sub-75% Package Push
**Goal**: Довести пакеты ниже 75% (**project**, **github**, **crm**, **hr**, **platform/outbox**, **platform/errors**, **integrations/onec**) до порога ≥75%; поднять overall COV.
**Depends on**: Phase 100
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. All target packages ≥75%
  2. overall COV ≥75%
  3. race PASS on touched packages
**Plans**: 3 plan(s) · source `.planning/phases/101-sub75-package-push/PLAN.md`

Plans:
- [x] All target packages ≥75%
- [x] overall COV ≥75%
- [x] race PASS on touched packages

### Phase 102: Dev Routes + Notification Push
**Goal**: Close deferred gaps: `internal/dev` (0%) and `notification` (74.8% → ≥75%)
**Depends on**: Phase 101 (sub-75% package push)
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. `dev` ≥75%
  2. `notification` ≥75%
  3. overall COV ≥76.3% (maintain/improve)
  4. race PASS on touched packages
**Plans**: 8 plan(s) · source `.planning/phases/102-dev-notification-push/PLAN.md`

Plans:
- [x] `methodOrder`, `methodColor`, `guessModule`
- [x] `collectRoutes`, `BuildAPIDocsHTML`, `BuildAPIDocsJSON`
- [x] DevMode guard → 404 when disabled
- [x] Status, Login validation/success, Seed idempotent
- [x] `handleAssigned` invalid JSON/UUID + debounce
- [x] `handleStatusChanged` no assignee, self-move, invalid JSON
- [x] `handleCommentAdded` invalid JSON, author-is-assignee skip
- [x] `handleNewLogin` invalid JSON/UUID, empty device, debounce

### Phase 103: GitHub Push + Mongo Integration Tests
**Goal**: Push boundary packages (github 75%→80%+) and add mongo integration tests (skip without MongoDB)
**Depends on**: Phase 102
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. github ≥78%
  2. COV overall ≥78.2%
  3. race PASS on touched packages (mongo tests skip gracefully)
**Plans**: 8 plan(s) · source `.planning/phases/103-github-mongo-integration/PLAN.md`

Plans:
- [x] `encryptField` round-trip + ciphertext in DB
- [x] `ListAllWebhookSecrets` cache path
- [x] `GetWorkflowRunDetail` httptest GitHub API
- [x] `GetWorkflowRunDetail` HTTP integration
- [x] `resolveExternalPreview` blocked URL + example.com (network)
- [x] CreateBranch / ListGitHubLinks edge cases
- [x] UpdateComment + DeleteComment flow
- [x] Invalid JSON / invalid comment ID / GetActivity

### Phase 104: Flaky GitHub Fix + Mongo Testutil
**Goal**: Eliminate network-dependent github handler tests; unify MongoDB test setup for CI
**Depends on**: Phase 103
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. `TestHandler_RegisterRoutes_IntegrationsCRUD` no network / no timeout
  2. github ≥80%
  3. race PASS github package
**Plans**: 8 plan(s) · source `.planning/phases/104-flaky-github-mongotest/PLAN.md`

Plans:
- [x] `newGitHubRoutesEnv` uses `NewGitHubServiceWithHTTPClient` + empty workflow runs mock
- [x] `TestListWorkflowRuns_MockGitHubAPI` success path
- [x] `TestListWorkflowRuns_ContinuesOnAPIError` 403 → empty list
- [x] `internal/testutil/mongotest/setup.go` — `Connect(t, prefix)`
- [x] Refactor: messenger, tms, calls, search mongo tests
- [x] `TestHandler_RegisterRoutes_IntegrationsCRUD` no network / no timeout
- [x] github ≥80%
- [x] race PASS github package

### Phase 105: Boundary Package Push
**Goal**: Поднять пакеты на грани 75% (**marketplace**, **project**, **tms**) и удержать overall COV ≥78.4%.
**Depends on**: Phase 104
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. marketplace ≥78%
  2. project ≥77%
  3. tms ≥76%
  4. overall COV ≥78.4%
  5. all new tests PASS
**Plans**: 8 plan(s) · source `.planning/phases/105-boundary-package-push/PLAN.md`

Plans:
- [x] `mapServiceError` — все 7 веток (unit + handler)
- [x] `WithdrawListing` / `ownerStatusChange` / `transitionReview` edge cases
- [x] Handler: withdraw 404/403/409/400, approve version conflict, invalid create
- [x] Service: Unarchive/Restore/Delete not-found + internal paths
- [x] `memberCountOrZero` fallback при ошибке MemberCount
- [x] `CreateTeam` constraint + internal error
- [x] Templates: GetTemplate not-found, DeleteTemplate custom success
- [x] Environment handlers: nil UUID, missing name, invalid env ID, delete 404

### Phase 106: Triple Boundary Push (search / messenger / tracker)
**Goal**: Поднять три пакета на грани 75% и overall COV ≥78.7%.
**Depends on**: Phase 105
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. messenger ≥75.4%
  2. tracker ≥75.2%
  3. overall COV ≥78.7%
  4. all new tests PASS
**Plans**: 8 plan(s) · source `.planning/phases/106-triple-boundary-push/PLAN.md`

Plans:
- [x] `Search()` all entity types parallel goroutines
- [x] Admin `searchUsers` → unscoped routing
- [x] `searchChannels` ILIKE fallback path (sqlite)
- [x] Closed DB errors for channel/project ID lookups
- [x] Handler error paths: GetChannel, CreateChannel, Join private, DM, Mute, AddMembers
- [x] ValidateInvite missing token → `valid: false`
- [x] ListComments after create
- [x] Issue link CRUD + time logs + workflow status update

### Phase 107: CRM + Dashboard + Bases Boundary Push
**Goal**: Push `crm`, `dashboard`, `bases` toward 79%+ and lift overall COV past 78.8%.
**Depends on**: Phase 106
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/107-crm-dashboard-bases/PLAN.md`

Plans:
- [x] `.planning/phases/107-crm-dashboard-bases/PLAN.md` (single plan)

### Phase 108: 79% COV Milestone
**Goal**: Cross **79%** overall statement coverage via accounting + search + tms boundary tests.
**Depends on**: Phase 107
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/108-cov-79-milestone/PLAN.md`

Plans:
- [x] `.planning/phases/108-cov-79-milestone/PLAN.md` (single plan)

### Phase 109: Triple Boundary Push (Search + Messenger + Tracker)
**Goal**: Push lowest packages toward 76%+ and maintain COV ≥79%.
**Depends on**: Phase 108
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/109-triple-boundary-push-2/PLAN.md`

Plans:
- [x] `.planning/phases/109-triple-boundary-push-2/PLAN.md` (single plan)

### Phase 110: TMS + AI + Project Boundary Push
**Goal**: Push `tms`, `ai`, `project` toward 78%+ and lift COV toward 79.5%.
**Depends on**: Phase 109
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/110-tms-ai-project/PLAN.md`

Plans:
- [x] `.planning/phases/110-tms-ai-project/PLAN.md` (single plan)

### Phase 111: Notification + HR + CRM Push
**Goal**: Triple push `notification (77.8%) + hr (77.9%) + crm (78%)` → COV **≥79.5%**.
**Depends on**: Phase 110
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/111-notification-hr-crm/PLAN.md`

Plans:
- [x] `.planning/phases/111-notification-hr-crm/PLAN.md` (single plan)

### Phase 112: Search + Messenger + Tracker Push
**Goal**: Triple push `search (75.8%) + messenger (75.8%) + tracker (75.9%)` → COV **≥80%**.
**Depends on**: Phase 111
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/112-search-messenger-tracker/PLAN.md`

Plans:
- [x] `.planning/phases/112-search-messenger-tracker/PLAN.md` (single plan)

### Phase 113: TMS + Project + Sprint + Messenger + Tracker Push
**Goal**: Cross-module push to COV **≥80%** (from 79.7%).
**Depends on**: Phase 112
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/113-tms-project-sprint/PLAN.md`

Plans:
- [x] `.planning/phases/113-tms-project-sprint/PLAN.md` (single plan)

### Phase 114: Search + Messenger + TMS + Tracker Push
**Goal**: Push COV **≥81%** (from 80.0%).
**Depends on**: Phase 113
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/114-search-messenger-tms/PLAN.md`

Plans:
- [x] `.planning/phases/114-search-messenger-tms/PLAN.md` (single plan)

### Phase 115: GitHub/GitLab + Auth + Messenger + Bases + AI Push
**Goal**: Push COV **≥81%** (from 80.2%).
**Depends on**: Phase 114
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/115-github-auth-messenger/PLAN.md`

Plans:
- [x] `.planning/phases/115-github-auth-messenger/PLAN.md` (single plan)

### Phase 116: Search PG Integration + Messenger/TMS Push
**Goal**: COV **≥81%** (from 80.7%).
**Depends on**: Phase 115
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/116-search-messenger-push/PLAN.md`

Plans:
- [x] `.planning/phases/116-search-messenger-push/PLAN.md` (single plan)

### Phase 117: Storage RBAC + Notification Quiet Hours + Tracker Time
**Goal**: Push low packages toward 82%; maintain COV **≥81%** (from 81.6%).
**Depends on**: Phase 116
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/117-storage-notification-tracker/PLAN.md`

Plans:
- [x] `.planning/phases/117-storage-notification-tracker/PLAN.md` (single plan)

### Phase 118: Notification Consumer + Tracker Automation Push
**Goal**: Push `notification`/`tracker` to **≥80%**; maintain COV **≥81%** (from 81.8%).
**Depends on**: Phase 117
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/118-notification-tracker-push/PLAN.md`

Plans:
- [x] `.planning/phases/118-notification-tracker-push/PLAN.md` (single plan)

### Phase 119: Notification/Tracker Push to 82%
**Goal**: Push `notification`/`tracker` toward **82%**; maintain COV **≥81%** (from 81.9%).
**Depends on**: Phase 118
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/119-notification-tracker-82/PLAN.md`

Plans:
- [x] `.planning/phases/119-notification-tracker-82/PLAN.md` (single plan)

### Phase 120: Tracker/TMS/Project Push to 82%
**Goal**: Push `tracker`/`tms`/`project` toward **82%**; maintain COV **≥82%** (from 82.0%).
**Depends on**: Phase 119
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/120-tracker-tms-project-82/PLAN.md`

Plans:
- [x] `.planning/phases/120-tracker-tms-project-82/PLAN.md` (single plan)

### Phase 121: Notification Consumer + Tracker Push to 82%
**Goal**: notification ≥82%, tracker ≥82%, COV ≥82%
**Depends on**: Phase 120
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. See PLAN.md exit criteria / tasks
**Plans**: 1 plan(s) · source `.planning/phases/121-notification-tracker-82/PLAN.md`

Plans:
- [x] `.planning/phases/121-notification-tracker-82/PLAN.md` (single plan)

### Phase 122: TMS + Project ≥82%
**Goal**: tms: 79.3% → ≥82%; project: 80.4% → ≥82%
**Depends on**: Phase 121
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. W1: `handler_wave122_test.go` — suite/run HTTP, ReopenTestRun service
  2. W2: `ent_repository_wave122_test.go` — UpdateSuite, MoveSuite, ReopenTestRun, UpdateTestRunStats
  3. W3: `project/handler_wave122_test.go` — unarchive, invites, trash
  4. W4: `project/service_wave122_test.go` — member role, trash, restore
  5. W5: Docker verify — tms **80.2%**, project **81.0%** (phase 124 closes gap to 82%)
**Plans**: 5 plan(s) · source `.planning/phases/122-tms-project-82/PLAN.md`

Plans:
- [x] W1: `handler_wave122_test.go` — suite/run HTTP, ReopenTestRun service
- [x] W2: `ent_repository_wave122_test.go` — UpdateSuite, MoveSuite, ReopenTestRun, UpdateTestRunStats
- [x] W3: `project/handler_wave122_test.go` — unarchive, invites, trash
- [x] W4: `project/service_wave122_test.go` — member role, trash, restore
- [x] W5: Docker verify — tms **80.2%**, project **81.0%** (phase 124 closes gap to 82%)

### Phase 123: Messenger ≥82%, Storage ≥82%
**Goal**: messenger: 77.6% → ≥82%; storage: 86.8% → ✅ already ≥82%
**Depends on**: Phase 122
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. W1: `ent_repository_wave123_test.go` — mute durations, invites, display name, project members
  2. W2: `service_wave123_test.go` — bot, mute, invites
  3. W3: `handler_wave123_test.go` — HTTP mute/bookmark/invite success paths
  4. W4: `mongo_repository_wave123_test.go` — edge cases (MONGODB_TEST_URL)
  5. W5: Docker verify — messenger **82.3%**, storage **86.8%**
**Plans**: 5 plan(s) · source `.planning/phases/123-messenger-storage-82/PLAN.md`

Plans:
- [x] W1: `ent_repository_wave123_test.go` — mute durations, invites, display name, project members
- [x] W2: `service_wave123_test.go` — bot, mute, invites
- [x] W3: `handler_wave123_test.go` — HTTP mute/bookmark/invite success paths
- [x] W4: `mongo_repository_wave123_test.go` — edge cases (MONGODB_TEST_URL)
- [x] W5: Docker verify — messenger **82.3%**, storage **86.8%**

### Phase 124: Floor ≥82% (tracker / tms / project)
**Goal**: tracker: 81.3% → ≥82%; tms: 81.7% → ≥82%; project: 81.0% → ≥82%
**Depends on**: Phase 123
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. W1: `tracker/repository_wave124_test.go` — CreateIssue all fields, ListIssues filters
  2. W2: `tracker/handler_wave124_test.go` — DeleteWorkflowTransition invalid ID
  3. W3: `tms/service_wave124_test.go` — ent rollback, link, update all fields
  4. W4: `tms/ent_repository_wave124_test.go` — CreateBugIssue, ReopenTestRun
  5. W5: `project/repository_wave124_test.go` — org + outbox Create
  6. W6: `project/service_wave124_test.go` — invite edges, team, restore
**Plans**: 7 plan(s) · source `.planning/phases/124-floor-82/PLAN.md`

Plans:
- [x] W1: `tracker/repository_wave124_test.go` — CreateIssue all fields, ListIssues filters
- [x] W2: `tracker/handler_wave124_test.go` — DeleteWorkflowTransition invalid ID
- [x] W3: `tms/service_wave124_test.go` — ent rollback, link, update all fields
- [x] W4: `tms/ent_repository_wave124_test.go` — CreateBugIssue, ReopenTestRun
- [x] W5: `project/repository_wave124_test.go` — org + outbox Create
- [x] W6: `project/service_wave124_test.go` — invite edges, team, restore
- [x] W7: Docker verify

### Phase 125: COV ≥82%, Security ≥80%, Regression Gate
**Goal**: COV overall: ≥82% → ✅ **82.6%**; auth: ≥80% → ✅ **82.4%**; ws: ≥80% → ✅ **87.7%**; rbac: ≥80% → ✅ **84.7%**
**Depends on**: Phase 124
**Requirements**: see PLAN.md
**Success Criteria** (what must be TRUE):
  1. ✅ COMPLETE (2026-06-21)
**Plans**: 1 plan(s) · source `.planning/phases/125-regression-gate/PLAN.md`

Plans:
- [x] `.planning/phases/125-regression-gate/PLAN.md` (single plan)
