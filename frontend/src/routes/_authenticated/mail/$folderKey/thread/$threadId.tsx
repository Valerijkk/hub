import { useInfiniteQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { ChevronLeft, Paperclip } from "lucide-react";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import {
	getMailThread,
	type MailAddress,
	type MailThreadDetail,
	type MailThreadMessageSummary,
} from "@/api/mail";
import { Button } from "@/components/ui/button";
import {
	isValidMailEntityID,
	isValidMailFolderKey,
	MailThreadList,
} from "@/modules/mail/mail-thread-list";
import { mailKeys } from "@/modules/mail/mail-query-keys";

const MAIL_THREAD_DETAIL_PAGE_SIZE = 25;

function uniqueMessages(pages: MailThreadDetail[] | undefined) {
	const known = new Set<string>();
	return (pages ?? []).flatMap((page) =>
		page.messages.filter((message) => {
			if (known.has(message.id)) return false;
			known.add(message.id);
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

function MessageCard({ message }: { message: MailThreadMessageSummary }) {
	const { i18n, t } = useTranslation("mail");
	const timestamp = message.received_at ?? message.sent_at;
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

	return (
		<article
			aria-labelledby={`mail-message-${message.id}`}
			className="rounded-[var(--r3)] border border-b1 bg-card p-4"
		>
			<div className="flex min-w-0 items-start justify-between gap-4">
				<div className="min-w-0">
					<h3 id={`mail-message-${message.id}`} className="truncate text-sm font-semibold text-t1">
						{message.subject || t("threadList.noSubject")}
					</h3>
					<p className="mt-1 break-words text-xs text-t2">
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
				</div>
				{timestamp ? (
					<time className="shrink-0 text-xs text-t3" dateTime={timestamp}>
						{formatTimestamp(timestamp, i18n.language)}
					</time>
				) : null}
			</div>

			{badges.length > 0 ? (
				<ul className="mt-3 flex flex-wrap gap-1" aria-label={t("reader.messageStatus")}>
					{badges.map((badge) => (
						<li key={badge} className="rounded-full bg-raised px-2 py-0.5 text-[11px] text-t2">
							{badge}
						</li>
					))}
				</ul>
			) : null}

			<p className="mt-4 whitespace-pre-wrap break-words text-sm leading-6 text-t2">
				{message.snippet || t("reader.noPreview")}
			</p>
			<p className="mt-2 text-xs text-t3">{t("reader.metadataOnly")}</p>

			<div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-t3">
				{message.has_attachments ? (
					<span className="inline-flex items-center gap-1">
						<Paperclip aria-hidden="true" className="h-3.5 w-3.5" />
						{t("threadList.hasAttachments")}
					</span>
				) : null}
				{message.size_bytes > 0 ? (
					<span>
						{t("reader.messageSize", {
							size: formatBytes(message.size_bytes, i18n.language),
						})}
					</span>
				) : null}
				{message.calendar_method ? (
					<span>{t(`reader.calendar.${message.calendar_method}`)}</span>
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
		() => uniqueMessages(detailQuery.data?.pages),
		[detailQuery.data?.pages],
	);
	const thread = detailQuery.data?.pages[0];
	const hasLoadedPages = (detailQuery.data?.pages.length ?? 0) > 0;

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
				<Button variant="outline" size="sm" onClick={() => void detailQuery.refetch()}>
					{t("retry")}
				</Button>
			</div>
		);
	} else if (detailQuery.isError && messages.length === 0) {
		content = (
			<div className="space-y-3 bg-[var(--amber)]/10 p-6" role="alert">
				<p className="text-sm text-t2">{t("reader.partialEmpty")}</p>
				<Button variant="outline" size="sm" onClick={() => void detailQuery.refetch()}>
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
						<Button variant="ghost" size="sm" onClick={() => void detailQuery.refetch()}>
							{t("retry")}
						</Button>
					</div>
				) : null}
				{messages.map((message) => (
					<MessageCard key={message.id} message={message} />
				))}
				<div className="py-2 text-center">
					{detailQuery.hasNextPage ? (
						<Button
							variant="ghost"
							size="sm"
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
				<Button asChild variant="ghost" size="sm" className="mb-2 -ml-2 md:hidden">
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
				<Button asChild variant="outline" className="mt-4">
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
