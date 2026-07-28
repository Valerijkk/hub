# PW1-REPORT — Правда об измеримом

**Дата:** 2026-07-27  
**Агенты:** Grok 4.5 · тесты покрытия только в Docker  
**Продакшн-код:** не менялся

## Сделано

| ID | Результат | Коммит |
|----|-----------|--------|
| A1 | Docker coverage gate: overall **81.9%** (2026-07-27). COVERAGE.md: current section + Historical baselines (2026-06-22). | `ea9811e` |
| A2 | Undated coverage claims date-stamped | `62d13a9` |
| A3 | `PHASE-STATUS-MATRIX.md` — 114 phase dirs, 52 without verification at scan time | `8e9a556` / `ebdad92` |
| A4 | STATE: `tasks_completed` / `goals_verified` / `phases_percent` + coverage/inventory | `c6aa14a` |
| A5 | `deploy_blockers` PG-03/FA-06 documented + pasted into STATE | `dc82a05` + A4 |
| A6 | Inventory measured: routes **72** tsx, tests **515**, Ent **82**, stores **13**, Go **1.25.0** | `1538470` |

## Числа (замерено сейчас)

| Метрика | Значение | Дата |
|---------|----------|------|
| Overall `internal/` statements | **81.9%** | 2026-07-27 |
| Baseline overall | 83.0% | 2026-06-22 |
| Delta | **−1.1 pp** | — |
| Gate floor ≥82% | **MISS** | 2026-07-27 |
| goals_verified (matrix PASS) | **55** | 2026-07-27 |
| phases without verification (scan) | **52** | 2026-07-27 |

## Хуже ожидаемого

- Coverage **просел** под floor 82% (было 83.0%).
- Per-package **statement** table for 2026-07-27: coverprofile export failed on second Docker run (`proxy.golang.org` EOF). Third run with GOMODCACHE was in flight; until then only overall is method-comparable to baseline.
- Do not compare func-mean package proxies to 2026-06-22 statement package %.

## Владельцу

- Принять, что live coverage = **81.9% (2026-07-27)**, не 83%.
- PG-03/FA-06 остаётся единственным prod-ops блокером (см. STATE `deploy_blockers`).
