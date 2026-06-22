import { api } from "./client";

/** Пользователь в очереди admin Pending (ORG-06). */
export interface PendingUser {
	id: string;
	email: string;
	pending_since: string;
}

/** Инвайт-ссылка организации для регистрации новых пользователей (ORG-06). */
export interface OrgInviteResponse {
	invite_token: string;
	invite_url: string;
}

export function listPendingUsers() {
	return api.get<PendingUser[]>("/admin/users/pending");
}

export function approvePendingUser(userId: string) {
	return api.post<{ message: string }>(`/admin/users/${userId}/approve`);
}

export function rejectPendingUser(userId: string) {
	return api.post<{ message: string }>(`/admin/users/${userId}/reject`);
}

export function getOrgInvite() {
	return api.get<OrgInviteResponse>("/admin/organization/invite");
}
