# PLAN: Phase 131 — frontend domain migration

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Перенести оставшиеся Vitest/RTL тесты в `frontend/tests/**`, сохранив dual-run,
static gates и behavioral parity полного baseline inventory.

## Depends on

- Phase 130 PASS (native harness и platform+auth wave).

## Success criteria

- [ ] Tracker/messenger/tms/calls и все оставшиеся frontend domains mapped.
- [ ] Native collected count равен принятому Phase 128 frontend baseline
  (snapshot: 180), либо ledger содержит verifier-approved split/consolidation.
- [ ] Legacy и native suites green с нулём неожиданных skips и accepted coverage delta.
- [ ] Legacy исходники и linker ещё не удалены.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-207 | Tracker, messenger, tms и calls migration по SUT с parity artifacts | fe1 | fe-ver | `frontend/tests/{tracker,messenger,tms,calls}/**`, `.planning/audits/test-rebase/inventory.json`, `.planning/evidence/test-rebase/131/wave-a/**` | Phase 130 | 1–1.5д |
| TASK-208 | Admin, ai, dashboard, hr, notifications, search, v4 и residual migration | fe2 | fe-ver | `frontend/tests/{admin,ai,dashboard,hr,notifications,search,v4}/**`, `frontend/tests/support/**`, `.planning/evidence/test-rebase/131/wave-b/**` | Phase 130 | 1–1.5д |

## Волны

- Wave 1: TASK-207 и TASK-208 параллельно; ownership directories не пересекаются.
- Wave 2: общий native regression, inventory reconciliation и verifier sign-off.

## Риски

| Риск | Mitigation |
|---|---|
| Broad `v4` tests имеют несколько SUT | Split только с сохранёнными IDs и accepted mapping; не терять assertions |
| Path-sensitive source guards | Обновить ожидаемый production path, не ослаблять guard |
| Механический объём >3 дней | Стоп и replan; Phase 131 не закрывать частичным inventory |

## Verification

- `fe-ver`: native type/lint/unit/coverage full green; legacy full green.
- `qa-auto-ver`: frontend `unmapped = 0`, parity required = 0, unexpected skip = 0.
