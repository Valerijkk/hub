package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Issue — определение Ent-схемы сущности Issue.
type Issue struct {
	ent.Schema
}

// Поля сущности Issue.
func (Issue) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Int("issue_number").
			Positive().
			Immutable(),
		field.String("issue_key").
			Unique().
			Immutable().
			Comment("Project-prefixed ID: HUB-1"),
		field.String("title").
			NotEmpty().
			MaxLen(500),
		field.JSON("description_json", map[string]interface{}{}).
			Optional().
			Comment("Plate editor JSON"),
		field.Text("description_text").
			Optional().
			Comment("Plain text for search"),
		field.Enum("issue_type").
			Values("bug", "task", "story", "epic", "sub_task", "improvement", "tech_debt", "spike", "release"),
		field.Enum("priority").
			Values("critical", "high", "medium", "low", "trivial").
			Default("medium"),
		field.String("rank").
			Optional().
			Comment("LexoRank for Kanban positioning"),
		field.Int("version").
			Default(1).
			Comment("Optimistic locking"),
		field.Int("story_points").
			Optional().
			Nillable().
			Comment("Story points (Fibonacci: 1,2,3,5,8,13,21)"),
		field.Int("estimate_minutes").
			Optional().
			Nillable(),
		field.Int("time_spent_minutes").
			Default(0),
		field.JSON("checklist", []map[string]interface{}{}).
			Optional().
			Comment("JSONB: [{id, text, checked}]"),
		field.JSON("custom_fields", map[string]interface{}{}).
			Optional().
			Comment("JSONB: custom_field_def_id -> value"),
		field.Time("due_date").
			Optional().
			Nillable(),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Время мягкого удаления (корзина 30 дней)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Рёбра (edges) сущности Issue.
func (Issue) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("issues").
			Unique().
			Required(),
		edge.From("status", WorkflowStatus.Type).
			Ref("issues").
			Unique().
			Required(),
		edge.To("assignee", User.Type).
			Unique(),
		edge.To("reporter", User.Type).
			Unique(),
		edge.To("children", Issue.Type).
			From("parent").
			Unique(),
		edge.To("comments", Comment.Type),
		edge.To("labels", Label.Type),
		edge.To("time_logs", TimeLog.Type),
		edge.To("activities", ActivityLog.Type),
		edge.To("files", FileMetadata.Type),
		edge.To("outgoing_links", IssueLink.Type),
		edge.From("incoming_links", IssueLink.Type).
			Ref("target"),
		edge.From("linked_test_cases", TestCase.Type).
			Ref("linked_issues"),
		edge.To("github_links", GitHubLink.Type),
		edge.To("gitlab_links", GitLabLink.Type),
		edge.To("message_links", IssueMessageLink.Type),
		edge.From("sprint", Sprint.Type).
			Ref("issues").
			Unique(),
	}
}

// Индексы сущности Issue.
func (Issue) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("issue_key"),
		index.Fields("issue_number"),
		index.Fields("rank"),
		index.Fields("created_at"),
		index.Fields("priority"),
		index.Fields("due_date"),
		index.Fields("deleted_at"),
		// A8-001: ListIssues WHERE project_issues = $1 AND deleted_at IS NULL
		index.Edges("project").Fields("deleted_at"),
	}
}
