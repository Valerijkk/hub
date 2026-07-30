# Layer 06 — Planning Ops

Источники: `.planning/README.md`, `.planning/STATE.md`, `.cursor/rules/57-backlog-tasks.mdc`,
`.cursor/rules/55-docs-sync.mdc`, Standard `playbooks/handoff-state.md`.

## Минимальный planning-контур в чужом репо

```
.planning/                 # или docs/planning/ — имя на выбор команды
├── STATE.md               # live позиция + session notes
├── README.md              # навигация
└── backlog/
    ├── README.md          # индексная таблица
    └── TASK-NNN-YYYY-MM-DD-slug.md
```

Не обязательно копировать всю структуру Hub (phases/, audits/, …). Обязательны
**STATE** и **задачи с ACCEPTANCE**.

## Задача исполнителю

Пишется по мастер-контракту. Extract:
[`../extracts/backlog-task.md`](../extracts/backlog-task.md).

Правила:

- Активные задачи — плоско в одной папке
- Исполнитель задачу не сочиняет
- «Закрыта» только с evidence (layer 05)
- Ссылка на файл задачи в теле коммита

## STATE

Каждая сессия: что сделано · какие тесты · какой doc · blockers · next.
Не дублировать романы — ссылки на артефакты.

## Docs sync

После user-visible / API / CI изменения — обновить domain doc в той же сессии.
Матрица «что → какой doc» — адаптация rule 55 под целевой репозиторий.

## DoD слоя

Новый участник находит задачу и STATE без чтения чатов.
