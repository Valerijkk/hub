package messenger_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	phase151LegacyPromoteToIssuePath = "/messages/{messageId}/promote-to-issue"
	phase151ScopedPromoteToIssuePath = "/messenger/channels/{channelId}/messages/{messageId}/promote-to-issue"
	phase151UnreadPath               = "/messenger/unread"
	phase151UnreadResponseRef        = "#/definitions/internal_messenger.UnreadResponse"
)

type phase151OpenAPISpec struct {
	Paths       map[string]map[string]phase151OpenAPIOperation `json:"paths" yaml:"paths"`
	Definitions map[string]phase151OpenAPISchema               `json:"definitions" yaml:"definitions"`
}

type phase151OpenAPIOperation struct {
	Parameters []phase151OpenAPIParameter         `json:"parameters" yaml:"parameters"`
	Responses  map[string]phase151OpenAPIResponse `json:"responses" yaml:"responses"`
}

type phase151OpenAPIParameter struct {
	Name     string                `json:"name" yaml:"name"`
	In       string                `json:"in" yaml:"in"`
	Required bool                  `json:"required" yaml:"required"`
	Type     string                `json:"type" yaml:"type"`
	Schema   phase151OpenAPISchema `json:"schema" yaml:"schema"`
}

type phase151OpenAPIResponse struct {
	Schema phase151OpenAPISchema `json:"schema" yaml:"schema"`
}

type phase151OpenAPISchema struct {
	Ref                  string                            `json:"$ref" yaml:"$ref"`
	Type                 string                            `json:"type" yaml:"type"`
	Format               string                            `json:"format" yaml:"format"`
	Properties           map[string]phase151OpenAPISchema `json:"properties" yaml:"properties"`
	AdditionalProperties any                               `json:"additionalProperties" yaml:"additionalProperties"`
}

func TestPhase151OpenAPIGeneratedSourcesShareMessengerContract(t *testing.T) {
	t.Parallel()

	sources := []struct {
		name string
		spec phase151OpenAPISpec
	}{
		{name: "swagger.json", spec: phase151ReadOpenAPIJSON(t, "swagger.json")},
		{name: "docs.go", spec: phase151ReadDocsGoOpenAPI(t)},
		{name: "swagger.yaml", spec: phase151ReadOpenAPIYAML(t)},
	}

	for _, source := range sources {
		source := source
		t.Run(source.name, func(t *testing.T) {
			phase151AssertOpenAPIMessengerContract(t, source.name, source.spec)
		})
	}
}

func phase151AssertOpenAPIMessengerContract(t *testing.T, sourceName string, spec phase151OpenAPISpec) {
	t.Helper()

	if _, found := spec.Paths[phase151LegacyPromoteToIssuePath]; found {
		t.Fatalf("%s legacy promote-to-issue path %q must be absent", sourceName, phase151LegacyPromoteToIssuePath)
	}

	scopedPath, found := spec.Paths[phase151ScopedPromoteToIssuePath]
	if !found {
		t.Fatalf("%s scoped promote-to-issue path %q is absent", sourceName, phase151ScopedPromoteToIssuePath)
	}
	scopedPost, found := scopedPath["post"]
	if !found {
		t.Fatalf("%s scoped promote-to-issue path %q has no POST operation", sourceName, phase151ScopedPromoteToIssuePath)
	}
	if len(scopedPost.Parameters) != 3 {
		t.Fatalf("%s scoped promote POST parameter count = %d, want 3", sourceName, len(scopedPost.Parameters))
	}

	parametersExpected := map[[2]string]struct{}{
		{"channelId", "path"}: {},
		{"messageId", "path"}: {},
		{"body", "body"}:      {},
	}
	parameters := make(map[[2]string]phase151OpenAPIParameter, len(scopedPost.Parameters))
	for _, parameter := range scopedPost.Parameters {
		key := [2]string{parameter.Name, parameter.In}
		if _, expected := parametersExpected[key]; !expected {
			t.Fatalf("%s scoped promote POST has unexpected parameter %#v", sourceName, parameter)
		}
		if _, duplicate := parameters[key]; duplicate {
			t.Fatalf("%s scoped promote POST has duplicate parameter %#v", sourceName, parameter)
		}
		parameters[key] = parameter
	}

	channelID, found := parameters[[2]string{"channelId", "path"}]
	if !found {
		t.Fatalf("%s scoped promote POST channelId path parameter is absent", sourceName)
	}
	if channelID.Name != "channelId" || channelID.In != "path" || !channelID.Required || channelID.Type != "string" || channelID.Schema.Ref != "" {
		t.Fatalf("%s scoped promote POST channelId parameter = %#v, want required string path parameter", sourceName, channelID)
	}
	messageID, found := parameters[[2]string{"messageId", "path"}]
	if !found {
		t.Fatalf("%s scoped promote POST messageId path parameter is absent", sourceName)
	}
	if messageID.Name != "messageId" || messageID.In != "path" || !messageID.Required || messageID.Type != "string" || messageID.Schema.Ref != "" {
		t.Fatalf("%s scoped promote POST messageId parameter = %#v, want required string path parameter", sourceName, messageID)
	}
	body, found := parameters[[2]string{"body", "body"}]
	if !found {
		t.Fatalf("%s scoped promote POST body parameter is absent", sourceName)
	}
	if body.Name != "body" || body.In != "body" || !body.Required || body.Type != "" || body.Schema.Ref != "#/definitions/internal_messenger.PromoteToIssueRequest" {
		t.Fatalf("%s scoped promote POST body parameter = %#v, want required PromoteToIssueRequest body", sourceName, body)
	}
	if _, found := scopedPost.Responses["201"]; !found {
		t.Fatalf("%s scoped promote POST has no 201 response", sourceName)
	}

	unreadPath, found := spec.Paths[phase151UnreadPath]
	if !found {
		t.Fatalf("%s unread path %q is absent", sourceName, phase151UnreadPath)
	}
	unreadGet, found := unreadPath["get"]
	if !found {
		t.Fatalf("%s unread path %q has no GET operation", sourceName, phase151UnreadPath)
	}
	unreadResponse, found := unreadGet.Responses["200"]
	if !found {
		t.Fatalf("%s unread GET has no 200 response", sourceName)
	}
	if unreadResponse.Schema.Ref != phase151UnreadResponseRef {
		t.Fatalf("%s unread GET 200 schema ref = %q, want %q", sourceName, unreadResponse.Schema.Ref, phase151UnreadResponseRef)
	}

	unreadDefinition, found := spec.Definitions["internal_messenger.UnreadResponse"]
	if !found {
		t.Fatalf("%s internal_messenger.UnreadResponse definition is absent", sourceName)
	}

	expectedMapProperties := map[string]struct{}{
		"channels": {},
		"dms":      {},
		"mentions": {},
		"threads":  {},
	}
	mapProperties := 0
	for name, property := range unreadDefinition.Properties {
		if _, isMap := phase151OpenAPIMapValueSchema(property); !isMap {
			continue
		}
		mapProperties++
		if _, expected := expectedMapProperties[name]; !expected {
			t.Fatalf("%s UnreadResponse has unexpected map-valued property %q", sourceName, name)
		}
	}
	if mapProperties != len(expectedMapProperties) {
		t.Fatalf("%s UnreadResponse map-valued property count = %d, want %d", sourceName, mapProperties, len(expectedMapProperties))
	}

	for name := range expectedMapProperties {
		property, found := unreadDefinition.Properties[name]
		if !found {
			t.Fatalf("%s UnreadResponse.%s is absent", sourceName, name)
		}
		valueSchema, isMap := phase151OpenAPIMapValueSchema(property)
		if !isMap {
			t.Fatalf("%s UnreadResponse.%s = %#v, want map-valued object", sourceName, name, property)
		}
		if valueSchema.Type != "integer" || valueSchema.Format != "int64" {
			t.Fatalf("%s UnreadResponse.%s values = %#v, want integer int64", sourceName, name, valueSchema)
		}
	}

	anyUnread, found := unreadDefinition.Properties["any_unread"]
	if !found {
		t.Fatalf("%s UnreadResponse.any_unread is absent", sourceName)
	}
	if anyUnread.Type != "boolean" {
		t.Fatalf("%s UnreadResponse.any_unread type = %q, want boolean", sourceName, anyUnread.Type)
	}
}

func phase151OpenAPIMapValueSchema(property phase151OpenAPISchema) (phase151OpenAPISchema, bool) {
	if property.Type != "object" || property.AdditionalProperties == nil {
		return phase151OpenAPISchema{}, false
	}

	payload, err := json.Marshal(property.AdditionalProperties)
	if err != nil {
		return phase151OpenAPISchema{}, false
	}

	var valueSchema phase151OpenAPISchema
	if err := json.Unmarshal(payload, &valueSchema); err != nil {
		return phase151OpenAPISchema{}, false
	}
	return valueSchema, true
}

func phase151ReadOpenAPIJSON(t *testing.T, name string) phase151OpenAPISpec {
	t.Helper()

	payload, err := os.ReadFile(phase151OpenAPIDocsPath(t, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return phase151UnmarshalOpenAPIJSON(t, name, payload)
}

func phase151ReadDocsGoOpenAPI(t *testing.T) phase151OpenAPISpec {
	t.Helper()

	source, err := os.ReadFile(phase151OpenAPIDocsPath(t, "docs.go"))
	if err != nil {
		t.Fatalf("read docs.go: %v", err)
	}

	const templatePrefix = "const docTemplate = " + string(rune(96))
	sourceText := string(source)
	if strings.Count(sourceText, templatePrefix) != 1 {
		t.Fatal("docs.go must contain exactly one docTemplate")
	}
	template := sourceText[strings.Index(sourceText, templatePrefix)+len(templatePrefix):]
	templateEnd := strings.Index(template, string(rune(96)))
	if templateEnd < 0 {
		t.Fatal("docs.go docTemplate closing delimiter is absent")
	}

	const schemesTemplate = "{{ marshal .Schemes }}"
	if strings.Count(template[:templateEnd], schemesTemplate) != 1 {
		t.Fatal("docs.go docTemplate schemes placeholder is malformed")
	}
	payload := []byte(strings.Replace(template[:templateEnd], schemesTemplate, "[]", 1))
	return phase151UnmarshalOpenAPIJSON(t, "docs.go docTemplate", payload)
}

func phase151ReadOpenAPIYAML(t *testing.T) phase151OpenAPISpec {
	t.Helper()

	payload, err := os.ReadFile(phase151OpenAPIDocsPath(t, "swagger.yaml"))
	if err != nil {
		t.Fatalf("read swagger.yaml: %v", err)
	}

	var spec phase151OpenAPISpec
	if err := yaml.Unmarshal(payload, &spec); err != nil {
		t.Fatalf("parse swagger.yaml: %v", err)
	}
	return spec
}

func phase151UnmarshalOpenAPIJSON(t *testing.T, sourceName string, payload []byte) phase151OpenAPISpec {
	t.Helper()

	var spec phase151OpenAPISpec
	if err := json.Unmarshal(payload, &spec); err != nil {
		t.Fatalf("parse %s: %v", sourceName, err)
	}
	return spec
}

func phase151OpenAPIDocsPath(t *testing.T, name string) string {
	t.Helper()

	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve phase151 OpenAPI test source path")
	}

	path := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "docs", name))
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("resolve %s from package test path: %v", name, err)
	}
	return path
}