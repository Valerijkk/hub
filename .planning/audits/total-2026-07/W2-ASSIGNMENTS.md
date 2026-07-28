# W2 Assignments — tms + bases

> Стадия A · read-only · 2026-07-28 · TASK-008
> Всего файлов: **109** · Агентов: **5**

## A1 — tms HTTP handlers

Отчёт: `A2-1-tms-handlers.md` · **Итого: 24**

### Prod (8)
- `backend/internal/tms/handler.go`
- `backend/internal/tms/handler_import.go`
- `backend/internal/tms/handler_reports.go`
- `backend/internal/tms/handler_runs.go`
- `backend/internal/tms/handler_suites.go`
- `backend/internal/tms/handler_test_cases.go`
- `backend/internal/tms/handler_webhook.go`
- `backend/internal/tms/dto.go`

### Tests (16)
- `backend/testframework/tms/handler_flow_test.go`
- `backend/testframework/tms/handler_import_flow_test.go`
- `backend/testframework/tms/handler_routes_test.go`
- `backend/testframework/tms/handler_wave108_test.go`
- `backend/testframework/tms/handler_wave110_test.go`
- `backend/testframework/tms/handler_wave113_test.go`
- `backend/testframework/tms/handler_wave114_test.go`
- `backend/testframework/tms/handler_wave115_test.go`
- `backend/testframework/tms/handler_wave120_test.go`
- `backend/testframework/tms/handler_wave122_test.go`
- `backend/testframework/tms/handler_wave124_test.go`
- `backend/testframework/tms/handler_wave88_test.go`
- `backend/testframework/tms/handler_webhook_flow_test.go`
- `backend/testframework/tms/handler_webhook_rbac_test.go`
- `backend/testframework/tms/handler_webhook_wave99_test.go`
- `backend/testframework/tms/helpers_test.go`

## A2 — tms services

Отчёт: `A2-2-tms-services.md` · **Итого: 22**

### Prod (8)
- `backend/internal/tms/service.go`
- `backend/internal/tms/service_converters.go`
- `backend/internal/tms/service_import.go`
- `backend/internal/tms/service_reports.go`
- `backend/internal/tms/service_runs.go`
- `backend/internal/tms/service_suites.go`
- `backend/internal/tms/service_test_cases.go`
- `backend/internal/tms/events.go`

### Tests (14)
- `backend/testframework/tms/service_cases_extra_test.go`
- `backend/testframework/tms/service_converters_test.go`
- `backend/testframework/tms/service_reports_test.go`
- `backend/testframework/tms/service_runs_extra_test.go`
- `backend/testframework/tms/service_suites_test.go`
- `backend/testframework/tms/service_test.go`
- `backend/testframework/tms/service_wave113_test.go`
- `backend/testframework/tms/service_wave114_test.go`
- `backend/testframework/tms/service_wave115_test.go`
- `backend/testframework/tms/service_wave116_test.go`
- `backend/testframework/tms/service_wave120_test.go`
- `backend/testframework/tms/service_wave122_extra_test.go`
- `backend/testframework/tms/service_wave124_test.go`
- `backend/testframework/tms/create_bug_nil_tc_a3009_test.go`

## A3 — tms repositories

Отчёт: `A2-3-tms-repos.md` · **Итого: 22**

### Prod (7)
- `backend/internal/tms/repository.go`
- `backend/internal/tms/ent_repository.go`
- `backend/internal/tms/mongo_repository.go`
- `backend/internal/tms/combined_repository.go`
- `backend/internal/tms/environment.go`
- `backend/internal/tms/provisioner.go`
- `backend/internal/tms/export_hooks.go`

### Tests (15)
- `backend/testframework/tms/ent_repository_batch_test.go`
- `backend/testframework/tms/ent_repository_extended_test.go`
- `backend/testframework/tms/ent_repository_list_update_test.go`
- `backend/testframework/tms/ent_repository_test.go`
- `backend/testframework/tms/ent_repository_wave122_extra_test.go`
- `backend/testframework/tms/ent_repository_wave122_test.go`
- `backend/testframework/tms/ent_repository_wave124_test.go`
- `backend/testframework/tms/mongo_repository_test.go`
- `backend/testframework/tms/postgres_ent_helper_test.go`
- `backend/testframework/tms/find_by_key_a3003_test.go`
- `backend/testframework/tms/link_issue_scope_a3003_ent_test.go`
- `backend/testframework/tms/result_scope_a3008_ent_test.go`
- `backend/testframework/tms/result_scope_a3008_test.go`
- `backend/testframework/tms/update_optimistic_a3004_ent_test.go`
- `backend/testframework/tms/update_optimistic_a3004_test.go`

## A4 — tms import/parsers + remaining scope tests

Отчёт: `A2-4-tms-import.md` · **Итого: 25**

### Prod (11)
- `backend/internal/tms/allure_matcher.go`
- `backend/internal/tms/allure_parser.go`
- `backend/internal/tms/allure_types.go`
- `backend/internal/tms/exporter.go`
- `backend/internal/tms/importer.go`
- `backend/internal/tms/junit_parser.go`
- `backend/internal/tms/testrail_import.go`
- `backend/internal/tms/export_hooks_converters.go`
- `backend/internal/tms/export_hooks_mongo.go`
- `backend/internal/tms/export_hooks_repo.go`
- `backend/internal/tms/export_hooks_testrail.go`

### Tests (14)
- `backend/testframework/tms/allure_matcher_wave99_test.go`
- `backend/testframework/tms/allure_parser_test.go`
- `backend/testframework/tms/coverage_wave_test.go`
- `backend/testframework/tms/environment_wave105_test.go`
- `backend/testframework/tms/execute_by_a3017_test.go`
- `backend/testframework/tms/exporter_test.go`
- `backend/testframework/tms/exporter_wave95_test.go`
- `backend/testframework/tms/import_run_scope_a3012_test.go`
- `backend/testframework/tms/importer_wave96_test.go`
- `backend/testframework/tms/junit_parser_test.go`
- `backend/testframework/tms/rollback_a3007_ent_test.go`
- `backend/testframework/tms/rollback_a3007_test.go`
- `backend/testframework/tms/run_scope_a3012_ent_test.go`
- `backend/testframework/tms/testrail_import_wave96_test.go`

## A5 — bases (полный модуль)

Отчёт: `A2-5-bases.md` · **Итого: 16**

### Prod (5)
- `backend/internal/bases/dto.go`
- `backend/internal/bases/export_hooks.go`
- `backend/internal/bases/handler.go`
- `backend/internal/bases/repository.go`
- `backend/internal/bases/service.go`

### Tests (11)
- `backend/testframework/bases/fixtures_test.go`
- `backend/testframework/bases/handler_flow_test.go`
- `backend/testframework/bases/handler_test.go`
- `backend/testframework/bases/handler_wave107_test.go`
- `backend/testframework/bases/handler_wave115_test.go`
- `backend/testframework/bases/rbac_test.go`
- `backend/testframework/bases/repository_table_cap_a13002_test.go`
- `backend/testframework/bases/repository_wave98_test.go`
- `backend/testframework/bases/service_test.go`
- `backend/testframework/bases/service_wave115_test.go`
- `backend/testframework/bases/service_wave98_test.go`

## Контроль

| Агент | Файлов |
|-------|-------:|
| A1 | 24 |
| A2 | 22 |
| A3 | 22 |
| A4 | 25 |
| A5 | 16 |
| **Σ** | **109** |
