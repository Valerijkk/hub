package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("email").
			Unique().
			NotEmpty(),
		field.String("name").
			Optional(),
		field.String("display_name").
			Optional(),
		field.String("avatar_url").
			Optional(),
		field.String("telegram_id").
			Optional().
			Unique().
			Nillable(),
		field.String("phone").
			Optional(),
		field.String("position").
			Optional(),
		field.String("department").
			Optional(),
		field.String("timezone").
			Default("UTC"),
		field.String("language").
			Default("ru"),
		field.Time("birthday").
			Optional().
			Nillable(),
		field.Enum("system_role").
			Values("admin", "member").
			Default("member"),
		field.Bool("onboarding_completed").
			Default(false),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sessions", Session.Type),
		edge.To("project_members", ProjectMember.Type),
		edge.To("team_members", TeamMember.Type),
		edge.To("files", FileMetadata.Type).
			StorageKey(edge.Column("uploader_id")),
		edge.To("invites", Invite.Type),
		edge.To("comments", Comment.Type),
		edge.To("reactions", Reaction.Type),
		edge.To("activity_logs", ActivityLog.Type),
		edge.To("time_logs", TimeLog.Type),
		edge.To("created_channels", Channel.Type),
		edge.To("channel_memberships", ChannelMember.Type),
		edge.To("bookmarks", Bookmark.Type),
		edge.To("created_test_cases", TestCase.Type),
		edge.To("created_runs", TestRun.Type),
		edge.To("executed_results", TestResult.Type),
		edge.To("test_case_versions", TestCaseVersion.Type),
		edge.To("created_calls", Call.Type),
		edge.To("call_participations", CallParticipant.Type),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("telegram_id"),
	}
}
