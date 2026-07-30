# START HERE — Touchstone Pro

Не читай всё дерево. Выбери сценарий.

## Сценарий A — поставить Pro поверх Standard

1. Установи [Standard](../prompt-engineering/INSTALL.md) в целевой репозиторий.
2. Скопируй эту папку рядом:

```text
<repo>/
├── prompt-engineering/          # Standard (обязателен)
└── prompt-engineering-pro/      # Pro (этот пакет)
```

3. Прочитай [RELATIONSHIP.md](RELATIONSHIP.md) — что откуда грузить.
4. Подключи адаптер Cursor/CI из Standard `adapters/`.
5. Для волн/оркестрации используй extracts:
   - [`extracts/prompt-contract.md`](extracts/prompt-contract.md)
   - [`extracts/master-delegate.md`](extracts/master-delegate.md)
   - [`extracts/wave-protocol.md`](extracts/wave-protocol.md)
   - [`extracts/mutation-evidence.md`](extracts/mutation-evidence.md)
   - [`extracts/backlog-task.md`](extracts/backlog-task.md)

## Сценарий B — понять, чем Pro лучше Standard

Открой [README.md](README.md) → таблица источников → [layers/02-gsd-max.md](layers/02-gsd-max.md)
и [layers/05-verification-mutation.md](layers/05-verification-mutation.md).

## Сценарий C — вести фазу как в Hub

1. [layers/02-gsd-max.md](layers/02-gsd-max.md) — lifecycle  
2. [layers/04-team-orchestration.md](layers/04-team-orchestration.md) — волны  
3. [layers/06-planning-ops.md](layers/06-planning-ops.md) — STATE / backlog  
4. [layers/05-verification-mutation.md](layers/05-verification-mutation.md) — доказательство

## Сценарий D — только тестовый фреймворк

Вернись в Standard: [`../prompt-engineering/START-HERE.md`](../prompt-engineering/START-HERE.md).
Pro не заменяет bootstrap/context/recipes — он их усиливает оркестрацией и
verification rigor.

## Запрещено

- Ставить Pro без Standard
- Копировать Hub `.cursor/rules/60-tz-constraints.mdc` в чужой продукт «как есть»
- Заявлять ROI Pro без отдельного paired eval
- Считать галочку в отчёте агента доказательством без mutation/layer evidence
