# PROD-01 — Workflow Automation Lite (v2.6 Spike)

> **Status:** Spike / design ready · **Target milestone:** v2.6  
> **Competitive gap:** #1 vs Jira (user-configurable rules)

---

## Goal (MVP)

Allow project admins to define **1–3 automation rules** without code:

| Trigger | Action (MVP) |
|---------|----------------|
| Issue moved to status X | Post system message to project channel Y |
| Issue assigned to user | Notify assignee (reuse notification pipeline) |
| Issue created with type Bug | Set default priority Critical (optional v2.6.1) |

---

## Existing infrastructure (reuse)

| Layer | Already exists |
|-------|----------------|
| Workflow | `WorkflowStatus`, `WorkflowTransition`, `WorkflowEngine.ValidateTransition` |
| Events | `IssueMovedEvent`, `IssueAssignedEvent` via outbox → NATS |
| Consumer | `notification/consumer.go` handles `issue.moved` → in-app notifications |
| Messenger | `messenger.Service.CreateMessage` for channel posts |
| UI shell | `/projects/:id/settings/workflow` — statuses + transitions CRUD |

---

## Proposed schema (v2.6)

```go
// ent/schema/automationrule.go
type AutomationRule struct {
  ProjectID   uuid.UUID
  Name        string
  Enabled     bool
  TriggerType enum // status_changed | issue_created | issue_assigned
  TriggerConfig JSON // { "to_status_id": "uuid" }
  ActionType  enum // post_to_channel | notify_user | set_field
  ActionConfig JSON // { "channel_id": "uuid" }
  CreatedByID uuid.UUID
}
```

**API:**
- `GET/POST /api/projects/:projectId/automation-rules`
- `PATCH/DELETE .../:ruleId`
- RBAC: `PermProjectManageWorkflow`

**Engine:**
- New `tracker/automation.go` — `RunRules(ctx, trigger, payload)`
- Called from `MoveIssue` / `UpdateIssue` after successful commit (or NATS consumer for async)

---

## Frontend (v2.6)

- Tab **Automation** on `/settings/workflow` (next to Statuses / Transitions)
- Rule builder: dropdown trigger → dropdown action → save
- i18n keys under `tracker:automation.*`
- E2E: create rule → move issue on Kanban → verify channel message

---

## Out of scope (v2.6)

- JQL conditions, cron, webhooks outbound, cross-project rules
- Visual rule graph (Jira Automation UI parity)

---

## Acceptance (PROD-01 done)

1. Admin creates rule: Done → #general channel message
2. User moves issue to Done on board → message appears within 5s (WS)
3. Rule disable stops future triggers
4. Unit tests: engine + handler; E2E smoke one rule

---

*Spike 2026-06-20 — cycle 10 audit prep for v2.6*
