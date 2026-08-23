---
name: coverage-audit
description: Матрица requirement/risk -> implementation -> observable test evidence. Обновляет .planning/COVERAGE.md без row-parity и historical percentage gates.
disable-model-invocation: true
---

# /coverage-audit — requirement и risk coverage

## Authority

Канон: `AGENTS.md`, `docs/test_plan.md`, current requirements/ADR/TZ и
`.cursor/playbooks/testing.md`. Исторические 82% Go coverage, package floors и
число legacy tests не являются migration gate или доказательством поведения.

Target taxonomy:

- genuine Go units: `backend/tests/**`;
- genuine Vitest/RTL units: `frontend/tests/**`;
- contract/API/integration/E2E/load: root Python `testframework/**`.

Legacy suites — delete-not-port inventory. Не запускать их для fallback,
parity или подтверждения current coverage.

## 1. Выбрать scope и источники

Зафиксировать task/phase/domain, exact source revision и применимые:

- `.planning/REQUIREMENTS.md`, current ROADMAP success criteria;
- `docs/system_specification.md`, `docs/test_plan.md`, security/TZ;
- current ADR/gray-zone decisions и open audit findings;
- implementation paths и public/boundary contracts.

Полный `docs/` не загружать без cross-domain причины.

## 2. Построить failure model

Для каждого requirement/risk проверить применимость:

- happy/invalid и boundary/empty/one/many;
- authn/authz/tenant isolation;
- timeout/unavailable/partial/retry/recovery;
- duplicate/reorder/replay/idempotency/version compatibility;
- cancellation/concurrency/resource exhaustion;
- privacy/redaction/security, accessibility/performance по scope.

`N/A` допустимо только с коротким rationale. Это не квота сценариев.

## 3. Найти implementation и evidence

Read-only поиск разрешён на host. Для каждого row записать exact path/symbol и
target test/TC. Наличие файла, grep hit, mock-only green или coverage percentage
не доказывают поведение.

Runnable checks use only `docker compose -f docker-compose.local.yml run --rm <exact-service>`
with non-zero collection, exact image/command/resources, duration, artifacts and cleanup.
Without actual runnable evidence the status is `NOT_RUN`, not a legacy PASS.

## 4. Матрица

```markdown
| Requirement/Risk | Implementation | TC / target job | Oracle | Status | Gap/owner |
|---|---|---|---|---|---|
| FR-... | path:symbol @ sha | TC-... / python-api | observable result + forbidden side effect | PASS|FAIL|NOT_RUN|NOT_READY|N/A | action |
```

Статусы:

- `PASS` — fresh executable evidence на exact revision;
- `FAIL` — oracle опровергнут;
- `NOT_RUN` — target существует, но проверка не выполнена;
- `NOT_READY` — target scaffold/job ещё отсутствует;
- `N/A` — risk неприменим и rationale проверен.

## 5. Gap classification

Разделять:

- requirement не реализован;
- implementation есть, но oracle/TC отсутствует;
- test есть, но нечувствителен/false-green;
- wrong taxonomy/path;
- infra/flake блокирует evidence;
- docs/traceability stale.

CRITICAL/HIGH нельзя закрывать процентом, self-report или planned test.

## 6. Update и handoff

Обновить `.planning/COVERAGE.md` только наблюдаемыми фактами: source SHA,
матрица, gaps, owner, next oracle и дата. Percentage допускается как secondary
diagnostic trend после target execution, но не заменяет rows и не переносит
historical baseline в greenfield gate.

Передать result `sa-ver`/`qa-auto-ver` conclusion-blind: exact sources, current
code, raw evidence, без готового убеждающего verdict.
