# W1 Assignments — tracker + sprint + project

> Стадия A · read-only · 2026-07-28 · TASK-008
> Всего файлов: **163** · Агентов: **5**

## A1 — tracker HTTP (handlers)

Отчёт: `A1-1-tracker-handlers.md`

### Prod (16)
- `backend/internal/tracker/handler.go`
- `backend/internal/tracker/handler_automation.go`
- `backend/internal/tracker/handler_boards.go`
- `backend/internal/tracker/handler_comments.go`
- `backend/internal/tracker/handler_custom_fields.go`
- `backend/internal/tracker/handler_issue_message_link.go`
- `backend/internal/tracker/handler_issues.go`
- `backend/internal/tracker/handler_labels.go`
- `backend/internal/tracker/handler_links.go`
- `backend/internal/tracker/handler_time.go`
- `backend/internal/tracker/handler_workflow.go`
- `backend/internal/tracker/dto.go`
- `backend/internal/tracker/events.go`
- `backend/internal/tracker/export_hooks.go`
- `backend/internal/tracker/provisioner.go`
- `backend/internal/tracker/workflow.go`

### Tests (20)
- `backend/testframework/tracker/handler_automation_flow_test.go`
- `backend/testframework/tracker/handler_flow_test.go`
- `backend/testframework/tracker/handler_jql_flow_test.go`
- `backend/testframework/tracker/handler_rbac_test.go`
- `backend/testframework/tracker/handler_routes_test.go`
- `backend/testframework/tracker/handler_wave103_test.go`
- `backend/testframework/tracker/handler_wave106_test.go`
- `backend/testframework/tracker/handler_wave109_test.go`
- `backend/testframework/tracker/handler_wave112_test.go`
- `backend/testframework/tracker/handler_wave114_test.go`
- `backend/testframework/tracker/handler_wave115_test.go`
- `backend/testframework/tracker/handler_wave117_test.go`
- `backend/testframework/tracker/handler_wave118_test.go`
- `backend/testframework/tracker/handler_wave119_test.go`
- `backend/testframework/tracker/handler_wave120_test.go`
- `backend/testframework/tracker/handler_wave121_test.go`
- `backend/testframework/tracker/handler_wave124_test.go`
- `backend/testframework/tracker/handler_wave95_test.go`
- `backend/testframework/tracker/handler_workflow_labels_test.go`
- `backend/testframework/tracker/provisioner_test.go`

**Итого A1: 36**

## A2 — tracker services

Отчёт: `A1-2-tracker-services.md`

### Prod (13)
- `backend/internal/tracker/service.go`
- `backend/internal/tracker/service_activity.go`
- `backend/internal/tracker/service_boards.go`
- `backend/internal/tracker/service_comments.go`
- `backend/internal/tracker/service_converters.go`
- `backend/internal/tracker/service_custom_fields.go`
- `backend/internal/tracker/service_issue_message_link.go`
- `backend/internal/tracker/service_issues.go`
- `backend/internal/tracker/service_labels.go`
- `backend/internal/tracker/service_links.go`
- `backend/internal/tracker/service_time.go`
- `backend/internal/tracker/service_workflow.go`
- `backend/internal/tracker/automation_service.go`

### Tests (20)
- `backend/testframework/tracker/service_converters_extended_test.go`
- `backend/testframework/tracker/service_converters_test.go`
- `backend/testframework/tracker/service_helpers_test.go`
- `backend/testframework/tracker/service_issue_message_link_test.go`
- `backend/testframework/tracker/service_test.go`
- `backend/testframework/tracker/service_wave100_test.go`
- `backend/testframework/tracker/service_wave109_test.go`
- `backend/testframework/tracker/service_wave112_test.go`
- `backend/testframework/tracker/service_wave113_test.go`
- `backend/testframework/tracker/service_wave114_test.go`
- `backend/testframework/tracker/service_wave115_test.go`
- `backend/testframework/tracker/service_wave117_test.go`
- `backend/testframework/tracker/service_wave118_test.go`
- `backend/testframework/tracker/service_wave119_test.go`
- `backend/testframework/tracker/service_wave121_test.go`
- `backend/testframework/tracker/service_workflow_extra_test.go`
- `backend/testframework/tracker/comment_scope_a3011_test.go`
- `backend/testframework/tracker/comment_soft_delete_a8006_test.go`
- `backend/testframework/tracker/log_time_a3005_test.go`
- `backend/testframework/tracker/start_timer_concurrent_a3005_test.go`

**Итого A2: 33**

## A3 — tracker repos + automation engine

Отчёт: `A1-3-tracker-repos.md`

### Prod (15)
- `backend/internal/tracker/repository.go`
- `backend/internal/tracker/repo_activity.go`
- `backend/internal/tracker/repo_automation.go`
- `backend/internal/tracker/repo_boards.go`
- `backend/internal/tracker/repo_comments.go`
- `backend/internal/tracker/repo_custom_fields.go`
- `backend/internal/tracker/repo_issue_message_link.go`
- `backend/internal/tracker/repo_issues.go`
- `backend/internal/tracker/repo_labels.go`
- `backend/internal/tracker/repo_links.go`
- `backend/internal/tracker/repo_time.go`
- `backend/internal/tracker/repo_workflow.go`
- `backend/internal/tracker/automation_consumer.go`
- `backend/internal/tracker/automation_dto.go`
- `backend/internal/tracker/automation_engine.go`

### Tests (22)
- `backend/testframework/tracker/automation_ack_a3006_test.go`
- `backend/testframework/tracker/automation_consumer_test.go`
- `backend/testframework/tracker/automation_engine_events_test.go`
- `backend/testframework/tracker/automation_engine_test.go`
- `backend/testframework/tracker/automation_engine_wave100_test.go`
- `backend/testframework/tracker/automation_service_extra_test.go`
- `backend/testframework/tracker/board_label_idor_blockf_test.go`
- `backend/testframework/tracker/link_scope_a3011_postgres_ent_test.go`
- `backend/testframework/tracker/link_scope_a3011_test.go`
- `backend/testframework/tracker/link_target_scope_a3011_postgres_ent_test.go`
- `backend/testframework/tracker/link_target_scope_a3011_test.go`
- `backend/testframework/tracker/postgres_ent_helper_test.go`
- `backend/testframework/tracker/reaction_scope_a3011_postgres_ent_test.go`
- `backend/testframework/tracker/reaction_scope_a3011_test.go`
- `backend/testframework/tracker/repository_board_cap_a13001_test.go`
- `backend/testframework/tracker/repository_extended_test.go`
- `backend/testframework/tracker/repository_more_test.go`
- `backend/testframework/tracker/repository_project_links_a13004_test.go`
- `backend/testframework/tracker/repository_test.go`
- `backend/testframework/tracker/repository_wave100_test.go`
- `backend/testframework/tracker/repository_wave120_test.go`
- `backend/testframework/tracker/repository_wave124_test.go`

**Итого A3: 37**

## A4 — sprint (полный модуль)

Отчёт: `A1-4-sprint.md`

### Prod (6)
- `backend/internal/sprint/dto.go`
- `backend/internal/sprint/export_hooks.go`
- `backend/internal/sprint/handler.go`
- `backend/internal/sprint/poker.go`
- `backend/internal/sprint/repository.go`
- `backend/internal/sprint/service.go`

### Tests (13)
- `backend/testframework/sprint/fail_closed_a3015_test.go`
- `backend/testframework/sprint/fail_closed_a3016_test.go`
- `backend/testframework/sprint/handler_flow_test.go`
- `backend/testframework/sprint/handler_rbac_test.go`
- `backend/testframework/sprint/handler_routes_test.go`
- `backend/testframework/sprint/handler_wave113_test.go`
- `backend/testframework/sprint/list_batch_counts_a13006_test.go`
- `backend/testframework/sprint/repository_ent_test.go`
- `backend/testframework/sprint/repository_get_in_project_a3001_test.go`
- `backend/testframework/sprint/repository_list_no_issues_a13006_test.go`
- `backend/testframework/sprint/service_extended_test.go`
- `backend/testframework/sprint/service_test.go`
- `backend/testframework/sprint/service_wave114_test.go`

**Итого A4: 19**

## A5 — project (полный модуль)

Отчёт: `A1-5-project.md`

### Prod (10)
- `backend/internal/project/dto.go`
- `backend/internal/project/events.go`
- `backend/internal/project/export_hooks.go`
- `backend/internal/project/handler.go`
- `backend/internal/project/invite.go`
- `backend/internal/project/provisioner.go`
- `backend/internal/project/repository.go`
- `backend/internal/project/service.go`
- `backend/internal/project/template_export.go`
- `backend/internal/project/templates.go`

### Tests (28)
- `backend/testframework/project/handler_flow_test.go`
- `backend/testframework/project/handler_wave110_test.go`
- `backend/testframework/project/handler_wave113_test.go`
- `backend/testframework/project/handler_wave120_test.go`
- `backend/testframework/project/handler_wave122_routes_test.go`
- `backend/testframework/project/handler_wave122_test.go`
- `backend/testframework/project/provisioner_test.go`
- `backend/testframework/project/repository_ent_extended_test.go`
- `backend/testframework/project/repository_ent_test.go`
- `backend/testframework/project/repository_wave124_extra_test.go`
- `backend/testframework/project/repository_wave124_test.go`
- `backend/testframework/project/service_extended_test.go`
- `backend/testframework/project/service_fa11_test.go`
- `backend/testframework/project/service_test.go`
- `backend/testframework/project/service_wave101_test.go`
- `backend/testframework/project/service_wave105_test.go`
- `backend/testframework/project/service_wave113_test.go`
- `backend/testframework/project/service_wave120_test.go`
- `backend/testframework/project/service_wave122_extra_test.go`
- `backend/testframework/project/service_wave122_test.go`
- `backend/testframework/project/service_wave124_test.go`
- `backend/testframework/project/template_export_ent_test.go`
- `backend/testframework/project/template_handler_fa11_test.go`
- `backend/testframework/project/template_handler_routes_test.go`
- `backend/testframework/project/template_handler_wave101_test.go`
- `backend/testframework/project/template_multipart_test.go`
- `backend/testframework/project/templates_test.go`
- `backend/testframework/project/templates_wave105_test.go`

**Итого A5: 38**

## Контроль покрытия W1

| Агент | Файлов |
|-------|-------:|
| A1 | 36 |
| A2 | 33 |
| A3 | 37 |
| A4 | 19 |
| A5 | 38 |
| **Σ** | **163** |
