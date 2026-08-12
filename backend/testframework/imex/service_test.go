package imex_test
import (
	"github.com/valerii/hub/backend/internal/imex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildFieldMap(t *testing.T) {
	headers := []string{"Name", "Summary", "Type", "Priority"}

	tests := []struct {
		name     string
		headers  []string
		mappings []imex.FieldMapping
		expected map[string]int
	}{
		{
			name:    "basic mapping",
			headers: headers,
			mappings: []imex.FieldMapping{
				{SourceColumn: "Summary", TargetField: "title"},
				{SourceColumn: "Type", TargetField: "type"},
			},
			expected: map[string]int{"title": 1, "type": 2},
		},
		{
			name:    "missing source column",
			headers: headers,
			mappings: []imex.FieldMapping{
				{SourceColumn: "NonExistent", TargetField: "title"},
			},
			expected: map[string]int{},
		},
		{
			name:     "empty mappings auto-maps Type/Priority by name (A5-001)",
			headers:  headers,
			mappings: []imex.FieldMapping{},
			expected: map[string]int{"type": 2, "priority": 3},
		},
		{
			name:     "nil mappings auto-maps Title/Type/Priority export headers (A5-001)",
			headers:  []string{"Key", "Title", "Type", "Priority", "Status"},
			mappings: nil,
			expected: map[string]int{"title": 1, "type": 2, "priority": 3},
		},
		{
			name:     "empty mappings auto-maps lowercase target headers (A5-001)",
			headers:  []string{"title", "type", "priority", "extra"},
			mappings: nil,
			expected: map[string]int{"title": 0, "type": 1, "priority": 2},
		},
		{
			name:    "all columns mapped",
			headers: headers,
			mappings: []imex.FieldMapping{
				{SourceColumn: "Name", TargetField: "title"},
				{SourceColumn: "Type", TargetField: "type"},
				{SourceColumn: "Priority", TargetField: "priority"},
			},
			expected: map[string]int{"title": 0, "type": 2, "priority": 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := imex.ExportBuildFieldMap(tt.mappings, tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetField(t *testing.T) {
	record := []string{"val0", "val1", "val2", "val3"}
	fieldMap := map[string]int{"title": 1, "type": 2}

	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{"existing field", "title", "val1"},
		{"another existing field", "type", "val2"},
		{"missing field", "priority", ""},
		{"empty field name", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := imex.ExportGetField(record, fieldMap, tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetField_OutOfBounds(t *testing.T) {
	record := []string{"only-one"}
	fieldMap := map[string]int{"title": 5} // index beyond record length

	result := imex.ExportGetField(record, fieldMap, "title")
	assert.Equal(t, "", result, "should return empty string when index exceeds record length")
}
