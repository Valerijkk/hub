/** Founder and admin may access system admin routes (ORG-04). */
export function isSystemAdmin(role: string | undefined | null): boolean {
	return role === "admin" || role === "founder";
}
