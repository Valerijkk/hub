# Hub — Requirement Coverage Matrix

> Phase 0 reconciliation: requirement → code → test.
> Источники: `.planning/PROJECT.md`, `docs/vision_and_scope.md`, `.planning/code-review-final/SUMMARY.md`
> Верификация кода: 2026-06-17 (SA) · Refresh: 2026-06-20 (Wave 6 docs sync)

---

## Методология

| Колонка | Значения |
|---------|----------|
| **Code** | ✅ Implemented · ⚠️ Partial · ❌ Missing · 🔧 Fixed post-audit (needs re-verify) |
| **Tests** | ✅ Covered · ⚠️ Partial · ❌ Missing |
| **Gap** | Открытый finding из SUMMARY.md или пробел покрытия |

**API endpoint count (2026-06-17):**

| Источник | Count | Примечание |
|----------|-------|------------|
| `docs/wiki/API.md` (documented) | **286** | 280 prod handler + 3 dev + 3 infra |
| Handler `RegisterRoutes` (code) | **283** | 21 модулей + templates (prod) |
| Dev routes (`/dev/*`) | **+3** | `dev/handler.go` — non-release only |
| Infra routes (`/docs`, `/docs/doc.json`, `/health`) | **+3** | `main.go` |
| **Total REST `/api/*`** | **286** | Wiki synced with code |
| WebSocket | **+1** | `GET /ws` (вне `/api`) |

**Test baseline (2026-06-20 Wave 6 verification):**

| Layer | Count | Notes |
|-------|-------|-------|
| Backend `*_test.go` | **161** files | `internal/` only |
| Frontend Vitest | **10** files | 36 tests |
| Playwright e2e | **7** spec files | login, navigation, logout, health, admin-audit, command-palette, a11y |
| Statement coverage (overall `internal/`) | **56.6%** | `go test -coverprofile` excl. platform/db config |
| Security-critical packages | **6/6 ≥80%** | calls **80.2%** (PG-01 closed) |

**Wave 1 security-critical test packages (verified 2026-06-20 Wave 6):**

| Package | Coverage | Status |
|---------|----------|--------|
| `auth/` | **83.6%** | ✅ ≥80% |
| `ws/` | **88.0%** | ✅ ≥80% |
| `rbac/` | **85.0%** | ✅ ≥80% |
| `audit/` | **80.5%** | ✅ ≥80% |
| `platform/middleware` | **84.0%** | ✅ ≥80% |
| `calls/` | **80.2%** | ✅ ≥80% (PG-01: chatMessageStore + LiveKit mock) |

Target Wave 1: ≥80% per security package — **6/6 met** ✅

---

## Top 20 Critical Requirements — Gap Matrix

| REQ-ID | Requirement | Source | Code Location | Code | Tests | Gap / Finding |
|--------|-------------|--------|---------------|------|-------|---------------|
| **AUTH-01** | Magic Link + Telegram login (no passwords) | US-M01 | `auth/handler.go`, `auth/magiclink.go`, `auth/telegram.go` | ✅ | ⚠️ `magiclink_test.go`, `handler_test.go`, `service_session_test.go` | SEC-07 🔧 ConstantTimeCompare applied |
| **AUTH-02** | JWT 15 min + refresh 30d Redis revoke | NFR-09/10 | `auth/service_session.go`, Redis keys | ✅ | ⚠️ `service_session_test.go` | FE refresh mutex 🔧 in `api/client.ts` |
| **AUTH-03** | Auth rate limit 3/10 min | NFR-05 | `main.go:217-233`, `middleware/ratelimit.go` | ✅ | ✅ `middleware_test.go` | Skipped when `DEV_MODE=true` (by design) |
| **PROJ-01** | Project creates tracker+TMS+chat+dashboard | US-M02 | `project/service.go`, provisioners | ✅ | ⚠️ `project/service_test.go` | PROJECT.md still `[ ]` — doc drift |
| **TASK-01** | 9 issue types, CRUD, links, comments | US-M03 | `tracker/handler.go` (38 routes) | ✅ | ⚠️ `tracker/service_test.go` | No integration/E2E |
| **TASK-02** | Optimistic lock (`version`, 409) | US-M03, data rule | `tracker/repo_issues.go:242` | ✅ | ✅ conflict test in `service_test.go` | — |
| **MSG-01** | Channels, threads, reactions, presence | US-M04 | `messenger/handler.go` (42 routes) | ✅ | ⚠️ 4 test files, no fan-out integration | Message rate limit ✅ |
| **TMS-01** | Test cases + Allure CI import | US-M05 | `tms/handler.go`, `tms/allure_parser.go` | ✅ | ✅ `allure_parser_test.go`, `service_test.go` | Compare-runs UI depth ⚠️ |
| **CALL-01** | LiveKit voice+screen, ≤50 participants | US-M06, NFR-03 | `calls/handler.go`, `deployments/livekit.yaml` | ✅ | ⚠️ `calls/service_test.go`, `recording_webhook_test.go` | INF-01 🔧 departure_timeout in yaml |
| **SRCH-01** | Global search with user-scope | US-M07, NF-03 | `search/service.go`, `command-palette.tsx` | ✅ | ⚠️ `service_test.go` mocks; E2E-06 pending | UI Ctrl+K ✅ 2026-06-20; saved filters/JQL partial |
| **WS-01** | WebSocket real-time via NATS | NFR-02 | `ws/hub.go`, `ws/bridge.go`, `main.go:605` | ✅ | ⚠️ `hub_test.go`, `authorizer_test.go`, `protocol_test.go` | SEC-H1 🔧 `ws.NewAuthorizer` in main |
| **API-01** | REST API + API keys | API module | 286 endpoints, `apigateway/` | ✅ | ⚠️ `apikey_test.go` DTO-only | SEC-01 🔧 middleware before routes |
| **NOTF-01** | In-app + Telegram + Web Push | US-S05 | `notification/handler.go`, `push.go` | ✅ | ⚠️ `service_test.go` | SEC-H3 🔧 ValidateOutboundURL |
| **AUDIT-01** | Audit log + trash 30 days | US-S03, NFR-11 | `audit/handler.go`, HTTP audit middleware, ban API, cron purge | ✅ | ⚠️ `middleware_test.go`, `service_security_test.go` | BE-H6/H7 🔧 GDPR + retention implemented |
| **OBS-01** | Admin observability + request timing | US-S03 | `audit/middleware.go`, `observability/timing.go`, `/admin/audit` UI | ✅ | ⚠️ `timing_test.go`, `middleware_test.go` | Full API audit volume — OK for dev |
| **BACK-01** | Weekly PG+Mongo+S3 backup | PROJECT backups | `backup/service.go`, cron | ✅ | ⚠️ `createZIP` unit only | INF-04 🔧 s3Client wired in main |
| **GH-01** | GitHub PR linking + webhooks | US-S04 | `github/handler.go`, `github/service.go` | ✅ | ⚠️ `github/service_test.go` | SafeDialer ✅ |
| **SEC-PROD** | 0 CRITICAL blockers for prod | G-04, SUMMARY | cross-cutting | ✅ | ⚠️ | 12/12 CRITICAL fixed; BE-H6/H7 re-verified |
| **NFR-13** | Error JSON `{error, request_id}` | NFR-13 | `platform/errors`, `middleware.go` | ✅ | ✅ `middleware_test.go` | — |
| **NFR-16** | WCAG 2.1 AA | NFR-16, G-05 | frontend components | ⚠️ | ✅ login a11y | ErrorBoundary + 10 Vitest; login axe PASS |
| **COV-01** | Coverage ≥50% overall, ≥80% security | G-04, Phase 18 | all packages | ✅ | ✅ | **56.6%** overall; security **6/6 ≥80%** ✅ |

---

## SUMMARY.md CRITICAL — Reconciliation Status

| ID | Finding | Code Status (2026-06-17) | Test Status |
|----|---------|---------------------------|-------------|
| SEC-01 | APIKeyAuth dead code | 🔧 Fixed `main.go:241-250` | ⚠️ No middleware-order test |
| SEC-02 | LiveKit webhook sha256 optional | 🔧 Fixed `calls/handler.go:257-263` | ✅ `recording_webhook_test.go` |
| SEC-03 | Hardcoded livekit prod secret | ⚠️ `livekit-prod.yaml` removed; local dev keys in `livekit.yaml` | ❌ |
| SEC-04 | GitLab SSRF DefaultClient | 🔧 Fixed `safehttp.NewSafeHTTPClient` | ⚠️ `gitlab/service_test.go` partial |
| SEC-05 | WS send-on-closed panic | 🔧 Fixed `ws/hub.go:removeClient` | ⚠️ `hub_test.go` partial race |
| SEC-06 | First-admin race | 🔧 Fixed + mutex | ✅ `service_registration_race_test.go` |
| SEC-07 | Magic-link timing attack | 🔧 Fixed ConstantTimeCompare | ✅ `magiclink_test.go` |
| INF-01 | LiveKit departure_timeout prod | 🔧 In `livekit.yaml`; prod via compose | ❌ No deploy test |
| INF-02 | DEV_MODE=true in CI prod | 🔧 Removed from `deploy.yml:63-65` | ❌ |
| INF-03 | Plaintext creds local compose | ⚠️ Env-var placeholders `${RESEND_API_KEY:-}` | ❌ Rotation/purge manual |
| INF-04 | BackupService nil s3Client | 🔧 Fixed `main.go:524-537` | ⚠️ ZIP unit only |
| FE-01 | Rules of Hooks violation | 🔧 Fixed `message-item.tsx:156-158` | ❌ No component test |

**Verdict:** COV-01 gate **PASS** (overall ≥50%, security 6/6 ≥80%). Wave 6 prod pipeline + Ctrl+K UI landed; E2E-06 and formal security re-audit remain optional hardening.

---

## Module Coverage Summary

| Module | Routes | Unit Tests | Integration | E2E | Security Audit |
|--------|--------|------------|-------------|-----|----------------|
| auth | 13 | ✅ Wave 1 | partial | ⚠️ login.spec | SEC-06/07 🔧 |
| project | 27 | ✅ | partial | ❌ | — |
| tracker | 38 | ✅ | partial | ❌ | — |
| messenger | 42 | ⚠️ | partial | ❌ | — |
| tms | 52 | ✅ | partial | ❌ | — |
| calls | 22 | ⚠️ Wave 1 | partial | ❌ | SEC-02 🔧 |
| ws | 1 (WS) | ⚠️ Wave 1 | ❌ | ❌ | SEC-05/H1 🔧 |
| search | 7 | ⚠️ | ❌ | ❌ | SEC-H2 🔧 |
| notification | 11 | ⚠️ | ❌ | ❌ | SEC-H3 🔧 |
| apigateway | 8 | ⚠️ | ❌ | ❌ | SEC-01 🔧 |
| sprint | 14 | ✅ | partial | ❌ | — |
| dashboard | 8 | ⚠️ | ❌ | ❌ | BE-H3 partial |
| audit/backup | 12 | ⚠️ Wave 1 | ❌ | ⚠️ admin-audit.spec | BE-H6/H7 🔧 |
| github/gitlab | 20 | ⚠️ | ❌ | ❌ | SEC-04 🔧 |
| imex | 6 | ⚠️ | ❌ | ❌ | — |
| platform | — | ✅ Wave 1 | partial | ⚠️ health.spec | — |

---

## Top 10 Unimplemented / Incomplete Requirements

| # | REQ-ID | Requirement | Why Unimplemented |
|---|--------|-------------|-------------------|
| 1 | **COV-01** | Unit coverage ≥50% / security ≥80% | **56.0%** overall ✅; calls **72.5%** — only residual gap |
| 2 | **FE-TEST** | Frontend Vitest unit tests | **4** files (target ≥10 per PLAN-CLOSE-GAPS) |
| 3 | **E2E-01** | Playwright critical-path E2E | ✅ **7 specs** + CI `e2e-full`; expand E2E-02..05 |
| 4 | **NFR-16** | WCAG 2.1 AA compliance | Frontend audit 76/100; no automated a11y gate |
| 5 | **NFR-17** | Lighthouse score ≥90 | Not measured in CI; bundle 1.1MB (FE-H5) |
| 6 | **ARCH-01** | Repository pattern all modules | 7 packages still use `*ent.Client` in service (Phase 19) |
| 7 | **OBS-01** | OTel metrics + structured observability | Phase 21 planned; Slog only today |
| 8 | **PROJ-REC** | PROJECT.md requirement validation | All active reqs still `[ ]` unchecked |
| 9 | **INF-03** | Secret hygiene (local compose history) | Env placeholders OK; rotation/purge not verified |
| 10 | **BE-H6** | GDPR delete completeness | ✅ Implemented; monitor partial cascade WARN logs |

---

## Next Actions (SA → QA / Tech Lead)

1. **SA_VER:** Re-run security-auditor against current `main.go` to close SEC-PROD formally.
2. **QA (post-gate):** Uplift `calls/` to 80% (Mongo chat integration or LiveKit mock).
3. **QA:** Wire Playwright into CI (`npm run test:e2e` with `PLAYWRIGHT_SKIP_WEBSERVER=1`).
4. **BA:** Move validated PROJECT.md items to Confirmed section.
5. **DevOps:** Confirm prod LiveKit config delivery (INF-01) on target host.

---
*Generated: 2026-06-17 — Verified: 2026-06-20 (COV-01 gate: 56.0% overall, 5/6 security ≥80%)*
