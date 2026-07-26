---
status: template            # template | filled
compatibility: deprecated
read_only: true
replacement: "context/projects/<profile>/{PROJECT,COMMANDS,CI-PIPELINE}.md"
filled_at: {{YYYY-MM-DD}}
---

# Project Context

> **DEPRECATED / READ-ONLY COMPATIBILITY.** Новые workflows не заполняют этот
> файл. Используйте active profile `PROJECT.md`, `COMMANDS.md` и
> `CI-PIPELINE.md`; неизвестное фиксируется как `UNKNOWN`.

## 1. Проект

- **Название:** {{project_name}}
- **Тип:** {{monorepo | single-app | multi-repo}}
- **Краткое описание (1–2 предложения):** {{what_it_does}}
- **Основная ветка:** {{default_branch}}

## 2. Стек

| Слой | Технология | Версия | Источник факта |
|------|-----------|--------|----------------|
| Backend | {{lang_framework}} | {{version}} | {{go.mod / package.json / ...}} |
| Frontend | {{lang_framework}} | {{version}} | {{...}} |
| БД | {{databases}} | {{...}} | {{...}} |
| Инфра | {{docker / k8s / compose}} | {{...}} | {{...}} |

## 3. Структура репозитория (только верхний уровень + тестовые директории)

```
{{repo_tree_top_level}}
```

- Тесты живут в: {{test_dirs}}
- Конфиги тест-раннеров: {{runner_config_paths}}

## 4. Команды (канонические, проверенные запуском)

| Действие | Команда | Откуда |
|----------|---------|--------|
| Сборка backend | `{{build_cmd}}` | {{Makefile / CI / docs}} |
| Юнит-тесты backend | `{{unit_test_cmd}}` | {{...}} |
| Интеграционные тесты | `{{integration_test_cmd}}` | {{...}} |
| Тесты frontend | `{{fe_test_cmd}}` | {{...}} |
| E2E | `{{e2e_cmd}}` | {{...}} |
| Линт | `{{lint_cmd}}` | {{...}} |
| Один тест точечно | `{{single_test_cmd_pattern}}` | {{...}} |

⛔ Агенты используют ТОЛЬКО команды из этой таблицы. Придумывать команды запрещено.

## 5. CI

- CI-система: {{github-actions | gitlab-ci | ...}}, конфиг: {{path}}
- Джобы, гоняющие тесты: {{job_names}}
- Гейты качества (coverage порог, обязательные джобы): {{quality_gates}}

## 6. Документация проекта

| Документ | Путь | Зачем агентам |
|----------|------|----------------|
| Спецификация / требования | {{path}} | источник истины для test-case-generator |
| Тест-план / стратегия | {{path}} | пирамида, приоритеты |
| Style guide | {{path}} | нейминг, форматирование |
| {{other}} | {{path}} | {{...}} |

## 7. Ограничения и запреты проекта

> Технологии/паттерны, которые в проекте запрещены или обязательны (из style guide, ADR, constraints-доков).

- {{constraint_1}}
- {{constraint_2}}
