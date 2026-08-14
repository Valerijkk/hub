package rbac_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/rbac"
)

func TestPermissionRegistry_ContainsUniqueNamesAndUniqueNonZeroBits(t *testing.T) {
	defs := rbac.ListAllPermissionDefs()
	require.NotEmpty(t, defs)

	names := make(map[string]struct{}, len(defs))
	bits := make(map[rbac.Permission]struct{}, len(defs))
	for _, def := range defs {
		assert.NotEmpty(t, def.Name)
		_, duplicateName := names[def.Name]
		assert.False(t, duplicateName, "duplicate permission name %q", def.Name)
		names[def.Name] = struct{}{}
		if def.Bit != 0 {
			_, duplicateBit := bits[def.Bit]
			assert.False(t, duplicateBit, "duplicate permission bit %d", def.Bit)
			bits[def.Bit] = struct{}{}
		}
	}
}

func TestPermissionRegistry_ValidatesKnownAndUnknownNames(t *testing.T) {
	for _, def := range rbac.ListAllPermissionDefs() {
		assert.True(t, rbac.ValidPermissionName(def.Name), def.Name)
	}

	assert.False(t, rbac.ValidPermissionName(""))
	assert.False(t, rbac.ValidPermissionName("task.escalate_everything"))
}

func TestPermissionRegistry_ReturnsDefensiveCopy(t *testing.T) {
	first := rbac.ListAllPermissionDefs()
	require.NotEmpty(t, first)
	originalName := first[0].Name
	first[0].Name = "mutated.by.caller"

	second := rbac.ListAllPermissionDefs()
	assert.Equal(t, originalName, second[0].Name)
}

func TestPermission_HasRejectsZeroAndRequiresWholeSubset(t *testing.T) {
	permissions := rbac.PermProjectView | rbac.PermTaskCreate

	assert.False(t, permissions.Has(0))
	assert.True(t, permissions.Has(rbac.PermProjectView))
	assert.True(t, permissions.Has(rbac.PermProjectView|rbac.PermTaskCreate))
	assert.False(t, permissions.Has(rbac.PermProjectView|rbac.PermTaskEdit))
}

func TestPermission_AddAndRemovePreserveOtherBits(t *testing.T) {
	permissions := rbac.PermProjectView
	withTaskCreate := permissions.Add(rbac.PermTaskCreate)

	assert.True(t, withTaskCreate.Has(rbac.PermProjectView))
	assert.True(t, withTaskCreate.Has(rbac.PermTaskCreate))
	assert.Equal(t, permissions, withTaskCreate.Remove(rbac.PermTaskCreate))
	assert.Equal(t, permissions, permissions.Remove(rbac.PermTaskCreate))
}

func TestDefaultRolePermissions_ViewerCannotCreateTasksButDeveloperCan(t *testing.T) {
	assert.False(t, rbac.DefaultRolePermissions["viewer"].Has(rbac.PermTaskCreate))
	assert.True(t, rbac.DefaultRolePermissions["developer"].Has(rbac.PermTaskCreate))
}
