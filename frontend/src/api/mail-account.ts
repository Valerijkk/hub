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
