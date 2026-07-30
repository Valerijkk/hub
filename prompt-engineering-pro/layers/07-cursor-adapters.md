# Layer 07 — Cursor / Runtime Adapters

Источники: Standard `adapters/`, `models/`, `playbooks/model-routing.md`,
`.cursor/REGULATION.md` §20, `.cursor/rules/70-prompt-engineering.mdc`.

## Адаптеры (канон в Standard)

| Runtime | Путь |
|---------|------|
| Cursor IDE | `prompt-engineering/adapters/cursor/` |
| CI bot | `prompt-engineering/adapters/ci/` |
| Generic chat | `prompt-engineering/adapters/generic-chat/` |

## Model routing (portable)

- Orchestrator default: GPT-5.6 Sol Pro (если доступен в catalog)
- Downstream: `inherit`; выбор slug — у launcher/router по risk/cost/context
- В теле промпта **не** задавать temperature/top-p/model id
- Выбор модели логировать в handoff

## Cursor wiring в целевом репо (рекомендация)

```
.cursor/
├── rules/70-touchstone.mdc     # ссылка на Standard + Pro layers
├── skills/                     # опционально: symlink/копии skills Standard
└── agents/                     # опционально: mapping на Standard agents
```

Не копировать Hub `60-tz-constraints` и полный REGULATION — только практики
из extracts.

## DoD слоя

Агент в Cursor видит rule на Touchstone и может открыть START-HERE Standard/Pro
без поиска по всему монорепо.
