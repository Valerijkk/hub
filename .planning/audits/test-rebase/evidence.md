# TASK-202 evidence — 2026-08-07

## Reproduction identity

- Git revision: `f7fbdd7ebb23bfaba2aee71d503a7c0280846023`
- Inventory tool version: `1.1.0`
- Interpreter used: bundled CPython `3.12.13`
- Generated ledger: `inventory.json`
- Ledger SHA-256: `17019e9fa0f6ff43b341c12c84b2677b476a4bafcac564e0ddd40f5def62668c`
- Ledger rows: `1028` (`828` Go + `180` Vitest + `20` Playwright)

## Commands and results

Run from the repository root:

```powershell
& 'C:\Users\Valerii\.cache\codex-runtimes\codex-primary-runtime\dependencies\python\python.exe' `
  .planning/audits/test-rebase/inventory.py --generate
# WROTE .../inventory.json sha256=17019e... rows=1028

& 'C:\Users\Valerii\.cache\codex-runtimes\codex-primary-runtime\dependencies\python\python.exe' `
  .planning/audits/test-rebase/inventory.py --check
# GREEN: ledger matches allow-listed inputs; sha256=17019e... rows=1028

& 'C:\Users\Valerii\.cache\codex-runtimes\codex-primary-runtime\dependencies\python\python.exe' `
  .planning/audits/test-rebase/inventory.py --self-test
# GREEN: clean inventory/schema sentinels passed; duplicate, removal,
# and missing-owner sentinels produced RED

powershell -NoProfile -ExecutionPolicy Bypass -File `
  .planning/audits/test-rebase/validate-ledger.ps1
# Ledger validator: .../dependencies/python/python.exe
# GREEN: ledger matches allow-listed inputs; sha256=17019e... rows=1028
# GREEN: clean inventory/schema sentinels passed; duplicate, removal,
# and missing-owner sentinels produced RED
```

No backend, frontend, browser, Docker, or load suite was executed by TASK-202.
`baseline_status` is therefore intentionally `not-executed-in-task-202` on
every row.

## Reconciled inventory

| Evidence | Observed |
|---|---:|
| Go `_test.go` source files | 828 |
| Go declared test/benchmark/example functions | 3663 |
| Go files classified `go-unit` by content | 605 |
| Go files classified `go-component` by content | 223 |
| Vitest source files | 180 |
| Vitest literal declared cases | 517 |
| Playwright TypeScript specs | 20 |
| Playwright literal declared cases | 53 |
| Playwright runtime skip sites (excluded from case count) | 6 |
| Canonical `export_hooks.go` files | 39 |
| Adjacent `export_hooks*.go` fragments | 4 |
| Export-hook files inspected in total | 43 |
| Exported hook symbols | 582 |
| Exported hook symbols with a resolved consumer | 550 |
| Target Python tests under `testframework/python/**` | 0 |

Canonical target classes reconcile to `605` `go-unit`, `223`
`go-component`, `180` `frontend-unit`, and `20` `python-e2e` rows.

## Owner-wave reconciliation

Every one of the 1,028 rows has exactly one deterministic `owner_wave`:

| Owner wave | Rows |
|---|---:|
| `P130/TASK-206` | 48 |
| `P131/TASK-207` | 91 |
| `P131/TASK-208` | 41 |
| `P133/TASK-210` | 131 |
| `P133/TASK-211` | 90 |
| `P134/TASK-212` | 219 |
| `P134/TASK-213` | 101 |
| `P135/TASK-214` | 92 |
| `P135/TASK-215` | 195 |
| `P138/TASK-195` | 15 |
| `P138/TASK-196` | 5 |

The 20/53 Playwright count is reconciled by requiring a literal title as the
first argument of a declared `test`/`it` call.  The six conditional
`test.skip(condition, reason)` calls are skip sites inside already-declared
cases and do not create six additional cases.

## Safety/parity gates encoded in the ledger

- Each source path is unique and is compared to the live, explicitly-scoped
  source set during `--check`.
- `inventory.json` is validated against the documented stdlib-only subset used
  by `inventory.schema.json`; unsupported assertion keywords are rejected.
- Runner → target class/root mappings are validated.
- `owner_wave` is required by schema and recomputed from the canonical mapping;
  missing, unknown, or mismatched ownership is RED.
- Exact 828/180/20/53/39/0 baseline facts are hard gates.  Drift is RED rather
  than silently accepted.
- Hook forward and reverse links are generated from Go import aliases; no hook
  deletion decision is made.
- Every legacy source disposition is `retain-until-parity`.
- `pro-qa-ai/**` is outside all read roots and any protected path leaked into
  the ledger is rejected.
- The sentinel mutates only in-memory copies and an OS temporary directory.  It
  proves duplicate/removal/missing-owner detection without touching
  `inventory.json`.

## Unresolved risks (for later tasks)

1. This is a static inventory, not a green runtime baseline.  Service-backed
   pass/fail/skip evidence belongs to TASK-203.
2. Go unit/component labels are deterministic content heuristics.  Borderline
   files still require owner review before wave assignment; directory names
   were intentionally not treated as evidence.
3. Thirty-two exported hook symbols have no resolved consumer in the current
   Go test source.  They are not declared dead and must not be removed without
   build/coverage/parity proof.
4. The four `backend/internal/tms/export_hooks_*.go` fragments are reported
   separately from the required 39 canonical `export_hooks.go` files so neither
   number can mask the other.
5. Literal-title extraction does not expand data-driven/dynamically-generated
   test cases.  The source hash and dependency hints preserve the evidence
   needed for manual parity review.
6. Python black-box tests cannot replace Go coverage evidence.  Coverage-bearing
   Go rows stay retained until the Go-native target and coverage gate are proven.
