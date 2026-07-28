//go:build hubtest

package intranet_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/auditlog"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/intranet"
)

// TestCoffeePartner_NoEmailPII_Task067 — W7-A2-014 / P3-BE-15:
// partner DTO без email. Мутация: вернуть Email в toCoffeePartner → JSON содержит "email".
func TestCoffeePartner_NoEmailPII_Task067(t *testing.T) {
	u := &struct {
		ID          uuid.UUID
		Email       string
		Name        string
		DisplayName string
		Department  string
	}{
		ID: uuid.New(), Email: "secret@hub.local", Name: "Alice", Department: "Eng",
	}
	// Используем Export через реальный ent.User-подобный путь: toCoffeePartner из export.
	client := enttest.Open(t, "sqlite3", "file:coffee_pii_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	entUser, err := client.User.Create().
		SetEmail(u.Email).
		SetName(u.Name).
		SetDisplayName("Alice D").
		SetDepartment(u.Department).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	partner := intranet.ExportToCoffeePartner(entUser)
	raw, err := json.Marshal(partner)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), `"email"`, "coffee partner must not leak email PII")
	assert.NotContains(t, string(raw), "secret@hub.local")
	assert.Equal(t, "Alice D", partner.DisplayName)
}

// TestDeleteNews_WritesAudit_Task067 — W7-A2-016 / P3-BE-15 / FR-INTR-07:
// soft-delete news → audit row (actor, post_id).
// Мутация: убрать Create AuditLog из SoftDeleteNewsPost → Count=0.
func TestDeleteNews_WritesAudit_Task067(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:news_audit_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("news-admin-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	repo := intranet.NewEntRepository(client)
	svc := intranet.NewService(repo, nil, nil)

	post, err := svc.CreateNews(ctx, admin.ID, intranet.CreateNewsRequest{
		Title: "Audit me", Body: "body", Published: true,
	})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteNews(ctx, admin.ID, post.ID))

	n, err := client.AuditLog.Query().
		Where(
			auditlog.ActionEQ("delete"),
			auditlog.EntityTypeEQ("news_post"),
			auditlog.EntityIDEQ(post.ID.String()),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n, "soft-delete news must write domain audit row")

	row, err := client.AuditLog.Query().
		Where(auditlog.EntityIDEQ(post.ID.String())).
		Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, row.UserID)
	assert.Equal(t, admin.ID, *row.UserID)
	assert.Equal(t, post.ID.String(), row.Details["post_id"])
	assert.Equal(t, admin.ID.String(), row.Details["actor_id"])
}
