# Changelog — Touchstone Standard (`prompt-engineering/`)

Формат: Keep a Changelog · версии SemVer.

## [Unreleased]

### Changed

- Мутация продуктового кода убрана: в репозитории тестового фреймворка её
  выполнить нельзя. Заменена проверкой чувствительности по управляемому
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
- Связка с изданием **Touchstone Pro** (sibling `prompt-engineering-pro/`)
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
