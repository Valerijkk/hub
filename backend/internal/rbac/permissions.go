package rbac

// Permission представляет битовую маску разрешений (uint64).
// Каждое разрешение — один бит, что позволяет хранить до 64 различных разрешений.
type Permission uint64

const (
	PermProjectView          Permission = 1 << iota // 1
	PermProjectEdit                                 // 2
	PermProjectDelete                               // 4
	PermProjectManageMembers                        // 8
	PermProjectManageRoles                          // 16
	PermProjectManageWorkflow                       // 32
	PermTaskCreate                                  // 64
	PermTaskEdit                                    // 128
	PermTaskEditAny                                 // 256 - edit any task, not just own
	PermTaskDelete                                  // 512
	PermTaskAssign                                  // 1024
	PermTaskComment                                 // 2048
	PermFileUpload                                  // 4096
	PermFileDelete                                  // 8192
	PermFileDeleteAny                               // 16384
	PermChannelCreate                               // 32768
	PermChannelManage                               // 65536
	PermMessageEdit                                 // 131072 - edit own messages
	PermMessageDeleteAny                            // 262144
	PermTestCaseCreate                              // 524288
	PermTestCaseEdit                                // 1048576
	PermTestRunCreate                               // 2097152
	PermTestRunExecute                              // 4194304
	PermInviteCreate                                // 8388608
	PermInviteManage                                // 16777216
	PermTeamManage                                  // 33554432
)

// DefaultRolePermissions сопоставляет фиксированные роли проекта с наборами разрешений по умолчанию.
var DefaultRolePermissions = map[string]Permission{
	"owner":     ^Permission(0), // all permissions (0xFFFFFFFFFFFFFFFF)
	"developer": PermProjectView | PermTaskCreate | PermTaskEdit | PermTaskAssign | PermTaskComment | PermFileUpload | PermFileDelete | PermChannelCreate | PermMessageEdit | PermTestCaseCreate | PermTestCaseEdit | PermTestRunCreate | PermTestRunExecute | PermInviteCreate,
	"tester":    PermProjectView | PermTaskCreate | PermTaskEdit | PermTaskComment | PermFileUpload | PermFileDelete | PermMessageEdit | PermTestCaseCreate | PermTestCaseEdit | PermTestRunCreate | PermTestRunExecute,
	"viewer":    PermProjectView,
}

// Has проверяет, содержит ли набор разрешений указанное разрешение.
func (p Permission) Has(perm Permission) bool {
	return p&perm == perm
}

// Add добавляет разрешение в набор.
func (p Permission) Add(perm Permission) Permission {
	return p | perm
}

// Remove удаляет разрешение из набора.
func (p Permission) Remove(perm Permission) Permission {
	return p &^ perm
}
