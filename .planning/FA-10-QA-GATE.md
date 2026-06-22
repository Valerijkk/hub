# FA-10 — QA Gate Report

**Date:** 2026-06-22 (local run)  
**Run at:** 2026-06-22T02:56 local · race rerun below  
**Branch:** `feat/v2.5-trust`

---

## Backend (release + hubtest)

| Step | Command | Result | Duration |
|------|---------|--------|----------|
| Build | `go build -tags release ./...` | **PASS** | ~10s |
| Tests | `go test -tags hubtest -count=1 ./testframework/...` | **PASS** (44 packages ok) | ~57s |
| Race | `go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...` | **PASS** (44 packages ok) | ~67s @ 2026-06-22T03:02 local |

---

## Frontend

| Step | Command | Result | Duration |
|------|---------|--------|----------|
| Typecheck | `npx tsc --noEmit` | **PASS** | ~33s |
| Lint | `npm run lint` (Biome) | **PASS** (359 files) | ~2s |
| Vitest | `npx vitest run` | **PASS** (99/99, 37 files) | ~14s |

---

## E2E smoke (Playwright)

| Step | Command | Result | Duration |
|------|---------|--------|----------|
| Link modules | `node scripts/link-e2e-node-modules.mjs` | **PASS** | — |
| Smoke | `npx playwright test -c playwright.config.ts login.spec.ts a11y.spec.ts --grep "login page\|full page has no critical"` | **PASS** (3/3) | ~8.5s |

**Specs passed:**
- `login.spec.ts` — login page loads with Hub branding
- `a11y.spec.ts` — login form panel has no critical or serious axe violations
- `a11y.spec.ts` — full page has no critical or serious axe violations

**Note:** Vite proxy warnings for `/api/dev/status` (backend not running locally); smoke tests still green.

---

## Not run locally (this session)

| Step | Reason |
|------|--------|
| Docker coverage-gate | CI-only (≥82% on master push) |
| Playwright e2e-full | requires Docker stack |

---

## Verdict: **PASS** (full FA-10 local gate)

All requested steps green: backend build + hubtest + race, frontend tsc/lint/vitest, Playwright e2e smoke 3/3.

---

*No git commits. No code fixes.*
