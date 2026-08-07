# PLAN: Phase 132 — backend native foundation

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Подготовить `backend/tests/**`, отделить Go-native support от service-backed
fixtures и доказать auth pilot без предположения, что новый путь устраняет
`hubtest` export hooks.

## Depends on

- Phase 128 PASS (inventory, baseline, taxonomy/coverage contract).

## Success criteria

- [ ] Все Go rows baseline inventory классифицированы по runtime dependency;
  shared support имеет явного владельца/target.
- [ ] `backend/tests/**` запускается отдельной командой, legacy suite не изменён.
- [ ] Auth pilot имеет old/new parity и одинаковую coverage instrumentation.
- [ ] Для каждого auth `Export*` зафиксирован consumer; ни один hook не удалён
  только из-за смены каталога.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-197 | Go semantic ledger и native support layout; env/factory/steps/mongotest разделены на Go-unit vs Python candidates | be1 | be-ver | `backend/tests/support/**`, `.planning/audits/test-rebase/{inventory.json,hooks.json}`, `.planning/evidence/test-rebase/132/ledger.md` | Phase 128 | 1д |
| TASK-209 | Auth pilot в `backend/tests/auth/**` с отдельным runner, parity, coverage diff и hook proof | be2 | be-ver | `backend/tests/auth/**`, `backend/internal/auth/export_hooks*.go`, `.planning/evidence/test-rebase/132/auth-pilot/**` | TASK-197 | 1–1.5д |

## Волны

- Wave 1: TASK-197.
- Wave 2: TASK-209.

## Риски

| Риск | Mitigation |
|---|---|
| External package не видит private symbols | Сохранить `-tags hubtest` и только доказанные minimal hooks |
| Integration helper ошибочно уедет в unit | Классифицировать по DB/broker/network dependency, не по каталогу |
| Двойной suite и data collision | Legacy/native запускаются раздельно с уникальными DB namespaces |

## Verification

- `be-ver`: legacy auth + native auth race/coverage, hook consumer ledger и
  deliberate failing auth sentinel.
- `qa-auto-ver`: auth inventory rows получают `PARITY_ACCEPTED`.
