import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
	createMailMessageMutationIntent,
	getMailFolders,
	mutateMailMessage,
	normalizeMailMessageMutationFolderIDForRoute,
	type MailMessageMutationAction,
	type MailMessageMutationIntent,
	type MailThreadMessageSummary,
} from "@/api/mail";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { mailKeys } from "./mail-query-keys";

type MailMessageActionNotice =
	| "unknown"
	| "partial"
	| "busy"
	| "conflict"
	| "failed";

interface MailMessageActionsProps {
	message: Pick<MailThreadMessageSummary, "id" | "seen" | "flagged">;
	threadID: string;
	onMovedOrTrashed: (
		action: "move" | "trash",
		folderID: string | null,
	) => void | Promise<void>;
}

function isMailMessageActionError(
	error: unknown,
	code: number,
	errorCode?: string,
): boolean {
	if (typeof error !== "object" || error === null) return false;
	const candidate = error as {
		code?: unknown;
		error_code?: unknown;
	};
	return (
		candidate.code === code &&
		(errorCode === undefined || candidate.error_code === errorCode)
	);
}

function retryAfterSeconds(error: unknown): number {
	if (typeof error !== "object" || error === null) return 1;
	const value = (error as { retry_after_seconds?: unknown }).retry_after_seconds;
	return typeof value === "number" && Number.isInteger(value) && value >= 1 && value <= 5
		? value
		: 1;
}

export function MailMessageActions({
	message,
	threadID,
	onMovedOrTrashed,
}: MailMessageActionsProps) {
	const { t } = useTranslation("mail");
	const queryClient = useQueryClient();
	const [moveOpen, setMoveOpen] = useState(false);
	const [moveFolderID, setMoveFolderID] = useState("");
	const [trashConfirmationOpen, setTrashConfirmationOpen] = useState(false);
	const [isMutating, setIsMutating] = useState(false);
	const [notice, setNotice] = useState<MailMessageActionNotice | null>(null);
	const [retryIntent, setRetryIntent] = useState<MailMessageMutationIntent | null>(null);
	const [retryAvailableAt, setRetryAvailableAt] = useState<number | null>(null);
	const [retryClock, setRetryClock] = useState(() => Date.now());
	const foldersQuery = useQuery({
		queryKey: mailKeys.folders(),
		queryFn: getMailFolders,
		enabled: moveOpen,
		retry: false,
		refetchOnWindowFocus: false,
	});

	useEffect(() => {
		if (retryAvailableAt === null) return;
		const remaining = retryAvailableAt - Date.now();
		if (remaining <= 0) {
			setRetryClock(Date.now());
			return;
		}
		const timer = globalThis.setTimeout(() => setRetryClock(Date.now()), remaining);
		return () => globalThis.clearTimeout(timer);
	}, [retryAvailableAt]);

	const folders = useMemo(() => {
		const seen = new Set<string>();
		return (foldersQuery.data?.items ?? []).flatMap((folder) => {
			try {
				const id = normalizeMailMessageMutationFolderIDForRoute(folder.id);
				if (seen.has(id)) return [];
				seen.add(id);
				return [{ id, name: folder.name }];
			} catch {
				return [];
			}
		});
	}, [foldersQuery.data?.items]);

	const invalidateMailMutation = async () => {
		await Promise.all([
			queryClient.invalidateQueries({ queryKey: mailKeys.message(message.id) }),
			queryClient.invalidateQueries({ queryKey: mailKeys.thread(threadID) }),
			queryClient.invalidateQueries({ queryKey: ["mail", "threads"] }),
			queryClient.invalidateQueries({ queryKey: mailKeys.folders() }),
		]);
	};

	const execute = async (intent: MailMessageMutationIntent): Promise<void> => {
		if (isMutating) return;
		setIsMutating(true);
		setNotice(null);
		try {
			const outcome = await mutateMailMessage(intent);
			if (outcome === "unknown") {
				setRetryIntent(null);
				setRetryAvailableAt(null);
				setNotice("unknown");
				return;
			}
			setRetryIntent(null);
			setRetryAvailableAt(null);
			await invalidateMailMutation();
			if (outcome === "partial" || outcome === "failed") {
				setNotice(outcome);
				return;
			}
			if (intent.action === "move" || intent.action === "trash") {
				await onMovedOrTrashed(intent.action, intent.folderID);
			}
		} catch (error) {
			if (isMailMessageActionError(error, 423, "MAIL_ACCOUNT_BUSY")) {
				const delay = retryAfterSeconds(error);
				setRetryIntent(intent);
				setRetryAvailableAt(Date.now() + delay * 1_000);
				setRetryClock(Date.now());
				setNotice("busy");
			} else {
				setRetryIntent(null);
				setRetryAvailableAt(null);
				setNotice(isMailMessageActionError(error, 409) ? "conflict" : "failed");
			}
			throw error;
		} finally {
			setIsMutating(false);
		}
	};

	const start = async (
		action: MailMessageMutationAction,
		folderID?: string,
	): Promise<void> => {
		setRetryIntent(null);
		setRetryAvailableAt(null);
		const intent = createMailMessageMutationIntent(message.id, action, folderID);
		await execute(intent);
	};

	const retrySeconds =
		retryAvailableAt === null
			? 0
			: Math.max(0, Math.ceil((retryAvailableAt - retryClock) / 1_000));
	const retryReady = retryIntent !== null && retrySeconds === 0 && !isMutating;

	return (
		<section className="mt-4 border-t border-b1 pt-3" aria-label={t("reader.actions.label")}>
			<div className="flex flex-wrap gap-2">
				<Button
					variant="outline"
					size="sm"
					className="min-h-11"
					disabled={isMutating}
					onClick={() => {
						void start(message.seen ? "mark_unread" : "mark_read").catch(() => undefined);
					}}
				>
					{t(message.seen ? "reader.actions.markUnread" : "reader.actions.markRead")}
				</Button>
				<Button
					variant="outline"
					size="sm"
					className="min-h-11"
					disabled={isMutating}
					onClick={() => {
						void start(message.flagged ? "unflag" : "flag").catch(() => undefined);
					}}
				>
					{t(message.flagged ? "reader.actions.unflag" : "reader.actions.flag")}
				</Button>
				<Button
					variant="outline"
					size="sm"
					className="min-h-11"
					disabled={isMutating}
					onClick={() => setMoveOpen((current) => !current)}
				>
					{t("reader.actions.move")}
				</Button>
				<Button
					variant="destructive"
					size="sm"
					className="min-h-11"
					disabled={isMutating}
					onClick={() => setTrashConfirmationOpen(true)}
				>
					{t("reader.actions.trash")}
				</Button>
			</div>
			{moveOpen ? (
				<div className="mt-3 flex flex-wrap items-center gap-2">
					{foldersQuery.isPending ? (
						<p className="text-xs text-t3">{t("reader.actions.moveLoading")}</p>
					) : foldersQuery.isError ? (
						<p className="text-xs text-t2" role="alert">
							{t("reader.actions.moveUnavailable")}
						</p>
					) : folders.length === 0 ? (
						<p className="text-xs text-t3">{t("reader.actions.moveEmpty")}</p>
					) : (
						<>
							<Select value={moveFolderID} onValueChange={setMoveFolderID} disabled={isMutating}>
								<SelectTrigger className="min-h-11 w-[min(100%,18rem)]">
									<SelectValue placeholder={t("reader.actions.movePlaceholder")} />
								</SelectTrigger>
								<SelectContent>
									{folders.map((folder) => (
										<SelectItem key={folder.id} value={folder.id}>
											{folder.name}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
							<Button
								variant="outline"
								size="sm"
								className="min-h-11"
								disabled={isMutating || moveFolderID.length === 0}
								onClick={() => {
									if (moveFolderID.length === 0) return;
									void start("move", moveFolderID).catch(() => undefined);
								}}
							>
								{t("reader.actions.moveConfirm")}
							</Button>
						</>
					)}
				</div>
			) : null}
			{notice ? (
				<div className="mt-3 flex flex-wrap items-center gap-2 rounded-[var(--r2)] bg-[var(--amber)]/10 p-3 text-xs text-t2" role="alert">
					<p>
						{notice === "unknown"
							? t("reader.actions.unknown")
							: notice === "busy"
								? t("reader.actions.busy", { seconds: Math.max(1, retrySeconds) })
								: notice === "partial"
									? t("reader.actions.partial")
								: notice === "conflict"
									? t("reader.actions.conflict")
									: t("reader.actions.failed")}
					</p>
					{retryIntent ? (
						<Button
							variant="outline"
							size="sm"
							className="min-h-11"
							disabled={!retryReady}
							onClick={() => {
								if (!retryIntent || !retryReady) return;
								void execute(retryIntent).catch(() => undefined);
							}}
						>
							{retryReady
								? t("reader.actions.retry")
								: t("reader.actions.retryWait", { seconds: Math.max(1, retrySeconds) })}
						</Button>
					) : null}
				</div>
			) : null}
			<ConfirmDialog
				open={trashConfirmationOpen}
				onOpenChange={setTrashConfirmationOpen}
				title={t("reader.actions.trashTitle")}
				description={t("reader.actions.trashDescription")}
				confirmLabel={t("reader.actions.trashConfirm")}
				cancelLabel={t("reader.actions.cancel")}
				destructive
				disabled={isMutating}
				onConfirm={() => start("trash")}
			/>
		</section>
	);
}
