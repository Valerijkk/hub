# Squash PR: `feat/v2.5-trust` → `main`

> **ARCHIVED** · Merged as **PR #1** · 2026-06-22  
> **Branch:** was `feat/v2.5-trust` · текущая ветка `master`  
> **Merge strategy:** squash merge ✅ выполнен  
> **Status:** ~~pending~~ → **MERGED**. Prod-блокер: FA-06 PG-03 manual ops.

---

## Summary

This branch delivers **v2.5 Trust** production hardening: full test-framework migration (Wave B), final audit remediation (Wave C FA-01–FA-12), RU documentation gate, and security fixes across auth, API gateway, and RBAC.

### Wave B — Test Framework Migration ✅

- Migrated all backend tests from `backend/internal/*_test.go` → `backend/testframework/` (`-tags hubtest`, `export_hooks.go` white-box pattern)
- **488** Go test files in `testframework/` · **0** remaining in `internal/` (TF-M08 verified)
- Vitest moved to `frontend/testframework/unit/` — **37** files, **99/99** pass, **0** in `src/`
- E2E consolidated under `testframework/e2e/specs/` — **20** Playwright specs (POM + Steps, FA-09)
- CI updated: `go test -tags hubtest ./testframework/...` only (TF-M09)

### Wave C — Final Audit FA-01–FA-12 ✅ (except FA-06 manual)

| Area | Deliverables |
|------|--------------|
| **FA-01** | Full file-by-file audit + remediation (`.planning/FA-01-AUDIT.md`) |
| **APIGW** | APIGW-01 webhook CRUD RBAC · APIGW-02 Ack/`NakWithDelay` · APIGW-03 JetStream FilterSubjects · APIGW-04 deactivated-user check on API key path |
| **AUTH** | AUTH-01 RefreshToken DB session + expiry validation before Redis rotation |
| **RBAC** | FA-RBAC-FIBER3 — Fiber v3 middleware order fixed on **123 routes** (`.planning/FA-RBAC-FIBER3-AUDIT.md`) |
| **Security** | FA-02 re-audit: **0 CRITICAL** · FA-04 github `safehttp` (no `DefaultClient`) · FA-02-03 LiveKit `iss` const-time · FA-02-04 `RequireAuth` fail-closed on DB error |
| **Infra** | FA-05 hub-mail · FA-08 wiring split · Playwright config → `frontend/` |
| **P3 cleanup** | WS-01 hub shutdown · APIGW-05 domain errors in apigateway |
| **Frontend** | FA-07 admin `console.error` → `toastApiError` |
| **Coverage** | FA-11 tms/project micro-gap closed (81.8% → **82.0%**) · FA-12 TestRail `mapTestRailType` → ent enum `classic`/`bdd` |
| **QA** | FA-10 `/qa-gate` PASS (local) — `.planning/FA-10-QA-GATE.md` |
| **Docs** | FA-03 docs sync (STATE, COVERAGE, API.md) |

**Not in this PR (manual ops):** FA-06 / PG-03 secret rotation — see [Blockers](#blockers).

### RU-V10 — Documentation Gate ✅

- Final RU comments/docs verification **PASS** (post RU-FIX-11–13) — `.planning/RU-V-REPORT.md`

### Security Fixes (highlights)

- **AUTH-01:** Refresh token rotation validates DB session existence and expiry
- **APIGW-04:** API key auth rejects keys owned by deactivated users (JWT parity)
- **WebhookDispatcher:** `inFlight sync.WaitGroup` — `Stop()` waits callbacks before `close(jobs)` (race fix)
- **FA-RBAC-FIBER3:** 123 routes — correct `[rbacMW, handler]` chain order (tracker 46, sprint 11, tms 50, rbac 6, github 5, gitlab 5)
- **ME-13 WS parity:** deactivated users blocked on WebSocket upgrade (`IsUserDeactivated` before `hub.Register`)
- **FA-02-03/04:** LiveKit webhook timing-safe compare · RequireAuth 503 fail-closed

---

## Key Metrics

| Metric | Value |
|--------|-------|
| Go test files (`backend/testframework/`) | **488** |
| Go test files (`backend/internal/`) | **0** |
| COV overall (`internal/`, Docker `-tags hubtest`) | **83.0%** |
| Security packages ≥80% | **6/6** ✅ |
| RBAC routes fixed (FA-RBAC-FIBER3) | **123** |
| Product score | **93/100** |
| REST API endpoints | **286+** |
| Frontend Vitest | **99/99** (37 files) |
| Playwright E2E specs | **20** |
| Ent schemas | **82** |

### Package coverage (Docker, post FA-11/12)

| Package | Coverage |
|---------|----------|
| ws | 87.7% |
| storage | 86.8% |
| rbac | 84.7% |
| notification | 82.5% |
| auth | 82.4% |
| tracker | 82.4% |
| messenger | 82.3% |
| tms | 82.0% |
| project | 82.0% |
| audit | 80.5% |

---

## Test Plan

### Pre-merge (reviewer / CI)

- [ ] **Commit** all uncommitted changes on `feat/v2.5-trust` before opening PR
- [ ] CI green: `backend-test` (`go test -tags hubtest ./testframework/...`)
- [ ] CI green: `coverage-gate` (≥82% on `internal/`)
- [ ] CI green: frontend — `tsc`, Biome lint, Vitest 99/99
- [ ] CI green: Playwright `e2e-full` (20 specs)
- [ ] CI green: `go build -tags release ./...`
- [ ] Race detector: `go test -tags hubtest -race ./testframework/...` (local or CI)

### Post-merge smoke (staging)

- [ ] Magic Link login + refresh token rotation (AUTH-01 path)
- [ ] API key auth with active vs deactivated user (APIGW-04)
- [ ] Tracker/TMS endpoints — member vs outsider 403 (FA-RBAC-FIBER3 sample)
- [ ] Webhook CRUD — RBAC enforced (APIGW-01)
- [ ] Admin UI errors surface as toasts, not `console.error` (FA-07)
- [ ] Ctrl+K global search (Wave 6 carryover)

### Production gate (before cutover)

- [ ] **FA-06 / PG-03:** Execute secret rotation runbook (`.planning/PG-03-SECRET-ROTATION.md`)
- [ ] Verify Resend/Telegram/LiveKit/JWT secrets rotated per runbook
- [ ] CTO sign-off on squash merge

---

## Blockers

| ID | Blocker | Owner | Status |
|----|---------|-------|--------|
| **FA-06** | **PG-03 secret rotation** — manual ops required before production cutover | Ops / CTO | ⏳ runbook ready (`.planning/PG-03-SECRET-ROTATION.md`) |

> **Note:** FA-06 does not block merge to `main` for staging/CI validation, but **must be completed before production deploy**. See `.planning/FA-06-EXECUTION-LOG.md`.

---

## Branch State

⚠️ **Uncommitted changes on branch** — working tree has pending modifications (backend, frontend, CI, docs, `.planning/`). Commit and push before opening the PR.

---

## References

| Document | Purpose |
|----------|---------|
| `.planning/STATE.md` | Current project state |
| `.planning/TASKS-PHASES-121-FINAL-AUDIT.md` | Wave A/B/C task breakdown |
| `.planning/FA-10-QA-GATE.md` | QA gate results |
| `.planning/FA-01-AUDIT.md` | File-by-file audit |
| `.planning/FA-RBAC-FIBER3-AUDIT.md` | 123-route RBAC fix |
| `.planning/FA-02-SECURITY-REAUDIT.md` | Security re-audit |
| `.planning/RU-V-REPORT.md` | RU documentation gate |
| `.planning/PG-03-SECRET-ROTATION.md` | FA-06 runbook |

---

*Draft for squash PR — no git commits in docs-writer session.*
