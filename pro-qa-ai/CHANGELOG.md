# Changelog — pro-qa-ai

Формат: Keep a Changelog · версии SemVer.

## [0.4.0-beta.2] — 2026-08-07

### Added

- Safe filesystem boundary и отдельный secret scanner с негативными тестами.
- 14-шаговый clean-room portable release protocol.
- Evidence/enforcement и negative-sentinel practices; хорошие и плохие
  примеры для изменяемого кода, функций и singleton trade-offs.

### Changed

- Portable publish теперь transactional: sibling staging, exact file/directory
  inventory, size/SHA-256 verification, atomic rename и fail-closed cleanup.
- Provenance gate отделяет adapted ideas от substantial copied GSD material;
  неизвестная exact revision блокирует только substantial material.
- Generated project profiles удаляются generic rule без Hub-specific assumption.
- SOLID/DRY/размер функций/mocks/POM — review guidance; correctness, security,
  failure boundaries и воспроизводимость остаются hard controls.
- Устаревающие file/test counts вынесены из canonical docs в release evidence.

## [0.4.0-beta.1] — 2026-08-04

### Changed (BREAKING)

- Папка переименована `prompt-engineering` → `pro-qa-ai`
- Структура верхнего уровня: 15 папок → 9. Прежние `rules/`, `practices/`,
  `patterns/`, `review-feedback/`, `playbooks/`, `recipes/`, `prompts/`,
  `planning/` на верхнем уровне не существуют — они внутри `standards/`,
  `workflows/` и `work/`
- `review-feedback/` переименован в `standards/team-review/`
- Переносимый пакет: 148 → 187 файлов
- Роли и скилы приведены к единой форме: 9 фиксированных секций у роли,
  7 у скила, с чеклистом готовности и вердиктом у каждого
- Удалён слой совместимости `context/*.md` (4 файла): ничто его не
  заполняло, а четыре скила несли ради него колонку fallback

- Вердикт проверки утверждения: `НЕ_УТВЕРЖДЕНИЕ_ЕСТЬ` → `УТВЕРЖДЕНИЯ_НЕТ`.
  Два написания одного значения жили параллельно: роли использовали одно,
  скилы и правило 40 — другое, неграмматичное. Отчёты со старым значением
  надо поправить

- `scripts/scan_secrets.py` — поиск конфиденциальных значений в рабочем
  дереве и во **всей истории**: убранное из последней версии остаётся в
  предыдущих ревизиях, и поиск по HEAD даёт ложное «чисто»
- `rules/50` §3.1 минимизация и §3.2 порядок действий при утечке:
  ротация раньше очистки, правку истории агент не выполняет никогда
- `work/journal/` — журнал по дням: единственная ось времени в `work/`,
  остальное разложено по темам. Обязательны пять классов записей, включая
  отказ с причиной: не сделанное иначе исчезает бесследно
- `standards/rules/60-agency-and-egress.md` — полномочия агента и выход
  наружу. Неотменяемо наравне с 40 и 50: отправленное не возвращается, а
  установленный пакет уже в проекте
- Проверка `egress`: упоминание сетевого запроса, установки пакета или
  отправки наружу без рядом стоящего разрешения либо запрета
- Проверка `handoff`: переход между ролями обязан быть описан с обеих
  сторон

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
- `standards/code/suites.md` — разделение сьютов по цене прогона
- `workflows/playbooks/phase-loop.md` — хребет: пять шагов на объём и
  ворота между ними; остальные плейбуки раскрывают отдельные шаги
- `standards/practices/portability.md` — тест не зависит от машины:
  временные каталоги, литералы путей, переводы строк, биты прав POSIX
- Свойства вместо примеров как техника тест-дизайна: обратимость,
  монотонность, идемпотентность, инвариант границы
- Часы как параметр вместо подмены системного времени
- Политика «плохой тест удаляется и заменяется, а не отключается»
- Правила, вынесенные из GSD (`SRC-GSD-CONTEXT-001`, `SRC-GSD-SUITES-001`,
  `SRC-GSD-STANDARDS-001`): запрет проверять исходник грепом вместо вызова;
  структурные утверждения вместо сравнения текста ошибки; обязательные
  граничные фикстуры для параметров-лимитов; тест паритета при намеренном
  дублировании; таймауты на внешние вызовы; поправка к решению отдельной
  записью без переписывания исходной
- Рецепты маршрутизировали реализацию в удалённую роль `qa-auto`; теперь
  в `test-sdet`, который для этого и заведён
- Проверка `handoff`: переход между ролями обязан быть описан с обеих
  сторон. Нашла три односторонних — ревьюер не значился отдающим для
  `test-sdet`, дебаггер не значился входом для ревьюера, документатор не
  значился входом для `test-sdet`
- Шапка «не раскрывать скрытую цепочку рассуждений» стояла в девяти
  рецептах и разошлась по формулировке; требование общее и теперь
  записано один раз в `_SCHEMA.md`
- Проверка `duplication`: повтор абзаца между файлами от 50% пересечения.
  Копия стоит контекста дважды и расходится при правке. Сразу нашла
  стопроцентный дубль — `tf-review` пересказывал условия блокировки из
  правила 30; заменено ссылкой
- Проверки `vocabulary`, `enums`, `roster`, `code_spans` — четыре класса
  смыслового расхождения, которые ловились только чтением: упразднённое
  понятие в тексте, два написания одного машинного значения, ссылка на
  несуществующую роль, разорванная вставка кода. Двенадцать мутаций
  подтверждают срабатывание
- Проверка `size_budget`: ярусы 38 / 54 / 90 КБ, исключение обязано нести
  записанную причину, потолок непробиваем. Три кода подтверждены мутацией
- `rules/00` §6 «Экономия собственного контекста»: грузить по задаче, а не
  впрок; передавать пути, а не содержимое; результат в артефакт;
  остановиться раньше, чем испортить
- Проверки `agent_contract` и `skill_contract`: состав секций, их порядок
  и наличие чеклиста. Все шесть кодов подтверждены мутацией
- Починена `check_recipes`: искала рецепты в `recipes/`, а они с
  реорганизации лежат в `workflows/recipes/` — проверка молчала

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
