import { useEffect, useRef } from "react";
import {
	type CalendarEvent,
	HOURS,
	formatDayHeader,
	isSameDay,
} from "./calendar-types";

// ---------------------------------------------------------------------------
// Week View
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
			<div
				className="grid min-w-[700px] w-full"
				style={{
					gridTemplateColumns: "56px repeat(7, 1fr)",
					gridTemplateRows: `52px repeat(${HOURS.length}, 60px)`,
				}}
			>
				{/* Header: empty top-left corner (time column header) */}
				<div className="sticky top-0 z-20 border-b border-r border-b1 bg-base" />

				{/* Day headers */}
				{weekDays.map((day, i) => {
					const { weekday, day: dayNum } = formatDayHeader(day, "ru-RU");
					const isToday = isSameDay(day, today);
					return (
						<div
							key={i}
							className={`h-[52px] flex flex-col items-center justify-center border-b border-b1 sticky top-0 bg-base z-[5] ${
								i < 6 ? "border-r border-b1" : ""
							}`}
						>
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
					);
				})}

				{/* Time rows */}
				{HOURS.map((hour) => (
					<>
						{/* Time label */}
						<div
							key={`label-${hour}`}
							className="h-[60px] flex items-start justify-end px-2 text-[10.5px] text-t4 font-medium -translate-y-[6px] border-r border-b border-b1"
						>
							{String(hour).padStart(2, "0")}:00
						</div>

						{/* Day cells for this hour */}
						{weekDays.map((day, dayIdx) => {
							const isToday = isCurrentWeek && isSameDay(day, today);
							const dayEvents = events.filter(
								(ev) =>
									isSameDay(ev.date, day) && ev.date.getHours() === hour,
							);

							return (
								<div
									key={`cell-${hour}-${dayIdx}`}
									className={`h-[60px] relative border-b border-b1 cursor-pointer transition-colors hover:bg-[var(--hover)] ${
										dayIdx < 6 ? "border-r border-b1" : ""
									} ${isToday ? "bg-[rgba(167,139,250,0.015)]" : ""}`}
									onClick={() => onSlotClick(day, hour)}
								>
									{/* Current time indicator */}
									{isToday && hour === currentHour && (
										<div
											ref={currentTimeRef}
											className="pointer-events-none absolute left-0 right-0 h-[2px] bg-red z-[4] shadow-[0_0_8px_var(--gR)]"
											style={{ top: `${(currentMinute / 60) * 100}%` }}
										>
											<div className="absolute -left-1 -top-[3px] w-2 h-2 rounded-full bg-red shadow-[0_0_8px_var(--gR)]" />
										</div>
									)}

									{/* Event blocks */}
									{dayEvents.map((ev) => {
										const topOffset = (ev.date.getMinutes() / 60) * 100;
										const height = (ev.durationMinutes / 60) * 100;
										return (
											<div
												key={ev.id}
												className={`absolute left-1 right-1 rounded-[var(--r1)] px-2 py-[5px] text-[11.5px] font-semibold cursor-pointer transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 overflow-hidden z-[2] hover:brightness-[1.15] hover:scale-[1.02] hover:z-[3] hover:shadow-[var(--sh2)] ${ev.colorClass}`}
												style={{
													top: `${topOffset}%`,
													height: `${Math.min(height, 100 - topOffset)}%`,
													minHeight: "20px",
												}}
												onClick={(e) => {
													e.stopPropagation();
													onEventClick(ev);
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
											</div>
										);
									})}
								</div>
							);
						})}
					</>
				))}
			</div>
		</div>
	);
}
