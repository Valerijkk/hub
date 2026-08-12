# TASK-202 evidence — 2026-08-07

## Reproduction identity

- Git revision: `ee6d696c2b48a6bcfbf77fe948174e819d282cce`
- Inventory schema/tool version: `2` / `2.2.0`
- Interpreter used: bundled CPython `3.12.13`
- Generated ledger: `inventory.json`
- Ledger SHA-256: `01dae4e2074f883bbebb7c075f299eba1b12d196380e151a81242c952489b95d`
- Ledger rows: `1047` (`830` Go + `180` Vitest + `3` frontend
  support + `20` Playwright + `14` E2E support)

## Commands and results

Run from the repository root:

```powershell
& 'C:\Users\Valerii\.cache\codex-runtimes\codex-primary-runtime\dependencies\python\python.exe' `
  .planning/audits/test-rebase/inventory.py --generate
# WROTE .../inventory.json sha256=01dae4e2... rows=1047

& 'C:\Users\Valerii\.cache\codex-runtimes\codex-primary-runtime\dependencies\python\python.exe' `
  .planning/audits/test-rebase/inventory.py --check
# GREEN: ledger matches allow-listed inputs; sha256=01dae4e2... rows=1047

& 'C:\Users\Valerii\.cache\codex-runtimes\codex-primary-runtime\dependencies\python\python.exe' `
  .planning/audits/test-rebase/inventory.py --self-test
# GREEN: root Python/frontend/E2E/component mappings and clean inventory/schema
# sentinels passed; duplicate, removal, misroute, and missing-owner
# sentinels produced RED

powershell -NoProfile -ExecutionPolicy Bypass -File `
  .planning/audits/test-rebase/validate-ledger.ps1
# Ledger validator: .../dependencies/python/python.exe
# GREEN: ledger matches allow-listed inputs; sha256=01dae4e2... rows=1047
# GREEN: root Python/frontend/E2E/component mappings and clean inventory/schema
# sentinels passed; duplicate, removal, misroute, and missing-owner
# sentinels produced RED
# planning-contract validator self-test: PASS
# planning-contract validation: PASS (1 contract)
```

No backend, frontend, browser, Docker, or load suite was executed by TASK-202.
`baseline_status` is therefore intentionally `not-executed-in-task-202` on
every row.

## Reconciled inventory

| Evidence | Observed |
|---|---:|
| Go `_test.go` source files | 830 |
| Legacy `backend/testframework/**` Go files / declarations | 829 / 3665 |
| Explicit `backend/cmd/server/**` Go files / declarations | 1 / 2 |
| Go declared test/benchmark/example functions | 3667 |
| Go files classified `go-unit` by content/dependencies | 594 |
| Go files classified `go-component` by content/dependencies | 236 |
| Go component conversion targets: `python-api` | 31 |
| Go component conversion targets: `python-integration` | 205 |
| Go component conversion targets: `python-contract` fallback | 0 |
| Vitest source files | 180 |
| Vitest literal declared cases | 519 |
| Vitest targets: `frontend-unit` | 156 |
| Vitest targets: `python-static-contract` | 24 |
| Frontend support/fixture source rows | 3 |
| Playwright TypeScript specs | 20 |
| Playwright literal declared cases | 53 |
| Playwright runtime skip sites (excluded from case count) | 6 |
| Playwright TypeScript support/cleanup rows | 14 |
| Canonical `export_hooks.go` files | 39 |
| Adjacent `export_hooks*.go` fragments | 4 |
| Export-hook files inspected in total | 43 |
| Exported hook symbols | 582 |
| Exported hook symbols with a resolved consumer | 550 |
| Target Python tests under root project `testframework/**` | 0 |
| Unmapped source rows | 0 |

Canonical target classes reconcile to `594` `go-unit`, `31` `python-api`,
`205` `python-integration`, `0` `python-contract`, `156` `frontend-unit`,
`24` `python-static-contract`, `3` `frontend-support`, `20` `python-e2e`,
and `14` `python-e2e-support` rows.
The Python project is rooted directly at `testframework/**`; translated
Playwright targets map to Python files under `testframework/e2e/**`.
Frontend unit targets map directly to `frontend/tests/<domain>/**` (the legacy
`domains/` segment is removed), while the three non-test support assets map to
`frontend/tests/support/**`.

### Go component conversion mapping

`backend/tests/**` is reserved for the 594 Go-unit rows.  All 236 Go component
rows retain their Go source until parity and receive an explicit Python
conversion target:

1. `http-server` service evidence maps to `testframework/api/**` (`31` rows).
2. Persistence, broker, object-storage, or external-process service evidence,
   plus `testcontainers`/`dockertest` or shared `testframework/env` /
   `testframework/mongotest` dependency evidence, maps to
   `testframework/integration/**` (`205` rows).
3. A future component marker without either resolved evidence set maps to
   `testframework/contract/**` as an explicit parity-required fallback; no
   current row needed that fallback.

The converted filename is deterministic: `<domain>/<name>_test.go` becomes
`<class>/<domain>/test_<name>.py`.  Validation rejects a Python component row
under `backend/tests/**`, rejects non-`.py` Python targets, and reports
unmapped rows as RED.  Current component-under-backend and unmapped counts are
both zero.

The separately allow-listed `backend/cmd/server/wiring_fiber_recover_test.go`
contains two declarations, is service-free under the same dependency rule, and
therefore maps to `backend/tests/server/wiring_fiber_recover_test.go` with the
genuine-unit owner `P133/TASK-211`.

### Frontend/static/support mapping

- The 156 browser-independent Vitest/RTL unit rows map directly to
  `frontend/tests/<domain>/**`.
- The 24 rows importing Node filesystem/path/process APIs map to
  `testframework/contract/frontend/<domain>/test_*.py`.  These are explicit
  static-contract conversion targets and remain `retain-until-parity`; mixed
  UI/static files may need splitting during migration.
- `render.tsx` and the two `*.fixture.ts` assets are first-class
  `frontend-support` rows under `frontend/tests/support/**`.
- All 14 TypeScript E2E config/shared helper files are first-class
  `python-e2e-support` rows under `testframework/e2e/support/**`; they are not
  silently dropped as cleanup.

## Owner-wave reconciliation

Every one of the 1,047 rows has exactly one deterministic `owner_wave`:

| Owner wave | Rows |
|---|---:|
| `P130/TASK-206` | 35 |
| `P131/TASK-207` | 85 |
| `P131/TASK-208` | 39 |
| `P131/TASK-230` | 24 |
| `P133/TASK-210` | 99 |
| `P133/TASK-211` | 61 |
| `P133/TASK-227` | 63 |
| `P134/TASK-212` | 179 |
| `P134/TASK-213` | 82 |
| `P134/TASK-228` | 59 |
| `P135/TASK-214` | 41 |
| `P135/TASK-215` | 132 |
| `P135/TASK-229` | 114 |
| `P137/TASK-194` | 14 |
| `P138/TASK-195` | 15 |
| `P138/TASK-196` | 5 |

The 236 Python component targets route by domain to TASK-227 (`63`), TASK-228
(`59`), or TASK-229 (`114`).  All 24 frontend static-contract rows route to
TASK-230.  Genuine Go/frontend units retain their original TASK-210…215 and
TASK-206…208 owner waves.

The 20/53 Playwright count is reconciled by requiring a literal title as the
first argument of a declared `test`/`it` call.  The six conditional
`test.skip(condition, reason)` calls are skip sites inside already-declared
cases and do not create six additional cases.

## Safety/parity gates encoded in the ledger

- Each source path is unique and is compared to the live, explicitly-scoped
  source set during `--check`.
- `inventory.json` is validated against the documented stdlib-only subset used
  by `inventory.schema.json`; unsupported assertion keywords are rejected.
- Runner → target class/root mappings are validated, including the invariant
  that only `go-unit` rows target `backend/tests/**`.
- `owner_wave` is required by schema and recomputed from the canonical mapping;
  missing, unknown, or mismatched ownership is RED.
- Exact 830 Go (829 legacy + 1 server) / 3667 declarations (3665 + 2) / 180
  Vitest / 3 frontend-support / 20 Playwright / 14
  E2E-support / 39 canonical-hook / 0 target-Python / 0-unmapped facts and the
  594/236 Go plus 156/24 frontend classifications are hard gates. Drift is RED.
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
2. Go unit/component labels and Python conversion classes are deterministic
   content/dependency heuristics. Borderline files still require owner review
   before wave assignment; directory names were intentionally not treated as
   evidence.
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
7. The 24 frontend static-contract rows include mixed guard/UI files. Their
   Python contract target records the non-unit dependency, but parity review
   must split any still-valid RTL assertions rather than discarding them.
