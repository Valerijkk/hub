# 🧹 C-W19: docs inventory recount + AGENTS/spec drift

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: [`C-W19-fe-ai-platform.md`](../audits/total-2026-07/C-W19-fe-ai-platform.md) §3 Docs  
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
- Live recount 2026-07-29 HEAD `0dfdc9f0` (`cursor-grok-4.5`): BE **812** / Vitest **179** (61 `.ts` + 118 `.tsx`)

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
HEAD=0dfdc9f0496ddd095fcd2d0580e5bed651e16de2 (short 0dfdc9f0)
branch=cursor-grok-4.5
```

Обновлены: AGENTS.md, docs/{README,test_plan,TESTING,testframework_strategy,TROUBLESHOOTING,system_specification,user_needs_analysis_2026,vision_and_scope,competitive_audit}.md, STATE.md, backlog README.

TASK-008 не закрывали.
