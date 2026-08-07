# PLAN: Phase 128 — status и baseline recovery

**Статус:** PLANNED · **Оценка:** 2–3 дня · **Блокирует:** Phase 129–141

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Получить воспроизводимый live-status, полный migration ledger и единый
taxonomy/coverage contract до любого scaffold или перемещения тестов.

## Depends on

- Нет фазовых зависимостей.
- Нужны Go 1.25, Node 22, Python runtime и Docker для полного baseline; их
  отсутствие фиксируется как BLOCKED/NOT_RUN, но не подменяется PASS.

## Success criteria

- [ ] Inventory охватывает baseline 828 Go / 180 Vitest / 20 E2E specs / 53
  cases / 0 Python либо объясняет датированный drift; `unmapped = 0`.
- [ ] Сохранён один полностью green coverage baseline с command, SHA, versions,
  services, pass/fail/skip, profile hash и per-package values; Go overall ≥82%.
  Значение ниже 82% блокирует фазу и любые перемещения, пока founder заранее
  не утвердит новый denominator/threshold в TASK-204.
- [ ] Каждая строка имеет taxonomy, target, coverage-bearing и hook consumers.
- [ ] До принятия Phase 128 не изменены test paths, runners и legacy contents.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-202 | Machine-readable inventory ledger с checksum, test IDs/counts, old→target mapping и live recount | qa-auto1 | qa-auto-ver | `.planning/audits/test-rebase/{inventory.py,schema.json,inventory.json,README.md}` | — | 0.75д |
| TASK-203 | Канонический backend/frontend/E2E baseline на одном SHA; failed/skip/runtime evidence не теряется | devops-engineer | qa-auto-ver | `.planning/audits/test-rebase/baseline/**`, `.planning/COVERAGE.md` | TASK-202 | 1д |
| TASK-204 | Утверждён taxonomy, parity, Go/frontend coverage и hook-consumer contract | arch | arch-ver | `.planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md`, `.planning/audits/test-rebase/schema.json` | TASK-202, TASK-203 | 0.5д |

## Волны

- Wave 1: TASK-202.
- Wave 2: TASK-203.
- Wave 3: TASK-204 и независимая проверка; только после PASS разблокируются 129–141.

## Риски

| Риск | Mitigation |
|---|---|
| Docker/Python недоступны | Сохранить точный blocker; повторить на чистом/CI runner, не продолжать migration |
| Coverage числа несравнимы | Один command+environment contract; исторические 83.0/81.9/80.2 только context |
| Inventory теряет source-guard/skip/hook | Полный путь, checksum, test count, skip и consumer fields обязательны schema gate |

## Verification

- `qa-auto-ver`: schema validation, recount и `unmapped = 0`.
- `be-ver`: воспроизводимость Go command/profile и hook counts.
- `team-lead`: PASS или явный BLOCKED; частичный baseline не принимается.
