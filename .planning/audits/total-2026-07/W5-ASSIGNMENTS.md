# W5 Assignments — ai + search

> Стадия A · read-only · TASK-008

## A1 — ai handlers (11)
Отчёт: `A5-1-ai-handlers.md`

- `backend/internal/ai/admin_handler.go`
- `backend/internal/ai/dto.go`
- `backend/internal/ai/export_hooks.go`
- `backend/internal/ai/handler.go`
- `backend/internal/ai/user_handler.go`
- `backend/testframework/ai/admin_handler_test.go`
- `backend/testframework/ai/admin_handler_wave110_test.go`
- `backend/testframework/ai/coverage_wave81_test.go`
- `backend/testframework/ai/handler_extended_test.go`
- `backend/testframework/ai/handler_wave115_test.go`
- `backend/testframework/ai/user_handler_test.go`

## A2 — ai adapters/stream (20)
Отчёт: `A5-2-ai-adapters.md`

- `backend/internal/ai/adapter_anthropic.go`
- `backend/internal/ai/adapter_gigachat.go`
- `backend/internal/ai/adapter_openai.go`
- `backend/internal/ai/adapter_stream.go`
- `backend/internal/ai/adapter_yandex.go`
- `backend/internal/ai/content_policy.go`
- `backend/internal/ai/fallback.go`
- `backend/internal/ai/privacy.go`
- `backend/internal/ai/provider_rpm.go`
- `backend/internal/ai/stream.go`
- `backend/testframework/ai/adapter_anthropic_test.go`
- `backend/testframework/ai/adapter_http_test.go`
- `backend/testframework/ai/adapter_openai_test.go`
- `backend/testframework/ai/adapter_wave110_test.go`
- `backend/testframework/ai/adapter_yandex_test.go`
- `backend/testframework/ai/content_policy_test.go`
- `backend/testframework/ai/fallback_test.go`
- `backend/testframework/ai/privacy_test.go`
- `backend/testframework/ai/provider_rpm_test.go`
- `backend/testframework/ai/stream_test.go`

## A3 — ai service/repo/bot (29)
Отчёт: `A5-3-ai-service.md`

- `backend/internal/ai/admin_settings.go`
- `backend/internal/ai/audit.go`
- `backend/internal/ai/bot_commands.go`
- `backend/internal/ai/bot_user.go`
- `backend/internal/ai/code_review.go`
- `backend/internal/ai/composer_suggest.go`
- `backend/internal/ai/doc_search.go`
- `backend/internal/ai/issue_lookup.go`
- `backend/internal/ai/messenger_bot.go`
- `backend/internal/ai/preference_input.go`
- `backend/internal/ai/repository.go`
- `backend/internal/ai/repository_ent.go`
- `backend/internal/ai/review_context.go`
- `backend/internal/ai/service.go`
- `backend/internal/ai/settings_store.go`
- `backend/testframework/ai/admin_settings_test.go`
- `backend/testframework/ai/audit_test.go`
- `backend/testframework/ai/bot_commands_test.go`
- `backend/testframework/ai/bot_user_test.go`
- `backend/testframework/ai/code_review_test.go`
- `backend/testframework/ai/code_review_unified_test.go`
- `backend/testframework/ai/composer_suggest_test.go`
- `backend/testframework/ai/doc_search_test.go`
- `backend/testframework/ai/issue_lookup_test.go`
- `backend/testframework/ai/messenger_bot_flow_test.go`
- `backend/testframework/ai/messenger_bot_test.go`
- `backend/testframework/ai/repository_ent_test.go`
- `backend/testframework/ai/service_test.go`
- `backend/testframework/ai/settings_store_test.go`

## A4 — search prod+half tests (22)
Отчёт: `A5-4-search-core.md`

- `backend/internal/search/dto.go`
- `backend/internal/search/export_hooks.go`
- `backend/internal/search/handler.go`
- `backend/internal/search/indexer.go`
- `backend/internal/search/jql.go`
- `backend/internal/search/meilisearch_indexer.go`
- `backend/internal/search/repository.go`
- `backend/internal/search/service.go`
- `backend/testframework/search/handler_flow_test.go`
- `backend/testframework/search/handler_wave100_test.go`
- `backend/testframework/search/handler_wave109_test.go`
- `backend/testframework/search/handler_wave112_test.go`
- `backend/testframework/search/handler_wave116_test.go`
- `backend/testframework/search/handler_wave93_test.go`
- `backend/testframework/search/jql_assignee_a4001_test.go`
- `backend/testframework/search/jql_predicates_extended_test.go`
- `backend/testframework/search/jql_test.go`
- `backend/testframework/search/jql_wave100_test.go`
- `backend/testframework/search/jql_wave116_test.go`
- `backend/testframework/search/meilisearch_indexer_test.go`
- `backend/testframework/search/repository_wave100_test.go`
- `backend/testframework/search/service_query_test.go`

## A5 — search remaining tests (14)
Отчёт: `A5-5-search-tests.md`

- `backend/testframework/search/service_scope_test.go`
- `backend/testframework/search/service_shard_errors_test.go`
- `backend/testframework/search/service_test.go`
- `backend/testframework/search/service_wave100_test.go`
- `backend/testframework/search/service_wave103_mongo_test.go`
- `backend/testframework/search/service_wave106_test.go`
- `backend/testframework/search/service_wave108_test.go`
- `backend/testframework/search/service_wave109_test.go`
- `backend/testframework/search/service_wave114_test.go`
- `backend/testframework/search/service_wave115_test.go`
- `backend/testframework/search/service_wave116_postgres_test.go`
- `backend/testframework/search/service_wave92_test.go`
- `backend/testframework/search/soft_delete_a4002_runtime_test.go`
- `backend/testframework/search/soft_delete_a4002_test.go`

**Σ файлов: 96**

