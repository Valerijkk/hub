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
	deleteMailAttachment,
	deleteMailDraft,
	getMailIdentities,
	type MailAttachmentScanStatus,
	type MailDraft,
	type MailMessagePriority,
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
}

interface DraftReference {
	id: string;
	revision: number;
}

interface ComposerSession {
	draftClientId: string;
	createKey: string;
	sendKey: string;
}

interface NormalizedRecipients {
	to: Array<{ email: string }>;
	cc: Array<{ email: string }>;
	bcc: Array<{ email: string }>;
}

interface MailComposerProps {
	open: boolean;
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

function newSession(): ComposerSession {
	return {
		draftClientId: createUUID(),
		createKey: createUUID(),
		sendKey: createUUID(),
	};
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

async function sha256Hex(bytes: ArrayBuffer): Promise<string> {
	if (typeof globalThis.crypto?.subtle?.digest !== "function") {
		throw new Error("SHA-256 unavailable");
	}
	const digest = await globalThis.crypto.subtle.digest("SHA-256", bytes);
	return Array.from(new Uint8Array(digest), (value) =>
		value.toString(16).padStart(2, "0"),
	).join("");
}

export function MailComposer({ open, onOpenChange }: MailComposerProps) {
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
	const [initialSession] = useState<ComposerSession>(newSession);
	const sessionRef = useRef<ComposerSession>(initialSession);
	const formRef = useRef(form);
	const attachmentsRef = useRef(attachments);
	const identityIdRef = useRef<string | null>(null);
	const draftRef = useRef<DraftReference | null>(null);
	const immutableCreatePayloadRef = useRef<SaveMailDraftRequest | null>(null);
	const lastSavedFingerprintRef = useRef<string | null>(null);
	const saveQueueRef = useRef<Promise<void>>(Promise.resolve());
	const uploadQueueRef = useRef<Promise<void>>(Promise.resolve());
	const autosaveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const fileInputRef = useRef<HTMLInputElement>(null);

	const identitiesQuery = useQuery({
		queryKey: ["mail", "identities"],
		queryFn: getMailIdentities,
		enabled: open,
		retry: false,
	});
	const senderIdentity = useMemo(() => {
		const candidates =
			identitiesQuery.data?.items.filter(
				(identity) =>
					identity.status === "active" && identity.send_allowed,
			) ?? [];
		return (
			candidates.find((identity) => identity.is_default) ??
			candidates[0] ??
			null
		);
	}, [identitiesQuery.data]);
	identityIdRef.current = senderIdentity?.id ?? null;

	const replaceForm = useCallback((next: ComposerForm) => {
		formRef.current = next;
		setForm(next);
	}, []);

	const replaceAttachments = useCallback((next: ComposerAttachment[]) => {
		attachmentsRef.current = next;
		setAttachments(next);
	}, []);

	const resetComposer = useCallback(() => {
		if (autosaveTimerRef.current) {
			clearTimeout(autosaveTimerRef.current);
			autosaveTimerRef.current = null;
		}
		sessionRef.current = newSession();
		draftRef.current = null;
		immutableCreatePayloadRef.current = null;
		lastSavedFingerprintRef.current = null;
		saveQueueRef.current = Promise.resolve();
		uploadQueueRef.current = Promise.resolve();
		replaceForm(emptyForm());
		replaceAttachments([]);
		setShowCc(false);
		setShowBcc(false);
		setSaveState("idle");
		setErrorKey(null);
		setIsSending(false);
		setIsDiscarding(false);
	}, [replaceAttachments, replaceForm]);

	const applyDraft = useCallback((draft: MailDraft) => {
		draftRef.current = { id: draft.id, revision: draft.revision };
	}, []);

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
								sessionRef.current,
								currentForm,
								currentAttachments,
								identityIdRef.current,
							);
						}
						const created = await createMailDraft(
							immutableCreatePayloadRef.current,
							{ idempotencyKey: sessionRef.current.createKey },
						);
						applyDraft(created);
						lastSavedFingerprintRef.current = draftFingerprint(
							immutableCreatePayloadRef.current,
						);
					}

					const latest = buildDraftRequest(
						sessionRef.current,
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
						});
						applyDraft(updated);
						lastSavedFingerprintRef.current = latestFingerprint;
					}

					const now = buildDraftRequest(
						sessionRef.current,
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
					const key =
						error instanceof ComposerValidationError
							? error.translationKey
							: safeErrorKey(error, "compose.errors.saveFailed");
					setSaveState("error");
					setErrorKey(key);
					throw error;
				}
			});
		},
		[applyDraft, enqueueSave],
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

	useEffect(() => {
		if (!open) return;
		if (autosaveTimerRef.current) clearTimeout(autosaveTimerRef.current);
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
			updateAttachment(item.localId, { state: "uploading" });
			try {
				await flushLatest(true);
				const bytes = await item.file.arrayBuffer();
				const sha256 = await sha256Hex(bytes);
				const mediaType = item.file.type || "application/octet-stream";
				const presigned = await presignMailAttachment(
					{
						draft_client_id: sessionRef.current.draftClientId,
						filename: item.file.name,
						media_type: mediaType,
						size_bytes: item.file.size,
						disposition: "attachment",
					},
					{ idempotencyKey: item.intentKey },
				);
				updateAttachment(item.localId, {
					attachmentId: presigned.attachment_id,
				});
				await putPresignedMailAttachment(presigned, bytes);
				const confirmed = await confirmMailAttachment(
					presigned.attachment_id,
					{ sha256, size_bytes: item.file.size },
				);
				updateAttachment(item.localId, {
					attachmentId: confirmed.id,
					state: "confirmed",
					scanStatus: confirmed.scan_status,
				});
				await flushLatest(true);
			} catch (error) {
				updateAttachment(item.localId, { state: "failed" });
				setErrorKey(safeErrorKey(error, "compose.errors.uploadFailed"));
			}
		},
		[flushLatest, updateAttachment],
	);

	const enqueueUpload = useCallback(
		(item: ComposerAttachment) => {
			const task = () => uploadAttachment(item);
			uploadQueueRef.current = uploadQueueRef.current.then(task, task);
		},
		[uploadAttachment],
	);

	const handleFiles = (event: ChangeEvent<HTMLInputElement>) => {
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
			});
		}
		replaceAttachments([...attachmentsRef.current, ...additions]);
		setErrorKey(null);
		for (const item of additions) enqueueUpload(item);
	};

	const removeAttachment = async (item: ComposerAttachment) => {
		if (item.state === "uploading" || item.state === "removing") return;
		updateAttachment(item.localId, { state: "removing" });
		try {
			if (item.attachmentId !== null) {
				await deleteMailAttachment(item.attachmentId);
			}
			replaceAttachments(
				attachmentsRef.current.filter(
					(candidate) => candidate.localId !== item.localId,
				),
			);
			await flushLatest();
		} catch (error) {
			updateAttachment(item.localId, { state: item.state });
			setErrorKey(safeErrorKey(error, "compose.errors.removeFailed"));
		}
	};

	const retryAttachment = (item: ComposerAttachment) => {
		if (item.state !== "failed") return;
		enqueueUpload(item);
	};

	const handleSend = async () => {
		const hasFailedAttachment = attachmentsRef.current.some(
			(item) => item.state === "failed",
		);
		if (
			isSending ||
			isDiscarding ||
			hasFailedAttachment ||
			attachmentsRef.current.some(
				(item) => item.state === "uploading" || item.state === "removing",
			)
		) {
			if (hasFailedAttachment) {
				setErrorKey("compose.errors.uploadFailed");
			}
			return;
		}
		try {
			const recipients = normalizeRecipients(formRef.current);
			if (recipients.to.length === 0) {
				throw new ComposerValidationError("compose.errors.toRequired");
			}
			setIsSending(true);
			setErrorKey(null);
			await flushLatest(true);
			const currentDraft = draftRef.current;
			if (currentDraft === null) {
				throw new Error("Draft unavailable");
			}
			const submission = await sendMailDraft(currentDraft.id, {
				idempotencyKey: sessionRef.current.sendKey,
				ifMatch: currentDraft.revision,
			});
			if (
				submission.status === "failed" ||
				submission.status === "unknown"
			) {
				setErrorKey(
					submission.error_code === "MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN"
						? "compose.errors.attachmentNotReady"
						: "compose.errors.sendFailed",
				);
				return;
			}
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: ["mail", "folders"] }),
				queryClient.invalidateQueries({ queryKey: ["mail", "threads"] }),
				queryClient.invalidateQueries({ queryKey: ["mail", "drafts"] }),
			]);
			toast.success(t("compose.sent"));
			onOpenChange(false);
			resetComposer();
		} catch (error) {
			const key =
				error instanceof ComposerValidationError
					? error.translationKey
					: safeErrorKey(error, "compose.errors.sendFailed");
			setErrorKey(key);
		} finally {
			setIsSending(false);
		}
	};

	const handleDiscard = async () => {
		if (isSending || isDiscarding) return;
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
						{ idempotencyKey: sessionRef.current.createKey },
					);
					applyDraft(recovered);
					lastSavedFingerprintRef.current = draftFingerprint(
						immutableCreatePayloadRef.current,
					);
				});
			} else {
				await saveQueueRef.current;
			}
			const currentDraft = draftRef.current;
			if (currentDraft !== null) {
				await deleteMailDraft(currentDraft.id, {
					ifMatch: currentDraft.revision,
				});
			}
			onOpenChange(false);
			resetComposer();
		} catch (error) {
			setErrorKey(safeErrorKey(error, "compose.errors.discardFailed"));
		} finally {
			setIsDiscarding(false);
		}
	};

	const requestClose = async () => {
		if (isSending || isDiscarding) return;
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
	const hasFailedAttachment = attachments.some(
		(item) => item.state === "failed",
	);
	const isBusy =
		saveState === "saving" ||
		isUploading ||
		isSending ||
		isDiscarding;

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
					<p className="text-xs text-t3">
						{senderIdentity
							? t("compose.fromValue", { address: senderIdentity.address })
							: t("compose.identityLoading")}
					</p>
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
					{attachments.length > 0 ? (
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
											item.state === "uploading" || item.state === "removing"
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
					<p
						role="alert"
						className="border-t border-b1 px-5 py-2 text-sm text-[var(--red)]"
					>
						{t(
							identitiesQuery.isError
								? "compose.errors.identityUnavailable"
								: (errorKey ?? "compose.errors.saveFailed"),
						)}
					</p>
				) : null}

				<footer className="flex flex-wrap items-center gap-2 border-t border-b1 px-4 py-3">
					<input
						ref={fileInputRef}
						type="file"
						multiple
						className="sr-only"
						onChange={handleFiles}
						aria-label={t("compose.attach")}
					/>
					<Button
						size="icon"
						variant="ghost"
						onClick={() => fileInputRef.current?.click()}
						disabled={attachments.length >= MAX_ATTACHMENTS || isBusy}
						aria-label={t("compose.attach")}
					>
						<Paperclip aria-hidden="true" />
					</Button>
					<Select
						value={form.priority}
						onValueChange={(value) =>
							updateField("priority", value as MailMessagePriority)
						}
						disabled={isSending || isDiscarding}
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
						disabled={isBusy}
						aria-label={t("compose.discard")}
					>
						<Trash2 aria-hidden="true" />
					</Button>
					<Button
						onClick={() => void handleSend()}
						disabled={
							isBusy ||
							hasFailedAttachment ||
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
