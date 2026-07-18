---
slug: prompt-engineering-toolkit
created: 2026-07-18
type: quick
status: in-progress
---

# Quick Task: Portable Prompt Engineering Toolkit

## Цель

Создать `prompt-engineering/` в корне hub — **обезличенный, переносимый в любой проект** тулкит промпт-инжиниринга для работы с тестовым фреймворком: агенты, шаблоны промптов, скилы, правила, контекст-темплейты, плейбуки и большая библиотека промптов (~1000). Архитектура заимствована у GSD (get-shit-done): templates / workflows / references / agent-contracts.

## Ключевые решения

1. **Портируемость** — ноль хардкода проекта: все проектные факты живут в `context/*.md`, которые заполняются скилом `context-intake` при онбординге тулкита в новый проект. Агенты читают только `context/`, не гадают.
2. **GSD-паттерны** — frontmatter-контракты агентов, fill-in шаблоны с `{{placeholders}}`, workflows-плейбуки, references-правила.
3. **Агенты (5)**: `test-code-reviewer` (паттерны фреймворка + нейминг функций/файлов), `test-case-generator` (TC из документации + верификация), `test-debugger`, `test-documenter`, + `context-intake` как скил-бутстрап.
4. **Скилы (4)**: `context-intake`, `anti-ai-check` (говнокод + AI-комментарии), `naming-check`, `test-verification`.
5. **Библиотека промптов**: 12 категорийных файлов × ~85 промптов ≈ 1000, все с `{{placeholders}}`.
6. **Интеграция Hub** — `.cursor/rules/70-prompt-engineering.mdc` + секция в `.cursor/REGULATION.md`; Hub-специфика ТОЛЬКО там, тулкит остаётся чистым.

## Задачи

- [ ] T1: Каркас: README.md, INSTALL.md, context/ (4 шаблона + README)
- [ ] T2: skills/ (4 скила + README)
- [ ] T3: agents/ (4 агента + README) + agents/templates/ (8+ шаблонов промптов + README)
- [ ] T4: rules/ (4 файла + README) + playbooks/ (5 + README)
- [ ] T5: prompts/ (12 категорий + README, ~1000 промптов)
- [ ] T6: .cursor интеграция (rule 70 + REGULATION §20)
- [ ] T7: SUMMARY.md + STATE.md + атомарные коммиты

## Верификация

- Ни один файл `prompt-engineering/` не содержит Hub-специфичных путей/доменов вне явно помеченных примеров-иллюстраций.
- Каждый агент ссылается только на `context/` и `rules/` тулкита.
- Библиотека промптов: суммарный счётчик в prompts/README.md ≈ 1000.
