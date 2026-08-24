import { api, type ApiRequestOptions } from "./client";

const MAIL_THREAD_MAX_LIMIT = 100;
const MAIL_THREAD_MAX_CURSOR_LENGTH = 512;
const MAIL_THREAD_MAX_SEARCH_BYTES = 128;
const MAIL_IDENTITY_MAX_LIMIT = 100;
const MAIL_IDENTITY_MAX_TOTAL = 200;
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

export interface MailRecipientInput {
	email: string;
	name?: string | null;
}

export type MailDraftStatus =
	| "active"
	| "saving"
	| "unknown"
	| "submitting"
	| "submitted";

export interface MailDraft {
	id: string;
	draft_client_id: string;
	revision: number;
	status: MailDraftStatus;
	from_identity_id: string;
	to: MailAddress[];
	cc: MailAddress[];
	bcc: MailAddress[];
	subject: string;
	body_format: "rich" | "plain";
	rich_body: string | null;
	plain_body: string;
	priority: MailMessagePriority;
	attachment_ids: string[];
	reply_to_message_id?: string | null;
	forward_message_id?: string | null;
	updated_at: string;
}

export interface SaveMailDraftRequest {
	draft_client_id: string;
	from_identity_id: string;
	to: MailRecipientInput[];
	cc: MailRecipientInput[];
	bcc: MailRecipientInput[];
	subject: string;
	body_format: "plain";
	plain_body: string;
	priority: MailMessagePriority;
	attachment_ids: string[];
}

export type MailSubmissionStatus =
	| "queued"
	| "submitting"
	| "sent"
	| "unknown"
	| "failed";

export type MailSubmissionErrorCode =
	| "MAIL_SUBMISSION_REJECTED"
	| "MAIL_SUBMISSION_OUTCOME_UNKNOWN"
	| "MAIL_SUBMISSION_RECONCILIATION_REQUIRED"
	| "MAIL_SUBMISSION_POLICY"
	| "MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN"
	| "MAIL_SUBMISSION_IDENTITY_UNAVAILABLE"
	| "MAIL_SUBMISSION_RECIPIENT_LIMIT"
	| "MAIL_SUBMISSION_SIZE_LIMIT"
	| "MAILBOX_UNAVAILABLE";

export interface MailSubmission {
	id: string;
	status: MailSubmissionStatus;
	message_id: string | null;
	error_code: MailSubmissionErrorCode | null;
	created_at: string;
	updated_at: string;
}

export type MailAttachmentDisposition = "attachment" | "inline";
export type MailAttachmentScanStatus =
	| "pending"
	| "clean"
	| "rejected"
	| "error";

export interface MailAttachment {
	id: string;
	filename: string;
	media_type: string;
	size_bytes: number;
	disposition: MailAttachmentDisposition;
	content_id: string | null;
	scan_status: MailAttachmentScanStatus;
	sha256: string | null;
}

export interface PresignMailAttachmentRequest {
	draft_client_id: string;
	filename: string;
	media_type: string;
	size_bytes: number;
	disposition: MailAttachmentDisposition;
	content_id?: string | null;
}

export interface PresignMailAttachmentResponse {
	attachment_id: string;
	upload_url: string;
	expires_at: string;
	required_headers: Record<string, string>;
}

export interface ConfirmMailAttachmentRequest {
	sha256: string;
	size_bytes: number;
}

export interface MailIdentity {
	id: string;
	address: string;
	display_name: string;
	kind: "primary" | "alias" | "group";
	send_allowed: boolean;
	is_default: boolean;
	status: "pending" | "active" | "error" | "deleting";
}

export interface MailIdentityPage {
	items: MailIdentity[];
	next_cursor: string | null;
}

export interface MailIdentityListParams {
	limit?: number;
	cursor?: string;
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
function normalizeMailIdentityLimit(value: number | undefined): number {
	if (value === undefined) return MAIL_IDENTITY_MAX_LIMIT;
	if (
		typeof value !== "number" ||
		!Number.isSafeInteger(value) ||
		value < 1 ||
		value > MAIL_IDENTITY_MAX_LIMIT
	) {
		throw new Error("Invalid mail identity limit");
	}
	return value;
}

function normalizeMailIdentityCursor(value: string | undefined): string | null {
	if (value === undefined) return null;
	if (
		typeof value !== "string" ||
		value.length === 0 ||
		value.length > MAIL_THREAD_MAX_CURSOR_LENGTH ||
		!MAIL_THREAD_CURSOR_PATTERN.test(value)
	) {
		throw new Error("Invalid mail identity cursor");
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

function normalizeMailDraftID(draftId: string): string {
	const normalized = normalizeMailUUID(draftId);
	if (normalized === null) {
		throw new Error("Invalid mail draft ID");
	}
	return normalized;
}

function normalizeMailAttachmentID(attachmentId: string): string {
	const normalized = normalizeMailUUID(attachmentId);
	if (normalized === null) {
		throw new Error("Invalid mail attachment ID");
	}
	return normalized;
}

export function getMailIdentities(
	params: MailIdentityListParams = {},
	options?: ApiRequestOptions,
): Promise<MailIdentityPage> {
	const query = new URLSearchParams({
		limit: String(normalizeMailIdentityLimit(params.limit)),
	});
	const cursor = normalizeMailIdentityCursor(params.cursor);
	if (cursor !== null) query.set("cursor", cursor);
	return api.get<MailIdentityPage>(`/mail/identities?${query.toString()}`, options);
}

export async function getAllMailIdentities(
	options?: ApiRequestOptions,
): Promise<MailIdentityPage> {
	const items: MailIdentity[] = [];
	let cursor: string | undefined;
	const seenCursors = new Set<string>();
	const maximumPages = Math.ceil(MAIL_IDENTITY_MAX_TOTAL / MAIL_IDENTITY_MAX_LIMIT);

	for (
		let page = 0;
		page < maximumPages && items.length < MAIL_IDENTITY_MAX_TOTAL;
		page += 1
	) {
		const response = await getMailIdentities(
			{
				limit: Math.min(
					MAIL_IDENTITY_MAX_LIMIT,
					MAIL_IDENTITY_MAX_TOTAL - items.length,
				),
				cursor,
			},
			options,
		);
		items.push(...response.items);
		if (response.next_cursor === null) {
			return { items, next_cursor: null };
		}
		const nextCursor = normalizeMailIdentityCursor(response.next_cursor);
		if (nextCursor === null || seenCursors.has(nextCursor)) {
			throw new Error("Invalid mail identity pagination");
		}
		seenCursors.add(nextCursor);
		cursor = nextCursor;
	}

	throw new Error("Mail identity pagination exceeded bounds");
}

export interface MailReplyDraftIntent {
	readonly sourceMessageID: string;
	readonly draftClientID: string;
	readonly fromIdentityID: string;
	readonly idempotencyKey: string;
}

function createMailReplyDraftUUID(): string {
	const value = globalThis.crypto?.randomUUID?.();
	if (typeof value !== "string" || !MAIL_UUID_V4_PATTERN.test(value)) {
		throw new Error("Unable to create mail reply draft intent");
	}
	return value.toLowerCase();
}

export function createMailReplyDraftIntent(
	sourceMessageID: string,
	fromIdentityID: string,
): MailReplyDraftIntent {
	const normalizedFromIdentityID = normalizeMailUUID(fromIdentityID);
	if (normalizedFromIdentityID === null) {
		throw new Error("Invalid mail identity ID");
	}
	return Object.freeze({
		sourceMessageID: normalizeMailMessageID(sourceMessageID),
		draftClientID: createMailReplyDraftUUID(),
		fromIdentityID: normalizedFromIdentityID,
		idempotencyKey: createMailReplyDraftUUID(),
	});
}

function mailReplyDraftString(value: unknown): value is string {
	return typeof value === "string" && value.length <= 8_388_608;
}

function mailReplyDraftAddress(value: unknown): value is MailAddress {
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		return false;
	}
	const candidate = value as Record<string, unknown>;
	return (
		mailReplyDraftString(candidate.email) &&
		candidate.email.length > 0 &&
		(candidate.name === null || mailReplyDraftString(candidate.name))
	);
}

function mailReplyDraftAddresses(value: unknown): value is MailAddress[] {
	return (
		Array.isArray(value) &&
		value.length <= 100 &&
		value.every(mailReplyDraftAddress)
	);
}

function mailReplyDraftResponseMatches(
	value: unknown,
	intent: MailReplyDraftIntent,
): value is MailDraft {
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		return false;
	}
	const candidate = value as Record<string, unknown>;
	const sameUUID = (actual: unknown, expected: string) =>
		typeof actual === "string" && normalizeMailUUID(actual) === expected;
	return (
		typeof candidate.id === "string" &&
		normalizeMailUUID(candidate.id) !== null &&
		sameUUID(candidate.draft_client_id, intent.draftClientID) &&
		sameUUID(candidate.from_identity_id, intent.fromIdentityID) &&
		sameUUID(candidate.reply_to_message_id, intent.sourceMessageID) &&
		candidate.forward_message_id === null &&
		candidate.status === "active" &&
		typeof candidate.revision === "number" &&
		Number.isSafeInteger(candidate.revision) &&
		candidate.revision >= 1 &&
		candidate.body_format === "plain" &&
		candidate.rich_body === null &&
		mailReplyDraftString(candidate.subject) &&
		mailReplyDraftString(candidate.plain_body) &&
		(candidate.priority === "low" ||
			candidate.priority === "normal" ||
			candidate.priority === "high") &&
		mailReplyDraftAddresses(candidate.to) &&
		mailReplyDraftAddresses(candidate.cc) &&
		mailReplyDraftAddresses(candidate.bcc) &&
		Array.isArray(candidate.attachment_ids) &&
		candidate.attachment_ids.length === 0 &&
		typeof candidate.updated_at === "string" &&
		Number.isFinite(new Date(candidate.updated_at).getTime())
	);
}

export async function createMailReplyDraft(
	intent: MailReplyDraftIntent,
): Promise<MailDraft> {
	const response = await api.postWithResponse<unknown>(
		`/mail/messages/${encodeURIComponent(intent.sourceMessageID)}/reply-draft`,
		{
			draft_client_id: intent.draftClientID,
			from_identity_id: intent.fromIdentityID,
		},
		{ idempotencyKey: intent.idempotencyKey },
	);
	if (
		response.status !== 201 ||
		!mailReplyDraftResponseMatches(response.data, intent)
	) {
		throw new Error("Invalid mail reply draft response");
	}
	return response.data;
}

export function createMailDraft(
	request: SaveMailDraftRequest,
	options: ApiRequestOptions,
): Promise<MailDraft> {
	return api.post<MailDraft>("/mail/drafts", request, options);
}

export function getMailDraft(
	draftId: string,
	options?: ApiRequestOptions,
): Promise<MailDraft> {
	return api.get<MailDraft>(
		`/mail/drafts/${encodeURIComponent(normalizeMailDraftID(draftId))}`,
		options,
	);
}

export function updateMailDraft(
	draftId: string,
	request: SaveMailDraftRequest,
	options: ApiRequestOptions,
): Promise<MailDraft> {
	return api.patch<MailDraft>(
		`/mail/drafts/${encodeURIComponent(normalizeMailDraftID(draftId))}`,
		request,
		options,
	);
}

export function deleteMailDraft(
	draftId: string,
	options: ApiRequestOptions,
): Promise<void> {
	return api.delete<void>(
		`/mail/drafts/${encodeURIComponent(normalizeMailDraftID(draftId))}`,
		options,
	);
}

export function sendMailDraft(
	draftId: string,
	options: ApiRequestOptions,
): Promise<MailSubmission> {
	return api.post<MailSubmission>(
		`/mail/drafts/${encodeURIComponent(normalizeMailDraftID(draftId))}/send`,
		undefined,
		options,
	);
}
export function getMailSubmission(
	submissionId: string,
	options?: ApiRequestOptions,
): Promise<MailSubmission> {
	const normalizedSubmissionID = normalizeMailUUID(submissionId);
	if (normalizedSubmissionID === null) {
		throw new Error("Invalid mail submission ID");
	}
	return api.get<MailSubmission>(
		`/mail/submissions/${encodeURIComponent(normalizedSubmissionID)}`,
		options,
	);
}


export function presignMailAttachment(
	request: PresignMailAttachmentRequest,
	options: ApiRequestOptions,
): Promise<PresignMailAttachmentResponse> {
	return api.post<PresignMailAttachmentResponse>(
		"/mail/attachments/presign",
		request,
		options,
	);
}

export async function putPresignedMailAttachment(
	contract: PresignMailAttachmentResponse,
	body: ArrayBuffer,
	options?: Pick<ApiRequestOptions, "signal">,
): Promise<void> {
	const response = await fetch(contract.upload_url, {
		method: "PUT",
		headers: contract.required_headers,
		body,
		credentials: "omit",
		redirect: "error",
		signal: options?.signal,
		referrerPolicy: "no-referrer",
	});
	if (!response.ok) {
		throw new Error("Mail attachment upload failed");
	}
}

export function confirmMailAttachment(
	attachmentId: string,
	request: ConfirmMailAttachmentRequest,
	options?: ApiRequestOptions,
): Promise<MailAttachment> {
	return api.post<MailAttachment>(
		`/mail/attachments/${encodeURIComponent(
			normalizeMailAttachmentID(attachmentId),
		)}/confirm`,
		request,
		options,
	);
}

export function deleteMailAttachment(
	attachmentId: string,
	options?: ApiRequestOptions,
): Promise<void> {
	return api.delete<void>(
		`/mail/attachments/${encodeURIComponent(
			normalizeMailAttachmentID(attachmentId),
		)}`,
		options,
	);
}


export type MailMessageMutationAction =
	| "move"
	| "mark_read"
	| "mark_unread"
	| "flag"
	| "unflag"
	| "trash";

export interface MailMessageMutationIntent {
	readonly targetType: "message";
	readonly targetIDs: readonly [string];
	readonly action: MailMessageMutationAction;
	readonly folderID: string | null;
	readonly idempotencyKey: string;
}

export type MailMessageMutationOutcome =
	| "succeeded"
	| "partial"
	| "failed"
	| "unknown";

const MAIL_UUID_V4_PATTERN =
	/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

function normalizeMailMessageMutationFolderID(folderID: string): string {
	const normalized = normalizeMailUUID(folderID);
	if (normalized === null) {
		throw new Error("Invalid mail mutation folder ID");
	}
	return normalized;
}

function mailMessageMutationActionValid(
	action: MailMessageMutationAction,
): boolean {
	return (
		action === "move" ||
		action === "mark_read" ||
		action === "mark_unread" ||
		action === "flag" ||
		action === "unflag" ||
		action === "trash"
	);
}

export function createMailMessageMutationIntent(
	messageID: string,
	action: MailMessageMutationAction,
	folderID?: string,
): MailMessageMutationIntent {
	if (!mailMessageMutationActionValid(action)) {
		throw new Error("Invalid mail message mutation action");
	}
	const normalizedMessageID = normalizeMailMessageID(messageID);
	const normalizedFolderID =
		action === "move"
			? normalizeMailMessageMutationFolderID(folderID ?? "")
			: null;
	if (action !== "move" && folderID !== undefined) {
		throw new Error("Unexpected mail mutation folder ID");
	}
	const idempotencyKey = globalThis.crypto?.randomUUID?.();
	if (
		typeof idempotencyKey !== "string" ||
		!MAIL_UUID_V4_PATTERN.test(idempotencyKey)
	) {
		throw new Error("Unable to create mail mutation idempotency key");
	}
	return Object.freeze({
		targetType: "message" as const,
		targetIDs: Object.freeze([normalizedMessageID]) as readonly [string],
		action,
		folderID: normalizedFolderID,
		idempotencyKey: idempotencyKey.toLowerCase(),
	});
}

function isMailMessageMutationRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isMailMessageMutationNonNegativeInteger(value: unknown): value is number {
	return typeof value === "number" && Number.isSafeInteger(value) && value >= 0;
}

function mailMessageMutationResponseMatches(
	value: unknown,
	intent: MailMessageMutationIntent,
	status: 200 | 202,
): MailMessageMutationOutcome | null {
	if (!isMailMessageMutationRecord(value) || !Array.isArray(value.results)) {
		return null;
	}
	if (value.results.length !== 1 || intent.targetIDs.length !== 1) {
		return null;
	}
	const target = value.results[0];
	if (
		!isMailMessageMutationRecord(target) ||
		target.target_type !== intent.targetType ||
		target.target_id !== intent.targetIDs[0] ||
		target.training_status !== "not_requested" ||
		!isMailMessageMutationNonNegativeInteger(target.affected_message_count) ||
		!isMailMessageMutationNonNegativeInteger(target.succeeded_message_count) ||
		!isMailMessageMutationNonNegativeInteger(target.failed_message_count) ||
		target.affected_message_count > 200 ||
		target.succeeded_message_count > 200 ||
		target.failed_message_count > 200 ||
		target.succeeded_message_count + target.failed_message_count >
			target.affected_message_count
	) {
		return null;
	}
	if (status === 200) {
		if (
			target.status === "succeeded" &&
			target.affected_message_count >= 1 &&
			target.succeeded_message_count === target.affected_message_count &&
			target.failed_message_count === 0 &&
			target.code === null
		) {
			return "succeeded";
		}
		if (
			target.status === "partial" &&
			target.affected_message_count >= 2 &&
			target.succeeded_message_count >= 1 &&
			target.failed_message_count >= 1 &&
			target.succeeded_message_count + target.failed_message_count ===
				target.affected_message_count &&
			target.code === "MAIL_ACTION_REJECTED"
		) {
			return "partial";
		}
		if (
			target.status === "failed" &&
			((target.affected_message_count === 0 &&
				target.succeeded_message_count === 0 &&
				target.failed_message_count === 0 &&
				(target.code === "MAIL_TARGET_UNAVAILABLE" ||
					target.code === "MAIL_FOLDER_UNAVAILABLE")) ||
				(target.affected_message_count >= 1 &&
					target.succeeded_message_count === 0 &&
					target.failed_message_count === target.affected_message_count &&
					target.code === "MAIL_ACTION_REJECTED"))
		) {
			return "failed";
		}
		return null;
	}
	return (
		target.status === "unknown" &&
		target.affected_message_count >= 1 &&
		target.succeeded_message_count + target.failed_message_count <
			target.affected_message_count &&
		target.code === "MAIL_ENGINE_OUTCOME_UNKNOWN"
			? "unknown"
			: null
	);
}

export async function mutateMailMessage(
	intent: MailMessageMutationIntent,
): Promise<MailMessageMutationOutcome> {
	const request = {
		target_type: intent.targetType,
		target_ids: [...intent.targetIDs],
		action: intent.action,
		...(intent.folderID === null ? {} : { folder_id: intent.folderID }),
	};
	const response = await api.postWithResponse<unknown>(
		"/mail/messages/bulk",
		request,
		{ idempotencyKey: intent.idempotencyKey },
	);
	if (response.status === 200) {
		const outcome = mailMessageMutationResponseMatches(response.data, intent, 200);
		if (outcome === "succeeded" || outcome === "partial" || outcome === "failed") {
			return outcome;
		}
	}
	if (response.status === 202) {
		const outcome = mailMessageMutationResponseMatches(response.data, intent, 202);
		if (outcome === "unknown") return outcome;
	}
	throw new Error("Invalid mail message mutation response");
}

export function normalizeMailMessageMutationFolderIDForRoute(
	folderID: string,
): string {
	return normalizeMailMessageMutationFolderID(folderID);
}
