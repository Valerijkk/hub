package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermission_Has(t *testing.T) {
	tests := []struct {
		name     string
		perm     Permission
		check    Permission
		expected bool
	}{
		{"single permission present", PermProjectView, PermProjectView, true},
		{"single permission absent", PermProjectView, PermProjectEdit, false},
		{"combined permissions", PermProjectView | PermTaskCreate, PermTaskCreate, true},
		{"zero permission", Permission(0), PermProjectView, false},
		{"owner has everything", ^Permission(0), PermProjectDelete, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.perm.Has(tt.check))
		})
	}
}

func TestPermission_Add(t *testing.T) {
	p := PermProjectView
	p = p.Add(PermTaskCreate)

	assert.True(t, p.Has(PermProjectView), "should still have original permission")
	assert.True(t, p.Has(PermTaskCreate), "should have added permission")
	assert.False(t, p.Has(PermProjectEdit), "should not have unrelated permission")
}

func TestPermission_Remove(t *testing.T) {
	p := PermProjectView | PermTaskCreate | PermTaskEdit
	p = p.Remove(PermTaskCreate)

	assert.True(t, p.Has(PermProjectView), "should still have remaining permission")
	assert.True(t, p.Has(PermTaskEdit), "should still have remaining permission")
	assert.False(t, p.Has(PermTaskCreate), "should have removed permission")
}

func TestPermission_AddRemoveIdempotent(t *testing.T) {
	p := PermProjectView
	p = p.Add(PermProjectView) // adding again
	assert.Equal(t, PermProjectView, p, "adding existing perm should be idempotent")

	p = p.Remove(PermTaskCreate) // removing non-existing
	assert.Equal(t, PermProjectView, p, "removing non-existing perm should be idempotent")
}

func TestDefaultRolePermissions(t *testing.T) {
	t.Run("owner has all permissions", func(t *testing.T) {
		owner := DefaultRolePermissions["owner"]
		assert.True(t, owner.Has(PermProjectView))
		assert.True(t, owner.Has(PermProjectDelete))
		assert.True(t, owner.Has(PermTeamManage))
		assert.Equal(t, ^Permission(0), owner)
	})

	t.Run("viewer has only view permission", func(t *testing.T) {
		viewer := DefaultRolePermissions["viewer"]
		assert.True(t, viewer.Has(PermProjectView))
		assert.False(t, viewer.Has(PermProjectEdit))
		assert.False(t, viewer.Has(PermTaskCreate))
	})

	t.Run("developer can create tasks", func(t *testing.T) {
		dev := DefaultRolePermissions["developer"]
		assert.True(t, dev.Has(PermProjectView))
		assert.True(t, dev.Has(PermTaskCreate))
		assert.True(t, dev.Has(PermFileUpload))
		assert.False(t, dev.Has(PermProjectDelete))
	})

	t.Run("tester can create test cases", func(t *testing.T) {
		tester := DefaultRolePermissions["tester"]
		assert.True(t, tester.Has(PermTestCaseCreate))
		assert.True(t, tester.Has(PermTestRunExecute))
		assert.False(t, tester.Has(PermProjectManageMembers))
	})
}
