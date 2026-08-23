# Hub Test Framework pipeline

## Current QA execution contract (2026-08-23)

Hub runs QA only through root Compose:

```text
docker compose -f docker-compose.local.yml run --rm <exact-service>
```

The exact services are `backend-unit`, `frontend-unit`, `python-contract`,
`python-api`, `python-integration`, `python-e2e-smoke`, `python-e2e-full`, and
`python-load`. They are disposable; integration and e2e use isolated QA
infrastructure. Windows-host test, build, lint, validator, Vite, Playwright,
pytest, and Locust execution is forbidden.

`python-contract` runs framework/package selftests only and has Docker evidence
of **181/181 PASS** on 2026-08-23. API, integration, e2e, load, SUT product
behavior, and Mail/Calendar runtime remain `NOT_RUN`. The five SUT contract
tests/cases were removed by founder direction and are not a merge gate.

`testframework/runner/**`, `qa-gate`, direct Docker snippets, and legacy
pipeline scripts are retired/archival only; none is a fallback entrypoint.

## Historical pipeline inventory

The pre-greenfield `pipeline/**`, `stages/**`, legacy Go `backend/testframework/**`,
TypeScript Playwright, and legacy job names remain historical inventory only.
They do not describe current command authority or a runnable CI surface.
