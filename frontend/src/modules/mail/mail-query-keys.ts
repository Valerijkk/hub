import {
	normalizeMailThreadDetailParams,
	normalizeMailThreadID,
	normalizeMailThreadListParams,
	type MailThreadDetailParams,
	type MailThreadListParams,
} from "@/api/mail";

export const mailKeys = {
	mailbox: () => ["mail", "mailbox"] as const,
	folders: () => ["mail", "folders"] as const,
	threads: (params: MailThreadListParams = {}) =>
		["mail", "threads", normalizeMailThreadListParams(params)] as const,
	thread: (threadId: string, params: MailThreadDetailParams = {}) =>
		[
			"mail",
			"thread",
			normalizeMailThreadID(threadId),
			normalizeMailThreadDetailParams(params),
		] as const,
};
