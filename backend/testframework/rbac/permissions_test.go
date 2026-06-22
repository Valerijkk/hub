package rbac_test

// Тесты модуля RBAC: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/rbac"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPermission_Has проверяет сценарий: permission has.
func TestPermission_Has(t *testing.T) {
	tests := []struct {
		name     string
		perm     rbac.Permission
		check    rbac.Permission
		expected bool
	}{
		{"single permission present", rbac.PermProjectView, rbac.PermProjectView, true},
		{"single permission absent", rbac.PermProjectView, rbac.PermProjectEdit, false},
		{"combined permissions", rbac.PermProjectView | rbac.PermTaskCreate, rbac.PermTaskCreate, true},
		{"zero permission", rbac.Permission(0), rbac.PermProjectView, false},
		{"owner has everything", ^rbac.Permission(0), rbac.PermProjectDelete, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.perm.Has(tt.check))
		})
	}
}

// TestPermission_Add проверяет сценарий: permission add.
func TestPermission_Add(t *testing.T) {
	p := rbac.PermProjectView
	p = p.Add(rbac.PermTaskCreate)

	assert.True(t, p.Has(rbac.PermProjectView), "should still have original permission")
	assert.True(t, p.Has(rbac.PermTaskCreate), "should have added permission")
	assert.False(t, p.Has(rbac.PermProjectEdit), "should not have unrelated permission")
}

// TestPermission_Remove проверяет сценарий: permission remove.
func TestPermission_Remove(t *testing.T) {
	p := rbac.PermProjectView | rbac.PermTaskCreate | rbac.PermTaskEdit
	p = p.Remove(rbac.PermTaskCreate)

	assert.True(t, p.Has(rbac.PermProjectView), "should still have remaining permission")
	assert.True(t, p.Has(rbac.PermTaskEdit), "should still have remaining permission")
	assert.False(t, p.Has(rbac.PermTaskCreate), "should have removed permission")
}

// TestPermission_AddRemoveIdempotent проверяет сценарий: permission add remove idempotent.
func TestPermission_AddRemoveIdempotent(t *testing.T) {
	p := rbac.PermProjectView
	p = p.Add(rbac.PermProjectView) // повторное добавление
	assert.Equal(t, rbac.PermProjectView, p, "adding existing perm should be idempotent")

	p = p.Remove(rbac.PermTaskCreate) // удаление несуществующего
	assert.Equal(t, rbac.PermProjectView, p, "removing non-existing perm should be idempotent")
}

// TestDefaultRolePermissions проверяет сценарий: default role права.
func TestDefaultRolePermissions(t *testing.T) {
	t.Run("owner has all permissions", func(t *testing.T) {
		owner := rbac.DefaultRolePermissions["owner"]
		assert.True(t, owner.Has(rbac.PermProjectView))
		assert.True(t, owner.Has(rbac.PermProjectDelete))
		assert.True(t, owner.Has(rbac.PermTeamManage))
		assert.Equal(t, ^rbac.Permission(0), owner)
	})

	t.Run("viewer has only view permission", func(t *testing.T) {
		viewer := rbac.DefaultRolePermissions["viewer"]
		assert.True(t, viewer.Has(rbac.PermProjectView))
		assert.False(t, viewer.Has(rbac.PermProjectEdit))
		assert.False(t, viewer.Has(rbac.PermTaskCreate))
	})

	t.Run("developer can create tasks", func(t *testing.T) {
		dev := rbac.DefaultRolePermissions["developer"]
		assert.True(t, dev.Has(rbac.PermProjectView))
		assert.True(t, dev.Has(rbac.PermTaskCreate))
		assert.True(t, dev.Has(rbac.PermFileUpload))
		assert.False(t, dev.Has(rbac.PermProjectDelete))
	})

	t.Run("tester can create test cases", func(t *testing.T) {
		tester := rbac.DefaultRolePermissions["tester"]
		assert.True(t, tester.Has(rbac.PermTestCaseCreate))
		assert.True(t, tester.Has(rbac.PermTestRunExecute))
		assert.False(t, tester.Has(rbac.PermProjectManageMembers))
	})
}

// ---------------------------------------------------------------------------
// Тесты реестра / rbac.ValidPermissionName (SC-14-4).
// ---------------------------------------------------------------------------

// TestValidPermissionName_KnownReturnsTrue проверяет сценарий: корректный permission name known returns true.
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
			assert.True(t, rbac.ValidPermissionName(n), "expected %q to be valid", n)
		})
	}
}

// TestValidPermissionName_UnknownReturnsFalse проверяет сценарий: корректный permission name unknown returns false.
func TestValidPermissionName_UnknownReturnsFalse(t *testing.T) {
	unknown := []string{
		"",
		" ",
		"project",
		"project.",
		".view",
		"Project.View",        // чувствительность к регистру
		"project.view ",       // завершающий пробел
		"unknown.permission",
		"task.create_extra",
		"file.delete_any.x",
		"team.manage; DROP TABLE users",
	}
	for _, n := range unknown {
		t.Run(n, func(t *testing.T) {
			assert.False(t, rbac.ValidPermissionName(n), "expected %q to be invalid", n)
		})
	}
}

// TestPermissionRegistry_AllNamesUnique проверяет сценарий: permission registry all names unique.
func TestPermissionRegistry_AllNamesUnique(t *testing.T) {
	seen := make(map[string]int, len(rbac.ExportPermissionRegistry()))
	for i, p := range rbac.ExportPermissionRegistry() {
		if prev, ok := seen[p.Name]; ok {
			t.Fatalf("duplicate permission name %q at indices %d and %d", p.Name, prev, i)
		}
		seen[p.Name] = i
	}
	assert.Equal(t, len(rbac.ExportPermissionRegistry()), len(seen),
		"every registry entry should contribute a unique name")
}

// TestPermissionRegistry_AllBitsUnique проверяет сценарий: permission registry all bits unique.
func TestPermissionRegistry_AllBitsUnique(t *testing.T) {
	// Биты со значением 0 намеренно "не распределены" (право хранится только
	// по имени, например когда 64-битное пространство исчерпано). Несколько записей
	// с Bit=0 остаются допустимыми — уникальными должны быть только назначенные (ненулевые) биты.
	seenBit := make(map[rbac.Permission]string)
	for _, p := range rbac.ExportPermissionRegistry() {
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

// TestPermissionRegistry_BitsAreSingleBitPositions проверяет сценарий: permission registry bits are single bit positions.
func TestPermissionRegistry_BitsAreSingleBitPositions(t *testing.T) {
	// Ненулевой Bit должен быть степенью двойки (ровно один бит установлен), иначе
	// семантика Has/Add/Remove ломается. Bit == 0 допустим и означает "нет слота".
	for _, p := range rbac.ExportPermissionRegistry() {
		if p.Bit == 0 {
			continue
		}
		b := uint64(p.Bit)
		assert.Equal(t, uint64(0), b&(b-1),
			"%s has Bit=%d which is not a single-bit value", p.Name, p.Bit)
	}
}

// TestBuildPermissionNameMap_CoversAll проверяет сценарий: build permission name map covers all.
func TestBuildPermissionNameMap_CoversAll(t *testing.T) {
	m := rbac.ExportBuildPermissionNameMap()
	require.Equal(t, len(rbac.ExportPermissionRegistry()), len(m),
		"map and registry must have the same cardinality")
	for _, p := range rbac.ExportPermissionRegistry() {
		got, ok := m[p.Name]
		require.True(t, ok, "registry entry %q missing from map", p.Name)
		assert.Equal(t, p.Bit, got, "bit mismatch for %q", p.Name)
	}
}

// TestPermissionNameMap_MatchesRegistry проверяет сценарий: permission name map matches registry.
func TestPermissionNameMap_MatchesRegistry(t *testing.T) {
	// Пакетный permissionNameMap (инициализируется при импорте) должен отражать
	// реестр. Это защищает от будущего расхождения, когда кто-то правит карту
	// напрямую вместо реестра.
	require.Equal(t, len(rbac.ExportPermissionRegistry()), len(rbac.ExportPermissionNameMap()))
	for _, p := range rbac.ExportPermissionRegistry() {
		got, ok := rbac.ExportPermissionNameMap()[p.Name]
		require.True(t, ok, "permissionNameMap missing %q", p.Name)
		assert.Equal(t, p.Bit, got)
	}
}

// TestListAllPermissionDefs_ReturnsDefensiveCopy проверяет сценарий: список all permission defs returns defensive copy.
func TestListAllPermissionDefs_ReturnsDefensiveCopy(t *testing.T) {
	a := rbac.ListAllPermissionDefs()
	require.Equal(t, len(rbac.ExportPermissionRegistry()), len(a))

	// Мутируем возвращённый срез и проверяем, что реестр не затронут.
	if len(a) > 0 {
		a[0] = rbac.PermissionDef{Name: "tampered", Bit: 0, Display: "x"}
	}

	b := rbac.ListAllPermissionDefs()
	require.Equal(t, len(rbac.ExportPermissionRegistry()), len(b))
	if len(b) > 0 {
		assert.NotEqual(t, "tampered", b[0].Name,
			"mutating the returned slice must not affect future calls")
		assert.Equal(t, rbac.ExportPermissionRegistry()[0].Name, b[0].Name)
	}
}

// ---------------------------------------------------------------------------
// rbac.ExportPermissionsFromNames / rbac.ExportPermissionNames — круговой обход (round-trip)
// ---------------------------------------------------------------------------

// TestPermissionsFromNames_IgnoresUnknownButCountsKnown проверяет сценарий: права from names ignores unknown but counts known.
func TestPermissionsFromNames_IgnoresUnknownButCountsKnown(t *testing.T) {
	// rbac.ExportPermissionsFromNames не должен паниковать на неизвестных именах — у вызывающих (legacy
	// fallback в toCustomRoleResponse) могут быть устаревшие данные. Она молча пропускает
	// неизвестные. rbac.ExportValidatePermissionNames — это шлюз, который их отклоняет.
	got := rbac.ExportPermissionsFromNames([]string{"project.view", "nope.nope", "task.create"})
	assert.True(t, got.Has(rbac.PermProjectView))
	assert.True(t, got.Has(rbac.PermTaskCreate))
	assert.False(t, got.Has(rbac.PermProjectDelete))
}

// TestPermissionNamesAndPermissionsFromNames_RoundTrip проверяет сценарий: permission names and права from names round trip.
func TestPermissionNamesAndPermissionsFromNames_RoundTrip(t *testing.T) {
	want := []string{"project.view", "task.create", "file.upload"}
	perm := rbac.ExportPermissionsFromNames(want)
	got := rbac.ExportPermissionNames(perm)

	assert.ElementsMatch(t, want, got,
		"round-trip through bitmask must preserve the set of names")
}

// ---------------------------------------------------------------------------
// Поведение rbac.ExportValidatePermissionNames (шлюз на уровне сервиса)
// ---------------------------------------------------------------------------

// TestValidatePermissionNames_AcceptsAllKnown проверяет сценарий: validate permission names accepts all known.
func TestValidatePermissionNames_AcceptsAllKnown(t *testing.T) {
	all := make([]string, 0, len(rbac.ExportPermissionRegistry()))
	for _, p := range rbac.ExportPermissionRegistry() {
		all = append(all, p.Name)
	}
	assert.NoError(t, rbac.ExportValidatePermissionNames(all))
}

// TestValidatePermissionNames_AcceptsEmpty проверяет сценарий: validate permission names accepts пустой.
func TestValidatePermissionNames_AcceptsEmpty(t *testing.T) {
	// Пустой набор прав допустим на уровне сервиса (роль без
	// прав осмысленна — например, роль-заглушка, ожидающая
	// настройки). Тег validate:"required" на уровне HTTP обрабатывает
	// случай "поле Permissions обязательно" до того, как мы сюда попадём.
	assert.NoError(t, rbac.ExportValidatePermissionNames(nil))
	assert.NoError(t, rbac.ExportValidatePermissionNames([]string{}))
}

// TestValidatePermissionNames_RejectsUnknown проверяет сценарий: validate permission names rejects unknown.
func TestValidatePermissionNames_RejectsUnknown(t *testing.T) {
	err := rbac.ExportValidatePermissionNames([]string{"project.view", "made.up.permission"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "made.up.permission")
}

// TestValidatePermissionNames_FailsOnFirstUnknown проверяет сценарий: validate permission names fails on first unknown.
func TestValidatePermissionNames_FailsOnFirstUnknown(t *testing.T) {
	// Порядок важен для ясности сообщения об ошибке.
	err := rbac.ExportValidatePermissionNames([]string{"first.bad", "second.bad"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "first.bad")
	assert.NotContains(t, err.Error(), "second.bad")
}
