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

## W13 — исполнение P0 (код готов локально, коммиты не сделаны)

| TASK | Тесты | Статус кода |
|------|-------|-------------|
| 009 project RBAC | `./testframework/project/...` ok | готово |
| 010 promote RBAC | `./testframework/messenger/...` ok | готово |
| 011 webhook match | `./testframework/apigateway/...` ok | готово |
| 012 webhook scope | вместе с 011 | готово |

Следующий шаг владельца: **коммиты** по TASK-009…012 (или «продолжай» = коммитить + P1).

## P2/P3

Сырые 422+245 остаются в отчётах A*; постановка задач — после закрытия P0/P1 или отдельной волной.
