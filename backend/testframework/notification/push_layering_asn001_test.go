//go:build hubtest

package notification_test

// TASK-188 / ASN-001: PushSender без *ent.Client (layering).

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPushSender_NoEntImport_ASN001 — layering: push.go не импортирует ent.
// Mutation: вернуть `db *ent.Client` / import ent в PushSender → тест краснеет.
func TestPushSender_NoEntImport_ASN001(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	pushPath := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "notification", "push.go"))

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, pushPath, nil, parser.ImportsOnly)
	require.NoError(t, err)

	for _, imp := range f.Imports {
		path := imp.Path.Value
		assert.NotEqual(t, `"github.com/valerii/hub/backend/ent"`, path,
			"push.go must not import ent (ASN-001); use PushSubscriptionStore")
		assert.NotContains(t, path, "/ent/pushsubscription",
			"push.go must not import ent/pushsubscription (ASN-001)")
	}
}
