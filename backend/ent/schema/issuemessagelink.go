package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// IssueMessageLink — связь задачи ↔ треда мессенджера (PROD-03).
type IssueMessageLink struct {
	ent.Schema
}

// Поля сущности IssueMessageLink.
func (IssueMessageLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("issue_id", uuid.UUID{}),
		field.UUID("channel_id", uuid.UUID{}),
		field.String("message_id").
			NotEmpty().
			MaxLen(64).
			Comment("Mongo thread root message ID"),
		field.UUID("created_by_id", uuid.UUID{}),
		field.Bool("mirror_to_thread").
			Default(true).
			Comment("Mirror new issue comments into linked thread"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности IssueMessageLink.
func (IssueMessageLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("issue", Issue.Type).
			Ref("message_links").
			Field("issue_id").
			Required().
			Unique(),
		edge.From("channel", Channel.Type).
			Ref("issue_message_links").
			Field("channel_id").
			Required().
			Unique(),
		edge.From("created_by", User.Type).
			Ref("issue_message_links").
			Field("created_by_id").
			Required().
			Unique(),
	}
}

// Индексы сущности IssueMessageLink.
func (IssueMessageLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("issue_id", "channel_id", "message_id").
			Unique(),
		index.Fields("channel_id", "message_id"),
	}
}
