# Phase 28 Verification — v2.6 Architecture Milestone Gate

**Date:** 2026-07-27  
**Auditor:** Grok 4.5 (PW2 retro-verify)  
**Mode:** documentation-only · measured against current HEAD  
**Verdict:** PARTIAL

Architecture exit criteria are met. Full `./testframework/...` suite is **not** fully green (1 failing package).

## Goal-backward

| Criterion | Expected | Actual (2026-07-27) | Status |
|-----------|----------|---------------------|--------|
| ARCH-01 verified | former holdouts on Repository | **7/7** packages from ARCHITECTURE.md §6 “without Repository” now expose Repository interfaces; services in those packages have **0** non-repository `*ent.Client` field leaks | PASS |
| No `*ent.Client` in rbac middleware | Repository injection | `RequireAdmin(repo Repository)`; `RequireSystemRole(role, repo Repository)`; middleware uses `ent.IsNotFound` only (error helper), not client field | PASS |
| Full backend regression green | `go test` hubtest suite | `./testframework/...` → **ok=41**, **fail=1**, **?=3** | FAIL |
| Milestone docs closed | STATE/ROADMAP mark | ROADMAP.md phases 21–28 checked `[x]` with ARCH-01 17/17 note on phase 27 | PASS (docs present) |

### Failing package (measured)

```text
FAIL  github.com/valerii/hub/backend/testframework/platform/cron  (4.559s)
--- FAIL: TestSchedule_RunsFnOnTick (1.00s)
```

Known timing-sensitive flake (also noted in `.planning/audits/grok-2026-07/FIX-W3.md` baseline). Unrelated to Repository migration.

### Domain packages for this wave (subset, all green)

| Package | Result |
|---------|--------|
| storage | ok 5.853s |
| apigateway | ok 48.494s |
| imex | ok 1.924s |
| dashboard | ok 2.051s |
| rbac | ok 2.367s |
| auth | ok 3.308s |

## Commands run

```text
# remaining-7 Repository interfaces
# auth dashboard backup imex storage apigateway rbac → 7/7

Select-String backend/internal/rbac/middleware.go 'func RequireAdmin'

cd backend
$env:POSTGRES_TEST_URL='postgresql://hub:hub@localhost:5432/hub?sslmode=disable'
$env:MONGODB_TEST_URL='mongodb://hub:hub@localhost:27017/?authSource=admin'
go test -tags hubtest -count=1 -timeout=900s ./testframework/...
# ok=41 fail=1 ?=3
```

## Residual

- Re-run / harden `TestSchedule_RunsFnOnTick` before claiming full-suite green for this gate.
- Do **not** treat PARTIAL as deploy blocker for ARCH-01 itself — structural goal is met.
