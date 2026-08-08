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

## Target runner contract

Один образ `hub-testframework:<sha>` должен быть собран из
`testframework/docker/**` и иметь immutable provenance. Каждый job использует
единый versioned outer launcher, task-owned name, CPU/RAM/PID/network limits,
read-only source, outer wall-clock timeout и unconditional `finally` cleanup.
Прямые copy/paste `docker run` и `compose up` не являются каноническим gate.

Каждый запуск заранее создаёт `testframework/reports/<task-run>/`, проверяет
owner marker с task/revision и монтирует только этот каталог как
`/reports/<task-run>`. JUnit, Allure, trace и screenshot не могут писать в
shared reports root.

Load исполняется в том же container contract, но inner command — принятый
Locust profile из `testframework/load/**`; host Locust запрещён. API,
integration, E2E и load services принадлежат Compose project
`hub-<job>-<task-run>`, имеют per-service CPU/RAM limits и удаляются outer
launcher только по exact task-owned names в `finally`, с post-cleanup proof.

До появления и независимой проверки реального launcher/scaffold runtime status
равен `NOT_READY`. Zero collection запрещено принимать как PASS.

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
