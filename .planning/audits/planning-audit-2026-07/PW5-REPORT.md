# PW5-REPORT — Миграция тестового фреймворка (август)

**Дата:** 2026-07-27  
**Код /tests не создавался** — только планы.

## Сделано

| ID | Артефакт | Коммит |
|----|----------|--------|
| E1 | `TF-MIGRATE-E1-INVENTORY.md` | `8533644` |
| E2 | `TF-MIGRATE-E2-EXPORT-HOOKS.md` | `e14374e` |
| E3 | `TF-MIGRATE-E3-PLAYWRIGHT.md` | `585354b` |
| E4 | `TF-MIGRATE-E4-TESTS-LAYOUT.md` | `a52fc2d` |
| E5 | `TF-MIGRATE-E5-LOCUST.md` | `199cc8d` |
| E6 | PLAN phases 128–136 + index | `8a5a563`…`9c67830` |

## Жёсткое условие

**Э1 / phase 128 не стартует**, пока не зафиксирован честный coverage baseline — **выполнен A1: 81.9% на 2026-07-27**.

## Хуже ожидаемого

- Baseline уже **ниже** 82% floor — миграция должна сравниваться с 81.9%, не с устаревшими 83.0%.

## Владельцу

Открытые вопросы из `02-TEST-FRAMEWORK-MIGRATION.md` (порог 82%, владелец `/tests`, Locust в CI, судьба guides).
