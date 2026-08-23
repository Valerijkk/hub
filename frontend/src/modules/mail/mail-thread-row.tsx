import { Paperclip, Star } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { MailThreadSummary } from "@/api/mail";
import { cn } from "@/lib/utils";

interface MailThreadRowProps {
	thread: MailThreadSummary;
	selected?: boolean;
}

function participantsLabel(participants: MailThreadSummary["participants"]): string {
	const addresses = participants
		.map((participant) => participant.email.trim())
		.filter((address) => address.length > 0);
	return addresses.join(", ");
}

function formatThreadTime(value: string, language: string): string {
	const timestamp = new Date(value);
	if (Number.isNaN(timestamp.getTime())) return "";

	const locale = language.startsWith("ru") ? "ru-RU" : "en-US";
	const now = new Date();
	const isToday = timestamp.toDateString() === now.toDateString();
	if (isToday) {
		return new Intl.DateTimeFormat(locale, {
			hour: "numeric",
			minute: "2-digit",
		}).format(timestamp);
	}

	const sameYear = timestamp.getFullYear() === now.getFullYear();
	return timestamp.toLocaleDateString(
		locale,
		sameYear
			? { month: "short", day: "numeric" }
			: { month: "short", day: "numeric", year: "numeric" },
	);
}

export function MailThreadRow({ thread, selected = false }: MailThreadRowProps) {
	const { i18n, t } = useTranslation("mail");
	const participants = participantsLabel(thread.participants);
	const time = formatThreadTime(thread.last_message_at, i18n.language);
	const unread = thread.unread_count > 0;
	const accessibleParts = [
		participants || t("threadList.unknownSender"),
		thread.subject || t("threadList.noSubject"),
		unread ? t("threadList.unreadCount", { count: thread.unread_count }) : null,
		thread.has_attachments ? t("threadList.hasAttachments") : null,
		thread.is_pinned ? t("threadList.starred") : null,
		thread.message_count > 1
			? t("threadList.messageCount", { count: thread.message_count })
			: null,
	]
		.filter(Boolean)
		.join(", ");

	return (
		<li
			aria-current={selected ? "page" : undefined}
			aria-label={accessibleParts}
			className={cn(
				"group relative min-h-[84px] border-b border-b1 px-3 py-3 outline-none transition-colors focus-visible:z-10 focus-visible:ring-2 focus-visible:ring-[var(--blue)] focus-visible:ring-inset",
				"hover:bg-raised/70",
				selected && "bg-[var(--blue)]/10 hover:bg-[var(--blue)]/15",
			)}
			tabIndex={0}
		>
			<div className="flex min-w-0 items-baseline gap-2">
				<span
					aria-hidden="true"
					className={cn(
						"h-2 w-2 shrink-0 rounded-full",
						unread ? "bg-[var(--blue)]" : "bg-transparent",
					)}
				/>
				<p
					className={cn(
						"min-w-0 flex-1 truncate text-sm text-t1",
						unread && "font-semibold",
					)}
				>
					{participants || t("threadList.unknownSender")}
				</p>
				<time
					className="shrink-0 text-xs tabular-nums text-t3"
					dateTime={thread.last_message_at}
				>
					{time}
				</time>
			</div>

			<div className="mt-1 flex min-w-0 items-center gap-1.5 pl-4">
				<p
					className={cn(
						"min-w-0 flex-1 truncate text-sm text-t2",
						unread && "font-medium text-t1",
					)}
				>
					{thread.subject || t("threadList.noSubject")}
				</p>
				{thread.is_pinned ? (
					<Star
						aria-label={t("threadList.starred")}
						className="h-3.5 w-3.5 shrink-0 fill-[var(--amber)] text-[var(--amber)]"
					/>
				) : null}
				{thread.has_attachments ? (
					<Paperclip
						aria-label={t("threadList.hasAttachments")}
						className="h-3.5 w-3.5 shrink-0 text-t3"
					/>
				) : null}
			</div>

			<div className="mt-1 flex min-w-0 items-center gap-2 pl-4">
				<p className="min-w-0 flex-1 truncate text-xs leading-5 text-t3">
					{thread.snippet}
				</p>
				{unread ? (
					<span
						className="shrink-0 rounded-full bg-[var(--blue)]/10 px-1.5 py-0.5 text-[11px] font-semibold tabular-nums text-[var(--blue)]"
						aria-label={t("threadList.unreadCount", {
							count: thread.unread_count,
						})}
					>
						{thread.unread_count}
					</span>
				) : null}
				{thread.message_count > 1 ? (
					<span
						className="shrink-0 rounded-full bg-raised px-1.5 py-0.5 text-[11px] font-medium tabular-nums text-t2"
						aria-label={t("threadList.messageCount", {
							count: thread.message_count,
						})}
					>
						{thread.message_count}
					</span>
				) : null}
			</div>

			{thread.labels.length > 0 ? (
				<div className="mt-2 flex min-w-0 gap-1 overflow-hidden pl-4" aria-label={t("threadList.labels")}>
					{thread.labels.slice(0, 2).map((label) => (
						<span
							key={label.id}
							className="max-w-28 truncate rounded-full border border-b1 px-1.5 py-0.5 text-[10px] font-medium text-t2"
							style={{ borderColor: label.color }}
						>
							{label.name}
						</span>
					))}
				</div>
			) : null}
		</li>
	);
}
