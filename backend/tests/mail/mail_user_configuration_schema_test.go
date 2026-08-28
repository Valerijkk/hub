package mail_test

import (
	"encoding/json"
	"strings"
	"testing"

	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	entschema "github.com/valerii/hub/backend/ent/schema"
	"github.com/google/uuid"
)

func TestMailUserConfigurationSchemas_HaveExactFieldsDefaultsAndBounds(t *testing.T) {
	settings := fieldDescriptors((entschema.MailUserSettings{}).Fields())
	assertExactNames(t, "MailUserSettings fields", fieldNames(settings), []string{
		"compose_format", "created_at", "default_identity_id", "default_signature_id", "id",
		"mail_account_id", "remote_images", "updated_at",
	})
	for _, name := range []string{"id", "mail_account_id", "default_identity_id", "default_signature_id"} {
		assertFieldType(t, name, settings[name], field.TypeUUID)
	}
	for _, name := range []string{"id", "mail_account_id", "default_identity_id"} {
		assertRequired(t, name, settings[name])
	}
	assertOptionalNillable(t, "default_signature_id", settings["default_signature_id"])
	if !settings["id"].Immutable || !settings["mail_account_id"].Immutable {
		t.Fatal("MailUserSettings id and mail_account_id must be immutable")
	}
	assertExactEnumValues(t, "remote_images", settings["remote_images"], []string{"blocked", "trusted_senders"})
	assertExactEnumValues(t, "compose_format", settings["compose_format"], []string{"rich", "plain"})
	assertExactDefaults(t, settings, map[string]any{"remote_images": "blocked", "compose_format": "rich"})
	assertVersionAndTimestamps(t, entschema.MailUserSettings{}, settings)

	signature := fieldDescriptors((entschema.MailSignature{}).Fields())
	assertExactNames(t, "MailSignature fields", fieldNames(signature), []string{
		"body_format", "created_at", "deleted_at", "id", "mail_account_id", "name", "plain_body",
		"rich_body", "updated_at",
	})
	for _, name := range []string{"id", "mail_account_id"} {
		assertFieldType(t, name, signature[name], field.TypeUUID)
		assertRequired(t, name, signature[name])
		if !signature[name].Immutable {
			t.Fatalf("MailSignature %s must be immutable", name)
		}
	}
	for _, name := range []string{"name", "plain_body"} {
		assertFieldType(t, name, signature[name], field.TypeString)
		assertRequired(t, name, signature[name])
	}
	assertFieldType(t, "body_format", signature["body_format"], field.TypeEnum)
	assertRequired(t, "body_format", signature["body_format"])
	assertOptionalNillable(t, "rich_body", signature["rich_body"])
	assertOptionalNillable(t, "deleted_at", signature["deleted_at"])
	assertExactEnumValues(t, "body_format", signature["body_format"], []string{"rich", "plain"})
	assertStringValidators(t, "name", signature["name"], []struct {
		value string
		want  bool
	}{
		{"signature", true},
		{strings.Repeat("x", 80), true},
		{"", false},
		{strings.Repeat("x", 81), false},
		{strings.Repeat("\U0001f600", 81), false},
	})
	for _, name := range []string{"plain_body", "rich_body"} {
		assertStringValidators(t, name, signature[name], []struct {
			value string
			want  bool
		}{
			{"body", true},
			{strings.Repeat("x", 100_000), true},
			{strings.Repeat("x", 100_001), false},
		})
	}
	assertVersionAndTimestamps(t, entschema.MailSignature{}, signature)

	rule := fieldDescriptors((entschema.MailRule{}).Fields())
	assertExactNames(t, "MailRule fields", fieldNames(rule), []string{
		"actions", "conditions", "created_at", "deleted_at", "enabled", "id", "mail_account_id", "name",
		"priority", "updated_at",
	})
	for _, name := range []string{"id", "mail_account_id"} {
		assertFieldType(t, name, rule[name], field.TypeUUID)
		assertRequired(t, name, rule[name])
		if !rule[name].Immutable {
			t.Fatalf("MailRule %s must be immutable", name)
		}
	}
	assertFieldType(t, "name", rule["name"], field.TypeString)
	assertRequired(t, "name", rule["name"])
	assertFieldType(t, "enabled", rule["enabled"], field.TypeBool)
	assertRequired(t, "enabled", rule["enabled"])
	assertFieldType(t, "priority", rule["priority"], field.TypeInt32)
	assertRequired(t, "priority", rule["priority"])
	assertInt32Validators(t, "priority", rule["priority"], []struct {
		value int32
		want  bool
	}{{-1, false}, {0, true}, {10_000, true}, {10_001, false}})
	for _, name := range []string{"conditions", "actions"} {
		assertFieldType(t, name, rule[name], field.TypeJSON)
		assertRequired(t, name, rule[name])
	}
	assertOptionalNillable(t, "deleted_at", rule["deleted_at"])
	assertVersionAndTimestamps(t, entschema.MailRule{}, rule)

	contact := fieldDescriptors((entschema.RecentContact{}).Fields())
	assertExactNames(t, "RecentContact fields", fieldNames(contact), []string{
		"created_at", "display_name", "id", "last_used_at", "mail_account_id", "normalized_email", "updated_at",
		"use_count", "user_id",
	})
	for _, name := range []string{"id", "mail_account_id", "user_id"} {
		assertFieldType(t, name, contact[name], field.TypeUUID)
	}
	for _, name := range []string{"id", "mail_account_id", "normalized_email", "last_used_at", "use_count"} {
		assertRequired(t, name, contact[name])
	}
	if !contact["id"].Immutable || !contact["mail_account_id"].Immutable {
		t.Fatal("RecentContact id and mail_account_id must be immutable")
	}
	for _, name := range []string{"display_name", "user_id"} {
		assertOptionalNillable(t, name, contact[name])
	}
	assertStringValidators(t, "normalized_email", contact["normalized_email"], []struct {
		value string
		want  bool
	}{
		{"contact@example.test", true},
		{"Contact@example.test", false},
		{"contact@Example.test", false},
		{"contact@", false},
	})
	assertStringValidators(t, "display_name", contact["display_name"], []struct {
		value string
		want  bool
	}{
		{"Contact", true},
		{"  Contact  ", true},
		{strings.Repeat("\U0001f600", 160), true},
		{strings.Repeat("\U0001f600", 161), false},
		{strings.Repeat("x", 641), false},
		{"name\r\nBcc: injected", false},
	})
	assertFieldType(t, "use_count", contact["use_count"], field.TypeInt32)
	assertInt32Validators(t, "use_count", contact["use_count"], []struct {
		value int32
		want  bool
	}{{-1, false}, {0, true}, {2_147_483_647, true}})
	assertTimestampsWithoutMixin(t, contact)
}

func TestMailUserConfigurationSchemas_HaveExactChecksIndexesAndEdges(t *testing.T) {
	assertExactChecks(t, (entschema.MailSignature{}).Annotations(), map[string]string{
		"mail_signatures_body_format":       "((body_format = 'rich' AND rich_body IS NOT NULL) OR (body_format = 'plain' AND rich_body IS NULL))",
		"mail_signatures_name_bounds":       "(char_length(name) BETWEEN 1 AND 80 AND octet_length(name) BETWEEN 1 AND 320)",
		"mail_signatures_plain_body_length": "(char_length(plain_body) <= 100000 AND octet_length(plain_body) <= 400000)",
		"mail_signatures_rich_body_length":  "(rich_body IS NULL OR (char_length(rich_body) <= 100000 AND octet_length(rich_body) <= 400000))",
	})
	assertExactChecks(t, (entschema.MailRule{}).Annotations(), map[string]string{
		"mail_rules_json_shape": "jsonb_typeof(conditions) = 'array' AND jsonb_typeof(actions) = 'array' AND jsonb_array_length(conditions) BETWEEN 1 AND 20 AND jsonb_array_length(actions) BETWEEN 1 AND 10 AND octet_length(conditions::text) + octet_length(actions::text) <= 65536",
		"mail_rules_name_bounds": "(char_length(name) BETWEEN 1 AND 120 AND octet_length(name) BETWEEN 1 AND 480)",
		"mail_rules_priority":    "priority BETWEEN 0 AND 10000",
	})
	assertExactChecks(t, (entschema.RecentContact{}).Annotations(), map[string]string{
		"recent_contacts_display_name": "display_name IS NULL OR (char_length(display_name) <= 160 AND octet_length(display_name) <= 640)",
		"recent_contacts_email_bounds": "(char_length(normalized_email) BETWEEN 1 AND 320 AND octet_length(normalized_email) BETWEEN 1 AND 320)",
		"recent_contacts_use_count":     "use_count BETWEEN 0 AND 2147483647",
	})

	assertExactIndexes(t, (entschema.MailUserSettings{}).Indexes(), []wantedIndex{{fields: []string{"mail_account_id"}, unique: true}})
	assertExactIndexes(t, (entschema.MailSignature{}).Indexes(), []wantedIndex{})
	assertExactIndexes(t, (entschema.MailRule{}).Indexes(), []wantedIndex{{fields: []string{"mail_account_id", "enabled", "priority", "id"}, where: "deleted_at IS NULL"}})
	assertExactIndexes(t, (entschema.RecentContact{}).Indexes(), []wantedIndex{
		{fields: []string{"mail_account_id", "normalized_email"}, unique: true},
		{fields: []string{"mail_account_id", "last_used_at", "id"}, where: "user_id IS NULL"},
	})

	settingsEdges := edgeDescriptors((entschema.MailUserSettings{}).Edges())
	assertExactNames(t, "MailUserSettings edges", edgeNames(settingsEdges), []string{"account", "default_identity", "default_signature"})
	assertConfigurationRequiredEdge(t, "MailUserSettings.account", settingsEdges["account"], "mail_account_id", "user_settings", entsql.Restrict)
	assertConfigurationRequiredEdge(t, "MailUserSettings.default_identity", settingsEdges["default_identity"], "default_identity_id", "default_for_settings", entsql.Restrict)
	assertConfigurationOptionalEdge(t, "MailUserSettings.default_signature", settingsEdges["default_signature"], "default_signature_id", "default_for_settings", entsql.Restrict)

	signatureEdges := edgeDescriptors((entschema.MailSignature{}).Edges())
	assertExactNames(t, "MailSignature edges", edgeNames(signatureEdges), []string{"account", "default_for_settings"})
	assertConfigurationRequiredEdge(t, "MailSignature.account", signatureEdges["account"], "mail_account_id", "signatures", entsql.Restrict)
	assertConfigurationPlainUniqueEdge(t, "MailSignature.default_for_settings", signatureEdges["default_for_settings"])

	ruleEdges := edgeDescriptors((entschema.MailRule{}).Edges())
	assertExactNames(t, "MailRule edges", edgeNames(ruleEdges), []string{"account"})
	assertConfigurationRequiredEdge(t, "MailRule.account", ruleEdges["account"], "mail_account_id", "rules", entsql.Restrict)

	contactEdges := edgeDescriptors((entschema.RecentContact{}).Edges())
	assertExactNames(t, "RecentContact edges", edgeNames(contactEdges), []string{"account", "user"})
	assertConfigurationRequiredEdge(t, "RecentContact.account", contactEdges["account"], "mail_account_id", "recent_contacts", entsql.Cascade)
	assertConfigurationOptionalEdge(t, "RecentContact.user", contactEdges["user"], "user_id", "recent_contacts", entsql.SetNull)
}

func TestMailRuleJSON_IsConcreteStrictAndBounded(t *testing.T) {
	var condition entschema.MailRuleCondition
	if err := json.Unmarshal([]byte(`{"field":"from","operator":"equals","value":"sender@example.test"}`), &condition); err != nil {
		t.Fatalf("unmarshal valid string condition: %v", err)
	}
	if condition.StringValue == nil || *condition.StringValue != "sender@example.test" || condition.BooleanValue != nil {
		t.Fatalf("condition value = %#v, want concrete string value", condition)
	}
	if err := json.Unmarshal([]byte(`{"field":"has_attachment","operator":"is","value":true}`), &condition); err != nil {
		t.Fatalf("unmarshal valid boolean condition: %v", err)
	}
	if condition.BooleanValue == nil || !*condition.BooleanValue || condition.StringValue != nil {
		t.Fatalf("condition value = %#v, want concrete boolean value", condition)
	}
	for _, raw := range []string{
		`{"field":"from","operator":"equals","value":"x","extra":true}`,
		`{"field":"has_attachment","operator":"equals","value":true}`,
		`{"field":"subject","operator":"contains","value":true}`,
		`{"field":"from","operator":"is","value":"x"}`,
	} {
		if err := json.Unmarshal([]byte(raw), &condition); err == nil {
			t.Fatalf("condition %s must be rejected", raw)
		}
	}

	target := uuid.New()
	var action entschema.MailRuleAction
	if err := json.Unmarshal([]byte(`{"type":"move","target_id":"`+target.String()+`"}`), &action); err != nil {
		t.Fatalf("unmarshal valid move action: %v", err)
	}
	if action.TargetID == nil || *action.TargetID != target {
		t.Fatalf("move target = %#v, want scalar UUID %s", action.TargetID, target)
	}
	encodedAction, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("marshal move action: %v", err)
	}
	if string(encodedAction) != `{"type":"move","target_id":"`+target.String()+`"}` {
		t.Fatalf("move JSON = %s, want scalar UUID target", encodedAction)
	}
	for _, raw := range []string{
		`{"type":"move"}`,
		`{"type":"mark_read","target_id":"` + target.String() + `"}`,
		`{"type":"add_label","target_id":null}`,
		`{"type":"mark_read","extra":true}`,
	} {
		if err := json.Unmarshal([]byte(raw), &action); err == nil {
			t.Fatalf("action %s must be rejected", raw)
		}
	}

	text := "<>&"
	conditions := entschema.MailRuleConditions{{Field: "subject", Operator: "contains", StringValue: &text}}
	actions := entschema.MailRuleActions{{Type: "mark_read"}}
	if err := entschema.ValidateMailRuleJSON(conditions, actions); err != nil {
		t.Fatalf("validate concrete JSON: %v", err)
	}
	encodedConditions, err := json.Marshal(conditions)
	if err != nil {
		t.Fatalf("marshal conditions: %v", err)
	}
	if !strings.Contains(string(encodedConditions), `\u003c\u003e\u0026`) {
		t.Fatalf("conditions must retain encoding/json default HTML escaping: %s", encodedConditions)
	}
	tooManyConditions := make(entschema.MailRuleConditions, 21)
	for index := range tooManyConditions {
		value := "x"
		tooManyConditions[index] = entschema.MailRuleCondition{Field: "subject", Operator: "contains", StringValue: &value}
	}
	if err := tooManyConditions.Validate(); err == nil {
		t.Fatal("21 rule conditions must be rejected")
	}
	tooManyActions := make(entschema.MailRuleActions, 11)
	for index := range tooManyActions {
		id := uuid.New()
		tooManyActions[index] = entschema.MailRuleAction{Type: "add_label", TargetID: &id}
	}
	if err := tooManyActions.Validate(); err == nil {
		t.Fatal("11 rule actions must be rejected")
	}
}

func assertConfigurationRequiredEdge(t *testing.T, name string, descriptor *edge.Descriptor, fieldName, ref string, onDelete entsql.ReferenceOption) {
	t.Helper()
	if descriptor == nil || !descriptor.Inverse || !descriptor.Unique || !descriptor.Required || descriptor.Field != fieldName || descriptor.RefName != ref {
		t.Fatalf("%s = %#v, want inverse required unique field %q ref %q", name, descriptor, fieldName, ref)
	}
	assertOnDelete(t, name, descriptor.Annotations, onDelete)
}

func assertConfigurationOptionalEdge(t *testing.T, name string, descriptor *edge.Descriptor, fieldName, ref string, onDelete entsql.ReferenceOption) {
	t.Helper()
	if descriptor == nil || !descriptor.Inverse || !descriptor.Unique || descriptor.Required || descriptor.Field != fieldName || descriptor.RefName != ref {
		t.Fatalf("%s = %#v, want inverse optional unique field %q ref %q", name, descriptor, fieldName, ref)
	}
	assertOnDelete(t, name, descriptor.Annotations, onDelete)
}

func assertConfigurationPlainUniqueEdge(t *testing.T, name string, descriptor *edge.Descriptor) {
	t.Helper()
	if descriptor == nil || descriptor.Inverse || !descriptor.Unique || descriptor.Required || descriptor.RefName != "" {
		t.Fatalf("%s = %#v, want plain unique edge", name, descriptor)
	}
}
