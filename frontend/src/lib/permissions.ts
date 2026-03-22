/**
 * Frontend permission constants mirroring backend/internal/rbac/permissions.go.
 *
 * IMPORTANT: The count (26) and order MUST match the backend exactly.
 * If backend permissions change, this file must be updated to match.
 */
export const Permission = {
	ProjectView: 1n << 0n,
	ProjectEdit: 1n << 1n,
	ProjectDelete: 1n << 2n,
	ProjectManageMembers: 1n << 3n,
	ProjectManageRoles: 1n << 4n,
	ProjectManageWorkflow: 1n << 5n,
	TaskCreate: 1n << 6n,
	TaskEdit: 1n << 7n,
	TaskEditAny: 1n << 8n,
	TaskDelete: 1n << 9n,
	TaskAssign: 1n << 10n,
	TaskComment: 1n << 11n,
	FileUpload: 1n << 12n,
	FileDelete: 1n << 13n,
	FileDeleteAny: 1n << 14n,
	ChannelCreate: 1n << 15n,
	ChannelManage: 1n << 16n,
	MessageEdit: 1n << 17n,
	MessageDeleteAny: 1n << 18n,
	TestCaseCreate: 1n << 19n,
	TestCaseEdit: 1n << 20n,
	TestRunCreate: 1n << 21n,
	TestRunExecute: 1n << 22n,
	InviteCreate: 1n << 23n,
	InviteManage: 1n << 24n,
	TeamManage: 1n << 25n,
} as const;

/**
 * Check whether an effective permission bitmask includes the given permission.
 */
export function hasPermission(effective: bigint, perm: bigint): boolean {
	return (effective & perm) === perm;
}
