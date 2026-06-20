/**
 * Widget "Unread Messages" -- displays unread message count with visual indicator.
 */
import { CheckCheck, MailOpen, MessageCircle } from "lucide-react";
import type { PersonalData } from "@/api/dashboard";

interface Props {
	data?: unknown;
}

/** Unread messages widget. */
export function UnreadMessagesWidget({ data }: Props) {
	const personalData = data as PersonalData | undefined;
	const count = personalData?.unread_count ?? 0;

	if (count === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-t2">
				<div className="relative">
					<MailOpen className="h-10 w-10 opacity-40" />
					<CheckCheck className="h-4 w-4 absolute -bottom-0.5 -right-0.5 text-green-500" />
				</div>
				<p className="text-sm font-medium">Все прочитано</p>
				<p className="text-xs opacity-60">Нет новых сообщений</p>
			</div>
		);
	}

	return (
		<div className="flex flex-col items-center justify-center h-full gap-3">
			<div className="relative">
				<MessageCircle className="h-10 w-10 text-amber-500 opacity-80" />
				<div className="absolute -top-1 -right-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-amber-500 px-1">
					<span className="text-[10px] font-bold text-white">
						{count > 99 ? "99+" : count}
					</span>
				</div>
			</div>
			<div className="text-center">
				<span className="text-3xl font-bold tabular-nums">{count}</span>
				<p className="text-sm text-t2 mt-0.5">непрочитанных сообщений</p>
			</div>
		</div>
	);
}
