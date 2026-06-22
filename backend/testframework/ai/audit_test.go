package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	auditmod "github.com/valerii/hub/backend/internal/audit"
)

func TestAuditRepoLogger_LogAIChat(t *testing.T) {
	dsn := "file:ai_audit_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("ai-audit-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	logger := ai.NewAuditRepoLogger(auditmod.NewRepository(client))
	providerID := uuid.New()
	require.NoError(t, logger.LogAIChat(ctx, user.ID, providerID, "gpt-4", 100, 50, 120*time.Millisecond, "127.0.0.1", "test-agent"))
}
