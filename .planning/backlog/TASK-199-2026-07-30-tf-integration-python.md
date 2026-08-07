# TASK-199 — 🧪 Python core и API/integration pilot

> Заведено: 2026-07-30 · Revision: **2026-08-07** · Приоритет: **P1** · Статус: Открыта · Фаза: **136**

## [OUTCOME]

Python framework имеет reusable core и по одному реальному API и infrastructure-backed integration pilot с изоляцией, строгими markers и dual-run mapping к legacy behavior.

## [ROLE]

- Owner: `qa-auto1` — core clients/factories, API pilot и DB/broker integration pilot.
- Verifier / handoff: `qa-auto-ver` — clean-env, isolation, mapping и mutation verification.

## [SCOPE AND CONSTRAINTS]

- Allowed: `testframework/{core,api,integration,contract,docker,reports}/**` и корневой config единого Python-проекта. Shared fixtures живут в `testframework/core/fixtures/**`; отдельный root `fixtures/` не создаётся.
- Read-only: corresponding Go legacy tests/ledger and public API docs.
- Do not touch: legacy Go tests, product code, CI, E2E.
- Tests interact through public HTTP/service boundaries; no Go imports, direct production internals or silent skips. Infrastructure profile must fail closed for required services.
- Depends on: TASK-193, TASK-204, TASK-209 through TASK-215 ledgers/statuses.

## [REFERENCES]

- TASK-197 backend ledger; `.planning/codebase/{TESTING,INTEGRATIONS,STRUCTURE}.md`
- `docs/wiki/API.md`, `docs/security_guidelines.md`, `.cursor/playbooks/testing.md`

## [ACCEPTANCE]

- Completion command: `cd testframework && python -m pytest -m 'api or integration' --strict-markers -q`; exit code 0 и заявленные counts обязательны.

- `pytest -m api` and `pytest -m integration` each collect non-zero and pass on documented clean environment; mandatory DB/broker outage fails setup, not skip/time out.
- Every pilot has legacy source row, TC/behavior mapping, unique data namespace, deterministic cleanup and sanitized artifact.
- Mutation proof: break an API contract guard and one real publish/persist boundary in a reversible local mutation — matching pilot RED; restore GREEN. If production mutation is unsafe, use approved local patch and prove restoration diff-clean.
- Regression: relevant legacy Go domain tests and Python smoke green; Go coverage contract remains governed by TASK-203/204 and is not claimed from Python.
- Evidence: environment/services, commands, counts/skips, wall time, mapping, mutation logs.
- Stop when: service unavailable or behavior lacks stable public contract; report blocker, do not replace with mocks.

## [OUTPUT SCHEMA]

1. Verdict and 2–5 sentence summary.
2. Core/fixture design and pilot mapping.
3. Commands/counts/timing/mutation RED→GREEN.
4. Regression, cleanup/redaction checks, assumptions.
5. Handoff `qa-auto-ver`, risks, skill-log.

## [OPERATING RULES]

- New behavior loop then legacy domain regression; classify FAIL before changes.
- No unexpected skips, production endpoints or secrets; test data must be isolated and cleaned.
- Legacy originals remain until TASK-219; docs-sync Python README/mapping.
- Atomic commits by pilot with TASK link, no AI attribution.
