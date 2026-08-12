# PLAN: Phase 134 — backend work и realtime waves

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Перенести genuine Go units work/realtime в `backend/tests/**` и одновременно
заменить component/integration rows Python-тестами корневого framework.

## Depends on

- Phase 133 PASS.

## Success criteria

- [ ] Все rows шести доменов имеют accepted Go-unit target или выполненный
  Python replacement; простой future-candidate link не считается parity.
- [ ] Legacy/native domain suites green с равными required behaviors и skips.
- [ ] Frozen Go coverage не просел без разобранной причины.
- [ ] Legacy sources и hooks остаются до Phase 141.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-212 | Tracker/project/sprint/tms migration и API/integration candidate mapping | be1 | be-ver | `backend/tests/{tracker,project,sprint,tms}/**`, `.planning/audits/test-rebase/inventory.json`, `.planning/evidence/test-rebase/134/work/**` | Phase 133 | 1–1.5д |
| TASK-213 | Messenger/notification migration, Redis/NATS classification и parity evidence | be2 | be-ver | `backend/tests/{messenger,notification}/**`, `.planning/audits/test-rebase/hooks.json`, `.planning/evidence/test-rebase/134/realtime/**` | Phase 133 | 1–1.5д |
| TASK-228 | Work/realtime API/integration/contract component parity в Python | qa-auto2 | qa-auto-ver | `testframework/{api,integration,contract}/{tracker,project,sprint,tms,messenger,notification}/**` | TASK-193/205/212/213 | 1–2д |

## Волны

- Wave 1: TASK-212 и TASK-213 unit batches.
- Wave 2: TASK-228 получает frozen classification/assertion handoff.
- Wave 3: combined legacy/native/Python domain/race/coverage reconciliation.

## Риски

| Риск | Mitigation |
|---|---|
| TMS/tracker объём превышает фазу | Ledger-first; stop/replan вместо partial PASS |
| Redis/NATS test ошибочно назван unit | Любой broker/network row получает TASK-228 Python integration target |
| Source-path guard сломан move | Обновить только path resolution; invariant/assertion не ослаблять |

## Verification

- `be-ver`: legacy/native counts, race, coverage, hook consumer diff.
- `qa-auto-ver`: phase-scope `unmapped = 0`, replacement links valid.
