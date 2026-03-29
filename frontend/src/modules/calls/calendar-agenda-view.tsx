import { useMemo } from "react";
import type { CalendarEvent } from "./calendar-types";

// ---------------------------------------------------------------------------
// Agenda View (chronological list of upcoming events)
// ---------------------------------------------------------------------------

export function AgendaView({
	events,
	onEventClick,
}: {
	events: CalendarEvent[];
	onEventClick: (ev: CalendarEvent) => void;
}) {
	// Group events by day, sorted chronologically
	const groupedByDay = useMemo(() => {
		const sorted = [...events].sort(
			(a, b) => a.date.getTime() - b.date.getTime(),
		);
		const groups = new Map<string, CalendarEvent[]>();
		for (const ev of sorted) {
			const key = `${ev.date.getFullYear()}-${ev.date.getMonth()}-${ev.date.getDate()}`;
			const arr = groups.get(key) ?? [];
			arr.push(ev);
			groups.set(key, arr);
		}
		return groups;
	}, [events]);

	if (groupedByDay.size === 0) {
		return (
			<div className="rounded-xl border bg-card p-8 text-center">
				<p className="text-t2">
					Нет запланированных звонков
				</p>
			</div>
		);
	}

	return (
		<div className="space-y-4">
			{Array.from(groupedByDay.entries()).map(([key, dayEvents]) => {
				const dayDate = dayEvents[0]!.date;
				const dayHeader = dayDate.toLocaleDateString("ru-RU", {
					weekday: "long",
					day: "numeric",
					month: "long",
					year: "numeric",
				});

				return (
					<div key={key} className="rounded-xl border bg-card overflow-hidden">
						{/* Day header */}
						<div className="border-b bg-raised/30 px-4 py-2">
							<span className="text-sm font-medium capitalize">
								{dayHeader}
							</span>
						</div>

						{/* Events for this day */}
						<div className="divide-y">
							{dayEvents.map((ev) => (
								<div
									key={ev.id}
									className="flex items-center gap-3 px-4 py-3 cursor-pointer transition-colors hover:bg-raised/50"
									onClick={() => onEventClick(ev)}
								>
									{/* Color dot */}
									<div
										className={`h-3 w-3 shrink-0 rounded-full border ${ev.colorClass}`}
									/>

									{/* Time */}
									<span className="w-14 shrink-0 text-sm font-medium tabular-nums">
										{ev.date.toLocaleTimeString("ru-RU", {
											hour: "2-digit",
											minute: "2-digit",
										})}
									</span>

									{/* Title */}
									<span className="flex-1 truncate text-sm font-medium">
										{ev.title}
									</span>

									{/* Duration badge */}
									<span className="shrink-0 text-xs text-t2">
										{ev.durationMinutes} мин
									</span>
								</div>
							))}
						</div>
					</div>
				);
			})}
		</div>
	);
}
