package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// LeaveRequest — заявка на отпуск (MOD-06 HR-lite).
type LeaveRequest struct {
	ent.Schema
}

func (LeaveRequest) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Enum("leave_type").
			Values("vacation", "sick", "personal", "other").
			Default("vacation"),
		field.Enum("status").
			Values("pending", "approved", "rejected", "cancelled").
			Default("pending"),
		field.Time("start_date").
			Comment("UTC date, inclusive"),
		field.Time("end_date").
			Comment("UTC date, inclusive"),
		field.Text("reason").
			Optional(),
		field.Text("review_note").
			Optional(),
		field.Time("reviewed_at").
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
func (LeaveRequest) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


func (LeaveRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("requester", User.Type).
			Ref("leave_requests").
			Unique().
			Required(),
		edge.From("reviewer", User.Type).
			Ref("leave_reviews").
			Unique(),
	}
}

func (LeaveRequest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "start_date"),
		index.Fields("start_date", "end_date"),
	}
}
