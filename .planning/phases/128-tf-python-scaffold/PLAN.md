# PLAN: Phase 128 — status и baseline recovery

**Статус:** BLOCKED — baseline red, root-layout correction active · **Оценка:** 4–6 дней · **Блокирует:** Phase 129–141

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Получить воспроизводимый live-status, полный migration ledger и единый
taxonomy/coverage contract до любого scaffold или перемещения тестов.

## Depends on

- Нет фазовых зависимостей.
- Нужны Go 1.25, Node 22, Python runtime и Docker для полного baseline; их
  отсутствие фиксируется как BLOCKED/NOT_RUN, но не подменяется PASS.

## Success criteria

- [ ] Inventory охватывает 1,047 rows: 830 Go / 3,667 tests (594 unit + 236
  component: 31 API + 205 integration), 180 Vitest / 519 cases (156 unit + 24
  static contract) + 3 support, 20 E2E specs / 53 cases + 14 support,
  39 canonical hooks / 0 Python; `unmapped = 0`.
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
| TASK-221 | Portable backend source/schema guards + no-network GitLab fixture | be1 | be-ver | 10 allowlisted Go test files | TASK-203 red attempt | 0.5д |
| TASK-222 | Актуальные frontend fixtures + semantic pipeline tokens | fe1 | fe-ver | 3 Vitest files + `pipeline-status.tsx` | TASK-203 red attempt | 0.5д |
| TASK-223 | Biome format/semantic debt до global lint green | fe2 | fe-ver | scoped frontend lint allowlist | TASK-222 | 1–2д |
| TASK-224 | Docker application/auth seed contract + real E2E baseline | devops-engineer | qa-auto-ver | test entrypoints/overlay/evidence | TASK-203 red attempt | 1д |
| TASK-225 | Audit behavior tests восстанавливают overall ≥82% и audit ≥80% | be2 | be-ver | новый audit coverage-gate test | TASK-221 | 0.5д |
| TASK-226 | Founder layout correction: единый root `testframework/**` Python-only, без `testframework/python/**`; ledger и target paths пересчитаны | arch + qa-auto1 | arch-ver + qa-auto-ver | канон, roadmap, target docs, inventory generator/ledger | TASK-202 | 0.5д |
| TASK-204 | Утверждён taxonomy, parity, Go/frontend coverage и hook-consumer contract | arch | arch-ver | `.planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md`, `.planning/audits/test-rebase/inventory.schema.json` | TASK-202, TASK-203, TASK-226 | 0.5д |

## Волны

- Wave 1: TASK-202.
- Wave 2: первая попытка TASK-203 и evidence-backed triage.
- Wave 2.1: TASK-221/222/224 параллельно; TASK-223 после TASK-222; TASK-225 после TASK-221 coverage evidence.
- Wave 2.2: TASK-226 исправляет root-layout contract и пересчитывает ledger;
  параллельно повтор TASK-203 на исправленном baseline.
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
