package schema

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type MailMessage struct {
	ent.Schema
}

func (MailMessage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).Immutable(),
		field.UUID("mail_thread_id", uuid.UUID{}),
		field.String("engine_message_id").Validate(validateMailMessageOpaqueID).Immutable(),
		field.String("engine_version").Validate(validateMailMessageOpaqueID),
		field.String("internet_message_id").Optional().Nillable().Validate(validateMailMessageID).Annotations(entsql.Annotation{Collation: "C"}),
		field.String("in_reply_to").Optional().Nillable().Validate(validateMailMessageID),
		field.JSON("references", MailMessageReferences{}).Default(MailMessageReferences{}).SchemaType(map[string]string{dialect.Postgres: "jsonb"}).Validate(func(value MailMessageReferences) error { return value.Validate() }),
		field.JSON("from", MailStoredAddress{}).SchemaType(map[string]string{dialect.Postgres: "jsonb"}).Validate(func(value MailStoredAddress) error { return value.Validate() }),
		field.JSON("to", MailStoredAddresses{}).Default(MailStoredAddresses{}).SchemaType(map[string]string{dialect.Postgres: "jsonb"}).Validate(func(value MailStoredAddresses) error { return value.Validate() }),
		field.JSON("cc", MailStoredAddresses{}).Default(MailStoredAddresses{}).SchemaType(map[string]string{dialect.Postgres: "jsonb"}).Validate(func(value MailStoredAddresses) error { return value.Validate() }),
		field.JSON("bcc", MailStoredAddresses{}).Default(MailStoredAddresses{}).SchemaType(map[string]string{dialect.Postgres: "jsonb"}).Validate(func(value MailStoredAddresses) error { return value.Validate() }),
		field.JSON("reply_to", MailStoredAddresses{}).Default(MailStoredAddresses{}).SchemaType(map[string]string{dialect.Postgres: "jsonb"}).Validate(func(value MailStoredAddresses) error { return value.Validate() }),
		field.String("subject").Validate(validateMailMessageSubject),
		field.String("snippet").Validate(validateMailProjectionSnippet),
		field.Time("sent_at").Optional().Nillable(),
		field.Time("received_at").Optional().Nillable(),
		field.Int64("size_bytes").Min(0).Max(52_428_800),
		field.Enum("priority").Values("low", "normal", "high"),
		field.Bool("has_attachments"),
		field.Bool("is_draft"),
		field.Bool("is_seen"),
		field.Bool("is_answered"),
		field.Bool("is_flagged"),
		field.Bool("is_junk"),
		field.String("engine_state").Optional().Nillable().Validate(validateMailMessageEngineState),
		field.Enum("projection_origin").Values("inbound_or_unknown", "confirmed_owner_submission", "confirmed_owner_draft"),
		field.Time("flags_synced_at").Optional().Nillable(),
		field.Enum("calendar_method").Values("request", "cancel", "reply").Optional().Nillable(),
		field.UUID("calendar_event_id", uuid.UUID{}).Optional().Nillable(),
		field.Int64("last_seen_sync_epoch").Default(0).Validate(validateMailProjectionEpoch),
		field.UUID("received_event_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (MailMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_messages_engine_ids":        "COALESCE(octet_length(engine_message_id) BETWEEN 1 AND 1024 AND octet_length(engine_version) BETWEEN 1 AND 1024, FALSE)",
			"mail_messages_message_ids":       "COALESCE((internet_message_id IS NULL OR octet_length(internet_message_id) BETWEEN 1 AND 998) AND (in_reply_to IS NULL OR octet_length(in_reply_to) BETWEEN 1 AND 998), FALSE)",
			"mail_messages_references":        "CASE WHEN jsonb_typeof(references) = 'array' THEN COALESCE(jsonb_array_length(references) <= 100 AND NOT jsonb_path_exists(references, '$[*] ? (@.type() != \"string\")'), FALSE) ELSE FALSE END",
			"mail_messages_from_shape":        "COALESCE(jsonb_typeof(\"from\") = 'object' AND jsonb_typeof(\"from\"->'email') = 'string' AND jsonb_typeof(\"from\"->'name') = 'string', FALSE)",
			"mail_messages_address_arrays":    "CASE WHEN jsonb_typeof(\"to\") = 'array' AND jsonb_typeof(cc) = 'array' AND jsonb_typeof(bcc) = 'array' AND jsonb_typeof(reply_to) = 'array' THEN COALESCE(jsonb_array_length(\"to\") <= 200 AND jsonb_array_length(cc) <= 200 AND jsonb_array_length(bcc) <= 200 AND jsonb_array_length(reply_to) <= 200 AND NOT jsonb_path_exists(\"to\", '$[*] ? (@.type() != \"object\" || !exists(@.email) || !exists(@.name) || @.email.type() != \"string\" || @.name.type() != \"string\")') AND NOT jsonb_path_exists(cc, '$[*] ? (@.type() != \"object\" || !exists(@.email) || !exists(@.name) || @.email.type() != \"string\" || @.name.type() != \"string\")') AND NOT jsonb_path_exists(bcc, '$[*] ? (@.type() != \"object\" || !exists(@.email) || !exists(@.name) || @.email.type() != \"string\" || @.name.type() != \"string\")') AND NOT jsonb_path_exists(reply_to, '$[*] ? (@.type() != \"object\" || !exists(@.email) || !exists(@.name) || @.email.type() != \"string\" || @.name.type() != \"string\")'), FALSE) ELSE FALSE END",
			"mail_messages_text":              "COALESCE(octet_length(subject) <= 998 AND char_length(snippet) <= 1200 AND octet_length(snippet) <= 4096 AND (engine_state IS NULL OR octet_length(engine_state) <= 4096), FALSE)",
			"mail_messages_size":              "COALESCE(size_bytes BETWEEN 0 AND 52428800, FALSE)",
			"mail_messages_epoch":             "COALESCE(last_seen_sync_epoch BETWEEN 0 AND 2147483647, FALSE)",
			"mail_messages_priority":          "COALESCE(priority IN ('low','normal','high'), FALSE)",
			"mail_messages_projection_origin": "COALESCE(projection_origin IN ('inbound_or_unknown','confirmed_owner_submission','confirmed_owner_draft'), FALSE)",
			"mail_messages_calendar_method":   "COALESCE(calendar_method IS NULL OR calendar_method IN ('request','cancel','reply'), FALSE)",
			"mail_messages_bcc_origin":        "CASE WHEN jsonb_typeof(bcc) = 'array' THEN COALESCE(projection_origin IN ('confirmed_owner_submission','confirmed_owner_draft') OR jsonb_array_length(bcc) = 0, FALSE) ELSE FALSE END",
		}),
	}
}

func (MailMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", MailAccount.Type).Ref("messages").Field("mail_account_id").Unique().Required().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("thread", MailThread.Type).Ref("messages").Field("mail_thread_id").Unique().Required().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("calendar_event", CalendarEvent.Type).Ref("mail_messages").Field("calendar_event_id").Unique().Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("folders", MailMessageFolder.Type),
		edge.To("labels", MailMessageLabel.Type),
		edge.To("attachments", MailAttachment.Type),
		edge.To("reply_drafts", MailDraft.Type),
		edge.To("forward_drafts", MailDraft.Type),
	}
}

func (MailMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mail_account_id", "engine_message_id").Unique().StorageKey("mail_messages_account_engine_uidx"),
		index.Fields("mail_account_id", "internet_message_id").StorageKey("mail_messages_live_internet_message_id_idx").Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("received_event_id").Unique().StorageKey("mail_messages_received_event_uidx"),
		index.Fields("mail_account_id", "received_at", "sent_at", "id").StorageKey("mail_messages_live_account_date_idx").Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("mail_thread_id", "received_at", "sent_at", "id").StorageKey("mail_messages_live_thread_date_idx").Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

type MailStoredAddress struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (address MailStoredAddress) Validate() error {
	if validateNormalizedMailAddress(address.Email) != nil || validateMailDisplayName(address.Name) != nil {
		return errors.New("invalid mail stored address")
	}
	for _, character := range address.Name {
		if unicode.IsControl(character) {
			return errors.New("invalid mail stored address")
		}
	}
	return nil
}

func (address MailStoredAddress) MarshalJSON() ([]byte, error) {
	if err := address.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}{Email: address.Email, Name: address.Name})
}

func (address *MailStoredAddress) UnmarshalJSON(data []byte) error {
	if err := preflightMailJSON(data, 6*(320+640)+len(`{"email":"","name":""}`), 6*640+2); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := requireJSONObject(decoder); err != nil {
		return err
	}
	var decoded MailStoredAddress
	var emailSeen, nameSeen bool
	for decoder.More() {
		name, err := nextJSONObjectName(decoder)
		if err != nil {
			return err
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return err
		}
		switch name {
		case "email":
			if emailSeen || bytes.Equal(raw, []byte("null")) {
				return errors.New("invalid mail stored address")
			}
			emailSeen = true
			if err := json.Unmarshal(raw, &decoded.Email); err != nil {
				return errors.New("invalid mail stored address")
			}
		case "name":
			if nameSeen || bytes.Equal(raw, []byte("null")) {
				return errors.New("invalid mail stored address")
			}
			nameSeen = true
			if err := json.Unmarshal(raw, &decoded.Name); err != nil {
				return errors.New("invalid mail stored address")
			}
		default:
			return errors.New("invalid mail stored address")
		}
	}
	if err := closeJSONObject(decoder); err != nil {
		return err
	}
	if !emailSeen || !nameSeen {
		return errors.New("invalid mail stored address")
	}
	if err := decoded.Validate(); err != nil {
		return err
	}
	*address = decoded
	return nil
}

type MailStoredAddresses []MailStoredAddress

func (addresses MailStoredAddresses) Validate() error {
	if addresses == nil || len(addresses) > 200 {
		return errors.New("invalid mail stored addresses")
	}
	seen := make(map[string]struct{}, len(addresses))
	for _, address := range addresses {
		if err := address.Validate(); err != nil {
			return err
		}
		if _, found := seen[address.Email]; found {
			return errors.New("duplicate mail stored address")
		}
		seen[address.Email] = struct{}{}
	}
	encoded, err := json.Marshal([]MailStoredAddress(addresses))
	if err != nil || len(encoded) > 65_536 {
		return errors.New("invalid mail stored addresses")
	}
	return nil
}

func (addresses MailStoredAddresses) MarshalJSON() ([]byte, error) {
	if err := addresses.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal([]MailStoredAddress(addresses))
}

func (addresses *MailStoredAddresses) UnmarshalJSON(data []byte) error {
	if err := preflightMailJSON(data, 6*65_536, 6*640+2); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return errors.New("invalid mail stored addresses")
	}
	decoded := MailStoredAddresses{}
	for decoder.More() {
		if len(decoded) == 200 {
			return errors.New("invalid mail stored addresses")
		}
		var address MailStoredAddress
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
		return errors.New("invalid mail stored addresses")
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("invalid mail stored addresses")
		}
		return err
	}
	if err := decoded.Validate(); err != nil {
		return err
	}
	*addresses = decoded
	return nil
}

type MailMessageReferences []string

func (references MailMessageReferences) Validate() error {
	if references == nil || len(references) > 100 {
		return errors.New("invalid mail message references")
	}
	seen := make(map[string]struct{}, len(references))
	for _, reference := range references {
		if err := validateMailMessageID(reference); err != nil {
			return err
		}
		if _, found := seen[reference]; found {
			return errors.New("duplicate mail message reference")
		}
		seen[reference] = struct{}{}
	}
	encoded, err := json.Marshal([]string(references))
	if err != nil || len(encoded) > 1_048_576 {
		return errors.New("invalid mail message references")
	}
	return nil
}

func (references MailMessageReferences) MarshalJSON() ([]byte, error) {
	if err := references.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal([]string(references))
}

func (references *MailMessageReferences) UnmarshalJSON(data []byte) error {
	if err := preflightMailJSON(data, 6*1_048_576, 6*998+2); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return errors.New("invalid mail message references")
	}
	decoded := MailMessageReferences{}
	for decoder.More() {
		if len(decoded) == 100 {
			return errors.New("invalid mail message references")
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return err
		}
		if bytes.Equal(raw, []byte("null")) {
			return errors.New("invalid mail message references")
		}
		var reference string
		if err := json.Unmarshal(raw, &reference); err != nil {
			return errors.New("invalid mail message references")
		}
		decoded = append(decoded, reference)
	}
	token, err = decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != ']' {
		return errors.New("invalid mail message references")
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("invalid mail message references")
		}
		return err
	}
	if err := decoded.Validate(); err != nil {
		return err
	}
	*references = decoded
	return nil
}

func validateMailMessageOpaqueID(value string) error {
	return validateMailMessageText(value, 1, 1_024)
}

func validateMailMessageEngineState(value string) error {
	return validateMailMessageText(value, 1, 4_096)
}

func validateMailMessageSubject(value string) error {
	return validateMailMessageText(value, 0, 998)
}

func validateMailMessageText(value string, minimum, maximum int) error {
	if !utf8.ValidString(value) || len(value) < minimum || len(value) > maximum {
		return errors.New("invalid mail message text")
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return errors.New("invalid mail message text")
		}
	}
	return nil
}

func validateMailMessageID(value string) error {
	if len(value) < 5 || len(value) > 998 || value[0] != '<' || value[len(value)-1] != '>' {
		return errors.New("invalid mail message id")
	}
	body := value[1 : len(value)-1]
	if strings.Count(body, "@") != 1 {
		return errors.New("invalid mail message id")
	}
	parts := strings.SplitN(body, "@", 2)
	if !isMailMessageDotAtom(parts[0]) || !isMailMessageDotAtom(parts[1]) {
		return errors.New("invalid mail message id")
	}
	return nil
}

func isMailMessageDotAtom(value string) bool {
	if len(value) == 0 || value[0] == '.' || value[len(value)-1] == '.' {
		return false
	}
	previousDot := false
	for _, character := range []byte(value) {
		if character == '.' {
			if previousDot {
				return false
			}
			previousDot = true
			continue
		}
		if !((character >= 'A' && character <= 'Z') || (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || isMailAText(character)) {
			return false
		}
		previousDot = false
	}
	return true
}

func preflightMailJSON(data []byte, maxCompact, maxStringToken int) error {
	compactBytes := 0
	stringTokenBytes := 0
	inString := false
	escaped := false
	for _, character := range data {
		if !inString && (character == ' ' || character == '\t' || character == '\r' || character == '\n') {
			continue
		}
		if compactBytes >= maxCompact {
			return errors.New("mail JSON lexical budget exceeded")
		}
		compactBytes++
		if !inString {
			if character == '"' {
				inString = true
				stringTokenBytes = 1
			}
			continue
		}
		if stringTokenBytes >= maxStringToken {
			return errors.New("mail JSON lexical budget exceeded")
		}
		stringTokenBytes++
		if escaped {
			escaped = false
			continue
		}
		if character == '\\' {
			escaped = true
			continue
		}
		if character == '"' {
			inString = false
		}
	}
	return nil
}
