# W12 COMPLETE — дедуп + постановка задач

> TASK-008 · Стадия B · 2026-07-28

## Артефакты

| Файл | Содержание |
|------|------------|
| [W12-REGISTRY-P0.md](W12-REGISTRY-P0.md) | 4×P0 → TASK-009…012 |
| [W12-REGISTRY-P1-BACKEND.md](W12-REGISTRY-P1-BACKEND.md) | 99 сырых → 49 классов → 13 задач |
| [W12-REGISTRY-P1-FRONTEND.md](W12-REGISTRY-P1-FRONTEND.md) | 42 сырых → 10 классов → 10 задач |

## Задачи заведены

- **P0:** TASK-009 … TASK-012
- **P1 BE:** TASK-013 … TASK-025
- **P1 FE:** TASK-026 … TASK-035
- **Blocked (нужен ADR/scope):** U01–U06 backend + 4 FE «уточнение» — в реестрах, без TASK

## W13 — исполнение P0 (закоммичено)

| TASK | Коммит | Тесты |
|------|--------|-------|
| 009 project RBAC | `b81ee8f7` | project ok |
| 010 promote RBAC | `a0fc6928` | messenger ok |
| 011+012 webhooks | `b7281200` | apigateway ok |

W12 docs: `929a01d2`

## W14 — P1 в работе (агенты)

| TASK | Агент |
|------|-------|
| 013 tracker IDOR | be1 |
| 014 project/sprint ACL | be2 |
| 015 tms IDOR | generalPurpose |

Следующий шаг после green: коммиты 013–015 → TASK-016+.


## P2/P3

Сырые 422+245 остаются в отчётах A*; постановка задач — после закрытия P0/P1 или отдельной волной.
