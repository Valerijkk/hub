# A2-W5 SUMMARY — ai + search (Stage A2)

> TASK-008 · Stage A2 · 2026-07-28  
> HEAD: `0d6a659fe9dd56a473cea0ace40ccae280970cbc`  
> Assignments: [`W5-ASSIGNMENTS.md`](W5-ASSIGNMENTS.md) · Stage A: [`W5-SUMMARY.md`](W5-SUMMARY.md)  
> Отчёты: `A2W5-1` … `A2W5-5`

## Покрытие

| Агент | Отчёт | Файлов | Прочитано | Пропущено |
|-------|-------|-------:|----------:|----------:|
| A1 ai handlers | [`A2W5-1-ai-handlers.md`](A2W5-1-ai-handlers.md) | 11 | 11 | 0 |
| A2 ai adapters | [`A2W5-2-ai-adapters.md`](A2W5-2-ai-adapters.md) | 20 | 20 | 0 |
| A3 ai service | [`A2W5-3-ai-service.md`](A2W5-3-ai-service.md) | 29 | 29 | 0 |
| A4 search core | [`A2W5-4-search-core.md`](A2W5-4-search-core.md) | 22 | 22 | 0 |
| A5 search tests | [`A2W5-5-search-tests.md`](A2W5-5-search-tests.md) | 14 | 14 | 0 |
| **Σ** | | **96** | **96** | **0** |

## Находки по агентам

| Агент | P0 | P1 | P2 | P3 | Σ | NEW | REGRESS | RESIDUAL |
|-------|---:|---:|---:|---:|--:|----:|--------:|---------:|
| A1 | 0 | 0 | 4 | 1 | 5 | 0 | 2 | 3 |
| A2 | 0 | 1 | 4 | 3 | 8 | 0 | 4 | 4 |
| A3 | 0 | 0 | 5 | 4 | 9 | 1 | 3 | 5 |
| A4 | 0 | 0 | 2 | 2 | 4 | 1 | 0 | 3 |
| A5 | 0 | 0 | 3 | 3 | 6 | 0 | 0 | 6 |
| **Σ** | **0** | **1** | **18** | **13** | **32** | **2** | **9** | **21** |

Сырых находок A2: **32**. Stage A (W5) было **68** — существенная доля P1 закрыта фиксами B/W21.

## Теги (смысл)

| Tag | Count | Суть |
|-----|------:|------|
| NEW | 2 | `code_review` access==nil fail-open; deleteFilter без sentinel map |
| REGRESS | 9 | Ложные закрытия P2-BE-07/10/17: AssistJQL, AI RL, RPM fail-open/ZAdd, stream timeout/fallback, composer slice, IncrementUsage, settings Upsert |
| RESIDUAL | 21 | Не добито / U-классы / слабые тесты / Meili SafeDial / filter HTTP status |

## FIXED-OK (кратко vs A1 P1)

| Класс | Итог |
|-------|------|
| P1-BE-29 content policy all roles | FIXED-OK |
| P1-BE-30 offline substring bypass | FIXED-OK |
| P1-BE-31 bot/issue/review authz | FIXED-OK (wiring SetAccessChecker; nil-access API — NEW residual) |
| P1-BE-32 search channel soft-delete + membership | FIXED-OK |
| P1-BE-U01 Ollama↔SafeDial | RESIDUAL (accepted U) |

## Топ ложных закрытий (REGRESS)

1. **provider_rpm** fail-open + ZAdd-before-check — в P2-BE-10, но TASK-040 не трогал `ai/`  
2. **AssistJQL** error map + **admin/BYOK AIRateLimit** — в P2-BE-07/17, код не держит  
3. **composer_suggest** byte trim + **IncrementUsage** race + stream **90s Timeout** / fallback-after-delta — заявлены P2-BE-17  

## Drift

- `backend/internal/ai/project_access.go` — вне `W5-ASSIGNMENTS` (добор TASK-008); живёт и wired.  
- Новые hubtest-файлы (`*_task037*`, `*_task044*`, `stream_cancel_*`) не в assignments — не считались в 96.

## Вердикт волны

P0 нет. Единственный P1 — **RESIDUAL** Ollama/SafeDial contradiction (U01). Главный риск A2-W5 — **9 REGRESS** (ложные «Исправлена» на AI stream/RPM/AssistJQL). Search P1 закрыты; остаток — HTTP status filters, Meili dial, hollow tests.
