# Stage A2 COMPLETE — повторный аудит после P3 W21

> TASK-008 · 2026-07-28 · READ-ONLY · код не менялся в A2  
> План: [`A2-PLAN.md`](A2-PLAN.md) · индекс: [`00-INDEX.md`](00-INDEX.md)

## Итог

| Метрика | Значение |
|---------|----------|
| Волн | **11/11** ✅ |
| Файлов | **1405 / 1405** |
| Open Σ (сырые, до дедупа) | **254** |
| P0 / P1 / P2 / P3 (open) | **0 / 24 / 172 / 58** |
| REGRESS (по сводкам волн) | см. ниже — есть в W4/W5/W6/W10; W7/W8 = **0** |

## По волнам

| Волна | Файлов | P0 | P1 | P2 | P3 | open Σ | Summary |
|-------|-------:|---:|---:|---:|---:|-------:|---------|
| A2-W1 | 163 | 0 | 0 | 5 | 3 | 8 | [A2-W1-SUMMARY](A2-W1-SUMMARY.md) |
| A2-W2 | 109 | 0 | 0 | 23 | 5 | 28 | [A2-W2-SUMMARY](A2-W2-SUMMARY.md) |
| A2-W3 | 131 | 0 | 0 | 12 | 4 | 16 | [A2-W3-SUMMARY](A2-W3-SUMMARY.md) |
| A2-W4 | 83 | 0 | 1 | 4 | 8 | 13 | [A2-W4-SUMMARY](A2-W4-SUMMARY.md) |
| A2-W5 | 96 | 0 | 1 | 18 | 13 | 32 | [A2-W5-SUMMARY](A2-W5-SUMMARY.md) |
| A2-W6 | 92 | 0 | 1 | 22 | 8 | 31 | [A2-W6-SUMMARY](A2-W6-SUMMARY.md) |
| A2-W7 | 284 | 0 | 4 | 32 | 4 | 40 | [A2-W7-SUMMARY](A2-W7-SUMMARY.md) |
| A2-W8 | 83 | 0 | 14 | 21 | 0 | 35 | [A2-W8-SUMMARY](A2-W8-SUMMARY.md) |
| A2-W9 | 127 | 0 | 0 | 4 | 7 | 11 | [A2-W9-SUMMARY](A2-W9-SUMMARY.md) |
| A2-W10 | 144 | 0 | 2 | 25 | 2 | 29 | [A2-W10-SUMMARY](A2-W10-SUMMARY.md) |
| A2-W11 | 93 | 0 | 1 | 6 | 4 | 11 | [A2-W11-SUMMARY](A2-W11-SUMMARY.md) |
| **Σ** | **1405** | **0** | **24** | **172** | **58** | **254** | |

## High-signal (фикс после дедупа)

Не backlog-ить U-классы без ADR / owner-approve migrate.

| Приоритет | ID / класс | Суть |
|-----------|------------|------|
| P1 REGRESS | A2-W4 | `CompleteTestRun` без `project_id` → TASK-012 fail-closed глушит TMS webhooks |
| P1 | A2-W5 | AI RPM / AssistJQL / rate-limit gaps (часть REGRESS) |
| P1 | A2-W6 | Telegram login positive test / legacy leftovers |
| P1 | A2-W7 | FR-INTR-07/08 author-delete+poll (U03); NF-09 size==0; Confirm scope; MaintenanceGate U04 |
| P1 | A2-W8 | VersionMixin gaps (~14/82); soft-delete / OnDelete residuals |
| P1 | A2-W10 / W11 | см. сводки (FE routes/modules + api) |
| NEW | A2W7-2-N01 P3 | Coffee `Pairs[]` triple-edges after FR-INTR-10 |
| — | A2-W8 | NEW=0 после финальной сводки; open=35 RESIDUAL (VersionMixin/OnDelete/trash) |

## Следующие шаги оркестратора

1. ~~**Дедуп**~~ → [`A2-DEDUP-2026-07-28.md`](A2-DEDUP-2026-07-28.md) · **14** code-TASK (**081…094**); сырые 254 ≠ 254 багов.
2. **Fix wave** приоритет: P1 REGRESS **081** → P1 **082–084** → P2 **085…093** → P3 **094**.
3. Deferred ADR/U* — **не** трогать без owner/ADR (VersionMixin, OnDelete, U03/U04, Telegram login, NF-09…).
4. **Не закрывать TASK-008** до remediation path + method-comparable COV vs 81.9%.
5. Naming задач: `TASK-<NNN>-<YYYY-MM-DD>-<slug>.md` — [`.cursor/rules/57-backlog-tasks.mdc`](../../.cursor/rules/57-backlog-tasks.mdc).
