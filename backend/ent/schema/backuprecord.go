package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// BackupRecord — определение Ent-схемы сущности BackupRecord.
type BackupRecord struct {
	ent.Schema
}

// Поля сущности BackupRecord.
func (BackupRecord) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("s3_key").
			NotEmpty().
			Comment("Ключ объекта в S3 (backups/{timestamp}.zip)"),
		field.Int64("file_size").
			Default(0).
			Comment("Размер архива в байтах"),
		field.Enum("status").
			Values("in_progress", "completed", "failed").
			Default("in_progress").
			Comment("Текущий статус резервного копирования"),
		field.Text("error_message").
			Optional().
			Nillable().
			Comment("Сообщение об ошибке при неудаче"),
		field.Time("restored_at").
			Optional().
			Nillable().
			Comment("W7-A4-018: момент успешного restore (не путать с error_message)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности BackupRecord.
func (BackupRecord) Edges() []ent.Edge {
	return nil // Системная сущность без связей
}

// Индексы сущности BackupRecord.
func (BackupRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("created_at"),
	}
}
