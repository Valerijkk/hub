package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

// Board holds the schema definition for the Board entity.
type Board struct {
	ent.Schema
}

// Fields of the Board.
func (Board) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(100),
		field.JSON("filters", map[string]interface{}{}).
			Optional().
			Comment("Saved filter config: status, assignee, priority, type"),
		field.Bool("is_default").
			Default(false),
		field.Enum("swimlane_mode").
			Values("none", "assignee", "priority", "type").
			Default("none").
			Comment("Режим swimlanes на доске"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Board.
func (Board) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("boards").
			Unique().
			Required(),
	}
}
