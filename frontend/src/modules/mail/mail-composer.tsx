import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Paperclip, RotateCcw, Send, Trash2, X } from "lucide-react";
import {
	type ChangeEvent,
	type KeyboardEvent,
	useCallback,
	useEffect,
	useMemo,
	useRef,
	useState,
} from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import type { ApiError } from "@/api/client";
import {
	confirmMailAttachment,
	createMailDraft,
	getMailDraft,
	deleteMailAttachment,
	deleteMailDraft,
	getAllMailIdentities,
	getMailSubmission,
	type MailAttachmentScanStatus,
	type MailDraft,
	type MailMessagePriority,
	type MailSubmission,
	type MailSubmissionStatus,
	presignMailAttachment,
	putPresignedMailAttachment,
	type SaveMailDraftRequest,
	sendMailDraft,
	updateMailDraft,
} from "@/api/mail";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";

const AUTOSAVE_DELAY_MS = 700;
const MAX_RECIPIENTS_PER_FIELD = 100;
const MAX_EMAIL_BYTES = 320;
const MAX_SUBJECT_CHARACTERS = 998;
const MAX_SUBJECT_BYTES = 998;
const MAX_BODY_CHARACTERS = 2_000_000;
const MAX_BODY_BYTES = 8_388_608;
const MAX_ATTACHMENTS = 50;
const MAX_ATTACHMENT_BYTES = 26_214_400;
const MAX_ATTACHMENT_NAME_CHARACTERS = 255;
const MAX_ATTACHMENT_METADATA_BYTES = 1_020;
const SUBMISSION_POLL_DELAYS_MS = [300, 600, 1_200, 2_400, 4_800];
const CONTROL_OR_SPACE_CHARACTERS = /[\p{White_Space}\p{Cc}]/u;
const LOCAL_PART_PATTERN =
	/^[A-Za-z0-9!#$%&'*+\-/=?^_`{|}~]+(?:\.[A-Za-z0-9!#$%&'*+\-/=?^_`{|}~]+)*$/;

type RecipientField = "to" | "cc" | "bcc";
type SaveState = "idle" | "dirty" | "saving" | "saved" | "error";
type AttachmentState = "uploading" | "confirmed" | "failed" | "removing";

interface ComposerForm {
	to: string;
	cc: string;
	bcc: string;
	subject: string;
	body: string;
	priority: MailMessagePriority;
}

interface ComposerAttachment {
	localId: string;
	intentKey: string;
	file: File;
	attachmentId: string | null;
	state: AttachmentState;
	scanStatus: MailAttachmentScanStatus | null;
	confirmation: {
		sha256: string;
		sizeBytes: number;
	} | null;
}

interface DraftReference {
	id: string;
	revision: number;
}

interface ComposerSession {
	draftClientId: string;
	createKey: string;
	sendKey: string;
	controller: AbortController;
}

interface SubmissionIntent {
	draftId: string;
	revision: number;
	idempotencyKey: string;
	snapshot: SaveMailDraftRequest;
	submissionId: string | null;
	status: Exclude<MailSubmissionStatus, "sent">;
}
interface NormalizedRecipients {
	to: Array<{ email: string }>;
	cc: Array<{ email: string }>;
	bcc: Array<{ email: string }>;
}

interface MailComposerProps {
	open: boolean;
	bootstrapDraft: MailDraft | null;
	onOpenChange: (open: boolean) => void;
}

class ComposerValidationError extends Error {
	constructor(readonly translationKey: string) {
		super(translationKey);
	}
}

function utf8Bytes(value: string): number {
	return new TextEncoder().encode(value).length;
}

function isWithinTextLimit(
	value: string,
	maxCharacters: number,
	maxBytes: number,
): boolean {
	return (
		value.length <= maxCharacters &&
		Array.from(value).length <= maxCharacters &&
		utf8Bytes(value) <= maxBytes
	);
}

function createUUID(): string {
	if (typeof globalThis.crypto?.randomUUID === "function") {
		return globalThis.crypto.randomUUID();
	}
	if (typeof globalThis.crypto?.getRandomValues !== "function") {
		throw new Error("Secure random unavailable");
	}
	const bytes = new Uint8Array(16);
	globalThis.crypto.getRandomValues(bytes);
	bytes[6] = ((bytes[6] ?? 0) & 0x0f) | 0x40;
	bytes[8] = ((bytes[8] ?? 0) & 0x3f) | 0x80;
	const hex = Array.from(bytes, (value) =>
		value.toString(16).padStart(2, "0"),
	).join("");
	return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(
		12,
		16,
	)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

function newSession(draftClientId = createUUID()): ComposerSession {
	return {
		draftClientId,
		createKey: createUUID(),
		sendKey: createUUID(),
		controller: new AbortController(),
	};
}

function isAbortError(error: unknown): boolean {
	return (
		error instanceof DOMException
			? error.name === "AbortError"
			: (error as { name?: unknown } | null)?.name === "AbortError"
	);
}

function sleepWithSignal(milliseconds: number, signal: AbortSignal): Promise<void> {
	return new Promise((resolve, reject) => {
		const timer = setTimeout(resolve, milliseconds);
		const abort = () => {
			clearTimeout(timer);
			reject(new DOMException("Aborted", "AbortError"));
		};
		if (signal.aborted) {
			abort();
			return;
		}
		signal.addEventListener("abort", abort, { once: true });
	});
}

function emptyForm(): ComposerForm {
	return {
		to: "",
		cc: "",
		bcc: "",
		subject: "",
		body: "",
		priority: "normal",
	};
}

function normalizedEmail(value: string): string | null {
	const email = value.normalize("NFC").trim();
	if (
		email.length === 0 ||
		CONTROL_OR_SPACE_CHARACTERS.test(email) ||
		utf8Bytes(email) > MAX_EMAIL_BYTES ||
		Array.from(email).length > MAX_EMAIL_BYTES
	) {
		return null;
	}
	const at = email.indexOf("@");
	if (at <= 0 || at !== email.lastIndexOf("@") || at === email.length - 1) {
		return null;
	}
	const localPart = email.slice(0, at);
	const domainPart = email.slice(at + 1);
	if (
		utf8Bytes(localPart) > 64 ||
		!LOCAL_PART_PATTERN.test(localPart) ||
		/[/?#:\\]/.test(domainPart)
	) {
		return null;
	}
	let canonicalDomain: string;
	try {
		canonicalDomain = new URL(`http://${domainPart}`).hostname.toLowerCase();
	} catch {
		return null;
	}
	if (canonicalDomain.length === 0 || canonicalDomain.length > 253) {
		return null;
	}
	for (const label of canonicalDomain.split(".")) {
		if (
			label.length === 0 ||
			label.length > 63 ||
			label.startsWith("-") ||
			label.endsWith("-") ||
			!/^[a-z0-9-]+$/.test(label)
		) {
			return null;
		}
	}
	const canonicalEmail =
		`${localPart.toLowerCase()}@${canonicalDomain}`;
	return utf8Bytes(canonicalEmail) <= MAX_EMAIL_BYTES
		? canonicalEmail
		: null;
}

function recipientParts(value: string): string[] {
	return value
		.split(/[,;\n]+/u)
		.map((part) => part.trim())
		.filter(Boolean);
}

function normalizeRecipients(form: ComposerForm): NormalizedRecipients {
	const seen = new Set<string>();
	const result: NormalizedRecipients = { to: [], cc: [], bcc: [] };
	for (const field of ["to", "cc", "bcc"] as const) {
		const parts = recipientParts(form[field]);
		if (parts.length > MAX_RECIPIENTS_PER_FIELD) {
			throw new ComposerValidationError("compose.errors.tooManyRecipients");
		}
		for (const value of parts) {
			const email = normalizedEmail(value);
			if (email === null) {
				throw new ComposerValidationError("compose.errors.invalidRecipient");
			}
			if (seen.has(email)) {
				throw new ComposerValidationError("compose.errors.duplicateRecipient");
			}
			seen.add(email);
			result[field].push({ email });
		}
	}
	return result;
}

function canonicalRecipientValue(
	form: ComposerForm,
	field: RecipientField,
): string {
	return normalizeRecipients(form)[field]
		.map(({ email }) => email)
		.join(", ");
}

function hasMeaningfulContent(
	form: ComposerForm,
	attachments: ComposerAttachment[],
): boolean {
	return (
		form.to.trim().length > 0 ||
		form.cc.trim().length > 0 ||
		form.bcc.trim().length > 0 ||
		form.subject.length > 0 ||
		form.body.length > 0 ||
		form.priority !== "normal" ||
		attachments.length > 0
	);
}

function buildDraftRequest(
	session: ComposerSession,
	form: ComposerForm,
	attachments: ComposerAttachment[],
	fromIdentityId: string | null,
): SaveMailDraftRequest {
	if (fromIdentityId === null) {
		throw new ComposerValidationError("compose.errors.identityUnavailable");
	}
	if (
		!isWithinTextLimit(
			form.subject,
			MAX_SUBJECT_CHARACTERS,
			MAX_SUBJECT_BYTES,
		)
	) {
		throw new ComposerValidationError("compose.errors.subjectLimit");
	}
	if (!isWithinTextLimit(form.body, MAX_BODY_CHARACTERS, MAX_BODY_BYTES)) {
		throw new ComposerValidationError("compose.errors.bodyLimit");
	}
	const recipients = normalizeRecipients(form);
	const attachmentIds = attachments
		.filter(
			(item): item is ComposerAttachment & { attachmentId: string } =>
				item.state === "confirmed" && item.attachmentId !== null,
		)
		.map((item) => item.attachmentId);
	if (
		attachmentIds.length > MAX_ATTACHMENTS ||
		new Set(attachmentIds).size !== attachmentIds.length
	) {
		throw new ComposerValidationError("compose.errors.attachmentCount");
	}
	return {
		draft_client_id: session.draftClientId,
		from_identity_id: fromIdentityId,
		to: recipients.to,
		cc: recipients.cc,
		bcc: recipients.bcc,
		subject: form.subject,
		body_format: "plain",
		plain_body: form.body,
		priority: form.priority,
		attachment_ids: attachmentIds,
	};
}

function draftFingerprint(request: SaveMailDraftRequest): string {
	return JSON.stringify(request);
}

function safeErrorKey(error: unknown, fallback: string): string {

	const apiError = error as Partial<ApiError> | null;
	const code =
		apiError && typeof apiError.error_code === "string"
			? apiError.error_code
			: "";
	if (code === "MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN") {
		return "compose.errors.attachmentNotReady";
	}
	if (
		code === "MAIL_DRAFT_VERSION_CONFLICT" ||
		code === "MAIL_DRAFT_RECONCILIATION_REQUIRED" ||
		code === "IDEMPOTENCY_KEY_REUSED" ||
		code === "MAIL_SUBMISSION_RECONCILIATION_REQUIRED"
	) {
		return "compose.errors.conflict";
	}
	if (
		code === "MAIL_SUBMISSION_RECIPIENT_LIMIT" ||
		code === "MAIL_SUBMISSION_SIZE_LIMIT" ||
		code === "MAIL_SUBMISSION_POLICY"
	) {
		return "compose.errors.rejected";
	}
	if (apiError?.code === 409) return "compose.errors.conflict";
	if (apiError?.code === 423) return "compose.errors.busy";
	if (apiError?.code === 413 || apiError?.code === 422) {
		return "compose.errors.rejected";
	}
	return fallback;
}

function isAuthoritativeSendFailure(error: unknown): boolean {
	const apiError = error as Partial<ApiError> | null;
	const code =
		apiError && typeof apiError.error_code === "string"
			? apiError.error_code
			: "";
	return (
		apiError?.code === 400 ||
		apiError?.code === 409 ||
		apiError?.code === 413 ||
		apiError?.code === 422 ||
		code === "MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN" ||
		code === "MAIL_SUBMISSION_POLICY" ||
		code === "MAIL_SUBMISSION_RECIPIENT_LIMIT" ||
		code === "MAIL_SUBMISSION_SIZE_LIMIT" ||
		code === "MAIL_SUBMISSION_IDENTITY_UNAVAILABLE" ||
		code === "MAIL_DRAFT_VERSION_CONFLICT" ||
		code === "MAIL_DRAFT_RECONCILIATION_REQUIRED"
	);
}

function draftRequestFromServer(draft: MailDraft): SaveMailDraftRequest {
	if (
		draft.status !== "active" ||
		draft.body_format !== "plain" ||
		draft.forward_message_id !== undefined &&
		draft.forward_message_id !== null
	) {
		throw new ComposerValidationError("compose.errors.conflict");
	}
	const recipients = (addresses: MailDraft["to"]) =>
		addresses.map(({ email, name }) => ({ email, name }));
	return {
		draft_client_id: draft.draft_client_id,
		from_identity_id: draft.from_identity_id,
		to: recipients(draft.to),
		cc: recipients(draft.cc),
		bcc: recipients(draft.bcc),
		subject: draft.subject,
		body_format: "plain",
		plain_body: draft.plain_body,
		priority: draft.priority,
		attachment_ids: draft.attachment_ids,
	};
}

function composerFormFromDraft(draft: MailDraft): ComposerForm {
	return {
		to: draft.to.map(({ email }) => email).join(", "),
		cc: draft.cc.map(({ email }) => email).join(", "),
		bcc: draft.bcc.map(({ email }) => email).join(", "),
		subject: draft.subject,
		body: draft.plain_body,
		priority: draft.priority,
	};
}

async function sha256Hex(bytes: ArrayBuffer): Promise<string> {
	if (typeof globalThis.crypto?.subtle?.digest !== "function") {
		throw new Error("SHA-256 unavailable");
	}
	const digest = await globalThis.crypto.subtle.digest("SHA-256", bytes);
	return Array.from(new Uint8Array(digest), (value) =>
		value.toString(16).padStart(2, "0"),
	).join("");
}

export function MailComposer({
	open,
	bootstrapDraft,
	onOpenChange,
}: MailComposerProps) {
	const { t } = useTranslation("mail");
	const queryClient = useQueryClient();
	const [form, setForm] = useState<ComposerForm>(emptyForm);
	const [attachments, setAttachments] = useState<ComposerAttachment[]>([]);
	const [showCc, setShowCc] = useState(false);
	const [showBcc, setShowBcc] = useState(false);
	const [saveState, setSaveState] = useState<SaveState>("idle");
	const [errorKey, setErrorKey] = useState<string | null>(null);
	const [isSending, setIsSending] = useState(false);
	const [isDiscarding, setIsDiscarding] = useState(false);
	const [isRemovingAttachment, setIsRemovingAttachment] = useState(false);
	const [initialSession] = useState<ComposerSession>(() => newSession());
	const [selectedIdentityId, setSelectedIdentityId] = useState<string | null>(null);
	const [submissionIntent, setSubmissionIntent] = useState<SubmissionIntent | null>(null);
	const [recoveryRequired, setRecoveryRequired] = useState(false);
	const [isLinkedReply, setIsLinkedReply] = useState(false);
	const sessionRef = useRef<ComposerSession>(initialSession);
	const formRef = useRef(form);
	const attachmentsRef = useRef(attachments);
	const identityIdRef = useRef<string | null>(null);
	const draftRef = useRef<DraftReference | null>(null);
	const immutableCreatePayloadRef = useRef<SaveMailDraftRequest | null>(null);
	const lastSavedFingerprintRef = useRef<string | null>(null);
	const saveRecoveryBlockedRef = useRef(false);
	const serverSnapshotRef = useRef<SaveMailDraftRequest | null>(null);
	const submissionIntentRef = useRef<SubmissionIntent | null>(null);
	const saveQueueRef = useRef<Promise<void>>(Promise.resolve());
	const uploadQueueRef = useRef<Promise<void>>(Promise.resolve());
	const attachmentRemovalInProgressRef = useRef(false);
	const autosaveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const fileInputRef = useRef<HTMLInputElement>(null);
	const bootstrapDraftKeyRef = useRef<string | null>(null);
	const linkedReplyRef = useRef(false);

	const identitiesQuery = useQuery({
		queryKey: ["mail", "identities"],
		queryFn: () => getAllMailIdentities({ signal: sessionRef.current.controller.signal }),
		enabled: open,
		retry: false,
	});
	const sendableIdentities = useMemo(
		() =>
			identitiesQuery.data?.items.filter(
				(identity) => identity.status === "active" && identity.send_allowed,
			) ?? [],
		[identitiesQuery.data],
	);

	const senderIdentity = useMemo(
		() =>
			sendableIdentities.find((identity) => identity.id === selectedIdentityId) ??
			null,
		[sendableIdentities, selectedIdentityId],
	);

	useEffect(() => {
		const next =
			sendableIdentities.find((identity) => identity.id === selectedIdentityId) ??
			sendableIdentities.find((identity) => identity.is_default) ??
			sendableIdentities[0] ??
			null;
		if (next?.id === identityIdRef.current) return;
		identityIdRef.current = next?.id ?? null;
		setSelectedIdentityId(next?.id ?? null);
	}, [sendableIdentities, selectedIdentityId]);

	const replaceForm = useCallback((next: ComposerForm) => {
		formRef.current = next;
		setForm(next);
	}, []);

	const replaceAttachments = useCallback((next: ComposerAttachment[]) => {
		attachmentsRef.current = next;
		setAttachments(next);
	}, []);

	const replaceSubmissionIntent = useCallback((next: SubmissionIntent | null) => {
		submissionIntentRef.current = next;
		setSubmissionIntent(next);
	}, []);

	const resetComposer = useCallback(() => {
		if (autosaveTimerRef.current) {
			clearTimeout(autosaveTimerRef.current);
			autosaveTimerRef.current = null;
		}
		if (submissionIntentRef.current !== null) return;
		sessionRef.current.controller.abort();
		setRecoveryRequired(false);
		sessionRef.current = newSession();
		draftRef.current = null;
		immutableCreatePayloadRef.current = null;
		lastSavedFingerprintRef.current = null;
		saveQueueRef.current = Promise.resolve();
		serverSnapshotRef.current = null;
		uploadQueueRef.current = Promise.resolve();
		saveRecoveryBlockedRef.current = false;
		linkedReplyRef.current = false;
		setIsLinkedReply(false);
		replaceForm(emptyForm());
		replaceAttachments([]);
		setShowCc(false);
		setShowBcc(false);
		identityIdRef.current = null;
		setSelectedIdentityId(null);
		setSaveState("idle");
		setErrorKey(null);
		setIsSending(false);
		setIsDiscarding(false);
		attachmentRemovalInProgressRef.current = false;
		setIsRemovingAttachment(false);
	}, [replaceAttachments, replaceForm]);

	const applyDraft = useCallback((draft: MailDraft) => {
		draftRef.current = { id: draft.id, revision: draft.revision };
	}, []);

	useEffect(() => {
		if (bootstrapDraft === null) {
			bootstrapDraftKeyRef.current = null;
			return;
		}
		if (!open) return;
		const bootstrapKey = `${bootstrapDraft.id}:${bootstrapDraft.revision}:${bootstrapDraft.updated_at}`;
		if (bootstrapDraftKeyRef.current === bootstrapKey) return;
		try {
			if (
				bootstrapDraft.reply_to_message_id == null ||
				bootstrapDraft.forward_message_id != null ||
				bootstrapDraft.attachment_ids.length !== 0
			) {
				throw new ComposerValidationError("compose.errors.conflict");
			}
			const snapshot = draftRequestFromServer(bootstrapDraft);
			if (autosaveTimerRef.current) {
				clearTimeout(autosaveTimerRef.current);
				autosaveTimerRef.current = null;
			}
			sessionRef.current.controller.abort();
			sessionRef.current = newSession(bootstrapDraft.draft_client_id);
			draftRef.current = null;
			immutableCreatePayloadRef.current = null;
			lastSavedFingerprintRef.current = draftFingerprint(snapshot);
			serverSnapshotRef.current = snapshot;
			saveQueueRef.current = Promise.resolve();
			uploadQueueRef.current = Promise.resolve();
			saveRecoveryBlockedRef.current = false;
			replaceSubmissionIntent(null);
			replaceForm(composerFormFromDraft(bootstrapDraft));
			replaceAttachments([]);
			linkedReplyRef.current = true;
			setIsLinkedReply(true);
			setShowCc(bootstrapDraft.cc.length > 0);
			setShowBcc(bootstrapDraft.bcc.length > 0);
			identityIdRef.current = bootstrapDraft.from_identity_id;
			setSelectedIdentityId(bootstrapDraft.from_identity_id);
			applyDraft(bootstrapDraft);
			setRecoveryRequired(false);
			setSaveState("saved");
			setErrorKey(null);
			bootstrapDraftKeyRef.current = bootstrapKey;
		} catch (error) {
			linkedReplyRef.current = false;
			setIsLinkedReply(false);
			setSaveState("error");
			setErrorKey(
				error instanceof ComposerValidationError
					? error.translationKey
					: "compose.errors.conflict",
			);
		}
	}, [
		applyDraft,
		bootstrapDraft,
		open,
		replaceAttachments,
		replaceForm,
		replaceSubmissionIntent,
	]);

	const reconcileDraftAfterSaveFailure = useCallback(
		async (session: ComposerSession): Promise<boolean> => {
			let currentDraft = draftRef.current;
			if (currentDraft === null) {
				const immutableCreate = immutableCreatePayloadRef.current;
				if (immutableCreate === null) return false;
				const recovered = await createMailDraft(immutableCreate, {
					idempotencyKey: session.createKey,
					signal: session.controller.signal,
				});
				if (session !== sessionRef.current || session.controller.signal.aborted) {
					throw new DOMException("Aborted", "AbortError");
				}
				applyDraft(recovered);
				currentDraft = draftRef.current;
			}
			if (currentDraft === null) return false;
			const authoritative = await getMailDraft(currentDraft.id, {
				signal: session.controller.signal,
			});
			if (
				session !== sessionRef.current ||
				session.controller.signal.aborted ||
				authoritative.draft_client_id !== session.draftClientId
			) {
				throw new ComposerValidationError("compose.errors.conflict");
			}
			const serverSnapshot = draftRequestFromServer(authoritative);
			applyDraft(authoritative);
			serverSnapshotRef.current = serverSnapshot;
			const localSnapshot = buildDraftRequest(
				session,
				formRef.current,
				attachmentsRef.current,
				identityIdRef.current,
			);
			if (draftFingerprint(localSnapshot) !== draftFingerprint(serverSnapshot)) {
				return false;
			}
			lastSavedFingerprintRef.current = draftFingerprint(serverSnapshot);
			return true;
		},
		[applyDraft],
	);

	const enqueueSave = useCallback(<T,>(task: () => Promise<T>): Promise<T> => {
		const queued = saveQueueRef.current.then(task, task);
		saveQueueRef.current = queued.then(
			() => undefined,
			() => undefined,
		);
		return queued;
	}, []);

	const persistLatest = useCallback(
		async (allowEmpty = false): Promise<void> => {
			return enqueueSave(async () => {
				const currentForm = formRef.current;
				const session = sessionRef.current;
				if (session.controller.signal.aborted) return;
				if (saveRecoveryBlockedRef.current) {
					setSaveState("error");
					setErrorKey("compose.errors.conflict");
					setRecoveryRequired(true);
					throw new ComposerValidationError("compose.errors.conflict");
				}

				const currentAttachments = attachmentsRef.current;
				if (
					!allowEmpty &&
					draftRef.current === null &&
					!hasMeaningfulContent(currentForm, currentAttachments)
				) {
					setSaveState("idle");
					return;
				}
				setSaveState("saving");
				setErrorKey(null);
				try {
					if (draftRef.current === null) {
						if (immutableCreatePayloadRef.current === null) {
							immutableCreatePayloadRef.current = buildDraftRequest(
								session,
								currentForm,
								currentAttachments,
								identityIdRef.current,
							);
						}
						const created = await createMailDraft(
							immutableCreatePayloadRef.current,
							{
								idempotencyKey: session.createKey,
								signal: session.controller.signal,
							},
						);
						if (session !== sessionRef.current || session.controller.signal.aborted) {
							return;
						}
						applyDraft(created);
						lastSavedFingerprintRef.current = draftFingerprint(
							immutableCreatePayloadRef.current,
						);
						serverSnapshotRef.current = immutableCreatePayloadRef.current;
					}

					const latest = buildDraftRequest(
						session,
						formRef.current,
						attachmentsRef.current,
						identityIdRef.current,
					);
					const latestFingerprint = draftFingerprint(latest);
					const currentDraft = draftRef.current;
					if (
						currentDraft !== null &&
						latestFingerprint !== lastSavedFingerprintRef.current
					) {
						const updated = await updateMailDraft(currentDraft.id, latest, {
							ifMatch: currentDraft.revision,
							signal: session.controller.signal,
						});
						if (session !== sessionRef.current || session.controller.signal.aborted) {
							return;
						}
						applyDraft(updated);
						lastSavedFingerprintRef.current = latestFingerprint;
						serverSnapshotRef.current = latest;
					}

					const now = buildDraftRequest(
						session,
						formRef.current,
						attachmentsRef.current,
						identityIdRef.current,
					);
					setSaveState(
						draftFingerprint(now) === lastSavedFingerprintRef.current
							? "saved"
							: "dirty",
					);
				} catch (error) {
					if (
						isAbortError(error) ||
						session !== sessionRef.current ||
						session.controller.signal.aborted
					) {
						return;
					}
					if (error instanceof ComposerValidationError) {
						setSaveState("error");
						setErrorKey(error.translationKey);
						throw error;
					}
					saveRecoveryBlockedRef.current = true;
					try {
						if (await reconcileDraftAfterSaveFailure(session)) {
							saveRecoveryBlockedRef.current = false;
							setRecoveryRequired(false);
							setSaveState("saved");
							return;
						}
					} catch (reconcileError) {
						if (
							isAbortError(reconcileError) ||
							session !== sessionRef.current ||
							session.controller.signal.aborted
						) {
							return;
						}
					}
					setRecoveryRequired(true);
					setSaveState("error");
					setErrorKey("compose.errors.conflict");
					throw error;
				}
			});
		},
		[applyDraft, enqueueSave, reconcileDraftAfterSaveFailure],
	);

	const flushLatest = useCallback(
		async (allowEmpty = false) => {
			if (autosaveTimerRef.current) {
				clearTimeout(autosaveTimerRef.current);
				autosaveTimerRef.current = null;
			}
			await persistLatest(allowEmpty);
		},
		[persistLatest],
	);

	const recoverDraft = useCallback(async () => {
		const session = sessionRef.current;
		const currentDraft = draftRef.current;
		if (currentDraft === null || session.controller.signal.aborted) {
			setErrorKey("compose.errors.conflict");
			return;
		}
		setSaveState("saving");
		try {
			const authoritative = await getMailDraft(currentDraft.id, {
				signal: session.controller.signal,
			});
			if (session !== sessionRef.current || session.controller.signal.aborted) return;
			if (authoritative.draft_client_id !== session.draftClientId) {
				throw new ComposerValidationError("compose.errors.conflict");
			}
			const serverSnapshot = draftRequestFromServer(authoritative);
			applyDraft(authoritative);
			serverSnapshotRef.current = serverSnapshot;
			const localSnapshot = buildDraftRequest(
				session,
				formRef.current,
				attachmentsRef.current,
				identityIdRef.current,
			);
			const localFingerprint = draftFingerprint(localSnapshot);
			if (localFingerprint !== draftFingerprint(serverSnapshot)) {
				const updated = await updateMailDraft(authoritative.id, localSnapshot, {
					ifMatch: authoritative.revision,
					signal: session.controller.signal,
				});
				if (session !== sessionRef.current || session.controller.signal.aborted) return;
				applyDraft(updated);
				serverSnapshotRef.current = localSnapshot;
			}
			lastSavedFingerprintRef.current = localFingerprint;
			saveRecoveryBlockedRef.current = false;
			setRecoveryRequired(false);
			setSaveState("saved");
			setErrorKey(null);
		} catch (error) {
			if (
				isAbortError(error) ||
				session !== sessionRef.current ||
				session.controller.signal.aborted
			) {
				return;
			}
			saveRecoveryBlockedRef.current = true;
			setRecoveryRequired(true);
			setSaveState("error");
			setErrorKey("compose.errors.conflict");
		}
	}, [applyDraft]);

	useEffect(() => {
		if (!open) return;
		if (autosaveTimerRef.current) clearTimeout(autosaveTimerRef.current);
		if (saveRecoveryBlockedRef.current) return;
		if (!hasMeaningfulContent(form, attachments)) {
			if (draftRef.current === null) setSaveState("idle");
			return;
		}
		try {
			buildDraftRequest(
				sessionRef.current,
				form,
				attachments,
				senderIdentity?.id ?? null,
			);
		} catch (error) {
			if (error instanceof ComposerValidationError) {
				setSaveState("error");
				setErrorKey(error.translationKey);
			}
			return;
		}
		setSaveState("dirty");
		setErrorKey(null);
		autosaveTimerRef.current = setTimeout(() => {
			void persistLatest().catch(() => undefined);
		}, AUTOSAVE_DELAY_MS);
		return () => {
			if (autosaveTimerRef.current) {
				clearTimeout(autosaveTimerRef.current);
				autosaveTimerRef.current = null;
			}
		};
	}, [attachments, form, open, persistLatest, senderIdentity?.id]);

	useEffect(
		() => () => {
			if (autosaveTimerRef.current) clearTimeout(autosaveTimerRef.current);
			if (submissionIntentRef.current === null) {
				sessionRef.current.controller.abort();
			}
		},
		[],
	);

	const updateField = <K extends keyof ComposerForm,>(
		field: K,
		value: ComposerForm[K],
	) => {
		replaceForm({ ...formRef.current, [field]: value });
	};

	const canonicalizeField = (field: RecipientField) => {
		try {
			const value = canonicalRecipientValue(formRef.current, field);
			updateField(field, value);
			setErrorKey(null);
		} catch (error) {
			if (error instanceof ComposerValidationError) {
				setErrorKey(error.translationKey);
			}
		}
	};

	const updateBoundedText = (
		field: "subject" | "body",
		value: string,
		maxCharacters: number,
		maxBytes: number,
		error: string,
	) => {
		if (!isWithinTextLimit(value, maxCharacters, maxBytes)) {
			setErrorKey(error);
			return;
		}
		updateField(field, value);
		setErrorKey(null);
	};

	const updateAttachment = useCallback(
		(localId: string, update: Partial<ComposerAttachment>) => {
			replaceAttachments(
				attachmentsRef.current.map((item) =>
					item.localId === localId ? { ...item, ...update } : item,
				),
			);
		},
		[replaceAttachments],
	);

	const uploadAttachment = useCallback(
		async (item: ComposerAttachment) => {
			if (linkedReplyRef.current) return;
			const session = sessionRef.current;
			if (session.controller.signal.aborted) return;
			updateAttachment(item.localId, { state: "uploading" });
			try {
				await flushLatest(true);
				if (session !== sessionRef.current || session.controller.signal.aborted) {
					return;
				}
				const bytes = await item.file.arrayBuffer();
				const sha256 = await sha256Hex(bytes);
				const mediaType = item.file.type || "application/octet-stream";
				const presigned = await presignMailAttachment(
					{
						draft_client_id: session.draftClientId,
						filename: item.file.name,
						media_type: mediaType,
						size_bytes: item.file.size,
						disposition: "attachment",
					},
					{
						idempotencyKey: item.intentKey,
						signal: session.controller.signal,
					},
				);
				if (session !== sessionRef.current || session.controller.signal.aborted) {
					return;
				}
				updateAttachment(item.localId, {
					attachmentId: presigned.attachment_id,
				});
				await putPresignedMailAttachment(presigned, bytes, { signal: session.controller.signal });
				let confirmed = await confirmMailAttachment(
					presigned.attachment_id,
					{ sha256, size_bytes: item.file.size },
					{ signal: session.controller.signal },
				);
				for (const delay of [250, 500, 1_000, 2_000]) {
					if (confirmed.scan_status !== "pending") break;
					await sleepWithSignal(delay, session.controller.signal);
					confirmed = await confirmMailAttachment(
						presigned.attachment_id,
						{ sha256, size_bytes: item.file.size },
						{ signal: session.controller.signal },
					);
				}
				if (session !== sessionRef.current || session.controller.signal.aborted) {
					return;
				}
				updateAttachment(item.localId, {
					attachmentId: confirmed.id,
					state: "confirmed",
					scanStatus: confirmed.scan_status,
					confirmation: {
						sha256,
						sizeBytes: item.file.size,
					},
				});
				await flushLatest(true);
			} catch (error) {
				if (
					isAbortError(error) ||
					session !== sessionRef.current ||
					session.controller.signal.aborted
				) return;
				updateAttachment(item.localId, { state: "failed" });
				setErrorKey(safeErrorKey(error, "compose.errors.uploadFailed"));
			}
		},
		[flushLatest, updateAttachment],
	);

	const enqueueUpload = useCallback(
		(item: ComposerAttachment) => {
			if (linkedReplyRef.current) return;
			const task = () => uploadAttachment(item);
			uploadQueueRef.current = uploadQueueRef.current.then(task, task);
		},
		[uploadAttachment],
	);

	const handleFiles = (event: ChangeEvent<HTMLInputElement>) => {
		if (linkedReplyRef.current || isInteractionLocked) {
			event.target.value = "";
			return;
		}
		const files = Array.from(event.target.files ?? []);
		event.target.value = "";
		if (files.length === 0) return;
		const available = MAX_ATTACHMENTS - attachmentsRef.current.length;
		if (files.length > available) {
			setErrorKey("compose.errors.attachmentCount");
			return;
		}
		const additions: ComposerAttachment[] = [];
		for (const file of files) {
			const mediaType = file.type || "application/octet-stream";
			if (
				file.size < 1 ||
				file.size > MAX_ATTACHMENT_BYTES ||
				Array.from(file.name).length > MAX_ATTACHMENT_NAME_CHARACTERS ||
				utf8Bytes(file.name) > MAX_ATTACHMENT_METADATA_BYTES ||
				Array.from(mediaType).length > MAX_ATTACHMENT_NAME_CHARACTERS ||
				utf8Bytes(mediaType) > MAX_ATTACHMENT_METADATA_BYTES
			) {
				setErrorKey("compose.errors.attachmentLimit");
				return;
			}
			additions.push({
				localId: createUUID(),
				intentKey: createUUID(),
				file,
				attachmentId: null,
				state: "uploading",
				scanStatus: null,
				confirmation: null,
			});
		}
		replaceAttachments([...attachmentsRef.current, ...additions]);
		setErrorKey(null);
		for (const item of additions) enqueueUpload(item);
	};

	const removeAttachment = async (item: ComposerAttachment) => {
		const session = sessionRef.current;
		const currentAttachments = attachmentsRef.current;
		const attachmentIndex = currentAttachments.findIndex(
			(candidate) => candidate.localId === item.localId,
		);
		const attachment = currentAttachments[attachmentIndex];
		if (
			session.controller.signal.aborted ||
			attachmentRemovalInProgressRef.current ||
			isInteractionLocked ||
			attachment === undefined ||
			attachment.state === "uploading" ||
			attachment.state === "removing"
		) {
			return;
		}

		attachmentRemovalInProgressRef.current = true;
		setIsRemovingAttachment(true);
		updateAttachment(attachment.localId, { state: "removing" });
		replaceAttachments(
			attachmentsRef.current.filter(
				(candidate) => candidate.localId !== attachment.localId,
			),
		);
		let isAuthoritativelyDetached = false;
		try {
			await flushLatest();
			if (session !== sessionRef.current || session.controller.signal.aborted) {
				return;
			}
			isAuthoritativelyDetached = true;
			if (attachment.attachmentId !== null) {
				await deleteMailAttachment(attachment.attachmentId, {
					signal: session.controller.signal,
				});
			}
		} catch (error) {
			if (
				isAbortError(error) ||
				session !== sessionRef.current ||
				session.controller.signal.aborted
			) return;
			if (!isAuthoritativelyDetached) {
				const nextAttachments = [...attachmentsRef.current];
				if (
					!nextAttachments.some(
						(candidate) => candidate.localId === attachment.localId,
					)
				) {
					nextAttachments.splice(
						Math.min(attachmentIndex, nextAttachments.length),
						0,
						attachment,
					);
					replaceAttachments(nextAttachments);
				}
			}
			setErrorKey("compose.errors.removeFailed");
		} finally {
			if (session === sessionRef.current && !session.controller.signal.aborted) {
				attachmentRemovalInProgressRef.current = false;
				setIsRemovingAttachment(false);
			}
		}
	};

	const retryAttachment = (item: ComposerAttachment) => {
		if (linkedReplyRef.current || isInteractionLocked || item.state !== "failed") return;
		enqueueUpload(item);
	};

	const reconcileFailedSubmission = useCallback(
		async (
			session: ComposerSession,
			intent: SubmissionIntent,
		): Promise<boolean> => {
			const authoritative = await getMailDraft(intent.draftId, {
				signal: session.controller.signal,
			});
			if (session !== sessionRef.current || session.controller.signal.aborted) {
				return false;
			}
			if (authoritative.draft_client_id !== session.draftClientId) {
				throw new ComposerValidationError("compose.errors.conflict");
			}
			const serverSnapshot = draftRequestFromServer(authoritative);
			applyDraft(authoritative);
			serverSnapshotRef.current = serverSnapshot;
			const localSnapshot = buildDraftRequest(
				session,
				formRef.current,
				attachmentsRef.current,
				identityIdRef.current,
			);
			if (draftFingerprint(localSnapshot) !== draftFingerprint(serverSnapshot)) {
				saveRecoveryBlockedRef.current = true;
				setRecoveryRequired(true);
				setSaveState("error");
				setErrorKey("compose.errors.conflict");
			} else {
				lastSavedFingerprintRef.current = draftFingerprint(serverSnapshot);
				setRecoveryRequired(false);
				setSaveState("saved");
			}
			session.sendKey = createUUID();
			replaceSubmissionIntent(null);
			return true;
		},
		[applyDraft, replaceSubmissionIntent],
	);

	const completeSentSubmission = useCallback(
		async (session: ComposerSession): Promise<void> => {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: ["mail", "folders"] }),
				queryClient.invalidateQueries({ queryKey: ["mail", "threads"] }),
				queryClient.invalidateQueries({ queryKey: ["mail", "drafts"] }),
			]);
			if (session !== sessionRef.current || session.controller.signal.aborted) {
				return;
			}
			replaceSubmissionIntent(null);
			toast.success(t("compose.sent"));
			onOpenChange(false);
			resetComposer();
		},
		[
			onOpenChange,
			queryClient,
			replaceSubmissionIntent,
			resetComposer,
			t,
		],
	);

	const handleSend = async () => {
		const hasUncleanAttachment = attachmentsRef.current.some(
			(item) => item.state !== "confirmed" || item.scanStatus !== "clean",
		);
		if (isSending || isDiscarding || hasUncleanAttachment) {
			if (hasUncleanAttachment) setErrorKey("compose.errors.attachmentNotReady");
			return;
		}
		const session = sessionRef.current;
		if (session.controller.signal.aborted) return;
		let intent: SubmissionIntent | null = submissionIntentRef.current;
		setIsSending(true);
		setErrorKey(null);
		try {
			if (intent === null) {
				const recipients = normalizeRecipients(formRef.current);
				if (recipients.to.length === 0) {
					throw new ComposerValidationError("compose.errors.toRequired");
				}
				await flushLatest(true);
				if (session !== sessionRef.current || session.controller.signal.aborted) return;
				const currentDraft = draftRef.current;
				if (currentDraft === null) throw new Error("Draft unavailable");
				intent = {
					draftId: currentDraft.id,
					revision: currentDraft.revision,
					idempotencyKey: session.sendKey,
					snapshot: buildDraftRequest(
						session,
						formRef.current,
						attachmentsRef.current,
						identityIdRef.current,
					),
					submissionId: null,
					status: "unknown",
				};
				replaceSubmissionIntent(intent);
			}
			if (intent === null) throw new Error("Submission intent unavailable");

			const obtainSubmission = async (): Promise<MailSubmission> => {
				let lastError: unknown;
				for (const delay of [0, ...SUBMISSION_POLL_DELAYS_MS]) {
					if (delay > 0) {
						await sleepWithSignal(delay, session.controller.signal);
					}
					try {
						return await sendMailDraft(intent.draftId, {
							idempotencyKey: intent.idempotencyKey,
							ifMatch: intent.revision,
							signal: session.controller.signal,
						});
					} catch (error) {
						if (isAbortError(error) || isAuthoritativeSendFailure(error)) {
							throw error;
						}
						lastError = error;
					}
				}
				throw lastError ?? new Error("Mail submission outcome unavailable");
			};

			let submission =
				intent.submissionId === null
					? await obtainSubmission()
					: await getMailSubmission(intent.submissionId, {
							signal: session.controller.signal,
						});
			if (session !== sessionRef.current || session.controller.signal.aborted) return;

			const recordSubmission = (candidate: MailSubmission) => {
				if (candidate.status === "sent") return;
				intent = {
					...intent,
					submissionId: candidate.id,
					status: candidate.status,
				};
				replaceSubmissionIntent(intent);
			};
			recordSubmission(submission);
			for (const delay of SUBMISSION_POLL_DELAYS_MS) {
				if (submission.status === "sent" || submission.status === "failed") {
					break;
				}
				await sleepWithSignal(delay, session.controller.signal);
				submission = await getMailSubmission(submission.id, {
					signal: session.controller.signal,
				});
				if (session !== sessionRef.current || session.controller.signal.aborted) return;
				recordSubmission(submission);
			}
			if (submission.status === "sent") {
				await completeSentSubmission(session);
				return;
			}
			if (submission.status === "failed") {
				if (await reconcileFailedSubmission(session, intent)) {
					setErrorKey(
						submission.error_code === "MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN"
							? "compose.errors.attachmentNotReady"
							: "compose.errors.sendFailed",
					);
				} else {
					setErrorKey("compose.errors.conflict");
				}
				return;
			}
			setErrorKey("compose.errors.sendFailed");
		} catch (error) {
			if (
				isAbortError(error) ||
				session !== sessionRef.current ||
				session.controller.signal.aborted
			) {
				return;
			}
			if (intent !== null && isAuthoritativeSendFailure(error)) {
				try {
					if (await reconcileFailedSubmission(session, intent)) {
						setErrorKey(safeErrorKey(error, "compose.errors.sendFailed"));
					} else {
						setErrorKey("compose.errors.conflict");
					}
				} catch (reconcileError) {
					if (
						isAbortError(reconcileError) ||
						session !== sessionRef.current ||
						session.controller.signal.aborted
					) {
						return;
					}
					setErrorKey("compose.errors.conflict");
				}
				return;
			}
			const key =
				error instanceof ComposerValidationError
					? error.translationKey
					: safeErrorKey(error, "compose.errors.sendFailed");
			setErrorKey(key);
		} finally {
			if (session === sessionRef.current && !session.controller.signal.aborted) {
				setIsSending(false);
			}
		}
	};

	const handleDiscard = async () => {
		if (isSending || isDiscarding) return;
		if (submissionIntentRef.current !== null) {
			setErrorKey("compose.errors.sendFailed");
			return;
		}
		const session = sessionRef.current;
		if (session.controller.signal.aborted) return;
		setIsDiscarding(true);
		setErrorKey(null);
		if (autosaveTimerRef.current) {
			clearTimeout(autosaveTimerRef.current);
			autosaveTimerRef.current = null;
		}
		try {
			if (
				draftRef.current === null &&
				immutableCreatePayloadRef.current !== null
			) {
				await enqueueSave(async () => {
					if (
						draftRef.current !== null ||
						immutableCreatePayloadRef.current === null
					) {
						return;
					}
					const recovered = await createMailDraft(
						immutableCreatePayloadRef.current,
						{
							idempotencyKey: session.createKey,
							signal: session.controller.signal,
						},
					);
					if (session !== sessionRef.current || session.controller.signal.aborted) {
						return;
					}
					applyDraft(recovered);
					lastSavedFingerprintRef.current = draftFingerprint(
						immutableCreatePayloadRef.current,
					);
				});
			} else {
				await saveQueueRef.current;
			}
			const currentDraft = draftRef.current;
			if (session !== sessionRef.current || session.controller.signal.aborted) {
				return;
			}
			if (currentDraft !== null) {
				await deleteMailDraft(currentDraft.id, {
					ifMatch: currentDraft.revision,
					signal: session.controller.signal,
				});
			}
			if (session !== sessionRef.current || session.controller.signal.aborted) {
				return;
			}
			onOpenChange(false);
			resetComposer();
		} catch (error) {
			if (
				isAbortError(error) ||
				session !== sessionRef.current ||
				session.controller.signal.aborted
			) {
				return;
			}
			setErrorKey(safeErrorKey(error, "compose.errors.discardFailed"));
		} finally {
			if (session === sessionRef.current && !session.controller.signal.aborted) {
				setIsDiscarding(false);
			}
		}
	};

	const requestClose = async () => {
		if (attachmentRemovalInProgressRef.current) return;
		if (isSending || isDiscarding) return;
		if (submissionIntentRef.current !== null) {
			setErrorKey("compose.errors.sendFailed");
			return;
		}
		if (!hasMeaningfulContent(formRef.current, attachmentsRef.current)) {
			onOpenChange(false);
			resetComposer();
			return;
		}
		try {
			await uploadQueueRef.current;
			if (
				attachmentsRef.current.some((item) => item.state === "failed")
			) {
				setErrorKey("compose.errors.uploadFailed");
				return;
			}
			await flushLatest();
			onOpenChange(false);
		} catch {
			// The safe inline error from flushLatest remains visible.
		}
	};

	const canonicalizeOnKey = (
		event: KeyboardEvent<HTMLInputElement>,
		field: RecipientField,
	) => {
		if (event.key === "Enter" || event.key === "," || event.key === ";") {
			event.preventDefault();
			canonicalizeField(field);
		}
	};

	const isUploading = attachments.some(
		(item) => item.state === "uploading" || item.state === "removing",
	);
	const hasUncleanAttachment = attachments.some(
		(item) => item.state !== "confirmed" || item.scanStatus !== "clean",
	);
	const isBusy =
		saveState === "saving" ||
		isUploading ||
		isRemovingAttachment ||
		isSending ||
		isDiscarding;
	const isSubmissionFrozen = submissionIntent !== null;
	const isInteractionLocked = isBusy || isSubmissionFrozen;

	const saveStatusKey =
		saveState === "idle"
			? "compose.status.idle"
			: saveState === "dirty"
				? "compose.status.dirty"
				: saveState === "saving"
					? "compose.status.saving"
					: saveState === "saved"
						? "compose.status.saved"
						: "compose.status.error";

	return (
		<Dialog
			open={open}
			onOpenChange={(nextOpen) => {
				if (nextOpen) onOpenChange(true);
				else void requestClose();
			}}
		>
			<DialogContent className="flex max-h-[calc(100dvh-1rem)] w-[calc(100vw-1rem)] max-w-3xl flex-col gap-0 overflow-hidden p-0 sm:max-h-[min(760px,calc(100dvh-2rem))]">
				<DialogHeader className="border-b border-b1 px-5 py-4 pr-12 text-left">
					<DialogTitle>{t("compose.title")}</DialogTitle>
					<div className="mt-2 flex items-center gap-2">
						<Label htmlFor="mail-compose-from" className="text-xs text-t3">
							{t("reader.from")}
						</Label>
						<Select
							value={selectedIdentityId ?? undefined}
							onValueChange={(value) => {
								identityIdRef.current = value;
								setSelectedIdentityId(value);
							}}
							disabled={
								isInteractionLocked || identitiesQuery.isPending || identitiesQuery.isError
							}
						>
							<SelectTrigger id="mail-compose-from" className="h-8 max-w-sm">
								<SelectValue placeholder={t("compose.identityLoading")} />
							</SelectTrigger>
							<SelectContent>
								{sendableIdentities.map((identity) => (
									<SelectItem key={identity.id} value={identity.id}>
										{identity.display_name
											? `${identity.display_name} <${identity.address}>`
											: identity.address}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>
				</DialogHeader>

				<div className="min-h-0 flex-1 overflow-y-auto">
					<div className="divide-y divide-b1 px-5">
						<div className="flex items-start gap-3 py-2">
							<Label htmlFor="mail-compose-to" className="w-10 pt-2.5 text-t3">
								{t("compose.to")}
							</Label>
							<Input
								id="mail-compose-to"
								value={form.to}
								onChange={(event) => updateField("to", event.target.value)}
								disabled={isInteractionLocked}
								onBlur={() => canonicalizeField("to")}
								onKeyDown={(event) => canonicalizeOnKey(event, "to")}
								placeholder={t("compose.recipientPlaceholder")}
								className="border-0 bg-transparent px-0 shadow-none focus-visible:scale-100 focus-visible:shadow-none"
								aria-required="true"
							/>
							<div className="flex pt-1">
								{!showCc ? (
									<Button
										size="sm"
										variant="ghost"
										onClick={() => setShowCc(true)}
									>
										{t("compose.cc")}
									</Button>
								) : null}
								{!showBcc ? (
									<Button
										size="sm"
										variant="ghost"
										onClick={() => setShowBcc(true)}
									>
										{t("compose.bcc")}
									</Button>
								) : null}
							</div>
						</div>
						{showCc ? (
							<div className="flex items-start gap-3 py-2">
								<Label htmlFor="mail-compose-cc" className="w-10 pt-2.5 text-t3">
									{t("compose.cc")}
								</Label>
								<Input
									id="mail-compose-cc"
									value={form.cc}
									onChange={(event) => updateField("cc", event.target.value)}
									disabled={isInteractionLocked}
									onBlur={() => canonicalizeField("cc")}
									onKeyDown={(event) => canonicalizeOnKey(event, "cc")}
									placeholder={t("compose.recipientPlaceholder")}
									className="border-0 bg-transparent px-0 shadow-none focus-visible:scale-100 focus-visible:shadow-none"
								/>
							</div>
						) : null}
						{showBcc ? (
							<div className="flex items-start gap-3 py-2">
								<Label htmlFor="mail-compose-bcc" className="w-10 pt-2.5 text-t3">
									{t("compose.bcc")}
								</Label>
								<Input
									id="mail-compose-bcc"
									value={form.bcc}
									onChange={(event) => updateField("bcc", event.target.value)}
									disabled={isInteractionLocked}
									onBlur={() => canonicalizeField("bcc")}
									onKeyDown={(event) => canonicalizeOnKey(event, "bcc")}
									placeholder={t("compose.recipientPlaceholder")}
									className="border-0 bg-transparent px-0 shadow-none focus-visible:scale-100 focus-visible:shadow-none"
								/>
							</div>
						) : null}
						<div className="py-2">
							<Label htmlFor="mail-compose-subject" className="sr-only">
								{t("compose.subject")}
							</Label>
							<Input
								id="mail-compose-subject"
								value={form.subject}
								disabled={isInteractionLocked}
								onChange={(event) =>
									updateBoundedText(
										"subject",
										event.target.value,
										MAX_SUBJECT_CHARACTERS,
										MAX_SUBJECT_BYTES,
										"compose.errors.subjectLimit",
									)
								}
								placeholder={t("compose.subject")}
								className="border-0 bg-transparent px-0 text-base font-medium shadow-none focus-visible:scale-100 focus-visible:shadow-none"
							/>
						</div>
					</div>
					<div className="px-5 py-4">
						<Label htmlFor="mail-compose-body" className="sr-only">
							{t("compose.body")}
						</Label>
						<Textarea
							id="mail-compose-body"
							value={form.body}
							disabled={isInteractionLocked}
							onChange={(event) =>
								updateBoundedText(
									"body",
									event.target.value,
									MAX_BODY_CHARACTERS,
									MAX_BODY_BYTES,
									"compose.errors.bodyLimit",
								)
							}
							placeholder={t("compose.bodyPlaceholder")}
							className="min-h-[260px] resize-none border-0 bg-transparent px-0 shadow-none focus-visible:scale-100 focus-visible:shadow-none"
						/>
					</div>
					{!isLinkedReply && attachments.length > 0 ? (
						<ul
							className="space-y-2 border-t border-b1 px-5 py-3"
							aria-label={t("compose.attachments")}
						>
							{attachments.map((item) => (
								<li
									key={item.localId}
									className="flex items-center gap-3 rounded-[var(--r2)] bg-raised px-3 py-2"
								>
									<Paperclip className="h-4 w-4 text-t3" aria-hidden="true" />
									<div className="min-w-0 flex-1">
										<p className="truncate text-sm font-medium text-t1">
											{item.file.name}
										</p>
										<p className="text-xs text-t3">
											{t(
												item.state === "uploading"
													? "compose.attachmentStatus.uploading"
													: item.state === "removing"
														? "compose.attachmentStatus.removing"
														: item.state === "failed"
															? "compose.attachmentStatus.failed"
															: `compose.attachmentStatus.${item.scanStatus ?? "pending"}`,
											)}
										</p>
									</div>
									{item.state === "failed" ? (
										<Button
											size="icon"
											variant="ghost"
											onClick={() => retryAttachment(item)}
											disabled={isInteractionLocked}
											aria-label={t("compose.retryAttachment")}
										>
											<RotateCcw aria-hidden="true" />
										</Button>
									) : null}
									<Button
										size="icon"
										variant="ghost"
										onClick={() => void removeAttachment(item)}
										disabled={
													isInteractionLocked || item.state === "uploading" || item.state === "removing"
										}
										aria-label={t("compose.removeAttachment", {
											name: item.file.name,
										})}
									>
										<X aria-hidden="true" />
									</Button>
								</li>
							))}
						</ul>
					) : null}
				</div>

				{errorKey || identitiesQuery.isError ? (
					<div
						role="alert"
						className="flex items-center gap-3 border-t border-b1 px-5 py-2 text-sm text-[var(--red)]"
					>
						<span className="flex-1">
							{t(
								identitiesQuery.isError
									? "compose.errors.identityUnavailable"
									: (errorKey ?? "compose.errors.saveFailed"),
							)}
						</span>
						{recoveryRequired ? (
							<Button size="sm" variant="outline" onClick={() => void recoverDraft()} disabled={isBusy || isSubmissionFrozen}>
								{t("retry")}
							</Button>
						) : null}
					</div>
				) : null}

				<footer className="flex flex-wrap items-center gap-2 border-t border-b1 px-4 py-3">
					<input
						ref={fileInputRef}
						type="file"
						multiple
						className="sr-only"
						onChange={handleFiles}
						disabled={isLinkedReply || isInteractionLocked}
						aria-label={t("compose.attach")}
					/>
					<Button
						size="icon"
						variant="ghost"
						onClick={() => fileInputRef.current?.click()}
						disabled={
							isLinkedReply ||
							attachments.length >= MAX_ATTACHMENTS ||
							isInteractionLocked
						}
						aria-label={t("compose.attach")}
					>
						<Paperclip aria-hidden="true" />
					</Button>
					<Select
						value={form.priority}
						onValueChange={(value) =>
							updateField("priority", value as MailMessagePriority)
						}
						disabled={isInteractionLocked}
					>
						<SelectTrigger
							className="w-[132px]"
							aria-label={t("compose.priority.label")}
						>
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="low">{t("compose.priority.low")}</SelectItem>
							<SelectItem value="normal">
								{t("compose.priority.normal")}
							</SelectItem>
							<SelectItem value="high">{t("compose.priority.high")}</SelectItem>
						</SelectContent>
					</Select>
					<p className="min-w-[120px] flex-1 text-xs text-t3" aria-live="polite">
						{t(saveStatusKey)}
					</p>
					<Button
						size="icon"
						variant="ghost"
						onClick={() => void handleDiscard()}
						disabled={isBusy || isSubmissionFrozen}
						aria-label={t("compose.discard")}
					>
						<Trash2 aria-hidden="true" />
					</Button>
					<Button
						onClick={() => void handleSend()}
						disabled={
							isBusy ||
							hasUncleanAttachment ||
							identitiesQuery.isPending ||
							identitiesQuery.isError ||
							senderIdentity === null
						}
					>
						<Send aria-hidden="true" />
						{isSending ? t("compose.sending") : t("compose.send")}
					</Button>
				</footer>
			</DialogContent>
		</Dialog>
	);
}
