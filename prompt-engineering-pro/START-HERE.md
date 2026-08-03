# START HERE — Touchstone Pro

Издание для **разработки продукта**. Не читай всё дерево — выбери сценарий.

## Сценарий A — поставить Pro поверх Standard

1. Установи [Standard](../pro-qa-ai/INSTALL.md) в целевой репозиторий.
2. Скопируй эту папку рядом:

```text
<repo>/
├── pro-qa-ai/          # Standard (обязателен)
└── prompt-engineering-pro/      # Pro (этот пакет)
```

3. Прочитай [RELATIONSHIP.md](RELATIONSHIP.md) — что откуда грузить.
4. Подключи роли: [`adapters/cursor/agents.map.yaml`](adapters/cursor/agents.map.yaml)
   ставит 12 ролей разработки в `.cursor/agents/`. Адаптер Standard ставит
   рядом четыре `test-*` — имена не пересекаются, конфликта нет.
5. Проверь `depends_on` в [VERSION.md](VERSION.md) против версии Standard.

## Сценарий B — собрать команду разработки

1. [`agents/README.md`](agents/README.md) — ростер и граница со Standard
2. [`playbooks/corporate-delivery.md`](playbooks/corporate-delivery.md) —
   scope → design → build → test → review → verify → handoff
3. [`agents/templates/`](agents/templates/) — шаблоны запуска фазы и волны

## Сценарий C — вести фазу

1. [`skills/phase-plan/SKILL.md`](skills/phase-plan/SKILL.md) — атомарный план
2. [`skills/wave-orchestrate/SKILL.md`](skills/wave-orchestrate/SKILL.md) — волна
3. [`skills/verify-goal/SKILL.md`](skills/verify-goal/SKILL.md) — goal-backward приёмка
4. [`extracts/mutation-evidence.md`](extracts/mutation-evidence.md) — доказательство фикса

Теория за этим порядком — [`layers/02-gsd-max.md`](layers/02-gsd-max.md).

## Сценарий D — только тестовый фреймворк

Pro не нужен. Иди в Standard:
[`../pro-qa-ai/START-HERE.md`](../pro-qa-ai/START-HERE.md).

Pro покрывает другой объект работы — продукт, а не фреймворк. Если задача
про сам фреймворк, роли отсюда её только запутают.

## Сценарий E — понять, что Pro добавляет

[README.md](README.md) → раздел «Приколы Cursor» → [SOURCES.md](SOURCES.md).

## Запрещено

- Ставить Pro без Standard
- Дублировать в Pro то, что уже есть в Standard (context, rules, recipes,
  evals, scripts) — только ссылка
- Копировать Hub `.cursor/rules/60-tz-constraints.mdc` в чужой продукт «как есть»
- Заявлять ROI Pro без отдельного paired eval
- Считать галочку в отчёте агента доказательством без mutation/layer evidence
- Отдавать ревью продуктового diff роли `test-code-reviewer`, а код
  фреймворка — роли `code-reviewer`: это ошибка маршрутизации, а не мелочь
