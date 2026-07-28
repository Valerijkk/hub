# Total Audit 2026-07 — индекс

> TASK-008 · A ✅ · **A2 🔄** ([A2-PLAN.md](A2-PLAN.md)) · B P0–P3 ✅ (W21) · C ⚠️ · FINAL: [00-FINAL.md](00-FINAL.md) · 2026-07-28

| Волна | Зона | Статус | Файлов | P0 | P1 | P2 | P3 |
|-------|------|--------|-------:|---:|---:|---:|---:|
| W1 | tracker+sprint+project | закрыта | 163 | 1 | 19 | 28 | 18 |
| W2 | tms+bases | закрыта | 109 | 0 | 14 | 36 | 16 |
| W3 | messenger+ws+calls | закрыта | 131 | 1 | 17 | 32 | 18 |
| W4 | platform+apigateway | закрыта | 83 | 2 | 6 | 24 | 23 |
| W5 | ai+search | закрыта | 96 | 0 | 9 | 31 | 28 |
| W6 | auth+rbac+audit | закрыта | 92 | 0 | 6 | 28 | 21 |
| W7 | remaining backend | закрыта | 284 | 0 | 9 | 38 | 28 |
| W8 | ent/schema | закрыта | 83 | 0 | 19 | 40 | 12 |
| W9 | frontend/components | закрыта | 127 | 0 | 11 | 52 | 47 |
| W10 | routes+modules | закрыта | 144 | 0 | 14 | 64 | 20 |
| W11 | api/stores/hooks/lib+docs | закрыта | 93 | 0 | 17 | 49 | 14 |

**Итого Stage A:** файлов **1405** · находок сырых **812** (P0=4 · P1=141 · P2=422 · P3=245)

Финальный отчёт стадии A: [`00-STAGE-A-COMPLETE.md`](00-STAGE-A-COMPLETE.md)

## Stage A2 (повтор после P3 W21)

| Волна | Зона | Статус | Файлов | P0 | P1 | P2 | P3 | open Σ |
|-------|------|--------|-------:|---:|---:|---:|---:|-------:|
| A2-W1 | tracker+sprint+project | 🔄 | | | | | | |
| A2-W2 | tms+bases | 🔄 | | | | | | |
| A2-W3 | messenger+ws+calls | ✅ | 131 | 0 | 0 | 12 | 4 | **16** |
| A2-W4 | platform+apigateway | ✅ | 83 | 0 | 1 | 4 | 8 | **13** |
| A2-W5 | ai+search | ✅ | 96 | 0 | 1 | 18 | 13 | **32** |
| A2-W6 | auth+rbac+audit | 🔄 | | | | | | |
| A2-W7 | remaining backend | 🔄 | | | | | | |
| A2-W8 | ent/schema | 🔄 | | | | | | |
| A2-W9 | frontend/components | ✅ | 127 | 0 | 0 | 4 | 7 | **11** |
| A2-W10 | routes+modules | 🔄 | | | | | | |
| A2-W11 | api/stores/hooks/lib+docs | ✅ | 93 | 0 | 1 | 6 | 4 | **11** |

Сводки: [`A2-W3`](A2-W3-SUMMARY.md) · [`A2-W4`](A2-W4-SUMMARY.md) · [`A2-W5`](A2-W5-SUMMARY.md) · [`A2-W9`](A2-W9-SUMMARY.md) · [`A2-W11`](A2-W11-SUMMARY.md)  
A2-W4 P1 REGRESS: CompleteTestRun без `project_id` (TMS webhooks). A2-W5: REGRESS=9.

## Стадии B/C

| Волна | Статус |
|-------|--------|
| W12 постановка | **COMPLETE** — [`W12-COMPLETE.md`](W12-COMPLETE.md) · TASK-009…056 + P3 057–078 |
| W13–W17 фиксы | **COMPLETE** (P0–P2 код; P3 backlog) |
| W18 мутация | **PARTIAL** — 39 PASS / 9 FAIL — [`W18-SUMMARY.md`](W18-SUMMARY.md) |
| W19 cross | **COMPLETE** — [`W19-SUMMARY.md`](W19-SUMMARY.md) |
| W20 FINAL | **COMPLETE** — [`00-FINAL.md`](00-FINAL.md) · hubtest local PASS · COV vs 81.9% NOT VERIFIED |

**TASK-008 не закрывать** до reopen 014/079/045… и method-comparable coverage.
