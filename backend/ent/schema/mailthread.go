package schema

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/apparentlymart/go-textseg/v13/textseg"
	"github.com/google/uuid"
)

type MailThread struct {
	ent.Schema
}

func (MailThread) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).Immutable(),
		field.Int32("thread_key_version").Default(1).Immutable(),
		field.Bytes("thread_key").MinLen(32).MaxLen(32).Immutable(),
		field.String("normalized_subject").Validate(validateMailThreadSubject),
		field.JSON("participant_addresses", MailThreadParticipants{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("last_message_at"),
		field.Int32("message_count").Min(1),
		field.Int32("unread_count").NonNegative(),
		field.Bool("is_pinned"),
		field.Bool("has_attachments"),
		field.String("snippet").Validate(validateMailProjectionSnippet),
		field.Int64("last_seen_sync_epoch").Default(0).Validate(validateMailProjectionEpoch),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (MailThread) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_threads_key":          "COALESCE(thread_key_version = 1 AND octet_length(thread_key) = 32, FALSE)",
			"mail_threads_subject":      "COALESCE(octet_length(normalized_subject) <= 998, FALSE)",
			"mail_threads_participants": "CASE WHEN jsonb_typeof(participant_addresses) = 'array' THEN COALESCE(jsonb_array_length(participant_addresses) BETWEEN 1 AND 401 AND NOT jsonb_path_exists(participant_addresses, '$[*] ? (@.type() != \"string\")'), FALSE) ELSE FALSE END",
			"mail_threads_counts":       "COALESCE(message_count >= 1 AND unread_count BETWEEN 0 AND message_count, FALSE)",
			"mail_threads_snippet":      "COALESCE(char_length(snippet) <= 1200 AND octet_length(snippet) <= 4096, FALSE)",
			"mail_threads_epoch":        "COALESCE(last_seen_sync_epoch BETWEEN 0 AND 2147483647, FALSE)",
		}),
	}
}

func (MailThread) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", MailAccount.Type).Ref("threads").Field("mail_account_id").Unique().Required().Immutable().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("messages", MailMessage.Type),
	}
}

func (MailThread) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mail_account_id", "thread_key_version", "thread_key").Unique().StorageKey("mail_threads_account_version_key_uidx"),
		index.Fields("mail_account_id", "is_pinned", "last_message_at", "id").StorageKey("mail_threads_live_pinned_idx").Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

type MailThreadParticipants []string

func (participants MailThreadParticipants) Validate() error {
	if len(participants) < 1 || len(participants) > 401 {
		return errors.New("mail thread participants must contain 1..401 addresses")
	}
	seen := make(map[string]struct{}, len(participants))
	for _, address := range participants {
		if validateNormalizedMailAddress(address) != nil {
			return errors.New("invalid mail thread participant address")
		}
		if _, found := seen[address]; found {
			return errors.New("duplicate mail thread participant address")
		}
		seen[address] = struct{}{}
	}
	encoded, err := json.Marshal(participants)
	if err != nil {
		return err
	}
	if len(encoded) > 1_048_576 {
		return errors.New("mail thread participants exceed 1048576 bytes")
	}
	return nil
}

func (participants MailThreadParticipants) MarshalJSON() ([]byte, error) {
	if err := participants.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal([]string(participants))
}

func (participants *MailThreadParticipants) UnmarshalJSON(data []byte) error {
	if err := preflightMailJSON(data, 6*1_048_576, 6*320+2); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return errors.New("mail thread participants must be an array")
	}
	decoded := make(MailThreadParticipants, 0, 401)
	for decoder.More() {
		if len(decoded) == 401 {
			return errors.New("mail thread participants must contain 1..401 addresses")
		}
		var address string
		if err := decoder.Decode(&address); err != nil {
			return err
		}
		decoded = append(decoded, address)
	}
	token, err = decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != ']' {
		return errors.New("mail thread participants must end with an array")
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("mail thread participants have trailing data")
		}
		return err
	}
	if err := decoded.Validate(); err != nil {
		return err
	}
	*participants = decoded
	return nil
}

func validateMailThreadSubject(value string) error {
	if !utf8.ValidString(value) || len(value) > 998 {
		return errors.New("invalid normalized mail thread subject")
	}
	return nil
}

func validateMailProjectionSnippet(value string) error {
	if !utf8.ValidString(value) || len(value) > 4096 {
		return errors.New("invalid mail projection snippet")
	}
	scanner := bufio.NewScanner(strings.NewReader(value))
	scanner.Buffer(make([]byte, 4097), 4097)
	scanner.Split(textseg.ScanGraphemeClusters)
	count := 0
	for scanner.Scan() {
		count++
		if count > 400 {
			return errors.New("mail projection snippet exceeds 400 grapheme clusters")
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.New("invalid mail projection snippet")
	}
	return nil
}
