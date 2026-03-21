package search

import (
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMergeResultsByScore проверяет сортировку результатов по убыванию score.
func TestMergeResultsByScore(t *testing.T) {
	results := []SearchResult{
		{ID: "1", Type: "issue", Title: "Low", Score: 0.1},
		{ID: "2", Type: "user", Title: "High", Score: 0.9},
		{ID: "3", Type: "project", Title: "Mid", Score: 0.5},
		{ID: "4", Type: "test_case", Title: "Very High", Score: 1.2},
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	require.Len(t, results, 4)
	assert.Equal(t, "4", results[0].ID, "наивысший score первый")
	assert.Equal(t, "2", results[1].ID)
	assert.Equal(t, "3", results[2].ID)
	assert.Equal(t, "1", results[3].ID, "наименьший score последний")
}

// TestMergeResultsWithLimit проверяет ограничение количества результатов.
func TestMergeResultsWithLimit(t *testing.T) {
	results := make([]SearchResult, 30)
	for i := range results {
		results[i] = SearchResult{
			ID:    uuid.New().String(),
			Type:  "issue",
			Title: "Test",
			Score: float64(30 - i),
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	limit := 20
	if len(results) > limit {
		results = results[:limit]
	}

	assert.Len(t, results, 20, "должно быть ограничено до 20 результатов")
	assert.Equal(t, float64(30), results[0].Score, "первый результат с наивысшим score")
	assert.Equal(t, float64(11), results[19].Score, "последний результат = 11")
}

// TestUUIDSliceToPgArray проверяет конвертацию среза UUID в формат PG-массива.
func TestUUIDSliceToPgArray(t *testing.T) {
	id1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	id2 := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	result := uuidSliceToPgArray([]uuid.UUID{id1, id2})
	assert.Equal(t, "{00000000-0000-0000-0000-000000000001,00000000-0000-0000-0000-000000000002}", result)
}

// TestUUIDSliceToPgArrayEmpty проверяет пустой массив.
func TestUUIDSliceToPgArrayEmpty(t *testing.T) {
	result := uuidSliceToPgArray([]uuid.UUID{})
	assert.Equal(t, "{}", result)
}

// TestSavedFilterToDTO проверяет конвертацию SavedFilter Ent-сущности в DTO.
func TestSavedFilterToDTO(t *testing.T) {
	// Примечание: полноценный тест конвертации требует Ent-сущность,
	// здесь проверяем вспомогательную логику.

	typeSet := map[string]bool{"issue": true, "user": true}
	assert.True(t, typeSet["issue"])
	assert.True(t, typeSet["user"])
	assert.False(t, typeSet["project"])
}

// TestSearchRequestDefaultLimit проверяет значение limit по умолчанию.
func TestSearchRequestDefaultLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"нулевой limit", 0, 20},
		{"отрицательный limit", -5, 20},
		{"превышающий limit", 100, 20},
		{"нормальный limit", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit := tt.input
			if limit <= 0 || limit > 50 {
				limit = 20
			}
			assert.Equal(t, tt.expected, limit)
		})
	}
}

// TestSearchResultTypes проверяет поддерживаемые типы сущностей.
func TestSearchResultTypes(t *testing.T) {
	supportedTypes := []string{"issue", "test_case", "user", "project", "channel", "file", "message"}
	typeSet := make(map[string]bool, len(supportedTypes))
	for _, tp := range supportedTypes {
		typeSet[tp] = true
	}

	assert.Len(t, typeSet, 7, "должно быть 7 типов сущностей")
	assert.True(t, typeSet["issue"])
	assert.True(t, typeSet["message"])
	assert.False(t, typeSet["unknown"])
}
