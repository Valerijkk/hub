import { useEffect, useRef } from "react";
import { activateOnEnterOrSpace } from "@/lib/keyboard";
import {
	type CalendarEvent,
	formatDayHeader,
	HOURS,
	isSameDay,
} from "./calendar-types";

// ---------------------------------------------------------------------------
// Вид «Неделя»
// ---------------------------------------------------------------------------

export function WeekView({
	weekDays,
	today,
	events,
	isCurrentWeek,
	onSlotClick,
	onEventClick,
}: {
	weekDays: Date[];
	today: Date;
	events: CalendarEvent[];
	isCurrentWeek: boolean;
	onSlotClick: (day: Date, hour: number) => void;
	onEventClick: (ev: CalendarEvent) => void;
}) {
	const currentTimeRef = useRef<HTMLDivElement>(null);

	const now = new Date();
	const currentHour = now.getHours();
	const currentMinute = now.getMinutes();

	useEffect(() => {
		currentTimeRef.current?.scrollIntoView({
			block: "center",
			behavior: "smooth",
		});
	}, []);

	return (
		<div className="flex flex-1 border border-b1 rounded-[var(--r3)] overflow-hidden overflow-x-auto">
			<table className="min-w-[700px] w-full border-collapse">
				<thead>
					<tr>
						<th
							scope="col"
							className="sticky top-0 z-20 w-14 border-b border-r border-b1 bg-base"
						/>
						{weekDays.map((day, i) => {
							const { weekday, day: dayNum } = formatDayHeader(day, "ru-RU");
							const isToday = isSameDay(day, today);
							const dayKey = `${day.getFullYear()}-${day.getMonth()}-${day.getDate()}`;
							return (
								<th
									key={dayKey}
									scope="col"
									className={`h-[52px] border-b border-b1 sticky top-0 bg-base z-[5] ${
										i < 6 ? "border-r border-b1" : ""
									}`}
								>
									<div className="flex flex-col items-center justify-center">
										<span className="text-[10.5px] text-t4 uppercase font-semibold tracking-[0.06em]">
											{weekday}
										</span>
										<span
											className={`text-[14px] font-bold ${
												isToday
													? "bg-purple text-white w-7 h-7 rounded-full flex items-center justify-center shadow-[0_2px_10px_var(--gP)]"
													: ""
											}`}
										>
											{dayNum}
										</span>
									</div>
								</th>
							);
						})}
					</tr>
				</thead>
				<tbody>
					{HOURS.map((hour) => (
						<tr key={hour}>
							<td className="h-[60px] align-top border-r border-b border-b1 px-2 text-[10.5px] text-t4 font-medium">
								{String(hour).padStart(2, "0")}:00
							</td>

							{weekDays.map((day, dayIdx) => {
								const isToday = isCurrentWeek && isSameDay(day, today);
								const dayEvents = events.filter(
									(ev) =>
										isSameDay(ev.date, day) && ev.date.getHours() === hour,
								);
								const cellKey = `cell-${hour}-${day.getFullYear()}-${day.getMonth()}-${day.getDate()}`;

								return (
									<td
										key={cellKey}
										className={`relative h-[60px] border-b border-b1 p-0 align-top ${
											dayIdx < 6 ? "border-r border-b1" : ""
										} ${isToday ? "bg-[rgba(167,139,250,0.015)]" : ""}`}
									>
										<button
											type="button"
											aria-label={`Создать встречу на ${day.toLocaleDateString("ru-RU", { weekday: "short", day: "numeric", month: "long" })} в ${String(hour).padStart(2, "0")}:00`}
											className="absolute inset-0 z-0 h-full w-full cursor-pointer border-0 bg-transparent transition-colors hover:bg-[var(--hover)] focus:outline-none focus:ring-2 focus:ring-inset focus:ring-primary/40"
											onClick={() => onSlotClick(day, hour)}
										/>
										{/* Индикатор текущего времени */}
										{isToday && hour === currentHour && (
											<div
												ref={currentTimeRef}
												className="pointer-events-none absolute left-0 right-0 h-[2px] bg-red z-[4] shadow-[0_0_8px_var(--gR)]"
												style={{ top: `${(currentMinute / 60) * 100}%` }}
											>
												<div className="absolute -left-1 -top-[3px] w-2 h-2 rounded-full bg-red shadow-[0_0_8px_var(--gR)]" />
											</div>
										)}

										{/* Блоки событий */}
										{dayEvents.map((ev) => {
											const topOffset = (ev.date.getMinutes() / 60) * 100;
											const height = (ev.durationMinutes / 60) * 100;
											return (
												<button
													key={ev.id}
													type="button"
													aria-label={`Встреча: ${ev.title}`}
													className={`absolute left-1 right-1 rounded-[var(--r1)] px-2 py-[5px] text-[11.5px] font-semibold cursor-pointer transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 overflow-hidden z-[2] hover:brightness-[1.15] hover:scale-[1.02] hover:z-[3] hover:shadow-[var(--sh2)] focus:outline-none focus:ring-2 focus:ring-primary/50 text-left ${ev.colorClass}`}
													style={{
														top: `${topOffset}%`,
														height: `${Math.min(height, 100 - topOffset)}%`,
														minHeight: "20px",
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
													<span className="overflow-hidden text-ellipsis whitespace-nowrap block">
														{ev.title}
													</span>
													{ev.durationMinutes >= 30 && (
														<span className="text-[10px] opacity-70 font-medium">
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
								);
							})}
						</tr>
					))}
				</tbody>
			</table>
		</div>
	);
}
