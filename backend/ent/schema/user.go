package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
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
		field.Enum("status").
			Values("active", "deactivated").
			Default("active").
			Comment("РЎС‚Р°С‚СѓСЃ РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ: active РёР»Рё deactivated"),
		field.Bool("onboarding_completed").
			Default(false),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.String("quiet_hours_start").
			Optional().
			Nillable().
			Comment("РќР°С‡Р°Р»Рѕ С‚РёС…РёС… С‡Р°СЃРѕРІ РІ С„РѕСЂРјР°С‚Рµ HH:MM"),
		field.String("quiet_hours_end").
			Optional().
			Nillable().
			Comment("РљРѕРЅРµС† С‚РёС…РёС… С‡Р°СЃРѕРІ РІ С„РѕСЂРјР°С‚Рµ HH:MM"),
		field.JSON("custom_fields", map[string]interface{}{}).
			Optional().
			Comment("РџРѕР»СЊР·РѕРІР°С‚РµР»СЊСЃРєРёРµ РєР°СЃС‚РѕРјРЅС‹Рµ РїРѕР»СЏ РїСЂРѕС„РёР»СЏ (JSONB)"),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
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
		edge.To("saved_filters", SavedFilter.Type),
		edge.To("notifications", Notification.Type),
		edge.To("notification_preferences", NotificationPreference.Type),
		edge.To("dashboard_layouts", DashboardLayout.Type),
		edge.To("api_keys", APIKey.Type),
		edge.To("recordings", Recording.Type),
		edge.To("scheduled_calls", ScheduledCall.Type),
		edge.To("project_templates", ProjectTemplate.Type),
		edge.To("polls", Poll.Type),
		edge.To("poll_votes", PollVote.Type),
		edge.To("push_subscriptions", PushSubscription.Type),
		edge.To("channel_invites", ChannelInvite.Type),
		edge.To("blocks_created", UserBlock.Type),
		edge.To("blocks_received", UserBlock.Type),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("telegram_id"),
		index.Fields("email"),
		index.Fields("status"),
	}
}
