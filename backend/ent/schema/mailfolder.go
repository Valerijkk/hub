package schema

import (
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
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

type MailFolder struct {
	ent.Schema
}

func (MailFolder) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).Immutable(),
		field.String("engine_folder_id").Optional().Nillable().Validate(validateMailFolderEngineID),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("role").Values("inbox", "sent", "drafts", "archive", "trash", "junk").Optional().Nillable().Immutable(),
		field.String("name").Validate(validateMailFolderName),
		field.String("normalized_name").Validate(validateMailFolderNormalizedName),
		field.String("color").Optional().Nillable().Validate(validateMailFolderColor),
		field.Int32("sort_order").Default(0),
		field.Int32("unread_count").Default(0).NonNegative(),
		field.Int32("total_count").Default(0).NonNegative(),
		field.Enum("control_state").Values("creating", "active", "unknown", "deleting", "error"),
		field.UUID("control_receipt_id", uuid.UUID{}).Optional().Nillable(),
		field.Int64("last_seen_sync_epoch").Default(0).Validate(validateMailProjectionEpoch),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (MailFolder) Mixin() []ent.Mixin {
	return []ent.Mixin{VersionMixin{}}
}

func (MailFolder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_folders_engine_folder_id":  "COALESCE(engine_folder_id IS NULL OR octet_length(engine_folder_id) BETWEEN 1 AND 1024, FALSE)",
			"mail_folders_text":              "COALESCE(char_length(name) BETWEEN 1 AND 120 AND octet_length(name) BETWEEN 1 AND 480 AND octet_length(normalized_name) BETWEEN 1 AND 480 AND (color IS NULL OR (char_length(color) BETWEEN 1 AND 32 AND octet_length(color) BETWEEN 1 AND 128)), FALSE)",
			"mail_folders_role":              "COALESCE(role IS NULL OR role IN ('inbox','sent','drafts','archive','trash','junk'), FALSE)",
			"mail_folders_control_shape":     "COALESCE((control_state = 'active' AND engine_folder_id IS NOT NULL AND control_receipt_id IS NULL AND deleted_at IS NULL) OR (control_state = 'creating' AND engine_folder_id IS NULL AND control_receipt_id IS NOT NULL AND deleted_at IS NULL) OR (control_state IN ('unknown','error') AND control_receipt_id IS NOT NULL AND deleted_at IS NULL) OR (control_state = 'deleting' AND engine_folder_id IS NOT NULL AND ((deleted_at IS NULL AND control_receipt_id IS NOT NULL) OR (deleted_at IS NOT NULL AND control_receipt_id IS NULL))), FALSE)",
			"mail_folders_counters":          "COALESCE(unread_count >= 0 AND total_count >= 0 AND unread_count <= total_count, FALSE)",
			"mail_folders_non_active_counts": "COALESCE(control_state = 'active' OR (unread_count = 0 AND total_count = 0), FALSE)",
			"mail_folders_role_parent":       "COALESCE(role IS NULL OR parent_id IS NULL, FALSE)",
			"mail_folders_not_self_parent":   "COALESCE(parent_id IS NULL OR parent_id <> id, FALSE)",
			"mail_folders_epoch":             "COALESCE(last_seen_sync_epoch BETWEEN 0 AND 2147483647, FALSE)",
			"mail_folders_version":           "COALESCE(version >= 1, FALSE)",
		}),
	}
}

func (MailFolder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", MailAccount.Type).Ref("folders").Field("mail_account_id").Unique().Required().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("parent", MailFolder.Type).Ref("children").Field("parent_id").Unique().Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.To("children", MailFolder.Type),
		edge.From("control_receipt", MailMutationReceipt.Type).Ref("folders").Field("control_receipt_id").Unique().Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.To("message_folders", MailMessageFolder.Type),
	}
}

func (MailFolder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mail_account_id", "engine_folder_id").Unique().StorageKey("mail_folders_account_engine_folder_uidx").Annotations(entsql.IndexWhere("engine_folder_id IS NOT NULL")),
		index.Fields("mail_account_id", "role").Unique().StorageKey("mail_folders_live_role_uidx").Annotations(entsql.IndexWhere("deleted_at IS NULL AND role IS NOT NULL")),
		index.Fields("mail_account_id", "normalized_name").Unique().StorageKey("mail_folders_live_custom_root_uidx").Annotations(entsql.IndexWhere("deleted_at IS NULL AND role IS NULL AND parent_id IS NULL")),
		index.Fields("mail_account_id", "parent_id", "normalized_name").Unique().StorageKey("mail_folders_live_custom_child_uidx").Annotations(entsql.IndexWhere("deleted_at IS NULL AND role IS NULL AND parent_id IS NOT NULL")),
	}
}

func validateMailFolderEngineID(value string) error {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > 1024 {
		return errors.New("invalid mail folder engine id")
	}
	return nil
}

func validateMailFolderName(value string) error {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > 480 || utf8.RuneCountInString(value) > 120 {
		return errors.New("invalid mail folder name")
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return errors.New("invalid mail folder name")
		}
	}
	return nil
}

func validateMailFolderNormalizedName(value string) error {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > 480 || normalizeMailNameV1(value) != value {
		return errors.New("invalid normalized mail folder name")
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return errors.New("invalid normalized mail folder name")
		}
	}
	return nil
}

func normalizeMailNameV1(value string) string {
	value = strings.Join(strings.FieldsFunc(value, unicode.IsSpace), " ")
	value = norm.NFC.String(value)
	value = cases.Fold().String(value)
	return norm.NFC.String(value)
}

func validateMailFolderColor(value string) error {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > 128 || utf8.RuneCountInString(value) > 32 {
		return errors.New("invalid mail folder color")
	}
	return nil
}

func validateMailProjectionEpoch(value int64) error {
	if value < 0 || value > 2_147_483_647 {
		return errors.New("invalid mail projection epoch")
	}
	return nil
}
