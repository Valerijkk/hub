# Changelog — Touchstone Pro

## [0.3.0-beta.1] — 2026-07-31

### Added

- Работающий состав вместо одной документации: 12 ролей разработки,
  скилы `phase-plan` / `wave-orchestrate` / `verify-goal`, пять
  плейбуков поставки, фрагменты категории 13, `gsd-port/`,
  `PORTABLE-HUB-KIT.md`, Cursor-адаптер на 12 ролей
- Индексы разделов: `agents/`, `skills/`, `playbooks/`, `prompts/`, `adapters/`

### Changed

- Позиционирование: «надстройка оркестрации над Standard» → **издание
  для разработки продукта**. Граница с Standard идёт по объекту работы
- Совместимость: жёсткое равенство major/minor заменено на минимальную
  версию Standard в `depends_on`
- `rc` → `beta` при `paired eval: NOT_RUN`

## [0.2.0-rc.1] — 2026-07-30

### Added

- Первая раскладка **Prompt Engineering Pro** для линейки Touchstone
- Слои 01–07: core, GSD max, prompt-contract, orchestration, verification/mutation,
  planning ops, Cursor adapters
- Extracts: prompt-contract, master-delegate, wave-protocol, mutation-evidence,
  backlog-task
- SOURCES.md — карта GSD / PE / Cursor / planning
- RELATIONSHIP.md — контракт зависимости от Standard `0.2.0-rc.1`

### Not claimed

- Отдельный paired eval Pro-пакета (`NOT_RUN`)
- GA `1.0.0`
