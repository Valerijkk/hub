package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Recording хранит метаданные записи звонка.
// Запись создаётся через LiveKit Egress и сохраняется в SeaweedFS (S3).
// Привязана к звонку, проекту и пользователю, начавшему запись.
type Recording struct {
	ent.Schema
}

// Fields описывает поля сущности Recording.
func (Recording) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("egress_id").
			NotEmpty().
			Comment("ID egress-процесса в LiveKit"),
		field.Enum("status").
			Values("recording", "processing", "ready", "failed").
			Default("recording").
			Comment("Статус записи: recording -> processing -> ready/failed"),
		field.String("file_path").
			Optional().
			Comment("Ключ файла в S3 (recordings/{projectID}/{callID}.mp4)"),
		field.Int64("file_size").
			Default(0).
			Comment("Размер файла записи в байтах"),
		field.Int("duration_seconds").
			Default(0).
			Comment("Длительность записи в секундах"),
		field.Time("started_at").
			Default(time.Now).
			Comment("Время начала записи"),
		field.Time("ended_at").
			Optional().
			Nillable().
			Comment("Время окончания записи"),
		field.UUID("call_id", uuid.UUID{}).
			Comment("ID звонка, к которому относится запись"),
		field.UUID("project_id", uuid.UUID{}).
			Comment("ID проекта для привязки записей к проекту"),
		field.UUID("started_by", uuid.UUID{}).
			Comment("ID пользователя, начавшего запись"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges описывает связи сущности Recording.
func (Recording) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("call", Call.Type).
			Ref("recordings").
			Field("call_id").
			Unique().
			Required(),
		edge.From("project", Project.Type).
			Ref("recordings").
			Field("project_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("recordings").
			Field("started_by").
			Unique().
			Required(),
	}
}

// Indexes описывает индексы сущности Recording.
func (Recording) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id", "created_at"),
		index.Fields("call_id"),
		index.Fields("egress_id"),
	}
}
