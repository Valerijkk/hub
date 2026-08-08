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

## Единый outer runner — обязательный gate

Все восемь jobs запускаются только через один versioned PowerShell-safe Docker
launcher из target scaffold. До появления и отдельной проверки этого launcher
любой runtime verdict `/qa-gate` — `NOT_READY`; прямые `docker run`/`compose up`
из документации не являются допустимой заменой.

Launcher принимает `job`, `task_run`, immutable image digest и wall-clock timeout
и обязан:

1. проверить безопасный формат `task_run` и вычислить exact container/Compose/volume names;
2. создать только `testframework/reports/<task-run>/` и owner marker с task/revision;
3. задать CPU, RAM/swap, PID, network и read-only-source ограничения;
4. enforce outer wall-clock deadline независимо от inner test timeout;
5. сохранить исходный exit code и raw stdout/stderr;
6. в безусловном `finally` остановить exact container/Compose project и удалить
   только task-owned volumes/networks;
7. после cleanup доказать отсутствие exact resources; cleanup failure делает job FAIL.

Никакой job не пишет в общий `testframework/reports/` root. JUnit/Allure/traces
пишутся только внутрь `/reports/<task-run>/`, смонтированного из одноимённого
task-owned host directory.

## Inner job contracts

`backend-unit` выполняет build/vet и service-free genuine Go units из
`backend/tests/**`. `frontend-unit` выполняет tsc/lint/genuine Vitest/RTL units
из `frontend/tests/**` и build; static/source contracts запрещены. Python jobs
используют immutable target image из `testframework/docker/**` и следующие
entrypoints внутри test container:

```text
python-api:          pytest -p no:cacheprovider -m api
python-integration:  pytest -p no:cacheprovider -m integration
python-e2e-smoke:    pytest -p no:cacheprovider -m "e2e and smoke"
python-e2e-full:     pytest -p no:cacheprovider -m "e2e and full"
python-load:         python -m locust -f load/locustfile.py --headless <approved profile>
```

Если launcher, `tests` service, profile или Locust entrypoint ещё не реализован
target scaffold, зафиксировать `NOT_READY`. Не изобретать legacy substitute и
не копировать внутренний command в небезопасный host-side запуск.

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
artifacts: testframework/reports/<task-run>/
owner_marker: <task/revision marker URI + verified value>
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
