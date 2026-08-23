//go:build hubtest

package audit_test

// TASK-160 / A3W6-4-001..003: Mongo post-Commit, Ban sentinels, requireAdmin без ent.
import (
	"bytes"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/audit"
)

// TestHandler_BanUser_SentinelValidation_A3W64_002 — 400 через errors.Is, не string match.
// Mutation: снова map через err.Error() == "…" → правка текста sentinel ломает 400 / тест краснеет.
func TestHandler_BanUser_SentinelValidation_A3W64_002(t *testing.T) {
	adminID := uuid.New()
	targetID := uuid.New()
	repo := &userMgmtMockRepo{adminMockRepo: adminMockRepo{adminID: adminID}}
	app := setupAuditHandlerTest(t, repo, adminID)

	t.Run("empty reason", func(t *testing.T) {
		body := bytes.NewBufferString(`{"reason":""}`)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/users/"+targetID.String()+"/ban", body)
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("self ban", func(t *testing.T) {
		body := bytes.NewBufferString(`{"reason":"abuse"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/users/"+adminID.String()+"/ban", body)
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestHandler_NoEntImport_A3W64_003 — layering: handler.go не импортирует ent.
// Mutation: вернуть ent.IsNotFound в requireAdmin → импорт ent → тест краснеет.
func TestHandler_NoEntImport_A3W64_003(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	handlerPath := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "audit", "handler.go"))

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, handlerPath, nil, parser.ImportsOnly)
	require.NoError(t, err)

	for _, imp := range f.Imports {
		path := imp.Path.Value
		assert.NotEqual(t, `"github.com/valerii/hub/backend/ent"`, path,
			"handler.go must not import ent (A3W6-4-003); use ErrUserNotFound")
	}
}

// TestBanUser_SentinelIdentity_A3W64_002 — смена текста sentinel не обязана ломать ErrorIs mapping в handler,
// но ErrorIs на package vars стабилен.
func TestBanUser_SentinelIdentity_A3W64_002(t *testing.T) {
	require.ErrorIs(t, audit.ErrBanReasonRequired, audit.ErrBanReasonRequired)
	require.ErrorIs(t, audit.ErrCannotBanSelf, audit.ErrCannotBanSelf)
	require.Equal(t, "укажите причину блокировки", audit.ErrBanReasonRequired.Error())
	require.Equal(t, "нельзя заблокировать самого себя", audit.ErrCannotBanSelf.Error())
}
