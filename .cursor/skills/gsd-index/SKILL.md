---
name: gsd-index
description: Индекс GSD skills для Hub-агентов. Читай при оркестрации волн или неясном маршруте.
---

# GSD Skills Index (Hub)

Полная матрица агентов: `.cursor/GSD-INTEGRATION.md`

GSD skills path: `~/.claude/skills/gsd-<name>/SKILL.md`

## Планирование и исполнение

| Skill | Когда |
|-------|-------|
| `gsd-do` | Не знаешь какую команду выбрать |
| `gsd-progress` | Статус проекта, что дальше |
| `gsd-next` | Следующий логический шаг |
| `gsd-plan-phase` | Детальный PLAN.md для фазы |
| `gsd-execute-phase` | Параллельное исполнение планов |
| `gsd-workstreams` | Несколько параллельных потоков |
| `gsd-quick` | Мелкая задача с STATE tracking |
| `gsd-fast` | Тривиальный inline фикс |

## Качество и тесты

| Skill | Когда |
|-------|-------|
| `gsd-add-tests` | Тесты после фазы |
| `gsd-validate-phase` | Nyquist validation gaps |
| `gsd-verify-work` | UAT / acceptance |
| `gsd-audit-fix` | Аудит → классификация → фикс |
| `gsd-audit-uat` | Cross-phase UAT gaps |
| `gsd-debug` | Сложный баг, persistent session |
| **Hub:** `test-triage` | FAIL теста → код vs тест |
| **Hub:** `qa-gate` | Полный регрессионный gate |

## Анализ и документация

| Skill | Когда |
|-------|-------|
| `gsd-coverage-audit` | Требование × код × тест |
| `gsd-intel` | Codebase intelligence refresh |
| `gsd-map-codebase` | Глубокий анализ кодовой базы |
| `gsd-docs-update` | Синхронизация docs с кодом |
| `gsd-code-review` | Peer review плана/кода |
| `gsd-secure-phase` | Threat verification фазы |

## Hub slash (Cursor)

`.cursor/skills/qa-gate`, `handoff-sync`, `coverage-audit`, `test-triage`, `style-check`

## Test Triage (обязательно)

При FAIL: **сначала** функционал по docs → **потом** код или тест.
См. `.cursor/skills/test-triage/SKILL.md` — REGULATION §10.
