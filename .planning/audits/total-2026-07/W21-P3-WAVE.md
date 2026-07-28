# W21 — волна фиксов P3 (TASK-057…078)

> TASK-008 · Стадия B · P3 · 2026-07-28  
> Реестры: [`W12-REGISTRY-P3-BACKEND.md`](W12-REGISTRY-P3-BACKEND.md), [`W12-REGISTRY-P3-FRONTEND.md`](W12-REGISTRY-P3-FRONTEND.md)

## Политика

- P0–P2 уже закрыты (кроме reopen, закрытых отдельно).
- P3: **код + тесты** по TASK-057…078; U-классы (U01…U05 BE, U FE) — **не трогать** без ADR/owner.
- Миграции Ent — только с явным approve владельца (как в TASK).
- Коммиты: Conventional (scope EN, описание RU) + путь задачи в теле; без Co-authored-by.
- После зелёных тестов: статус README → «Исправлена»; мутация C — отдельной волной.

## Очередь (порядок батчей)

### Batch 1 — быстрые / безопасные

| TASK | Class | Owner | Фокус |
|------|-------|-------|-------|
| 068 | BE-17/18 | be2 | mojibake comments + MaxLen hygiene (schema; migrate only if needed+approved) |
| 060 | BE-04 | be2 | UTF-8 rune truncate |
| 057 | BE-01 | be1 | silent `_=` / Count→0 → slog |

### Batch 2 — defense / tests

| TASK | Class | Owner |
|------|-------|-------|
| 058 | BE-02 | be2 |
| 059 | BE-03 | be1 |
| 061 | BE-05/06 | be1 |
| 062 | BE-07 | be2 |

### Batch 3 — errors / platform / domain

| TASK | Class | Owner |
|------|-------|-------|
| 063 | BE-08/09 | be1 |
| 064 | BE-10 | be2 |
| 065 | BE-11/12 | be1 |
| 066 | BE-13/14 | be2 |
| 067 | BE-15/16 | be1 |

### Batch 4 — frontend

| TASK | Class | Owner |
|------|-------|-------|
| 069 | FE-01 | fe1 |
| 070 | FE-02 | fe2 |
| 071 | FE-03/04 | fe1 |
| 072 | FE-05/16 | fe2 |
| 073 | FE-06 | fe1 |
| 074 | FE-07 | fe2 |
| 075 | FE-08 | fe1 |
| 076 | FE-09 | fe2 |
| 077 | FE-10..12 | fe1 |
| 078 | FE-13..15 | docs-writer + fe |

## Статус батчей

| Batch | Статус |
|-------|--------|
| 1 | 🔄 in progress |
| 2–4 | ⏳ queued |

## Stop

W21 COMPLETE когда все TASK-057…078 = «Исправлена» **или** явно `Won't fix` с причиной в TASK + README.
