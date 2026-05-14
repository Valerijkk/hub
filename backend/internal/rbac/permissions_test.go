package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// ---------------------------------------------------------------------------
// Registry / ValidPermissionName tests (SC-14-4).
// ---------------------------------------------------------------------------

func TestValidPermissionName_KnownReturnsTrue(t *testing.T) {
	known := []string{
		"project.view",
		"project.edit",
		"project.delete",
		"project.manage_members",
		"project.manage_roles",
		"project.manage_workflow",
		"task.create",
		"task.edit",
		"task.edit_any",
		"task.delete",
		"task.assign",
		"task.comment",
		"file.upload",
		"file.delete",
		"file.delete_any",
		"channel.create",
		"channel.manage",
		"message.edit",
		"message.delete_any",
		"testcase.create",
		"testcase.edit",
		"testrun.create",
		"testrun.execute",
		"invite.create",
		"invite.manage",
		"team.manage",
	}
	for _, n := range known {
		t.Run(n, func(t *testing.T) {
			assert.True(t, ValidPermissionName(n), "expected %q to be valid", n)
		})
	}
}

func TestValidPermissionName_UnknownReturnsFalse(t *testing.T) {
	unknown := []string{
		"",
		" ",
		"project",
		"project.",
		".view",
		"Project.View",        // case-sensitive
		"project.view ",       // trailing whitespace
		"unknown.permission",
		"task.create_extra",
		"file.delete_any.x",
		"team.manage; DROP TABLE users",
	}
	for _, n := range unknown {
		t.Run(n, func(t *testing.T) {
			assert.False(t, ValidPermissionName(n), "expected %q to be invalid", n)
		})
	}
}

func TestPermissionRegistry_AllNamesUnique(t *testing.T) {
	seen := make(map[string]int, len(permissionRegistry))
	for i, p := range permissionRegistry {
		if prev, ok := seen[p.Name]; ok {
			t.Fatalf("duplicate permission name %q at indices %d and %d", p.Name, prev, i)
		}
		seen[p.Name] = i
	}
	assert.Equal(t, len(permissionRegistry), len(seen),
		"every registry entry should contribute a unique name")
}

func TestPermissionRegistry_AllBitsUnique(t *testing.T) {
	// Bits with value 0 are intentionally "unallocated" (permission stored only
	// by name, e.g. when the 64-bit space is exhausted). Multiple Bit=0 entries
	// must remain legal — only assigned (non-zero) bits must be unique.
	seenBit := make(map[Permission]string)
	for _, p := range permissionRegistry {
		if p.Bit == 0 {
			continue
		}
		if prev, ok := seenBit[p.Bit]; ok {
			t.Fatalf("permissions %q and %q share bit %d",
				prev, p.Name, p.Bit)
		}
		seenBit[p.Bit] = p.Name
	}
}

func TestPermissionRegistry_BitsAreSingleBitPositions(t *testing.T) {
	// A non-zero Bit must be a power of two (exactly one bit set), otherwise
	// Has/Add/Remove semantics break. Bit == 0 is allowed and means "no slot".
	for _, p := range permissionRegistry {
		if p.Bit == 0 {
			continue
		}
		b := uint64(p.Bit)
		assert.Equal(t, uint64(0), b&(b-1),
			"%s has Bit=%d which is not a single-bit value", p.Name, p.Bit)
	}
}

func TestBuildPermissionNameMap_CoversAll(t *testing.T) {
	m := buildPermissionNameMap()
	require.Equal(t, len(permissionRegistry), len(m),
		"map and registry must have the same cardinality")
	for _, p := range permissionRegistry {
		got, ok := m[p.Name]
		require.True(t, ok, "registry entry %q missing from map", p.Name)
		assert.Equal(t, p.Bit, got, "bit mismatch for %q", p.Name)
	}
}

func TestPermissionNameMap_MatchesRegistry(t *testing.T) {
	// Package-level permissionNameMap (initialised at import) must reflect the
	// registry. This guards against future drift where someone edits the map
	// directly instead of the registry.
	require.Equal(t, len(permissionRegistry), len(permissionNameMap))
	for _, p := range permissionRegistry {
		got, ok := permissionNameMap[p.Name]
		require.True(t, ok, "permissionNameMap missing %q", p.Name)
		assert.Equal(t, p.Bit, got)
	}
}

func TestListAllPermissionDefs_ReturnsDefensiveCopy(t *testing.T) {
	a := ListAllPermissionDefs()
	require.Equal(t, len(permissionRegistry), len(a))

	// Mutate the returned slice and assert the registry is unaffected.
	if len(a) > 0 {
		a[0] = PermissionDef{Name: "tampered", Bit: 0, Display: "x"}
	}

	b := ListAllPermissionDefs()
	require.Equal(t, len(permissionRegistry), len(b))
	if len(b) > 0 {
		assert.NotEqual(t, "tampered", b[0].Name,
			"mutating the returned slice must not affect future calls")
		assert.Equal(t, permissionRegistry[0].Name, b[0].Name)
	}
}

// ---------------------------------------------------------------------------
// permissionsFromNames / permissionNames round-trip
// ---------------------------------------------------------------------------

func TestPermissionsFromNames_IgnoresUnknownButCountsKnown(t *testing.T) {
	// permissionsFromNames must not panic on unknown names — callers (legacy
	// fallback in toCustomRoleResponse) may have stale data. It silently skips
	// unknowns. validatePermissionNames is the gate that rejects them.
	got := permissionsFromNames([]string{"project.view", "nope.nope", "task.create"})
	assert.True(t, got.Has(PermProjectView))
	assert.True(t, got.Has(PermTaskCreate))
	assert.False(t, got.Has(PermProjectDelete))
}

func TestPermissionNamesAndPermissionsFromNames_RoundTrip(t *testing.T) {
	want := []string{"project.view", "task.create", "file.upload"}
	perm := permissionsFromNames(want)
	got := permissionNames(perm)

	assert.ElementsMatch(t, want, got,
		"round-trip through bitmask must preserve the set of names")
}

// ---------------------------------------------------------------------------
// validatePermissionNames behaviour (service-layer gate)
// ---------------------------------------------------------------------------

func TestValidatePermissionNames_AcceptsAllKnown(t *testing.T) {
	all := make([]string, 0, len(permissionRegistry))
	for _, p := range permissionRegistry {
		all = append(all, p.Name)
	}
	assert.NoError(t, validatePermissionNames(all))
}

func TestValidatePermissionNames_AcceptsEmpty(t *testing.T) {
	// An empty permission set is allowed at the service layer (a role with no
	// permissions is meaningful — e.g. a placeholder role waiting to be
	// configured). The HTTP layer's validate:"required" tag handles the
	// "Permissions field is mandatory" case before we get here.
	assert.NoError(t, validatePermissionNames(nil))
	assert.NoError(t, validatePermissionNames([]string{}))
}

func TestValidatePermissionNames_RejectsUnknown(t *testing.T) {
	err := validatePermissionNames([]string{"project.view", "made.up.permission"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "made.up.permission")
}

func TestValidatePermissionNames_FailsOnFirstUnknown(t *testing.T) {
	// Order matters for error message clarity.
	err := validatePermissionNames([]string{"first.bad", "second.bad"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "first.bad")
	assert.NotContains(t, err.Error(), "second.bad")
}
