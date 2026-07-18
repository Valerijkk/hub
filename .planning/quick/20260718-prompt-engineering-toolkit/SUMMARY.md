---
slug: prompt-engineering-toolkit
created: 2026-07-18
completed: 2026-07-19
type: quick
status: complete
---

# Summary: Portable Prompt Engineering Toolkit

## Что сделано

Создан `prompt-engineering/` — обезличенный (project-agnostic), переносимый в любой проект тулкит промпт-инжиниринга для работы с тестовым фреймворком. **53 Markdown-файла, 1000 промптов.** Архитектура по мотивам GSD (templates/rules/playbooks/context), вся проектная специфика вынесена в заполняемые `context/*.md`.

## Структура

| Блок | Файлы | Что |
|------|-------|-----|
| Корень | `README.md`, `INSTALL.md` | Обзор, философия, перенос в новый проект (5 шагов) |
| `context/` | 4 шаблона + README | Заполняемые факты: PROJECT-CONTEXT, TESTFRAMEWORK-MAP, NAMING-CONVENTIONS, PATTERNS-LIBRARY (индустриальный baseline A1–A12 + проектный overlay) |
| `agents/` | 4 агента + README | test-code-reviewer (паттерны+нейминг), test-case-generator (TC из доки+верификация), test-debugger (triage), test-documenter |
| `agents/templates/` | `_MASTER-PROMPT-TEMPLATE` + 9 сценарных + README | Единый шаблон «кто/что/где + best practices» и его специализации |
| `skills/` | 4 скила + README | context-intake, anti-ai-check, naming-check, test-verification |
| `rules/` | 5 правил + README | 00 глобальные, 10 нейминг, 20 качество, 30 severity, 40 верификация |
| `playbooks/` | 5 + README | onboarding, test-generation, review-cycle, debugging, documentation |
| `prompts/` | 12 категорий + README | **1000 промптов** со `{{слотами}}` |

## Ключевые механизмы

1. **Обезличенность + `context-intake`.** Ноль хардкода проекта; скил сканирует репо и заполняет `context/` фактами с источниками. Перенос = копирование папки + один прогон скила.
2. **Единый мастер-шаблон промпта** (`_MASTER-PROMPT-TEMPLATE.md`) — role prompting, контекст-инъекция, формат вывода, few-shot, chain-of-thought, step-back/ReAct, параметры генерации, верификация. Обезличенное зеркало канона `.cursor/agents/team-lead.md` §5.2.
3. **10-этапный протокол верификации** (`rules/40`) — вшит секцией во все 4 агента и 4 скила; режимы self-check и independent. Компактный блок верификации — обязательная часть каждого выхода.

## Библиотека промптов (1000)

01 design (85), 02 unit (85), 03 integration+GraphQL/gRPC/WS/observability (85), 04 api+cache/versioning (85), 05 e2e/ui+a11y/mobile/realtime/perf (85), 06 review/quality+self-check (85), 07 debugging/flaky+chaos (70), 08 test-data/mocking+builders (85), 09 coverage branch/mutation (80), 10 ci+environments/merge-queue (85), 11 perf/chaos/OWASP-security (85), 12 docs/reporting+onboarding (85).

## Интеграция Hub

- `.cursor/rules/70-prompt-engineering.mdc` — карта тулкита, правила работы с тестами через агентов, связь с Hub-канонами.
- `.cursor/REGULATION.md` §20 — назначение, состав, механизмы, правила для агентов Hub.
- Тулкит остаётся чистым: Hub-специфика только в `context/*.md` (заполнит `context-intake`) и rule 70.

## Коммиты (ветка master)

```
e8d17c1 scaffold — README, INSTALL, context templates
fc5b16c skills — context-intake, anti-ai-check, naming-check, test-verification
e3ee28c agents — reviewer, tc-generator, debugger, documenter + 9 templates
99afba7 rules + 5 playbooks
9da8d7b prompts part 1 (325)
299e8dc prompts complete — 1000
0b59609 10-stage verification protocol wired into all agents/skills
4a69e80 master prompt template
d36198d cursor integration — rule 70 + REGULATION §20
```

## Верификация выполнения

- `find prompt-engineering -name '*.md' | wc -l` → 53 файла.
- Счётчик промптов → ровно 1000 (`grep -cE '^[0-9]+\. '`).
- Все агенты/скилы содержат секцию верификации (rules/40).
- Все launch-шаблоны ссылаются на `_MASTER-PROMPT-TEMPLATE.md`.
- `.cursor` rule 70 + REGULATION §20 связаны с тулкитом двусторонне.

## Что дальше (для пользователя)

Тулкит готов к переносу. Для активации на Hub: прогнать `skills/context-intake` — он заполнит `context/*.md` реальными фактами Hub (backend Go/testframework/, frontend vitest/playwright, нейминг из style guide). После этого агенты работают на Hub-фактах.
