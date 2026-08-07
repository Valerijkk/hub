# PLAN: Phase 131 — frontend domain migration

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Перенести только настоящие Vitest/RTL unit-тесты в
`frontend/tests/<domain>/**`, а source/docs/compile guards вывести в Python
contract/static lane корневого `testframework/**`.

## Depends on

- Phase 129 PASS (root Python contract lane).
- Phase 130 PASS (native harness и platform+auth wave).

## Success criteria

- [ ] Tracker/messenger/tms/calls и все оставшиеся frontend domains mapped.
- [ ] Каждый из 180 legacy Vitest-файлов классифицирован: настоящий unit имеет
  native target, source/docs/compile guard — Python/static replacement.
- [ ] Support-файлы находятся в `frontend/tests/support/**`; служебный legacy
  сегмент `domains/` в target не сохраняется.
- [ ] Legacy и native suites green с нулём неожиданных skips и accepted coverage delta.
- [ ] Legacy исходники и linker ещё не удалены.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-207 | Tracker, messenger, tms и calls migration по SUT с parity artifacts | fe1 | fe-ver | `frontend/tests/{tracker,messenger,tms,calls}/**`, `.planning/audits/test-rebase/inventory.json`, `.planning/evidence/test-rebase/131/wave-a/**` | Phase 130 | 1–1.5д |
| TASK-208 | Admin, ai, dashboard, hr, notifications, search, v4 и residual migration | fe2 | fe-ver | `frontend/tests/{admin,ai,dashboard,hr,notifications,search,v4}/**`, `frontend/tests/support/**`, `.planning/evidence/test-rebase/131/wave-b/**` | Phase 130 | 1–1.5д |
| TASK-230 | Frontend source/docs/compile guards → root Python contract/static lane с parity | qa-auto2 | qa-auto-ver | `testframework/contract/frontend/**`, inventory/evidence | TASK-193/205/207/208 | 1д |

## Волны

- Wave 1: TASK-207 и TASK-208 параллельно; ownership directories не пересекаются.
- Wave 2: TASK-230 заменяет все исключённые static/contract guards.
- Wave 3: общий native/Python/legacy regression, inventory reconciliation и verifier sign-off.

## Риски

| Риск | Mitigation |
|---|---|
| Broad `v4` tests имеют несколько SUT | Split только с сохранёнными IDs и accepted mapping; не терять assertions |
| Path-sensitive source guards ошибочно названы UI unit | Не переносить во frontend; TASK-230 создаёт path-independent contract replacement |
| Механический объём >3 дней | Стоп и replan; Phase 131 не закрывать частичным inventory |

## Verification

- `fe-ver`: native type/lint/unit/coverage full green; legacy full green.
- `qa-auto-ver`: frontend + contract `unmapped = 0`, parity required = 0,
  unexpected skip = 0.
