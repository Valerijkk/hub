# GPT-6 Astra: рабочие правила Hub

Дата сверки: 2026-09-06. Источники: [Model guidance](https://developers.openai.com/api/docs/guides/latest-model), [модель GPT-6 Astra](https://developers.openai.com/api/docs/models/gpt-6-astra), [Codex config schema](https://learn.chatgpt.com/docs/config-schema.json), [config reference](https://learn.chatgpt.com/docs/config-file/config-reference).

## Маршрутизация

- Root использует `gpt-6-astra` / `max`; у Astra допустимы `low`, `medium`, `high`, `xhigh`, `max`.
- При миграции API `none` или `minimal` переводятся в `low`; иной текущий эффективный effort сохраняется.
- Tiered routing downstream остаётся действующим. Существующие workload GPT-5.6 не переводятся в Astra молча.
- В Codex `service_tier = "default"`, default downstream — Terra / `medium`, threads = 3. Эти значения здесь не меняются.
- Схема принимает nonempty `ReasoningEffort`; список Astra выше подтверждён страницей модели, включая `max`.
- Локальные безопасные model-поля уже совпадают. Repo TOML не переключает работающую сессию: launch/managed override и observed telemetry учитываются отдельно.
- `agents.max_depth` в схеме относится к V1 и игнорируется V2. Глубину и fan-out V2 ограничивает `TaskContract`; не делайте ложный вывод, что TOML ограничит V2 runtime.

## Если позже мигрируется API-клиент

- Astra tool calling требует Responses API. Удаляйте `temperature`, `top_p`, `top_logprobs`; для Chat Completions также `logprobs`, для Responses — `message.output_text.logprobs` из `include`.
- Только при миграции с GPT-5.5 или более ранней модели заменяйте `prompt_cache_retention` на `prompt_cache_options.ttl: "30m"`. Это API-параметр, не настройка Codex TOML.
- Async tool calling, steering и cache APIs — возможности harness. Не добавляйте для них спекулятивные настройки.
- Этот документ не меняет продуктовый AI-клиент: отдельная миграция требует scope, проверки и evidence.

## Практика prompting

- Продолжайте согласованную незаблокированную работу; фиксируйте допущения, вопросы и блокеры в артефактах.
- Спрашивайте пользователя только о недостающем вводе, который существенно меняет итог; steering сохраняет готовую часть работы.
- Пишите коротко по-русски, с нормальными пробелами. Делегируйте только ограниченную независимую работу, когда это разрешено и полезно.
- Проверки соразмеряйте риску; не создавайте повторяющие реализацию косметические тесты. Проверяйте достижимые skills/rules на конфликты, приоритет у прямых инструкций пользователя.

## Provenance

Для этой синхронизации requested runtime: `gpt-5.6-terra` / `medium`; actual/provider telemetry: `UNKNOWN`; provenance: `dispatcher-enforced`. Runtime-проверки не выполнялись.

Связанное устаревшее упоминание маршрутизации в `.cursor/rules/70-pro-qa-ai.mdc` оставлено без изменений: protected policy и security boundaries не входят в этот MR.
