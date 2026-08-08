# Tooling, Docker evidence и отчёты

## Root Python project

`testframework/pyproject.toml` — единый authority для pytest markers,
dependencies, formatting/type policy и test discovery. Не создавать отдельные
Python projects по stage/domain и не переносить legacy runners.

| Каталог | Назначение |
|---|---|
| `core/` | config, clients, fixtures, traceability/evidence infrastructure |
| `contract/` | API/schema/event/source compatibility |
| `api/` | public backend/BFF behavior |
| `integration/` | DB/broker/storage boundaries |
| `e2e/` | pytest + playwright-python journeys |
| `load/` | Locust profiles |
| `docker/` | pinned images/Compose/entrypoints |
| `reports/` | task-owned generated results, не source authority |

## Target commands

Один образ `hub-testframework:<sha>` должен быть собран из
`testframework/docker/**` и иметь immutable provenance. Каждый job использует
task-owned name, CPU/RAM limits, read-only source и `--rm`:

```powershell
docker run --rm --name hub-python-<job>-<task> `
  --cpus 2.0 --memory 4g --memory-swap 4g `
  --mount type=bind,src="${PWD}/testframework",dst=/work/testframework,readonly `
  --mount type=bind,src="${PWD}/testframework/reports",dst=/reports `
  -e PYTHONDONTWRITEBYTECODE=1 -w /work/testframework hub-testframework:<sha> `
  pytest -p no:cacheprovider -m "<contract|api|integration|e2e>" `
  --junitxml=/reports/<task>-<job>.xml
```

Load исполняется в том же container contract, но inner command — принятый
Locust profile из `testframework/load/**`; host Locust запрещён. API,
integration, E2E и load services принадлежат Compose project
`hub-<job>-<task>`, имеют per-service CPU/RAM limits и завершаются:

```powershell
docker compose -p hub-<job>-<task> -f testframework/docker/compose.yaml down -v --remove-orphans
```

До появления реального scaffold эти команды — target contract, не claim о
готовом runner. Zero collection запрещено принимать как PASS.

## Evidence record

Для каждого запуска сохранить:

- source SHA и dirty-worktree provenance;
- image digest/tag и Compose/service profile;
- exact outer/inner command;
- CPU/RAM, timeout, workers/seed;
- collected/passed/failed/skipped/xfailed и duration;
- JUnit/Allure/trace/screenshot URI с redaction;
- container/volume/network cleanup result.

Allure — формат отчёта, не evidence сам по себе. Match выполняется стабильным
TC ID; fullName/className не является единственным ключом. Секреты, tokens,
cookies и PII в attachments запрещены.

## Parallelism и flake

- namespace data/schema/tenant/port/artifact по task/run/worker;
- записывать seed и shard;
- bounded eventual assertions вместо sleep;
- retry budget конечен и не превращает flaky в green;
- quarantine имеет owner, reason, expiry и blocking remediation;
- unexpected skip, collection drift и leaked resource = FAIL.

## Coverage

Coverage используется для поиска слепых зон, но не доказывает behavior.
Исторический Go 82% и row parity не являются migration gate. Новый baseline
публикуется честно после retain genuine units. Python quality gate —
requirement/risk/TC traceability, sensitive oracle, non-zero collection и
Docker provenance.

См. [`qa-gate`](../../skills/qa-gate/SKILL.md),
[`C20`](../../knowledge/code-practices/controls/hard-controls.md) и
[`adoption-status`](../../knowledge/code-practices/controls/adoption-status.md).
