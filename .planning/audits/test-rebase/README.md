# TASK-202 — reproducible test inventory

> **Historical inventory notice (superseded 2026-08-23):** records below retain
> TASK-202/TASK-204 discovery provenance and are not the current verdict.
> Current minimal-slice scope, Docker evidence, deletion reconciliation and
> publication rule are in `FINAL-MERGE-EVIDENCE-2026-08-23.md`.

This directory is the evidence boundary for the test-architecture rebase.  The
inventory is descriptive only: no legacy source is deleted or moved by this
tool.

## Files

- `inventory.py` — Python 3.12 stdlib-only generator, validator, and sentinel.
- `inventory.schema.json` — JSON Schema for every ledger row and hook record.
- `inventory.json` — deterministic generated ledger pinned to a Git revision.
- `validate-ledger.ps1` — safe PowerShell entrypoint for check + sentinel.
- `validate_planning_contracts.py` — cross-platform fail-closed ADR marker gate.
- `evidence.md` — commands, observed totals, ledger digest, and unresolved risks.

## Explicit read boundary

The generator never walks the repository root.  It reads only:

- `backend/testframework/**/*_test.go`;
- `backend/cmd/server/**/*_test.go` (the explicit server wiring unit/support
  source; it targets `backend/tests/server/**` only while service-free);
- `testframework/stages/unit/frontend-test/**/*.{test,spec}.{ts,tsx}`;
- `testframework/stages/unit/frontend-test/render.tsx` and
  `**/*.fixture.{ts,tsx}` frontend support assets;
- `testframework/stages/e2e/domains/**/*.spec.ts`;
- `testframework/stages/e2e/{playwright.config.ts,shared/**/*.{ts,tsx}}` E2E
  support/config sources;
- `backend/internal/**/export_hooks*.go`;
- `testframework/**/*.py` only to count already-created target tests;
- `git rev-parse HEAD` for the current revision.

The recorded revision identifies the source snapshot and must be an ancestor
of the current HEAD; literal equality is intentionally not required because
committing the ledger itself advances HEAD. Live completeness, content hashes,
schema checks, and duplicate/removal sentinels prove that the snapshot still
matches the executable test sources.

`pro-qa-ai/**` is outside every scan root and is additionally rejected by the
validator if a protected path appears anywhere in the ledger.

## Classification and targets

Target paths are mechanical proposals, not proof that a replacement exists:

| Current runner | Target class | Target root |
|---|---|---|
| Go genuine unit | `go-unit` | `backend/tests/<domain>/**` |
| Go component/API/integration | `python-api`, `python-integration` or `python-contract` | `testframework/{api,integration,contract}/**` |
| Vitest behavior unit | `frontend-unit` | `frontend/tests/<domain>/**` |
| Vitest source/docs/compile guard | `python-static-contract` | `testframework/contract/frontend/**` |
| Frontend fixtures/render support | `frontend-support` | `frontend/tests/support/**` |
| Playwright TypeScript | `python-e2e` | `testframework/e2e/**` |
| Playwright config/shared support | `python-e2e-support` | `testframework/e2e/support/**` |

Python black-box suites live directly under root `testframework/**`; Locust remains
separate under `testframework/load/**`.  All legacy rows have disposition
`retain-until-parity`.

Go classification is content-based.  A file is `go-component` only when the
source contains a concrete persistence/process/network boundary marker:
`enttest.Open`, a test-service URL, testcontainers/dockertest, an in-process
HTTP/TCP server, an external process, an embedded NATS server, or imports the
shared service-backed `testframework/env` / `testframework/mongotest` support.
Otherwise it is `go-unit`.  The explicit `backend/cmd/server` source follows
the same rule; its current service-free row maps to `backend/tests/server/**`.
Neither the directory nor filename alone turns a test into a component.

`test.skip(condition, reason)` inside a Playwright test is recorded as a skip
site, but is not counted as a declared test case.  A declared case must pass a
literal title as the first argument to `test(...)`/`it(...)` (including an
explicitly declared `.skip("title", callback)`).

Hook records cover both the 39 canonical `export_hooks.go` files and adjacent
`export_hooks*.go` fragments.  Each exported symbol lists consumers resolved
through the test file's Go import alias.  This is dependency evidence only;
TASK-202 makes no hook deletion decision.

Every ledger row has one mandatory `owner_wave`.  It is a migration owner, not
a claim that parity already exists:

| Source group | Owner wave |
|---|---|
| Go genuine units: auth/rbac/audit/ws/calls | `P133/TASK-210` |
| Go genuine units: platform/apigateway/dev/ent/env/status/server | `P133/TASK-211` |
| Go genuine units: tracker/project/sprint/tms | `P134/TASK-212` |
| Go genuine units: messenger/notification | `P134/TASK-213` |
| Go genuine units: calendar/intranet/accounting/crm/hr/bases/dashboard | `P135/TASK-214` |
| Go genuine units: AI/search/storage/backup/imex/GitHub/GitLab/1C/marketplace/mail | `P135/TASK-215` |
| Frontend auth/platform | `P130/TASK-206` |
| Frontend tracker/messenger/tms/calls | `P131/TASK-207` |
| Remaining frontend domains | `P131/TASK-208` |
| Frontend static/source contracts (all domains) | `P131/TASK-230` |
| Go component targets: security/platform domains | `P133/TASK-227` |
| Go component targets: work/realtime domains | `P134/TASK-228` |
| Go component targets: business/integration domains | `P135/TASK-229` |
| Playwright config/shared support | `P137/TASK-194` |
| 15 mechanical Playwright specs | `P138/TASK-195` |
| 5 complex Playwright specs | `P138/TASK-196` |

The five complex specs are explicitly allow-listed in `inventory.py`; all other
Playwright sources go to the mechanical wave.  An unknown domain is a generator
error rather than an invented assignment.

## Reproduce and validate

From the repository root, using any Python 3.12+ interpreter:

```powershell
python .planning/audits/test-rebase/inventory.py --generate
python .planning/audits/test-rebase/inventory.py --check
python .planning/audits/test-rebase/inventory.py --self-test

# Or run the non-writing validation pair through PowerShell:
pwsh -File .planning/audits/test-rebase/validate-ledger.ps1

# Planning-only decisions referenced by runtime tests:
python .planning/audits/test-rebase/validate_planning_contracts.py --self-test
python .planning/audits/test-rebase/validate_planning_contracts.py
```

`--check` first validates `inventory.json` against `inventory.schema.json`, then
rebuilds the expected ledger in memory and validates duplicate/missing
sources, target roots/classes, exact baseline totals, hook reverse links,
Playwright 20-file/53-case reconciliation, and protected-path exclusion, then
byte-compares it with `inventory.json`.  Any drift exits non-zero.

Schema validation is a documented stdlib-only subset, not a claim of general
JSON Schema support.  It enforces the keywords used by this schema (`type`,
`required`, `properties`, `additionalProperties`, `items`, local `$ref`,
`enum`, `const`, `pattern`, and `minimum`) and reports any unsupported schema
assertion keyword instead of silently ignoring it.

The sentinel uses only an in-memory object plus an OS temporary directory.  It
must prove clean → GREEN, duplicate → RED, and removal → RED without touching
the real ledger.

`validate-ledger.ps1` also runs the planning-contract validator. This keeps
docs-only decisions executable without making the backend unit container
depend on a `.planning/**` mount. PR Guard runs the same validator; removal of
the ChannelMember LWW approval markers is therefore fail-closed.
