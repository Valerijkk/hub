//go:build hubtest

package rbac

func ExportRBACServiceCache(s *RBACService) *PermissionCache {
	return s.cache
}

func ExportCacheKey(userID, projectID string) string {
	return cacheKey(userID, projectID)
}

func ExportPermissionRegistry() []PermissionDef {
	return permissionRegistry
}

func ExportBuildPermissionNameMap() map[string]Permission {
	return buildPermissionNameMap()
}

func ExportPermissionNameMap() map[string]Permission {
	return permissionNameMap
}

func ExportPermissionsFromNames(names []string) Permission {
	return permissionsFromNames(names)
}

func ExportPermissionNames(perm Permission) []string {
	return permissionNames(perm)
}

func ExportValidatePermissionNames(names []string) error {
	return validatePermissionNames(names)
}

// ExportInvalidateByPatternScript — Lua SCAN+DEL (A3W6-3-003 comment/script sync).
func ExportInvalidateByPatternScript() string {
	return invalidateByPatternScript
}
