# Phase 2: Task Tracker - Research

**Researched:** 2026-03-21
**Domain:** Issue CRUD, Kanban boards, workflows, comments, activity log, time tracking (Go + Fiber v3 + Ent + React 19 + pragmatic-drag-and-drop + Plate)
**Confidence:** HIGH

## Summary

Phase 2 builds the task tracker module on top of the Phase 1 foundation (auth, RBAC, projects, WebSocket gateway, outbox pattern, file storage). The tracker is the highest-value module in Hub and the foundation for cross-module linking with TMS (Phase 4) and messenger (Phase 3). It covers issue CRUD with 9 templates, custom fields via JSONB, unlimited parent/child nesting via Ent self-referencing edges, issue linking (7 link types + custom), customizable workflows (project-scoped state machine), comments with @mentions and reactions, full change log, time tracking with start/stop timer, project-level labels, file attachments (reusing Phase 1 storage), and Kanban boards with drag-and-drop positioning via LexoRank.

The existing codebase provides strong patterns to follow: handler/service/repository layered architecture (see `project` module), Ent schemas with UUID primary keys and created_at/updated_at timestamps, outbox event publishing for cross-module communication, RBAC permission bitmap checks (`PermTaskCreate`, `PermTaskEdit`, etc. already defined in Phase 1), and the API client with auto-refresh on the frontend. The tracker module (`backend/internal/tracker/`) will mirror the `project` module structure. The frontend will add routes under `_authenticated/projects/$projectId/` for issue views and Kanban boards.

**Primary recommendation:** Follow the existing `project` module pattern exactly (handler.go, service.go, repository.go, dto.go, events.go) with Repository interface for testability. Use Ent self-referencing O2M edges for parent/child hierarchy, a `workflow_status` schema with `allowed_transitions` JSONB for the state machine, `lexorank` Go library for card positioning, and Plate editor for rich-text issue descriptions/comments stored as JSON with a plain_text column for search.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| TASK-01 | 9 issue templates: Bug, Task, Story, Epic, Sub-task, Improvement, Tech Debt, Spike, Release | Ent `IssueType` enum schema; seeded per-project on creation |
| TASK-02 | Fixed + custom fields on issues | Fixed fields as Ent columns + `custom_field_definitions` table + JSONB `custom_field_values` on issue |
| TASK-03 | Unlimited parent/child nesting | Ent self-referencing O2M edge (`parent`/`children` on Issue schema) |
| TASK-04 | Issue links: parent/child, blocks, relates_to, duplicates, clones, causes + custom | `IssueLink` schema with source_id, target_id, link_type enum + custom type string |
| TASK-05 | Project-prefixed sequential IDs (HUB-1, MOBILE-1) | Atomic `task_counter` increment on Project (already exists in schema) + `issue_number` field |
| TASK-06 | Customizable workflow (statuses + transitions) | `WorkflowStatus` + `WorkflowTransition` schemas per project; state machine validation in service layer |
| TASK-07 | Comments with @mentions, internal comments, quoting, reactions | `Comment` schema with Plate JSON body, `is_internal` flag, `mentions` JSONB, `Reaction` schema |
| TASK-08 | Full change log (who changed what when) | `ActivityLog` schema capturing field-level diffs as JSONB, published via outbox |
| TASK-09 | Time tracking: estimate + actual + start/stop timer | `TimeLog` schema with start/end timestamps + `estimate` and `time_spent` fields on Issue |
| TASK-10 | Labels: project-level, colors, groups, shared across modules | `Label` + `LabelGroup` schemas at project level; M2M edge to Issue |
| TASK-11 | File attachments with preview | Reuse Phase 1 `FileMetadata` schema + presigned URLs; add edge from Issue to FileMetadata |
| KANB-01 | Kanban columns = workflow statuses | Board view reads workflow statuses as columns; cards grouped by status |
| KANB-04 | Multiple boards per project | `Board` schema with saved filter configuration (JSONB); references project workflow |
| KANB-05 | Board filters: my tasks, bugs only, by priority | Filter as query params on board endpoint; persisted in Board.filters JSONB |
| UI-09 | Pagination for issue lists, virtualization for large lists | Cursor-based pagination on backend; TanStack Virtual on frontend for large lists |
</phase_requirements>

## Standard Stack

### Core (New for Phase 2)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/cristalhq/lexorank | latest | LexoRank card positioning | Clean Go implementation; generates lexicographically sortable strings between two existing keys; avoids full reindex on drag |
| @atlaskit/pragmatic-drag-and-drop | 1.7.x | Kanban drag-and-drop | From Atlassian (powers Jira/Trello), ~4.7kB core, headless (no UI lock-in), framework-agnostic core with React adapter |
| @udecode/plate | 49.x | Rich-text editor (descriptions, comments) | Shadcn-native, built on Slate.js, free @mentions + tables, Plate JSON for storage |
| motion | 12.x | Drag animations for Kanban cards | Layout animations for card reorder, smooth transitions |
| @tanstack/react-virtual | 3.x | Virtualization for large issue lists | Virtual scrolling for 1000+ items without DOM bloat |

### Already Available (from Phase 1)

| Library | Purpose | Notes |
|---------|---------|-------|
| Ent v0.14.5 | ORM, schema, migrations | Self-referencing edges for hierarchy, JSONB fields for custom data |
| Fiber v3.1.0 | HTTP handlers | Same handler pattern as project module |
| NATS 1.49.0 | Event publishing | Outbox pattern already implemented |
| TanStack Query 5.x | Server state | Optimistic updates for drag-and-drop |
| TanStack Router 1.x | Routing | New routes under `_authenticated/projects/$projectId/` |
| Zustand 5.x | Client state | Board filter state, active timer state |
| aws-sdk-go-v2 | File storage | Presigned URLs for attachments |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| cristalhq/lexorank | Southclaws/lexorank | Southclaws has ReorderableList utility but is heavier; cristalhq is simpler for our single-column reorder use case |
| JSONB custom fields | EAV pattern (separate rows) | EAV is normalized but creates N+1 queries and joins; JSONB is simpler, faster reads, good enough for 50 users |
| Ent self-ref edges for hierarchy | Materialized path / nested sets | Self-ref edges are simplest with Ent, support unlimited depth; materialized path better for deep reads but complex writes |
| Plate | Tiptap | Tiptap tables/mentions are paid extensions; Plate is fully free and shadcn-native |

**Installation (new packages):**
```bash
# Backend
go get github.com/cristalhq/lexorank

# Frontend
npm install @atlaskit/pragmatic-drag-and-drop @udecode/plate motion @tanstack/react-virtual
```

## Architecture Patterns

### Recommended Module Structure

```
backend/internal/tracker/
  handler.go            # Fiber route handlers (issues, comments, boards, labels, workflow)
  service.go            # Business logic (issue CRUD, workflow validation, time tracking)
  repository.go         # Repository interface + EntRepository implementation
  workflow.go           # State machine: validate transitions, enforce rules
  dto.go                # Request/response types
  events.go             # NATS event definitions (tracker.issue.created, etc.)

backend/ent/schema/
  issue.go              # Core issue entity with all fixed fields
  issue_type.go         # Enum-like: Bug, Task, Story, Epic, Sub-task, etc.
  issue_link.go         # M2M self-referencing: source -> target with type
  workflow_status.go    # Project-scoped statuses (category: todo/in_progress/done)
  workflow_transition.go # Allowed transitions between statuses
  comment.go            # Issue comments with rich text
  reaction.go           # Emoji reactions on comments
  activity_log.go       # Field-level change tracking
  time_log.go           # Time tracking entries
  label.go              # Project-level labels
  label_group.go        # Label groups
  board.go              # Kanban board configuration
  custom_field_def.go   # Custom field definitions per project
  custom_field_val.go   # Custom field values (JSONB) per issue

frontend/src/
  routes/_authenticated/projects/$projectId/
    issues/
      index.tsx         # Issue list with pagination
      $issueKey.tsx     # Issue detail view
      new.tsx           # Create issue form
    board/
      index.tsx         # Default Kanban board
      $boardId.tsx      # Specific board view
    settings/
      workflow.tsx      # Workflow editor
      labels.tsx        # Label management
      custom-fields.tsx # Custom field management
  api/
    tracker.ts          # API client for tracker endpoints
  stores/
    board-store.ts      # Active board filters, drag state
    timer-store.ts      # Active time tracking timer
  components/tracker/
    issue-card.tsx      # Kanban card component
    issue-form.tsx      # Create/edit issue form
    board-column.tsx    # Kanban column
    comment-list.tsx    # Comment thread
    activity-feed.tsx   # Activity log
    plate-editor.tsx    # Plate rich-text editor wrapper
```

### Pattern 1: Ent Self-Referencing Edge for Issue Hierarchy (TASK-03)

**What:** Unlimited parent/child nesting using Ent's O2M self-referencing edge pattern.
**When to use:** Any entity that needs tree structure -- issues, test case folders, comment threads.

```go
// ent/schema/issue.go
func (Issue) Edges() []ent.Edge {
    return []ent.Edge{
        // Self-referencing: parent/children hierarchy
        edge.To("children", Issue.Type).
            From("parent").
            Unique(),
        // Other edges...
        edge.From("project", Project.Type).Ref("issues").Unique().Required(),
        edge.To("assignee", User.Type).Unique(),
        edge.To("reporter", User.Type).Unique(),
        edge.To("comments", Comment.Type),
        edge.To("labels", Label.Type),
        edge.To("time_logs", TimeLog.Type),
        edge.To("activities", ActivityLog.Type),
        edge.To("files", FileMetadata.Type),
    }
}
```

**Querying children recursively:**
```go
// For display, load 2 levels deep (direct children + their children)
issue, err := client.Issue.Query().
    Where(issue.IDEQ(issueID)).
    WithChildren(func(q *ent.IssueQuery) {
        q.WithChildren() // 2nd level
    }).
    WithParent().
    Only(ctx)
```

### Pattern 2: Project-Prefixed Issue IDs (TASK-05)

**What:** Atomic counter on Project entity generates sequential numbers. Issue stores both `issue_number` (int) and computed `issue_key` (string).
**When to use:** Creating new issues.

```go
// repository.go -- inside a transaction
func (r *EntRepository) CreateIssue(ctx context.Context, projectID uuid.UUID, input CreateIssueInput) (*ent.Issue, error) {
    tx, err := r.client.Tx(ctx)
    if err != nil {
        return nil, err
    }

    // Atomically increment project counter
    p, err := tx.Project.UpdateOneID(projectID).
        AddTaskCounter(1).
        Save(ctx)
    if err != nil {
        _ = tx.Rollback()
        return nil, err
    }

    issueNumber := p.TaskCounter
    issueKey := fmt.Sprintf("%s-%d", p.Key, issueNumber)

    issue, err := tx.Issue.Create().
        SetProjectID(projectID).
        SetIssueNumber(issueNumber).
        SetIssueKey(issueKey).
        SetTitle(input.Title).
        // ... other fields
        Save(ctx)
    if err != nil {
        _ = tx.Rollback()
        return nil, err
    }

    // Outbox event
    if r.publisher != nil {
        _ = r.publisher.PublishInTx(ctx, tx, "tracker.issue.created", IssueCreatedEvent{
            IssueID:   issue.ID.String(),
            IssueKey:  issueKey,
            ProjectID: projectID.String(),
        })
    }

    return issue, tx.Commit()
}
```

### Pattern 3: Workflow State Machine (TASK-06)

**What:** Per-project customizable workflow stored as statuses + allowed transitions. Validation at service layer.
**When to use:** Every status change on an issue.

```go
// workflow.go
type WorkflowEngine struct {
    repo Repository
}

// ValidateTransition checks if the status change is allowed.
func (w *WorkflowEngine) ValidateTransition(ctx context.Context, projectID uuid.UUID, fromStatusID, toStatusID uuid.UUID) error {
    transition, err := w.repo.FindTransition(ctx, projectID, fromStatusID, toStatusID)
    if err != nil {
        return apperrors.NewBadRequest("transition not allowed from current status")
    }
    _ = transition
    return nil
}
```

**Default workflow seeded per project:**
```
To Do (category: todo) -> In Progress (category: in_progress) -> In Review (category: in_progress) -> Done (category: done)
```

Status categories (`todo`, `in_progress`, `done`) enable Kanban column grouping and reporting without knowing specific status names.

### Pattern 4: LexoRank for Kanban Card Positioning (Kanban)

**What:** Cards store a `rank` string field. When a card is dragged between two others, generate a rank between them.
**When to use:** Every drag-and-drop reorder on the Kanban board.

```go
import "github.com/cristalhq/lexorank"

// When moving card between cardA (rank "0|aaaaaa:") and cardB (rank "0|bbbbbb:")
newRank, err := lexorank.Between(cardA.Rank, cardB.Rank)
// newRank = "0|aaaaab:" -- sorts between A and B lexicographically

// When moving to first position (before all cards)
newRank, err = lexorank.Before(firstCard.Rank)

// When moving to last position (after all cards)
newRank, err = lexorank.After(lastCard.Rank)
```

**Frontend optimistic update:**
```typescript
// Use TanStack Query optimistic update
const moveCard = useMutation({
    mutationFn: (data: { issueId: string; statusId: string; rank: string }) =>
        api.put(`/projects/${projectId}/issues/${data.issueId}/position`, data),
    onMutate: async (data) => {
        // Cancel outgoing refetches
        await queryClient.cancelQueries({ queryKey: ['board', projectId] });
        // Snapshot
        const previous = queryClient.getQueryData(['board', projectId]);
        // Optimistically update
        queryClient.setQueryData(['board', projectId], (old) => {
            // Move card in local state
            return reorderBoard(old, data);
        });
        return { previous };
    },
    onError: (err, data, context) => {
        // Rollback on error
        queryClient.setQueryData(['board', projectId], context?.previous);
    },
    onSettled: () => {
        queryClient.invalidateQueries({ queryKey: ['board', projectId] });
    },
});
```

### Pattern 5: Custom Fields via JSONB (TASK-02)

**What:** Two schemas: `CustomFieldDefinition` (metadata) and JSONB column on Issue for values.
**When to use:** Project-level custom field configuration + per-issue values.

```go
// ent/schema/custom_field_def.go
func (CustomFieldDefinition) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.String("name").NotEmpty(),
        field.Enum("field_type").Values(
            "text", "number", "date", "checkbox",
            "select", "multi_select", "user", "url",
        ),
        field.JSON("options", []string{}).Optional(), // for select/multi_select
        field.Bool("required").Default(false),
        field.Int("sort_order").Default(0),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

// On Issue schema:
field.JSON("custom_fields", map[string]interface{}{}).
    Optional().
    Comment("JSONB map of custom_field_def_id -> value"),
```

### Pattern 6: Activity Log with Field-Level Diffs (TASK-08)

**What:** Every field change is recorded as a JSONB diff in the `ActivityLog` table.
**When to use:** Any issue update -- status, assignee, title, description, custom fields.

```go
// ent/schema/activity_log.go
func (ActivityLog) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.String("action").NotEmpty(), // "created", "updated", "commented", "status_changed"
        field.JSON("changes", []FieldChange{}).Optional(),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

// types
type FieldChange struct {
    Field    string `json:"field"`
    OldValue any    `json:"old_value,omitempty"`
    NewValue any    `json:"new_value"`
}
```

### Anti-Patterns to Avoid

- **Querying full tree on every request:** Never recursively load the entire issue hierarchy. Load 1-2 levels and lazy-load deeper levels on demand.
- **Storing rich text as HTML:** Store Plate JSON (structured) + plain text column (for search). Never store raw HTML -- it invites XSS and makes diffing impossible.
- **Reindexing all card positions on drag:** Use LexoRank to compute new rank between neighbors. Only update the dragged card's rank. Full rebalance only when rank space exhausted (rare).
- **Workflow validation only on frontend:** Always validate transitions on the backend. Frontend shows available transitions but backend enforces them.
- **Eager loading all edges on list endpoints:** Issue list should NOT load comments, activity logs, or children. Only load assignee, labels, status (WithAssignee, WithLabels, WithStatus).
- **Synchronous activity logging:** Log changes via outbox event, not synchronously in the update transaction. Keeps writes fast.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Card positioning | Custom float-based sorting | LexoRank (cristalhq/lexorank) | Float precision issues after many inserts; LexoRank handles millions of reorders |
| Rich-text editor | Custom contenteditable | Plate editor | Mentions, tables, slash commands, undo/redo, serialization all handled |
| Drag-and-drop | Custom mousedown/mousemove handlers | pragmatic-drag-and-drop | Accessibility, touch support, auto-scroll, drop indicators -- months of work |
| State machine | if/else chains for status transitions | Table-driven workflow (DB-stored transitions) | Hardcoded transitions can't be customized per project; DB-driven approach matches TASK-06 |
| Optimistic UI | Manual cache manipulation | TanStack Query onMutate pattern | Rollback, deduplication, background refetch all handled |
| UUID generation | Custom ID schemes | google/uuid (already used) | Consistent with Phase 1 pattern |
| Sequential numbering | Application-level counter | Atomic DB increment in transaction | Race conditions with concurrent issue creation |
| Pagination | Offset-based | Cursor-based (issue_key or created_at) | Offset breaks when items inserted/deleted during pagination |

**Key insight:** The tracker is deceptively complex -- 15+ related Ent schemas, a state machine, rich text, drag-and-drop, real-time updates, and optimistic UI. Using battle-tested libraries for each sub-problem is the only way to ship in reasonable time.

## Common Pitfalls

### Pitfall 1: Ent N+1 Queries on Issue Lists

**What goes wrong:** Listing 50 issues makes 50+1 queries (one per issue for assignee, labels, status).
**Why it happens:** Default Ent queries don't eager-load edges.
**How to avoid:** Always use `With*()` methods for list queries:
```go
issues, err := client.Issue.Query().
    Where(issue.HasProjectWith(project.IDEQ(projectID))).
    WithAssignee().
    WithLabels().
    WithStatus().
    WithParent().  // only parent ref, not children tree
    Limit(limit).
    Offset(offset).
    All(ctx)
```
**Warning signs:** Issue list page takes >500ms with 50+ issues.

### Pitfall 2: LexoRank Space Exhaustion

**What goes wrong:** After thousands of reorders between the same two cards, rank strings get very long or no space remains.
**Why it happens:** Each "between" operation extends the string. Extreme case: card dragged back and forth 1000 times.
**How to avoid:** Monitor rank string length. When any rank exceeds 50 characters, trigger async background rebalance that reassigns evenly-spaced ranks to all cards in the column. This is a background job, not blocking the user.
**Warning signs:** Rank strings exceeding 30+ characters in the database.

### Pitfall 3: Concurrent Issue Creation Race Condition

**What goes wrong:** Two users create issues simultaneously, both get the same issue number.
**Why it happens:** Read-then-increment pattern without transaction isolation.
**How to avoid:** Use Ent transaction with `AddTaskCounter(1)` -- this uses PostgreSQL atomic UPDATE ... RETURNING which guarantees unique counters under concurrency. Already demonstrated in the project schema which has `task_counter` field.

### Pitfall 4: Optimistic Locking Not Enforced

**What goes wrong:** User A and User B edit the same issue. User B overwrites User A's changes silently.
**Why it happens:** No version check on update.
**How to avoid:** Add `version` field to Issue schema. On every update, check `WHERE version = expectedVersion` and increment. Return 409 Conflict if version mismatch.
```go
field.Int("version").Default(1),

// In update:
updated, err := tx.Issue.UpdateOneID(issueID).
    Where(issue.VersionEQ(expectedVersion)).
    AddVersion(1).
    SetTitle(newTitle).
    Save(ctx)
if ent.IsNotFound(err) {
    return apperrors.NewConflict("issue was modified by another user")
}
```

### Pitfall 5: WebSocket Broadcast Storm on Board

**What goes wrong:** 20 users viewing the same Kanban board, one drags a card. All 20 get WebSocket update, causing 20 TanStack Query invalidations and 20 refetch requests to the API.
**Why it happens:** Naive "invalidate board query on any board event" pattern.
**How to avoid:** Send granular WebSocket events (`card_moved`, `card_updated`) with the diff payload. Frontend applies the diff to the cached query data directly without refetching. Only refetch on reconnect or explicit refresh.
```typescript
// ws message handler
ws.on('tracker.issue.moved', (data) => {
    queryClient.setQueryData(['board', projectId], (old) =>
        applyCardMove(old, data.issueId, data.newStatusId, data.newRank)
    );
});
```

### Pitfall 6: Rich Text Storage Without Plain Text

**What goes wrong:** Full-text search over issue descriptions becomes impossible or requires deserializing Plate JSON for every row.
**Why it happens:** Only storing the Plate JSON blob without a derived plain text column.
**How to avoid:** Store both columns: `description_json` (Plate JSON for rendering) and `description_text` (extracted plain text for search + notifications). Update both on every save. Use PG `tsvector` index on `description_text` for full-text search.

### Pitfall 7: Comment @Mentions Not Validated

**What goes wrong:** User mentions @deleted-user or @non-member in a comment. Notification fails or sends to wrong person.
**Why it happens:** Frontend sends mention user IDs without backend validation.
**How to avoid:** Backend parses Plate JSON, extracts mention node user IDs, validates each is a current project member. Strip invalid mentions before saving. Only send notifications for validated mentions.

## Code Examples

### Issue Ent Schema (Core)

```go
// ent/schema/issue.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
    "time"
)

type Issue struct {
    ent.Schema
}

func (Issue) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.Int("issue_number").Positive().Immutable(),
        field.String("issue_key").Unique().Immutable().
            Comment("Project-prefixed ID: HUB-1, MOBILE-42"),
        field.String("title").NotEmpty().MaxLen(500),
        field.JSON("description_json", map[string]interface{}{}).
            Optional().
            Comment("Plate editor JSON"),
        field.Text("description_text").Optional().
            Comment("Plain text extracted from Plate JSON for search"),
        field.Enum("issue_type").Values(
            "bug", "task", "story", "epic", "sub_task",
            "improvement", "tech_debt", "spike", "release",
        ),
        field.Enum("priority").Values("critical", "high", "medium", "low", "trivial").
            Default("medium"),
        field.String("rank").Optional().
            Comment("LexoRank string for Kanban card positioning"),
        field.Int("version").Default(1).
            Comment("Optimistic locking version"),
        field.Int("estimate_minutes").Optional().Nillable(),
        field.Int("time_spent_minutes").Default(0),
        field.JSON("custom_fields", map[string]interface{}{}).
            Optional().
            Comment("JSONB: custom_field_def_id -> value"),
        field.Time("due_date").Optional().Nillable(),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (Issue) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("project", Project.Type).Ref("issues").Unique().Required(),
        edge.From("status", WorkflowStatus.Type).Ref("issues").Unique().Required(),
        edge.To("assignee", User.Type).Unique(),
        edge.To("reporter", User.Type).Unique(),
        edge.To("children", Issue.Type).From("parent").Unique(),
        edge.To("comments", Comment.Type),
        edge.To("labels", Label.Type),
        edge.To("time_logs", TimeLog.Type),
        edge.To("activities", ActivityLog.Type),
        edge.To("files", FileMetadata.Type),
        edge.To("outgoing_links", IssueLink.Type),
        edge.From("incoming_links", IssueLink.Type).Ref("target"),
    }
}

func (Issue) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("issue_key"),
        index.Fields("issue_number"),
        index.Fields("rank"),
        index.Fields("created_at"),
        index.Fields("priority"),
    }
}
```

### Tracker Route Registration

```go
// handler.go (RegisterRoutes pattern matching project module)
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
    tracker := api.Group("/projects/:projectId", authMw)

    // Issues
    issues := tracker.Group("/issues")
    issues.Post("/", rbac.RequirePermission(rbac.PermTaskCreate, rbacSvc), h.CreateIssue)
    issues.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListIssues)
    issues.Get("/:issueKey", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetIssue)
    issues.Put("/:issueKey", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.UpdateIssue)
    issues.Delete("/:issueKey", rbac.RequirePermission(rbac.PermTaskDelete, rbacSvc), h.DeleteIssue)
    issues.Put("/:issueKey/position", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.MoveIssue)

    // Comments
    issues.Get("/:issueKey/comments", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListComments)
    issues.Post("/:issueKey/comments", rbac.RequirePermission(rbac.PermTaskComment, rbacSvc), h.CreateComment)

    // Activity
    issues.Get("/:issueKey/activity", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetActivity)

    // Time tracking
    issues.Post("/:issueKey/time", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.LogTime)
    issues.Post("/:issueKey/timer/start", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.StartTimer)
    issues.Post("/:issueKey/timer/stop", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.StopTimer)

    // Boards
    boards := tracker.Group("/boards")
    boards.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListBoards)
    boards.Post("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.CreateBoard)
    boards.Get("/:boardId", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetBoard)

    // Workflow
    workflow := tracker.Group("/workflow")
    workflow.Get("/statuses", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListStatuses)
    workflow.Post("/statuses", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.CreateStatus)
    workflow.Get("/transitions", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTransitions)
    workflow.Post("/transitions", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.CreateTransition)

    // Labels
    labels := tracker.Group("/labels")
    labels.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListLabels)
    labels.Post("/", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.CreateLabel)

    // Custom Fields
    customFields := tracker.Group("/custom-fields")
    customFields.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListCustomFields)
    customFields.Post("/", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.CreateCustomField)
}
```

### Cursor-Based Pagination (UI-09)

```go
// dto.go
type ListIssuesRequest struct {
    Cursor   string   `query:"cursor"`      // issue_key of last seen item
    Limit    int      `query:"limit"`        // default 25, max 100
    Status   []string `query:"status"`
    Priority []string `query:"priority"`
    Assignee string   `query:"assignee"`
    Type     string   `query:"type"`
    Search   string   `query:"search"`
    OrderBy  string   `query:"order_by"`     // created_at, priority, updated_at
}

type PaginatedResponse[T any] struct {
    Items      []T    `json:"items"`
    NextCursor string `json:"next_cursor,omitempty"`
    HasMore    bool   `json:"has_more"`
    Total      int    `json:"total"`
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| dnd-kit for drag-and-drop | pragmatic-drag-and-drop | 2024 (dnd-kit abandoned) | Must use pragmatic-dnd; dnd-kit has no maintainer |
| Float/integer position | LexoRank string positioning | Standard since Jira adopted | Eliminates full-column reindex on card move |
| Tiptap for rich text | Plate 49 (shadcn-native) | 2025-2026 | Free tables/mentions vs Tiptap paid extensions |
| HTML rich text storage | Structured JSON (Plate/Slate) | Industry standard | Better diffing, search, rendering flexibility |
| Offset pagination | Cursor-based pagination | Best practice since 2020 | Stable under concurrent inserts/deletes |

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing + testify v1.10 |
| Config file | None needed (Go stdlib) |
| Quick run command | `go test ./backend/internal/tracker/... -count=1 -race` |
| Full suite command | `go test ./backend/... -count=1 -race` |

### Phase Requirements to Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TASK-01 | Issue creation with 9 types | unit | `go test ./backend/internal/tracker/ -run TestCreateIssue -count=1` | No - Wave 0 |
| TASK-02 | Custom field values stored/retrieved | unit | `go test ./backend/internal/tracker/ -run TestCustomFields -count=1` | No - Wave 0 |
| TASK-03 | Parent/child nesting (3+ levels) | unit | `go test ./backend/internal/tracker/ -run TestHierarchy -count=1` | No - Wave 0 |
| TASK-04 | Issue link creation + types | unit | `go test ./backend/internal/tracker/ -run TestIssueLinks -count=1` | No - Wave 0 |
| TASK-05 | Sequential IDs under concurrency | unit | `go test ./backend/internal/tracker/ -run TestSequentialIDs -count=1 -race` | No - Wave 0 |
| TASK-06 | Workflow transition validation | unit | `go test ./backend/internal/tracker/ -run TestWorkflow -count=1` | No - Wave 0 |
| TASK-07 | Comment CRUD + internal flag | unit | `go test ./backend/internal/tracker/ -run TestComments -count=1` | No - Wave 0 |
| TASK-08 | Activity log records changes | unit | `go test ./backend/internal/tracker/ -run TestActivityLog -count=1` | No - Wave 0 |
| TASK-09 | Time tracking start/stop | unit | `go test ./backend/internal/tracker/ -run TestTimeTracking -count=1` | No - Wave 0 |
| TASK-10 | Label CRUD + assignment | unit | `go test ./backend/internal/tracker/ -run TestLabels -count=1` | No - Wave 0 |
| TASK-11 | File attachment to issue | unit | `go test ./backend/internal/tracker/ -run TestAttachments -count=1` | No - Wave 0 |
| KANB-01 | Board columns match workflow statuses | unit | `go test ./backend/internal/tracker/ -run TestBoardColumns -count=1` | No - Wave 0 |
| KANB-04 | Multiple boards per project | unit | `go test ./backend/internal/tracker/ -run TestMultipleBoards -count=1` | No - Wave 0 |
| KANB-05 | Board filters | unit | `go test ./backend/internal/tracker/ -run TestBoardFilters -count=1` | No - Wave 0 |
| UI-09 | Cursor-based pagination | unit | `go test ./backend/internal/tracker/ -run TestPagination -count=1` | No - Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./backend/internal/tracker/... -count=1 -race`
- **Per wave merge:** `go test ./backend/... -count=1 -race`
- **Phase gate:** Full suite green before verify-work

### Wave 0 Gaps

- [ ] `backend/internal/tracker/service_test.go` -- covers TASK-01 through TASK-11, KANB-01, KANB-04, KANB-05
- [ ] `backend/internal/tracker/workflow_test.go` -- covers TASK-06 workflow transition validation
- [ ] Mock repository implementation following `project/service_test.go` pattern
- [ ] Framework install: `go get github.com/cristalhq/lexorank` -- LexoRank dependency

## Open Questions

1. **LexoRank rebalancing strategy**
   - What we know: LexoRank strings grow when inserting between adjacent items repeatedly
   - What's unclear: Exact threshold for triggering background rebalance
   - Recommendation: Set 50-char threshold; rebalance assigns evenly-spaced ranks to all cards in column

2. **Plate editor version pinning**
   - What we know: Plate is at v49.x, actively developed with frequent releases
   - What's unclear: Breaking changes between minor versions
   - Recommendation: Pin to specific minor version in package.json, test before upgrading

3. **Activity log granularity for description changes**
   - What we know: Recording field-level diffs is required
   - What's unclear: How to diff Plate JSON (deeply nested objects)
   - Recommendation: For description, store "description changed" without inline diff; link to version history if needed

## Sources

### Primary (HIGH confidence)
- Ent O2M recursive example: https://pkg.go.dev/entgo.io/ent/examples/o2mrecur
- Ent schema edges docs: https://entgo.io/docs/schema-edges/
- TanStack Query optimistic updates: https://tanstack.com/query/v5/docs/react/guides/optimistic-updates
- pragmatic-drag-and-drop: https://atlassian.design/components/pragmatic-drag-and-drop/
- Plate editor docs: https://platejs.org/docs
- PostgreSQL JSONB docs: https://www.postgresql.org/docs/current/datatype-json.html

### Secondary (MEDIUM confidence)
- cristalhq/lexorank Go library: https://github.com/cristalhq/lexorank
- LexoRank algorithm explanation: https://medium.com/whisperarts/lexorank-what-are-they-and-how-to-use-them-for-efficient-list-sorting-a48fc4e7849f
- Kanban board with shadcn + pragmatic-dnd: https://marmelab.com/blog/2026/01/15/building-a-kanban-board-with-shadcn.html
- react-kanban-kit (pragmatic-dnd reference): https://github.com/braiekhazem/react-kanban-kit

### Tertiary (LOW confidence)
- Southclaws/lexorank (alternative): https://github.com/Southclaws/lexorank -- claimed production use but fewer stars
- Plate v49 specific features: training data only, verify against official docs before implementation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- all libraries verified via npm/pkg.go.dev, consistent with Phase 1 patterns
- Architecture: HIGH -- mirrors existing project module pattern exactly, Ent self-ref edges documented officially
- Pitfalls: HIGH -- N+1 queries, LexoRank exhaustion, concurrent counters are well-documented problems with known solutions
- Workflow engine: MEDIUM -- custom table-driven approach is standard but requires careful schema design
- Plate integration: MEDIUM -- version 49 is actively moving; pin version and test

**Research date:** 2026-03-21
**Valid until:** 2026-04-21 (stable domain, well-known patterns)
