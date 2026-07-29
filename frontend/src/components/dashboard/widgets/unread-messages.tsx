/**
 * Виджет «Непрочитанные» — счётчик с визуальным индикатором.
 */
import { CheckCheck, MailOpen, MessageCircle } from "lucide-react";
import {
	type DashboardWidgetData,
	isPersonalData,
} from "@/components/dashboard/widget-data";

interface Props {
	data?: DashboardWidgetData;
}

/** Виджет непрочитанных сообщений. */
export function UnreadMessagesWidget({ data }: Props) {
	const personalData = isPersonalData(data) ? data : undefined;
	const count = personalData?.unread_count ?? 0;

	if (count === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-t2">
				<div className="relative">
					<MailOpen className="h-10 w-10 opacity-40" />
					<CheckCheck className="h-4 w-4 absolute -bottom-0.5 -right-0.5 text-[var(--green)]" />
				</div>
				<p className="text-sm font-medium">Все прочитано</p>
				<p className="text-xs opacity-60">Нет новых сообщений</p>
			</div>
		);
	}

	return (
		<div className="flex flex-col items-center justify-center h-full gap-3">
			<div className="relative">
				<MessageCircle className="h-10 w-10 text-[var(--amber)] opacity-80" />
				<div className="absolute -top-1 -right-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-[var(--amber)] px-1">
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
