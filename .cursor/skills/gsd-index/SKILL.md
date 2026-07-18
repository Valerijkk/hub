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

`.cursor/skills/qa-gate`, `handoff-sync`, `coverage-audit`, `test-triage`, `style-check`, `anti-ai-check`, `amend-push`

## Test Triage (обязательно)

При FAIL: **сначала** функционал по docs → **потом** код или тест.
См. `.cursor/skills/test-triage/SKILL.md` — REGULATION §14, §18 (примеры по auth/tracker/bases/messenger/tms/calls).

## Быстрый маршрут по ролям (кто чаще какой skill вызывает)

| Роль | Чаще всего | Реже |
|------|-----------|------|
| `team-lead` | `gsd-do`, `gsd-execute-phase`, `gsd-workstreams`, `gsd-progress` | `gsd-autonomous`, `gsd-audit-milestone` |
| `tech-lead` | `gsd-plan-phase`, `gsd-analyze-dependencies` | `gsd-insert-phase` |
| `ba1`/`ba2` | `gsd-discuss-phase`, `gsd-spec-phase` | `gsd-review-backlog`, `gsd-plant-seed` |
| `sa1`/`sa2` | `gsd-coverage-audit`, `gsd-intel` | `gsd-audit-uat`, `gsd-map-codebase` |
| `arch` | `gsd-ai-integration-phase`, `gsd-research-phase` | `gsd-ui-phase`, `gsd-select-framework` |
| `designer` | `gsd-sketch`, `gsd-ui-phase` | `gsd-ui-review` |
| `be1`/`be2` | Hub `qa-gate`, `test-triage`, `gsd-quick` | `gsd-debug`, `gsd-add-tests` |
| `fe1`/`fe2` | Hub `qa-gate`, `gsd-quick` | `gsd-ui-review`, `gsd-sketch` |
| `qa-manual` | `gsd-audit-uat` | `gsd-verify-work` |
| `qa-auto1`/`qa-auto2` | `gsd-add-tests`, Hub `test-triage` | `gsd-validate-phase`, `gsd-audit-fix` |
| `devops-engineer` | `gsd-ship` | — |
| `security-auditor` | `gsd-secure-phase` | `gsd-audit-fix` |
| `debugger` | `gsd-debug` | `gsd-forensics` |
| `code-reviewer` | `gsd-code-review` | `gsd-code-review-fix` |
| `docs-writer` | `gsd-docs-update` | `gsd-milestone-summary` |

## Примеры вызова (реальные ситуации)

```
"Тесты падают в bases после рефакторинга RBAC"
  → .cursor/skills/test-triage/SKILL.md (классификация) → при сложном root cause: gsd-debug

"Нужно решить какой AI-фреймворк для gateway BYOK"
  → gsd-framework-selector → gsd-ai-integration-phase (arch)

"Волна из 4 модулей (auth+tracker+bases+tms) готова, надо смержить"
  → /qa-gate (Hub, полный) → gsd-code-review → team-lead приёмка → gsd-ship

"Сессия заканчивается, нужно зафиксировать прогресс"
  → /handoff-sync (Hub) — не gsd-milestone-summary (это только для завершённого milestone)
```

Полная детализация use-case таблиц и глоссарий GSD↔Hub: `.cursor/GSD-INTEGRATION.md`.
