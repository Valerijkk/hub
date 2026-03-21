package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// AuditLog хранит записи журнала аудита всех действий в системе.
type AuditLog struct {
	ent.Schema
}

// Fields of the AuditLog.
func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("action").
			NotEmpty().
			Comment("Действие: login, logout, create, update, delete"),
		field.String("entity_type").
			NotEmpty().
			Comment("Тип сущности: issue, project, test_case, channel, file, user"),
		field.String("entity_id").
			Optional().
			Comment("ID сущности (UUID или ключ)"),
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID пользователя, выполнившего действие"),
		field.UUID("project_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID проекта (если действие привязано к проекту)"),
		field.JSON("details", map[string]interface{}{}).
			Optional().
			Comment("Дополнительные данные о действии в JSON"),
		field.String("ip_address").
			Optional().
			Comment("IP-адрес клиента"),
		field.String("user_agent").
			Optional().
			Comment("User-Agent клиента"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Indexes of the AuditLog.
func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("action"),
		index.Fields("entity_type"),
		index.Fields("user_id"),
		index.Fields("project_id"),
		index.Fields("created_at"),
	}
}
