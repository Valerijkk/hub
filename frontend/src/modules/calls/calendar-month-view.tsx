import { useMemo } from "react";
import {
	type CalendarEvent,
	WEEKDAY_NAMES_SHORT,
	getMonthGridDays,
	isSameDay,
	isSameMonth,
} from "./calendar-types";

// ---------------------------------------------------------------------------
// Month View
// ---------------------------------------------------------------------------

export function MonthView({
	year,
	month,
	today,
	events,
	onDayClick,
	onEventClick,
}: {
	year: number;
	month: number;
	today: Date;
	events: CalendarEvent[];
	onDayClick: (day: Date) => void;
	onEventClick: (ev: CalendarEvent) => void;
}) {
	const gridDays = useMemo(() => getMonthGridDays(year, month), [year, month]);
	const currentMonth = new Date(year, month, 1);

	// Group events by day key for quick lookup
	const eventsByDay = useMemo(() => {
		const map = new Map<string, CalendarEvent[]>();
		for (const ev of events) {
			const key = `${ev.date.getFullYear()}-${ev.date.getMonth()}-${ev.date.getDate()}`;
			const arr = map.get(key) ?? [];
			arr.push(ev);
			map.set(key, arr);
		}
		return map;
	}, [events]);

	return (
		<div className="rounded-xl border bg-card overflow-hidden">
			{/* Weekday headers */}
			<div className="grid grid-cols-7 border-b">
				{WEEKDAY_NAMES_SHORT.map((name) => (
					<div
						key={name}
						className="py-2 text-center text-xs font-medium uppercase text-t2"
					>
						{name}
					</div>
				))}
			</div>

			{/* Day cells */}
			<div className="grid grid-cols-7">
				{gridDays.map((day, idx) => {
					const isToday = isSameDay(day, today);
					const isCurrentMonth = isSameMonth(day, currentMonth);
					const dayKey = `${day.getFullYear()}-${day.getMonth()}-${day.getDate()}`;
					const dayEvents = eventsByDay.get(dayKey) ?? [];

					return (
						<div
							key={dayKey}
							className={`relative min-h-[80px] border-b border-r p-1.5 cursor-pointer transition-colors hover:bg-raised/50 ${
								!isCurrentMonth ? "opacity-40" : ""
							} ${idx % 7 === 6 ? "border-r-0" : ""}`}
							onClick={() => onDayClick(day)}
						>
							{/* Day number */}
							<div className="flex items-center justify-center">
								<span
									className={`inline-flex h-7 w-7 items-center justify-center rounded-full text-sm ${
										isToday
											? "bg-primary text-primary-foreground font-semibold"
											: "text-t1"
									}`}
								>
									{day.getDate()}
								</span>
							</div>

							{/* Event indicators */}
							{dayEvents.length > 0 && (
								<div className="mt-0.5 space-y-0.5">
									{dayEvents.slice(0, 3).map((ev) => (
										<div
											key={ev.id}
											className={`truncate rounded px-1 py-0.5 text-[10px] leading-tight cursor-pointer transition-opacity hover:opacity-80 ${ev.colorClass}`}
											title={`${ev.title} ${ev.date.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" })}`}
											onClick={(e) => {
												e.stopPropagation();
												onEventClick(ev);
											}}
										>
											{ev.date.toLocaleTimeString("ru-RU", {
												hour: "2-digit",
												minute: "2-digit",
											})}{" "}
											{ev.title}
										</div>
									))}
									{dayEvents.length > 3 && (
										<div className="px-1 text-[10px] text-t2">
											+{dayEvents.length - 3}
										</div>
									)}
								</div>
							)}
						</div>
					);
				})}
			</div>
		</div>
	);
}
