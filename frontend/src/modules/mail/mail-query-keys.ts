import {
	normalizeMailMessageID,
	normalizeMailThreadDetailParams,
	normalizeMailThreadID,
	normalizeMailThreadListParams,
	type MailThreadDetailParams,
	type MailThreadListParams,
} from "@/api/mail";

export const mailKeys = {
	setup: () => ["mail", "setup"] as const,
	mailbox: () => ["mail", "mailbox"] as const,
	oauthConnection: () => [...mailKeys.mailbox(), "oauth-connection"] as const,
	folders: () => ["mail", "folders"] as const,
	threadsRoot: () => ["mail", "threads"] as const,
	threads: (params: MailThreadListParams = {}) =>
		["mail", "threads", normalizeMailThreadListParams(params)] as const,
	threadRoot: (threadId: string) =>
		["mail", "thread", normalizeMailThreadID(threadId)] as const,
	thread: (threadId: string, params: MailThreadDetailParams = {}) =>
		[
			"mail",
			"thread",
			normalizeMailThreadID(threadId),
			normalizeMailThreadDetailParams(params),
		] as const,
	message: (messageId: string) =>
		["mail", "message", normalizeMailMessageID(messageId)] as const,
	messagesRoot: () => ["mail", "message"] as const,
	drafts: () => ["mail", "drafts"] as const,
	directMessage: (actorId: string | null, messageId: string | null) =>
		["mail", "direct-message", actorId, messageId] as const,
	identities: () => ["mail", "identities"] as const,
};
