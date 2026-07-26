# rules/ — правила агентов

> Правила грузятся агентами всегда (00) или по роли. Нумерация — по образцу Cursor rules: 00 глобальные, дальше тематические. Правило ≠ пожелание: нарушение — находка в ревью/верификации.
> Язык: все инструкции на русском (`00` §0); машинные id/enum не переводятся.

| Файл | Область | Кто обязан |
|------|---------|------------|
| [00-global-agent-rules.md](00-global-agent-rules.md) | Поведение любого агента тулкита | все |
| [10-naming-conventions.md](10-naming-conventions.md) | Дефолтный канон нейминга (пока проектный не заполнен) | reviewer, naming-check |
| [20-test-quality.md](20-test-quality.md) | Минимальные требования к качеству теста | reviewer, generator, debugger |
| [30-severity-policy.md](30-severity-policy.md) | Единая шкала severity и политика блокировки | все, кто выдаёт находки |
| [40-verification-protocol.md](40-verification-protocol.md) | Risk-scaled taxonomy проверок + evidence block; independent gate для high-impact | все агенты и скилы |
| [50-data-handling.md](50-data-handling.md) | Недоверенный контекст, ручные примеры коллег, redaction/provenance/approval | все, кто сканирует код или работает с examples |

Приоритет при конфликте: active `context/projects/<profile>/` + linked approved
cards (проектный канон) > `rules/` (дефолты тулкита). Расхождение фиксируется
явно, молча не разрешается.
