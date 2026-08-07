# PLAN: Phase 130 — frontend native foundation

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Поднять независимый `frontend/tests/**` harness с раздельным legacy/native
discovery и перенести platform+auth wave без удаления исходников.

## Depends on

- Phase 128 PASS.
- Phase 129 не блокирует: Python и frontend native — независимые ветви.

## Success criteria

- [ ] `frontend/tests/**` входит в Vitest, TypeScript, Biome и coverage.
- [ ] Legacy и native scripts запускаются отдельно и публикуют comparable counts.
- [ ] Platform+auth inventory rows имеют replacement result и verifier parity.
- [ ] Junction/linker сохраняется, пока legacy frontend tree имеет consumers.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-198 | Native harness, support layer, test typecheck/lint/coverage и dual discovery commands | fe1 | fe-ver | `frontend/tests/support/**`, `frontend/vite.config.ts`, `frontend/tsconfig.test.json`, `frontend/biome.json`, `frontend/package.json` | Phase 128 | 1д |
| TASK-206 | Platform+auth files переносятся по ledger с сохранением names/assertions/source guards | fe2 | fe-ver | `frontend/tests/platform/**`, `frontend/tests/auth/**`, `.planning/audits/test-rebase/inventory.json`, `.planning/evidence/test-rebase/130/**` | TASK-198 | 1–1.5д |

## Волны

- Wave 1: TASK-198.
- Wave 2: TASK-206 (platform и auth можно исполнять параллельно без общих files).

## Риски

| Риск | Mitigation |
|---|---|
| Двойная discovery даёт дубликаты | Раздельные `test:legacy`/`test:native`; compare artifacts, не один mixed run |
| Tests вне static gates | Native include одновременно во Vitest, test-tsconfig и Biome |
| `@testframework`/renderer ломается | Канонический `frontend/tests/support`; legacy alias сохраняется до Phase 141 |

## Verification

- `fe-ver`: legacy/native counts, typecheck, Biome, coverage diff и deliberate
  failing native test.
- `qa-auto-ver`: inventory rows platform+auth имеют `PARITY_ACCEPTED`.
