---
agent: любой + скил refresh-test-context
scenario: обновление существующего profile после изменений в проекте/фреймворке
---

# Шаблон: refresh контекста

```
Прогони skills/refresh-test-context/SKILL.md для profile="{{profile}}".

## Что изменилось (подсказка, где перечитывать)
- {{«реорганизовали e2e» | «сменили раннер X→Y» | «новый модуль Z» | «не знаю —
  проверь всё, сверь каждый факт с источником»}}

## Правила refresh
1. Сохрани byte-for-byte все BEGIN/END MANUAL blocks.
2. Изменившиеся факты обнови с current source IDs; исчезнувшие evidence-backed
   claims замени на `UNKNOWN` и зафиксируй в отчёте.
3. Обнови affected artifacts в `context/projects/<profile>/`. Не создавай и не
   approve pattern cards: feedback и pattern curation остаются отдельными gates.

## Ожидаемый выход
Refresh report: source delta, changed artifacts, preserved manual blocks,
new `UNKNOWN`, stale pattern refs и human follow-ups.
```
