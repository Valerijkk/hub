---
gsd_state_version: 1.0
milestone: v4
milestone_name: Hub One Super-app
status: wave-c-mostly-complete
stopped_at: Wave B ✅ · Wave C FA-01–FA-12 ✅ (except FA-06 manual) · RU-V10 PASS · FA-10 qa-gate PASS
last_updated: "2026-06-23T00:30:00.000Z"
last_activity: 2026-06-22/23 AUTH-01 · APIGW-04 · WebhookDispatcher inFlight Stop race fix · FA-10 race PASS
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

**Текущий фокус:** Wave C final audit — **FA-01–FA-12 ✅** (кроме **FA-06** manual ops) · squash PR `feat/v2.5-trust`

---

## Текущая позиция (2026-06-22)

| Параметр | Значение |
|----------|----------|
| Фаза | **Wave C** — FA-01–FA-12 ✅ (кроме FA-06 manual) · RU-V10 PASS · FA-10 qa-gate PASS |
| Ветка | `feat/v2.5-trust` (uncommitted) |
| COV overall (`internal/`) | **83.0%** (Docker, `-tags hubtest`, post FA-11/12) |
| Security packages ≥80% | **6/6** ✅ |
| Product score | **93/100** |
| REST API | **286+** endpoints |
| Backend tests | **488** files в `backend/testframework/` · **0** в `internal/*_test.go` (TF-M08 ✅) |
| Playwright E2E | **20** specs (`testframework/e2e/specs/`) |
| Frontend Vitest | **99/99** pass · 37 files `frontend/testframework/unit/` · 0 in `src/` |

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
| Test framework migration | TF-M01–M10 | ✅ Wave B COMPLETE (0 internal tests) | 2026-06-22 |
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
| TF-M06 | Vitest → `frontend/testframework/unit/` | ✅ 37 files · 99/99 · 0 in `src/` |
| TF-M07 | E2E in testframework/e2e/specs/ | ✅ 20 specs |
| TF-M08 | delete `internal/*_test.go` | ✅ **0 remaining** (verified) |
| TF-M09 | CI: `go test -tags hubtest ./testframework/...` | ✅ full PASS |

**Итог (verified 2026-06-22):** `backend/testframework/` = **488** `*_test.go` · `backend/internal/` = **0** · `frontend/testframework/unit/` = **37** · `testframework/e2e/specs/` = **20**.

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

- **FA-06 / PG-03:** Secret rotation (manual ops, runbook ready) — единственный блокер prod cutover
- **SSH deploy:** GitHub Actions deploy job — firewall :22 / `SSH_HOST` secrets (continue-on-error on master)
- **Post-merge CI:** master `64dce64` — e2e-full failed Docker Hub pull timeout (transient); retry in `ci-e2e-backend.sh`
- ~~**UX-1:**~~ ✅ **DONE 2026-06-22** — PR #15 merged (`calls`, `settings`, `admin` in Ctrl+K; i18n; E2E shortcuts)
- ~~**Squash PR v2.5-trust:**~~ ✅ merged #1 (2026-06-22) + hotfixes #12–#14

---

## Недавние сессии

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
| Uncommitted v4 block on branch | squash PR after FA-06 + CTO approval |
| ~~main.go ~629 LOC tech debt~~ | ✅ FA-08 wiring split 2026-06-22 |

---

## Session Continuity

**Last session:** 2026-06-22/23  
**Stopped at:** AUTH-01 · APIGW-04 closed · WebhookDispatcher Stop race fix · FA-10 race PASS  
**Resume:** FA-06 PG-03 secret rotation (manual) → squash PR

**Latest session (2026-06-22/23):**
- **AUTH-01** RefreshToken DB + expiry validation
- **APIGW-04** API key deactivated-user check
- **WebhookDispatcher** `inFlight.Wait()` before `close(jobs)` — race suite green

**Wave C highlights (2026-06-22):**
- **APIGW-01** webhook CRUD RBAC · **APIGW-02** Ack/`NakWithDelay` (`webhook_dispatcher_apigw02_test.go`)
- **FA-RBAC-FIBER3** 123 routes — Fiber v3 `[rbacMW, handler]` order (`.planning/FA-RBAC-FIBER3-AUDIT.md`)
- **FA-10** qa-gate PASS local: build · hubtest · tsc · Biome · Vitest 99/99 (`.planning/FA-10-QA-GATE.md`)
- **RU-V10** PASS post RU-FIX-11–13 (`.planning/RU-V-REPORT.md`)

**Next session priorities:**
1. FA-06: PG-03 secret rotation (manual ops)
2. FA-10 follow-up: ~~race suite~~ ✅ · ~~coverage-gate CI~~ ✅ · Playwright E2E
3. Squash PR `feat/v2.5-trust` → main when CTO approves
4. FA-02 backlog: LiveKit `iss` const-time (FA-02-03), `RequireAuth` fail-open (FA-02-04)

---

*Обновляется после каждой фазы и `/handoff-sync`*
