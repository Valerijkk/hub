import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

import {
	createMailThreadMutationIntent,
	mutateMailThreads,
	type MailMessageMutationAction,
	type MailThreadMutationIntent,
	type MailThreadMutationResult,
} from "@/api/mail";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";

import { mailKeys } from "./mail-query-keys";

const RETRY_DELAY_MIN_SECONDS = 1;
const RETRY_DELAY_MAX_SECONDS = 5;
const RETRY_CLOCK_INTERVAL_MS = 250;

type BulkAction = Extract<
	MailMessageMutationAction,
	"mark_read" | "mark_unread" | "trash"
>;

type RetryState = {
	readonly intent: MailThreadMutationIntent;
	readonly availableAt: number;
};

type StatusNotice =
	| {
			readonly kind: MailThreadMutationResult["outcome"];
			readonly result: MailThreadMutationResult;
	  }
	| {
			readonly kind: "busy" | "busyExhausted" | "conflict" | "requestFailed";
	  };

type MailMutationApiError = {
	readonly code: number;
	readonly error_code?: string;
	readonly retry_after_seconds?: number;
};

interface MailThreadBulkActionsProps {
	readonly selectedThreadIDs: readonly string[];
	readonly onClearSelection: () => void;
	readonly onSelectionLockedChange: (locked: boolean) => void;
}

function getApiError(error: unknown): MailMutationApiError | null {
	if (typeof error !== "object" || error === null) {
		return null;
	}

	const candidate = error as Partial<MailMutationApiError>;
	return typeof candidate.code === "number"
		? (candidate as MailMutationApiError)
		: null;
}

function isBusyError(error: unknown): error is MailMutationApiError {
	const apiError = getApiError(error);
	return apiError?.code === 423 && apiError.error_code === "MAIL_ACCOUNT_BUSY";
}

function isConflictError(error: unknown): boolean {
	return getApiError(error)?.code === 409;
}

function getRetryDelaySeconds(error: MailMutationApiError): number {
	const requested = error.retry_after_seconds;

	if (
		typeof requested !== "number" ||
		!Number.isFinite(requested) ||
		requested <= 0
	) {
		return RETRY_DELAY_MIN_SECONDS;
	}

	return Math.min(
		RETRY_DELAY_MAX_SECONDS,
		Math.max(RETRY_DELAY_MIN_SECONDS, Math.ceil(requested)),
	);
}

export function MailThreadBulkActions({
	selectedThreadIDs,
	onClearSelection,
	onSelectionLockedChange,
}: MailThreadBulkActionsProps) {
	const { t } = useTranslation("mail");
	const queryClient = useQueryClient();
	const inFlightRef = useRef(false);
	const retryAttemptedRef = useRef(false);
	const [pendingTrashIntent, setPendingTrashIntent] =
		useState<MailThreadMutationIntent | null>(null);
	const [retryState, setRetryState] = useState<RetryState | null>(null);
	const [isExecuting, setIsExecuting] = useState(false);
	const [clockNow, setClockNow] = useState(() => Date.now());
	const [notice, setNotice] = useState<StatusNotice | null>(null);

	const retrySeconds =
		retryState === null
			? 0
			: Math.max(0, Math.ceil((retryState.availableAt - clockNow) / 1000));
	const selectionLocked =
		pendingTrashIntent !== null || retryState !== null || isExecuting;

	useEffect(() => {
		onSelectionLockedChange(selectionLocked);
		return () => onSelectionLockedChange(false);
	}, [onSelectionLockedChange, selectionLocked]);

	useEffect(() => {
		if (retryState === null) {
			return;
		}

		setClockNow(Date.now());
		const intervalID = window.setInterval(
			() => setClockNow(Date.now()),
			RETRY_CLOCK_INTERVAL_MS,
		);

		return () => window.clearInterval(intervalID);
	}, [retryState]);

	const invalidateMailData = useCallback(async () => {
		await Promise.all([
			queryClient.invalidateQueries({ queryKey: mailKeys.threadsRoot() }),
			queryClient.invalidateQueries({ queryKey: mailKeys.folders() }),
		]);
	}, [queryClient]);

	const execute = useCallback(
		async (intent: MailThreadMutationIntent, isRetry: boolean) => {
			if (inFlightRef.current) {
				return;
			}

			inFlightRef.current = true;
			setIsExecuting(true);
			setNotice(null);

			try {
				const result = await mutateMailThreads(intent);
				await invalidateMailData();

				retryAttemptedRef.current = false;
				setRetryState(null);
				setNotice({ kind: result.outcome, result });

				if (result.outcome === "succeeded") {
					onClearSelection();
					window.requestAnimationFrame(() => {
						document.getElementById("mail-thread-list-heading")?.focus();
					});
				}
			} catch (error) {
				if (isBusyError(error)) {
					if (!isRetry && !retryAttemptedRef.current) {
						const retryDelaySeconds = getRetryDelaySeconds(error);

						retryAttemptedRef.current = true;
						setRetryState({
							intent,
							availableAt: Date.now() + retryDelaySeconds * 1000,
						});
						setNotice({ kind: "busy" });
					} else {
						retryAttemptedRef.current = false;
						setRetryState(null);
						setNotice({ kind: "busyExhausted" });
					}
					return;
				}

				retryAttemptedRef.current = false;
				setRetryState(null);

				if (isConflictError(error)) {
					setNotice({ kind: "conflict" });
					return;
				}
				setNotice({ kind: "requestFailed" });
			} finally {
				inFlightRef.current = false;
				setIsExecuting(false);
			}
		},
		[invalidateMailData, onClearSelection],
	);

	const startAction = useCallback(
		(action: BulkAction) => {
			if (selectionLocked || selectedThreadIDs.length === 0) {
				return;
			}

			retryAttemptedRef.current = false;
			setNotice(null);
			const intent = createMailThreadMutationIntent(selectedThreadIDs, action);

			if (action === "trash") {
				setPendingTrashIntent(intent);
				return;
			}
			void execute(intent, false);
		},
		[execute, selectedThreadIDs, selectionLocked],
	);

	const handleTrashDialogOpenChange = useCallback((open: boolean) => {
		if (!open && !inFlightRef.current) {
			setPendingTrashIntent(null);
		}
	}, []);

	const handleTrashConfirm = useCallback(async () => {
		if (pendingTrashIntent === null) {
			return;
		}
		await execute(pendingTrashIntent, false);
	}, [execute, pendingTrashIntent]);

	const handleRetry = useCallback(() => {
		if (retryState === null || retrySeconds > 0 || inFlightRef.current) {
			return;
		}
		setRetryState(null);
		void execute(retryState.intent, true);
	}, [execute, retrySeconds, retryState]);

	const handleCancelRetry = useCallback(() => {
		retryAttemptedRef.current = false;
		setRetryState(null);
		setNotice(null);
	}, []);

	let noticeText: string | null = null;
	if (notice !== null) {
		switch (notice.kind) {
			case "succeeded":
				noticeText = t("threadList.bulk.success", {
					succeeded: notice.result.succeededMessageCount,
				});
				break;
			case "partial":
				noticeText = t("threadList.bulk.partial", {
					succeeded: notice.result.succeededMessageCount,
					affected: notice.result.affectedMessageCount,
					failed: notice.result.failedMessageCount,
				});
				break;
			case "failed":
				noticeText = t("threadList.bulk.failed");
				break;
			case "unknown":
				noticeText = t("threadList.bulk.unknown");
				break;
			case "busy":
				noticeText = t("threadList.bulk.busy");
				break;
			case "busyExhausted":
				noticeText = t("threadList.bulk.busyExhausted");
				break;
			case "conflict":
				noticeText = t("threadList.bulk.conflict");
				break;
			case "requestFailed":
				noticeText = t("threadList.bulk.requestFailed");
				break;
		}
	}

	return (
		<>
			{selectedThreadIDs.length > 0 ? (
				<section
					aria-label={t("threadList.bulk.label")}
					className="flex flex-wrap items-center gap-2 border-b border-b1 bg-card px-3 py-2"
				>
					<p className="mr-auto text-sm text-t2">
						{t("threadList.bulk.selected", {
							count: selectedThreadIDs.length,
						})}
					</p>
					<div
						aria-label={t("threadList.bulk.label")}
						className="flex flex-wrap items-center gap-2"
						role="group"
					>
						<Button
							className="min-h-11"
							disabled={selectionLocked || selectedThreadIDs.length === 0}
							onClick={() => startAction("mark_read")}
							size="sm"
							type="button"
						>
							{t("threadList.bulk.markRead")}
						</Button>
						<Button
							className="min-h-11"
							disabled={selectionLocked || selectedThreadIDs.length === 0}
							onClick={() => startAction("mark_unread")}
							size="sm"
							type="button"
						>
							{t("threadList.bulk.markUnread")}
						</Button>
						<Button
							className="min-h-11"
							disabled={selectionLocked || selectedThreadIDs.length === 0}
							onClick={() => startAction("trash")}
							size="sm"
							type="button"
						>
							{t("threadList.bulk.trash")}
						</Button>
					</div>
				</section>
			) : null}

			{noticeText !== null ? (
				<p
					{...(notice?.kind === "succeeded"
						? { "aria-live": "polite" }
						: { role: "alert" })}
					className="px-3 py-2 text-sm text-t2"
				>
					{noticeText}
				</p>
			) : null}

			{retryState !== null ? (
				<div className="flex flex-wrap items-center gap-2 px-3 py-2">
					{retrySeconds > 0 ? (
						<span className="text-sm text-t2">
							{t("threadList.bulk.retryWait", { seconds: retrySeconds })}
						</span>
					) : (
						<Button
							className="min-h-11"
							disabled={isExecuting}
							onClick={handleRetry}
							size="sm"
							type="button"
						>
							{t("threadList.bulk.retry")}
						</Button>
					)}
					<Button
						className="min-h-11"
						disabled={isExecuting}
						onClick={handleCancelRetry}
						size="sm"
						type="button"
					>
						{t("threadList.bulk.cancelRetry")}
					</Button>
				</div>
			) : null}

			<ConfirmDialog
				cancelLabel={t("threadList.bulk.cancel")}
				confirmLabel={t("threadList.bulk.confirm")}
				description={t("threadList.bulk.confirmDescription")}
				destructive
				disabled={isExecuting}
				onConfirm={handleTrashConfirm}
				onOpenChange={handleTrashDialogOpenChange}
				open={pendingTrashIntent !== null}
				title={t("threadList.bulk.confirmTitle", {
					count:
						pendingTrashIntent?.targetIDs.length ?? selectedThreadIDs.length,
				})}
			/>
		</>
	);
}
