# PLAN: Phase 134 — backend work и realtime waves

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Перенести Go-native work/realtime tests в `backend/tests/**` с полным
disposition ledger для tracker/project/sprint/tms/messenger/notification.

## Depends on

- Phase 133 PASS.

## Success criteria

- [ ] Все rows шести доменов имеют accepted Go target или связанный Python
  replacement candidate; `unmapped = 0` в scope фазы.
- [ ] Legacy/native domain suites green с равными required behaviors и skips.
- [ ] Frozen Go coverage не просел без разобранной причины.
- [ ] Legacy sources и hooks остаются до Phase 141.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-212 | Tracker/project/sprint/tms migration и API/integration candidate mapping | be1 | be-ver | `backend/tests/{tracker,project,sprint,tms}/**`, `.planning/audits/test-rebase/inventory.json`, `.planning/evidence/test-rebase/134/work/**` | Phase 133 | 1–1.5д |
| TASK-213 | Messenger/notification migration, Redis/NATS classification и parity evidence | be2 | be-ver | `backend/tests/{messenger,notification}/**`, `.planning/audits/test-rebase/hooks.json`, `.planning/evidence/test-rebase/134/realtime/**` | Phase 133 | 1–1.5д |

## Волны

- Wave 1: TASK-212 и TASK-213 параллельно.
- Wave 2: combined domain/race/coverage reconciliation.

## Риски

| Риск | Mitigation |
|---|---|
| TMS/tracker объём превышает фазу | Ledger-first; stop/replan вместо partial PASS |
| Redis/NATS test ошибочно назван unit | Real broker/network rows остаются legacy до Phase 136 |
| Source-path guard сломан move | Обновить только path resolution; invariant/assertion не ослаблять |

## Verification

- `be-ver`: legacy/native counts, race, coverage, hook consumer diff.
- `qa-auto-ver`: phase-scope `unmapped = 0`, replacement links valid.
