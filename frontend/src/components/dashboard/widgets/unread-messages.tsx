/**
 * Виджет "Непрочитанные сообщения" -- отображает количество непрочитанных сообщений.
 */
import type { PersonalData } from "@/api/dashboard";

interface Props {
	data?: unknown;
}

/** Виджет непрочитанных сообщений. */
export function UnreadMessagesWidget({ data }: Props) {
	const personalData = data as PersonalData | undefined;
	const count = personalData?.unread_count ?? 0;

	return (
		<div className="flex flex-col items-center justify-center h-full">
			<span className="text-4xl font-bold">{count}</span>
			<span className="text-sm text-muted-foreground mt-1">
				{count === 0 ? "Все прочитано" : "непрочитанных"}
			</span>
		</div>
	);
}
