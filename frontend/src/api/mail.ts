import { api } from "./client";

const MAIL_THREAD_MAX_LIMIT = 100;
const MAIL_THREAD_MAX_CURSOR_LENGTH = 512;
const MAIL_THREAD_MAX_SEARCH_BYTES = 128;
const MAIL_UUID_PATTERN =
	/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const MAIL_NIL_UUID = "00000000-0000-0000-0000-000000000000";
const MAIL_THREAD_CURSOR_PATTERN = /^[A-Za-z0-9_-]+$/;
const MAIL_QUERY_CONTROL_CHARACTERS = /[\u0000-\u001F\u007F]/;

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

export type MailProjectionState =
	| "bootstrapping"
	| "ready"
	| "degraded"
	| "rebuilding";

export type MailFolderProjectionState = MailProjectionState;

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

export interface MailThreadParticipant {
	email: string;
}

export interface MailThreadLabel {
	id: string;
	name: string;
	color: string;
	version: number;
}

export interface MailThreadSummary {
	id: string;
	subject: string;
	participants: MailThreadParticipant[];
	message_count: number;
	unread_count: number;
	last_message_at: string;
	snippet: string;
	is_pinned: boolean;
	has_attachments: boolean;
	labels: MailThreadLabel[];
}

export interface MailThreadPage {
	items: MailThreadSummary[];
	next_cursor: string | null;
	projection_state: MailProjectionState;
	projection_as_of: string | null;
}

export interface MailAddress {
	email: string;
	name: string | null;
}

export type MailMessagePriority = "low" | "normal" | "high";

export type MailCalendarMethod = "request" | "cancel" | "reply";

export interface MailThreadMessageSummary {
	id: string;
	thread_id: string;
	from: MailAddress;
	to: MailAddress[];
	cc: MailAddress[];
	subject: string;
	snippet: string;
	sent_at: string | null;
	received_at: string | null;
	priority: MailMessagePriority;
	size_bytes: number;
	has_attachments: boolean;
	is_draft: boolean;
	folder_ids: string[];
	seen: boolean;
	flagged: boolean;
	answered: boolean;
	junk: boolean;
	calendar_method: MailCalendarMethod | null;
	calendar_event_id: string | null;
}

export interface MailThreadDetail extends MailThreadSummary {
	messages: MailThreadMessageSummary[];
	next_cursor: string | null;
}

export type MailMessageAttachmentDisposition = "attachment" | "inline";

export interface MailMessageAttachment {
	id: string;
	filename: string;
	media_type: string;
	size_bytes: number;
	disposition: MailMessageAttachmentDisposition;
	content_id: string | null;
	scan_status: "clean";
	sha256: string;
}

export interface MailMessageDetail extends MailThreadMessageSummary {
	plain_body: string;
	sanitized_html: string | null;
	remote_images_blocked: boolean;
	attachments: MailMessageAttachment[];
	calendar: Record<string, never> | null;
}

export interface MailThreadListParams {
	folder_id?: string;
	label_id?: string;
	search?: string;
	starred?: boolean;
	unread?: boolean;
	has_attachments?: boolean;
	limit?: number;
	cursor?: string;
}

export interface MailThreadDetailParams {
	limit?: number;
	cursor?: string;
}

export interface MailThreadListKeyParams {
	folder_id: string | null;
	label_id: string | null;
	search: string | null;
	starred: boolean | null;
	unread: boolean | null;
	has_attachments: boolean | null;
	limit: number | null;
	cursor: string | null;
}

export interface MailThreadDetailKeyParams {
	limit: number | null;
	cursor: string | null;
}

function normalizeMailUUID(value: string | undefined): string | null {
	if (typeof value !== "string" || !MAIL_UUID_PATTERN.test(value)) {
		return null;
	}
	const normalized = value.toLowerCase();
	return normalized === MAIL_NIL_UUID ? null : normalized;
}

function normalizeMailThreadLimit(value: number | undefined): number | null {
	if (
		typeof value !== "number" ||
		!Number.isSafeInteger(value) ||
		value < 1 ||
		value > MAIL_THREAD_MAX_LIMIT
	) {
		return null;
	}
	return value;
}

function normalizeMailThreadCursor(value: string | undefined): string | null {
	if (
		typeof value !== "string" ||
		value.length === 0 ||
		value.length > MAIL_THREAD_MAX_CURSOR_LENGTH ||
		!MAIL_THREAD_CURSOR_PATTERN.test(value)
	) {
		return null;
	}
	return value;
}

function normalizeMailThreadSearch(value: string | undefined): string | null {
	if (typeof value !== "string" || MAIL_QUERY_CONTROL_CHARACTERS.test(value)) {
		return null;
	}
	const normalized = value
		.normalize("NFC")
		.toLocaleLowerCase("und")
		.normalize("NFC")
		.replace(/\s+/gu, " ")
		.trim();
	if (
		normalized.length === 0 ||
		MAIL_QUERY_CONTROL_CHARACTERS.test(normalized) ||
		/[%_\\]/.test(normalized) ||
		new TextEncoder().encode(normalized).length > MAIL_THREAD_MAX_SEARCH_BYTES
	) {
		return null;
	}
	return normalized;
}

function normalizeMailThreadBoolean(value: boolean | undefined): boolean | null {
	return typeof value === "boolean" ? value : null;
}

export function normalizeMailThreadListParams(
	params: MailThreadListParams = {},
): MailThreadListKeyParams {
	return {
		folder_id: normalizeMailUUID(params.folder_id),
		label_id: normalizeMailUUID(params.label_id),
		search: normalizeMailThreadSearch(params.search),
		starred: normalizeMailThreadBoolean(params.starred),
		unread: normalizeMailThreadBoolean(params.unread),
		has_attachments: normalizeMailThreadBoolean(params.has_attachments),
		limit: normalizeMailThreadLimit(params.limit),
		cursor: normalizeMailThreadCursor(params.cursor),
	};
}

export function normalizeMailThreadDetailParams(
	params: MailThreadDetailParams = {},
): MailThreadDetailKeyParams {
	return {
		limit: normalizeMailThreadLimit(params.limit),
		cursor: normalizeMailThreadCursor(params.cursor),
	};
}

export function normalizeMailThreadID(threadId: string): string {
	const normalized = normalizeMailUUID(threadId);
	if (normalized === null) {
		throw new Error("Invalid mail thread ID");
	}
	return normalized;
}

export function normalizeMailMessageID(messageId: string): string {
	const normalized = normalizeMailUUID(messageId);
	if (normalized === null) {
		throw new Error("Invalid mail message ID");
	}
	return normalized;
}

function appendMailThreadBoolean(
	params: URLSearchParams,
	name: "starred" | "unread" | "has_attachments",
	value: boolean | null,
): void {
	if (value !== null) {
		params.set(name, value ? "true" : "false");
	}
}

function mailThreadListSearchParams(
	query: MailThreadListKeyParams,
): URLSearchParams {
	const params = new URLSearchParams();
	if (query.folder_id !== null) params.set("folder_id", query.folder_id);
	if (query.label_id !== null) params.set("label_id", query.label_id);
	if (query.search !== null) params.set("search", query.search);
	appendMailThreadBoolean(params, "starred", query.starred);
	appendMailThreadBoolean(params, "unread", query.unread);
	appendMailThreadBoolean(params, "has_attachments", query.has_attachments);
	if (query.limit !== null) params.set("limit", String(query.limit));
	if (query.cursor !== null) params.set("cursor", query.cursor);
	return params;
}

function mailThreadDetailSearchParams(
	query: MailThreadDetailKeyParams,
): URLSearchParams {
	const params = new URLSearchParams();
	if (query.limit !== null) params.set("limit", String(query.limit));
	if (query.cursor !== null) params.set("cursor", query.cursor);
	return params;
}

function mailThreadQueryPath(path: string, params: URLSearchParams): string {
	const query = params.toString();
	return query.length === 0 ? path : `${path}?${query}`;
}

export function getMailThreads(
	params: MailThreadListParams = {},
): Promise<MailThreadPage> {
	return api.get<MailThreadPage>(
		mailThreadQueryPath(
			"/mail/threads",
			mailThreadListSearchParams(normalizeMailThreadListParams(params)),
		),
	);
}

export function getMailThread(
	threadId: string,
	params: MailThreadDetailParams = {},
): Promise<MailThreadDetail> {
	const normalizedThreadID = normalizeMailThreadID(threadId);
	return api.get<MailThreadDetail>(
		mailThreadQueryPath(
			`/mail/threads/${encodeURIComponent(normalizedThreadID)}`,
			mailThreadDetailSearchParams(normalizeMailThreadDetailParams(params)),
		),
	);
}

export function getMailMessage(messageId: string): Promise<MailMessageDetail> {
	const normalizedMessageID = normalizeMailMessageID(messageId);
	return api.get<MailMessageDetail>(
		`/mail/messages/${encodeURIComponent(normalizedMessageID)}`,
	);
}
