# FEATURE BACKLOG 2026 — Hub v2.5 → v3.0

> Детальный backlog для dev-команды · 2026-06-21  
> Источники: competitive_analysis_2026.md, user_needs_analysis_2026.md, PROD-01-SPIKE, MILESTONE-3  
> **PO LOCKED 2026-06-21:** `.planning/DISCUSS-ROADMAP-LOCKED.md`  
> **Out of scope (still):** SSO, native apps (PWA OK), SaaS billing (later), multi-org (v5)  
> **In scope by phase:** Gantt/webcam (16), email hub-mail (15.2+), AI (17), v4 accounting/no-code DB

---

## Waves overview

| Wave | Timeline | Theme | Items |
|------|----------|-------|-------|
| **QW** | 2 weeks | Quick wins post v2.0.2 | FB-01..FB-06 |
| **v2.5** | 4–6 weeks | Production hardening + UX | FB-07..FB-14 |
| **v2.6 Core** | 8–12 weeks | Differentiation | FB-15..FB-22 |
| **v3.0 AI** | 12+ weeks | MILESTONE-3 | FB-23..FB-28 |

---

## Quick Wins (2 weeks)

### FB-01 — Shortcuts overlay (`?`)

| Field | Value |
|-------|-------|
| **Persona** | Developer, Member |
| **Problem** | Пользователи не знают Ctrl+K, hotkeys — снижение productivity vs Linear |
| **User story** | Как пользователь, хочу нажать `?` и увидеть список горячих клавиш |
| **Acceptance** | Given authenticated, When press `?`, Then modal с группами (nav, tracker, messenger); Escape закрывает; i18n ru/en |
| **Technical scope** | FE: `shortcuts-dialog.tsx`, hook global keydown; route `/help` link; no BE |
| **Effort** | S · **P1** · v2.5 |
| **Dependencies** | help.json exists |
| **Metrics** | Help page views +30% |
| **Risk** | Conflict with Plate editor `?` — guard when focus in input |

### FB-02 — Skeleton loaders (tables/board)

| Field | Value |
|-------|-------|
| **Persona** | All |
| **Problem** | PLSH-03: текст «Загрузка...» вместо skeleton — perceived slowness |
| **Acceptance** | Given slow query, When loading issues/cases/messages list, Then skeleton rows 3–5; no layout shift |
| **Technical scope** | FE: extend `QueryStateView`, kanban-board, case-table, channel-list |
| **Effort** | S · **P2** · v2.5 |
| **Metrics** | Lighthouse CLS ≤ 0.1 |

### FB-03 — Rate limit countdown in toast

| Field | Value |
|-------|-------|
| **Persona** | All |
| **Problem** | 429 без понятного retry-after |
| **Acceptance** | Given API 429 with Retry-After, When toastApiError, Then show «Повторите через N сек» + i18n |
| **Technical scope** | FE: `format-api-error.ts`; BE: ensure Retry-After header on rate limit middleware |
| **Effort** | S · **P1** · v2.5 |
| **Dependencies** | ratelimit.go |

### FB-04 — Ctrl+N new issue (global)

| Field | Value |
|-------|-------|
| **Persona** | Developer |
| **Problem** | US-C03 partial — нет global create |
| **Acceptance** | Given project context or last project in store, When Ctrl+N, Then issue-form dialog opens |
| **Technical scope** | FE: command-palette action + hotkey in app-shell |
| **Effort** | S · **P2** · v2.5 |

### FB-05 — Admin: rate limit hit viewer

| Field | Value |
|-------|-------|
| **Persona** | Admin |
| **Problem** | Нет visibility кто упёрся в лимит |
| **Acceptance** | Given admin, When open /admin/audit filtered entity=rate_limit, Then list user_id, endpoint, timestamp last 24h |
| **Technical scope** | BE: audit event on rate limit; FE: filter preset in audit-log |
| **Effort** | M · **P2** · v2.5 |
| **Dependencies** | audit middleware |

### FB-06 — Lighthouse CI artifact (manual gate)

| Field | Value |
|-------|-------|
| **Persona** | Admin |
| **Problem** | G-05 Lighthouse ≥90 not verified |
| **Acceptance** | Script `npm run lighthouse:ci` outputs JSON; documented thresholds in test_plan |
| **Technical scope** | FE: scripts/lighthouse-ci.mjs; docs only in CI (not blocking initially) |
| **Effort** | S · **P1** · v2.5 |

---

## v2.5 — Production hardening

### FB-07 — E2E-02 Project CRUD flow

| Field | Value |
|-------|-------|
| **Persona** | PO, Admin |
| **Acceptance** | Create project → verify board + channel exist → archive → trash restore |
| **Technical scope** | `e2e/project-lifecycle.spec.ts`; fixtures projectFactory |
| **Effort** | M · **P0** · v2.5 |
| **Metrics** | CI e2e-full green |

### FB-08 — E2E-03 Messenger send + thread

| Field | Value |
|-------|-------|
| **Acceptance** | Login → create channel → send message → reply thread → reaction |
| **Technical scope** | `e2e/messenger.spec.ts`; WS wait helpers |
| **Effort** | M · **P0** · v2.5 |

### FB-09 — E2E-04 TMS case + run

| Field | Value |
|-------|-------|
| **Acceptance** | Create manual case → start run → mark pass/fail → create bug link |
| **Effort** | M · **P0** · v2.5 |

### FB-10 — E2E-05 Calls join smoke

| Field | Value |
|-------|-------|
| **Acceptance** | Schedule call → join room → mute → leave (mock LiveKit if needed) |
| **Effort** | M · **P1** · v2.5 |
| **Note** | May skip in CI without LiveKit — tag `@livekit` |

### FB-11 — RHF + Zod: issue-form

| Field | Value |
|-------|-------|
| **Persona** | Developer |
| **Problem** | Inline validation partial |
| **Acceptance** | Zod schema title 3–200 chars, type required; errors i18n; submit disabled invalid |
| **Technical scope** | FE: add zod + react-hook-form; migrate issue-form.tsx |
| **Effort** | M · **P1** · v2.5 |
| **Out of scope** | All forms at once |

### FB-12 — RHF + Zod: onboarding + new project

| Field | Value |
|-------|-------|
| **Acceptance** | Extend isProjectStepValid to Zod; new.tsx join.tsx same pattern |
| **Effort** | M · **P1** · v2.5 |
| **Dependencies** | FB-11 patterns |

### FB-13 — WCAG audit sign-off

| Field | Value |
|-------|-------|
| **Acceptance** | axe e2e ≥15 routes; zero critical; document exceptions in test_plan |
| **Effort** | M · **P0** · v2.5 |
| **Dependencies** | a11y.spec.ts expansion |

### FB-14 — PG-03 Secret rotation execution

| Field | Value |
|-------|-------|
| **Persona** | Admin |
| **Acceptance** | Checklist in PG-03-SECRET-ROTATION.md executed; livekit-prod templated |
| **Effort** | S · **P0** · v2.5 (ops) |

---

## v2.6 Core — Differentiation

### FB-15 — PROD-01 Workflow Automation Lite ⭐

| Field | Value |
|-------|-------|
| **Persona** | PO, Team Lead, DevOps |
| **Problem** | #1 gap vs Jira — нет «when status Done → notify channel» |
| **User story** | Как PO, хочу правило: при переходе в Done → сообщение в #releases |
| **Acceptance (G/W/T)** | **Given** rule enabled status→Done→channel, **When** issue moved on Kanban, **Then** system message in channel ≤5s via WS; **Given** rule disabled, **Then** no message |
| **Technical scope** | |
| BE | `ent/schema/automationrule.go`; migration; `tracker/automation.go` RunRules; hooks in MoveIssue; API CRUD `/automation-rules`; RBAC PermProjectManageWorkflow |
| FE | Tab Automation on `/settings/workflow`; rule builder dropdowns; i18n `tracker:automation.*` |
| Tests | Unit engine 90%+; handler tests; e2e one rule smoke |
| **Effort** | **L** · **P0** · v2.6 |
| **Dependencies** | outbox/NATS, messenger.CreateMessage |
| **Risks** | Infinite loops — max 1 action per event; idempotency key |
| **Metrics** | ≥30% projects with ≥1 rule in 90 days |
| **MVP triggers** | status_changed, issue_created, issue_assigned |
| **MVP actions** | post_to_channel, notify_user, set_field (v2.6.1) |
| **Out of scope** | JQL conditions, cron, visual graph |

### FB-16 — PROD-02 Flaky Test Detection

| Field | Value |
|-------|-------|
| **Persona** | QA Lead |
| **Problem** | US-C02; Allure TestOps ML flaky |
| **Acceptance** | Given ≥5 runs in 14 days, When success rate <100% and >0%, Then case flagged flaky; dashboard widget top 10 flaky |
| **Technical scope** | BE: `tms/flaky.go` analyzer job daily; store `flaky_score` on case; FE: badge + trends chart |
| **Effort** | L · **P0** · v2.6 |
| **Metrics** | QA weekly report uses flaky widget |

### FB-17 — PROD-03 Issue ↔ Message Bidirectional Sync

| Field | Value |
|-------|-------|
| **Persona** | Developer, Member |
| **Problem** | Deep links есть; нет sync comments thread↔issue |
| **Acceptance** | Given linked thread, When comment in Hub issue, Then optional mirror in thread (setting per link) |
| **Technical scope** | BE: `issue_message_link` table; sync service; FE: link UI in thread header |
| **Effort** | L · **P1** · v2.6 |
| **Risks** | Loop prevention; feature flag |

### FB-18 — PROD-04 Webhook Delivery Dashboard

| Field | Value |
|-------|-------|
| **Persona** | PO, DevOps |
| **Acceptance** | List outbound webhooks: status, payload hash, retry count, last error; manual retry |
| **Technical scope** | BE: `webhook_delivery` ent; apigateway persistence; FE: settings/webhooks tab |
| **Effort** | M · **P1** · v2.6 |

### FB-19 — Shared JQL Saved Filters (US-C04)

| Field | Value |
|-------|-------|
| **Acceptance** | Save filter → share with project role; appear in issue list sidebar |
| **Technical scope** | BE: saved_filter.share_scope; FE: saved-filters.tsx share dialog |
| **Effort** | M · **P1** · v2.6 |

### FB-20 — GitHub Actions Status in Issue (US-C06)

| Field | Value |
|-------|-------|
| **Acceptance** | Sidebar shows latest workflow run status for linked PR branch |
| **Technical scope** | BE: github service GetWorkflowRuns; FE: pipeline-status widget extend |
| **Effort** | M · **P2** · v2.6 |
| **Dependencies** | GitHub token scopes |

### FB-21 — OpenTelemetry Export

| Field | Value |
|-------|-------|
| **Persona** | Admin, DevOps |
| **Acceptance** | OTEL_EXPORTER_OTLP_ENDPOINT set → traces for HTTP handlers + service spans |
| **Technical scope** | BE: `platform/observability/otel.go`; middleware span; deploy env docs |
| **Effort** | L · **P2** · v2.6 |

### FB-22 — Notification Digest (daily)

| Field | Value |
|-------|-------|
| **Persona** | Member |
| **Acceptance** | User opts in → 09:00 local summary Telegram/in-app of unread mentions |
| **Technical scope** | BE: cron digest job; FE: notification-settings toggle |
| **Effort** | M · **P2** · v2.6 |
| **Note** | Not email — Telegram + in-app only per constraints |

---

## v2.7 (stretch)

### FB-22b — Custom Fields Builder UI

| Effort | XL · P2 · v2.7 |
| Ent `custom_field_definition` exists — UI wizard for PO |

### FB-22c — Timeline Roadmap

| Effort | L · P2 · v2.7 |
| Epic/initiative bars like Linear Initiatives |

### FB-22d — Triage Inbox

| Effort | M · P2 · v2.7 |
| Incoming bugs queue with assign/reject |

---

## v3.0 — AI Chat Hub (MILESTONE-3 subset)

### FB-23 — AI Provider Abstraction (M3-001..008)

| Field | Value |
|-------|-------|
| **Acceptance** | Admin selects GigaChat/YandexGPT/DeepSeek/Ollama; fallback chain; rate limits |
| **Technical scope** | BE: `internal/ai/provider.go`; config encrypted keys; audit AI requests (no content log) |
| **Effort** | XL · **P0** · v3.0 |

### FB-24 — @hub-ai Messenger Bot (M3-009..014)

| Field | Value |
|-------|-------|
| **Acceptance** | DM + @mention; streaming; markdown+code copy; 20 msg context |
| **Technical scope** | BE: bot user, WS stream; FE: bot message renderer |
| **Effort** | XL · **P0** · v3.0 |

### FB-25 — /ai test-cases command (M3-016)

| Field | Value |
|-------|-------|
| **Persona** | QA |
| **Acceptance** | `/ai test-cases HUB-123` → draft cases → import to TMS |
| **Effort** | L · **P1** · v3.0 |

### FB-26 — AI Panel Ctrl+Shift+A (M3-022)

| Effort | L · P1 · v3.0 |

### FB-27 — Inline AI in issue editor (M3-023)

| Effort | M · P2 · v3.0 |

### FB-28 — RAG project docs (M3-021)

| Effort | XL · P1 · v3.0 |
| Vector store pgvector or mongo embeddings |

---

## v3.5 — AI Provider Hub (multi-model)

### AI-01 — Provider registry + admin UI

| Field | Value |
|-------|-------|
| **Spec** | `.planning/AI-PROVIDER-HUB-SPEC.md` |
| **Acceptance** | Admin CRUD providers; encrypted keys; test ping; DeepSeek + Ollama presets |
| **Effort** | L · **P0** · v3.0 |

### AI-02 — Model picker + OpenRouter preset

| **Acceptance** | User selects any admin-enabled model; OpenRouter «100 models» optional |
| **Effort** | M · **P0** · v3.5 |

### AI-03 — OpenAI-compatible `/api/ai/v1/chat/completions`

| **Acceptance** | Hub JWT auth; proxies to selected provider; streaming SSE |
| **Effort** | M · **P1** · v3.5 |

### AI-04 — Anthropic + OpenAI adapters (admin opt-in)

| **Note** | VPN disclaimer in admin UI |
| **Effort** | S · **P2** · v3.5 |

---

## v4.0 — Corporate modules

### MOD-01 — Intranet news feed

| Persona | All employees |
| **Acceptance** | Admin/HR posts news; pin; reactions; RBAC publish |
| **BE** | `internal/intranet/` or extend `project` → `org` scope |
| **Effort** | L · **P1** · v4.0 |

### MOD-02 — Org directory + org chart

| **Acceptance** | Tree by department; profile cards; search in Ctrl+K |
| **Effort** | M · **P1** · v4.0 |

### MOD-03 — Kudos / thanks

| **Acceptance** | Send kudos → recipient notif + optional public feed |
| **Effort** | S · **P2** · v4.0 |

### MOD-04 — Unified calendar module

| **Acceptance** | `/calendar` — sprints, calls, deadlines; iCal export |
| **Effort** | L · **P1** · v4.0 |

### MOD-05 — CRM-lite (contacts + deals)

| **Acceptance** | Kanban deals; link to project; no invoicing |
| **Out of scope** | Full accounting |
| **Effort** | L · **P2** · v4.0 |

### MOD-06 — HR-lite (leave requests)

| **Acceptance** | Submit leave → approver chain via PROD-01; calendar block |
| **Effort** | M · **P2** · v4.0 |

### MOD-07 — 1C connector (webhook)

| **Acceptance** | Approved leave → POST to 1C endpoint; mapping config |
| **Effort** | L · **P3** · v4.5 |
| **Service** | hub-integrations microservice |

### MOD-08 — Internal marketplace (classifieds)

| **Acceptance** | Post listing; mod queue; optional #marketplace channel |
| **Effort** | M · **P3** · v4.0 |

---

## Security roadmap (SEC-R*)

| ID | Item | Milestone |
|----|------|-----------|
| SEC-R1 | PG-03 rotation | v2.5 |
| SEC-R2 | AI key encryption + SSRF | v3.0 |
| SEC-R8 | Prompt injection filters | v3.5 |

---

## NOTIF — Telegram bind + Email alerts (v2.5–v2.6)

> Spec: [NOTIF-TG-EMAIL-SPEC.md](./NOTIF-TG-EMAIL-SPEC.md)

| ID | Feature | Effort | Priority |
|----|---------|--------|----------|
| NOTIF-01 | ADR email transactional (Resend alerts) | S | P0 |
| NOTIF-02 | Telegram bind deep link + bot `/start bind_` | M | P0 |
| NOTIF-03 | Telegram `/link` code + confirm API | S | P0 |
| NOTIF-04 | Unlink, status, test ping, audit | S | P1 |
| NOTIF-05 | EmailSender + run_failed templates | M | P1 |
| NOTIF-06 | Channel `email` in preferences UI | S | P1 |
| NOTIF-07 | Verify real email in profile (non-placeholder) | M | P1 |
| NOTIF-08 | `run_failed` / `run_regression` event types | M | P0 |
| NOTIF-09 | TelegramLinkCard + settings UX | M | P1 |
| NOTIF-10 | E2E notification bind smoke | M | P2 |

---

## Dependency graph (critical path)

```
v2.0.2 SHIPPED
    │
    ├─► FB-07..10 E2E ──► FB-13 WCAG
    │
    ├─► FB-11 RHF ──► FB-12 onboarding
    │
    └─► FB-15 PROD-01 ──► FB-20 CI automation triggers
              │
              ├─► FB-17 issue↔msg
              └─► FB-18 webhooks
                        │
                        v2.6 SHIP ──► FB-23..28 AI v3.0 ──► AI-01..04 ──► MOD-01..08 v4.0
```

---

## Extended dependency graph note

See [PLATFORM-VISION-v4.md](./PLATFORM-VISION-v4.md) for microservices split triggers.

---

## Priority matrix (RICE simplified)

| ID | Reach | Impact | Confidence | Effort | Score |
|----|-------|--------|------------|--------|-------|
| FB-15 PROD-01 | 5 | 5 | 4 | 8 | **12.5** |
| FB-07 E2E project | 5 | 4 | 5 | 4 | **25** |
| FB-16 Flaky | 3 | 5 | 4 | 8 | **7.5** |
| FB-23 AI provider | 5 | 5 | 3 | 13 | **5.8** |
| FB-01 Shortcuts | 5 | 2 | 5 | 2 | **25** |

---

## Verification checklist (ba-ver)

- [x] No forbidden features (SSO, Gantt, webcams, billing)
- [x] Personas mapped from vision_and_scope.md
- [x] PROD-01 aligned with PROD-01-SPIKE.md
- [x] v3.0 aligned with MILESTONE-3.md
- [x] ≥25 items with acceptance criteria
- [x] Technical scope names real Hub modules

---

*Maintainers: update after each milestone ship · Link from IMPROVEMENT-MASTER-PLAN.md Wave PRODUCT*
