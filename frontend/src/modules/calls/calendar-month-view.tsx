import { useMemo } from "react";
import { activateOnEnterOrSpace } from "@/lib/keyboard";
import {
	type CalendarEvent,
	getMonthGridDays,
	isSameDay,
	isSameMonth,
	WEEKDAY_NAMES_SHORT,
} from "./calendar-types";

// ---------------------------------------------------------------------------
// Вид «Месяц»
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

	// Индекс событий по ключу дня для O(1) lookup в сетке
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
			<table className="w-full border-collapse">
				<thead>
					<tr className="border-b">
						{WEEKDAY_NAMES_SHORT.map((name) => (
							<th
								key={name}
								scope="col"
								className="py-2 text-center text-xs font-medium uppercase text-t2"
							>
								{name}
							</th>
						))}
					</tr>
				</thead>
				<tbody>
					{Array.from(
						{ length: Math.ceil(gridDays.length / 7) },
						(_, rowIdx) => {
							const rowStart = gridDays[rowIdx * 7];
							const rowKey = rowStart
								? `${rowStart.getFullYear()}-${rowStart.getMonth()}-${rowStart.getDate()}`
								: `row-${rowIdx}`;
							return (
								<tr key={rowKey}>
									{gridDays
										.slice(rowIdx * 7, rowIdx * 7 + 7)
										.map((day, idx) => {
											const isToday = isSameDay(day, today);
											const isCurrentMonth = isSameMonth(day, currentMonth);
											const dayKey = `${day.getFullYear()}-${day.getMonth()}-${day.getDate()}`;
											const dayEvents = eventsByDay.get(dayKey) ?? [];

											return (
												<td
													key={dayKey}
													className={`relative min-h-[80px] border-b border-r p-0 align-top ${
														!isCurrentMonth ? "opacity-40" : ""
													} ${idx === 6 ? "border-r-0" : ""}`}
												>
													<button
														type="button"
														aria-label={`Выбрать день: ${day.toLocaleDateString("ru-RU", { day: "numeric", month: "long" })}`}
														className="absolute inset-0 z-0 h-full w-full cursor-pointer border-0 bg-transparent p-1.5 text-left transition-colors hover:bg-raised/50 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-primary/40"
														onClick={() => onDayClick(day)}
													/>
													<div className="relative z-[1] pointer-events-none p-1.5">
														{/* Номер дня */}
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

														{/* Индикаторы событий */}
														{dayEvents.length > 0 && (
															<div className="mt-0.5 space-y-0.5 pointer-events-auto">
																{dayEvents.slice(0, 3).map((ev) => (
																	<button
																		key={ev.id}
																		type="button"
																		aria-label={`Встреча: ${ev.title}`}
																		className={`block w-full truncate rounded px-1 py-0.5 text-[10px] leading-tight cursor-pointer transition-opacity hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-primary/50 text-left ${ev.colorClass}`}
																		title={`${ev.title} ${ev.date.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" })}`}
																		onClick={(e) => {
																			e.stopPropagation();
																			onEventClick(ev);
																		}}
																		onKeyDown={(e) => {
																			if (e.key === "Enter" || e.key === " ") {
																				e.stopPropagation();
																			}
																			activateOnEnterOrSpace(() =>
																				onEventClick(ev),
																			)(e);
																		}}
																	>
																		{ev.date.toLocaleTimeString("ru-RU", {
																			hour: "2-digit",
																			minute: "2-digit",
																		})}{" "}
																		{ev.title}
																	</button>
																))}
																{dayEvents.length > 3 && (
																	<div className="px-1 text-[10px] text-t2">
																		+{dayEvents.length - 3}
																	</div>
																)}
															</div>
														)}
													</div>
												</td>
											);
										})}
								</tr>
							);
						},
					)}
				</tbody>
			</table>
		</div>
	);
}
