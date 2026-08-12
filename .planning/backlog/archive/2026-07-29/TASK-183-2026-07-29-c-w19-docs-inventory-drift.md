# 🧹 C-W19: docs inventory recount + AGENTS/spec drift

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: [`C-W19-fe-ai-platform.md`](../../../audits/total-2026-07/C-W19-fe-ai-platform.md) §3 Docs  
> Исполнитель: **docs-writer** · Finding ID: `C-W19-DOC-001`

## [OUTCOME]

Канонические inventory-цифры BE `*_test.go` и Vitest files совпадают с shell recount HEAD; AGENTS.md / system_specification / user_needs / vision не держат stale 515/99/94/bbdaf16/Phase16 webcam.

## [ROLE]

- Role: docs-writer · Handoff: docs-ver / product-ver wording

## [SCOPE AND CONSTRAINTS]

- Allowed: `docs/README.md`, `test_plan.md`, `TESTING.md`, `testframework_strategy.md`, `TROUBLESHOOTING.md`, `AGENTS.md`, `docs/system_specification.md`, `docs/user_needs_analysis_2026.md`, `docs/vision_and_scope.md`, `docs/competitive_audit.md` (cosmetic); README backlog + этот файл
- Do not touch: product code
- Shell recount обязателен в evidence; дата measure

## [REFERENCES]

- После TASK-132: 805 уже drift → **807**; Vitest 167→**170**
- AGENTS 515 / Vitest 99; system_spec 94; user_needs bbdaf16; vision webcam Phase 16
- Live recount 2026-07-29 HEAD `d868ba77` (`cursor-grok-4.5`): BE **812** / Vitest **179** (61 `.ts` + 118 `.tsx`)

## [ACCEPTANCE]

- Одна цифра BE + одна Vitest во всех затронутых docs = live recount
- Anti-stale: нет `515`/`99/99`/`bbdaf16` как live HEAD inventory
- Webcam wording согласован с shipped (не Phase 16 IN)
- Stop: FIXED · «На проверке»

## [EVIDENCE] — 2026-07-29

```
BE=812  (backend/testframework/**/*_test.go)
VITEST=179  (61 .test.ts + 118 .test.tsx under stages/unit/frontend-test)
HOOKS=39  (backend/internal/**/export_hooks.go)
HEAD=d868ba776dd1d1ac3521c3c355339a2390723f07 (short d868ba77)
branch=cursor-grok-4.5
```

Обновлены: AGENTS.md, docs/{README,test_plan,TESTING,testframework_strategy,TROUBLESHOOTING,system_specification,user_needs_analysis_2026,vision_and_scope,competitive_audit}.md, STATE.md, backlog README.

TASK-008 не закрывали.

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 1 refill / be1 · criteria only (no mutate) · 2026-07-30 · **docs-only → N/A code mutate**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| C-W19-DOC-001 | **docs-only** — inventory цифры в `AGENTS.md` + `docs/{README,test_plan,TESTING,testframework_strategy,TROUBLESHOOTING,system_specification,user_needs_analysis_2026,vision_and_scope,competitive_audit}.md` = live shell recount HEAD | docs | N/A hubtest RED — evidence = shell recount ≡ docs (см. `C-VERIFY-docs-183-2026-07-29.md`) | **N/A code mutate** — не ломать prod; verifier = recount vs canon; stale `515`/`99`/`bbdaf16` as live HEAD → FAIL |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 001 | Shell recount + doc text | **docs-only** |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: C-W19-DOC-001`

### Gaps

- Docs-only: **C-W19-DOC-001** → **N/A** code mutation (inventory/docs sync).
- C-VERIFY 2026-07-29: live BE=821 / Vitest=180 vs docs 812/179 — re-sync before FIXED; не путать с code-mutation PROVEN.

## [MUTATION RESULT] — 2026-07-30 W13 slot 1 refill (be1)

- Verdict: **N/A** (docs-only) — **не** fake-PROVEN.
- Report: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be1-docs-183.md`
- Status → **Закрыта**.
