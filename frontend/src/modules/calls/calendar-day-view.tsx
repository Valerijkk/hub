import { useEffect, useMemo, useRef } from "react";
import { activateOnEnterOrSpace } from "@/lib/keyboard";
import { type CalendarEvent, HOURS_FULL, isSameDay } from "./calendar-types";

// ---------------------------------------------------------------------------
// Вид «День» — одна сетка на 24 часа
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
			<table className="w-full min-w-[300px] border-collapse">
				<tbody>
					{HOURS_FULL.map((hour) => {
						const hourEvents = dayEvents.filter(
							(ev) => ev.date.getHours() === hour,
						);
						return (
							<tr key={hour}>
								<td className="relative w-[60px] align-top border-r border-b pr-2 pt-0.5 text-[11px] text-t2">
									{String(hour).padStart(2, "0")}:00
								</td>

								<td
									className={`relative border-b p-0 align-top ${
										isToday ? "bg-primary/[0.02]" : ""
									}`}
									style={{ height: "60px" }}
								>
									<button
										type="button"
										aria-label={`Создать встречу на ${String(hour).padStart(2, "0")}:00`}
										className="absolute inset-0 z-0 w-full cursor-pointer border-0 bg-transparent transition-colors hover:bg-raised/50 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-primary/40"
										onClick={() => onSlotClick(day, hour)}
									/>
									{/* Индикатор текущего времени */}
									{isToday && hour === currentHour && (
										<div
											ref={currentTimeRef}
											className="pointer-events-none absolute left-0 right-0 z-10 border-t-2 border-red-500"
											style={{ top: `${(currentMinute / 60) * 100}%` }}
										>
											<div className="absolute -left-1 -top-1.5 h-3 w-3 rounded-full bg-red-500" />
										</div>
									)}

									{/* Блоки событий */}
									{hourEvents.map((ev) => {
										const topOffset = (ev.date.getMinutes() / 60) * 100;
										const height = (ev.durationMinutes / 60) * 100;
										return (
											<button
												key={ev.id}
												type="button"
												aria-label={`Встреча: ${ev.title}`}
												className={`absolute left-1 right-1 z-[5] overflow-hidden rounded-md border px-2 py-1 text-xs cursor-pointer transition-opacity hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-primary/50 text-left ${ev.colorClass}`}
												style={{
													top: `${topOffset}%`,
													height: `${Math.min(height, 100 - topOffset)}%`,
													minHeight: "24px",
												}}
												onClick={(e) => {
													e.stopPropagation();
													onEventClick(ev);
												}}
												onKeyDown={(e) => {
													if (e.key === "Enter" || e.key === " ") {
														e.stopPropagation();
													}
													activateOnEnterOrSpace(() => onEventClick(ev))(e);
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
											</button>
										);
									})}
								</td>
							</tr>
						);
					})}
				</tbody>
			</table>
		</div>
	);
}
