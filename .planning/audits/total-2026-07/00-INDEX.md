# Total Audit 2026-07 — индекс

> TASK-008 · A ✅ · **A2 ✅** ([A2-PLAN.md](A2-PLAN.md) · [00-STAGE-A2-COMPLETE.md](00-STAGE-A2-COMPLETE.md)) · B P0–P3 ✅ (W21) · C ⚠️ · FINAL: [00-FINAL.md](00-FINAL.md) · 2026-07-28

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
| A2-W1 | tracker+sprint+project | OK | 163 | 0 | 0 | 5 | 3 | **8** |
| A2-W2 | tms+bases | OK | 109 | 0 | 0 | 23 | 5 | **28** |
| A2-W3 | messenger+ws+calls | OK | 131 | 0 | 0 | 12 | 4 | **16** |
| A2-W4 | platform+apigateway | OK | 83 | 0 | 1 | 4 | 8 | **13** |
| A2-W5 | ai+search | OK | 96 | 0 | 1 | 18 | 13 | **32** |
| A2-W6 | auth+rbac+audit | OK | 92 | 0 | 1 | 22 | 8 | **31** |
| A2-W7 | remaining backend | OK | 284 | 0 | 4 | 33 | 4 | **41** |
| A2-W8 | ent/schema | OK | 83 | 0 | 14 | 21 | 0 | **35** |
| A2-W9 | frontend/components | OK | 127 | 0 | 0 | 4 | 7 | **11** |
| A2-W10 | routes+modules | OK | 144 | 0 | 2 | 25 | 2 | **29** |
| A2-W11 | api/stores/hooks/lib+docs | OK | 93 | 0 | 1 | 6 | 4 | **11** |

**Итого Stage A2:** файлов **1405** · open Σ (сырые, до дедупа) **255** (P0=0 · P1=24 · P2=173 · P3=58).  
Сводки: `A2-W{n}-SUMMARY.md` · финал: [`00-STAGE-A2-COMPLETE.md`](00-STAGE-A2-COMPLETE.md).

## Стадии B/C

| Волна | Статус |
|-------|--------|
| W12 постановка | **COMPLETE** — [`W12-COMPLETE.md`](W12-COMPLETE.md) · TASK-009…056 + P3 057–078 |
| W13–W17 фиксы | **COMPLETE** (P0–P2 код; P3 backlog) |
| W18 мутация | **PARTIAL** — 39 PASS / 9 FAIL — [`W18-SUMMARY.md`](W18-SUMMARY.md) |
| W19 cross | **COMPLETE** — [`W19-SUMMARY.md`](W19-SUMMARY.md) |
| W20 FINAL | **COMPLETE** — [`00-FINAL.md`](00-FINAL.md) · hubtest local PASS · COV vs 81.9% NOT VERIFIED |

**TASK-008 не закрывать** до reopen 014/079/045… и method-comparable coverage.
