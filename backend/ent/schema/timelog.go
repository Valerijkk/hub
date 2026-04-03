package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// TimeLog holds the schema definition for the TimeLog entity.
type TimeLog struct {
	ent.Schema
}

// Fields of the TimeLog.
func (TimeLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Int("minutes").
			Positive(),
		field.String("description").
			Optional().
			MaxLen(500),
		field.Time("started_at").
			Optional().
			Nillable().
			Comment("Set if timer-based"),
		field.Time("ended_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the TimeLog.
func (TimeLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("issue", Issue.Type).
			Ref("time_logs").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("time_logs").
			Unique().
			Required(),
	}
}

// Indexes of the TimeLog.
func (TimeLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("issue"),
	}
}
