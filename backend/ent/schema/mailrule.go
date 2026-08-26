package schema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

const maxMailRuleJSONBytes = 65_536

type MailRule struct {
	ent.Schema
}

func (MailRule) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).
			Immutable(),
		field.String("name").
			MinLen(1).
			MaxLen(480).
			Validate(validateMailRuleName),
		field.Bool("enabled"),
		field.Int32("priority").
			Min(0).
			Max(10_000),
		field.JSON("conditions", MailRuleConditions{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("actions", MailRuleActions{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

func (MailRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

func (MailRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_rules_json_shape":  "jsonb_typeof(conditions) = 'array' AND jsonb_typeof(actions) = 'array' AND jsonb_array_length(conditions) BETWEEN 1 AND 20 AND jsonb_array_length(actions) BETWEEN 1 AND 10 AND octet_length(conditions::text) + octet_length(actions::text) <= 65536",
			"mail_rules_name_bounds": "(char_length(name) BETWEEN 1 AND 120 AND octet_length(name) BETWEEN 1 AND 480)",
			"mail_rules_priority":    "priority BETWEEN 0 AND 10000",
		}),
	}
}

func (MailRule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", MailAccount.Type).
			Ref("rules").
			Field("mail_account_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Restrict)),
	}
}

func (MailRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mail_account_id", "enabled", "priority", "id").
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

type MailRuleConditions []MailRuleCondition

func (conditions MailRuleConditions) Validate() error {
	if len(conditions) < 1 || len(conditions) > 20 {
		return errors.New("mail rule conditions must contain 1..20 entries")
	}
	for index, condition := range conditions {
		if err := condition.Validate(); err != nil {
			return fmt.Errorf("mail rule condition %d: %w", index, err)
		}
	}
	return nil
}

type MailRuleCondition struct {
	Field        string
	Operator     string
	StringValue  *string
	BooleanValue *bool
}

func (condition MailRuleCondition) Validate() error {
	if (condition.StringValue == nil) == (condition.BooleanValue == nil) {
		return errors.New("mail rule condition must have exactly one typed value")
	}
	switch condition.Field {
	case "from", "to":
		if condition.StringValue == nil || (condition.Operator != "equals" && condition.Operator != "contains" && condition.Operator != "matches_domain") {
			return errors.New("invalid mail rule address condition")
		}
	case "subject":
		if condition.StringValue == nil || (condition.Operator != "equals" && condition.Operator != "contains") {
			return errors.New("invalid mail rule subject condition")
		}
	case "has_attachment":
		if condition.BooleanValue == nil || condition.Operator != "is" {
			return errors.New("invalid mail rule attachment condition")
		}
	default:
		return errors.New("invalid mail rule condition field")
	}
	if condition.StringValue != nil {
		return validateMailRuleConditionString(*condition.StringValue)
	}
	return nil
}

func (condition MailRuleCondition) MarshalJSON() ([]byte, error) {
	if err := condition.Validate(); err != nil {
		return nil, err
	}
	var value json.RawMessage
	var err error
	if condition.StringValue != nil {
		value, err = json.Marshal(*condition.StringValue)
	} else {
		value, err = json.Marshal(*condition.BooleanValue)
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Field    string          `json:"field"`
		Operator string          `json:"operator"`
		Value    json.RawMessage `json:"value"`
	}{Field: condition.Field, Operator: condition.Operator, Value: value})
}

func (condition *MailRuleCondition) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := requireJSONObject(decoder); err != nil {
		return err
	}
	var decoded MailRuleCondition
	var fieldSeen, operatorSeen, valueSeen bool
	for decoder.More() {
		name, err := nextJSONObjectName(decoder)
		if err != nil {
			return err
		}
		switch name {
		case "field":
			if fieldSeen {
				return errors.New("duplicate mail rule condition field")
			}
			fieldSeen = true
			if err := decoder.Decode(&decoded.Field); err != nil {
				return err
			}
		case "operator":
			if operatorSeen {
				return errors.New("duplicate mail rule condition operator")
			}
			operatorSeen = true
			if err := decoder.Decode(&decoded.Operator); err != nil {
				return err
			}
		case "value":
			if valueSeen {
				return errors.New("duplicate mail rule condition value")
			}
			valueSeen = true
			var raw json.RawMessage
			if err := decoder.Decode(&raw); err != nil {
				return err
			}
			if bytes.Equal(raw, []byte("null")) {
				return errors.New("mail rule condition value must not be null")
			}
			var stringValue string
			if err := json.Unmarshal(raw, &stringValue); err == nil {
				decoded.StringValue = &stringValue
				break
			}
			var booleanValue bool
			if err := json.Unmarshal(raw, &booleanValue); err != nil {
				return errors.New("mail rule condition value must be a string or boolean")
			}
			decoded.BooleanValue = &booleanValue
		default:
			return fmt.Errorf("unknown mail rule condition field %q", name)
		}
	}
	if err := closeJSONObject(decoder); err != nil {
		return err
	}
	if !fieldSeen || !operatorSeen || !valueSeen {
		return errors.New("mail rule condition requires field, operator and value")
	}
	if err := decoded.Validate(); err != nil {
		return err
	}
	*condition = decoded
	return nil
}

type MailRuleActions []MailRuleAction

func (actions MailRuleActions) Validate() error {
	if len(actions) < 1 || len(actions) > 10 {
		return errors.New("mail rule actions must contain 1..10 entries")
	}
	moves := 0
	markReads := 0
	labels := make(map[uuid.UUID]struct{}, len(actions))
	for index, action := range actions {
		if err := action.Validate(); err != nil {
			return fmt.Errorf("mail rule action %d: %w", index, err)
		}
		switch action.Type {
		case "move":
			moves++
		case "mark_read":
			markReads++
		case "add_label":
			if _, found := labels[*action.TargetID]; found {
				return errors.New("mail rule add_label targets must be distinct")
			}
			labels[*action.TargetID] = struct{}{}
		}
	}
	if moves > 1 || markReads > 1 {
		return errors.New("mail rule allows at most one move and mark_read action")
	}
	return nil
}

type MailRuleAction struct {
	Type     string
	TargetID *uuid.UUID
}

func (action MailRuleAction) Validate() error {
	switch action.Type {
	case "move", "add_label":
		if action.TargetID == nil {
			return errors.New("mail rule target action requires target_id")
		}
	case "mark_read":
		if action.TargetID != nil {
			return errors.New("mail rule mark_read forbids target_id")
		}
	default:
		return errors.New("invalid mail rule action type")
	}
	return nil
}

func (action MailRuleAction) MarshalJSON() ([]byte, error) {
	if err := action.Validate(); err != nil {
		return nil, err
	}
	if action.TargetID == nil {
		return json.Marshal(struct {
			Type string `json:"type"`
		}{Type: action.Type})
	}
	return json.Marshal(struct {
		Type     string     `json:"type"`
		TargetID *uuid.UUID `json:"target_id"`
	}{Type: action.Type, TargetID: action.TargetID})
}

func (action *MailRuleAction) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := requireJSONObject(decoder); err != nil {
		return err
	}
	var decoded MailRuleAction
	var typeSeen, targetSeen bool
	for decoder.More() {
		name, err := nextJSONObjectName(decoder)
		if err != nil {
			return err
		}
		switch name {
		case "type":
			if typeSeen {
				return errors.New("duplicate mail rule action type")
			}
			typeSeen = true
			if err := decoder.Decode(&decoded.Type); err != nil {
				return err
			}
		case "target_id":
			if targetSeen {
				return errors.New("duplicate mail rule action target_id")
			}
			targetSeen = true
			var raw json.RawMessage
			if err := decoder.Decode(&raw); err != nil {
				return err
			}
			if bytes.Equal(raw, []byte("null")) {
				return errors.New("mail rule action target_id must not be null")
			}
			var target uuid.UUID
			if err := json.Unmarshal(raw, &target); err != nil {
				return errors.New("mail rule action target_id must be a UUID")
			}
			decoded.TargetID = &target
		default:
			return fmt.Errorf("unknown mail rule action field %q", name)
		}
	}
	if err := closeJSONObject(decoder); err != nil {
		return err
	}
	if !typeSeen {
		return errors.New("mail rule action requires type")
	}
	if err := decoded.Validate(); err != nil {
		return err
	}
	*action = decoded
	return nil
}

func ValidateMailRuleJSON(conditions MailRuleConditions, actions MailRuleActions) error {
	if err := conditions.Validate(); err != nil {
		return err
	}
	if err := actions.Validate(); err != nil {
		return err
	}
	encodedConditions, err := json.Marshal(conditions)
	if err != nil {
		return err
	}
	encodedActions, err := json.Marshal(actions)
	if err != nil {
		return err
	}
	if len(encodedConditions)+len(encodedActions) > maxMailRuleJSONBytes {
		return errors.New("mail rule JSON exceeds 65536 bytes")
	}
	return nil
}

func validateMailRuleName(value string) error {
	return validateMailRuleText(value, 120, 480)
}

func validateMailRuleConditionString(value string) error {
	return validateMailRuleText(value, 500, 2_000)
}

func validateMailRuleText(value string, maxRunes, maxBytes int) error {
	if !utf8.ValidString(value) || len(value) == 0 || len(value) > maxBytes || utf8.RuneCountInString(value) > maxRunes {
		return errors.New("invalid mail rule text")
	}
	for _, character := range value {
		if unicode.In(character, unicode.Cc, unicode.Cf, unicode.Cs, unicode.Co) || !unicode.IsGraphic(character) {
			return errors.New("invalid mail rule text")
		}
	}
	return nil
}

func requireJSONObject(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return errors.New("mail rule JSON value must be an object")
	}
	return nil
}

func nextJSONObjectName(decoder *json.Decoder) (string, error) {
	token, err := decoder.Token()
	if err != nil {
		return "", err
	}
	name, ok := token.(string)
	if !ok {
		return "", errors.New("mail rule JSON object key must be a string")
	}
	return name, nil
}

func closeJSONObject(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '}' {
		return errors.New("mail rule JSON value must end with an object")
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("mail rule JSON value has trailing data")
		}
		return err
	}
	return nil
}
