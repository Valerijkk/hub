import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import {
	CalendarDays,
	ChevronDown,
	ChevronLeft,
	ImageOff,
	Paperclip,
} from "lucide-react";
import { useCallback, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
	getMailMessage,
	getMailThread,
	type MailAddress,
	type MailMessageDetail,
	type MailThreadDetail,
	type MailThreadMessageSummary,
} from "@/api/mail";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
	isValidMailEntityID,
	isValidMailFolderKey,
	MailThreadList,
} from "@/modules/mail/mail-thread-list";
import { MailMessageActions } from "@/modules/mail/mail-message-actions";
import { MailMessageReplyAction } from "@/modules/mail/mail-message-reply-action";
import { mailKeys } from "@/modules/mail/mail-query-keys";

const MAIL_THREAD_DETAIL_PAGE_SIZE = 25;

function uniqueMessages(
	pages: MailThreadDetail[] | undefined,
	threadId: string,
) {
	const known = new Set<string>();
	return (pages ?? []).flatMap((page) =>
		page.messages.filter((message) => {
			const messageID = message.id.toLowerCase();
			if (
				!isValidMailEntityID(messageID) ||
				!isValidMailEntityID(message.thread_id) ||
				message.thread_id.toLowerCase() !== threadId ||
				known.has(messageID)
			) {
				return false;
			}
			known.add(messageID);
			return true;
		}),
	);
}

function formatAddress(address: MailAddress): string {
	return address.name?.trim()
		? `${address.name.trim()} <${address.email}>`
		: address.email;
}

function formatAddresses(addresses: MailAddress[]): string {
	return addresses.map(formatAddress).join(", ");
}

function formatTimestamp(value: string | null, language: string): string {
	if (!value) return "";
	const timestamp = new Date(value);
	if (Number.isNaN(timestamp.getTime())) return "";
	return timestamp.toLocaleString(language.startsWith("ru") ? "ru-RU" : "en-US", {
		dateStyle: "medium",
		timeStyle: "short",
	});
}

function formatBytes(value: number, language: string): string {
	if (!Number.isFinite(value) || value < 0) return "";
	const units = ["B", "KB", "MB", "GB"];
	let size = value;
	let unit = 0;
	while (size >= 1024 && unit < units.length - 1) {
		size /= 1024;
		unit += 1;
	}
	return `${new Intl.NumberFormat(language.startsWith("ru") ? "ru-RU" : "en-US", {
		maximumFractionDigits: unit === 0 ? 0 : 1,
	}).format(size)} ${units[unit]}`;
}

interface MessageCardProps {
	message: MailThreadMessageSummary;
	expanded: boolean;
	onToggle: () => void;
	onMovedOrTrashed: (
		action: "move" | "trash",
		folderID: string | null,
	) => void | Promise<void>;
}

function messageDetailMatchesSummary(
	detail: MailMessageDetail,
	message: MailThreadMessageSummary,
): boolean {
	return (
		isValidMailEntityID(detail.id) &&
		isValidMailEntityID(detail.thread_id) &&
		detail.id.toLowerCase() === message.id.toLowerCase() &&
		detail.thread_id.toLowerCase() === message.thread_id.toLowerCase()
	);
}

function MessageCard({
	message,
	expanded,
	onToggle,
	onMovedOrTrashed,
}: MessageCardProps) {
	const { i18n, t } = useTranslation("mail");
	const timestamp = message.received_at ?? message.sent_at;
	const messageID = message.id.toLowerCase();
	const bodyQuery = useQuery({
		queryKey: mailKeys.message(messageID),
		queryFn: () => getMailMessage(messageID),
		enabled: expanded,
		retry: false,
		refetchOnWindowFocus: false,
	});
	const detail =
		bodyQuery.data && messageDetailMatchesSummary(bodyQuery.data, message)
			? bodyQuery.data
			: undefined;
	const badges = [
		message.is_draft ? t("reader.badges.draft") : null,
		!message.seen ? t("reader.badges.unread") : null,
		message.flagged ? t("reader.badges.flagged") : null,
		message.answered ? t("reader.badges.answered") : null,
		message.junk ? t("reader.badges.junk") : null,
		message.priority !== "normal"
			? t(`reader.badges.priority.${message.priority}`)
			: null,
	].filter((badge): badge is string => typeof badge === "string");
	const triggerID = `mail-message-trigger-${messageID}`;
	const panelID = `mail-message-panel-${messageID}`;
	const attachmentsHeadingID = `mail-message-attachments-${messageID}`;

	let bodyContent;
	if (bodyQuery.isPending) {
		bodyContent = (
			<div className="space-y-3 py-4" aria-busy="true" aria-live="polite">
				<p className="text-sm text-t2">{t("reader.body.loading")}</p>
				<div className="h-4 w-4/5 rounded animate-shimmer" />
				<div className="h-4 w-3/5 rounded animate-shimmer" />
			</div>
		);
	} else if (!detail) {
		bodyContent = (
			<div className="space-y-3 py-4" role="alert">
				<p className="text-sm text-t2">{t("reader.body.unavailable")}</p>
				<Button
					variant="outline"
					size="sm"
					className="min-h-11"
					onClick={() => void bodyQuery.refetch()}
				>
					{t("retry")}
				</Button>
			</div>
		);
	} else {
		const calendarEventID =
			detail.calendar_event_id && isValidMailEntityID(detail.calendar_event_id)
				? detail.calendar_event_id.toLowerCase()
				: null;
		bodyContent = (
			<div className="space-y-4 py-4">
				{bodyQuery.isError ? (
					<div className="flex flex-wrap items-center justify-between gap-3 rounded-[var(--r2)] bg-[var(--amber)]/10 p-3" role="alert">
						<p className="text-xs text-t2">{t("reader.body.partialUnavailable")}</p>
						<Button
							variant="ghost"
							size="sm"
							className="min-h-11 shrink-0"
							onClick={() => void bodyQuery.refetch()}
						>
							{t("retry")}
						</Button>
					</div>
				) : null}
				{detail.remote_images_blocked ? (
					<div className="flex items-start gap-2 rounded-[var(--r2)] bg-raised px-3 py-2 text-xs text-t2">
						<ImageOff aria-hidden="true" className="mt-0.5 h-4 w-4 shrink-0" />
						<p>{t("reader.body.remoteImagesBlocked")}</p>
					</div>
				) : null}
				<div className="rounded-[var(--r2)] bg-base/60 p-4">
					{detail.plain_body.length > 0 ? (
						<div className="whitespace-pre-wrap break-words text-sm leading-7 text-t1">
							{detail.plain_body}
						</div>
					) : (
						<p className="text-sm text-t3">{t("reader.body.plainUnavailable")}</p>
					)}
					{detail.plain_body.length === 0 &&
					detail.sanitized_html !== null &&
					detail.sanitized_html.length > 0 ? (
						<p className="mt-2 text-xs leading-5 text-t3">
							{t("reader.body.richUnavailable")}
						</p>
					) : null}
				</div>
				{detail.attachments.length > 0 ? (
					<section aria-labelledby={attachmentsHeadingID}>
						<h4 id={attachmentsHeadingID} className="text-xs font-semibold uppercase tracking-[0.06em] text-t3">
							{t("reader.body.attachments", { count: detail.attachments.length })}
						</h4>
						<ul className="mt-2 space-y-2">
							{detail.attachments.map((attachment) => (
								<li key={attachment.id} className="flex flex-wrap items-center gap-3 rounded-[var(--r2)] border border-b1 bg-card px-3 py-3">
									<Paperclip aria-hidden="true" className="h-4 w-4 shrink-0 text-t3" />
									<div className="min-w-0 flex-1">
										<p className="break-words text-sm font-medium text-t1">{attachment.filename}</p>
										<p className="mt-0.5 text-xs text-t3">
											{attachment.media_type} · {formatBytes(attachment.size_bytes, i18n.language)} · {t(`reader.body.disposition.${attachment.disposition}`)}
										</p>
									</div>
									<span className="rounded-full bg-raised px-2 py-1 text-[11px] text-t3">
										{t("reader.body.downloadUnavailable")}
									</span>
								</li>
							))}
						</ul>
					</section>
				) : null}
				{calendarEventID ? (
					<Button asChild variant="outline" className="min-h-11">
						<Link
							to="/calendar/events/$eventId"
							params={{ eventId: calendarEventID }}
						>
							<CalendarDays aria-hidden="true" className="h-4 w-4" />
							{t("reader.body.openCalendar")}
						</Link>
					</Button>
				) : null}
			</div>
		);
	}

	return (
		<article
			aria-labelledby={triggerID}
			className={cn(
				"overflow-hidden rounded-[var(--r3)] border border-b1 bg-card transition-[border-color,box-shadow]",
				expanded && "border-[var(--accent)]/60 shadow-sm",
			)}
		>
			<h3>
				<button
					id={triggerID}
					type="button"
					aria-expanded={expanded}
					aria-controls={panelID}
					className="flex min-h-11 w-full items-start gap-3 px-4 py-3 text-left outline-none transition-colors hover:bg-raised/60 focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-[var(--blue)]"
					onClick={onToggle}
				>
					<span className="min-w-0 flex-1">
						<span className="block truncate text-sm font-semibold text-t1">
							{message.subject || t("threadList.noSubject")}
						</span>
						<span className="mt-1 block truncate text-xs text-t2">
							{formatAddress(message.from)}
						</span>
						<span className="mt-2 block truncate text-sm font-normal text-t2">
							{message.snippet || t("reader.noPreview")}
						</span>
					</span>
					<span className="flex shrink-0 flex-col items-end gap-2 text-xs font-normal text-t3">
						{timestamp ? (
							<time dateTime={timestamp}>
								{formatTimestamp(timestamp, i18n.language)}
							</time>
						) : null}
						<span className="inline-flex items-center gap-1">
							{t(expanded ? "reader.body.collapse" : "reader.body.expand")}
							<ChevronDown
								aria-hidden="true"
								className={cn("h-4 w-4 transition-transform", expanded && "rotate-180")}
							/>
						</span>
					</span>
				</button>
			</h3>
			<div id={panelID} role="region" aria-labelledby={triggerID} hidden={!expanded}>
				{expanded ? (
					<div className="px-4 pb-4">
						<div className="border-t border-b1 pt-3">
							<p className="break-words text-xs text-t2">
								<span className="font-medium text-t1">{t("reader.from")}:</span>{" "}
								{formatAddress(message.from)}
							</p>
							{message.to.length > 0 ? (
								<p className="mt-1 break-words text-xs text-t3">
									<span className="font-medium text-t2">{t("reader.to")}:</span>{" "}
									{formatAddresses(message.to)}
								</p>
							) : null}
							{message.cc.length > 0 ? (
								<p className="mt-1 break-words text-xs text-t3">
									<span className="font-medium text-t2">{t("reader.cc")}:</span>{" "}
									{formatAddresses(message.cc)}
								</p>
							) : null}
							{badges.length > 0 ? (
								<ul className="mt-3 flex flex-wrap gap-1" aria-label={t("reader.messageStatus")}>
									{badges.map((badge) => (
										<li key={badge} className="rounded-full bg-raised px-2 py-0.5 text-[11px] text-t2">
											{badge}
										</li>
									))}
								</ul>
							) : null}
							<div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-t3">
								{message.has_attachments ? (
									<span className="inline-flex items-center gap-1">
										<Paperclip aria-hidden="true" className="h-3.5 w-3.5" />
										{t("threadList.hasAttachments")}
									</span>
								) : null}
								{message.size_bytes > 0 ? (
									<span>{t("reader.messageSize", { size: formatBytes(message.size_bytes, i18n.language) })}</span>
								) : null}
								{message.calendar_method ? (
									<span>{t(`reader.calendar.${message.calendar_method}`)}</span>
								) : null}
							</div>
							<MailMessageReplyAction messageID={messageID} />
							<MailMessageActions
								message={message}
								threadID={message.thread_id.toLowerCase()}
								onMovedOrTrashed={onMovedOrTrashed}
							/>
						</div>
						{bodyContent}
					</div>
				) : null}
			</div>
		</article>
	);
}

function ThreadReader({
	folderKey,
	threadId,
}: {
	folderKey: string;
	threadId: string;
}) {
	const { t } = useTranslation("mail");
	const navigate = useNavigate();
	const [selection, setSelection] = useState<{
		threadId: string;
		messageId: string;
	} | null>(null);
	const detailQuery = useInfiniteQuery({
		queryKey: mailKeys.thread(threadId, {
			limit: MAIL_THREAD_DETAIL_PAGE_SIZE,
		}),
		queryFn: ({ pageParam }) =>
			getMailThread(threadId, {
				limit: MAIL_THREAD_DETAIL_PAGE_SIZE,
				cursor: pageParam,
			}),
		initialPageParam: undefined as string | undefined,
		getNextPageParam: (
			lastPage,
			_allPages,
			_lastPageParam,
			allPageParams,
		) => {
			const nextCursor = lastPage.next_cursor ?? undefined;
			return nextCursor && !allPageParams.includes(nextCursor)
				? nextCursor
				: undefined;
		},
		retry: false,
		refetchOnWindowFocus: false,
	});
	const messages = useMemo(
		() => uniqueMessages(detailQuery.data?.pages, threadId),
		[detailQuery.data?.pages, threadId],
	);
	const selectedMessageID =
		selection?.threadId === threadId ? selection.messageId : null;
	const thread = detailQuery.data?.pages[0];
	const hasLoadedPages = (detailQuery.data?.pages.length ?? 0) > 0;
	const navigateAfterRelocation = useCallback(
		async (action: "move" | "trash", destinationFolderID: string | null) => {
			const targetFolderKey =
				action === "move" &&
				destinationFolderID !== null &&
				isValidMailEntityID(destinationFolderID)
					? destinationFolderID.toLowerCase()
					: folderKey;
			await navigate({
				to: "/mail/$folderKey",
				params: { folderKey: targetFolderKey },
				replace: true,
			});
		},
		[folderKey, navigate],
	);

	let content;
	if (detailQuery.isPending) {
		content = (
			<div className="space-y-3 p-4" aria-busy="true" aria-live="polite">
				<p className="sr-only">{t("reader.loading")}</p>
				<div className="h-5 w-2/3 rounded animate-shimmer" />
				<div className="h-28 w-full rounded-[var(--r3)] animate-shimmer" />
				<div className="h-28 w-full rounded-[var(--r3)] animate-shimmer" />
			</div>
		);
	} else if ((detailQuery.isError || !thread) && !hasLoadedPages) {
		content = (
			<div className="space-y-3 p-6" role="alert">
				<p className="text-sm text-t2">{t("reader.unavailable")}</p>
				<Button variant="outline" size="sm" className="min-h-11" onClick={() => void detailQuery.refetch()}>
					{t("retry")}
				</Button>
			</div>
		);
	} else if (detailQuery.isError && messages.length === 0) {
		content = (
			<div className="space-y-3 bg-[var(--amber)]/10 p-6" role="alert">
				<p className="text-sm text-t2">{t("reader.partialEmpty")}</p>
				<Button variant="outline" size="sm" className="min-h-11" onClick={() => void detailQuery.refetch()}>
					{t("retry")}
				</Button>
			</div>
		);
	} else if (messages.length === 0) {
		content = (
			<div className="p-8 text-center">
				<p className="text-sm font-medium text-t2">{t("reader.noMessagesTitle")}</p>
				<p className="mt-1 text-xs leading-5 text-t3">
					{t("reader.noMessagesDescription")}
				</p>
			</div>
		);
	} else {
		content = (
			<div className="space-y-3 p-4">
				{detailQuery.isError ? (
					<div className="flex items-center justify-between gap-3 rounded-[var(--r2)] bg-[var(--amber)]/10 p-3" role="alert">
						<p className="text-xs text-t2">{t("reader.partialUnavailable")}</p>
						<Button variant="ghost" size="sm" className="min-h-11 shrink-0" onClick={() => void detailQuery.refetch()}>
							{t("retry")}
						</Button>
					</div>
				) : null}
				{messages.map((message) => {
					const messageID = message.id.toLowerCase();
					return (
						<MessageCard
							key={messageID}
							message={message}
							expanded={selectedMessageID === messageID}
							onToggle={() =>
								setSelection((current) =>
									current?.threadId === threadId && current.messageId === messageID
										? null
										: { threadId, messageId: messageID },
								)
							}
							onMovedOrTrashed={navigateAfterRelocation}
						/>
					);
				})}
				<div className="py-2 text-center">
					{detailQuery.hasNextPage ? (
						<Button
							variant="ghost"
							size="sm"
							className="min-h-11"
							disabled={detailQuery.isFetchingNextPage}
							onClick={() => void detailQuery.fetchNextPage()}
						>
							{detailQuery.isFetchingNextPage
								? t("reader.loadingMore")
								: t("reader.loadMore")}
						</Button>
					) : (
						<p className="text-xs text-t3">{t("reader.end")}</p>
					)}
				</div>
			</div>
		);
	}

	return (
		<section className="min-w-0 flex-1 bg-base/30" aria-labelledby="mail-thread-reader-heading">
			<header className="border-b border-b1 bg-card px-4 py-3">
				<Button asChild variant="ghost" size="sm" className="mb-2 -ml-2 min-h-11 md:hidden">
					<Link to="/mail/$folderKey" params={{ folderKey }}>
						<ChevronLeft aria-hidden="true" className="h-4 w-4" />
						{t("reader.back")}
					</Link>
				</Button>
				<h2 id="mail-thread-reader-heading" className="truncate text-lg font-semibold text-t1">
					{thread
						? thread.subject || t("threadList.noSubject")
						: t("reader.title")}
				</h2>
				{thread ? (
					<p className="mt-1 truncate text-xs text-t3">
						{thread.participants.map((participant) => participant.email).join(", ")}
					</p>
				) : null}
			</header>
			{content}
		</section>
	);
}

function InvalidThread() {
	const { t } = useTranslation("mail");
	return (
		<section className="flex min-h-[20rem] items-center justify-center rounded-[var(--r3)] border border-b1 bg-card p-6 text-center" role="alert">
			<div className="max-w-sm">
				<h2 className="text-lg font-semibold text-t1">{t("reader.invalidTitle")}</h2>
				<p className="mt-2 text-sm leading-6 text-t2">{t("reader.invalidDescription")}</p>
				<Button asChild variant="outline" className="mt-4 min-h-11">
					<Link to="/mail">{t("folder.backToAll")}</Link>
				</Button>
			</div>
		</section>
	);
}

function ThreadRoute() {
	const { folderKey, threadId } = Route.useParams();
	if (!isValidMailFolderKey(folderKey) || !isValidMailEntityID(threadId)) {
		return <InvalidThread />;
	}

	const normalizedFolderKey =
		folderKey === "all" ? "all" : folderKey.toLowerCase();
	const normalizedThreadID = threadId.toLowerCase();
	return (
		<div className="flex min-h-[32rem] min-w-0 overflow-hidden rounded-[var(--r3)] border border-b1 bg-card">
			<MailThreadList
				folderKey={normalizedFolderKey}
				selectedThreadId={normalizedThreadID}
				className="hidden md:block"
			/>
			<ThreadReader folderKey={normalizedFolderKey} threadId={normalizedThreadID} />
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/mail/$folderKey/thread/$threadId",
)({ component: ThreadRoute });
