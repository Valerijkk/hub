import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import type { ApiError } from "@/api/client";
import {
	createMailReplyDraft,
	createMailReplyDraftIntent,
	getAllMailIdentities,
	type MailReplyDraftIntent,
} from "@/api/mail";
import { Button } from "@/components/ui/button";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { useMailComposerContext } from "./mail-composer-context";

type MailReplyNotice = "busy" | "conflict" | "unavailable" | "unknown";

interface MailMessageReplyActionProps {
	messageID: string;
}

function mailReplyError(error: unknown): Partial<ApiError> | null {
	return typeof error === "object" && error !== null
		? (error as Partial<ApiError>)
		: null;
}

function mailReplyRetryAfterSeconds(error: unknown): number {
	const value = mailReplyError(error)?.retry_after_seconds;
	return typeof value === "number" && Number.isInteger(value) && value >= 1 && value <= 5
		? value
		: 1;
}

export function MailMessageReplyAction({
	messageID,
}: MailMessageReplyActionProps) {
	const { t } = useTranslation("mail");
	const { openReplyDraft } = useMailComposerContext();
	const [selectedIdentityID, setSelectedIdentityID] = useState<string | null>(
		null,
	);
	const [intent, setIntent] = useState<MailReplyDraftIntent | null>(null);
	const [notice, setNotice] = useState<MailReplyNotice | null>(null);
	const [isCreating, setIsCreating] = useState(false);
	const [retryAvailableAt, setRetryAvailableAt] = useState<number | null>(null);
	const [retryClock, setRetryClock] = useState(() => Date.now());
	const identitiesQuery = useQuery({
		queryKey: ["mail", "identities"],
		queryFn: getAllMailIdentities,
		retry: false,
		refetchOnWindowFocus: false,
	});
	const sendableIdentities = useMemo(
		() =>
			(identitiesQuery.data?.items ?? []).filter(
				(identity) => identity.status === "active" && identity.send_allowed,
			),
		[identitiesQuery.data],
	);
	const selectedIdentity = useMemo(
		() =>
			sendableIdentities.find((identity) => identity.id === selectedIdentityID) ??
			null,
		[sendableIdentities, selectedIdentityID],
	);

	useEffect(() => {
		const next =
			sendableIdentities.find((identity) => identity.id === selectedIdentityID) ??
			sendableIdentities.find((identity) => identity.is_default) ??
			sendableIdentities[0] ??
			null;
		setSelectedIdentityID((current) =>
			current === (next?.id ?? null) ? current : (next?.id ?? null),
		);
	}, [selectedIdentityID, sendableIdentities]);

	useEffect(() => {
		if (retryAvailableAt === null) return;
		const remaining = retryAvailableAt - Date.now();
		if (remaining <= 0) {
			setRetryClock(Date.now());
			return;
		}
		const timer = globalThis.setTimeout(
			() => setRetryClock(Date.now()),
			remaining,
		);
		return () => globalThis.clearTimeout(timer);
	}, [retryAvailableAt]);

	const execute = async (nextIntent: MailReplyDraftIntent): Promise<void> => {
		if (isCreating) return;
		setIsCreating(true);
		setNotice(null);
		try {
			const draft = await createMailReplyDraft(nextIntent);
			setIntent(null);
			setRetryAvailableAt(null);
			openReplyDraft(draft);
		} catch (error) {
			const apiError = mailReplyError(error);
			if (
				apiError?.code === 423 &&
				apiError.error_code === "MAIL_ACCOUNT_BUSY"
			) {
				setIntent(nextIntent);
				setRetryAvailableAt(Date.now() + mailReplyRetryAfterSeconds(error) * 1_000);
				setRetryClock(Date.now());
				setNotice("busy");
			} else if (apiError?.code === 409) {
				setIntent(null);
				setRetryAvailableAt(null);
				setNotice("conflict");
			} else {
				setIntent(null);
				setRetryAvailableAt(null);
				setNotice("unknown");
			}
		} finally {
			setIsCreating(false);
		}
	};

	const start = () => {
		if (selectedIdentity === null || isCreating || intent !== null) return;
		try {
			void execute(createMailReplyDraftIntent(messageID, selectedIdentity.id));
		} catch {
			setNotice("unavailable");
		}
	};

	const retrySeconds =
		retryAvailableAt === null
			? 0
			: Math.max(0, Math.ceil((retryAvailableAt - retryClock) / 1_000));
	const retryReady = intent !== null && retrySeconds === 0 && !isCreating;
	const identitiesUnavailable =
		identitiesQuery.isError ||
		(!identitiesQuery.isPending && sendableIdentities.length === 0);

	return (
		<div className="mt-3 flex flex-wrap items-center gap-2" aria-label={t("reader.reply.label")}>
			{sendableIdentities.length > 1 ? (
				<Select
					value={selectedIdentityID ?? undefined}
					onValueChange={setSelectedIdentityID}
					disabled={isCreating || intent !== null}
				>
					<SelectTrigger className="h-11 max-w-xs" aria-label={t("reader.reply.identity")}>
						<SelectValue placeholder={t("reader.reply.identityLoading")} />
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
			) : null}
			<Button
				variant="outline"
				size="sm"
				className="min-h-11"
				disabled={
					isCreating ||
					identitiesQuery.isPending ||
					identitiesUnavailable ||
					selectedIdentity === null ||
					intent !== null
				}
				onClick={start}
			>
				{t(isCreating ? "reader.reply.creating" : "reader.reply.action")}
			</Button>
			{identitiesQuery.isPending ? (
				<p className="text-xs text-t3">{t("reader.reply.identityLoading")}</p>
			) : null}
			{identitiesUnavailable ? (
				<p className="text-xs text-t2" role="alert">
					{t("reader.reply.identityUnavailable")}
				</p>
			) : null}
			{notice ? (
				<div className="flex flex-wrap items-center gap-2 text-xs text-t2" role="alert">
					<p>
						{notice === "busy"
							? t("reader.reply.busy", { seconds: Math.max(1, retrySeconds) })
							: notice === "conflict"
								? t("reader.reply.conflict")
								: notice === "unavailable"
									? t("reader.reply.unavailable")
									: t("reader.reply.unknown")}
					</p>
					{intent ? (
						<Button
							variant="outline"
							size="sm"
							className="min-h-11"
							disabled={!retryReady}
							onClick={() => {
								if (intent === null || !retryReady) return;
								void execute(intent);
							}}
						>
							{retryReady
								? t("reader.reply.retry")
								: t("reader.reply.retryWait", { seconds: Math.max(1, retrySeconds) })}
						</Button>
					) : null}
				</div>
			) : null}
		</div>
	);
}
