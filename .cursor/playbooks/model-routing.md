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

## Конфигурация и параллельная работа

Project defaults находятся в `.codex/config.toml`: root Astra/Max, стандартный
service tier, downstream default Terra/Medium, максимум три субагента плюс root.
Лимит применяется совместно с фактически доступными слотами. Только root
создаёт субагентов; downstream не запускает вложенную волну.

Делегируй независимую bounded-задачу, когда это ускоряет работу или даёт нужное
независимое review. Не заполняй слоты дублирующим поиском или формальными задачами.
Разводи ownership; тяжёлые Docker-проверки выполняй последовательно с согласованными
CPU/RAM limits и учётом соседних веток. Число агентов не разрешает параллельные сборки.

`model_profile: "inherit"` в `.planning/config.json` — selector GSD, а не разрешение
наследовать runtime при dispatch. Defaults клиента и ответы GSD resolver не
заменяют явные `model` + `reasoning_effort` из текущего routing contract.
Если adapter возвращает `inherit` или устаревшую модель, root задаёт разрешённый
точный runtime через tool override до запуска. Недоступный runtime даёт
`BLOCKED`; смена модели требует нового явного routing, не silent fallback.

Область действия defaults и приоритет explicit spawn описаны в
[Codex Config Reference](https://learn.chatgpt.com/docs/config-file/config-reference).
Наличие TOML не доказывает, что работающая сессия его загрузила: для этого нужна
отдельная telemetry. Project config применяется только в trusted project;
явные launch overrides имеют приоритет над ним. См.
[Config basics](https://learn.chatgpt.com/docs/config-file/config-basic).
