# agents/ — системные промпты агентов

> Каждый файл — готовый системный промпт роли. Frontmatter совместим с Cursor custom agents; для других инструментов содержимое просто вставляется как system prompt. Запуск — через шаблоны из [templates/](templates/).

## Ростер

| Агент | Роль | Режим | Ключевой выход |
|-------|------|-------|-----------------|
| [test-code-reviewer](test-code-reviewer.md) | Ревью кода тестов: паттерны фреймворка + нейминг функций/файлов | readonly | Вердикт approve / request changes с находками |
| [test-case-generator](test-case-generator.md) | Тест-кейсы из документации + самоверификация | пишет TC-доки | TC-набор + verification report |
| [test-debugger](test-debugger.md) | Систематический дебаг упавших/flaky тестов | пишет минимальный фикс | Root cause + фикс + регрессия |
| [test-documenter](test-documenter.md) | Документация тестового фреймворка в синхроне с кодом | пишет docs | Обновлённые доки + changelog |

## Общий контракт всех агентов

1. **Перед задачей грузить**: `context/PROJECT-CONTEXT.md`, `context/TESTFRAMEWORK-MAP.md`, `context/NAMING-CONVENTIONS.md`, `context/PATTERNS-LIBRARY.md`, `rules/00-global-agent-rules.md` + правила своей роли.
2. **Контекст не заполнен** (status: template / `TODO(intake)` в нужной секции) → остановиться и запросить `/context-intake`. Не гадать.
3. **Выход по формату** своей секции «Формат вердикта/отчёта» — машинно-парсируемый, с severity.
4. **Границы роли** — секция «Что я НЕ делаю» обязательна к соблюдению; выход за границы = передача другому агенту, не самодеятельность.

## Конвейер (кто за кем)

```
документация фичи
   → test-case-generator (TC + self-check)
   → /test-verification (независимый гейт)
   → автоматизация (человек или кодер-агент по TC)
   → test-code-reviewer (ревью кода тестов + naming)
   → CI
   → упало? → test-debugger (triage: код vs тест)
   → test-documenter (синхрон доков после изменений)
```

## Подключение

| Инструмент | Действие |
|-----------|----------|
| Cursor | Скопировать файл в `.cursor/agents/`, при желании — проставить `model:` в frontmatter |
| Claude Code | В `.claude/agents/` проекта |
| Чат / API | Содержимое файла = system prompt; шаблон из templates/ = первое user-сообщение |
