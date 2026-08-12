# A3-W2 SUMMARY — tms + bases (FINAL)

> TASK-008 · Stage A3 FINAL · 2026-07-29 · READ-ONLY  
> HEAD: `ee353287` (`cursor-grok-4.5`)  
> Plan: [A3-PLAN.md](A3-PLAN.md) · Assignments: [W2-ASSIGNMENTS.md](W2-ASSIGNMENTS.md)  
> Отчёты: [A3W2-1](A3W2-1-tms-handlers.md) · [A3W2-2](A3W2-2-tms-services.md) · [A3W2-3](A3W2-3-tms-repos.md) · [A3W2-4](A3W2-4-tms-import.md) · [A3W2-5](A3W2-5-bases.md)

## Покрытие

| Агент | Зона | В списке | Прочитано | Пропущено |
|-------|------|--------:|----------:|----------:|
| A1 | tms handlers | 24 | 24 | 0 |
| A2 | tms services | 22 | 22 | 0 |
| A3 | tms repos | 22 | 22 | 0 |
| A4 | tms import | 25 | 25 | 0 |
| A5 | bases | 16 | 16 | 0 |
| **Σ** | | **109** | **109** | **0** |

## Активные находки (сырые)

| Severity | Count |
|----------|------:|
| P0 | **0** |
| P1 | **0** |
| P2 | **7** |
| P3 | **11** |
| **Σ open** | **18** |

| Слот | P0 | P1 | P2 | P3 | Σ |
|------|---:|---:|---:|---:|--:|
| A1 handlers | 0 | 0 | 0 | 1 | 1 |
| A2 services | 0 | 0 | 4 | 2 | 6 |
| A3 repos | 0 | 0 | 1 | 2 | 3 |
| A4 import | 0 | 0 | 2 | 3 | 5 |
| A5 bases | 0 | 0 | 0 | 3 | 3 |

## High-signal

| ID | Sev | Суть |
|----|-----|------|
| A3W2-2-001 | P2 | AssignTesters без membership |
| A3W2-2-002 | P2 | FileIDs без project scope |
| A3W2-2-003 | P2 | ImportAllureToExistingRun без tx |
| A3W2-4-001/002 | P2 | export Steps empty / TestRail depth |
| A3W2-3-001 | P2 | residual unbounded list/batch (partial TASK-111) |

## A2 → A3

Большинство A2 open → **FIXED-OK** (TASK-091/111…). REGRESS 0. RESIDUAL: list caps / bases trash nuances.

## Кумулятив A3

| Волна | Файлов | open Σ |
|-------|-------:|-------:|
| W1 | 163 | 19 |
| W2 | 109 | 18 |
| **Σ** | **272** | **37** |

## Следующий шаг

A3-W3 (messenger+ws+calls, 131 файлов).
