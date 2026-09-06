# Hub — маршрутизация моделей

`team-lead`/root использует `gpt-6-astra` с `max`; `ultra` не используется.

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

В handoff указывай observed runtime и telemetry; без неё — `actual_model: UNKNOWN`.
Нерешённый architecture, security или product вопрос возвращается оркестратору:
это граница scope, а не правило конкретной модели.

Исторические model ADR описывают прошлые решения и не переопределяют этот файл
или текущую founder-инструкцию.
