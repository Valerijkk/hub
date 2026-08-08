# Hub — QA и тестовая автоматизация

> Founder greenfield cutover (LOCKED, 2026-08-07). Канон стратегии:
> [`docs/test_plan.md`](../../docs/test_plan.md) и
> [`06-TEST-ARCHITECTURE-REBASE.md`](../../.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).
> Evidence-backed правила: [`15-test-automation.md`](../knowledge/code-practices/15-test-automation.md),
> [`01-evidence-model.md`](../knowledge/code-practices/01-evidence-model.md).

Этот playbook определяет активный QA-контур Hub. Старые пути могут встречаться
в истории и inventory, но не являются target, шаблоном или источником кода для
новых тестов.

## 1. Неподвижная архитектура

| CI job | Scope | Target | Владелец |
|---|---|---|---|
| `backend-unit` | genuine isolated Go units | `backend/tests/{domain}/**` | `be1`/`be2` |
| `frontend-unit` | genuine Vitest/RTL units | `frontend/tests/{domain}/**`, `frontend/tests/support/**` | `fe1`/`fe2` |
| `python-contract` | API/schema/event/source contracts | `testframework/tests/contract/**` | `qa-auto2` |
| `python-api` | публичное поведение backend/BFF | `testframework/tests/api/**` | `qa-auto1` |
| `python-integration` | DB/broker/storage boundaries | `testframework/tests/integration/**` | `qa-auto1` |
| `python-e2e-smoke/full` | браузерные user journeys | `testframework/tests/e2e/**`; `smoke` marker, `full` launcher profile | `qa-auto2` |
| `python-load` | load/resilience profiles | `testframework/load/**` | `qa-auto1` + DevOps |

Root `testframework/` — один greenfield Python-проект:

```text
testframework/
  pyproject.toml
  core/
  contract/
  api/
  integration/
  e2e/
  load/
  docker/
  reports/
```

Обязательные запреты:

- `backend/tests/**` не принимает HTTP/component/integration tests, БД,
  broker, storage или service containers;
- `frontend/tests/**` не принимает static/source-contract tests и проверки,
  которые просто grep/парсят production source;
- legacy Go component/API/integration, static-contract Vitest, TypeScript
  Playwright и их support/config/entrypoints удаляются по exact manifest и
  **не портируются**;
- нет row parity, dual-run и «перевода» legacy case-by-case;
- исторические 82% Go statement coverage и 80% package floors — baseline для
  контекста, не migration gate и не замена requirement/risk coverage;
- `pro-qa-ai/**` не читается и не меняется без явного founder handoff.

## 2. Как выбрать suite

Сначала назвать failure mode и наблюдаемый boundary, затем выбрать самый
быстрый hermetic scope, который способен обнаружить риск. `size` (ресурсы и
изоляция) и `scope` (объём системы) — разные оси.

| Риск/граница | Предпочтительный evidence |
|---|---|
| pure calculation, validation, reducer, component behavior | native unit |
| публичный HTTP/BFF contract | Python API |
| schema/event/source compatibility | Python contract |
| PostgreSQL/Mongo/Redis/NATS/storage semantics | Python integration |
| пользовательский результат через браузер | Python E2E |
| throughput, saturation, recovery under load | Python load |

Пирамида, Testing Trophy и Honeycomb допустимы как модели для обсуждения, но
не задают обязательные пропорции. Количество тестов по слоям не является gate.

## 3. Контракт test case

Каждый новый Python case имеет стабильный `TC-*` и metadata:

- `task`, `requirement` (`FR-*`, `NFR-*`, `GZ-*`, ADR/API/spec section);
- `risk`, `layer`, `domain`, `priority`, `owner`;
- предусловия и данные;
- observable oracle и запрещённые side effects;
- cleanup и ссылка на human-readable case/catalog.

Каждый применимый requirement/risk рассматривается через failure model:

- happy и invalid;
- authn, authz и tenant isolation;
- boundary; empty/one/many; null/Unicode/time/timezone;
- duplicate/reorder и version compatibility;
- timeout/unavailable/partial completion/retry/recovery;
- cancellation/concurrency/replay/idempotency;
- resource exhaustion;
- security/privacy/redaction, accessibility и performance — когда применимо.

Это пространство выбора, не квота. `N/A` допустимо с кратким rationale.
Техники equivalence classes, boundary values, decision tables, state machines
и pairwise применяются только там, где они уменьшают реальный риск.

## 4. Качество тестового кода

- Один test выражает один behavioral concept, а не ритуально один assert.
  Несколько assertions допустимы, если вместе доказывают один outcome.
- DAMP для scenario, DRY для механического setup. Не выносить важные значения
  и ожидания в DSL, который скрывает рассказ теста.
- Stub/fake/spy/mock выбирается по contract. Interaction assert допустим,
  только если collaboration наблюдаема клиентом или является явным contract.
- Не привязывать тест к private call order или incidental implementation.
- UI сначала использует accessible semantic locator. Tiny action helper,
  Component Object или Page Object появляются после повторения meaningful
  UI-knowledge. POM/Steps не обязательны и сами по себе не являются quality.
- Fixed sleep не синхронизирует test. Использовать controlled time,
  deterministic condition, bounded eventual assertion и explicit cleanup.
- Shared mutable accounts, records, schema, artifact paths и ports запрещены;
  namespace включает task/run/worker ID.

Подробнее: [backend Go](testing/backend-go.md), [frontend](testing/frontend.md),
[Python E2E](testing/e2e.md), [test design](testing/test-design.md),
[tooling/evidence](testing/tooling.md), [Hub risks](testing/reference.md).

## 5. Workflow

```text
requirement + task + threat/failure model
  -> qa-manual: TC и observable oracle
  -> qa-manual-ver: independent requirement/TC review
  -> owner выбирает suite по boundary
  -> be/fe пишут genuine units; qa-auto1/2 пишут greenfield Python non-unit
  -> Docker-only execution + exact evidence + negative sentinel
  -> qa-auto-ver: independent verdict
  -> target CI jobs
```

Автоматизация может начаться только с однозначного requirement/oracle. Это не
означает quota «один TC на каждую строку»: один case может закрывать несколько
связанных требований, а один риск — требовать несколько cases.

При FAIL обязателен [`/test-triage`](../skills/test-triage/SKILL.md): сначала
проверить поведение и docs, затем классифицировать `code`, `test`, `infra`,
`flaky` или `stop`. Assertion не меняют ради green.

## 6. Docker-only execution

На Windows host разрешены чтение/редактирование файлов, Git и Docker CLI.
Tests, validators, linters, build/vet, Vite, backend/frontend processes,
pytest/Playwright и Locust исполняются только в task-owned containers/Compose:

- уникальное имя/project: `hub-<job>-<task>`;
- явные CPU/RAM и bounded timeout;
- source mount read-only, если runner не обязан писать;
- isolated data/ports/network/artifacts;
- `--rm` либо guaranteed `down -v --remove-orphans`;
- zero collection, unexpected skip или leaked resource = FAIL.

Единственная точка исполнения — versioned outer launcher из
[`/qa-gate`](../skills/qa-gate/SKILL.md). Он enforce wall-clock timeout,
сохраняет exit code и выполняет exact-resource cleanup в `finally`; artifacts
изолированы в `testframework/reports/<task-run>/` с owner marker. Пока launcher
не реализован и независимо не проверен, любой runtime verdict — `NOT_READY`.
Прямой `docker run`/`compose up` не является допустимым substitute.

## 7. Evidence и false confidence

Минимальная запись runnable evidence:

```yaml
job: python-api
source_sha: <sha>
image: <immutable tag or digest>
command: <exact versioned launcher invocation + recorded inner command>
resources: {cpus: 2.0, memory: 4g}
service_profile: <none or compose profile>
result: {collected: N, passed: N, failed: 0, skipped: 0}
duration: <seconds>
artifacts: <task-owned URI>
cleanup: PASS
```

Gate — точная traceability `task/requirement/TC`, чувствительный observable
oracle, non-zero collection, expected failure от временного sentinel/mutation
для high-risk behavior и воспроизводимый Docker provenance. Нельзя выдавать
наличие файла, raw grep, зелёный mock-only assert, coverage percentage или
старый host-run за доказательство поведения.

## 8. Definition of Done

- путь и job соответствуют таблице §1;
- suite выбран по risk/boundary/feedback, а не по геометрии или quota;
- TC и metadata разрешаются в существующие task/requirement;
- oracle проверяет observable outcome и forbidden side effects;
- применимые scenario families рассмотрены, `N/A` объяснены;
- Docker evidence содержит exact command/result/duration/artifact/cleanup;
- zero collection/unexpected skips/leaks отсутствуют;
- high-risk sentinel действительно красный без защищаемого behavior;
- docs/traceability синхронизированы;
- independent verifier дал verdict по evidence, а не по self-report.

## 9. Source precedence

Founder locks → Hub safety/authority → GSD lifecycle → совместимые
Superpowers practices → локальные heuristics. Исследовательский control catalog
не означает, что control уже blocking: сверять
[`adoption-status.md`](../knowledge/code-practices/controls/adoption-status.md).

Связанные документы: [`docs/TESTING.md`](../../docs/TESTING.md),
[`docs/test_plan.md`](../../docs/test_plan.md),
[`test-triage`](../skills/test-triage/SKILL.md),
[`qa-gate`](../skills/qa-gate/SKILL.md),
[`verification.md`](verification.md).
