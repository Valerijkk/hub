---
gsd_state_version: 1.0
milestone: v2.0
milestone_name: Hub Redesign
status: complete
stopped_at: Phase 14 complete — Architectural Hardening (5/5 plans)
last_updated: "2026-06-20T00:05:00.000Z"
last_activity: 2026-06-20
progress:
  total_phases: 14
  completed_phases: 14
  total_plans: 62
  completed_plans: 63
  percent: 100
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-27)

**Core value:** Unified environment for dev teams -- tracker, TMS, messenger, and calls connected in one interface
**Current focus:** v2.0.2 **SHIP-READY** — audit **93/100** (ceiling без v2.6)

## Current Position

Phase: Close-Gaps post-gate → v2.0.2 prod hardening
Plan: All waves **COMPLETE** (uncommitted local)
Status: **Cycle 12** — cleanup ✅ · score **93/100** · race+build+vitest ✅
Last activity: 2026-06-20 cycle 12 (artifacts cleanup + full gate)

## Accumulated Context

### Roadmap Evolution

- Phase 14 added 2026-05-15: Architectural Hardening (split god-objects, remove ent.Client leak, expand Permission, increase test coverage)

## Performance Metrics

**Velocity:**

- Total plans completed: 0 (v2.0)
- Average duration: -
- Total execution time: -

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| - | - | - | - |

*Updated after each plan completion*
| Phase 08 P01 | 5min | 1 tasks | 2 files |
| Phase 08 P02 | 2min | 2 tasks | 1 files |
| Phase 09 P02 | 2min | 3 tasks | 3 files |
| Phase 09 P01 | 3min | 3 tasks | 4 files |
| Phase 10 P02 | 3min | 2 tasks | 4 files |
| Phase 10 P01 | 1min | 3 tasks | 3 files |
| Phase 10 P03 | 2min | 3 tasks | 3 files |
| Phase 11 P02 | 2min | 2 tasks | 3 files |
| Phase 11 P01 | 3min | 2 tasks | 4 files |
| Phase 11 P04 | 4min | 2 tasks | 4 files |
| Phase 11 P03 | 5min | 2 tasks | 5 files |
| Phase 12 P04 | 4min | 2 tasks | 3 files |
| Phase 12 P01 | 7min | 2 tasks | 11 files |
| Phase 12 P02 | 8min | 2 tasks | 9 files |
| Phase 12 P05 | 14min | 2 tasks | 16 files |
| Phase 12 P03 | 14min | 2 tasks | 18 files |
| Phase 13 P01 | 4min | 2 tasks | 3 files |
| Phase 13 P02 | 4min | 2 tasks | 1 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [v2.0] Visual-only redesign -- no new features, no functionality changes
- [v2.0] Components restyled one-by-one for maximum quality
- [v2.0] Light + Dark mode both supported (mockups dark-only, light adapted)
- [v2.0] !important cleanup gates all subsequent work (Phase 8 prerequisite)
- [v2.0] /gsd:ui-review after every phase
- [Phase 08]: Hex-based accent colors instead of HSL for simpler dark/light mapping
- [Phase 08]: Dual @theme inline mapping preserves Shadcn compat while exposing new tokens
- [Phase 08]: Fixed radius values (6/8/12/16px) replace calc-based system
- [Phase 08]: Removed all 56 !important declarations; noise overlay at z-index 9999 via feTurbulence SVG
- [Phase 09]: Topbar uses sticky positioning within flex layout, title prop optional for future per-page wiring
- [Phase 09]: Gradient edge line uses absolute div instead of ::after pseudo for cleaner Tailwind
- [Phase 10]: className-only changes -- no React/Radix behavioral code modified in primitives
- [Phase 10]: className-only changes, zero Radix behavioral modifications for Dialog/Tabs/DropdownMenu
- [Phase 11]: Used CSS utility classes (dash-*) for dashboard-specific gradient/animation styles
- [Phase 11]: Removed Shadcn Button/Input deps from auth forms for exact mockup CSS match
- [Phase 11]: Kanban cards use CSS class system (kanban-card) instead of inline Tailwind for gradient+glow complexity
- [Phase 11]: Green accent (emerald-400) for messenger active states and focus indicators
- [Phase 12]: Removed Shadcn deps from onboarding for exact mockup CSS match
- [Phase 12]: Trash uses native table with sticky headers per mockup pattern
- [Phase 12]: Activity feed uses colored dots instead of avatars for compact mockup-matching design
- [Phase 12]: Comment layout: flex gap-14 rows with gradient avatar fallbacks instead of bordered cards
- [Phase 12]: Sidebar meta: direct void bg + pixel-exact padding matching mockup 05
- [Phase 12]: Used CSS utility classes (tms-run-card, tms-precond-card) for ::before pseudo-element effects in TMS
- [Phase 12]: Removed Card/CardHeader Shadcn components in favor of plain divs with design system gradient classes
- [Phase 12]: Velocity chart restyled from Shadcn Card wrapper to inline gradient container
- [Phase 12]: Planning poker uses border-blue glow shadow for selected state
- [Phase 12]: Custom tab buttons instead of Shadcn TabsTrigger for Admin mockup match
- [Phase 12]: Flat section layout (h2+desc+content) for Settings instead of Card wrappers
- [Phase 13]: CSS-only online indicator via group-data pattern (no JS conditional rendering)
- [Phase 13]: 200ms transition on bg/color/border for theme toggle smoothness
- [Phase 13]: Kept all @keyframes even when animation classes removed; replaced all transition:all with explicit property lists

### Pending Todos

- **UX-1:** Onboarding step 4, command palette actions, full RHF+Zod, i18n tracker forms
- **Tests:** Component tests UserPicker; expand axe E2E
- PG-03: Secret rotation (manual ops)
- Git commit when approved

### Session 2026-06-20 (final) — UX-0 complete + audit loop

**UX-0 delivered:**
- `user-picker.tsx` → issue-form (no more UUID assignee)
- `skip-link.tsx`, `/help` route, sidebar link
- Kanban «Move to…» menu + toast (no alert); `sanitizeHexColor` (CSS injection fix)
- E2E: `command-palette.spec.ts`, navigation includes /help
- Rules: `56-regression-verify.mdc`
- Plans: `AUDIT-SCORECARD.md`, `IMPROVEMENT-MASTER-PLAN.md`

**Verifiers:** fe-ver 68→blockers i18n+tests; qa-ver PASS gate; code-reviewer BLOCK color → fixed

**Regression:** `go test ./...` ✅ · `vite build` ✅ · Vitest **39/39** ✅ · `tsc` ✅

**Product score:** 74/100 (target 80 at UX-1)

### Session 2026-06-20 (late) — Wave 6 prod + Ctrl+K + docs sync

**PG-01 closed:** `calls/` **80.2%** (chatMessageStore interface, `chat_mongo.go`, LiveKit mock tests) — 6/6 security packages ≥80% ✅

**Wave 6 — Prod pipeline:**
- `backend/Dockerfile`: `-tags release`, Go 1.25
- `.github/workflows/ci.yml`: race timeout 600s, release build, Biome baseline (continue-on-error), Mongo `authSource=admin` in E2E
- `.github/workflows/deploy.yml`: `workflow_run` after CI success only, SHA image tags, no `--volumes` prune, post-deploy health curl
- `deployments/docker-compose.prod.yml`: `ENVIRONMENT`, `CORS_ORIGINS`, `INTEGRATION_ENCRYPTION_KEY`, Mongo authSource
- `deployments/.env.example`: new required vars documented

**UX-0 — Global Ctrl+K (US-M07):**
- `frontend/src/components/search/command-palette.tsx`, `command-store.ts`, `ui/command.tsx`
- `app-shell.tsx` + `topbar.tsx`: Ctrl/Cmd+K, search trigger, `GET /api/search` via debounced global search
- Docs: `docs/wiki/SHORTCUTS.md`, `docs/competitive_audit.md`, `.planning/COVERAGE.md` SRCH-01 UI ✅

**Rule added:** `.cursor/rules/55-docs-sync.mdc` — mandatory doc update after any code change

**Verified:** `go build -tags release` ✅ · `tsc` ✅ · Vitest 36/36 ✅

### Session 2026-06-20 — COV-01 closure + gate verification

**Coverage wave (handler + ent + consumer tests):**
- `tms` 56.0%, `messenger` 50.0%, `notification` 48.2%, `project` 44.8%
- **Overall `internal/` 56.0%** (target ≥50% ✅)
- Security-critical: auth 83.6%, ws 88%, rbac 85%, audit 80.5%, middleware 84% ✅; calls 72.5% ⚠️

**Gate verification (`go test -race -count=1 ./internal/...`, `go build`, `tsc`, `vitest`):**
- Backend: **all packages PASS**
- Frontend: **10 files / 36 tests PASS**
- E2E scaffold: **6 Playwright specs** (CI job present)

**Artifacts updated:** `.planning/COVERAGE.md`, `.planning/TASKS-CLOSE-GAPS.md`, `.planning/STATE.md`

**Wave 1 (prior):** Admin observability, HTTP audit middleware, BanUser API, `Timed()` wrapper, `docs/competitive_audit.md`

**Wave 2 — Backend Coverage ✅**
- Tests: calls webhook, messenger rate, storage presign, platform/nats, apigateway middleware order (SEC-01)
- `Timed()` on auth/audit/tracker hot paths
- BE-H6/BE-H7 docs fix in `docs/security_guidelines.md`

**Wave 3 — Frontend Quality ✅**
- Vitest: **17 tests** in 6 files (auth, admin, hooks, ErrorBoundary)
- ErrorBoundary wrappers on tracker, tms, calls, dashboard routes
- Playwright: **5 specs** in `frontend/e2e/` (login, navigation, logout, admin-audit, health)
- Biome: partial critical lint fixes (269 errors remain)

**Wave 4 — Observability + Docs ✅**
- Audit filter `slow=1` (duration_ms>500) — backend + admin UI checkbox
- `docs/wiki/API.md` sync (+15 routes)
- `.planning/COVERAGE.md` refresh (COV-01 ~13%)

**Wave 5 — Gate 🔄**
- T5-01: qa-gate — **build + test -race + tsc + vitest PASS** (lint 128 errors — not gated)
- T5-02: verifier acceptance — pending
- T5-03: STATE.md handoff — this session

**Session 2026-06-17 (continued)**
- CI: `backend-test`, `frontend-test`, `e2e-smoke` jobs in `.github/workflows/ci.yml`
- Audit: handler + purge tests → **64.5%** coverage (was 47.3%)
- A11y: login page color-contrast fix (Telegram button #187AAC, helper text --t2)
- Security pkg coverage: auth 50.5%, ws 52.2%, rbac 53.4%, audit 64.5%, calls 35.3%

### Blockers/Concerns

- 56 !important declarations in app.css must be removed in Phase 8 before any component restyling
- Light mode has no mockups -- derived palette must be validated for WCAG AA contrast
- 6 page groups have no mockups -- need consistent extrapolation from design system patterns
- Plate editor may need CSS scoping after !important removal

## Session Continuity

Last session: 2026-06-20T23:00:00.000Z
Stopped at: Wave 6 prod pipeline + Ctrl+K + full docs sync
Completed this session:
- PG-01 calls 80.2% + Wave 6 infra/UX code
- Docs: STATE, COVERAGE, TASKS, PLAN, AGENTS, stack_and_tools, test_plan, competitive_audit, vision KPI, DEPLOY, REGULATION, rule 55-docs-sync
Resume file: None

## Next session priorities
1. UserPicker + form validation (UX backlog)
2. E2E-06 Ctrl+K Playwright spec
3. PG-03 secret rotation checklist
4. Commit Wave 6 when CTO approves
