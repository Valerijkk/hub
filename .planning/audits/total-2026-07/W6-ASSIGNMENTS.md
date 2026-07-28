# W6 Assignments — auth + rbac + audit

> Стадия A · read-only · TASK-008

## A1 — auth prod + half tests (30)
Отчёт: `A6-1-auth-core.md`

- `backend/internal/auth/dto.go`
- `backend/internal/auth/events.go`
- `backend/internal/auth/export_hooks.go`
- `backend/internal/auth/handler.go`
- `backend/internal/auth/magiclink.go`
- `backend/internal/auth/middleware.go`
- `backend/internal/auth/repository.go`
- `backend/internal/auth/resend.go`
- `backend/internal/auth/service.go`
- `backend/internal/auth/service_login.go`
- `backend/internal/auth/service_organization.go`
- `backend/internal/auth/service_profile.go`
- `backend/internal/auth/service_registration.go`
- `backend/internal/auth/service_session.go`
- `backend/internal/auth/telegram.go`
- `backend/internal/auth/telegram_bot.go`
- `backend/testframework/auth/handler_admin_extra_test.go`
- `backend/testframework/auth/handler_admin_test.go`
- `backend/testframework/auth/handler_extra_test.go`
- `backend/testframework/auth/handler_integration_test.go`
- `backend/testframework/auth/handler_invite_coverage_test.go`
- `backend/testframework/auth/handler_invite_test.go`
- `backend/testframework/auth/handler_magiclink_runes_test.go`
- `backend/testframework/auth/handler_organization_test.go`
- `backend/testframework/auth/handler_profile_test.go`
- `backend/testframework/auth/handler_test.go`
- `backend/testframework/auth/handler_wave115_test.go`
- `backend/testframework/auth/magiclink_test.go`
- `backend/testframework/auth/middleware_failclosed_test.go`
- `backend/testframework/auth/middleware_test.go`

## A2 — auth remaining tests (13)
Отчёт: `A6-2-auth-tests.md`

- `backend/testframework/auth/repository_organization_test.go`
- `backend/testframework/auth/repository_test.go`
- `backend/testframework/auth/resend_telegram_test.go`
- `backend/testframework/auth/service_login_test.go`
- `backend/testframework/auth/service_organization_test.go`
- `backend/testframework/auth/service_profile_test.go`
- `backend/testframework/auth/service_registration_race_test.go`
- `backend/testframework/auth/service_registration_test.go`
- `backend/testframework/auth/service_session_test.go`
- `backend/testframework/auth/service_test.go`
- `backend/testframework/auth/telegram_bot_coverage_test.go`
- `backend/testframework/auth/telegram_bot_polling_test.go`
- `backend/testframework/auth/telegram_widget_test.go`

## A3 — rbac full (15)
Отчёт: `A6-3-rbac.md`

- `backend/internal/rbac/cache.go`
- `backend/internal/rbac/custom_roles.go`
- `backend/internal/rbac/export_hooks.go`
- `backend/internal/rbac/handler.go`
- `backend/internal/rbac/middleware.go`
- `backend/internal/rbac/permissions.go`
- `backend/internal/rbac/repository.go`
- `backend/internal/rbac/service.go`
- `backend/testframework/rbac/cache_test.go`
- `backend/testframework/rbac/custom_roles_a1002_test.go`
- `backend/testframework/rbac/handler_test.go`
- `backend/testframework/rbac/middleware_integration_test.go`
- `backend/testframework/rbac/middleware_test.go`
- `backend/testframework/rbac/permissions_test.go`
- `backend/testframework/rbac/service_test.go`

## A4 — audit prod + half tests (22)
Отчёт: `A6-4-audit-core.md`

- `backend/internal/audit/consumer.go`
- `backend/internal/audit/dto.go`
- `backend/internal/audit/export_hooks.go`
- `backend/internal/audit/handler.go`
- `backend/internal/audit/handler_alerts.go`
- `backend/internal/audit/middleware.go`
- `backend/internal/audit/repository.go`
- `backend/internal/audit/retention.go`
- `backend/internal/audit/service.go`
- `backend/testframework/audit/consumer_ack_a1008_test.go`
- `backend/testframework/audit/consumer_lifecycle_test.go`
- `backend/testframework/audit/consumer_misc_test.go`
- `backend/testframework/audit/consumer_test.go`
- `backend/testframework/audit/dto_mapping_test.go`
- `backend/testframework/audit/founder_ban_a1009_test.go`
- `backend/testframework/audit/gdpr_founder_a1004_test.go`
- `backend/testframework/audit/handler_alerts_test.go`
- `backend/testframework/audit/handler_founder_ban_a1009_test.go`
- `backend/testframework/audit/handler_limits_a1013_test.go`
- `backend/testframework/audit/handler_test.go`
- `backend/testframework/audit/handler_user_mgmt_test.go`
- `backend/testframework/audit/middleware_test.go`

## A5 — audit remaining tests (12)
Отчёт: `A6-5-audit-tests.md`

- `backend/testframework/audit/repository_extra_test.go`
- `backend/testframework/audit/repository_newspost_trash_test.go`
- `backend/testframework/audit/repository_scheduled_call_trash_test.go`
- `backend/testframework/audit/repository_test.go`
- `backend/testframework/audit/repository_trash_test.go`
- `backend/testframework/audit/retention_test.go`
- `backend/testframework/audit/service_deactivate_test.go`
- `backend/testframework/audit/service_limits_test.go`
- `backend/testframework/audit/service_purge_test.go`
- `backend/testframework/audit/service_security_test.go`
- `backend/testframework/audit/service_test.go`
- `backend/testframework/audit/service_trash_test.go`

**Σ файлов: 92**

