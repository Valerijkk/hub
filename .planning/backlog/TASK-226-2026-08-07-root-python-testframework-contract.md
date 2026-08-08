# TASK-226 — Root Python testframework contract

**Фаза:** 128 · **Приоритет:** P1 · **Статус:** OPEN — remediation verification pending
**Исполнитель:** `arch` + `qa-auto1` · **Проверка:** `arch-ver` + `qa-auto-ver`
**Depends on:** `TASK-202`

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

- [ ] Канон, roadmap, phase plans и onboarding docs используют один root
  `testframework/**`, без target-ссылок на `testframework/python/**`.
- [ ] Python layout: root package/config + `core/`, `api/`, `integration/`,
  `contract/`, `e2e/`, `load/`, `docker/`, `reports/`.
- [ ] `go-component` rows не попадают в `backend/tests/**` и удаляются без replacement:
  они входят в exact delete manifest. Новые Python API/integration/contract
  suites проектируются с нуля по актуальным требованиям, рискам и TC IDs.
- [ ] Frontend target не сохраняет служебный legacy-сегмент `domains/`:
  `frontend/tests/<domain>/**` + `frontend/tests/support/**`. Все 24 legacy
  static-contract Vitest rows и их source/docs/compile guards удаляются по
exact manifest без порта. Любой будущий `testframework/tests/contract/frontend/**`
  suite является независимым greenfield design по current requirements/risks,
  а не replacement/translation этих строк.
- [ ] Inventory/cutover manifest относит legacy Playwright TS specs, support и
  entrypoints только в exact `delete_legacy` set без replacement target path.
Новый Python E2E catalog в `testframework/tests/e2e/**` создаётся отдельно из
  актуальных requirements, risks и stable TC IDs.
- [ ] Ledger учитывает не только test specs, но и legacy support-файлы, чтобы
  финальная очистка не оставила TypeScript runtime внутри `testframework/**`.
- [ ] Ledger пересоздан, schema/check/self-test проходят, `unmapped = 0`.
- [ ] Current-vs-target distinction сохранён; файлы legacy не удалены этой
  задачей.
- [ ] `pro-qa-ai/**` не читался и не изменялся.

## Runtime readiness

Документный contract можно исправлять и рецензировать сейчас. Любая повторная
validator/runtime acceptance остаётся `NOT_READY`, пока
`testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` и
`testframework/docker/images.lock.json` не будут независимо приняты
`qa-auto-ver`. Прямой Docker/Compose, host test/validator, mutable-image и
ad-hoc cleanup не являются fallback. Launcher обязан задавать уникальные
task-owned имена, CPU/RAM/PID limits, wall-clock timeout, isolated
`testframework/reports/<task-run>/` с owner marker и unconditional
exact-resource cleanup в `finally`.

## Atomic scope

Только planning/docs/inventory contract. Перемещение тестов, scaffold и CI
выполняются следующими фазами отдельными атомарными коммитами.
