package rbac

// Permission представляет битовую маску разрешений (uint64).
// Каждое разрешение — один бит, что позволяет хранить до 64 различных разрешений
// в быстром bitmask-представлении. Разрешения, не помещающиеся в 64 бита, хранятся
// исключительно как имена в `Role.permission_names` (см. ent/schema/role.go); их
// `Bit` в реестре равен 0 — они невидимы для bitmask-fast-path, но сохраняются
// через storage-layer.
type Permission uint64

// Битовые маски прав доступа (RBAC).
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

// PermissionDef описывает одно разрешение в реестре.
//
// Bit == 0 означает, что разрешение не имеет места в 64-битной маске
// (используется для permissions с индексом >= 64). Такие разрешения существуют
// только как имена в `Role.permission_names` и не участвуют в bitmask-арифметике
// middleware. Когда 64 бита будет исчерпано, новые разрешения добавляются с Bit=0
// без breaking change для существующего кода.
type PermissionDef struct {
	Name    string
	Bit     Permission
	Display string
}

// permissionRegistry — единый source of truth для всех известных разрешений.
//
// При добавлении нового разрешения:
//   1. Если осталось < 64 битов: объявить новую константу `Perm...` и добавить
//      запись с её Bit.
//   2. Если все 64 бита заняты: добавить запись с Bit=0 — разрешение будет
//      жить только в permission_names. Middleware-уровень должен мигрировать
//      на чтение имени напрямую (followup-план).
var permissionRegistry = []PermissionDef{
	{Name: "project.view", Bit: PermProjectView, Display: "View project"},
	{Name: "project.edit", Bit: PermProjectEdit, Display: "Edit project"},
	{Name: "project.delete", Bit: PermProjectDelete, Display: "Delete project"},
	{Name: "project.manage_members", Bit: PermProjectManageMembers, Display: "Manage project members"},
	{Name: "project.manage_roles", Bit: PermProjectManageRoles, Display: "Manage project roles"},
	{Name: "project.manage_workflow", Bit: PermProjectManageWorkflow, Display: "Manage project workflow"},
	{Name: "task.create", Bit: PermTaskCreate, Display: "Create tasks"},
	{Name: "task.edit", Bit: PermTaskEdit, Display: "Edit own tasks"},
	{Name: "task.edit_any", Bit: PermTaskEditAny, Display: "Edit any task"},
	{Name: "task.delete", Bit: PermTaskDelete, Display: "Delete tasks"},
	{Name: "task.assign", Bit: PermTaskAssign, Display: "Assign tasks"},
	{Name: "task.comment", Bit: PermTaskComment, Display: "Comment on tasks"},
	{Name: "file.upload", Bit: PermFileUpload, Display: "Upload files"},
	{Name: "file.delete", Bit: PermFileDelete, Display: "Delete own files"},
	{Name: "file.delete_any", Bit: PermFileDeleteAny, Display: "Delete any file"},
	{Name: "channel.create", Bit: PermChannelCreate, Display: "Create channels"},
	{Name: "channel.manage", Bit: PermChannelManage, Display: "Manage channels"},
	{Name: "message.edit", Bit: PermMessageEdit, Display: "Edit own messages"},
	{Name: "message.delete_any", Bit: PermMessageDeleteAny, Display: "Delete any message"},
	{Name: "testcase.create", Bit: PermTestCaseCreate, Display: "Create test cases"},
	{Name: "testcase.edit", Bit: PermTestCaseEdit, Display: "Edit test cases"},
	{Name: "testrun.create", Bit: PermTestRunCreate, Display: "Create test runs"},
	{Name: "testrun.execute", Bit: PermTestRunExecute, Display: "Execute test runs"},
	{Name: "invite.create", Bit: PermInviteCreate, Display: "Create invites"},
	{Name: "invite.manage", Bit: PermInviteManage, Display: "Manage invites"},
	{Name: "team.manage", Bit: PermTeamManage, Display: "Manage teams"},
}

// permissionNameMap построен из реестра один раз при инициализации пакета.
var permissionNameMap = buildPermissionNameMap()

func buildPermissionNameMap() map[string]Permission {
	m := make(map[string]Permission, len(permissionRegistry))
	for _, p := range permissionRegistry {
		m[p.Name] = p.Bit
	}
	return m
}

// ValidPermissionName возвращает true, если имя зарегистрировано в реестре.
// Используется в CreateCustomRole/UpdateCustomRole для отказа от опечаток.
func ValidPermissionName(name string) bool {
	_, ok := permissionNameMap[name]
	return ok
}

// ListAllPermissionDefs возвращает копию реестра — для UI и документации.
// Возвращается копия, чтобы вызывающий код не мог изменить внутреннее состояние.
func ListAllPermissionDefs() []PermissionDef {
	out := make([]PermissionDef, len(permissionRegistry))
	copy(out, permissionRegistry)
	return out
}

// Has проверяет, содержит ли набор разрешений указанное разрешение.
// W6-A3-005 / P2-BE-19: Bit=0 → всегда false (раньше p&0==0 давало fail-open).
func (p Permission) Has(perm Permission) bool {
	if perm == 0 {
		return false
	}
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
