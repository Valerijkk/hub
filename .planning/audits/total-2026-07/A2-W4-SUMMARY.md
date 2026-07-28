# A2-W4-SUMMARY — platform + apigateway (re-audit)

> TASK-008 · Stage A2 · 2026-07-28 · READ-ONLY  
> Периметр: `W4-ASSIGNMENTS.md` · **83/83** файлов (1 assignment-файл удалён с диска — зафиксировано)

## Counts

| Агент | Отчёт | Файлов | NEW | REGRESS | RESIDUAL | FIXED-OK | P0 | P1 | P2 | P3 |
|-------|-------|-------:|----:|--------:|---------:|---------:|---:|---:|---:|---:|
| A1 http+mw | `A2W4-1-platform-http.md` | 15 | 0 | 0 | 3 | 9 | 0 | 0 | 0 | 3 |
| A2 nats/outbox | `A2W4-2-platform-nats-outbox.md` | 15 | 1 | 0 | 2 | 8 | 0 | 0 | 2 | 1 |
| A3 observability | `A2W4-3-platform-obs.md` | 14 | 1 | 0 | 0 | 10 | 0 | 0 | 0 | 1 |
| A4 platform core | `A2W4-4-platform-core.md` | 20 | 0 | 0 | 1 | 12 | 0 | 0 | 0 | 1 |
| A5 apigateway | `A2W4-5-apigateway.md` | 19 | 0 | 1 | 4 | 8 | 0 | 1 | 2 | 2 |
| **Σ** | | **83** | **2** | **1** | **10** | **47** | **0** | **1** | **4** | **8** |

**Активных находок (NEW+REGRESS+RESIDUAL): 13**  
**По severity активных: P0=0 · P1=1 · P2=4 · P3=8**

Сравнение со Stage A (`W4-SUMMARY.md`): было сырых **55** (P0=2, P1=6, P2=24, P3=23). После B/W21 большинство закрыто; A2 фиксирует **1 регресс** и **false closures**.

## Топ приоритет

1. **A2W4-5-001 (P1 REGRESS)** — `CompleteTestRun` без `project_id` → fail-closed TASK-012 глушит `test_run.completed` webhooks.  
2. **A2W4-5-002 / A2W4-5-003 (P2 RESIDUAL)** — P2-BE-16 заявлен, но `NewRequestWithContext` и `X-Hub-Delivery` отсутствуют.  
3. **A2W4-2-001 (P2 NEW)** — outbox SKIP LOCKED держит row-lock на время NATS Flush.

## False closures / registry drift

| Claim | Реальность |
|-------|------------|
| P2-BE-16 включает W4-A5-005, W4-A5-006 | **не исправлено** в `webhook.go` |
| P3-BE-U01 / U02 | HSTS, docs CTC, AES AAD — осознанный residual |
| Assignment A3 `metrics_allow_middleware_test.go` | файл удалён (ожидаемо после P3-BE-07) |

## Артефакты

| Файл |
|------|
| `.planning/audits/total-2026-07/A2W4-1-platform-http.md` |
| `.planning/audits/total-2026-07/A2W4-2-platform-nats-outbox.md` |
| `.planning/audits/total-2026-07/A2W4-3-platform-obs.md` |
| `.planning/audits/total-2026-07/A2W4-4-platform-core.md` |
| `.planning/audits/total-2026-07/A2W4-5-apigateway.md` |
| `.planning/audits/total-2026-07/A2-W4-SUMMARY.md` |

Код продукта **не менялся**.
