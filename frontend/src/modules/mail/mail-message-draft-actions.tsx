import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import type { ApiError } from "@/api/client";
import {
	createMailForwardDraft,
	createMailForwardDraftIntent,
	createMailReplyAllDraft,
	createMailReplyAllDraftIntent,
	createMailReplyDraft,
	createMailReplyDraftIntent,
	getAllMailIdentities,
	type MailForwardDraftIntent,
	type MailLinkedDraftKind,
	type MailReplyAllDraftIntent,
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

type MailLinkedDraftIntent =
	| MailReplyDraftIntent
	| MailReplyAllDraftIntent
	| MailForwardDraftIntent;
type MailDraftActionNoticeKind = "busy" | "conflict" | "unavailable" | "unknown";

interface MailDraftActionNotice {
	readonly kind: MailDraftActionNoticeKind;
	readonly intentKind: MailLinkedDraftKind;
}

interface MailMessageDraftActionsProps {
	messageID: string;
}

function mailDraftActionError(error: unknown): Partial<ApiError> | null {
	return typeof error === "object" && error !== null
		? (error as Partial<ApiError>)
		: null;
}

function mailDraftActionRetryAfterSeconds(error: unknown): number {
	const value = mailDraftActionError(error)?.retry_after_seconds;
	return typeof value === "number" && Number.isInteger(value) && value >= 1 && value <= 5
		? value
		: 1;
}

function draftActionLabelKey(kind: MailLinkedDraftKind): string {
	switch (kind) {
		case "reply":
			return "reader.draftActions.reply";
		case "reply_all":
			return "reader.draftActions.replyAll";
		case "forward":
			return "reader.draftActions.forward";
	}
}

function draftActionCreatingKey(kind: MailLinkedDraftKind): string {
	switch (kind) {
		case "reply":
			return "reader.draftActions.creatingReply";
		case "reply_all":
			return "reader.draftActions.creatingReplyAll";
		case "forward":
			return "reader.draftActions.creatingForward";
	}
}

export function MailMessageDraftActions({
	messageID,
}: MailMessageDraftActionsProps) {
	const { t } = useTranslation("mail");
	const { openLinkedDraft } = useMailComposerContext();
	const [selectedIdentityID, setSelectedIdentityID] = useState<string | null>(
		null,
	);
	const [intent, setIntent] = useState<MailLinkedDraftIntent | null>(null);
	const [notice, setNotice] = useState<MailDraftActionNotice | null>(null);
	const [creatingKind, setCreatingKind] = useState<MailLinkedDraftKind | null>(
		null,
	);
	const creatingRef = useRef(false);
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

	const execute = async (nextIntent: MailLinkedDraftIntent): Promise<void> => {
		if (creatingRef.current || creatingKind !== null) return;
		creatingRef.current = true;
		setCreatingKind(nextIntent.kind);
		setNotice(null);
		try {
			const draft = await (async () => {
				switch (nextIntent.kind) {
					case "reply":
						return createMailReplyDraft(nextIntent);
					case "reply_all":
						return createMailReplyAllDraft(nextIntent);
					case "forward":
						return createMailForwardDraft(nextIntent);
				}
			})();
			setIntent(null);
			setRetryAvailableAt(null);
			openLinkedDraft(draft);
		} catch (error) {
			const apiError = mailDraftActionError(error);
			if (
				apiError?.code === 423 &&
				apiError.error_code === "MAIL_ACCOUNT_BUSY"
			) {
				setIntent(nextIntent);
				setRetryAvailableAt(
					Date.now() + mailDraftActionRetryAfterSeconds(error) * 1_000,
				);
				setRetryClock(Date.now());
				setNotice({ kind: "busy", intentKind: nextIntent.kind });
			} else if (apiError?.code === 409) {
				setIntent(null);
				setRetryAvailableAt(null);
				setNotice({ kind: "conflict", intentKind: nextIntent.kind });
			} else {
				setIntent(null);
				setRetryAvailableAt(null);
				setNotice({ kind: "unknown", intentKind: nextIntent.kind });
			}
		} finally {
			creatingRef.current = false;
			setCreatingKind(null);
		}
	};

	const start = (kind: MailLinkedDraftKind) => {
		if (selectedIdentity === null || creatingKind !== null || intent !== null) {
			return;
		}
		try {
			const nextIntent = (() => {
				switch (kind) {
					case "reply":
						return createMailReplyDraftIntent(messageID, selectedIdentity.id);
					case "reply_all":
						return createMailReplyAllDraftIntent(
							messageID,
							selectedIdentity.id,
						);
					case "forward":
						return createMailForwardDraftIntent(messageID, selectedIdentity.id);
				}
			})();
			void execute(nextIntent);
		} catch {
			setNotice({ kind: "unavailable", intentKind: kind });
		}
	};

	const retrySeconds =
		retryAvailableAt === null
			? 0
			: Math.max(0, Math.ceil((retryAvailableAt - retryClock) / 1_000));
	const retryReady = intent !== null && retrySeconds === 0 && creatingKind === null;
	const identitiesUnavailable =
		identitiesQuery.isError ||
		(!identitiesQuery.isPending && sendableIdentities.length === 0);
	const controlsDisabled =
		creatingKind !== null ||
		identitiesQuery.isPending ||
		identitiesUnavailable ||
		selectedIdentity === null ||
		intent !== null;

	return (
		<div
			className="mt-3 flex flex-wrap items-center gap-2"
			role="group"
			aria-label={t("reader.draftActions.label")}
		>
			{sendableIdentities.length > 1 ? (
				<Select
					value={selectedIdentityID ?? undefined}
					onValueChange={setSelectedIdentityID}
					disabled={creatingKind !== null || intent !== null}
				>
					<SelectTrigger
						className="h-11 max-w-xs"
						aria-label={t("reader.draftActions.identity")}
					>
						<SelectValue
							placeholder={t("reader.draftActions.identityLoading")}
						/>
					</SelectTrigger>
					<SelectContent>
						{sendableIdentities.map((identity) => (
							<SelectItem key={identity.id} value={identity.id}>
								{identity.display_name
									? identity.display_name + " <" + identity.address + ">"
									: identity.address}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			) : null}
			{(["reply", "reply_all", "forward"] as const).map((kind) => (
				<Button
					key={kind}
					variant="outline"
					size="sm"
					className="min-h-11"
					disabled={controlsDisabled}
					onClick={() => start(kind)}
				>
					{creatingKind === kind
						? t(draftActionCreatingKey(kind))
						: t(draftActionLabelKey(kind))}
				</Button>
			))}
			{identitiesQuery.isPending ? (
				<p className="text-xs text-t3">
					{t("reader.draftActions.identityLoading")}
				</p>
			) : null}
			{identitiesUnavailable ? (
				<p className="text-xs text-t2" role="alert">
					{t("reader.draftActions.identityUnavailable")}
				</p>
			) : null}
			{notice ? (
				<div
					className="flex flex-wrap items-center gap-2 text-xs text-t2"
					role="alert"
				>
					<p>
						{notice.kind === "busy"
							? t("reader.draftActions.busy", {
									action: t(draftActionLabelKey(notice.intentKind)),
									seconds: Math.max(1, retrySeconds),
								})
							: notice.kind === "conflict"
								? t("reader.draftActions.conflict", {
										action: t(draftActionLabelKey(notice.intentKind)),
									})
								: notice.kind === "unavailable"
									? t("reader.draftActions.unavailable", {
											action: t(draftActionLabelKey(notice.intentKind)),
										})
									: t("reader.draftActions.unknown", {
											action: t(draftActionLabelKey(notice.intentKind)),
										})}
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
								? t("reader.draftActions.retry", {
										action: t(draftActionLabelKey(intent.kind)),
									})
								: t("reader.draftActions.retryWait", {
										action: t(draftActionLabelKey(intent.kind)),
										seconds: Math.max(1, retrySeconds),
									})}
						</Button>
					) : null}
				</div>
			) : null}
		</div>
	);
}
