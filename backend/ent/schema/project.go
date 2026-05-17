package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Fields of the Project.
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
			Comment("РЎС‡С‘С‚С‡РёРє С‚РµСЃС‚-РєРµР№СЃРѕРІ РґР»СЏ РіРµРЅРµСЂР°С†РёРё РєР»СЋС‡РµР№"),
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

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", ProjectMember.Type),
		edge.To("roles", Role.Type),
		edge.To("teams", Team.Type),
		edge.To("invites", Invite.Type),
		edge.To("files", FileMetadata.Type),
		edge.To("issues", Issue.Type),
		edge.To("workflow_statuses", WorkflowStatus.Type),
		edge.To("labels", Label.Type),
		edge.To("label_groups", LabelGroup.Type),
		edge.To("boards", Board.Type),
		edge.To("custom_fields", CustomFieldDefinition.Type),
		edge.To("channels", Channel.Type),
		edge.To("test_cases", TestCase.Type),
		edge.To("test_runs", TestRun.Type),
		edge.To("test_suites", TestSuite.Type),
		edge.To("components", Component.Type),
		edge.To("saved_filters", SavedFilter.Type),
		edge.To("github_integrations", GitHubIntegration.Type),
		edge.To("gitlab_integrations", GitLabIntegration.Type),
		edge.To("recordings", Recording.Type),
		edge.To("scheduled_calls", ScheduledCall.Type),
		edge.To("webhook_endpoints", WebhookEndpoint.Type),
		edge.To("environments", Environment.Type),
		edge.To("sprints", Sprint.Type),
	}
}

// Indexes of the Project.
func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("deleted_at"),
	}
}
