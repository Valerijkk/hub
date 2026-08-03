# playbooks/ — сквозные процессы

> Плейбук отвечает на «как прогнать процесс целиком»: какие агенты/скилы, в каком порядке, с какими гейтами. Агенты знают свою роль; плейбук знает конвейер.

| Плейбук | Процесс | Участники |
|---------|---------|-----------|
| [test-generation.md](test-generation.md) | Документация → верифицированные TC → автотесты | tc-generator, test-verification, reviewer |
| [review-cycle.md](review-cycle.md) | Ревью-цикл тестового кода до мержа | anti-ai-check, naming-check, reviewer |
| [debugging.md](debugging.md) | Красный CI → root cause → зелёный с профилактикой | debugger, reviewer, documenter |
| [documentation.md](documentation.md) | Поддержание доков фреймворка в синхроне | documenter, refresh-test-context |

Плейбуки поставки — corporate-delivery, GSD lifecycle, волны, handoff/STATE
и dry-run portable kit — переехали в издание Pro
(`prompt-engineering-pro/playbooks/`) вместе с ролями разработки.

Общие правила конвейеров:
1. Гейт не пропускает — процесс возвращается на шаг владельца дефекта, не «проталкивается».
2. Каждый переход между агентами — через артефакт (отчёт/набор/вердикт), не через пересказ.
3. Верификатор ≠ автор (независимость проверки).
4. **Verification по риску** ([rules/40](../../standards/rules/40-verification-protocol.md)):
   утверждения имеют evidence; непрогнанное = `NOT_RUN`; critical/high-impact
   требует independent gate.
5. Human input может быть natural language; строгий master contract обязателен для delegation/CI/handoff.
6. Модель выбирается только из актуального runtime catalog; core agents model-agnostic.
7. External/destructive/prod действия остаются за явным approval gate.
