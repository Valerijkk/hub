# Changelog — pro-qa-ai

Формат: Keep a Changelog · версии SemVer.

## [0.4.0-beta.1] — 2026-08-04

### Changed (BREAKING)

- Папка переименована `prompt-engineering` → `pro-qa-ai`
- Структура верхнего уровня: 15 папок → 9. Прежние `rules/`, `practices/`,
  `patterns/`, `review-feedback/`, `playbooks/`, `recipes/`, `prompts/`,
  `planning/` на верхнем уровне не существуют — они внутри `standards/`,
  `workflows/` и `work/`
- `review-feedback/` переименован в `standards/team-review/`
- Переносимый пакет: 148 → 184 файла

### Removed (BREAKING)

- `evals/` — протокол замера, ни разу не прогнанный
- `model-routing.md` и `models/`: выбор модели задаёт среда запуска,
  а не текст промпта
- `adapters/` и машинерия task-packet: существовала только ради адаптера
  generic-chat
- Проверка `cursor_adapter` из валидатора

- Слой ревью схлопнут: `standards/team-review/{inbox,approved}` и
  `TEMPLATE.md` со схемой `RF-*` удалены, остались `threads/` и `checks/`
  со схемой `THREAD-*` / `CHECK-*`. Кто заполнял старый inbox —
  переносит записи в `threads/inbox/`
- Англоязычный обзор `EXECUTIVE-OVERVIEW` удалён: дублировал RU-версию
  по языку, а GOVERNANCE и ROI-METRICS — по содержанию, и уже разошёлся
  с фактами (55 тестов вместо 41, 162 файла вместо 184)

### Added

- `standards/team-review/threads/` — треды из реальных ревью, читаются
  ревьюером ПЕРЕД общими практиками
- `standards/practices/` — 10 файлов пар «плохо / хорошо», включая
  security-tests, coverage, anti-ai
- `work/planning/` — backlog, bugs, discussions, state
- `work/reasoning/` — размышления агента, необязательные к исполнению
- Команды `tf-*`: init, status, cover, cases, triage, review, gate
- `COMMANDS.md` — маршрутизатор команд
- Роль `test-sdet` — пишет код автотестов и фреймворка под них
- `standards/code/` — 13 файлов: ООП, SOLID, DRY/KISS, параметризация,
  техники тест-дизайна, шаги плюс устройство фреймворка
- `standards/test-cases/` — формат кейса и три разобранных примера
- `standards/team-review/checks/` — реестр проверяемых правил `CHECK-*`
- Проверка `inline_paths` в валидаторе: пути в обратных кавычках раньше
  не проверялись, и три из них протухли незамеченными

## [Unreleased]

### Changed

- Проверка утверждения продуктового кода убрана: в репозитории тестового фреймворка её
  выполнить нельзя. Заменена проверкой осмысленности утверждения по управляемому
  входу теста — фикстура, стаб, записанный контракт, тестовые данные
  (`rules/40`). Уровень L2, не L3: она доказывает зависимость утверждения
  от наблюдаемого значения, но не соответствие стаба реальному контракту
- `test-verification` научился разбирать значения матрицы генератора;
  добавлен статус `APPROVED WITH QUESTIONS`
- Каталог рецептов переведён на роли, существующие в Standard
- Четыре рецепта получили процедуру и проверяемый критерий готовности
  (bump `1.0.0` → `1.1.0` у каждого)

## [0.3.0-beta.1] — 2026-07-31

### Removed (BREAKING)

- Слой разработки вынесен в издание Pro — 26 файлов: 12 продуктовых ролей
  (orchestrator, product-owner, architect, backend-dev, frontend-dev,
  qa-manual, qa-auto, code-reviewer, debugger, security-reviewer, devops,
  docs-writer), скилы `phase-plan` / `wave-orchestrate` / `verify-goal`,
  плейбуки corporate-delivery / gsd-lifecycle / wave-execution /
  handoff-state / portable-kit-dry-run, категория промптов 13,
  `gsd-port/`, `PORTABLE-HUB-KIT.md`, три corporate-шаблона запуска
- Cursor-адаптер разделён: в Standard остались 4 маппинга `test-*`,
  12 продуктовых уехали в Pro

### Changed

- Граница изданий проведена по **объекту работы**: Standard — тестовый
  фреймворк, Pro — продукт, который разрабатывают
- Версия `rc` → `beta`: release candidate обещает больше, чем позволяют
  собственные ворота при `paired eval: NOT_RUN`
- Portable package: 174 → 148 файлов

### Migration

Ставившим `0.2.0-rc.1` ради корпоративного ростера — установить Pro
рядом со Standard. Роли не удалены, а переехали.

## [0.2.0-rc.1] — 2026-07-30

### Added

- Продуктовая идентичность **Touchstone Standard** (`TOUCHSTONE.md`)
- Production entry: обновлённые START-HERE / README / VERSION / MANIFEST

### Changed

- Позиционирование: pre-release toolkit → foundation-ready Standard для
  линейки Touchstone (тестовые фреймворки)
- Честный release gate: structural PASS ≠ effectiveness PASS

### Not claimed

- Paired eval / ROI по-прежнему `NOT_RUN`
- GA `1.0.0` не объявлен до eval + owner sign-off

## [0.1.0-alpha.1] — 2026-07-26

### Added

- AI Quality Layer: agents, skills, rules, playbooks, recipes, prompts
- Context bootstrap/refresh, pattern cards, manual colleague feedback
- Structural validator, packager, context inventory, task-packet builder
- Eval protocol + fixtures; corporate/GSD portable layer
- Adapters: Cursor, CI, generic-chat
