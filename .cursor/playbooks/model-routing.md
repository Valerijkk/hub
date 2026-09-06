# Hub — маршрутизация моделей

`team-lead`/root использует `gpt-6-astra` с `max`; `ultra` не используется.

Официальный [guide GPT-6 Astra](https://developers.openai.com/api/docs/guides/latest-model)
подтверждает уровни `low`, `medium`, `high`, `xhigh`, `max`. При миграции значение
`none` или `minimal` заменяют на `low`; в остальных случаях сохраняют текущий
эффективный effort. Это не меняет root `max` и не сворачивает существующие
workload GPT-5.6 в Astra без отдельного решения и проверки.
Рабочие границы Hub: [OPENAI-ASTRA-WORKFLOW.md](../../docs/ai/OPENAI-ASTRA-WORKFLOW.md).

Для каждой самостоятельной bounded-подзадачи оркестратор явно выбирает runtime
и task-routed effort по риску, объёму и нужному evidence:

| Работа | Допустимый выбор |
|---|---|
| Сфокусированный анализ, реализация или review | `gpt-6-astra` / `medium` или `high` |
| Bounded-задача, где GPT-5.6 достаточно по evidence и риску | `gpt-5.6-sol` или `gpt-5.6-terra` / `low`, `medium` или `high` |

Роли описывают ownership и evidence, но не выбирают downstream runtime. В routing
фиксируются `requested_model`, `reasoning_effort` и причина выбора. Каждый
subagent dispatch использует bounded/no-history fork: model и effort передаются
явно и не наследуются от root. Не подменяй запрошенные model или effort молча.

`agents.max_depth` в `.codex/config.toml` — параметр V1 и игнорируется V2. Для
V2 depth и fan-out всегда явно ограничиваются в `TaskContract`; это не обещание,
что конфигурация ограничит V2 runtime.

В handoff указывай observed runtime и telemetry; без неё — `actual_model: UNKNOWN`.
Нерешённый architecture, security или product вопрос возвращается оркестратору:
это граница scope, а не правило конкретной модели.

Исторические model ADR описывают прошлые решения и не переопределяют этот файл
или текущую founder-инструкцию.
