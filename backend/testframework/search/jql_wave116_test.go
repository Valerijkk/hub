package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJQL_OrderWithoutBy(t *testing.T) {
	_, err := search.ParseJQL(`status = todo ORDER created DESC`)
	require.Error(t, err)
}

func TestDateFieldPredicate_UnknownField(t *testing.T) {
	assert.Nil(t, search.ExportDateFieldPredicate("2024-06-01", "=", "deleted"))
}

func TestDateFieldPredicate_UpdatedLessThan(t *testing.T) {
	pred := search.ExportDateFieldPredicate("2024-06-01", "<", "updated")
	require.NotNil(t, pred)
}

func TestDateFieldPredicate_CreatedGreaterEqual(t *testing.T) {
	pred := search.ExportDateFieldPredicate("2024-06-01", ">=", "created")
	require.NotNil(t, pred)
}

func TestParseJQL_UnclosedString(t *testing.T) {
	_, err := search.ParseJQL(`status = "open`)
	require.Error(t, err)
}
