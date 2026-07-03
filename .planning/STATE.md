---
gsd_state_version: 1.0
milestone: v4
milestone_name: Hub One Super-app
status: team-reorg-complete-ready-for-wave1
stopped_at: Реорганизация AI-команды (30 агентов) + planning-пакет ГОТОВЫ · ждёт ревью пользователя → коммит → Волна 1
last_updated: "2026-07-03T02:10:00.000Z"
last_activity: 2026-07-03 team reorg 30 agents · 10 playbooks · A/B Fable-vs-Opus (Fable=team-lead) · ADR-MICROSERVICES accepted · gray-zones ~90 решений · BP-2026 дельты в playbooks · ROADMAP-FINAL + TF-PY(Python/Locust) + DOCS-EXCELLENCE
progress:
  total_phases: 125
  completed_phases: 125
  total_plans: 250+
  completed_plans: 250+
  percent: 100
---

# Состояние проекта Hub

## Справка

См.: `.planning/PROJECT.md` (обновлено 2026-03-27)

**Ключевая ценность:** единая self-hosted среда для dev-команды и сотрудников компании — трекер, TMS, мессенджер, звонки, интранет, CRM, HR, учёт, AI — в одном интерфейсе.

**Текущий фокус (2026-07-03):** planning-пакет реорганизации готов → ревью пользователя → коммит → **Волна 1** (`.planning/ROADMAP-FINAL-2026-07-03.md`). FA-06 — ручная ротация по runbook (пользователь), больше не гейтит разработку.

## Session note — 2026-07-03 (be1 · Волна 2 · задача 2.6 OTel trace через NATS)

- **Ветка:** `feat/platform/otel-nats-trace` — сквозной trace-context REST → outbox → NATS → consumer (жёсткий GATE пилота микросервисов).
- **Outbox:** новая nullable JSON-колонка `trace_context` в `OutboxEvent` (map traceparent/tracestate/baggage); `Publisher.Publish/PublishInTx` сохраняют контекст из ctx. Envelope в payload отвергнут — сломал бы `json.Unmarshal` во всех consumers.
- **NATS:** `platform/nats/trace.go` — W3C-пропагатор (явный, не глобальный: глобальный no-op без OTLP endpoint), `Inject/ExtractTraceContext` поверх `nats.Header`, `TraceContextMap`/`ExtractTraceContextFromMap` для персистентности; header `Hub-Event-Id`. Client: `PublishMsg`, `PublishWithTraceContext`, `PublishJSONContext`. Worker кладёт traceparent + event_id в headers при publish.
- **Consumers:** `observability.StartConsumerSpan/EndConsumerSpan` (span `consume <subject>`, kind=Consumer, атрибуты messaging.system/destination + hub.event_id) — применено в notification consumer, audit consumer, ws bridge (обёртка `traced`).
- **Тесты:** round-trip header/map, outbox store/restore, worker headers, consumer parent span — `testframework/platform/{outbox,observability}`. Гейты: build release ✅ · vet hubtest ✅ · race platform+notification+audit+ws ✅.

## Session note — 2026-07-03 (реорганизация команды и планирование)

- **Команда:** 30 агентов (`.cursor/agents/`), параллельные потоки be1/be2, fe1/fe2, ba1/2, sa1/2, qa-auto1/2 + верификаторы у всех; `team-lead` = **Fable 5** (по итогам 5 A/B-задач против Opus 4.8), `team-lead-opus` = эскалация; остальные Sonnet 5.
- **Playbooks:** 10 файлов `.cursor/playbooks/` (~15k строк) + интегрированы 27 дельт best practices 2026 (Fiber v3 stable, govulncheck/gitleaks, versioned Atlas, synctest, TanStack preload-канон, React 19 Actions, WCAG 2.2, Playwright шардинг/storageState).
- **Архитектура:** `ADR-ARCH-MICROSERVICES-2026` accepted — schema-per-service, strangler-fig, каталог: core + hub-ai + realtime-трио (ws/messenger/calls) + notification-пилот; tracker исключён; Go/No-Go композит.
- **Серые зоны:** ~90 вопросов → решения в `GRAY-ZONES-DECISIONS-2026-07-03.md`; правила 50/56/60 обновлены (агенты коммитят/пушат фича-ветки, **push в master — только пользователь**; rebase-merge; регрессия домена у агента + полный gate на волну; SSO за ADR).
- **QA-конвейер:** бизнес-дока → TC (qa-manual) → верификация доки → автотесты → верификация; **юниты пишут разработчики**. Целевой TF: **Python** (pytest+allure+httpx+playwright-python) + **Locust** — фаза TF-PY после микросервисного дробления.
- **Roadmap:** `.planning/ROADMAP-FINAL-2026-07-03.md` — Волны 1–2 атомарно (24 плана), эпик DOCS-EXCELLENCE (LICENSE ✅ создана — проприетарная, README/SECURITY/тест-стратегия/онбординг — в работе), TF-PY, дальние волны укрупнённо.
- **`.cursor/` вынут из `.gitignore`** — команда трекается целиком.
- Research-артефакты: `EXPANSION-ANALYSIS`, `MARKET-AND-NEEDS`, `BEST-PRACTICES-2026-*`, A/B-планы (`ab-microservices/`, `ab-tests/`), gray-zones (6 областей).

---

## Текущая позиция (2026-06-22)

| Параметр | Значение |
|----------|----------|
| Фаза | **Wave C** — FA-01–FA-12 ✅ (кроме FA-06 manual) · RU-V10 PASS · FA-10 qa-gate PASS |
| Ветка | `master` · HEAD `bbdaf16` |
| COV overall (`internal/`) | **83.0%** (Docker, `-tags hubtest`, post FA-11/12) |
| Security packages ≥80% | **6/6** ✅ |
| Product score | **89/100** (post P1 backlog close) |
| REST API | **286+** endpoints |
| Backend tests | **495** files в `backend/testframework/` · **0** в `internal/*_test.go` (TF-M08 ✅) |
| Playwright E2E | **20** specs (`testframework/stages/e2e/domains/`) |
| Frontend Vitest | **99/99** pass · 37 files `testframework/stages/unit/frontend-test/domains/` · 0 in `src/` |

### Покрытие ключевых пакетов (Docker, Phase 125)

| Пакет | Coverage |
|-------|----------|
| notification | **82.5%** |
| auth | **82.4%** |
| tracker | **82.4%** |
| messenger | **82.3%** |
| ws | **87.7%** |
| storage | **86.8%** |
| rbac | **84.7%** |
| audit | **80.5%** |
| tms | **82.0%** (FA-11 ✅ micro-gap closed) |
| project | **82.0%** (FA-11 ✅ micro-gap closed) |

---

## Milestones — статус

| Milestone | Phases | Статус | Дата |
|-----------|--------|--------|------|
| v1.0 MVP | 1–7 | ✅ SHIPPED | 2026-03-21 |
| v2.0 Hub Redesign | 8–14 | ✅ COMPLETE | 2026-05-15 |
| v2.5 Trust | 15 | ✅ COMPLETE | 2026-06-22 |
| v2.6 Automate + Arch + OBS | 16–38 | ✅ COMPLETE | 2026-06-24 |
| v3.0 AI Native | 17–19 | ✅ COMPLETE | 2026-06-22 |
| v4 Super-app | 39–58 | ✅ COMPLETE | 2026-06-25 |
| Coverage wave | 59–125 | ✅ COMPLETE | 2026-06-21 |
| Test framework migration | TF-M01–M10 | ✅ Wave B + pipeline domains (TF-M10) | 2026-06-23 |
| Final audit | FA-01–FA-12 | ✅ Wave C COMPLETE (кроме FA-06 manual) | 2026-06-22 |

Архив v4: `.planning/milestones/v4-super-app/SUMMARY.md`

---

## PO Decisions (locked)

См.: `.planning/DISCUSS-ROADMAP-LOCKED.md`

- **AUTH:** Magic Link only · org table v2.5 · founder role · Telegram только bind/delivery
- **NOTIF:** TG bind · hub-mail Phase 15.2 · email marketing v2.6
- **Phase 15 exit:** AUTH+NOTIF+E2E+WCAG+run_failed+Biome+LiveKit CI · coverage ≥50% ✅
- **TZ rewrite:** Gantt, webcam, marketing, no-code, accounting — IN (by phase)
- **Coverage target:** ≥82% overall · security 6/6 ≥80% ✅

---

## Test Framework Migration — ✅ COMPLETE (Wave B)

| ID | Задача | Статус |
|----|--------|--------|
| TF-M01–M03 | export_hooks + auth/ws/rbac/audit/notification | ✅ |
| TF-M04 | tracker/tms/project/messenger → testframework/ | ✅ Docker PASS |
| TF-M05 | remaining internal tests | ✅ migrated |
| TF-M06 | Vitest → `testframework/stages/unit/frontend-test/domains/` | ✅ 37 files · 99/99 · 0 in `src/` |
| TF-M07 | E2E in `testframework/stages/e2e/domains/` | ✅ 20 specs |
| TF-M10 | Pipeline manifests + shared POM in `stages/e2e/shared/` | ✅ 2026-06-22 |
| TF-M08 | delete `internal/*_test.go` | ✅ **0 remaining** (verified) |
| TF-M09 | CI: `go test -tags hubtest ./testframework/...` | ✅ full PASS |

**Итог (verified 2026-06-22):** `backend/testframework/` = **495** `*_test.go` · `backend/internal/` = **0** · `testframework/stages/unit/frontend-test/domains/` = **37** · `testframework/stages/e2e/domains/` = **20**.

## Wave C Final Audit — ✅ MOSTLY COMPLETE (2026-06-22)

| ID | Задача | Статус |
|----|--------|--------|
| FA-01 | Full file-by-file audit + remediation | ✅ 2026-06-22 · `.planning/FA-01-AUDIT.md` |
| FA-01 P0 | APIGW-01 webhook CRUD RBAC | ✅ 2026-06-22 |
| FA-01 P1 | APIGW-02 webhook Ack semantics (`NakWithDelay`) | ✅ 2026-06-22 · `webhook_dispatcher_apigw02_test.go` |
| FA-01 P1 | APIGW-03 JetStream FilterSubjects on webhook consumer | ✅ 2026-06-22 · `webhook_dispatcher_apigw03_test.go` |
| FA-01 P1 | AUTH-01 RefreshToken session validation | ✅ 2026-06-22/23 · `service_session.go` · `service_session_test.go` |
| FA-01 P2 | APIGW-04 API key deactivated user check | ✅ 2026-06-22/23 · `apikey.go` · `apikey_auth_test.go` |
| FA-01 P2 | WebhookDispatcher `Stop` inFlight race fix | ✅ 2026-06-22/23 · `webhook.go` (`inFlight.Wait()` before `close(jobs)`) |
| FA-01 P2 | BOOT-01 notification consumer Stop cleanup | ✅ 2026-06-22 · `wiring_notification.go` |
| FA-01 P3 | WS-01 Hub graceful shutdown | ✅ 2026-06-22 · `ws/hub.go` · `hub_shutdown_test.go` |
| FA-01 P3 | APIGW-05 domain errors (no Fiber in services) | ✅ 2026-06-22 · `apigateway/errors.go` |
| FA-01 CRIT | FA-RBAC-FIBER3 Fiber v3 middleware order | ✅ 2026-06-22 · **123 routes** · `.planning/FA-RBAC-FIBER3-AUDIT.md` |
| FA-02 | Security re-audit (post TF) | ✅ 2026-06-22 · 0 CRITICAL · `.planning/FA-02-SECURITY-REAUDIT.md` |
| FA-03 | Docs sync (STATE/COVERAGE/API) | ✅ 2026-06-22 |
| FA-04 | github SafeHTTP client (no DefaultClient) | ✅ 2026-06-22 · `github/service.go:55` |
| FA-05 | hub-mail in prod compose | ✅ 2026-06-22 |
| FA-06 | PG-03 secret rotation | ⏳ **manual ops** (runbook ready) |
| FA-07 | Admin console.error → toastApiError | ✅ 2026-06-22 |
| FA-08 | main.go bootstrap extraction | ✅ 2026-06-22 · `wiring_*.go` split |
| FA-09 | E2E 20 specs → full POM+Steps | ✅ 2026-06-22 |
| FA-10 | `/qa-gate` Docker + Playwright + tsc | ✅ 2026-06-22 **PASS** (local) · `.planning/FA-10-QA-GATE.md` |
| FA-11 | tms/project 81.8% → 82% micro-gap | ✅ 2026-06-22 · `service_fa11_test.go` |
| FA-12 | TestRail `mapTestRailType` → ent `classic`/`bdd` | ✅ 2026-06-22 · `tms/testrail_import.go:69` |
| RU-V10 | RU comments/docs final gate | ✅ 2026-06-22 PASS · `.planning/RU-V-REPORT.md` |

**FA-10 qa-gate:** build ✅ · hubtest ✅ · tsc ✅ · Biome ✅ · Vitest 99/99 ✅ · race ✅ · e2e-smoke 3/3 ✅ · coverage-gate CI ✅ · e2e-full/Docker cov — CI only.

Детали: `.planning/TASKS-PHASES-121-FINAL-AUDIT.md`

---

## Pending Todos

- **FA-06 / PG-03:** Phase 0 ✅ · scripts: `fa06-rotate-p1.sh`, `fa06-rotate-p1-k3s.sh`, `fa06-rotate-p2-smoke.sh`, `v25-release-gate.sh` · **Phase 1–3 manual**
- **SSH deploy:** `scripts/deploy-preflight.sh` + GitHub Secrets `SSH_HOST`/`SSH_KEY` · firewall :22
- ~~**UX-1:**~~ ✅ PR #15 merged · ~~**CI flake:**~~ ✅ #16 docker retry · #22 gitlab test · **master CI green** (run 27962930096)
- ~~**Dependabot batch:**~~ ✅ #17–#21 merged 2026-06-22 (recharts, happy-dom, radix select/tooltip, atlaskit dnd 2.x)
- **v2.5.0 tag:** after FA-06 Phase 1–3 + CTO sign-off

---

## Недавние сессии

### 2026-06-22 — Quality processes audit + docs sync ✅

- Полный аудит 15 процессов качества (REGULATION, rules 45/55/56, `/qa-gate`, verifiers, CI)
- Отчёт: `.planning/AUDIT-2026-06-22-quality-processes.md` · scorecard обновлён
- Локальная регрессия: build ✅ · vet ✅ · tsc ✅ · lint ✅ · vitest **99/99** ✅
- Docs sync: `test_plan`, `TESTING`, `stack_and_tools`, `code_style_guide`, `vision_and_scope`, `COVERAGE`, `AGENTS`, `REGULATION`
- Backlog: **TF-CLEANUP-01** (legacy `testframework/e2e/specs/`), **RU-COMMENTS-01**, **FA-06**

### 2026-06-22 — PR #25 K3s runbook + v2.5 release gate ✅

- **HEAD** `8317412` — `fa06-rotate-p1-k3s.sh`, `v25-release-gate.sh`
- Ops entry: `scripts/fa06.sh` (p1-compose | p1-k3s | p2-smoke | deploy | release-gate)

### 2026-06-22 — FA-06 ops scripts + release gate ✅

- **PR #24** `deploy-preflight.sh`, `fa06-rotate-p2-smoke.sh` → HEAD `beb0022`
- **Build & Deploy** run 27965584506: **success** (images pushed; deploy best-effort)
- K3s + v2.5 tag gate scripts added on master

### 2026-06-22 — Master CI green + Dependabot batch ✅

- **Master CI** run 27962930096: all jobs pass incl. **coverage-gate ≥82%** · e2e-full
- **PR #16** docker compose retry · **PR #22** gitlab pipelines mock · HEAD `a9f931c`
- **Dependabot #17–#21** merged: recharts, happy-dom, radix select/tooltip, atlaskit/pragmatic-drag-and-drop 2.x
- **FA-06 Phase 0** git scan filled → `.planning/FA-06-EXECUTION-LOG.md` (local; no secret values)

### 2026-06-22 — UX-1 command palette ✅ PR #15

- **PR #15** merged → `64dce64` — Ctrl+K actions: Calls, Settings, Admin + i18n keys (en/ru)
- E2E: `command-palette.spec.ts` action shortcuts → `/calls`, `/settings/profile`
- Onboarding Zod step 4 + issue-form RHF+Zod already on master (FA-06 wave)

### 2026-06-22 — APIGW-05 domain errors ✅

- **APIGW-05:** Removed `fiber.NewError` from `APIKeyService` / `WebhookService`; added `errors.go` with typed sentinels + `MapToFiber()` in handler/middleware only.
- **Tests:** `go test -tags hubtest -count=1 ./testframework/apigateway/...` — PASS (38s).

### 2026-06-22/23 — AUTH-01, APIGW-04, WebhookDispatcher Stop race fix ✅

- **AUTH-01 (P1):** `RefreshToken` — DB session lookup + expiry validation; reject missing/expired sessions before Redis rotation (`service_session.go`, `service_session_test.go`)
- **APIGW-04 (P2):** API key auth path rejects deactivated owner — parity with JWT `RequireAuth` (`apikey.go`, `apikey_auth_test.go`)
- **WebhookDispatcher Stop race:** `inFlight sync.WaitGroup` — consumer callbacks tracked; `Stop()` waits `inFlight` before `close(jobs)` to prevent DATA RACE on shutdown (`webhook.go`)
- **FA-10:** `go test -tags hubtest -race` full suite **PASS** post-fix → `.planning/FA-10-QA-GATE.md`

### 2026-06-22 — Wave C complete (FA-01–FA-12, кроме FA-06) ✅

- **FA-01** file-by-file audit + remediation → `.planning/FA-01-AUDIT.md`
- **APIGW-01** webhook CRUD RBAC · **APIGW-02** Ack/`NakWithDelay` semantics
- **FA-RBAC-FIBER3** Fiber v3 middleware order → **123 routes** fixed (tracker 46, sprint 11, rbac 6, github 5, gitlab 5 + tms 50 prior) → `.planning/FA-RBAC-FIBER3-AUDIT.md`
- **FA-02** security re-audit: 0 CRITICAL · **FA-04** github safehttp · **FA-11/12** tms/project + TestRail enum
- **FA-05** hub-mail prod compose · **FA-07** admin toast · **FA-08** main.go `wiring_*.go` · **FA-09** E2E POM+Steps
- **FA-10** qa-gate **PASS** (local): build · hubtest · tsc · Biome · Vitest 99/99 → `.planning/FA-10-QA-GATE.md`
- **RU-V10** final gate **PASS** (post RU-FIX-11–13) → `.planning/RU-V-REPORT.md`
- **FA-03** docs sync: STATE + COVERAGE + API.md; COV **83.0%** · 488 testframework / 0 internal tests
- **Остаётся:** FA-06 PG-03 secret rotation (manual ops)

### 2026-06-22 — Wave B Test Framework Migration ✅

- TF-M04–M08: all internal tests migrated · **0** `internal/*_test.go`
- `go test -tags hubtest ./testframework/...` — full PASS (Docker)
- TF-M06: Vitest **99/99** · 37 files `frontend/testframework/unit/` · 0 in `src/`
- Next: RU-V FIX wave → Wave C FA-01–FA-12

### 2026-06-21 — Phases 121–125 Coverage Wave ✅

- Phase 121: notification **82.5%**, tracker **82.4%**
- Phase 123: messenger **82.3%**, storage **86.8%**
- Phase 125: COV **82.6%**, security **6/6 ≥80%**
- TF-M02–M03: testframework/notification, auth, ws, rbac, audit — Docker PASS
- TF-M04: 154 files moved; compile fix in progress

### 2026-06-20 — Wave 6 prod + Ctrl+K

- PG-01 calls **80.2%** · Dockerfile `-tags release` · deploy gated on CI
- Ctrl+K global search UI (`command-palette.tsx`)
- Rule: `.cursor/rules/55-docs-sync.mdc`

### 2026-06-20 — UX-0 + audit loop

- UserPicker, skip-link, `/help`, Kanban «Move to…»
- Product score 74→93 path via IMPROVEMENT-MASTER-PLAN
- Vitest 39→99+ tests

---

## Блокеры / риски

| Риск | Mitigation |
|------|------------|
| ~~RU-V gaps (EN comments, encoding)~~ | ✅ RU-V10 PASS 2026-06-22 |
| ~~tms/project −0.2% gap~~ | ✅ FA-11 closed (82.0%) |
| ~~TMS CI webhook без RBAC (FA-02-02)~~ | ✅ FA-RBAC-FIBER3 + APIGW-01 2026-06-22 |
| ~~Fiber v3 RBAC middleware order~~ | ✅ FA-RBAC-FIBER3 123 routes 2026-06-22 |
| PG-03 not executed | FA-06 manual ops before prod cutover |
| ~~Uncommitted v4 block on branch~~ | ✅ PR #1 squash merged 2026-06-22 |
| ~~main.go ~629 LOC tech debt~~ | ✅ FA-08 wiring split 2026-06-22 |

---

## Session Continuity

**Last session:** 2026-06-22/23  
**Stopped at:** AUTH-01 · APIGW-04 closed · WebhookDispatcher Stop race fix · FA-10 race PASS  
**Resume:** FA-06 PG-03 Phase 1–3 (manual) → `bash scripts/fa06.sh release-gate` → `git tag v2.5.0`

### 2026-06-23 — Docs audit + cleanup ✅

- **AUDIT:** `.planning/AUDIT-2026-06-23.md` — полный аудит, P0/P1/P2 остаток
- **Cleanup:** удалено 169+ coverage-артефактов (`backend/cov*`, `cover*`, `*.txt` отчёты)
- **Docs sync:** AGENTS, system_spec, vision, competitive, test_plan → master `bbdaf16`
- **PR #27:** Lighthouse CI artifact (non-blocking)

**Next session priorities:**
1. FA-06: PG-03 secret rotation (manual ops) — **P0**
2. SSH deploy + `git tag v2.5.0`
3. NFR-16 WCAG full sign-off · NFR-17 Lighthouse blocking gate
4. FA-02 backlog: LiveKit `iss` (FA-02-03), RequireAuth fail-open (FA-02-04)

### 2026-07-03 — Best practices 2026 → playbooks ✅

- Источник: `.planning/research/BEST-PRACTICES-2026-frontend-testing.md` (21 дельта)
- **frontend.md:** WCAG 2.2 AA (§9), Router preload `intent` + `ensureQueryData`/`useSuspenseQuery` канон (§4.4), Actions-триада React 19 (§3.9), React Compiler «пилот за ADR» (§3.7), CWV-бюджеты INP<150ms (§11.0), Tailwind `@theme` (§8.1) + container queries (§8.5); Приложение B — 8 дельт «на обсуждение»
- **testing.md:** `test.step()` (§6.1, §8.4), storageState setup-project (§6.5), шардинг + blob/merge-reports (§6.8), Vitest 4 Browser Mode note (§5.1), Allure 3 + allure-vitest wrapper (§8.6)
- **design.md:** цель WCAG 2.2 AA, §5.7 новые критерии (2.5.7 Dragging / 2.5.8 Target Size ≥24px / 2.4.11 Focus Not Obscured), target size в §5.5
- Принято в playbooks: 13 дельт · на обсуждение (Приложение B frontend.md): 8 · отклонённых нет

---

*Обновляется после каждой фазы и `/handoff-sync`*
