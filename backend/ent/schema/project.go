package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Project — определение Ent-схемы сущности Project.
type Project struct {
	ent.Schema
}

// Поля сущности Project.
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MinLen(2).
			MaxLen(100),
		field.String("key").
			Unique().
			NotEmpty().
			Match(keyRegexp),
		field.String("description").
			Optional().
			MaxLen(2000),
		field.String("avatar_url").
			Optional(),
		field.Enum("status").
			Values("active", "archived", "deleted").
			Default("active"),
		field.Int("task_counter").
			Default(0),
		field.Int("test_case_counter").
			Default(0).
			Comment("Счётчик тест-кейсов для генерации ключей"),
		field.Time("deleted_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (Project) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Рёбра (edges) сущности Project.
// OnDelete Cascade на всех owned children: soft-delete не вызывает FK OnDelete;
// PermanentDelete / hard-purge — намеренный wipe (ADR-ENT-ONDELETE-CASCADE-2026 / A2W8-4-002).
func (Project) Edges() []ent.Edge {
	cascade := entsql.OnDelete(entsql.Cascade)
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("projects").
			Unique(),
		edge.To("members", ProjectMember.Type).Annotations(cascade),
		edge.To("roles", Role.Type).Annotations(cascade),
		edge.To("teams", Team.Type).Annotations(cascade),
		edge.To("invites", Invite.Type).Annotations(cascade),
		edge.To("files", FileMetadata.Type).Annotations(cascade),
		edge.To("issues", Issue.Type).Annotations(cascade),
		edge.To("workflow_statuses", WorkflowStatus.Type).Annotations(cascade),
		edge.To("labels", Label.Type).Annotations(cascade),
		edge.To("label_groups", LabelGroup.Type).Annotations(cascade),
		edge.To("boards", Board.Type).Annotations(cascade),
		edge.To("custom_fields", CustomFieldDefinition.Type).Annotations(cascade),
		edge.To("channels", Channel.Type).Annotations(cascade),
		edge.To("test_cases", TestCase.Type).Annotations(cascade),
		edge.To("test_runs", TestRun.Type).Annotations(cascade),
		edge.To("test_suites", TestSuite.Type).Annotations(cascade),
		edge.To("components", Component.Type).Annotations(cascade),
		edge.To("saved_filters", SavedFilter.Type).Annotations(cascade),
		edge.To("github_integrations", GitHubIntegration.Type).Annotations(cascade),
		edge.To("gitlab_integrations", GitLabIntegration.Type).Annotations(cascade),
		edge.To("recordings", Recording.Type).Annotations(cascade),
		edge.To("scheduled_calls", ScheduledCall.Type).Annotations(cascade),
		edge.To("webhook_endpoints", WebhookEndpoint.Type).Annotations(cascade),
		edge.To("environments", Environment.Type).Annotations(cascade),
		edge.To("sprints", Sprint.Type).Annotations(cascade),
		edge.To("automation_rules", AutomationRule.Type).Annotations(cascade),
		edge.To("crm_deals", CrmDeal.Type).Annotations(cascade),
	}
}

// Индексы сущности Project.
func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("deleted_at"),
	}
}
