# Roadmap: Hub

> **Текущая product-truth (founder, 2026-08-12):** Hub остаётся
> незавершённым MVP. Ни один продуктовый модуль не принят как `complete`,
> `stable` или `production-ready`. Метки `SHIPPED`/`COMPLETE`, milestone/audit
> `PASS` и строки «реализовано» ниже фиксируют только историческое выполнение
> конкретных planning/technical slices; они не подтверждают готовность продукта,
> модуля или production. PG-03/FA-06 — отдельный manual ops-блокер production
> cutover, а не индикатор готовности MVP.

## Overview

Hub проектируется как единая платформа для команд разработки: общий фундамент,
продуктовые модули и сквозные функции развиваются по цепочке зависимостей.
Закрытие фазы означает завершение ограниченного planning/technical slice по её
критериям; оно не является приёмкой всего модуля, продукта или production.

## Milestones

- :white_check_mark: **v1.0 MVP** - Phases 1-7 (shipped 2026-03-21)
- :white_check_mark: **v2.0 Hub Redesign** - Phases 8-14 (shipped 2026-05-15)
- :white_check_mark: **v2.5 Trust** - Phase 15 (shipped 2026-06-22)
- :white_check_mark: **v2.6 Automate** - Phase 16 (shipped 2026-06-22)
- :white_check_mark: **v3.0 AI Native** - Phases 17–19 (shipped 2026-06-22)
- :white_check_mark: **v4 Super-app** - Phases 39–58 (shipped 2026-06-25, branch uncommitted)
- :calendar: **v4.5+** - Accounting advanced, Lark Base formulas · 2027+

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

> **Phase artifact locations:** Phases **01–13** live under `.planning/milestones/` (`v1.0-phases/` for 01–07, `v2.0-phases/` for 08–13). Phases **14–125** (and later TF/backlog dirs) live under `.planning/phases/`. Convention `total_phases: 125` = milestones 01–13 (13) + phases dirs 14–125 (112) — see `planning-audit-2026-07/00-PLANNING-AUDIT.md` §0 and `planning-audit-2026-07/PHASE-STATUS-MATRIX.md`.

<details>
<summary>v1.0 MVP (Phases 1-7) - SHIPPED 2026-03-21</summary>

- [x] **Phase 1: Foundation & Infrastructure** - Аутентификация, проекты, RBAC, инфраструктура реального времени, хранилище, среда разработки
- [x] **Phase 2: Task Tracker** - CRUD задач, Kanban-доски, рабочие процессы, комментарии, лог активности, учёт времени (completed 2026-03-21)
- [x] **Phase 3: Messenger** - Личные сообщения, каналы, треды, реакции, голосовые сообщения, присутствие, поиск
- [x] **Phase 4: TMS & Allure** - Тест-кейсы, тестовые прогоны, загрузка Allure-отчётов, связь с задачами, отчётность, импорт/экспорт
- [x] **Phase 5: Voice Calls** - Аудиозвонки, демонстрация экрана, управление звонками, эфемерный чат (completed 2026-03-21)
- [x] **Phase 6: Notifications, Search & Dashboards** - Уведомления в приложении, глобальный поиск, персональные/проектные/административные дашборды, аудит
- [x] **Phase 7: Polish & Extended Features** - Спринты, интеграция с GitHub, запись звонков, интернационализация, темы, PWA, API, резервное копирование (completed 2026-03-21)

</details>

### v2.0 Hub Redesign (Complete 2026-05-15)

**Milestone Goal:** Pixel-perfect редизайн всего фронтенда Hub по 11 HTML-макетам без ломания существующего функционала

- [x] **Phase 8: Design System Foundation** - CSS-токены, шрифты, !important-очистка, тёмная и светлая палитры, радиусы, тени, анимации (drift fixed 2026-07-03: подтверждено по STATE/AGENTS)
- [x] **Phase 9: Layout Shell** - Сайдбар, топбар, AppShell с новым дизайном (drift fixed 2026-07-03: подтверждено по STATE/AGENTS)
- [x] **Phase 10: UI Primitives** - Все Shadcn-компоненты рестайлированы под новый дизайн-язык (drift fixed 2026-07-03)
- [x] **Phase 11: Core Mockup Pages** - Login, Dashboard, Messenger, Kanban (drift fixed 2026-07-03)
- [x] **Phase 12: Remaining Mockup Pages + Unmocked Pages** - Issue Detail, Calls, TMS, Settings, Admin, Onboarding, Trash + Projects, Sprint, JQL, Project Settings, Project Dashboard (drift fixed 2026-07-03: подтверждено по STATE/AGENTS)
- [x] **Phase 13: Polish & QA** - Финальный визуальный QA, доступность, мобильная проверка, очистка мёртвого CSS (drift fixed 2026-07-03)
- [x] **Phase 14: Architectural Hardening** - Расщепление god-объектов (AuthService, CallService), удаление ent.Client leak из сервисов, расширение Permission, покрытие тестами критичных путей

### v2.5+ Production & Super-app (Planned — PO LOCKED 2026-06-21)

> Decisions: `.planning/archive/roadmaps/DISCUSS-ROADMAP-LOCKED.md` · Branch: `feat/v2.5-trust`

- [x] **Phase 15: Trust (v2.5)** - Magic Link only, organization, NOTIF TG bind, hub-mail, E2E FB-07..09, WCAG, run_failed, Biome, LiveKit CI, coverage ≥50% (2026-06-22)
- [x] **Phase 16: Automate (v2.6)** - PROD automation, flaky, issue↔message, webhooks log, Gantt, webcam, email marketing, triage inbox, Meilisearch (2026-06-22)
- [x] **Phase 17: AI Native (v3.0)** - AI gateway (Ollama + BYOK), @hub-ai + dialog, full JQL automation (W1–W11, 2026-06-22)
- [x] **Phase 18: AI Privacy (v3.5)** - M3-030 external AI consent, settings UI, E2E ai-hub (2026-06-22)
- [x] **Phase 19: AI Polish** - M3-024 composer Tab suggest, dev seed Ollama, MILESTONE-3 reconcile (2026-06-22)
- [x] **Phase 20: Milestone Gate** - ROADMAP sync, composer E2E, COVERAGE refresh (2026-06-22)
- [x] **Phase 21: Architecture Completion** - backup Repository (ARCH-01 W1), cron unify (2026-06-22)
- [x] **Phase 22: Architecture storage** - storage + orphan cleanup Repository (2026-06-22)
- [x] **Phase 23: Architecture apigateway** - APIKey + Webhook Repository (2026-06-23)
- [x] **Phase 24: Architecture imex** - import/export Repository, Jira atomic keys (2026-06-23)
- [x] **Phase 25: Architecture dashboard** - dashboard Repository (2026-06-23)
- [x] **Phase 26: Architecture rbac** - RBAC + custom roles Repository (2026-06-23)
- [x] **Phase 27: Architecture auth** - auth Repository, middleware repo (2026-06-23) — **ARCH-01 17/17**
- [x] **Phase 28: v2.6 Architecture Gate** - rbac RequireAdmin→Repository, regression (2026-06-23)
- [x] **Phase 29: OBS-01** - Timed metrics + `/admin/observability/ops` (2026-06-23)
- [x] **Phase 30: OBS-01 FE** - Admin metrics tab + Vitest + E2E smoke (2026-06-23)
- [x] **Phase 31: OBS-01 Prometheus** - `/metrics` + HTTP middleware + E2E (2026-06-23)
- [x] **Phase 32: OBS-01 OTel** - OTLP traces + HTTP/Timed spans (2026-06-24)
- [x] **Phase 33: Monitoring stack** - Grafana/Prometheus compose + metrics CIDR + K3s NP (2026-06-24)
- [x] **Phase 34: Loki + alerts** - Loki/Promtail + Prometheus rules + logs panel (2026-06-24)
- [x] **Phase 35: Alertmanager** - AM routing + Grafana datasource + metrics-down (2026-06-24)
- [x] **Phase 36: Alert webhook** - AM → audit log + prod env wiring (2026-06-24)
- [x] **Phase 37: Biome gate** - PG-02 lint 0 errors / 0 warnings (2026-06-24)
- [x] **Phase 38: v2.6 Milestone Gate** - API docs + qa-gate + alert E2E (2026-06-24)
- [x] **Phase 39: Intranet News** - Hub One v4 kickoff — corporate news feed (MOD-INTRA first slice). (2026-06-24)
- [x] **Phase 40: Org directory** - MOD-02 directory + Ctrl+K (2026-06-24)
- [x] **Phase 41: Unified calendar** - MOD-04 `/calendar` + iCal (2026-06-24)
- [x] **Phase 42: CRM-lite** - MOD-05 contacts + deals kanban (2026-06-24)
- [x] **Phase 43: HR-lite** - MOD-06 leave requests + calendar (2026-06-24)
- [x] **Phase 44: Kudos** - MOD-03 thanks feed + notifications (2026-06-24)
- [x] **Phase 45: Marketplace** - MOD-08 classifieds + mod queue (2026-06-24)
- [x] **Phase 46: Status page** - V4-06 public /status (2026-06-24)
- [x] **Phase 47: Random Coffee** - V4-05 intranet matching (2026-06-24)
- [x] **Phase 48: 1C Connector** - MOD-07 leave webhook + admin settings (2026-06-24)
- [x] **Phase 49: Lark Base** - V4-04 no-code tables `/bases` (2026-06-25)
- [x] **Phase 50: v4 Milestone Gate** - qa-gate + Vitest v4 smoke (2026-06-25)
- [x] **Phase 51: Accounting MVP** - V4-03 ledger + trial balance (2026-06-25)
- [x] **Phase 52: Accounting v4.5** - invoices + P&L + CRM won hook (2026-06-25)
- [x] **Phase 53: v4 Module Gate** - CRM↔invoice tests + E2E v4.5 (2026-06-25)
- [x] **Phase 54: v4 Coverage Uplift** - bases 61% + marketplace 69% (2026-06-25)
- [x] **Phase 55: onec + intranet** - 72% + 61% coverage (2026-06-25)
- [x] **Phase 56: v4 Final Gate** - CRM 73% + qa-gate green (2026-06-25)
- [x] **Phase 57: Post-v4 CI Handoff** - v4 E2E in CI + admin fix (2026-06-25)
- [x] **Phase 58: v4 Milestone Archive** - SUMMARY + qa-gate `-race` + coverage refresh (2026-06-25)
- [x] **Phase 59: Coverage + UX Close-out** - tracker 52% + storage 56% + issue-form Vitest (2026-06-25)
- [x] **Phase 60: UX-1 Gate** - a11y calls/TMS + shortcuts E2E + onboarding Zod (2026-06-25)
- [x] **Phase 61: Coverage + Context Tooltips** - project 59% + search 61% + UX-P1-04 (2026-06-25)
- [x] **Phase 62: Onboarding Profile RHF** - UX-P1-07 profile step RHF+Zod + E2E (2026-06-25)
- [x] **Phase 63: Production Release Gate** - qa-gate -race + 55.7% + VERIFICATION (2026-06-25)
- [x] **Phase 64: Auth Coverage Uplift** - auth 74.7% → **81.6%** (2026-06-25)
- [x] **Phase 65: UX-P1-05 Rate Limit Timer + Messenger Coverage** - 1. **UX-P1-05 / FB-03:** Live countdown toast on 429 + actionable `request_id` in error toasts 2. **Coverage:** `mess… (2026-06-25)
- [x] **Phase 66: Notification Coverage Uplift (42.5% → 50%)** - Close COV gap: `notification/` package **≥50%** + telegram handler route fix. (2026-06-25)
- [x] **Phase 67: Production Release Gate (Phases 64–66)** - Re-run full `/qa-gate` after Phases 64–66; publish `PRODUCTION-HANDOFF.md`; refresh coverage baseline. (2026-06-25)
- [x] **Phase 68: Messenger Fan-out & Coverage (52.3% → 60%)** - PLAN-CLOSE-GAPS **2.5**: messenger fan-out integration + coverage **≥60%**. (2026-06-26)
- [x] **Phase 69: Apigateway Coverage + COV Baseline Refresh** - 1. **SEC-01 / API-01:** `apigateway/` **54% → ≥60%** 2. Refresh overall COV baseline post Phase 68 (2026-06-26)
- [x] **Phase 70: GitLab/GitHub Coverage (SEC-04)** - 1. **SEC-04 / GH-01:** `gitlab/` and `github/` **48% / 45% → ≥50%** 2. Cover `EntRepository` (0%) and `RegisterRoutes… (2026-06-21)
- [x] **Phase 71: Mail + Dashboard Coverage (MAIL-01 / BE-H3)** - 1. **MAIL-01 / Phase 15.2:** `mail/` **44.6% → ≥50%** — hub-mail handlers + admin API 2. **BE-H3:** `dashboard/` proj… (2026-06-21)
- [x] **Phase 72: Production QA Gate (Post Phase 71)** - Full `/qa-gate` re-run after Phases 70–71 (gitlab/github/mail/dashboard coverage). (2026-06-21)
- [x] **Phase 73: Platform HTTP/Outbox Coverage (SEC-04 / NF-07)** - 1. `platform/http` **20.4% → ≥50%** — SafeDialer + OpenGraph (SSRF / NF-07) 2. `platform/outbox` **28.6% → ≥50%** — P… (2026-06-21)
- [x] **Phase 74: AI Module Coverage (M3 / COV-01)** - `internal/ai/` **24.8% → ≥50%** — handlers, repository, settings, privacy, adapters (unit paths without live LLM). (2026-06-21)
- [x] **Phase 75: Post-AI QA Gate Refresh** - Re-run full `/qa-gate` after Phase 74 AI coverage additions; refresh overall COV baseline. (2026-06-21)
- [x] **Phase 76: Sprint Ent + Handler Routes Coverage** - `internal/sprint/` **58.0% → ≥60%** — EntRepository + RegisterRoutes (RBAC) (2026-06-21)
- [x] **Phase 77: Storage Handler Routes (CR-04 / Presign)** - `internal/storage/` **55.7% → ≥60%** — NewHandler, RegisterRoutes, RBAC authorizeProject, NewStorageService (2026-06-21)
- [x] **Phase 78: Tracker + TMS Coverage Wave** - - `internal/tracker/` **52.1% → ≥60%** - `internal/tms/` **55.5% → ≥60%** (2026-06-21)
- [x] **Phase 79: Notification + Project Coverage Wave** - - `internal/notification/` **55.7% → ≥60%** - `internal/project/` **59.4% → ≥60%** (2026-06-21)
- [x] **Phase 80: Backup + HR + IMEX Coverage Wave** - - `internal/backup/` **53.4% → ≥60%** - `internal/hr/` **53.1% → ≥60%** - `internal/imex/` **57.8% → ≥60%** (2026-06-21)
- [x] **Phase 81: AI Coverage + QA Gate** - - `internal/ai/` **50.7% → ≥60%** - Full `/qa-gate` PASS on branch (2026-06-21)
- [x] **Phase 82: AI Bot Hook + Doc Search RAG** - - Close Phase 81 deferred: `@hub-ai` messenger bot hook + `SearchDocProvider` full path - `internal/ai/` **68.5% → ≥7… (2026-06-21)
- [x] **Phase 83: AI Bot Commands + Search Query** - - Complete `@hub-ai` slash commands (test-cases, review, explain, translate, sql) - `internal/search/` **60.5% → ≥60%… (2026-06-21)
- [x] **Phase 84: VCS Review + Bases + Intranet Coffee** - - `UnifiedCodeReviewProvider.FetchReview` GitLab path via mock HTTP - `internal/bases/` **60.8% → ≥65%** - `internal/… (2026-06-21)
- [x] **Phase 85: GitHub PR Review + Coverage Uplift** - - `UnifiedCodeReviewProvider.FetchReview` **GitHub** path via mock HTTP - `FetchPRForReview` httptest coverage - Quie… (2026-06-21)
- [x] **Phase 86: Search + Tracker + TMS + Notification Uplift** - Поднять пакеты ниже 65%: `search`, `tracker`, `tms`, `notification`. (2026-06-21)
- [x] **Phase 87: TMS + Notification + NATS Uplift** - Поднять покрытие `tms`, `notification`, `platform/nats` — закрыть пробелы в consumer events, ent batch repo и NATS pu… (2026-06-21)
- [x] **Phase 88: TMS + Notification ≥65%** - Довести `tms` и `notification` до **≥65%** statement coverage. (2026-06-21)
- [x] **Phase 89: E2E Wave 3 + Dashboard Uplift** - Закрыть **E2E-01** (5 critical Playwright paths) и поднять `dashboard/` coverage. (2026-06-21)
- [x] **Phase 90: IMEX Jira Import Uplift** - Поднять `imex/` — ImportJira JSON/handler flows и edge cases. (2026-06-21)
- [x] **Phase 91: Intranet + Observability Uplift** - Поднять `intranet/` и `platform/observability/` до ≥80% — coffee/news handler gaps и metrics allowlist. (2026-06-21)
- [x] **Phase 92: Messenger Polls + Search Repo** - Закрыть 0% gaps в messenger poll repo + handler DeleteChannel/AddMembers; GetSavedFilter в search. (2026-06-21)
- [x] **Phase 93: Auto-Channels NATS + Handler Gaps** - Поднять `messenger/` до ≥70% — auto_channels NATS lifecycle, thread replies, revoke invite, poll handlers; search ind… (2026-06-21)
- [x] **Phase 94: Search Meili + Bot + Notification JetStream** - Поднять `search/` до ≥70%, `service_bot` coverage, notification consumer Start с JetStream. (2026-06-21)
- [x] **Phase 95: Notification Push/TG + Tracker Automation + TMS Export** - Поднять `notification/`, `tracker/`, `tms/` до ≥70% через push/telegram bind, automation consumer, issue message link… (2026-06-21)
- [x] **Phase 96: TMS TestRail/Mongo + Excel Import** - Закрыть deferred Phase 95: `mongo_repository`, `testrail_import`, `ImportFromExcel`; подготовить messenger mongo inte… (2026-06-21)
- [x] **Phase 97: Mongo CI + Webhook Dispatcher + Marketplace** - Mongo integration tests в CI, закрыть 0% webhook dispatcher, marketplace handlers, telegram sendMessage. (2026-06-21)
- [x] **Phase 98: Accounting + Calendar + COV 75%** - Закрыть gaps **accounting** (65%) и **calendar** (65.6%); довести overall `internal/` COV до **≥75%** (deferred из Ph… (2026-06-21)
- [x] **Phase 99: Backup + Package 75% Push** - Довести **backup** (65.7%) и пакеты на грани 75% (**notification**, **tms**, **marketplace**) до целевого порога; укр… (2026-06-21)
- [x] **Phase 100: RestoreBackup + Tracker/Messenger/Search 75%** - Закрыть **RestoreBackup** S3 happy path; довести **tracker**, **messenger**, **search** до ≥75%; удержать overall COV… (2026-06-21)
- [x] **Phase 101: Sub-75% Package Push** - Довести пакеты ниже 75% (**project**, **github**, **crm**, **hr**, **platform/outbox**, **platform/errors**, **integr… (2026-06-21)
- [x] **Phase 102: Dev Routes + Notification Push** - Close deferred gaps: `internal/dev` (0%) and `notification` (74.8% → ≥75%)
- [x] **Phase 103: GitHub Push + Mongo Integration Tests** - Push boundary packages (github 75%→80%+) and add mongo integration tests (skip without MongoDB)
- [x] **Phase 104: Flaky GitHub Fix + Mongo Testutil** - Eliminate network-dependent github handler tests; unify MongoDB test setup for CI
- [x] **Phase 105: Boundary Package Push** - Поднять пакеты на грани 75% (**marketplace**, **project**, **tms**) и удержать overall COV ≥78.4%.
- [x] **Phase 106: Triple Boundary Push (search / messenger / tracker)** - Поднять три пакета на грани 75% и overall COV ≥78.7%.
- [x] **Phase 107: CRM + Dashboard + Bases Boundary Push** - Push `crm`, `dashboard`, `bases` toward 79%+ and lift overall COV past 78.8%.
- [x] **Phase 108: 79% COV Milestone** - Cross **79%** overall statement coverage via accounting + search + tms boundary tests.
- [x] **Phase 109: Triple Boundary Push (Search + Messenger + Tracker)** - Push lowest packages toward 76%+ and maintain COV ≥79%.
- [x] **Phase 110: TMS + AI + Project Boundary Push** - Push `tms`, `ai`, `project` toward 78%+ and lift COV toward 79.5%.
- [x] **Phase 111: Notification + HR + CRM Push** - Triple push `notification (77.8%) + hr (77.9%) + crm (78%)` → COV **≥79.5%**.
- [x] **Phase 112: Search + Messenger + Tracker Push** - Triple push `search (75.8%) + messenger (75.8%) + tracker (75.9%)` → COV **≥80%**.
- [x] **Phase 113: TMS + Project + Sprint + Messenger + Tracker Push** - Cross-module push to COV **≥80%** (from 79.7%).
- [x] **Phase 114: Search + Messenger + TMS + Tracker Push** - Push COV **≥81%** (from 80.0%).
- [x] **Phase 115: GitHub/GitLab + Auth + Messenger + Bases + AI Push** - Push COV **≥81%** (from 80.2%).
- [x] **Phase 116: Search PG Integration + Messenger/TMS Push** - COV **≥81%** (from 80.7%).
- [x] **Phase 117: Storage RBAC + Notification Quiet Hours + Tracker Time** - Push low packages toward 82%; maintain COV **≥81%** (from 81.6%).
- [x] **Phase 118: Notification Consumer + Tracker Automation Push** - Push `notification`/`tracker` to **≥80%**; maintain COV **≥81%** (from 81.8%).
- [x] **Phase 119: Notification/Tracker Push to 82%** - Push `notification`/`tracker` toward **82%**; maintain COV **≥81%** (from 81.9%).
- [x] **Phase 120: Tracker/TMS/Project Push to 82%** - Push `tracker`/`tms`/`project` toward **82%**; maintain COV **≥82%** (from 82.0%).
- [x] **Phase 121: Notification Consumer + Tracker Push to 82%** - notification ≥82%, tracker ≥82%, COV ≥82%
- [x] **Phase 122: TMS + Project ≥82%** - tms: 79.3% → ≥82%; project: 80.4% → ≥82%
- [x] **Phase 123: Messenger ≥82%, Storage ≥82%** - messenger: 77.6% → ≥82%; storage: 86.8% → ✅ already ≥82% (2026-06-21)
- [x] **Phase 124: Floor ≥82% (tracker / tms / project)** - tracker: 81.3% → ≥82%; tms: 81.7% → ≥82%; project: 81.0% → ≥82%
- [x] **Phase 125: COV ≥82%, Security ≥80%, Regression Gate** - COV overall: ≥82% → ✅ **82.6%**; auth: ≥80% → ✅ **82.4%**; ws: ≥80% → ✅ **87.7%**; rbac: ≥80% → ✅ **84.7%** (2026-06-21)

### August 2026 — test architecture rebase (TASK-193…234)

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](active/august-2026/06-TEST-ARCHITECTURE-REBASE.md) ·
карточки: [`backlog/README.md`](backlog/README.md). Старый недельный test-раздел
`05-AUGUST-DETAILED.md` superseded этим каноном; остальные августовские темы
сохраняются. Микросервисный split остаётся отложенным до re-audit после тестового
фундамента.

**Фактическая стадия (founder, 2026-08-12):** незавершённый MVP; ни один
продуктовый модуль не принят как complete, stable или production-ready.
Исторические milestone COMPLETE не равны product/production acceptance.
PG-03/FA-06 — отдельный manual ops-блокер production cutover, а не proxy
готовности MVP. Новая native/Python migration начата только с Phase 128 inventory.
Historical ≥82% сохранён как discovery evidence, но founder-approved greenfield
contract отменяет его как migration gate; destructive batches разрешены только
по current-SHA manifest.

| Phase | Status | Depends on | Outcome / tasks |
|------:|--------|------------|-----------------|
| 126 | IMPLEMENTED 2026-07-28 | — | Intranet NewsPost trash · TASK-001 |
| 127 | IMPLEMENTED 2026-07-28 | 126 | Calendar series trash · TASK-002 |
| 128 | IN PROGRESS · contract/manifests | — | TASK-202 archived/verified; TASK-203/224 baseline evidence closed; TASK-204 accepted contract + ordered `cutover-manifest.json` verification pending |
| 129 | PLANNED · 7 ordered plans | 128 PASS | exact cutover: per-batch operation + CI/docs switch + manifest evidence; move 350 hook-free Go units, preserve/refactor 245 via TASK-234, move frontend units, delete legacy non-units, root Python scaffold · TASK-193/197/198/201/219/234 |
| 130 | PLANNED | 129 | task/requirement/risk → TC traceability and fail-closed quality gates · TASK-206/218 |
| 131 | PLANNED | 130 | Python core + health/auth/account API foundation · TASK-205/207 |
| 132 | PLANNED | 131 | authz, tenant isolation, admin/platform and abuse cases · TASK-208/209/227 |
| 133 | PLANNED | 131 | PostgreSQL/Mongo/Redis/NATS/jobs integration · TASK-210/211 |
| 134 | PLANNED | **133** | work + realtime domains from current requirements · TASK-212/213/228 |
| 135 | PLANNED | **133** | business + external integrations from current requirements · TASK-214/215/229 |
| 136 | PLANNED | 131 | API/schema/frontend static contracts in Python · TASK-216/230 |
| 137 | PLANNED | 131 | Python Playwright foundation + risk-based critical smoke · TASK-194/217 |
| 138 | PLANNED | 137, 134, 135 | new E2E domain and recovery journeys · TASK-195/196 |
| 139 | PLANNED | 132, 138 | accessibility + cross-cutting security/privacy · TASK-231/232 |
| 140 | PLANNED | 131, 134, 135 | load/performance/resilience baselines · TASK-200 |
| 141 | PLANNED | 132–140 | CI quality gates + clean-room verification + docs/handoff · TASK-220/233 |

```text
128 -> 129 -> 130 -> 131
                    |-> 132 security/platform -----------------------> 139 security/a11y -> 141
                    |-> 133 infra -> 134 work/realtime --+          /
                    |              -> 135 business/ext --+-> 138 journeys
                    |-> 136 contracts                                  -> 141
                    |-> 137 E2E ---------------------------> 138 journeys
                    |-> 140 load/resilience ----------------------------> 141
```

- [x] **Phase 126:** Intranet NewsPost trash
- [x] **Phase 127:** Calendar series trash
- [ ] **Phase 128:** accepted founder contract + regenerated current exact
  keep/move/delete manifest; legacy baseline хранится только как discovery.
- [ ] **Phase 129:** ранний exact-manifest cutover: genuine native units →
  `backend/tests/**` / `frontend/tests/**`; legacy non-units/runners удаляются;
  единый Python-проект создаётся прямо в `testframework/**`. 245 hook-dependent
  genuine Go units не удаляются: сохраняются и переносятся через TASK-234.
- [ ] **Phases 130–140:** greenfield Python suites из current requirements,
  tasks, verified TC и risks; никакого port/parity/dual-run.
- [ ] **Phase 141:** CI quality gates, clean-room verification и docs/handoff.

### August 2026 — Hub Mail + Calendar Sync (TASK-235…241)

Канон: [`ADR-MAIL-01`](decisions/ADR-MAIL-01-mail-engine-calendar-sync.md) ·
контракт: [`MAIL-SPEC.md`](decisions/specs/MAIL-SPEC.md) · QA handoff:
[`docs/mail/QA-HANDOFF.md`](../docs/mail/QA-HANDOFF.md).

Эти фазы принадлежат продуктовой mail-ветке и не меняют защищённую тестовую программу
128–141. Разработка идёт параллельно; финальный rebase выполняется только после того,
как владелец сообщит о merge тестовой ветки. Runtime mail stack остаётся `NOT_RUN`, пока
не принят единый Docker-only launcher и не закрыты license/provenance gates.

| Phase | Status | Depends on | Outcome / task |
|------:|--------|------------|----------------|
| 142 | CONTRACT COMPLETE · static PASS · runtime NOT_RUN | — | ADR, ERD, OpenAPI, NATS, UI/security/QA contracts · TASK-235 |
| 143 | DISPATCH READY · CODE BLOCKED BY EXACT SPARK RUNTIME | 142 | CalendarEvent/Attendee/Exception, deterministic iTIP and optimistic RSVP · TASK-236 |
| 144 | DISPATCH READY · CODE BLOCKED BY EXACT SPARK RUNTIME | 142 | Mailbox domain, engine port/fake, folders/threads/drafts/submission · TASK-237 |
| 145 | BLOCKED BY EXTERNAL GATES | 144 + license + launcher | Stalwart adapter, instance-domain mail, bounded Docker profile · TASK-238 |
| 146 | DISPATCH READY · CODE/HARNESS BLOCKED | 143, 144 | Outlook three-pane UI, composer, attachments, search/realtime · TASK-239 |
| 147 | CONTRACT 0.1.4 · ROOT STATIC PASS `46cfa26c` · INDEPENDENT REVIEW PENDING · CODE BLOCKED | 143–146 | recurrence/import recovery, busy/free, reminders, L1/CRM seam · TASK-240 |
| 148 | GATE CONTRACT PREPARED · EXECUTION BLOCKED | 143–147 + test-rebase merge notice | security/reconciliation/docs/verification/rebase gate · TASK-241 |

```text
142 contracts -> 143 calendar core ----+-> 146 UI -> 147 reliability -> 148 gate
              -> 144 mailbox domain -> 145 engine/runtime -----------+
```

- [x] **Phase 142:** freeze product, architecture, API, event, security and QA contracts.
- [ ] **Phase 143:** implement canonical calendar RSVP/iTIP core.
- [ ] **Phase 144:** implement engine-neutral mailbox domain and fake adapter.
- [ ] **Phase 145:** pass external gates and implement/measure the mail runtime.
- [ ] **Phase 146:** implement the responsive Outlook-class Hub mail workspace.
- [ ] **Phase 147:** close recurrence/import/busy-free/reminder and L1/CRM behavior.
- [ ] **Phase 148:** close security and verification, then rebase after founder signal.

## Phase Details

<details>
<summary>v1.0 MVP Phase Details (Phases 1-7)</summary>

### Phase 1: Foundation & Infrastructure
**Goal**: Пользователи могут аутентифицироваться, создавать проекты с командами и ролями, а платформа имеет работающую инфраструктуру реального времени, файловое хранилище и CI-пайплайн — субстрат, от которого зависит каждый модуль
**Depends on**: Nothing (first phase)
**Requirements**: AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-05, AUTH-06, AUTH-07, AUTH-08, PROJ-01, PROJ-02, PROJ-03, PROJ-04, PROJ-06, PROJ-07, UI-01, UI-07, UI-08
**Success Criteria** (что должно быть ИСТИННО):
  1. Пользователь может войти через Magic Link по email и через Telegram, попасть на рабочую оболочку дашборда и выйти
  2. Пользователь может создать проект и увидеть автоматически сгенерированные доску трекера, пространство TMS, чат проекта и заглушки каналов
  3. Владелец проекта может приглашать пользователей по ссылке, назначать роли (Owner/Developer/Tester/Viewer), и ограничения по ролям применяются
  4. Файлы можно загружать перетаскиванием или вставкой, они сохраняются в SeaweedFS с рабочим предпросмотром
  5. WebSocket-соединение доставляет события реального времени из NATS в браузер (проверено тестовым событием)
**Plans**: 9 plans

Plans:
- [x] 01-01: Среда разработки, структура проекта, CI-пайплайн, Docker Compose со всеми сервисами
- [x] 01-02: Схемы базы данных (Ent + MongoDB), настройка NATS, Redis, интеграция с SeaweedFS
- [x] 01-03: Модуль аутентификации (Magic Link + вход через Telegram, сессии, RBAC-мидлвара)
- [x] 01-04: Модуль проектов (CRUD, членство, роли, команды, архив/корзина)
- [x] 01-05: WebSocket gateway (NATS bridge) + frontend API client, stores, hooks, permissions
- [x] 01-07: React UI оболочка (маршрутизация, sidebar, topbar, login, проекты, файлы)
- [x] 01-06: Документация фазы 1 (бизнес-дока, тестовая дока, техническая дока)
- [x] 01-08: [GAP] Подключение модуля хранилища в main.go, исправление auth middleware в storage handler
- [x] 01-09: [GAP] Реализация CRUD команд (teams) вместо заглушек 501

### Phase 2: Task Tracker
**Goal**: Пользователи могут управлять работой над проектом через полнофункциональный трекер задач с Kanban-досками, настраиваемыми рабочими процессами, расширенными комментариями и учётом времени
**Depends on**: Phase 1
**Requirements**: TASK-01, TASK-02, TASK-03, TASK-04, TASK-05, TASK-06, TASK-07, TASK-08, TASK-09, TASK-10, TASK-11, KANB-01, KANB-04, KANB-05, UI-09
**Success Criteria** (что должно быть ИСТИННО):
  1. Пользователь может создавать задачи всех 9 типов с описаниями в формате rich text, вложениями файлов и ID с префиксом проекта (HUB-123)
  2. Пользователь может перетаскивать задачи на Kanban-доске между колонками, соответствующими статусам рабочего процесса, с обновлениями в реальном времени, видимыми другим пользователям
  3. Владелец проекта может настраивать рабочий процесс, добавляя/удаляя статусы и определяя допустимые переходы
  4. Пользователь может создавать иерархии родитель-потомок и межзадачные связи (блокирует, связан с, дублирует и т.д.)
  5. Пользователь может комментировать с @упоминаниями и реакциями, просматривать полную историю изменений и вести учёт времени с таймером старт/стоп
**Plans**: 5 plans

Plans:
- [x] 02-01-PLAN.md — Ent-схемы трекера (15 схем), Repository interface, DTOs, events
- [x] 02-02-PLAN.md — Issue CRUD backend (handler/service/repository), workflow engine, тесты
- [x] 02-03-PLAN.md — Frontend API client, Zustand stores, Plate editor wrapper
- [x] 02-04-PLAN.md — Kanban-доска (drag-and-drop), списки задач, детальный вид, комментарии, настройки
- [x] 02-05-PLAN.md — Документация фазы 2 (бизнес, тестовая, техническая) на русском

### Phase 3: Messenger
**Goal**: Пользователи могут общаться в реальном времени через личные сообщения, каналы и треды с форматированным текстом, реакциями, голосовыми сообщениями и индикаторами присутствия
**Depends on**: Phase 1, Phase 2
**Requirements**: MSG-01, MSG-02, MSG-03, MSG-04, MSG-05, MSG-06, MSG-07, MSG-08, MSG-09, MSG-10, MSG-12, MSG-13
**Success Criteria** (что должно быть ИСТИННО):
  1. Пользователь может отправлять и получать сообщения в реальном времени через личные сообщения, групповые чаты, публичные и приватные каналы
  2. Пользователь может отвечать в тредах, ставить реакции эмодзи, форматировать текст и записывать голосовые сообщения
  3. Каналы проекта создаются автоматически при создании проекта; каналы только для чтения работают для сообщений ботов
  4. Пользователь может редактировать/удалять свои сообщения, закреплять, добавлять в закладки и искать
  5. Статус онлайн/офлайн/отошёл виден в реальном времени и счётчики непрочитанных корректно обновляются
**Plans**: 5 plans

Plans:
- [x] 03-01-PLAN.md — Бэкенд ядра мессенджера
- [x] 03-02-PLAN.md — Управление каналами
- [x] 03-03-PLAN.md — Фронтенд чата
- [x] 03-04-PLAN.md — Присутствие, непрочитанные, поиск, пины, закладки, пересылка
- [x] 03-05-PLAN.md — Документация фазы 3

### Phase 4: TMS & Allure
**Goal**: Пользователи могут управлять тест-кейсами, выполнять тестовые прогоны, загружать результаты Allure и видеть связи между тестами и задачами
**Depends on**: Phase 1, Phase 2
**Requirements**: TMS-01, TMS-02, TMS-03, TMS-04, TMS-05, TMS-06, TMS-08, TMS-09, TMS-10, ALLU-01, ALLU-02, ALLU-03, IMEX-02
**Success Criteria** (что должно быть ИСТИННО):
  1. Пользователь может создавать тест-кейсы в классическом и BDD форматах с иерархией папок и версионированием
  2. Пользователь может создавать и выполнять ручные тестовые прогоны с результатами по каждому шагу
  3. JSON-результаты Allure загружаются через API и сопоставляются с тест-кейсами
  4. Тест-кейсы двусторонне связаны с задачами
  5. Отчёты показывают процент прохождения, тренды, обнаружение нестабильных тестов
**Plans**: 5 plans

Plans:
- [x] 04-01-PLAN.md — Ent-схемы TMS, handler/service/repository
- [x] 04-02-PLAN.md — Тестовые прогоны + фронтенд TMS UI
- [x] 04-03-PLAN.md — Интеграция с Allure
- [x] 04-04-PLAN.md — Кросс-модульные связи, отчётность, импорт/экспорт
- [x] 04-05-PLAN.md — Документация фазы 4

### Phase 5: Voice Calls
**Goal**: Пользователи могут присоединяться к аудиозвонкам с демонстрацией экрана
**Depends on**: Phase 1, Phase 3
**Requirements**: CALL-01, CALL-03, CALL-06
**Success Criteria** (что должно быть ИСТИННО):
  1. Пользователь может начать или присоединиться к аудиозвонку с участием до 50 человек
  2. Пользователь может включать/выключать микрофон, поднимать руку и демонстрировать экран
  3. Эфемерный чат звонка доступен во время звонка и исчезает после завершения
  4. События звонка отображаются в канале мессенджера
**Plans**: 4 plans

Plans:
- [x] 05-01-PLAN.md — Интеграция с LiveKit
- [x] 05-02-PLAN.md — UI звонков
- [x] 05-03-PLAN.md — Жизненный цикл звонка
- [x] 05-04-PLAN.md — Документация фазы 5

### Phase 6: Notifications, Search & Dashboards
**Goal**: Пользователи имеют уведомления, глобальный поиск и дашборды
**Depends on**: Phase 1-5
**Requirements**: NOTF-01, NOTF-02, NOTF-03, NOTF-04, SRCH-01, SRCH-02, DASH-01, DASH-02, DASH-03, AUDT-01, AUDT-02, AUDT-03, AUDT-04, UI-05
**Success Criteria** (что должно быть ИСТИННО):
  1. Пользователь получает уведомления в приложении о назначениях, @упоминаниях и сменах статуса
  2. Ctrl+K открывает глобальный поиск по всей платформе
  3. Персональный дашборд показывает задачи, непрочитанные, звонки, прогоны и дедлайны
  4. Дашборд проекта показывает метрики; административный дашборд показывает метрики системы
**Plans**: 5 plans

Plans:
- [x] 06-01-PLAN.md — Система уведомлений
- [x] 06-02-PLAN.md — Глобальный поиск
- [x] 06-03-PLAN.md — Дашборды
- [x] 06-04-PLAN.md — Журнал аудита, лимиты, корзина, GDPR
- [x] 06-05-PLAN.md — Документация фазы 6

### Phase 7: Polish & Extended Features
**Goal**: Спринты, интеграция с GitHub, запись звонков, i18n, темы, PWA, API, бэкапы
**Depends on**: Phase 1-6
**Requirements**: PROJ-05, KANB-02, KANB-03, SCRM-01, SCRM-02, SCRM-03, SCRM-04, SCRM-05, MSG-11, TMS-07, CALL-02, CALL-04, CALL-05, GHUB-01, GHUB-02, GHUB-03, GHUB-04, GHUB-05, SRCH-03, API-01, API-02, API-03, API-04, IMEX-01, IMEX-03, BKUP-01, BKUP-02, BKUP-03, UI-02, UI-03, UI-04, UI-06
**Success Criteria** (что должно быть ИСТИННО):
  1. Scrum-команды могут создавать спринты с Planning Poker и графиками burndown/velocity
  2. PR/коммиты GitHub автоматически связываются с задачами по ID
  3. Пользователи могут записывать звонки
  4. Приложение работает на русском и английском с переключением тем и PWA
  5. REST API документирован, API-ключи, вебхуки, автоматические бэкапы
**Plans**: 6 plans

Plans:
- [x] 07-01-PLAN.md — Управление спринтами
- [x] 07-02-PLAN.md — Интеграция с GitHub
- [x] 07-03-PLAN.md — Запись звонков, запланированные/повторяющиеся звонки
- [x] 07-04-PLAN.md — Интернационализация, темизация, мобильная раскладка, PWA, превью ссылок
- [x] 07-05-PLAN.md — REST API, API-ключи, вебхуки, hub-cli, импорт/экспорт, шаблоны, бэкапы, WIP, swimlanes, окружения, JQL
- [x] 07-06-PLAN.md — Документация фазы 7

</details>

### Phase 8: Design System Foundation
**Goal**: Приложение переходит на новую дизайн-систему — все поверхности, текст, границы, тени, радиусы, анимации и шрифты соответствуют новой палитре, а тёмная и светлая темы работают корректно
**Depends on**: Phase 7 (v1.0 complete)
**Requirements**: DSFN-01, DSFN-02, DSFN-03, DSFN-04, DSFN-05, DSFN-06, DSFN-07, DSFN-08, DSFN-09, THME-01, THME-02, THME-03, THME-04
**Success Criteria** (what must be TRUE):
  1. User sees Inter font for all UI text and JetBrains Mono for code elements across the entire application
  2. User sees the new 5-level surface palette (void/base/raised/card/input), 4-level text hierarchy, and 3-level border system in dark mode matching mockup values
  3. User can switch between dark/light/system themes without visual glitches, and light theme passes WCAG AA contrast ratios
  4. User can select accent color and see it applied consistently across all themed surfaces
  5. No !important declarations remain in app.css, and spring/bounce easing is applied on hover/focus/enter animations
**Plans**: 2 plans

Plans:
- [x] 08-01-PLAN.md — Font swap + CSS token foundation (surfaces, text, borders, radii, shadows, easing, accents, dual Shadcn mapping)
- [x] 08-02-PLAN.md — !important removal (56 declarations) + noise texture overlay

**UI hint**: yes

### Phase 9: Layout Shell
**Goal**: Пользователь видит обновлённый сайдбар с градиентной линией и свечением активного пункта, топбар с backdrop blur, и вся оболочка работает на всех размерах экрана
**Depends on**: Phase 8
**Requirements**: LYOT-01, LYOT-02, LYOT-03, LYOT-04, LYOT-05
**Success Criteria** (what must be TRUE):
  1. User sees redesigned sidebar with gradient edge line, new background, and active navigation item with accent glow indicator
  2. User sees redesigned topbar with backdrop blur effect and gradient bottom edge
  3. Sidebar collapse/expand works without visual regression at all viewport sizes
  4. Mobile navigation (bottom nav, overlay sidebar) works without regression
**Plans**: 2 plans

Plans:
- [x] 09-01-PLAN.md — Sidebar redesign (background, gradient edge, active glow, header, user/projects tokens)
- [x] 09-02-PLAN.md — Topbar redesign (backdrop blur, gradient edge), AppShell topbar visibility, mobile nav tokens

**UI hint**: yes

### Phase 10: UI Primitives
**Goal**: Все 10 базовых Shadcn-компонентов рестайлированы под новый дизайн-язык, и любая страница, использующая эти компоненты, автоматически выглядит обновлённой
**Depends on**: Phase 8, Phase 9
**Requirements**: PRIM-01, PRIM-02, PRIM-03, PRIM-04, PRIM-05, PRIM-06, PRIM-07, PRIM-08, PRIM-09, PRIM-10
**Success Criteria** (what must be TRUE):
  1. User sees redesigned Button (gradient primary, ghost/secondary variants, glow hover) and Card (gradient bg, top-edge highlight, hover glow) across the app
  2. User sees redesigned Input (blue focus glow ring), Badge (status-specific glow), and Avatar (online indicator glow)
  3. User sees redesigned Dialog/Popover (entrance animation, backdrop blur), DropdownMenu, Tabs (active indicator), Tooltip, and ScrollArea (slim 5px scrollbar)
  4. All existing page functionality that uses these components continues to work without regression
**Plans**: 3 plans

Plans:
- [x] 10-01-PLAN.md — Button, Card, Input restyle (gradient primary, card gradient+highlight, input focus glow)
- [x] 10-02-PLAN.md — Badge, ScrollArea, Avatar, Tooltip restyle (status glows, 5px scrollbar, raised surfaces)
- [x] 10-03-PLAN.md — Dialog, Tabs, DropdownMenu restyle (backdrop blur, active indicator, raised menus)
**UI hint**: yes

### Phase 11: Core Mockup Pages
**Goal**: Четыре самые важные и сложные страницы (Login, Dashboard, Messenger, Kanban) pixel-perfect соответствуют HTML-макетам, а вся существующая функциональность работает без регрессий
**Depends on**: Phase 10
**Requirements**: LOGN-01, LOGN-02, LOGN-03, LOGN-04, LOGN-05, DASH-01, DASH-02, DASH-03, DASH-04, DASH-05, MSGR-01, MSGR-02, MSGR-03, MSGR-04, MSGR-05, MSGR-06, KANB-01, KANB-02, KANB-03, KANB-04
**Success Criteria** (what must be TRUE):
  1. User sees Login page with split layout (animated gradient left panel with floating feature cards, form right panel), mesh gradient background, and redesigned email/code flow -- login via magic link and Telegram works without regression
  2. User sees Dashboard with 3-column widget grid, greeting section with gradient text, quick action buttons, and staggered entrance animations -- widget drag-and-drop works without regression
  3. User sees Messenger with redesigned 3-panel layout, channel list with search/filter, message items with new styling, glassmorphic composer, and redesigned reactions/typing/reply -- all messenger functionality (send, react, thread, pin, forward, voice, search) works without regression
  4. User sees Kanban with redesigned columns, issue cards with priority dots and glow hover, filter/view options bar -- drag-and-drop works without regression
**Plans**: 4 plans

Plans:
- [x] 11-01-PLAN.md — Login page restyle (split layout, mesh gradient, floating cards, form)
- [x] 11-02-PLAN.md — Dashboard page restyle (greeting gradient, quick actions, widget grid)
- [x] 11-03-PLAN.md — Messenger page restyle (channel list, messages, glassmorphic composer)
- [x] 11-04-PLAN.md — Kanban page restyle (columns, issue cards, filter bar)
**UI hint**: yes

### Phase 12: Remaining Mockup Pages + Unmocked Pages
**Goal**: Все оставшиеся 7 страниц с макетами pixel-perfect соответствуют HTML-макетам, а 6 страниц без макетов приведены к единому дизайн-языку -- вся функциональность работает без регрессий
**Depends on**: Phase 11
**Requirements**: ISSU-01, ISSU-02, ISSU-03, ISSU-04, CALL-01, CALL-02, CALL-03, CALL-04, TMSP-01, TMSP-02, TMSP-03, TMSP-04, STNG-01, STNG-02, STNG-03, STNG-04, STNG-05, ADMN-01, ADMN-02, ADMN-03, ADMN-04, ADMN-05, ONBD-01, ONBD-02, ONBD-03, ONBD-04, ONBD-05, TRSH-01, TRSH-02, TRSH-03, XTRA-01, XTRA-02, XTRA-03, XTRA-04, XTRA-05, XTRA-06, XTRA-07
**Success Criteria** (what must be TRUE):
  1. User sees redesigned Issue Detail (two-column layout, comments/activity, right sidebar meta) -- all issue functionality (edit, comment, time track, checklist, attachments) works without regression
  2. User sees redesigned Calls (participant grid, controls, status indicators), TMS (test management interface, case grid, execution tracking) -- all call and TMS functionality works without regression
  3. User sees redesigned Settings (sidebar nav + content, profile, appearance, notifications), Admin (KPI cards, tabbed interface, user table, audit log), Onboarding (centered modal, profile form, feature tour, theme selection), and Trash (table, restore/delete buttons) -- all functionality works without regression
  4. User sees Projects list, Project overview, Sprint management, JQL search, Project settings, and Project dashboard pages styled consistently with the design system -- all existing functionality on these pages works without regression
**Plans**: 5 plans

Plans:
- [x] 12-01-PLAN.md — Issue Detail + Calls restyle (drift fixed 2026-07-03: подтверждено по STATE/AGENTS)
- [x] 12-02-PLAN.md — TMS pages restyle (drift fixed 2026-07-03: подтверждено по STATE/AGENTS)
- [x] 12-03-PLAN.md — Settings + Admin restyle (drift fixed 2026-07-03: подтверждено по STATE/AGENTS)
- [x] 12-04-PLAN.md — Onboarding + Trash restyle
- [x] 12-05-PLAN.md — Unmocked pages (Projects, Sprints, JQL, Project settings, Project dashboard) (drift fixed 2026-07-03: подтверждено по STATE/AGENTS)
**UI hint**: yes

### Phase 13: Polish & QA
**Goal**: Финальная визуальная проверка качества, доступность, мобильная адаптация и очистка кодовой базы -- приложение визуально безупречно в обеих темах на всех размерах экрана
**Depends on**: Phase 12
**Requirements**: (cross-cutting quality -- validates all previous phases)
**Success Criteria** (what must be TRUE):
  1. All animations run at 60fps on 4x CPU throttle (Dashboard and Kanban benchmark pages)
  2. Lighthouse accessibility audit shows zero contrast violations in both dark and light themes across all pages
  3. All pages render correctly at 375px, 768px, and 1024px viewports without layout breaks or overflow
  4. Dead CSS classes removed from app.css; no unused custom utilities remain
  5. Theme toggle transition is smooth (200ms background/color transition on key surfaces)
**Plans**: 2 plans

Plans:
- [x] 13-01-PLAN.md — Avatar online glow + smooth theme toggle (1/1 SUMMARY)
- [x] 13-02-PLAN.md — Dead CSS cleanup + final QA
**UI hint**: yes

### Phase 14: Architectural Hardening
**Goal**: Расщепить god-объекты бэкенда (AuthService 580 строк, CallService) согласно SRP, убрать ent.Client из сервисов в репозитории (LO-03), расширить Permission за пределы uint64, покрыть тестами критичные пути, выявленные в аудите REVIEW.md (webhook SSRF/dispatcher, RBAC route integration, magic-link brute force, outbox panic recovery, LiveKit webhook signature).
**Depends on**: Phase 13
**Requirements**: cross-cutting code-quality. Не вводит новые фичи; цель — снизить технический долг и повысить покрытие тестами.
**Success Criteria** (what must be TRUE):
  1. AuthService разбит минимум на 3 субсервиса (Session/Login/Profile), каждый <= 250 строк, с собственными интерфейсами
  2. CallService разделён на CallRoomService (LiveKit RPC) + CallRepository wrapper + CallEventPublisher (NATS); call.service.go <= 250 строк
  3. MessengerService, AuditService, SearchService не импортируют `*ent.Client` напрямую — все запросы через Repository интерфейсы
  4. Permission представление поддерживает > 64 разрешений (миграция на `[]string` или `bigint[]`)
  5. Покрытие тестами: webhook dispatcher (SSRF + retry + signing), RBAC route integration (anonymous/non-member/member на ключевых маршрутах), magic-link verify brute-force lockout, outbox worker panic recovery
  6. `go test -count=1 ./...` зелёный, `go test -race ./...` без data race, общий test count вырос минимум на 15 тест-функций
**Plans**: TBD (планируется через /gsd-plan-phase)
**Tech hint**: no UI changes; backend refactor only


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
  3. notification coverage unchanged (+0.1%)
  4. race PASS all packages
  5. `/qa-gate` PASS (backend)
**Plans**: 5 plan(s) · source `.planning/phases/87-tms-notif-nats/PLAN.md`

Plans:
- [x] nats ≥65%
- [x] tms ≥62%
- [x] notification coverage unchanged (+0.1%)
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

### Phase 142: Mail Contract Foundation
**Goal**: Заморозить source-of-truth, ERD, OpenAPI, NATS, security, UI и QA-контракты
до реализации почтового модуля.
**Depends on**: founder mail brief; does not depend on Phases 128–141 execution
**Requirements**: TASK-235, FR-MAIL-001…016, FR-CALSYNC-001…013 (contract only)
**Success Criteria** (what must be TRUE):
  1. ADR-MAIL-01 records candidate comparison and explicit operational gates.
  2. MAIL-SPEC freezes entities, invariants, state machines and traceability.
  3. OpenAPI and NATS JSON Schema cover all promised public/admin surfaces.
  4. QA handoff and founder question queue are complete without runtime overclaims.
**Plans**: source `.planning/phases/142-mail-contract-foundation/PLAN.md`

### Phase 143: Calendar RSVP Core
**Goal**: First-class events and attendees provide deterministic RFC 5545/5546
invite/update/cancel/reply behavior with optimistic RSVP.
**Depends on**: Phase 142
**Requirements**: TASK-236, FR-CALSYNC-001…010
**Success Criteria** (what must be TRUE):
  1. Ent model and call backfill policy match the frozen ERD.
  2. REQUEST/CANCEL/REPLY and recurrence instance behavior have genuine units.
  3. RSVP writes are owner/sender-bound, versioned, idempotent and outbox-backed.
  4. Calendar API/aggregator exposes canonical events without regressing calls.
**Plans**: source `.planning/phases/143-calendar-rsvp-core/PLAN.md`

### Phase 144: Mailbox Domain
**Goal**: Engine-neutral Hub mailbox domain supports owned folders/messages/threads,
drafts, submissions, labels, attachments, rules, contacts and projections.
**Depends on**: Phase 142
**Requirements**: TASK-237, FR-MAIL-001…016
**Success Criteria** (what must be TRUE):
  1. MailEngine port and deterministic fake freeze engine isolation.
  2. Domain/services/handlers match OpenAPI and store no raw body in PostgreSQL.
  3. Submission and event consumers are idempotent; search/realtime projections repair.
  4. Genuine mail units cover pure policies and fake-adapter service behavior.
**Plans**: source `.planning/phases/144-mailbox-domain/PLAN.md`

### Phase 145: Stalwart Gateway and Docker Mail Profile
**Goal**: After explicit gates, connect the preferred engine and prove bounded persistent
instance-domain mail runtime.
**Depends on**: Phase 144; founder/legal license verdict; accepted Docker launcher
**Requirements**: TASK-238, FR-MAIL-001, FR-MAIL-009, FR-MAIL-015
**Success Criteria** (what must be TRUE):
  1. Immutable image provenance and license choice are recorded.
  2. JMAP adapter and provisioning pass accepted integration evidence.
  3. Optional `mail` profile is healthy, persistent and ≤4 CPU/≤4 GiB by measured run.
  4. External-domain claims remain blocked until DNS/PTR/network acceptance exists.
**Plans**: source `.planning/phases/145-stalwart-gateway/PLAN.md`

### Phase 146: Outlook-class Mail UI
**Goal**: Hub users can triage, read, compose, search and RSVP through an accessible,
responsive three-pane mail workspace.
**Depends on**: Phases 143, 144
**Requirements**: TASK-239, applicable FR-MAIL and FR-CALSYNC-006
**Success Criteria** (what must be TRUE):
  1. Desktop/tablet/mobile information architecture matches the approved design.
  2. Draft autosave/conflict/send ambiguity and attachment scan states are honest.
  3. Invitation card handles success, conflict, cancellation, stale and untrusted states.
  4. Feature units live only in `frontend/tests/mail/**` and pass in Docker.
**Plans**: source `.planning/phases/146-mail-outlook-ui/PLAN.md`

### Phase 147: Mail/Calendar Reliability and Integrations
**Goal**: Close recurrence/import/recovery, busy/free with leave, reminders, L1 feeds and
the CRM send/manual-log seam.
**Depends on**: Phases 143–146
**Requirements**: TASK-240, FR-CALSYNC-009…013, FR-CRM-MAIL-001…002,
CRM GZ-V4-08…10
**Success Criteria** (what must be TRUE):
  1. Out-of-order/redelivered recurrence and iTIP flows converge.
  2. Availability combines events, calls and approved leave without detail leakage.
  3. Reminder and CRM paths reuse existing policy boundaries.
  4. L1 feeds are owner-controlled and SSRF/bounds protected or remain explicitly deferred.
**Plans**: source `.planning/phases/147-mail-calendar-reliability/PLAN.md` · contract
`.planning/phases/147-mail-calendar-reliability/CONTRACT.md` · atomic dispatch index
`.planning/phases/147-mail-calendar-reliability/DISPATCH-INDEX.md` · root static review
`.planning/phases/147-mail-calendar-reliability/STATIC-REVIEW.md`

### Phase 148: Mail Merge-readiness Gate
**Goal**: Resolve independent review findings, prove security/recovery/docs, hand off QA,
and rebase after the test-program merge signal.
**Depends on**: Phases 143–147 and founder notice that Claude's branch is in master
**Requirements**: TASK-241, all implemented P0/P1 mail requirements
**Success Criteria** (what must be TRUE):
  1. Applicable Docker-only unit/integration/security evidence is current-SHA green.
  2. QA handoff and code/docs/contracts are synchronized.
  3. Branch is rebased after the protected test migration and conflicts are resolved.
  4. No push/PR/merge/deploy or readiness overclaim occurs without founder authority.
**Plans**: source `.planning/phases/148-mail-release-gate/PLAN.md` · gate matrix
`.planning/phases/148-mail-release-gate/GATE-MATRIX.md` · independent review packets
`.planning/phases/148-mail-release-gate/REVIEW-PACKETS.md`

## Progress

**Execution Order:**
Phases execute in numeric order (1 → 125). Decimal insertions (if any) slot between surrounding integers.

### Honest counters (synced 2026-07-28 · TASK-004/005)

Sources: [`PHASE-STATUS-MATRIX.md`](audits/planning-audit-2026-07/PHASE-STATUS-MATRIX.md) (live scan A3 + counting method in header), [`STATE.md`](STATE.md) (`goals_verified: 63`).

| Metric | Value | Source |
|--------|------:|--------|
| `total_phases` (project convention) | **125** | milestones 01–13 (**13**) + `.planning/phases/` dirs **14–125** (**112**) = 125. August phases **126–136** are additional (not yet in `total_phases` convention). |
| Phase dirs under `.planning/phases/` (live) | **124** | Matrix scan 2026-07-27 (includes TF 126–136 + `999-test-framework` within the 124 dirs). |
| `goals_verified` (`VERIFIED_PASS`) | **63** | Matrix **classification** column only (TASK-004). Matches `STATE.md`. |
| `VERIFIED_PARTIAL` | **5** | Matrix |
| `CODE_GATE_DEPLOY_BLOCKED` | **2** | Phases **63**, **67** — code gate PASS, deploy blocked |
| `NO_VERIFICATION` | **54** | Matrix |
| `VERIFICATION.md` present (dir root) | **70** | Matrix summary at scan; 126/127 VERIFICATION added 2026-07-28 not yet re-scanned into matrix |

**Deploy blockers** (do not read Progress 100% as “prod unlocked”):

| Phase | Verdict | Blocker |
|------:|---------|---------|
| 63 | PASS (code gate) | deploy blocked on **user commit/PR** + **PG-03** ops |
| 67 | PASS (code gate) | same — **PG-03 / FA-06** secret rotation still manual (`production/PG-03-SECRET-ROTATION.md`, `FA-06-EXECUTION-LOG.md`) |

> **STATE.md note (TASK-004):** `goals_verified: 63` — считать только колонку `classification` в матрице; метод в шапке `PHASE-STATUS-MATRIX.md`.

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Foundation & Infrastructure | v1.0 | 9/9 | Complete | 2026-03-21 |
| 2. Task Tracker | v1.0 | 5/5 | Complete | 2026-03-21 |
| 3. Messenger | v1.0 | 5/5 | Complete | 2026-03-21 |
| 4. TMS & Allure | v1.0 | 5/5 | Complete | 2026-03-21 |
| 5. Voice Calls | v1.0 | 4/4 | Complete | 2026-03-21 |
| 6. Notifications, Search & Dashboards | v1.0 | 5/5 | Complete | 2026-03-21 |
| 7. Polish & Extended Features | v1.0 | 6/6 | Complete | 2026-03-21 |
| 8. Design System Foundation | v2.0 | 2/2 | Complete | 2026-03-28 |
| 9. Layout Shell | v2.0 | 2/2 | Complete | 2026-03-28 |
| 10. UI Primitives | v2.0 | 3/3 | Complete | 2026-03-28 |
| 11. Core Mockup Pages | v2.0 | 4/4 | Complete | 2026-03-28 |
| 12. Remaining Mockup Pages + Unmocked Pages | v2.0 | 5/5 | Complete | 2026-03-28 |
| 13. Polish & QA | v2.0 | 2/2 | Complete | 2026-03-28 |
| 14. Architectural Hardening | v2.0 | 5/5 | Verified PASS | 2026-05-15 |
| 15. Trust (v2.5) | v2.5 | Complete | Verified PASS | 2026-06-22 |
| 16. Automate (v2.6) | v2.6 | Complete | Verified PASS | 2026-06-22 |
| 17. AI Native (v3.0) | v3.x | Complete | Verified PASS | 2026-06-22 |
| 18. AI Privacy (v3.5) | v3.x | see PLAN | Verified PASS | 2026-06-22 |
| 19. AI Polish | v3.x | see PLAN | Verified PASS | 2026-06-22 |
| 20. Milestone Gate | v3.x | see PLAN | Verified PASS | 2026-06-22 |
| 21. Architecture Completion | arch/obs | see PLAN | Verified PASS | 2026-06-22 |
| 22. Architecture storage | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-22 |
| 23. Architecture apigateway | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-23 |
| 24. Architecture imex | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-23 |
| 25. Architecture dashboard | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-23 |
| 26. Architecture rbac | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-23 |
| 27. Architecture auth | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-23 |
| 28. v2.6 Architecture Gate | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-23 |
| 29. OBS-01 | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-23 |
| 30. OBS-01 FE | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-23 |
| 31. OBS-01 Prometheus | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-23 |
| 32. OBS-01 OTel | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 33. Monitoring stack | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 34. Loki + alerts | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 35. Alertmanager | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 36. Alert webhook | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 37. Biome gate | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 38. v2.6 Milestone Gate | arch/obs | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 39. Intranet News | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 40. Org directory | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 41. Unified calendar | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 42. CRM-lite | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 43. HR-lite | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 44. Kudos | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 45. Marketplace | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 46. Status page | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 47. Random Coffee | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 48. 1C Connector | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-24 |
| 49. Lark Base | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-25 |
| 50. v4 Milestone Gate | v4 | see PLAN | Verified PASS | 2026-06-25 |
| 51. Accounting MVP | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-25 |
| 52. Accounting v4.5 | v4 | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-25 |
| 53. v4 Module Gate | v4 | see PLAN | Verified PASS | 2026-06-25 |
| 54. v4 Coverage Uplift | v4 | see PLAN | Verified PASS | 2026-06-25 |
| 55. onec + intranet | v4 | see PLAN | Verified PASS | 2026-06-25 |
| 56. v4 Final Gate | v4 | see PLAN | Verified PASS | 2026-06-25 |
| 57. Post-v4 CI Handoff | v4 | see PLAN | Verified PASS | 2026-06-25 |
| 58. v4 Milestone Archive | v4 | see PLAN | Verified PASS | 2026-06-25 |
| 59. Coverage + UX Close-out | cov/gates | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-25 |
| 60. UX-1 Gate | cov/gates | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-25 |
| 61. Coverage + Context Tooltips | cov/gates | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-25 |
| 62. Onboarding Profile RHF | cov/gates | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-25 |
| 63. Production Release Gate | cov/gates | see PLAN | Code PASS / deploy blocked | 2026-06-25 |
| 64. Auth Coverage Uplift | cov/gates | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-25 |
| 65. UX-P1-05 Rate Limit Timer + Messenger Coverage | cov/gates | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-25 |
| 66. Notification Coverage Uplift (42.5% → 50%) | cov/gates | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-25 |
| 67. Production Release Gate (Phases 64–66) | cov/gates | see PLAN | Code PASS / deploy blocked | 2026-06-25 |
| 68. Messenger Fan-out & Coverage (52.3% → 60%) | cov/gates | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-26 |
| 69. Apigateway Coverage + COV Baseline Refresh | cov/gates | see PLAN | Verified PASS | 2026-06-26 |
| 70. GitLab/GitHub Coverage (SEC-04) | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 71. Mail + Dashboard Coverage (MAIL-01 / BE-H3) | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 72. Production QA Gate (Post Phase 71) | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 73. Platform HTTP/Outbox Coverage (SEC-04 / NF-07) | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 74. AI Module Coverage (M3 / COV-01) | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 75. Post-AI QA Gate Refresh | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 76. Sprint Ent + Handler Routes Coverage | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 77. Storage Handler Routes (CR-04 / Presign) | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 78. Tracker + TMS Coverage Wave | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 79. Notification + Project Coverage Wave | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 80. Backup + HR + IMEX Coverage Wave | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 81. AI Coverage + QA Gate | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 82. AI Bot Hook + Doc Search RAG | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 83. AI Bot Commands + Search Query | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 84. VCS Review + Bases + Intranet Coffee | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 85. GitHub PR Review + Coverage Uplift | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 86. Search + Tracker + TMS + Notification Uplift | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 87. TMS + Notification + NATS Uplift | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 88. TMS + Notification ≥65% | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 89. E2E Wave 3 + Dashboard Uplift | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 90. IMEX Jira Import Uplift | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 91. Intranet + Observability Uplift | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 92. Messenger Polls + Search Repo | cov/gates | see PLAN | Verified PARTIAL | 2026-06-21 |
| 93. Auto-Channels NATS + Handler Gaps | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 94. Search Meili + Bot + Notification JetStream | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 95. Notification Push/TG + Tracker Automation + TMS Export | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 96. TMS TestRail/Mongo + Excel Import | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 97. Mongo CI + Webhook Dispatcher + Marketplace | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 98. Accounting + Calendar + COV 75% | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 99. Backup + Package 75% Push | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 100. RestoreBackup + Tracker/Messenger/Search 75% | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 101. Sub-75% Package Push | cov/gates | see PLAN | Verified PASS | 2026-06-21 |
| 102. Dev Routes + Notification Push | cov/gates | see PLAN | Verified PASS | — |
| 103. GitHub Push + Mongo Integration Tests | cov/gates | see PLAN | Verified PASS | — |
| 104. Flaky GitHub Fix + Mongo Testutil | cov/gates | see PLAN | Verified PASS | — |
| 105. Boundary Package Push | cov/gates | see PLAN | Verified PASS | — |
| 106. Triple Boundary Push (search / messenger / tracker) | cov/gates | see PLAN | Verified PASS | — |
| 107. CRM + Dashboard + Bases Boundary Push | cov/gates | see PLAN | Verified PASS | — |
| 108. 79% COV Milestone | cov/gates | see PLAN | Verified PASS | — |
| 109. Triple Boundary Push (Search + Messenger + Tracker) | cov/gates | see PLAN | Verified PASS | — |
| 110. TMS + AI + Project Boundary Push | cov/gates | see PLAN | Verified PASS | — |
| 111. Notification + HR + CRM Push | cov/gates | see PLAN | Verified PASS | — |
| 112. Search + Messenger + Tracker Push | cov/gates | see PLAN | Verified PASS | — |
| 113. TMS + Project + Sprint + Messenger + Tracker Push | cov/gates | see PLAN | Verified PASS | — |
| 114. Search + Messenger + TMS + Tracker Push | cov/gates | see PLAN | Verified PARTIAL | — |
| 115. GitHub/GitLab + Auth + Messenger + Bases + AI Push | cov/gates | see PLAN | Verified PARTIAL | — |
| 116. Search PG Integration + Messenger/TMS Push | cov/gates | see PLAN | Verified PASS | — |
| 117. Storage RBAC + Notification Quiet Hours + Tracker Time | cov/gates | see PLAN | Verified PASS | — |
| 118. Notification Consumer + Tracker Automation Push | cov/gates | see PLAN | Verified PASS | — |
| 119. Notification/Tracker Push to 82% | cov/gates | see PLAN | Verified PARTIAL | — |
| 120. Tracker/TMS/Project Push to 82% | cov/gates | see PLAN | Verified PASS | — |
| 121. Notification Consumer + Tracker Push to 82% | cov/gates | see PLAN | Verified PARTIAL | — |
| 122. TMS + Project ≥82% | cov/gates | see PLAN | Tasks done / no VERIFICATION.md | — |
| 123. Messenger ≥82%, Storage ≥82% | cov/gates | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-21 |
| 124. Floor ≥82% (tracker / tms / project) | cov/gates | see PLAN | Tasks done / no VERIFICATION.md | — |
| 125. COV ≥82%, Security ≥80%, Regression Gate | cov/gates | see PLAN | Tasks done / no VERIFICATION.md | 2026-06-21 |

> **Сноска о сверке (drift fixed 2026-07-03, Волна 1 задача 1.13):** чекбоксы Phase 8/9/12 и строки Progress приведены в соответствие с фактом — v2.0 Hub Redesign завершён 2026-05-15 (подтверждено по `.planning/STATE.md` «v2.0 Hub Redesign ✅ COMPLETE» и `AGENTS.md`; планы 08-01/08-02, 09-01/09-02, 12-01…12-05 отмечены [x] в Phase Details). Дополнительно удалены пять дублирующихся блоков «Plans: 12-0X-PLAN.md», ошибочно вставленных в Phase Details фаз 2, 3, 4 и 6 (v1.0), — они дублировали планы Phase 12 с противоречивыми чекбоксами. Строки Phase 15–17 обновлены по факту milestone-статусов (v2.5/v2.6/v3.0 ✅ COMPLETE, см. раздел Milestones выше).

> **Сноска C2 (2026-07-27):** Progress-таблица расширена до Phase 125; статусы VERIFIED_* / deploy-blocked взяты из `PHASE-STATUS-MATRIX.md`. Чекбокс «Complete» в Phases overview ≠ `goals_verified`.

---


---

## Backlog

> Unsequenced ideas (999.x). Promote via `/gsd-review-backlog` when ready.
