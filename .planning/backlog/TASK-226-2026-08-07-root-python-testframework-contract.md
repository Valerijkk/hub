# TASK-226 — Root Python testframework contract

**Фаза:** 128 · **Приоритет:** P1 · **Статус:** VERIFIED
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
Вложенного `testframework/python/**` в target нет. Legacy TypeScript E2E может
существовать только во время dual-run и удаляется после Python parity.

## Acceptance criteria

- [ ] Канон, roadmap, phase plans и onboarding docs используют один root
  `testframework/**`, без target-ссылок на `testframework/python/**`.
- [ ] Python layout: root package/config + `core/`, `api/`, `integration/`,
  `contract/`, `e2e/`, `load/`, `docker/`, `reports/`.
- [ ] `go-component` rows не попадают в `backend/tests/**`: они получают
  target в `testframework/api|integration|contract/**` либо отдельное
  доказанное преобразование в настоящий unit.
- [ ] Frontend target не сохраняет служебный legacy-сегмент `domains/`:
  `frontend/tests/<domain>/**` + `frontend/tests/support/**`. Source/docs/compile
  guards не маскируются под RTL unit и получают отдельный Python/static target.
- [ ] Inventory generator направляет legacy Playwright TS в
  `testframework/e2e/**` и считает Python по всему root framework.
- [ ] Ledger учитывает не только test specs, но и legacy support-файлы, чтобы
  финальная очистка не оставила TypeScript runtime внутри `testframework/**`.
- [ ] Ledger пересоздан, schema/check/self-test проходят, `unmapped = 0`.
- [ ] Current-vs-target distinction сохранён; файлы legacy не удалены этой
  задачей.
- [ ] `pro-qa-ai/**` не читался и не изменялся.

## Atomic scope

Только planning/docs/inventory contract. Перемещение тестов, scaffold и CI
выполняются следующими фазами отдельными атомарными коммитами.
