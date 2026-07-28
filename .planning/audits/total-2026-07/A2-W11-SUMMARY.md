# A2 W11 SUMMARY — api/stores/hooks/lib + docs

> TASK-008 · Stage A2 · **93/93** · 2026-07-28  
> Зеркало Stage A: `W11-SUMMARY.md` · Assignments: `W11-ASSIGNMENTS.md`

## Агенты

| Агент | Отчёт | Files | NEW | REGRESS | RESIDUAL | FIXED-OK | Open Σ |
|-------|-------|------:|----:|--------:|---------:|---------:|-------:|
| A1 | `A2W11-1-fe-lib-docs.md` | 19 | 0 | 0 | 0 | 10 | **0** |
| A2 | `A2W11-2-fe-lib-docs.md` | 19 | 0 | 0 | 2 | 13 | **2** |
| A3 | `A2W11-3-fe-lib-docs.md` | 19 | 1 | 0 | 2 | 16 | **3** |
| A4 | `A2W11-4-fe-lib-docs.md` | 19 | 0 | 0 | 0 | 19 | **0** |
| A5 | `A2W11-5-fe-lib-docs.md` | 17 | 0 | 0 | 2 | 22 | **2** |
| **Σ** | | **93** | **1** | **0** | **6** | **80** | **7** |

## Open by severity

| Severity | NEW | REGRESS | RESIDUAL | Σ |
|----------|----:|--------:|---------:|--:|
| P0 | 0 | 0 | 0 | **0** |
| P1 | 0 | 0 | 1 | **1** |
| P2 | 1 | 0 | 4 | **5** |
| P3 | 0 | 0 | 1 | **1** |
| **Σ** | **1** | **0** | **6** | **7** |

## Топ открытых

1. **A2W11-5-001 (RESIDUAL P1)** — `vision` G-07 «US-WIKI реализованы» ↔ KPI wiki planned.
2. **A2W11-3-001 (RESIDUAL P2)** — `use-presence` без `isError`; ложное закрытие в TASK-047 TODO.
3. **A2W11-2-001 / A2W11-3-002 (RESIDUAL P2)** — JWT в `sessionStorage` + `/ws?token=` (нужен ADR/BE).
4. **A2W11-2-002 / A2W11-5-002 (RESIDUAL P2)** — TZ offset для задач vs cursor в коде.

## CLOSED by TASK-123 (2026-07-28)

- A2W11-4-001 / A2W11-4-002 — BUSINESS webcam consistency + dead `useRooms` removed
- A2W11-5-003 / A2W11-5-004 — NF-11 → `wiring_fiber.go`; REST summary ~252 OpenAPI

## REGRESS

**0** — ранее закрытые W11 P1/P2/P3 в api/client/calendar/calls/stores/hooks/docs test-path **не откатились**.

## Вердикт волны

Stage A2 W11 **закрыта** (read-only). Код api-клиента и большинство stores/hooks после B/C — **FIXED-OK**. TASK-123 закрыл dead `useRooms` + docs drift (webcam/NF-11/REST). Остаток: ADR-security (tokens/WS), docs PO (G-07, TZ pagination), false-close presence `isError`.
