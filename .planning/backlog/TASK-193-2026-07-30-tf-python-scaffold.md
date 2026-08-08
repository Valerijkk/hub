# TASK-193 — 🐍 Root Python scaffold at cutover

> Revision: **2026-08-07 founder greenfield rebase** · Priority: **P1** · Status: Open · Phase: **129**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

`testframework/` becomes one reproducible Python QA project directly at root;
no nested Python project and no legacy executable test code coexist after Phase 129.

## [ROLE]

- Owner: `qa-auto1`; verifier: `qa-auto-ver`.
- Depends on: Phase 128 exact manifests.

## [SCOPE]

Root `pyproject.toml`, lock, pytest config, `conftest.py`,
`{core,api,integration,contract,e2e,load,docker,reports}/**`, README and pinned
Docker image/profile. Do not read/change `pro-qa-ai/**`.

## [OWNED TARGET PATHS]

`testframework/pyproject.toml`, `testframework/tests/conftest.py`, `testframework/{src/hub_testkit,tests/{api,integration,contract,e2e},selftests,cases,load,docker,reports}/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Clean Docker install/collect/run is non-zero with strict markers and TC metadata.
- Runtime config is typed/fail-closed; artifacts redact secret/PII canaries.
- Temporary failing test makes runner non-zero; restore green.
- Executable test/load code under root is Python-only.
- Evidence records image digest, SHA, limits, counts/duration and resource cleanup.

## [OPERATING RULES]

Docker-only execution with CPU/RAM limits; task resources removed after use.
Atomic allowlist commit, no push/AI attribution. Legacy row/file/case counts are not acceptance criteria.
