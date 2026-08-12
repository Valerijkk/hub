# A3-W11 SUMMARY — fe lib + docs (FINAL)

> TASK-008 · Stage A3 FINAL · 2026-07-29 · READ-ONLY  
> HEAD: `ee353287` · Assignments: [W11-ASSIGNMENTS.md](W11-ASSIGNMENTS.md)  
> Отчёты: [A3W11-1](A3W11-1-fe-lib-docs.md) · [A3W11-2](A3W11-2-fe-lib-docs.md) · [A3W11-3](A3W11-3-fe-lib-docs.md) · [A3W11-4](A3W11-4-fe-lib-docs.md) · [A3W11-5](A3W11-5-fe-lib-docs.md)

## Покрытие

| Агент | В списке | Прочитано | Σ open |
|-------|--------:|----------:|-------:|
| A1 | 19 | 19 | 1 |
| A2 | 19 | 19 | 2 |
| A3 | 19 | 19 | 2 |
| A4 | 19 | 19 | 3 |
| A5 | 17 | 17 | 3 |
| **Σ** | **93** | **93** | **11** |

## Severity

| P0 | P1 | P2 | P3 |
|---:|---:|---:|---:|
| 0 | **2** | **5** | **4** |

## High-signal

| ID | Sev | Tag | Суть |
|----|-----|-----|------|
| A3W11-1-001 | P1 | REGRESS | `client.ts` не мапит wire `error`→`ApiError.message` |
| A3W11-2-* | P1/P2 | — | см. слот A2 (auth/storage) |
| A3W11-5-001…003 | P2 | NEW/RES | inventory/strategy/spec drift |

## Кумулятив A3 (W1–W6 + W9–W11; W7/W8 ещё open)

| Волна | Файлов | open Σ |
|-------|-------:|-------:|
| W1–W6 | 674 | 101 |
| W9 | 127 | 23 |
| W10 | 144 | 25 |
| W11 | 93 | 11 |
| **Σ (без W7/W8)** | **1038** | **160** |

Ждём W7 (4/5) · W8 (4/5) → `00-STAGE-A3-COMPLETE.md`.
