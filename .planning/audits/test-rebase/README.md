# TASK-202 — reproducible test inventory

This directory is the evidence boundary for the test-architecture rebase.  The
inventory is descriptive only: no legacy source is deleted or moved by this
tool.

## Files

- `inventory.py` — Python 3.12 stdlib-only generator, validator, and sentinel.
- `inventory.schema.json` — JSON Schema for every ledger row and hook record.
- `inventory.json` — deterministic generated ledger pinned to a Git revision.
- `validate-ledger.ps1` — safe PowerShell entrypoint for check + sentinel.
- `evidence.md` — commands, observed totals, ledger digest, and unresolved risks.

## Explicit read boundary

The generator never walks the repository root.  It reads only:

- `backend/testframework/**/*_test.go`;
- `testframework/stages/unit/frontend-test/**/*.{test,spec}.{ts,tsx}`;
- `testframework/stages/e2e/domains/**/*.spec.ts`;
- `backend/internal/**/export_hooks*.go`;
- `testframework/python/**/*.py` only to count already-created target tests;
- `git rev-parse HEAD` for the current revision.

`pro-qa-ai/**` is outside every scan root and is additionally rejected by the
validator if a protected path appears anywhere in the ledger.

## Classification and targets

Target paths are mechanical proposals, not proof that a replacement exists:

| Current runner | Target class | Target root |
|---|---|---|
| Go | `go-unit` or `go-component` | `backend/tests/**` |
| Vitest | `frontend-unit` | `frontend/tests/**` |
| Playwright TypeScript | `python-e2e` | `testframework/python/e2e/**` |

Python black-box suites remain under `testframework/python/**`; Locust remains
separate under `testframework/load/**`.  All legacy rows have disposition
`retain-until-parity`.

Go classification is content-based.  A file is `go-component` only when the
source contains a concrete persistence/process/network boundary marker:
`enttest.Open`, a test-service URL, testcontainers/dockertest, an in-process
HTTP/TCP server, an external process, or an embedded NATS server.  Otherwise it
is `go-unit`.  Neither the directory nor filename influences this decision.

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
| Go auth/rbac/audit/ws/calls | `P133/TASK-210` |
| Go platform/apigateway/dev/ent/env/status | `P133/TASK-211` |
| Go tracker/project/sprint/tms | `P134/TASK-212` |
| Go messenger/notification | `P134/TASK-213` |
| Go calendar/intranet/accounting/crm/hr/bases/dashboard | `P135/TASK-214` |
| Go AI/search/storage/backup/imex/GitHub/GitLab/1C/marketplace/mail | `P135/TASK-215` |
| Frontend auth/platform | `P130/TASK-206` |
| Frontend tracker/messenger/tms/calls | `P131/TASK-207` |
| Remaining frontend domains | `P131/TASK-208` |
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
