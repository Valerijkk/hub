# A2 W11 SUMMARY — api/stores/hooks/lib + docs

> TASK-008 · Stage A2 · **93/93** · 2026-07-28  
> Зеркало Stage A: `W11-SUMMARY.md` · Assignments: `W11-ASSIGNMENTS.md`

## Агенты

| Агент | Отчёт | Files | NEW | REGRESS | RESIDUAL | FIXED-OK | Open Σ |
|-------|-------|------:|----:|--------:|---------:|---------:|-------:|
| A1 | `A2W11-1-fe-lib-docs.md` | 19 | 0 | 0 | 0 | 10 | **0** |
| A2 | `A2W11-2-fe-lib-docs.md` | 19 | 0 | 0 | 2 | 13 | **2** |
| A3 | `A2W11-3-fe-lib-docs.md` | 19 | 1 | 0 | 2 | 16 | **3** |
| A4 | `A2W11-4-fe-lib-docs.md` | 19 | 1 | 0 | 1 | 17 | **2** |
| A5 | `A2W11-5-fe-lib-docs.md` | 17 | 0 | 0 | 4 | 20 | **4** |
| **Σ** | | **93** | **2** | **0** | **9** | **76** | **11** |

## Open by severity

| Severity | NEW | REGRESS | RESIDUAL | Σ |
|----------|----:|--------:|---------:|--:|
| P0 | 0 | 0 | 0 | **0** |
| P1 | 0 | 0 | 1 | **1** |
| P2 | 1 | 0 | 5 | **6** |
| P3 | 1 | 0 | 3 | **4** |
| **Σ** | **2** | **0** | **9** | **11** |

## Топ открытых

1. **A2W11-3-001 (RESIDUAL P2)** — `use-presence` без `isError`; ложное закрытие в TASK-047 TODO.
2. **A2W11-2-001 / A2W11-3-002 (RESIDUAL P2)** — JWT в `sessionStorage` + `/ws?token=` (нужен ADR/BE).
3. **A2W11-2-002 / A2W11-5-002 (RESIDUAL P2)** — TZ offset для задач vs cursor в коде.
4. **A2W11-4-001 (NEW P2)** — `BUSINESS.md` внутренний drift webcam/Phase 16.

Закрыто docs: **A2W11-5-001** → FIXED (TASK-105, G-07 Partial).

## REGRESS

**0** — ранее закрытые W11 P1/P2/P3 в api/client/calendar/calls/stores/hooks/docs test-path **не откатились**.

## Вердикт волны

Stage A2 W11 **закрыта** (read-only). Код api-клиента и большинство stores/hooks после B/C — **FIXED-OK**. Остаток: ADR-security (tokens/WS), docs TZ pagination (A2W11-5-002), false-close presence `isError`, мелкий docs/API dead-export. **G-07 wiki KPI drift закрыт** (TASK-105).
