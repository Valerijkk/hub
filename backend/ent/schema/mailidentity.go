package schema

import (
	"errors"
	"strings"
	"unicode/utf8"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type MailIdentity struct {
	ent.Schema
}

func (MailIdentity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).
			Immutable(),
		field.String("normalized_address").
			MinLen(1).
			MaxLen(320).
			Validate(validateNormalizedMailAddress).
			Immutable(),
		field.String("display_name").
			Optional().
			Nillable().
			MaxLen(640).
			Validate(validateMailDisplayName),
		field.Enum("kind").
			Values("primary", "alias", "group").
			Immutable(),
		field.Bool("send_allowed").
			Default(false),
		field.Enum("status").
			Values("pending", "active", "error", "deleting").
			Default("pending"),
		field.String("engine_identity_id").
			Optional().
			Nillable().
			MinLen(1).
			MaxLen(1024),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

func (MailIdentity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_identities_send_allowed":       "send_allowed = FALSE OR status = 'active'",
			"mail_identities_group_not_sendable": "kind <> 'group' OR send_allowed = FALSE",
			"mail_identities_deleting_alias_only": "status <> 'deleting' OR kind = 'alias'",
		}),
	}
}

func (MailIdentity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", MailAccount.Type).
			Ref("identities").
			Field("mail_account_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("default_for_settings", MailUserSettings.Type).
			Unique(),
	}
}

func (MailIdentity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("normalized_address").
			Unique(),
		index.Fields("mail_account_id", "engine_identity_id").
			Unique().
			Annotations(entsql.IndexWhere("engine_identity_id IS NOT NULL")),
		index.Fields("mail_account_id").
			Unique().
			Annotations(entsql.IndexWhere("kind = 'primary'")),
		index.Fields("mail_account_id", "status"),
	}
}

func validateNormalizedMailAddress(value string) error {
	if len(value) == 0 || len(value) > 320 || strings.Count(value, "@") != 1 {
		return errors.New("invalid normalized mail address")
	}
	parts := strings.SplitN(value, "@", 2)
	if validateNormalizedMailLocalPart(parts[0]) != nil || validateNormalizedMailDomain(parts[1]) != nil {
		return errors.New("invalid normalized mail address")
	}
	return nil
}

func validateMailDisplayName(value string) error {
	if !utf8.ValidString(value) || len(value) > 640 || utf8.RuneCountInString(value) > 160 {
		return errors.New("invalid mail display name")
	}
	return nil
}
