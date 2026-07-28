# W12 — реестр P0 (стадия B)

> TASK-008 · 2026-07-28 · Источник: `00-STAGE-A-COMPLETE.md`

| Source ID | Class | Title | Primary | Mutation criterion | Backlog |
|-----------|-------|-------|---------|--------------------|---------|
| W1-A5-001 | AUTHZ-PROJECT | project routes без RBAC | `backend/internal/project/handler.go:23-57` | убрать `RequirePermission` с Delete → stranger 200 | [TASK-009](../../backlog/TASK-009-project-routes-rbac.md) |
| W3-A2-001 | AUTHZ-PROMOTE | promote-to-issue без PermTaskCreate | `backend/internal/messenger/service_promote.go:90-174` | убрать вызов `promoteCheck` → cross-project create 201 | [TASK-010](../../backlog/TASK-010-promote-to-issue-rbac.md) |
| W4-A5-001 | WEBHOOK-MATCH | UI events ≠ NATS subjects | `backend/internal/apigateway/webhook.go:325-331` | вернуть exact-only match → job не enqueued | [TASK-011](../../backlog/TASK-011-webhook-event-match.md) |
| W4-A5-002 | WEBHOOK-SCOPE | dispatch без project-scope | `backend/internal/apigateway/webhook.go:209-234` | убрать фильтр project → hook A получает event B | [TASK-012](../../backlog/TASK-012-webhook-project-scope.md) |

P1+ реестры: `W12-REGISTRY-P1-BACKEND.md`, `W12-REGISTRY-P1-FRONTEND.md` (агенты W12).
