# TASK-207 — 🧪 Frontend tracker/messenger/tms/calls wave

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **131**

## [OUTCOME]

Frontend tests for tracker, messenger, tms and calls live in `frontend/tests/**` with complete ledger parity and unchanged 180-test legacy dual-run.

## [ROLE]

- Owner: `fe1` — four-domain migration and evidence.
- Verifier / handoff: `fe-ver` — type/lint/test/coverage and mutation verification.

## [SCOPE AND CONSTRAINTS]

- Allowed: `frontend/tests/{tracker,messenger,tms,calls}/**`, required shared test support only.
- Read-only: matching legacy tests and production SUT.
- Do not touch: legacy/linker/CI, platform/auth, remaining domains or product behavior.
- Cross-domain legacy file is split only with explicit one-to-many ledger mapping and no assertion loss.
- Depends on: TASK-198, TASK-206.

## [REFERENCES]

- TASK-202 ledger; four legacy domain directories
- `.planning/codebase/{STRUCTURE,TESTING,CONVENTIONS}.md`, `.cursor/playbooks/testing/frontend.md`

## [ACCEPTANCE]

- Completion command: `cd frontend && npx vitest run tests/tracker tests/messenger tests/tms tests/calls`; exit code 0 и заявленные counts обязательны.

- All four-domain rows are `PARITY_PROVEN`; old/new cases/assertions/TC IDs reconcile, duplicates/unmapped = 0.
- Target tests pass test typecheck, Biome and Vitest; coverage per domain does not regress beyond approved contract.
- Mutation proof: one meaningful product guard in each domain REDs its target test; all four restored GREEN/diff-clean.
- Regression: full legacy 180 and prior target platform/auth wave green; no unexpected skips.
- Evidence: commands/counts/timing/coverage, ledger status and four mutation logs.
- Stop when: source guard/path or behavior cannot be preserved without production change; report row, do not weaken test.

## [OUTPUT SCHEMA]

1. Verdict. 2. Four-domain parity table. 3. Static/test/coverage/mutation evidence. 4. Regression/risks. 5. Handoff `fe-ver`; skill-log.

## [OPERATING RULES]

- Atomic domain batches, originals retained, test triage for every FAIL.
- No snapshot/assertion weakening, duplicate discovery or unapproved skip.
- Docs-sync ledger; commits link TASK and contain no AI attribution.
