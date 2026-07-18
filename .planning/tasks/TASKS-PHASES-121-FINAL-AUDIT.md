# TASKS — Phases 121+ · Test Framework Migration · Final Audit

> Updated: 2026-06-23 (AUTH-01 · APIGW-04 · WebhookDispatcher Stop race fix) · Branch: `feat/v2.5-trust` · Tests: **Docker only** (`-tags hubtest`)

---

## Wave A — Coverage Phases ✅ COMPLETE

| Phase | Target | Result (Docker) | Status |
|-------|--------|-----------------|--------|
| **121** | notification ≥82%, tracker ≥82% | notif **82.5%** · tracker **82.4%** | ✅ |
| **122** | tms ≥82%, project ≥82% | tms **81.8%** · project **81.7%** | ✅ (gap → 124) |
| **123** | messenger ≥82%, storage ≥82% | messenger **82.3%** · storage **86.8%** | ✅ |
| **124** | tracker/tms/project floor ≥82% | tracker **82.4%** · tms **81.8%** · project **81.7%** | ✅ tracker; tms/project −0.2% |
| **125** | COV ≥82%, security ≥80% | COV **82.6%** · auth **82.4%** · ws **87.7%** · rbac **84.7%** · audit **80.5%** | ✅ |

**Docker verify (standard):**
```powershell
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default `
  -v "${PWD}/backend:/app" -v hub-go-mod-cache:/go/pkg/mod -w /app `
  -e CGO_ENABLED=1 `
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" `
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" `
  golang:1.25-bookworm `
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./testframework/... && go test -tags hubtest -coverprofile=/tmp/cov -coverpkg=./internal/... ./testframework/... && go tool cover -func=/tmp/cov | tail -1"
```

---

## Wave B — Test Framework Migration ✅ COMPLETE

| ID | Task | Scope | Status |
|----|------|-------|--------|
| TF-M01 | `export_hooks.go` + `-tags hubtest` per module | hooks for external tests | ✅ auth/ws/rbac/audit/tracker/messenger |
| TF-M02 | Migrate **notification** | `backend/testframework/notification/` (36 files) | ✅ Docker PASS |
| TF-M03 | Migrate **auth, ws, rbac, audit** | security-critical | ✅ Docker PASS (audit 18 files) |
| TF-M04 | Migrate **tracker, tms, project, messenger** | core modules | ✅ Docker PASS |
| TF-M05 | Migrate remaining modules | storage/calls/search/ai/… | ✅ migrated |
| TF-M06 | Move Vitest → `frontend/testframework/unit/` | 37 files | ✅ 99/99 · 0 in `src/` |
| TF-M07 | E2E in `testframework/e2e/specs/` | 20 specs | ✅ |
| TF-M08 | Delete `internal/*/_test.go` after migration | cleanup | ✅ **0 remaining** (verified) |
| TF-M09 | CI: `go test -tags hubtest ./testframework/...` | ci.yml | ✅ full PASS |
| TF-M10 | `scripts/migrate-go-tests.ps1` + `fix-testframework-prefix.ps1` | automation | ✅ |

**Verified 2026-06-22:** `backend/testframework/` = **488** `*_test.go` · `backend/internal/` = **0**.

### TF-M03 subtasks (after script fix)

| ID | Module | Files | export_hooks | Status |
|----|--------|-------|--------------|--------|
| TF-M03a | auth | 26 | ✅ | ✅ testframework/auth — Docker PASS |
| TF-M03b | ws | 8 | ✅ | ✅ testframework/ws — Docker PASS |
| TF-M03c | rbac | 6 | ✅ | ✅ testframework/rbac — Docker PASS |
| TF-M03d | audit | 18 | ✅ | ✅ testframework/audit — Docker PASS |

**Go constraint:** white-box → `export_hooks.go` (`//go:build hubtest`) + `package mod_test` in `backend/testframework/<mod>/`.

---

## Wave C — Final Audit — ✅ MOSTLY COMPLETE (2026-06-22)

> Полный file-by-file аудит, best practices, docs. **Единственный остаток:** FA-06 manual ops.

| ID | Task | Agent/Skill | Status |
|----|------|-------------|--------|
| FA-01 | Full file-by-file audit (DRY/KISS/OOP/SOLID) | gsd-audit-fix + code-reviewer | ✅ 2026-06-22 · `.planning/audits/2026-06/FA-01-AUDIT.md` |
| FA-01 P0 | APIGW-01 webhook CRUD RBAC | be | ✅ 2026-06-22 |
| FA-01 P1 | APIGW-02 webhook Ack/`NakWithDelay` semantics | be | ✅ 2026-06-22 · `webhook_dispatcher_apigw02_test.go` |
| FA-01 P1 | APIGW-03 JetStream FilterSubjects (webhook consumer) | be | ✅ 2026-06-22 · `webhook_dispatcher_apigw03_test.go` |
| FA-01 P1 | AUTH-01 RefreshToken DB session + hash validation | be | ✅ 2026-06-22/23 · `service_session.go` · `service_session_test.go` |
| FA-01 P2 | APIGW-04 deactivated user on API key path | be | ✅ 2026-06-22/23 · `apikey.go` · `apikey_auth_test.go` |
| FA-01 P2 | WebhookDispatcher `Stop` inFlight race fix | be | ✅ 2026-06-22/23 · `webhook.go` (`inFlight.Wait()` before `close(jobs)`) |
| FA-01 CRIT | FA-RBAC-FIBER3 Fiber v3 middleware order (123 routes) | be | ✅ 2026-06-22 · `.planning/audits/2026-06/FA-RBAC-FIBER3-AUDIT.md` |
| FA-02 | Security re-audit all modules | security-auditor + gsd-secure-phase | ✅ 2026-06-22 PASS · 0 CRITICAL · `.planning/audits/2026-06/FA-02-SECURITY-REAUDIT.md` |
| FA-03 | Update `docs/*`, `.planning/STATE.md`, `COVERAGE.md`, `API.md` | gsd-docs-update | ✅ 2026-06-22 |
| FA-04 | Fix github `http.DefaultClient` fallback | be | ✅ 2026-06-22 · `github/service.go:55` safehttp |
| FA-05 | hub-mail in prod compose | devops-engineer | ✅ 2026-06-22 |
| FA-06 | PG-03 secret rotation execute | manual ops | ⏳ **manual ops** (runbook ready) |
| FA-07 | Admin console.error → toastApiError | fe | ✅ 2026-06-22 |
| FA-08 | main.go bootstrap extraction | refactor-specialist | ✅ 2026-06-22 · `wiring_*.go` split |
| FA-09 | E2E 20 specs → full POM+Steps | qa | ✅ 2026-06-22 |
| FA-10 | `/qa-gate` Docker + Playwright + tsc | qa-gate skill | ✅ 2026-06-22 **PASS** (local) · `.planning/audits/2026-06/FA-10-QA-GATE.md` |
| FA-11 | Close tms/project 81.8% → 82% micro-gap | be + qa | ✅ 2026-06-22 · `service_fa11_test.go` |
| FA-12 | Fix TestRail `mapTestRailType` → ent enum `classic`/`bdd` | be | ✅ 2026-06-22 · `tms/testrail_import.go:69` |
| RU-V10 | RU comments/docs final gate | verifier | ✅ 2026-06-22 PASS · `.planning/audits/2026-06/RU-V-REPORT.md` |

### FA-RBAC-FIBER3 — 123 routes fixed (2026-06-22)

| File | Routes |
|------|-------:|
| `internal/tms/handler.go` | 50 (FA-02-02) |
| `internal/tracker/handler.go` | 46 |
| `internal/sprint/handler.go` | 11 |
| `internal/rbac/handler.go` | 6 |
| `internal/github/handler.go` | 5 |
| `internal/gitlab/handler.go` | 5 |
| **Total** | **123** |

Fiber v3: `Post(path, handler, rbacMW)` → chain `[rbacMW, handler]`. Details: `.planning/audits/2026-06/FA-RBAC-FIBER3-AUDIT.md`

### FA-10 qa-gate (2026-06-22/23)

| Step | Result |
|------|--------|
| `go build -tags release ./...` | ✅ PASS |
| `go test -tags hubtest ./testframework/...` | ✅ PASS |
| `go test -tags hubtest -race ./testframework/...` | ✅ PASS (~82s full suite, post WebhookDispatcher Stop fix) |
| `npx tsc --noEmit` | ✅ PASS |
| `npm run lint` (Biome) | ✅ PASS |
| `npx vitest run` | ✅ PASS (99/99) |
| **Verdict** | **PASS** (local) |

**Closed this session:** AUTH-01 (P1) · APIGW-04 (P2) · WebhookDispatcher `Stop` inFlight race fix.

**Not run locally:** Playwright E2E · Docker coverage job. CI: e2e-full **20 specs** · backend-test `./testframework/...` only. Full report: `.planning/audits/2026-06/FA-10-QA-GATE.md`

### FA-02 backlog (post-fix, LOW)

| ID | Severity | Finding | Owner | Status |
|----|----------|---------|-------|--------|
| FA-02-02 | MEDIUM | TMS CI webhook skips RBAC / project membership (IDOR) | be + sa | ✅ 2026-06-22 · FA-RBAC-FIBER3 + APIGW-01 |
| FA-02-03 | LOW | LiveKit webhook `iss` compare uses `!=` (not const-time) | be | ✅ 2026-06-22 · `calls/handler.go` SHA-256 compare |
| FA-02-04 | LOW | `RequireAuth` fail-open on deactivated-user lookup error | be | ✅ 2026-06-22 · 503 fail-closed |

---

## Current Metrics (Docker, 2026-06-22, Wave C verified)

| Metric | Value |
|--------|-------|
| COV overall | **83.0%** (post FA-11/12) |
| notification | **82.5%** (testframework) |
| auth | **82.4%** |
| ws | **87.7%** |
| rbac | **84.7%** |
| audit | **80.5%** |
| tracker | **82.4%** |
| messenger | **82.3%** |
| storage | **86.8%** |
| tms | **82.0%** (FA-11 ✅) |
| project | **82.0%** (FA-11 ✅) |
| Go test files in `backend/testframework/` (total) | **488** (verified) |
| Go test files in `backend/internal/` (remaining) | **0** (verified, TF-M08 ✅) |
| Frontend Vitest `frontend/testframework/unit/` | **37** files · 99/99 |
| Playwright E2E `testframework/e2e/specs/` | **20** specs |

---

## Next Actions (priority)

1. **FA-06 / PG-03** — secret rotation (manual ops) — единственный блокер Wave C
2. **Squash PR** — `feat/v2.5-trust` → main when CTO approves (after FA-06)
3. **CI verify** — Playwright e2e-full (20 specs) + Docker coverage on PR
4. **Backlog:** FA-02 backlog closed (FA-02-03 iss const-time · FA-02-04 RequireAuth fail-closed)

### New tooling
- `scripts/fix-testframework-prefix.ps1` — re-apply `mod.Symbol` prefixing + ent/project alias
- `backend/testframework/messenger/mock_repository_test.go` — local mock (white-box fields)
- `backend/testframework/messenger/stubs_test.go` — `setupChannelWithMembers`, `testMessengerProjectKey`
- `backend/testframework/tracker/handler_rbac_test.go` — FA-RBAC-FIBER3 regression (member vs outsider 403)
- `backend/testframework/sprint/handler_rbac_test.go` — sprint RBAC regression

*Opus 4.8: недоступен в subagent list — работа на текущей модели, Docker verify.*
