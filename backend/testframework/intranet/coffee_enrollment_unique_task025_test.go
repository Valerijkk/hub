//go:build hubtest

package intranet_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

// TestCoffeeEnrollment_UniquePerUser_TASK025 — двойной Create на одного user → constraint.
// Mutation: убрать index.Edges("user").Unique() → второй Create проходит → FAIL.
func TestCoffeeEnrollment_UniquePerUser_TASK025(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:coffee_unique_t025?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("coffee-unique@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.CoffeeEnrollment.Create().
		SetUserID(u.ID).
		SetIsActive(true).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.CoffeeEnrollment.Create().
		SetUserID(u.ID).
		SetIsActive(false).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err), "second enroll must hit unique on user: %v", err)
}
