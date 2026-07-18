# Production Handoff — v2.5 Trust (`feat/v2.5-trust`)

**Date:** 2026-06-21  
**Gate:** Phase 86 `/qa-gate` ✅ · COV **69.7%** (Phases 76–86)  
**Branch:** `feat/v2.5-trust` (uncommitted — **requires user-approved commit + PR**)

---

## Executive summary

Hub v2.5 Trust wave (Phases 59–73) is **code-complete and gate-ready**:

- Backend: release build + race tests green; **69.7%** statement coverage (target ≥50%)
- Security-critical packages: **6/6 ≥80%**
- Frontend: tsc + Biome 0 errors; **99 Vitest** tests
- CI: backend-test, frontend-test, e2e-full (16 Playwright specs)

**Blockers before prod deploy:**

1. Squash commit + PR merge to `main`
2. PG-03 secret rotation if credentials ever touched git history (`.planning/production/PG-03-SECRET-ROTATION.md`)

---

## What shipped (Phases 59–73)

| Phase | Deliverable |
|-------|-------------|
| 59 | tracker/storage coverage, issue-form Vitest |
| 60 | a11y E2E, shortcuts CI, onboarding step 4 Zod |
| 61 | project/search coverage, context-help tooltips (UX-P1-04) |
| 62 | onboarding profile RHF+Zod (UX-P1-07) |
| 63 | first production gate, schema route hygiene |
| 64 | auth **81.9%** (org/pending/Telegram facade tests) |
| 65 | UX-P1-05 rate-limit countdown toast; messenger **52.3%** |
| 66 | notification **55.7%**; telegram route fix (`DELETE /telegram`) |
| 67 | re-gate + PRODUCTION-HANDOFF |
| 68 | messenger **62.1%**, fan-out unread integration |
| 69 | apigateway **67.0%**, COV baseline **58.1%** |
| 70 | gitlab **68.5%**, github **67.8%**, COV **59.2%** |
| 71 | mail **86.0%**, dashboard **73.7%**, COV **59.8%** |
| 72 | full `/qa-gate` PASS (race ~48s, Vitest 99/99) |
| 73 | platform/http **87.6%**, outbox **74.3%** |
| 74 | ai **50.7%** (21 test files, race PASS) |
| 75 | full `/qa-gate` PASS · COV **62.2%** |
| 76 | sprint **79.8%** EntRepository + RegisterRoutes |
| 77 | storage **79.6%** CR-04 RBAC presign routes |
| 78 | tracker **63.2%** · tms **61.0%** RegisterRoutes + ent |
| 79 | notification **63.2%** · project **71.7%** templates + cross-project links |
| 80 | backup **65.7%** · hr **73.3%** · imex **77.7%** PDF export + repo paths |
| 81 | ai **68.5%** adapter HTTP mocks + `/qa-gate` PASS · COV **67.6%** |
| 82 | ai **75.0%** @hub-ai bot hook + doc search RAG · COV **68.3%** |
| 83 | ai **76.4%** slash commands + search query · COV **68.4%** |
| 84 | bases **71.5%** · gitlab **79.1%** · intranet coffee · COV **69.1%** |
| 85 | github **72.6%** · AI GitHub `/ai review` · COV **69.3%** |
| 86 | search **66.1%** · tracker **65.1%** · COV **69.7%** |

---

## Quality metrics (verified 2026-06-21)

| Layer | Metric | Value |
|-------|--------|-------|
| Backend | `go test -race ./internal/...` | PASS (~47s) |
| Backend | Overall coverage | **64.5%** |
| Backend | tracker / tms | **63.2%** / **61.0%** |
| Backend | platform/http · outbox | **87.6%** · **74.3%** |
| Backend | mail / dashboard | **86.0%** / **73.7%** |
| Backend | apigateway | **67.0%** |
| Backend | messenger | **62.1%** |
| Backend | auth / ws / rbac / audit / calls | 81.9% / 88.0% / 84.7% / 80.5% / 80.2% |
| Frontend | Vitest | **99/99** |
| Frontend | Biome | 0 errors (396 files) |
| E2E | CI e2e-full | 16 specs |
| Infra | `DEV_MODE` in deploy.yml | absent ✅ |

---

## Pre-merge checklist (developer)

- [ ] Review diff on `feat/v2.5-trust` (~large uncommitted block)
- [ ] Squash commit with conventional message, e.g. `feat(trust): v2.5 coverage, UX gate, notification fixes`
- [ ] Open PR → `main`; ensure CI green
- [ ] Code-reviewer approval (per `.cursor/rules/50-git.mdc`)

---

## Pre-deploy checklist (ops)

1. [ ] Merge PR to `main`
2. [ ] Run PG-03 if applicable (JWT, Telegram, Resend, LiveKit, DB passwords)
3. [ ] Verify K3s secrets from templates — **no** hardcoded `devkey: secret`
4. [ ] Deploy via GitHub Actions `deploy.yml`
5. [ ] Smoke test:
   - Magic Link login
   - Create issue + Kanban move
   - Messenger send + reaction
   - TMS run import
   - Voice call join (LiveKit)
   - Admin audit log visible
   - Notification Telegram bind (if bot configured)
6. [ ] Monitor: Loki alerts, `/admin/metrics`, slow audit filter

---

## Known non-blocking gaps (post v2.5)

| Gap | Status | Plan |
|-----|--------|------|
| messenger fan-out integration | ✅ Phase 68 | 62.1% |
| PG-03 rotation | manual | ops runbook |
| Secret purge from git history | manual | BFG/git-filter-repo |
| v3.0 AI Chat Hub | future | `.planning/roadmaps/MILESTONE-3.md` |

---

## Rollback

- K3s: redeploy previous image tag from registry
- DB: Atlas migrations are forward-only; restore from backup if schema rollback needed (`backup` module)
- Feature flags: none required for this release

---

## Contacts & docs

| Doc | Purpose |
|-----|---------|
| `docs/security_guidelines.md` | OWASP, blockers |
| `docs/wiki/API.md` | REST reference |
| `.planning/COVERAGE.md` | requirement → test matrix |
| `.planning/production/PG-03-SECRET-ROTATION.md` | secret rotation |
| `.planning/phases/67-production-gate/VERIFICATION.md` | gate evidence |

---

*Generated by Phase 67 GSD executor — no commit performed (awaiting user approval).*
