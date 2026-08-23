import { api } from "./client";

export type MailAccountLifecycle =
	| "not_configured"
	| "provisioning"
	| "active"
	| "suspended"
	| "error"
	| "deleting";

export type MailAccountRecoveryCode =
	| "MAILBOX_ENGINE_DISABLED"
	| "MAILBOX_DOMAIN_REQUIRED"
	| "MAILBOX_PROVISIONING_RETRY"
	| "MAILBOX_SUSPENDED"
	| "MAILBOX_UNAVAILABLE";

export interface MailAccountCapabilities {
	rich_content: boolean;
	rules: boolean;
	spam_training: boolean;
	remote_images: boolean;
}

export interface MailAccountStatus {
	status: MailAccountLifecycle;
	primary_address: string | null;
	retry_after_seconds: number | null;
	error_code: MailAccountRecoveryCode | null;
	capabilities: MailAccountCapabilities;
}

export function getMailAccountStatus(): Promise<MailAccountStatus> {
	return api.get<MailAccountStatus>("/mail/account");
}

export type MailFolderProjectionState =
	| "bootstrapping"
	| "ready"
	| "degraded"
	| "rebuilding";

export interface MailFolder {
	id: string;
	parent_id: string | null;
	role: string | null;
	name: string;
	color: string | null;
	sort_order: number;
	unread_count: number;
	total_count: number;
	state: string;
	version: number;
}

export interface MailFolderProjection {
	items: MailFolder[];
	projection_state: MailFolderProjectionState;
	projection_as_of: string | null;
}

export function getMailFolders(): Promise<MailFolderProjection> {
	return api.get<MailFolderProjection>("/mail/folders");
}
