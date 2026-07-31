# Extract — Backlog Task (portable)

> Сжатие Hub практика backlog-задач Hub (провенанс — SOURCES.md) + master-delegate.
> Одна папка · одна таблица · один формат.

## Имя файла

```
TASK-<NNN>-<YYYY-MM-DD>-<functional-slug>.md
```

Активные — плоско в `backlog/`. Закрытые — в `archive/<date>/` (опционально).

## Тело

Обязательные блоки (пустые = `N/A: причина`):

```markdown
# <emoji> <краткий заголовок>

[OUTCOME]
...

[ROLE]
...

[SCOPE AND CONSTRAINTS]
...

[REFERENCES]
...

[ACCEPTANCE]
- Done:
- Evidence:
- Mutation (если фикс/тест): что ломать / какой тест / слой
- Stop when:

[OUTPUT SCHEMA]
...

[OPERATING RULES]
...
```

## Индекс

`backlog/README.md` — таблица: ID · суть · приоритет · статус · исполнитель · источник.
Обновляется в том же коммите, что статус задачи.

## Статусы

`Открыта` · `В работе` · `На проверке` · `Закрыта` · `Отменена`

«Закрыта» без evidence (см. mutation-evidence) — дефект процесса.

## Кто пишет

Orchestrator / architect / tech-lead. Исполнитель не сочиняет задачу — выполняет.
