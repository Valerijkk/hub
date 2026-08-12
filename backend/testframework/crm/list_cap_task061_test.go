//go:build hubtest

package crm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/crm"
)

// TestEntRepository_ListContactsAndDeals_Caps_P3BE06 — N>MaxCRMList → Limit (W7-A2-013 / TASK-118).
// Mutation: снять Limit → FAIL (len > MaxCRMList).
func TestEntRepository_ListContactsAndDeals_Caps_P3BE06(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_listcap?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()
	repo := crm.NewEntRepository(client)

	owner, err := client.User.Create().
		SetEmail("crm-cap@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	capN := crm.MaxCRMList
	for i := 0; i < capN+3; i++ {
		_, err := repo.CreateContact(ctx, owner.ID, crm.CreateContactRequest{
			Name: fmt.Sprintf("C-%d", i), Email: fmt.Sprintf("c%d@test.com", i),
		})
		require.NoError(t, err)
		_, err = repo.CreateDeal(ctx, owner.ID, crm.CreateDealRequest{
			Title: fmt.Sprintf("D-%d", i),
		})
		require.NoError(t, err)
	}

	contacts, err := repo.ListContacts(ctx, &owner.ID)
	require.NoError(t, err)
	assert.Equal(t, capN, len(contacts))

	deals, err := repo.ListDeals(ctx, &owner.ID)
	require.NoError(t, err)
	assert.Equal(t, capN, len(deals))
}
