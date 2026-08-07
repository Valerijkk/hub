# TASK-217 — 🧪 Python 3-spec smoke parity

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **137**

## [OUTCOME]

Python Playwright reproduces the exact current three-spec smoke contract with fail-closed preflight, zero unexpected skips and artifacts comparable to legacy smoke.

## [ROLE]

- Owner: `qa-auto2` — smoke port, preflight and parity report.
- Verifier / handoff: `qa-auto-ver` — exact selection/case/assertion and failure sensitivity verification.

## [SCOPE AND CONSTRAINTS]

- Allowed: `testframework/python/e2e/**`, smoke manifest/report.
- Read-only: three legacy smoke specs, active Playwright config/shared helpers.
- Do not touch: legacy files/config, product code, CI, remaining 17 specs.
- Smoke selection comes from current executable config, not stale docs; preserve frontend-only vs backend-required semantics explicitly.
- Depends on: TASK-194.

## [REFERENCES]

- `frontend/playwright.config.ts`, selected legacy specs, TASK-202 ledger
- `.planning/codebase/{ARCHITECTURE,TESTING,INTEGRATIONS}.md`

## [ACCEPTANCE]

- Completion command: `cd testframework/python && python -m pytest e2e -m smoke --strict-markers -q`; exit code 0 и заявленные counts обязательны.

- Manifest maps exactly three legacy specs and all their cases/assertions to Python; old/new collected/executed/pass/fail/skip reconcile.
- Missing frontend/backend/dev-mode prerequisite fails preflight where mandatory; unexpected skips = 0; trace/screenshot redacted.
- Mutation proof: break one observable selector/assertion per smoke spec — mapped Python test RED; restore GREEN.
- Regression: legacy smoke still green; Python API/integration/fixtures green.
- Evidence: exact commands, counts, timings, artifacts, mapping and three mutation logs.
- Stop when: current smoke selection is ambiguous or parity needs product change; do not widen/narrow silently.

## [OUTPUT SCHEMA]

1. Verdict. 2. Three-spec mapping. 3. Dual-run/preflight/mutation evidence. 4. Regression/risks. 5. Handoff `qa-auto-ver`; skill-log.

## [OPERATING RULES]

- Port without deleting original; event-driven waits; triage/re-run old+new.
- Empty collection/unexpected skip/secret artifact = FAIL.
- Docs-sync smoke manifest; TASK-linked atomic commit, no AI attribution.
