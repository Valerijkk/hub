import { useEffect, useMemo, useRef } from "react";
import {
	type CalendarEvent,
	HOURS_FULL,
	isSameDay,
} from "./calendar-types";

// ---------------------------------------------------------------------------
// Day View (single day, 24-hour grid)
// ---------------------------------------------------------------------------

export function DayView({
	day,
	today,
	events,
	onSlotClick,
	onEventClick,
}: {
	day: Date;
	today: Date;
	events: CalendarEvent[];
	onSlotClick: (day: Date, hour: number) => void;
	onEventClick: (ev: CalendarEvent) => void;
}) {
	const currentTimeRef = useRef<HTMLDivElement>(null);
	const isToday = isSameDay(day, today);
	const now = new Date();
	const currentHour = now.getHours();
	const currentMinute = now.getMinutes();

	const dayEvents = useMemo(
		() => events.filter((ev) => isSameDay(ev.date, day)),
		[events, day],
	);

	useEffect(() => {
		currentTimeRef.current?.scrollIntoView({
			block: "center",
			behavior: "smooth",
		});
	}, []);

	return (
		<div className="overflow-auto rounded-xl border bg-card">
			<div
				className="grid min-w-[300px]"
				style={{
					gridTemplateColumns: "60px 1fr",
					gridTemplateRows: `repeat(${HOURS_FULL.length}, 60px)`,
				}}
			>
				{HOURS_FULL.map((hour) => {
					const hourEvents = dayEvents.filter(
						(ev) => ev.date.getHours() === hour,
					);
					return (
						<>
							{/* Time label */}
							<div
								key={`label-${hour}`}
								className="relative flex items-start justify-end border-r border-b pr-2 pt-0.5 text-[11px] text-t2"
							>
								{String(hour).padStart(2, "0")}:00
							</div>

							{/* Hour cell */}
							<div
								key={`cell-${hour}`}
								className={`relative border-b cursor-pointer transition-colors hover:bg-raised/50 ${
									isToday ? "bg-primary/[0.02]" : ""
								}`}
								onClick={() => onSlotClick(day, hour)}
							>
								{/* Current time indicator */}
								{isToday && hour === currentHour && (
									<div
										ref={currentTimeRef}
										className="pointer-events-none absolute left-0 right-0 z-10 border-t-2 border-red-500"
										style={{ top: `${(currentMinute / 60) * 100}%` }}
									>
										<div className="absolute -left-1 -top-1.5 h-3 w-3 rounded-full bg-red-500" />
									</div>
								)}

								{/* Event blocks */}
								{hourEvents.map((ev) => {
									const topOffset = (ev.date.getMinutes() / 60) * 100;
									const height = (ev.durationMinutes / 60) * 100;
									return (
										<div
											key={ev.id}
											className={`absolute left-1 right-1 z-[5] overflow-hidden rounded-md border px-2 py-1 text-xs cursor-pointer transition-opacity hover:opacity-80 ${ev.colorClass}`}
											style={{
												top: `${topOffset}%`,
												height: `${Math.min(height, 100 - topOffset)}%`,
												minHeight: "24px",
											}}
											onClick={(e) => {
												e.stopPropagation();
												onEventClick(ev);
											}}
										>
											<span className="font-medium leading-tight line-clamp-2">
												{ev.title}
											</span>
											{ev.durationMinutes >= 30 && (
												<span className="opacity-70 ml-1">
													{ev.date.toLocaleTimeString("ru-RU", {
														hour: "2-digit",
														minute: "2-digit",
													})}
												</span>
											)}
										</div>
									);
								})}
							</div>
						</>
					);
				})}
			</div>
		</div>
	);
}
