package schema

import (
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// FileMetadata — определение Ent-схемы сущности FileMetadata.
type FileMetadata struct {
	ent.Schema
}

// validateFileEntityID — UUID или sentinel storage («me»/«temp»).
// Пустая строка допустима: поле Optional (entity ещё не привязан).
func validateFileEntityID(s string) error {
	if s == "" {
		return nil
	}
	switch s {
	case "me", "temp":
		return nil
	}
	if _, err := uuid.Parse(s); err != nil {
		return fmt.Errorf("entity_id must be UUID or sentinel me/temp: %w", err)
	}
	return nil
}

// Поля сущности FileMetadata.
func (FileMetadata) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("filename").
			NotEmpty(),
		field.String("content_type").
			NotEmpty(),
		// A3W8-2-001: size >= 0 (defense-in-depth к storage/service).
		field.Int64("size").
			NonNegative(),
		field.String("s3_key").
			NotEmpty(),
		field.String("s3_bucket").
			NotEmpty(),
		field.String("entity_type").
			Optional().
			Comment("task, message, testcase, avatar, etc."),
		// A3W8-2-001: UUID или sentinel me/temp (storage avatar/voice).
		field.String("entity_id").
			Optional().
			Validate(validateFileEntityID),
		field.Enum("status").
			Values("pending", "confirmed", "deleted").
			Default("pending"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Soft-delete timestamp (30-day trash)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности FileMetadata.
func (FileMetadata) Edges() []ent.Edge {
	return []ent.Edge{
		// A3W8-2-001: uploader обязателен.
		edge.From("uploader", User.Type).
			Ref("files").
			Unique().
			Required(),
		edge.From("project", Project.Type).
			Ref("files").
			Unique(),
	}
}

// Индексы сущности FileMetadata.
func (FileMetadata) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("entity_type", "entity_id"),
		index.Fields("status"),
		index.Fields("s3_key"),
		index.Fields("deleted_at"),
	}
}
