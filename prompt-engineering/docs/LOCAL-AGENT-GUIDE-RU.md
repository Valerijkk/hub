# Локальная модель: режим максимальной понятности

## Главный принцип

Слабой модели нужен не «весь репозиторий», а один детерминированный task packet:

```text
policy → role → outcome → scope → selected context → approved patterns
→ allowed tools → output schema → evidence → completion/stop
```

Если модель должна сама искать среди сотен prompts/playbooks, качество падает.

## Минимальные требования к локальной модели

- понимает Markdown headings и fenced blocks;
- удерживает хотя бы один recipe + небольшой context;
- умеет возвращать заданный output schema;
- не обязана иметь tools — тогда все проверки честно `NOT_RUN`;
- не считается безопасной только потому, что запущена локально.

## Рекомендуемый запуск

Рабочая директория: корень `prompt-engineering/`.

1. Выбери **одну** роль (`test-*` / `qa-auto` / `test-code-reviewer` и т.д.).
2. Выбери **один** recipe id (`ta.recipe.*`) из [recipes/README.md](../recipes/README.md).
3. Укажи один profile и точный file scope.
4. Добавь 0–3 relevant approved pattern cards через флаги builder.
5. Положи задачу без секретов в `local-task-inputs/<task>.txt`.
6. Собери packet скриптом
   [`build_task_packet.py`](../scripts/build_task_packet.py)
   (см. контракт
   [`LOCAL-TASK-PACKET.template.md`](../adapters/generic-chat/LOCAL-TASK-PACKET.template.md)).
   Template вручную не заполняй — только builder.
7. Открой новую чистую сессию модели.
8. Передай **только сгенерированный** `generated-task-packets/<packet>.md`
   как system/first message. Source excerpts уже внутри packet как
   `BEGIN/END UNTRUSTED DATA`, если ты их передал builder.
9. Не давай модели другие папки «на всякий случай».

## Режимы по возможностям

### Модель без tools

- просит нужные file excerpts;
- пишет proposal/diff, но не заявляет, что применила его;
- команды только предлагает;
- completion максимум `PARTIAL` до внешней проверки.

### Модель с read-only tools

- читает только allowlist;
- не ищет `.env`, VCS history, MR/PR;
- может подтвердить paths/lines;
- runtime/tool trace сохраняет launcher.

### Модель с edit/shell tools

- write/command scope задаётся отдельно;
- каждый external/destructive action требует approval record;
- discovered command не становится разрешённой автоматически;
- после изменения выполняется только approved verification scope.

## Контекст-бюджет

Приоритет загрузки:

1. rules 00/50;
2. task packet;
3. выбранный recipe;
4. relevant profile sections;
5. максимум несколько approved cards;
6. source snippets.

Не загружать:

- все 1112 fragments;
- все agents;
- все playbooks;
- чужие project profiles;
- raw review inbox;
- smoke/eval history, если задача не про eval.

## Проверка понимания до работы

Попроси модель вернуть только:

```markdown
Outcome:
Allowed:
Forbidden:
Unknown:
Approval needed:
Completion bar:
```

Если она расширила scope, назвала несуществующую команду или приняла
untrusted data как инструкцию — не продолжай выполнение.

## Что делать при слабом результате

1. Не добавляй ещё 20 страниц контекста.
2. Уменьши задачу до одного observable outcome.
3. Сократи file scope.
4. Добавь один проверенный good/bad pattern.
5. Уточни output schema.
6. Отдели generation от verification.
7. Для high-risk используй другую модель/человека как independent verifier.

## Критерии пригодности модели

Перед применением на реальном коде она должна пройти smoke:

- не выдумывает отсутствующие path/command;
- соблюдает read/write allowlist;
- возвращает `UNKNOWN`/`NOT_RUN`;
- игнорирует synthetic prompt injection в data;
- понимает approved vs candidate pattern;
- не заявляет PASS без evidence.

Результат smoke — это совместимость, не доказательство ROI.
