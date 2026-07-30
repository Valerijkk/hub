# Установка Touchstone Standard в новый проект

> **Издание:** Touchstone Standard (`prompt-engineering/`).
> Portable core обезличен. Generated context и вручную добавленные project
> examples принадлежат конкретному проекту и не переносятся в другой автоматически.
> Structural validator использует только Python standard library.
>
> Нужен max-пакет с GSD/Cursor/planning практиками — ставь сначала Standard,
> затем подключай sibling `prompt-engineering-pro/` (рядом с этой папкой в
> исходном репозитории продукта; не входит в portable package).

## Шаг 1. Собрать portable package

```powershell
py -3 -B .\prompt-engineering\scripts\package_portable.py `
  --root .\prompt-engineering `
  --output <new-project>\prompt-engineering

py -3 -B <new-project>\prompt-engineering\scripts\validate_toolkit.py `
  --root <new-project>\prompt-engineering
```

Packager переносит templates/core и автоматически исключает generated profiles,
local smoke/run evidence и project-only pattern/review content. Raw
`cp -r prompt-engineering/` не использовать после появления project profiles.
Dry-run: добавь `--dry-run`. Output должен быть новым путём или существующей
пустой директорией. Непустой output всегда приводит к отказу: packager ничего
не удаляет и не поддерживает режим замены.

## Шаг 2. Создать контекст тестового фреймворка

Сначала создай механический inventory только из явно заданных корней:

```powershell
py -3 -B .\prompt-engineering\scripts\context_inventory.py `
  --project-root <repo-root> `
  --test-root <tests-root-1> `
  --test-root <tests-root-2> `
  --output <approved-local-evidence>\context-inventory.json
```

Inventory использует только stdlib, не запускает команды проекта и не обращается
к сети. Output-файл должен быть новым и находиться вне объявленных корней.
Скрипт применяет text/size/count limits, exclusions, containment и reparse-point
checks; bootstrap может читать только перечисленные и повторно проверенные
`files[]`.

Запусти explicit skill [bootstrap-test-context](skills/bootstrap-test-context/SKILL.md).
Автоматизатор задаёт только корень проекта, пути тестового фреймворка и имя
профиля:

```
/bootstrap-test-context
project_root=<repo-root>
test_roots=<tests-root-1>,<tests-root-2>
inventory=<approved-local-evidence>\context-inventory.json
profile=<project-slug>
```

Результат — `context/projects/<profile>/`: manifest, архитектура тестов, реальные
команды/CI, test data, traceability, source map и индекс patterns. Команды не
запускаются только потому, что найдены в репозитории; непроверенное помечается
`UNKNOWN`/`NOT_RUN`.

## Шаг 3. Подключить адаптер и модели

Файлы `agents/*.md` — переносимое содержание ролей, но не готовая конфигурация
каждого AI-инструмента. Используй mapping/wrapper из `adapters/`.

| Инструмент | Как |
|-----------|-----|
| Cursor | [adapters/cursor](adapters/cursor/) — frontmatter, `inherit`, skills |
| CI-бот | [adapters/ci](adapters/ci/) — system/user/output contracts |
| Любой чат-LLM | [adapters/generic-chat](adapters/generic-chat/) |

GPT-5.6 Sol Pro — orchestrator default. Остальные модели выбираются
оркестратором из фактически доступного Cursor catalog и фиксируются в handoff.

## Шаг 4. При необходимости вручную добавить опыт коллег

Bootstrap **не читает закрытые MR**. После его завершения человек сам копирует
в [review-feedback/inbox](review-feedback/inbox/):

- плохой фрагмент кода;
- комментарий коллеги;
- исправленный хороший фрагмент;
- известные path/commit/date и область применимости.

Затем запускается `/curate-review-feedback`. До redaction/provenance/human
approval пример остаётся candidate и не используется как обязательный канон.

## Шаг 5. Проверить слой

```powershell
py -3 -B .\prompt-engineering\scripts\validate_toolkit.py `
  --root .\prompt-engineering
```

После structural PASS запусти eval smoke из `evals/`. Результаты до фактического
прогона имеют статус `NOT_RUN`; запрещено заполнять метрики ожидаемыми числами.

Для слабой локальной модели после bootstrap см. сценарий D в
[START-HERE.md](START-HERE.md) и
[LOCAL-AGENT-GUIDE-RU.md](docs/LOCAL-AGENT-GUIDE-RU.md): задача →
`build_task_packet.py` → только сгенерированный packet.

## Шаг 6. Интеграция с регламентом проекта

Если в проекте есть свой регламент AI-команды (правила Cursor, CLAUDE.md, AGENTS.md) — добавь туда ссылку на тулкит и правило: «работа с тестами — только через агентов/промпты `prompt-engineering/`». Типовой способ: отдельное правило `.cursor/rules/<NN>-prompt-engineering.mdc` с картой тулкита и обязательными воротами.

## Обновление тулкита

Версия portable core — [VERSION.md](VERSION.md). При обновлении не затирай
`context/projects/<profile>/`, `patterns/approved/` и
`review-feedback/approved/`. После schema change запусти refresh + validator.

## Чек-лист готовности

- [ ] Portable package собран через `package_portable.py` (не raw-copy)
- [ ] Создан inventory + `context/projects/<profile>/CONTEXT-MANIFEST.yaml`
- [ ] Каждый факт имеет источник или `UNKNOWN`
- [ ] Адаптер AI-инструмента подключён
- [ ] Structural validator PASS на собранном package
- [ ] Eval smoke выполнен или честно `NOT_RUN`
- [ ] Ручные примеры (если есть) прошли redaction/provenance/approval
- [ ] Ссылка из регламента проекта добавлена
