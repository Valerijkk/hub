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
		mappings []imex.FieldMapping
		expected map[string]int
	}{
		{
			"basic mapping",
			[]imex.FieldMapping{
				{SourceColumn: "Summary", TargetField: "title"},
				{SourceColumn: "Type", TargetField: "type"},
			},
			map[string]int{"title": 1, "type": 2},
		},
		{
			"missing source column",
			[]imex.FieldMapping{
				{SourceColumn: "NonExistent", TargetField: "title"},
			},
			map[string]int{},
		},
		{
			"empty mappings",
			[]imex.FieldMapping{},
			map[string]int{},
		},
		{
			"all columns mapped",
			[]imex.FieldMapping{
				{SourceColumn: "Name", TargetField: "title"},
				{SourceColumn: "Type", TargetField: "type"},
				{SourceColumn: "Priority", TargetField: "priority"},
			},
			map[string]int{"title": 0, "type": 2, "priority": 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := imex.ExportBuildFieldMap(tt.mappings, headers)
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
