package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// User — определение Ent-схемы сущности User.
type User struct {
	ent.Schema
}

// Поля сущности User.
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
			Values("founder", "admin", "member").
			Default("member"),
		field.Enum("status").
			Values("active", "pending", "deactivated").
			Default("active").
			Comment("User lifecycle: active, pending admin approval, deactivated"),
		field.Time("pending_since").
			Optional().
			Nillable().
			Comment("When user entered pending queue (7d auto-delete)"),
		field.Bool("onboarding_completed").
			Default(false),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.String("quiet_hours_start").
			Optional().
			Nillable().
			Comment("Начало тихих часов в формате HH:MM"),
		field.String("quiet_hours_end").
			Optional().
			Nillable().
			Comment("Конец тихих часов в формате HH:MM"),
		field.JSON("custom_fields", map[string]interface{}{}).
			Optional().
			Comment("Пользовательские кастомные поля профиля (JSONB)"),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Рёбра (edges) сущности User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("users").
			Unique(),
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
		edge.To("automation_rules", AutomationRule.Type),
		edge.To("issue_message_links", IssueMessageLink.Type),
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
		edge.To("ai_providers", AIProvider.Type),
		edge.To("ai_usage_daily", AIUsageDaily.Type),
		edge.To("ai_preference", AIUserPreference.Type),
		edge.To("news_posts", NewsPost.Type),
		edge.To("crm_contacts_created", CrmContact.Type),
		edge.To("crm_contact_link", CrmContact.Type),
		edge.To("crm_deals_owned", CrmDeal.Type),
		edge.To("leave_requests", LeaveRequest.Type),
		edge.To("leave_reviews", LeaveRequest.Type),
		edge.To("kudos_sent", Kudo.Type),
		edge.To("kudos_received", Kudo.Type),
		edge.To("marketplace_listings", MarketplaceListing.Type),
		edge.To("marketplace_reviews", MarketplaceListing.Type),
		edge.To("coffee_enrollment", CoffeeEnrollment.Type).Unique(),
		edge.To("coffee_rounds_triggered", CoffeeRound.Type),
		edge.To("coffee_pairs_a", CoffeePair.Type),
		edge.To("coffee_pairs_b", CoffeePair.Type),
		edge.To("hub_bases", HubBase.Type),
		edge.To("hub_records", HubRecord.Type),
		edge.To("journal_entries", JournalEntry.Type),
		edge.To("invoices_created", Invoice.Type),
	}
}

// Индексы сущности User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("telegram_id"),
		index.Fields("email"),
		index.Fields("status"),
	}
}
