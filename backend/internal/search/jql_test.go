package search

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseJQLSimple проверяет парсинг простого условия.
func TestParseJQLSimple(t *testing.T) {
	q, err := ParseJQL(`priority = high`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 1)

	assert.Equal(t, "priority", q.Conditions[0].Field)
	assert.Equal(t, "=", q.Conditions[0].Operator)
	assert.Equal(t, "high", q.Conditions[0].Value)
}

// TestParseJQLMultipleConditions проверяет парсинг нескольких условий с AND/OR.
func TestParseJQLMultipleConditions(t *testing.T) {
	q, err := ParseJQL(`priority = high AND type = bug`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 2)

	assert.Equal(t, "priority", q.Conditions[0].Field)
	assert.Equal(t, "=", q.Conditions[0].Operator)
	assert.Equal(t, "high", q.Conditions[0].Value)
	assert.Equal(t, "", q.Conditions[0].Logic)

	assert.Equal(t, "type", q.Conditions[1].Field)
	assert.Equal(t, "=", q.Conditions[1].Operator)
	assert.Equal(t, "bug", q.Conditions[1].Value)
	assert.Equal(t, "AND", q.Conditions[1].Logic)
}

// TestParseJQLIN проверяет оператор IN.
func TestParseJQLIN(t *testing.T) {
	q, err := ParseJQL(`priority IN (high, critical)`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 1)

	assert.Equal(t, "priority", q.Conditions[0].Field)
	assert.Equal(t, "IN", q.Conditions[0].Operator)
	assert.Equal(t, []string{"high", "critical"}, q.Conditions[0].Values)
}

// TestParseJQLNotEqual проверяет оператор !=.
func TestParseJQLNotEqual(t *testing.T) {
	q, err := ParseJQL(`status != done`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 1)

	assert.Equal(t, "!=", q.Conditions[0].Operator)
	assert.Equal(t, "done", q.Conditions[0].Value)
}

// TestParseJQLOrderBy проверяет ORDER BY.
func TestParseJQLOrderBy(t *testing.T) {
	q, err := ParseJQL(`priority = high ORDER BY created DESC`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 1)
	require.NotNil(t, q.OrderBy)

	assert.Equal(t, "created", q.OrderBy.Field)
	assert.Equal(t, "DESC", q.OrderBy.Direction)
}

// TestParseJQLCurrentUser проверяет функцию currentUser().
func TestParseJQLCurrentUser(t *testing.T) {
	q, err := ParseJQL(`assignee = currentUser()`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 1)

	assert.Equal(t, "assignee", q.Conditions[0].Field)
	assert.Equal(t, "currentUser()", q.Conditions[0].Value)
}

// TestParseJQLQuotedValue проверяет значение в кавычках.
func TestParseJQLQuotedValue(t *testing.T) {
	q, err := ParseJQL(`status = "in progress"`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 1)

	assert.Equal(t, "in progress", q.Conditions[0].Value)
}

// TestParseJQLNotIN проверяет оператор NOT IN.
func TestParseJQLNotIN(t *testing.T) {
	q, err := ParseJQL(`priority NOT IN (low, trivial)`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 1)

	assert.Equal(t, "NOT IN", q.Conditions[0].Operator)
	assert.Equal(t, []string{"low", "trivial"}, q.Conditions[0].Values)
}

// TestParseJQLComparisonOps проверяет операторы сравнения.
func TestParseJQLComparisonOps(t *testing.T) {
	tests := []struct {
		input string
		op    string
	}{
		{"priority >= high", ">="},
		{"priority <= low", "<="},
		{"priority > medium", ">"},
		{"priority < high", "<"},
	}

	for _, tt := range tests {
		q, err := ParseJQL(tt.input)
		require.NoError(t, err, "input: %s", tt.input)
		assert.Equal(t, tt.op, q.Conditions[0].Operator)
	}
}

// TestParseJQLEmpty проверяет ошибку на пустой запрос.
func TestParseJQLEmpty(t *testing.T) {
	_, err := ParseJQL("")
	require.Error(t, err)
}

// TestToEntPredicates проверяет конвертацию в предикаты Ent.
func TestToEntPredicates(t *testing.T) {
	q, err := ParseJQL(`priority = high AND type = bug`)
	require.NoError(t, err)

	preds := ToEntPredicates(q, uuid.New())
	assert.Len(t, preds, 2)
}

// TestToEntPredicatesIN проверяет предикат для IN.
func TestToEntPredicatesIN(t *testing.T) {
	q, err := ParseJQL(`priority IN (high, critical)`)
	require.NoError(t, err)

	preds := ToEntPredicates(q, uuid.New())
	assert.Len(t, preds, 1)
}

// TestParseJQLOR проверяет OR логику.
func TestParseJQLOR(t *testing.T) {
	q, err := ParseJQL(`priority = high OR priority = critical`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 2)
	assert.Equal(t, "OR", q.Conditions[1].Logic)
}
