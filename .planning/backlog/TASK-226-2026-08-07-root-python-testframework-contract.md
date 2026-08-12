# TASK-226 — Root Python testframework contract

**Фаза:** 128 · **Приоритет:** P1 · **Статус:** CLOSED — documentation/package-contract boundary only
**Исполнитель:** `arch` + `qa-auto1` · **Проверка:** `arch-ver` + `qa-auto-ver`
**Depends on:** `TASK-202`

**Closeout evidence:** accepted founder ADR originally pinned to
`ff069dee371f40409c30e62e79e2f5b065005894`, independent conclusion-blind
`APPROVED` review, and synthetic provenance parent
`507e4bd691eb513c0f796e1eac5235a1a6178ff5`. These are documentation-contract
provenance only; they do not establish launcher admission or runtime readiness.

**Closeout boundary:** this closes only TASK-226 documentation/package-contract
metadata. It does not close or implement package bootstrap, framework selftests,
SUT suites, the launcher, or runtime.

## Outcome

Зафиксировать уточнение founder как единственный целевой контракт тестовой
архитектуры и пересчитать machine-readable ledger без старого вложенного пути.

```text
backend/tests/**    Go unit tests only
frontend/tests/**   Vitest/RTL unit tests only
testframework/**    one root Python QA project
```

В `testframework/**` исполняемый тестовый и нагрузочный код только Python.
Допустимы конфигурация, Docker-файлы, lockfile, схемы, отчёты и документация.
Вложенного `testframework/python/**` в target нет. Legacy TypeScript E2E
удаляется по exact manifest в первом безопасном cleanup batch; сравнение
legacy и нового наборов не является критерием.

## Owned target paths

Planning/inventory contract files only: the founder ADR, live Phase 128–141
planning contract and `.planning/audits/test-rebase/**` inventory/manifest
records named by the task. This card authorizes no test/code move or deletion
and no path beyond its live PLAN/manifest.

## Acceptance criteria

- [x] Канон, roadmap, phase plans и onboarding docs используют один root
  `testframework/**`, без target-ссылок на `testframework/python/**`.
- [x] Python layout: `src/hub_testkit/**`, planned
  `tests/{contract,api,integration,e2e}/**`, `selftests/**`, `cases/**`,
  `schemas/**`, `policies/**`, `load/**`, `docker/**`, `runner/**` source candidate,
  and
  `reports/<task-run>/**`. Direct root `core/`, `api/`, `integration/`,
  `contract/`, and `e2e/` trees are not target architecture.
- [x] `go-component` rows не попадают в `backend/tests/**` и удаляются без replacement:
  они входят в exact delete manifest. Новые Python API/integration/contract
  suites проектируются с нуля по актуальным требованиям, рискам и TC IDs.
- [x] Frontend target не сохраняет служебный legacy-сегмент `domains/`:
  `frontend/tests/<domain>/**` + `frontend/tests/support/**`. Все 24 legacy
  static-contract Vitest rows и их source/docs/compile guards удаляются по
exact manifest без порта. Любой будущий `testframework/tests/contract/frontend/**`
  suite является независимым greenfield design по current requirements/risks,
  а не replacement/translation этих строк.
- [x] Inventory/cutover manifest относит legacy Playwright TS specs, support и
  entrypoints только в exact `delete_legacy` set без replacement target path.
Новый Python E2E catalog в `testframework/tests/e2e/**` создаётся отдельно из
  актуальных requirements, risks и stable TC IDs.
- [x] Ledger/support closure is transferred to TASK-204/TASK-219/TASK-220; it is
  not a TASK-226 docs-closeout gate.
- [x] Ledger regeneration, schema/check/self-test and `unmapped = 0` validation
  are transferred to TASK-204/TASK-220; they are not TASK-226 docs-closeout gates.
- [x] Current-vs-target distinction сохранён; файлы legacy не удалены этой
  задачей.
- [x] `pro-qa-ai/**` protected-zone verification is recorded by the independent
  closeout; it is not a runtime or launcher-readiness claim.

## Readiness

The documentation contract can be corrected and reviewed now. The
structure-only scaffold does not authorize package bootstrap, framework
selftests, SUT suites, launcher, or runtime; their statuses are independent:

- package bootstrap: `NOT_READY`;
- framework selftests: `NOT_READY`;
- SUT suite layers: `NOT_READY`;
- versioned launcher API: `NOT_READY`;
- runtime: `NOT_RUN`.

A runner source candidate exists only under `testframework/runner/**`, which
alone owns future launcher authority; it is not independently accepted and its
admission remains `NOT_READY`. `testframework/docker/**` contains only immutable
execution assets and container contracts. Required `images.lock.v3.json` is
absent; docker contains no launcher or alternate entrypoint. No supported
command, Docker-directory entrypoint, or direct Docker/Compose, host
test/validator, mutable-image, or ad-hoc-cleanup fallback exists. After
separate independent acceptance, the launcher must provide unique task-owned
names, CPU/RAM/PID limits, wall-clock timeout, isolated
`testframework/reports/<task-run>/` with an owner marker, and unconditional
exact-resource cleanup in `finally`.

## Atomic scope

Только planning/docs/inventory contract. Перемещение тестов, scaffold и CI
выполняются следующими фазами отдельными атомарными коммитами.
