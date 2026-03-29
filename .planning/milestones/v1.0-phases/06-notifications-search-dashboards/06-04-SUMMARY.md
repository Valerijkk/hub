---
phase: 06-notifications-search-dashboards
plan: 04
subsystem: audit
tags: [audit-log, jetstream, soft-delete, gdpr, admin, trash, resource-limits]

requires:
  - phase: 01-core-infrastructure
    provides: Ent ORM, NATS JetStream, SystemSetting schema, auth middleware
  - phase: 02-tracker
    provides: Issue schema, tracker repository
  - phase: 04-tms
    provides: TestCase schema, TMS repository
provides:
  - AuditLog Ent schema with JetStream durable consumer
  - Soft delete (deleted_at) on Issue, TestCase, FileMetadata
  - 30-day trash with restore and auto-purge
  - Resource limits via SystemSetting
  - User deactivation, task reassignment, GDPR delete
  - Admin UI for audit log, limits, trash, user management
affects: [07-api-docs, dashboard]

tech-stack:
  added: []
  patterns: [soft-delete-with-deleted-at, jetstream-durable-consumer-audit, admin-role-middleware]

key-files:
  created:
    - backend/ent/schema/auditlog.go
    - backend/internal/audit/handler.go
    - backend/internal/audit/service.go
    - backend/internal/audit/consumer.go
    - backend/internal/audit/repository.go
    - backend/internal/audit/dto.go
    - backend/internal/audit/service_test.go
    - frontend/src/api/admin.ts
    - frontend/src/components/admin/audit-log.tsx
    - frontend/src/components/admin/resource-limits.tsx
    - frontend/src/components/admin/trash-manager.tsx
    - frontend/src/components/admin/user-management.tsx
    - frontend/src/routes/_authenticated/admin/audit.tsx
    - frontend/src/routes/_authenticated/admin/limits.tsx
    - frontend/src/routes/_authenticated/admin/trash.tsx
    - frontend/src/routes/_authenticated/admin/users.tsx
  modified:
    - backend/ent/schema/issue.go
    - backend/ent/schema/testcase.go
    - backend/ent/schema/filemetadata.go
    - backend/ent/schema/user.go
    - backend/internal/auth/middleware.go
    - backend/internal/tracker/repository_impl.go
    - backend/internal/tms/ent_repository.go
    - backend/internal/storage/service.go
    - backend/cmd/server/main.go

key-decisions:
  - "JetStream durable consumer 'audit-logger' on AUDIT stream for at-least-once audit recording"
  - "Soft delete via deleted_at field (same pattern as Project) for Issue, TestCase, FileMetadata"
  - "Auth middleware checks user.Status to reject deactivated users"
  - "DeleteIssue/DeleteTestCase changed from hard delete to soft delete (set deleted_at)"
  - "PurgeExpired runs as goroutine with 24h ticker in main.go"

patterns-established:
  - "Soft delete pattern: deleted_at nullable Time field + all queries filter DeletedAtIsNil()"
  - "Admin-only middleware: check user.SystemRole == admin via handler method"
  - "GDPR delete: clear assignee -> delete sessions -> delete user record"

requirements-completed: [AUDT-01, AUDT-02, AUDT-03, AUDT-04]

duration: 10min
completed: 2026-03-21
---

# Phase 06 Plan 04: Audit, Limits, Trash & User Management Summary

**Full audit trail with JetStream durable consumer, 30-day soft-delete trash for issues/test-cases/files, configurable resource limits, and GDPR-compliant user lifecycle management with admin UI.**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-21T03:58:28Z
- **Completed:** 2026-03-21T04:08:28Z
- **Tasks:** 2
- **Files modified:** 25

## Accomplishments

### Task 1: Audit Backend
- Created AuditLog Ent schema with indexes on action, entity_type, user_id, project_id, created_at
- Implemented JetStream durable consumer ("audit-logger") on AUDIT stream consuming events from all modules (project, tracker, messenger, tms, calls, auth)
- Added `deleted_at` nullable Time field to Issue, TestCase, FileMetadata schemas with indexes
- Updated all existing queries in tracker (FindIssueByKey, FindIssueByID, ListIssues, GetBoardData), TMS (GetTestCase, FindTestCaseByKey, ListTestCases), and storage (GenerateDownloadURL) to filter DeletedAtIsNil()
- Changed DeleteIssue and DeleteTestCase from hard delete to soft delete (SetDeletedAt)
- Added User.status enum field (active/deactivated) with auth middleware check
- Built audit service with resource limits (SystemSetting), trash operations (list/restore/soft-delete/purge), user management (deactivate/reassign/GDPR)
- Admin-only HTTP routes under /api/admin/* with role check middleware
- Wired audit consumer and purge cron in main.go
- Unit tests for subject mapping, CRUD suffix conversion, days remaining calculation, field extraction

### Task 2: Admin Frontend
- API client with typed functions for all admin endpoints
- AuditLogViewer: filterable paginated table with CSV export and 30-second auto-refresh
- ResourceLimits: numeric input form with save and current value display
- TrashManager: tabbed interface (issues/test-cases/files/projects) with restore and days-remaining indicator
- UserManagement: user list with deactivation, reassignment dialog, and multi-step GDPR delete confirmation
- 4 admin route pages under /admin/* with admin role guard in beforeLoad

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed duplicate User edges**
- **Found during:** Task 1 (go generate)
- **Issue:** User schema had duplicate "notifications" and "notification_preferences" edges from parallel notification agent
- **Fix:** Removed duplicate edges (also auto-fixed by linter)
- **Files modified:** backend/ent/schema/user.go

**2. [Rule 1 - Bug] Fixed RegisterRoutes signature**
- **Found during:** Task 1 (go build)
- **Issue:** Handler.RegisterRoutes had invalid Go syntax with inline interface type parameter
- **Fix:** Removed unused RegisterRoutes method, kept RegisterRoutesSimple
- **Files modified:** backend/internal/audit/handler.go

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | 71e0147 | feat(06-04): audit backend |
| 2 | 0fb6118 | feat(06-04): admin frontend |
