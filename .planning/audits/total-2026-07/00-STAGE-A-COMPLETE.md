# STAGE A COMPLETE — волны 1–11

> TASK-008 · 2026-07-28 · **read-only по коду соблюдён**

## Покрытие файлов

| Волна | Зона | Файлов | P0 | P1 | P2 | P3 | Сырых |
|------:|------|-------:|---:|---:|---:|---:|------:|
| 1 | tracker+sprint+project | 163 | 1 | 19 | 28 | 18 | 66 |
| 2 | tms+bases | 109 | 0 | 14 | 36 | 16 | 66 |
| 3 | messenger+ws+calls | 131 | 1 | 17 | 32 | 18 | 68 |
| 4 | platform+apigateway | 83 | 2 | 6 | 24 | 23 | 55 |
| 5 | ai+search | 96 | 0 | 9 | 31 | 28 | 68 |
| 6 | auth+rbac+audit | 92 | 0 | 6 | 28 | 21 | 55 |
| 7 | remaining backend | 284 | 0 | 9 | 38 | 28 | 75 |
| 8 | ent/schema | 83 | 0 | 19 | 40 | 12 | 71 |
| 9 | frontend/components | 127 | 0 | 11 | 52 | 47 | 110 |
| 10 | routes+modules | 144 | 0 | 14 | 64 | 20 | 98 |
| 11 | api/stores/hooks/lib+docs | 93 | 0 | 17 | 49 | 14 | 80 |
| **Σ** | | **1405** | **4** | **141** | **422** | **245** | **812** |

Оценка TASK (~2031) шире (все docs/wiki рекурсивно + возможный drift замеров). По волнам A1–A11 списки закрыты **0 пропусков** в отчётах агентов.

## P0 (приоритет стадии B)

1. **W1-A5-001** — `project` routes без RBAC / mutate без membership  
2. **W3-A2-001** — promote-to-issue без `PermTaskCreate` (PrivEsc)  
3. **W4-A5** ×2 — apigateway webhook/signing (см. A4-5)

## Evidence тестов

Все прогнанные `testframework/{domain}` пакеты волн 1–7 — **ok** (`-tags hubtest -race`, Postgres/Mongo/Redis/NATS подняты).

## Handoff → Стадия B (W12+)

1. Дедуп 812 сырых → реестр уникальных классов  
2. Завести задачи с TASK-009+ по правилу 57 (с критерием мутации)  
3. Чинить **P0 → P1 → P2 → P3**  
4. Код трогать только со стадии B

Индекс: [`00-INDEX.md`](00-INDEX.md) · Отчёты: `A{wave}-{agent}-*.md` · Сводки: `W{N}-SUMMARY.md`
