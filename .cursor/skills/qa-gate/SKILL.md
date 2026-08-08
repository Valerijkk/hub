---
name: qa-gate
description: Founder-cutover quality gate — genuine Go/Vitest units плюс greenfield Python contract/API/integration/E2E/load, только в task-owned Docker с evidence и cleanup.
disable-model-invocation: true
---

# /qa-gate — target quality gate

Канон: `AGENTS.md` target table, `docs/test_plan.md`,
`.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md` и
`.cursor/playbooks/testing.md`.

## Founder locks

- На Windows host не запускать tests, validators, linters, build/vet, Vite,
  backend/frontend processes, pytest/Playwright или Locust. Host — только
  чтение/редактирование, Git и Docker CLI.
- Каждый runner task-owned, имеет CPU/RAM/timeout, isolated resources и cleanup.
- Не использовать legacy `backend/testframework/**`, static-contract Vitest,
  TypeScript Playwright, их configs/linkers/runners как gate или fallback.
- Исторический 82% Go coverage и row parity не gate.
- Zero collection, unexpected skip или leaked resource = FAIL.

До появления Phase 129 scaffold отсутствие target path/job — FAIL/NOT_READY,
а не повод прогнать legacy suite и назвать gate green.

## Target jobs

| Порядок | Job | Target | Обязательность |
|---:|---|---|---|
| 1 | `backend-unit` | `backend/tests/{domain}/**` | PR при backend change |
| 2 | `frontend-unit` | `frontend/tests/{domain}/**`, support | PR при frontend change |
| 3 | `python-contract` | `testframework/contract/**` | PR по contract/schema/event/source change |
| 4 | `python-api` | `testframework/api/**` | PR по public backend/BFF behavior |
| 5 | `python-integration` | `testframework/integration/**` | PR по DB/broker/storage boundary |
| 6 | `python-e2e-smoke` | `testframework/e2e/**`, marker `smoke` | PR/release gate по критическим journeys |
| 7 | `python-e2e-full` | `testframework/e2e/**`, marker `full` | merge wave/release |
| 8 | `python-load` | `testframework/load/**` | change-specific/nightly/manual baseline |

Выбор запускаемых jobs следует changed risk и observable boundary; он не
создаёт обязательную Pyramid/Trophy/Honeycomb ratio. Skipped job записывается
как `NOT_APPLICABLE` с rationale, не как PASS.

## 1. `backend-unit`

```powershell
docker run --rm --name hub-backend-unit-<task> `
  --cpus 2.0 --memory 4g --memory-swap 4g `
  --mount type=bind,src="${PWD}/backend",dst=/work/backend,readonly `
  -w /work/backend golang:1.25-bookworm `
  sh -lc "go build -tags release ./... && go vet ./... && go test -race -count=1 -timeout=600s ./tests/..."
```

Только service-free genuine units. `--rm` очищает container.

## 2. `frontend-unit`

```powershell
docker run --rm --name hub-frontend-unit-<task> `
  --cpus 2.0 --memory 4g --memory-swap 4g `
  --mount type=bind,src="${PWD}/frontend",dst=/work/frontend,readonly `
  --mount type=volume,src=hub-node-modules-<task>,dst=/work/frontend/node_modules `
  --mount type=volume,src=hub-frontend-dist-<task>,dst=/work/frontend/dist `
  -w /work/frontend mcr.microsoft.com/playwright:v1.61.0-noble `
  sh -lc "npm ci && npx tsc --noEmit && npm run lint && npx vitest run tests && npm run build"

docker volume rm hub-node-modules-<task> hub-frontend-dist-<task>
```

Только genuine Vitest/RTL units. Static/source contracts запрещены.

## 3. Python contract job

```powershell
docker run --rm --name hub-python-contract-<task> `
  --cpus 2.0 --memory 4g --memory-swap 4g `
  --mount type=bind,src="${PWD}/testframework",dst=/work/testframework,readonly `
  --mount type=bind,src="${PWD}/testframework/reports",dst=/reports `
  -e PYTHONDONTWRITEBYTECODE=1 -w /work/testframework hub-testframework:<sha> `
  pytest -p no:cacheprovider -m contract --junitxml=/reports/<task>-contract.xml
```

## 4. Python jobs с services

API/integration/E2E используют target Compose authority из
`testframework/docker/**`. Compose YAML обязан задавать CPU/RAM/healthcheck и
task-isolated data. Примеры markers/entrypoints внутри test container:

```text
python-api:          pytest -p no:cacheprovider -m api
python-integration:  pytest -p no:cacheprovider -m integration
python-e2e-smoke:    pytest -p no:cacheprovider -m "e2e and smoke"
python-e2e-full:     pytest -p no:cacheprovider -m "e2e and full"
python-load:         python -m locust -f load/locustfile.py --headless <approved profile>
```

Outer execution contract:

```powershell
docker compose -p hub-<job>-<task> -f testframework/docker/compose.yaml `
  --profile <job> up --abort-on-container-exit --exit-code-from tests

docker compose -p hub-<job>-<task> -f testframework/docker/compose.yaml `
  down -v --remove-orphans
```

Если `tests`/profile/Locust entrypoint ещё не реализован target scaffold,
зафиксировать FAIL/NOT_READY. Не изобретать legacy substitute.

## FAIL protocol

Остановиться на первой FAIL и применить
[`test-triage`](../test-triage/SKILL.md): сначала product behavior/docs/repro,
затем `ROOT=code|test|infra|flaky|stop`. После исправления повторить упавший job,
его смежные risk jobs и затем полный выбранный gate. Не менять assert/skip/retry
ради green.

## Evidence record

```yaml
source_sha: <sha>
job: <target job>
image: <digest or immutable tag>
command: <exact outer and inner command>
resources: {cpus: 2.0, memory: 4g, timeout: <value>}
services: <profile or none>
result: {collected: N, passed: N, failed: N, skipped: N}
duration: <seconds>
artifacts: <task-owned URI>
cleanup: PASS|FAIL
triage: <root/fix or none>
```

`PASS` требует non-zero collection, только expected skips, exact traceability
`task/requirement/TC`, cleanup PASS и для high-risk behavior чувствительный
sentinel/mutation evidence. Host/legacy/self-report evidence не принимается.

## Result

```text
QA GATE: PASS|FAIL|NOT_READY
Jobs: backend-unit=... frontend-unit=... python-contract=... python-api=...
      python-integration=... python-e2e-smoke=... python-e2e-full=... python-load=...
Evidence: source SHA, images, exact commands/results/durations/artifacts
Cleanup: containers/volumes/networks
NOT_RUN/NOT_APPLICABLE: job + rationale
Next: triage/remediation owner
```
