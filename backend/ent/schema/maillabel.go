package schema

import (
	"encoding/base32"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type MailLabel struct {
	ent.Schema
}

func (MailLabel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).Immutable(),
		field.String("name").Validate(validateMailLabelName),
		field.String("normalized_name").Validate(validateMailLabelNormalizedName),
		field.String("color").Validate(validateMailLabelColor),
		field.String("engine_keyword").Validate(validateMailLabelKeyword).Immutable(),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (MailLabel) Mixin() []ent.Mixin {
	return []ent.Mixin{VersionMixin{}}
}

func (MailLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_labels_text":    "COALESCE(char_length(name) BETWEEN 1 AND 60 AND octet_length(name) BETWEEN 1 AND 240 AND octet_length(normalized_name) BETWEEN 1 AND 240 AND char_length(color) BETWEEN 1 AND 32 AND octet_length(color) BETWEEN 1 AND 128 AND octet_length(engine_keyword) = 57 AND engine_keyword ~ '^hubk-[a-z2-7]{52}$', FALSE)",
			"mail_labels_version": "COALESCE(version >= 1, FALSE)",
		}),
	}
}

func (MailLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", MailAccount.Type).Ref("labels").Field("mail_account_id").Unique().Required().Immutable().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("message_labels", MailMessageLabel.Type),
	}
}

func (MailLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mail_account_id", "normalized_name").Unique().StorageKey("mail_labels_live_normalized_name_uidx").Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("mail_account_id", "engine_keyword").Unique().StorageKey("mail_labels_account_keyword_uidx"),
	}
}

func validateMailLabelName(value string) error {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > 240 || utf8.RuneCountInString(value) > 60 {
		return errors.New("invalid mail label name")
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return errors.New("invalid mail label name")
		}
	}
	return nil
}

func validateMailLabelNormalizedName(value string) error {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > 240 || normalizeMailNameV1(value) != value {
		return errors.New("invalid normalized mail label name")
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return errors.New("invalid normalized mail label name")
		}
	}
	return nil
}

func validateMailLabelColor(value string) error {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > 128 || utf8.RuneCountInString(value) > 32 {
		return errors.New("invalid mail label color")
	}
	return nil
}

func validateMailLabelKeyword(value string) error {
	if len(value) != 57 || !strings.HasPrefix(value, "hubk-") {
		return errors.New("invalid mail label engine keyword")
	}
	suffix := value[len("hubk-"):]
	for _, character := range suffix {
		if (character < 'a' || character > 'z') && (character < '2' || character > '7') {
			return errors.New("invalid mail label engine keyword")
		}
	}
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(suffix))
	if err != nil || len(decoded) != 32 {
		return errors.New("invalid mail label engine keyword")
	}
	if strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(decoded)) != suffix {
		return errors.New("invalid mail label engine keyword")
	}
	return nil
}
