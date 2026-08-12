# Founder greenfield cutover — keep native units, rebuild the rest in Python

**Status:** VERIFIED / HOOK-BLOCKED FOR 245 ROWS — no tests moved or deleted  
**Date:** 2026-08-07  
**Machine manifest:** `founder-greenfield-cutover-manifest.json`

## Outcome

The founder decision materially simplifies the test rebase:

- retain only real service-free Go units and genuine Vitest/RTL units;
- relocate them to `backend/tests/<domain>/**` and `frontend/tests/<domain>/**`;
- delete every existing component/API/integration/static-contract/TypeScript
  Playwright test and its closed legacy runner/support chain;
- do **not** translate those deleted tests into Python;
- design the new root `testframework/**` suites from current requirements,
  product risks, tasks and test cases, with happy/negative/edge/corner coverage.

This document is deliberately fail-closed. The current TASK-202 ledger is a
classification input, not migration authority: its SHA is pinned, its live
drift is exposed, and any unknown row blocks destructive execution.

## Exact ledger partition

Pinned input: 1,049 rows / 4,246 declared cases, ledger SHA-256
`72de6a81b2bd5da8873a98e784a5069a661e9caea083f10722aa85e9f91b4511`,
source revision `5708671f00411c7c9e9dbee88b725de2092cbacd`.

| Action | Rows | Declared cases | Decision |
|---|---:|---:|---|
| Move | **507** | **1,673** | 350 hook-free Go units + 156 frontend units + `render.tsx` |
| Keep blocked | **245** | **1,428** | Go unit-class rows that still consume private export hooks |
| Delete, do not port | **297** | **1,145** | 237 Go component + 24 static Vitest + 2 compile fixtures + 20 TS E2E + 14 TS E2E support |
| **Total** | **1,049** | **4,246** | exact, mutually exclusive reconciliation |

The raw TASK-202 label `go-unit` is not sufficient for immediate movement.
Among 595 such rows, 245 consume 434 private symbols through 653 links across
42 `export_hooks*.go` files. Moving them unchanged would preserve a hidden
white-box test API and violate the isolated-unit rule. Those rows stay where
they are until a public/injected seam makes `hook_consumers=[]`; then the
ledger is regenerated and they enter the move partition. They are not deleted
silently.

Frontend support was also tightened rather than trusted blindly:

- `render.tsx` has 62 observed consumers: 58 moving real units and four
  deleting static-contract tests, so it moves to `frontend/tests/support/`;
- the two `*.fixture.ts` rows have no runtime test consumer and are referenced
  only by two task-specific compile `tsconfig` files, so all four files delete
  together.

## What is deleted

The row-level delete set is derived exactly from the pinned ledger selectors:

- Go `python-api`: 31 files / 125 cases;
- Go `python-integration`: 206 files / 857 cases;
- frontend `python-static-contract`: 24 files / 110 cases;
- two compile-only frontend fixtures;
- Playwright TypeScript: 20 specs / 53 cases;
- Playwright TypeScript support/config: 14 ledger rows.

The machine manifest additionally lists 28 exact non-ledger paths for removal:
four backend component-support helpers, old Playwright/link runners, seven old
migration/E2E scripts, six pipeline YAML manifests, two compile `tsconfig`
files, `MIGRATION.md` and five obsolete `stages/**` READMEs. `package.json`,
`vite.config.ts`, lockfile, GitHub/GitLab CI and current docs are edited in
place; they are not deleted.

The four backend support helpers are a closed deletion set relative to moving
units: `env` has two legacy component consumers, `factory` one, `mongotest`
eleven and `steps` zero; **none** is imported by the 350 immediately movable
Go units.

## Hooks: no bulk deletion

TASK-202 records 43 hook files, 582 symbols, 550 consumed symbols and 892
consumer links. Deleting the component tests removes 239 links, but 653 links
remain from the 245 blocked unit rows. Therefore:

1. no hook is bulk-deleted with component tests;
2. `backend/internal/tms/export_hooks_mongo.go` is the only hook file currently
   consumed solely by component rows and is eligible only after a regenerated
   reverse-link inventory proves zero consumers;
3. the other 42 hook files remain until every blocked unit is refactored or
   explicitly reclassified;
4. hook cleanup gets its own atomic commits and zero-consumer evidence.

## New Python tests are not ports

Every new test is specified from live product truth, not from a legacy test
body. Its metadata must include:

- stable `TC-*` id;
- `TASK-*` or accepted backlog link;
- FR/NFR, user story, API, event or schema reference;
- domain owner, priority and risk;
- scenario category and expected artifact.

The design checklist covers happy path, negative, edge, corner, authorization,
validation, idempotency, concurrency, timeouts/retries, dependency failures,
security, accessibility and observability. A category may be absent only with
a documented risk-based N/A reason. Missing infrastructure is a failure, not a
skip. Bare skips and live external egress are forbidden.

Runtime code under root `testframework/**` is Python-only. Dockerfiles,
YAML/TOML/JSON, Markdown and sanitized report data remain declarative. All
test, validator, lint and build commands execute only in task-owned Docker
containers with CPU/RAM limits; unused containers, networks, volumes and
temporary images are removed after ownership inspection.

## Atomic execution and rollback

Execution order is designed so each commit has one reversible purpose:

1. amend and accept the ADR/machine contract/active plan for the founder rule;
2. move hook-free backend units per domain with Docker evidence;
3. move frontend units per domain, then move shared render support once;
4. delete the exact Go component/support closure without touching blocked
   units or hooks;
5. delete frontend static/TS E2E and runner/config closure while updating
   package, Vite, CI and docs in the same commit;
6. build new root Python slices from current requirements and risks;
7. refactor hook-dependent units, move them, then prune zero-consumer hooks.

Rollback is `git revert` in reverse order, one exact atomic commit at a time.
Ad-hoc copies, `git reset --hard`, checkout over dirty work, broad recursive
deletion and cleanup of shared Docker resources are forbidden.

## Ledger regeneration evidence

TASK-202 was regenerated by inventory tool `2.2.1` on exact HEAD
`5708671f00411c7c9e9dbee88b725de2092cbacd` inside a task-owned Docker
container with no network, 1 CPU, 1 GiB RAM and `--rm`. The two new test
sources are now explicit rows: `admin_users_test.go` is a hook-dependent Go
unit (five cases), while `telegram_availability_test.go` is a deleting
integration row (two cases).

The subsequent read-only Docker run passed both canonical gates:

- `inventory.py --check`: GREEN, SHA
  `72de6a81b2bd5da8873a98e784a5069a661e9caea083f10722aa85e9f91b4511`,
  1,049 rows;
- `inventory.py --self-test`: GREEN; clean inventory/schema sentinels passed,
  while duplicate, removal, misroute and missing-owner sentinels produced RED.

The prior eight source hash mismatches are incorporated and live drift is now
exactly zero. B1 ledger drift is resolved. The ADR is now `Accepted`, the
executable contract uses `greenfield-rebuild` with parity/dual-run disabled,
and the active plan has been rebased; the former B2 and B4 blockers are also
resolved. The Docker generation/check containers were removed automatically.

## Remaining scoped blocker

1. **Hook-dependent units:** 245 rows are keep/block, not move, until their
   private hook dependency is removed or the founder explicitly reclassifies
   them.

No production code, test source, CI file, hook or protected `pro-qa-ai/**`
content was changed by this audit.
