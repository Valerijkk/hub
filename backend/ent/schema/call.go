package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Call хранит метаданные голосового звонка.
// Каждый звонок привязан к каналу мессенджера и содержит информацию
// о комнате LiveKit, статусе и количестве участников.
type Call struct {
	ent.Schema
}

// Fields описывает поля сущности Call.
func (Call) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("room_name").
			Unique().
			NotEmpty().
			Comment("Уникальное имя комнаты LiveKit"),
		field.UUID("channel_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID канала мессенджера (nil для instant-звонков)"),
		field.String("invite_code").
			Optional().
			Unique().
			MaxLen(12).
			Comment("Код приглашения для instant-звонка (например hub-abc123)"),
		field.String("title").
			Optional().
			MaxLen(255).
			Comment("Название звонка (для instant-звонков без канала)"),
		field.UUID("created_by", uuid.UUID{}).
			Comment("ID пользователя, создавшего звонок"),
		field.Enum("status").
			Values("active", "ended").
			Default("active").
			Comment("Статус звонка: активен или завершён"),
		field.Time("started_at").
			Optional().
			Nillable().
			Comment("Время начала звонка (устанавливается по webhook room_started)"),
		field.Time("ended_at").
			Optional().
			Nillable().
			Comment("Время завершения звонка"),
		field.Int("participant_count").
			Default(0).
			Comment("Текущее количество участников в звонке"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges описывает связи сущности Call.
func (Call) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", Channel.Type).
			Ref("calls").
			Field("channel_id").
			Unique(),
		edge.From("creator", User.Type).
			Ref("created_calls").
			Field("created_by").
			Unique().
			Required(),
		edge.To("participants", CallParticipant.Type),
		edge.To("recordings", Recording.Type),
	}
}

// Indexes описывает индексы сущности Call.
func (Call) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "status"),
		index.Fields("room_name"),
		index.Fields("invite_code"),
	}
}
