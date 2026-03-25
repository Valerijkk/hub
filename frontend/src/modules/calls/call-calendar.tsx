import { useQuery } from "@tanstack/react-query";
import { ChevronLeft, ChevronRight, Plus } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { listScheduledCallsForChannels } from "@/api/calls";
import { getChannels } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import { CreateCallDialog } from "./create-call-dialog";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

/** Scheduled call record from API. */
interface ScheduledCallRecord {
	id: string;
	title: string;
	description?: string;
	scheduled_at: string;
	rrule?: string;
	channel_id: string;
	project_id: string;
	created_by: string;
	is_active: boolean;
}

/** Parsed call event positioned in the calendar grid. */
interface CalendarEvent {
	id: string;
	title: string;
	description?: string;
	date: Date;
	durationMinutes: number;
	channelId: string;
	rrule?: string;
	/** Pastel color class for the block. */
	colorClass: string;
}

type CalendarView = "month" | "week" | "day" | "agenda";

// ---------------------------------------------------------------------------
// Color palette for call types
// ---------------------------------------------------------------------------

const CALL_COLORS = [
	"bg-violet-400/20 border-violet-400/40 text-violet-700 dark:text-violet-300",
	"bg-blue-400/20 border-blue-400/40 text-blue-700 dark:text-blue-300",
	"bg-emerald-400/20 border-emerald-400/40 text-emerald-700 dark:text-emerald-300",
	"bg-amber-400/20 border-amber-400/40 text-amber-700 dark:text-amber-300",
	"bg-rose-400/20 border-rose-400/40 text-rose-700 dark:text-rose-300",
	"bg-cyan-400/20 border-cyan-400/40 text-cyan-700 dark:text-cyan-300",
] as const;


function getColorForEvent(id: string): string {
	let hash = 0;
	for (let i = 0; i < id.length; i++) {
		hash = (hash * 31 + id.charCodeAt(i)) | 0;
	}
	return CALL_COLORS[Math.abs(hash) % CALL_COLORS.length] as string;
}



// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const HOUR_START = 8;
const HOUR_END = 20;
const HOURS = Array.from(
	{ length: HOUR_END - HOUR_START },
	(_, i) => HOUR_START + i,
);
const HOURS_FULL = Array.from({ length: 24 }, (_, i) => i);
const DEFAULT_DURATION = 60; // minutes

const WEEKDAY_NAMES_SHORT = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];

/** Get the Monday of the week containing `date`. */
function getWeekStart(date: Date): Date {
	const d = new Date(date);
	const day = d.getDay();
	const diff = day === 0 ? -6 : 1 - day;
	d.setDate(d.getDate() + diff);
	d.setHours(0, 0, 0, 0);
	return d;
}

/** Get 7 days starting from `weekStart`. */
function getWeekDays(weekStart: Date): Date[] {
	return Array.from({ length: 7 }, (_, i) => {
		const d = new Date(weekStart);
		d.setDate(d.getDate() + i);
		return d;
	});
}

function isSameDay(a: Date, b: Date): boolean {
	return (
		a.getFullYear() === b.getFullYear() &&
		a.getMonth() === b.getMonth() &&
		a.getDate() === b.getDate()
	);
}

function isSameMonth(a: Date, b: Date): boolean {
	return (
		a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth()
	);
}

function formatDayHeader(
	date: Date,
	locale: string,
): { weekday: string; day: string } {
	const weekday = date.toLocaleDateString(locale, { weekday: "short" });
	const day = date.getDate().toString();
	return { weekday, day };
}

/** Get all days to display in a month grid (including padding days from prev/next months). */
function getMonthGridDays(year: number, month: number): Date[] {
	const firstDay = new Date(year, month, 1);
	const lastDay = new Date(year, month + 1, 0);

	// Find Monday before or on the 1st
	let startDay = firstDay.getDay(); // 0=Sun
	// Convert to Mon=0 format
	startDay = startDay === 0 ? 6 : startDay - 1;

	const days: Date[] = [];

	// Add padding days from previous month
	for (let i = startDay - 1; i >= 0; i--) {
		const d = new Date(year, month, 1);
		d.setDate(d.getDate() - i - 1);
		days.push(d);
	}

	// Add all days of the month
	for (let d = 1; d <= lastDay.getDate(); d++) {
		days.push(new Date(year, month, d));
	}

	// Pad to complete the last week (fill to 35 or 42 cells)
	while (days.length % 7 !== 0) {
		const last = days[days.length - 1]!;
		const next = new Date(last);
		next.setDate(next.getDate() + 1);
		days.push(next);
	}

	return days;
}

/**
 * Simple RRULE expansion within a date range.
 * Only handles FREQ=DAILY and FREQ=WEEKLY (with optional BYDAY).
 */
function expandRRule(
	baseDate: Date,
	rrule: string,
	rangeStart: Date,
	rangeEnd: Date,
): Date[] {
	const dates: Date[] = [];
	const isDaily = rrule.includes("FREQ=DAILY");
	const isWeekly = rrule.includes("FREQ=WEEKLY");

	if (!isDaily && !isWeekly) {
		if (baseDate >= rangeStart && baseDate <= rangeEnd) dates.push(baseDate);
		return dates;
	}

	const byDayMatch = rrule.match(/BYDAY=([A-Z,]+)/);
	const byDays = byDayMatch
		? byDayMatch[1]!.split(",").map((d) => {
				const map: Record<string, number> = {
					MO: 1,
					TU: 2,
					WE: 3,
					TH: 4,
					FR: 5,
					SA: 6,
					SU: 0,
				};
				return map[d] ?? -1;
			})
		: null;

	const cursor = new Date(rangeStart);
	cursor.setHours(baseDate.getHours(), baseDate.getMinutes(), 0, 0);

	while (cursor <= rangeEnd) {
		if (cursor >= rangeStart) {
			if (isDaily) {
				if (!byDays || byDays.includes(cursor.getDay())) {
					dates.push(new Date(cursor));
				}
			} else if (isWeekly) {
				if (!byDays || byDays.includes(cursor.getDay())) {
					dates.push(new Date(cursor));
				}
			}
		}
		cursor.setDate(cursor.getDate() + 1);
	}

	return dates;
}

// ---------------------------------------------------------------------------
// Month View
// ---------------------------------------------------------------------------

function MonthView({
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
						className="py-2 text-center text-xs font-medium uppercase text-muted-foreground"
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
							key={idx}
							className={`relative min-h-[80px] border-b border-r p-1.5 cursor-pointer transition-colors hover:bg-muted/50 ${
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
											: "text-foreground"
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
										<div className="px-1 text-[10px] text-muted-foreground">
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

// ---------------------------------------------------------------------------
// Week View (extracted from old CallCalendar)
// ---------------------------------------------------------------------------

function WeekView({
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
		<div className="overflow-auto rounded-xl border bg-card">
			<div
				className="grid min-w-[700px]"
				style={{
					gridTemplateColumns: "60px repeat(7, 1fr)",
					gridTemplateRows: `40px repeat(${HOURS.length}, 60px)`,
				}}
			>
				{/* Header: empty top-left corner */}
				<div className="sticky top-0 z-20 border-b border-r bg-card" />

				{/* Day headers */}
				{weekDays.map((day, i) => {
					const { weekday, day: dayNum } = formatDayHeader(day, "ru-RU");
					const isToday = isSameDay(day, today);
					return (
						<div
							key={i}
							className={`sticky top-0 z-20 flex flex-col items-center justify-center border-b bg-card text-xs ${
								i < 6 ? "border-r" : ""
							} ${isToday ? "bg-primary/5" : ""}`}
						>
							<span className="uppercase text-muted-foreground">
								{weekday}
							</span>
							<span
								className={`text-sm font-semibold ${
									isToday
										? "flex h-6 w-6 items-center justify-center rounded-full bg-primary text-primary-foreground"
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
							className="relative flex items-start justify-end border-r border-b pr-2 pt-0.5 text-[11px] text-muted-foreground"
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
									className={`relative border-b cursor-pointer transition-colors hover:bg-muted/50 ${
										dayIdx < 6 ? "border-r" : ""
									} ${isToday ? "bg-primary/[0.02]" : ""}`}
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
									{dayEvents.map((ev) => {
										const topOffset = (ev.date.getMinutes() / 60) * 100;
										const height = (ev.durationMinutes / 60) * 100;
										return (
											<div
												key={ev.id}
												className={`absolute left-0.5 right-0.5 z-[5] overflow-hidden rounded-md border px-1.5 py-0.5 text-xs cursor-pointer transition-opacity hover:opacity-80 ${ev.colorClass}`}
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
												<span className="font-medium leading-tight line-clamp-2">
													{ev.title}
												</span>
												{ev.durationMinutes >= 30 && (
													<span className="opacity-70">
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

// ---------------------------------------------------------------------------
// Day View (single day, 24-hour grid)
// ---------------------------------------------------------------------------

function DayView({
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
								className="relative flex items-start justify-end border-r border-b pr-2 pt-0.5 text-[11px] text-muted-foreground"
							>
								{String(hour).padStart(2, "0")}:00
							</div>

							{/* Hour cell */}
							<div
								key={`cell-${hour}`}
								className={`relative border-b cursor-pointer transition-colors hover:bg-muted/50 ${
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

// ---------------------------------------------------------------------------
// Agenda View (chronological list of upcoming events)
// ---------------------------------------------------------------------------

function AgendaView({
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
				<p className="text-muted-foreground">
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
						<div className="border-b bg-muted/30 px-4 py-2">
							<span className="text-sm font-medium capitalize">
								{dayHeader}
							</span>
						</div>

						{/* Events for this day */}
						<div className="divide-y">
							{dayEvents.map((ev) => (
								<div
									key={ev.id}
									className="flex items-center gap-3 px-4 py-3 cursor-pointer transition-colors hover:bg-muted/50"
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
									<span className="shrink-0 text-xs text-muted-foreground">
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

// ---------------------------------------------------------------------------
// CallCalendar component
// ---------------------------------------------------------------------------

/**
 * CallCalendar -- calendar view for scheduled calls.
 * Supports month (default) and week views with a view switcher.
 */
export function CallCalendar() {
	const { t } = useTranslation("calls");

	const [view, setView] = useState<CalendarView>("month");
	const [weekOffset, setWeekOffset] = useState(0);
	const [monthOffset, setMonthOffset] = useState(0);
	const [dayOffset, setDayOffset] = useState(0);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [dialogDefaults, setDialogDefaults] = useState<{
		date?: string;
		time?: string;
	}>({});
	const [selectedEvent, setSelectedEvent] = useState<CalendarEvent | null>(
		null,
	);

	const today = useMemo(() => new Date(), []);

	// --- Month view state ---
	const displayMonth = useMemo(() => {
		const d = new Date(today.getFullYear(), today.getMonth() + monthOffset, 1);
		return { year: d.getFullYear(), month: d.getMonth() };
	}, [today, monthOffset]);

	const monthRangeStart = useMemo(() => {
		const days = getMonthGridDays(displayMonth.year, displayMonth.month);
		return days[0]!;
	}, [displayMonth]);

	const monthRangeEnd = useMemo(() => {
		const days = getMonthGridDays(displayMonth.year, displayMonth.month);
		const last = days[days.length - 1]!;
		const end = new Date(last);
		end.setHours(23, 59, 59, 999);
		return end;
	}, [displayMonth]);

	// --- Week view state ---
	const weekStart = useMemo(() => {
		const base = getWeekStart(today);
		base.setDate(base.getDate() + weekOffset * 7);
		return base;
	}, [today, weekOffset]);

	const weekDays = useMemo(() => getWeekDays(weekStart), [weekStart]);

	const weekRangeEnd = useMemo(() => {
		const end = new Date(weekDays[6]!);
		end.setHours(23, 59, 59, 999);
		return end;
	}, [weekDays]);

	// --- Day view state ---
	const displayDay = useMemo(() => {
		const d = new Date(today);
		d.setDate(d.getDate() + dayOffset);
		d.setHours(0, 0, 0, 0);
		return d;
	}, [today, dayOffset]);

	const dayRangeEnd = useMemo(() => {
		const end = new Date(displayDay);
		end.setHours(23, 59, 59, 999);
		return end;
	}, [displayDay]);

	// --- Agenda view state (next 30 days from today) ---
	const agendaRangeStart = useMemo(() => {
		const d = new Date(today);
		d.setHours(0, 0, 0, 0);
		return d;
	}, [today]);

	const agendaRangeEnd = useMemo(() => {
		const d = new Date(today);
		d.setDate(d.getDate() + 30);
		d.setHours(23, 59, 59, 999);
		return d;
	}, [today]);

	// --- Compute visible range based on active view ---
	const rangeStart = useMemo(() => {
		switch (view) {
			case "month": return monthRangeStart;
			case "week": return weekStart;
			case "day": return displayDay;
			case "agenda": return agendaRangeStart;
		}
	}, [view, monthRangeStart, weekStart, displayDay, agendaRangeStart]);

	const rangeEnd = useMemo(() => {
		switch (view) {
			case "month": return monthRangeEnd;
			case "week": return weekRangeEnd;
			case "day": return dayRangeEnd;
			case "agenda": return agendaRangeEnd;
		}
	}, [view, monthRangeEnd, weekRangeEnd, dayRangeEnd, agendaRangeEnd]);

	// Загружаем каналы для получения списка их ID
	const { data: channels } = useQuery({
		queryKey: ["messenger", "channels"],
		queryFn: getChannels,
	});

	const channelIds = useMemo(
		() => (channels ?? []).map((ch) => ch.id),
		[channels],
	);

	// Загружаем запланированные звонки из всех каналов
	const { data: scheduledCallsData } = useQuery({
		queryKey: ["calls", "scheduled", channelIds],
		queryFn: () => listScheduledCallsForChannels(channelIds),
		enabled: channelIds.length > 0,
	});

	const scheduledCalls = useMemo<ScheduledCallRecord[]>(
		() =>
			(scheduledCallsData ?? []).map((c) => ({
				id: c.id,
				title: c.title,
				description: c.description,
				scheduled_at: c.scheduled_at,
				rrule: c.rrule,
				channel_id: c.channel_id,
				project_id: c.project_id,
				created_by: c.created_by,
				is_active: c.is_active,
			})),
		[scheduledCallsData],
	);

	// Expand calls into calendar events for the visible range
	const events = useMemo<CalendarEvent[]>(() => {
		if (!scheduledCalls) return [];

		const result: CalendarEvent[] = [];

		for (const call of scheduledCalls) {
			const baseDate = new Date(call.scheduled_at);
			const colorClass = getColorForEvent(call.id);

			if (call.rrule) {
				const occurrences = expandRRule(
					baseDate,
					call.rrule,
					rangeStart,
					rangeEnd,
				);
				for (const occ of occurrences) {
					result.push({
						id: `${call.id}-${occ.toISOString()}`,
						title: call.title,
						description: call.description,
						date: occ,
						durationMinutes: DEFAULT_DURATION,
						channelId: call.channel_id,
						rrule: call.rrule,
						colorClass,
					});
				}
			} else {
				if (baseDate >= rangeStart && baseDate <= rangeEnd) {
					result.push({
						id: call.id,
						title: call.title,
						description: call.description,
						date: baseDate,
						durationMinutes: DEFAULT_DURATION,
						channelId: call.channel_id,
						colorClass,
					});
				}
			}
		}

		return result;
	}, [scheduledCalls, rangeStart, rangeEnd]);

	// --- Handlers ---

	const handleDayClick = useCallback((day: Date) => {
		const d = new Date(day);
		d.setHours(10, 0, 0, 0);
		setDialogDefaults({
			date: `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`,
			time: "10:00",
		});
		setSelectedEvent(null);
		setDialogOpen(true);
	}, []);

	const handleSlotClick = useCallback((day: Date, hour: number) => {
		const d = new Date(day);
		d.setHours(hour, 0, 0, 0);
		setDialogDefaults({
			date: `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`,
			time: `${String(hour).padStart(2, "0")}:00`,
		});
		setSelectedEvent(null);
		setDialogOpen(true);
	}, []);

	const handleEventClick = useCallback((ev: CalendarEvent) => {
		setSelectedEvent(ev);
	}, []);

	const handleNavigateToday = useCallback(() => {
		if (view === "month") setMonthOffset(0);
		else if (view === "week") setWeekOffset(0);
		else if (view === "day") setDayOffset(0);
	}, [view]);

	const handleNavigatePrev = useCallback(() => {
		if (view === "month") setMonthOffset((m) => m - 1);
		else if (view === "week") setWeekOffset((w) => w - 1);
		else if (view === "day") setDayOffset((d) => d - 1);
	}, [view]);

	const handleNavigateNext = useCallback(() => {
		if (view === "month") setMonthOffset((m) => m + 1);
		else if (view === "week") setWeekOffset((w) => w + 1);
		else if (view === "day") setDayOffset((d) => d + 1);
	}, [view]);

	// --- Header text ---
	const headerText = useMemo(() => {
		if (view === "month") {
			const d = new Date(displayMonth.year, displayMonth.month, 1);
			return d.toLocaleDateString("ru-RU", {
				month: "long",
				year: "numeric",
			});
		}
		if (view === "day") {
			return displayDay.toLocaleDateString("ru-RU", {
				weekday: "long",
				day: "numeric",
				month: "long",
				year: "numeric",
			});
		}
		if (view === "agenda") {
			return "Ближайшие 30 дней";
		}
		return `${weekDays[0]!.toLocaleDateString("ru-RU", {
			day: "numeric",
			month: "long",
		})} \u2014 ${weekDays[6]!.toLocaleDateString("ru-RU", {
			day: "numeric",
			month: "long",
			year: "numeric",
		})}`;
	}, [view, displayMonth, weekDays, displayDay]);

	const isCurrentWeek = weekOffset === 0;

	return (
		<div className="space-y-4">
			{/* Navigation bar */}
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-2">
					{view !== "agenda" && (
						<>
							<Button
								variant="outline"
								size="icon"
								onClick={handleNavigatePrev}
							>
								<ChevronLeft className="h-4 w-4" />
							</Button>
							<Button variant="outline" size="sm" onClick={handleNavigateToday}>
								{t("calendarToday", { defaultValue: "Сегодня" })}
							</Button>
							<Button
								variant="outline"
								size="icon"
								onClick={handleNavigateNext}
							>
								<ChevronRight className="h-4 w-4" />
							</Button>
						</>
					)}

					<span className="ml-2 text-sm font-medium text-foreground capitalize">
						{headerText}
					</span>
				</div>

				<div className="flex items-center gap-2">
					{/* View switcher */}
					<div className="inline-flex h-9 items-center rounded-lg border bg-muted p-1">
						{(
							[
								{ key: "month" as const, label: t("calendarMonth", { defaultValue: "Месяц" }) },
								{ key: "week" as const, label: t("calendarWeek", { defaultValue: "Неделя" }) },
								{ key: "day" as const, label: t("calendarDay", { defaultValue: "День" }) },
								{ key: "agenda" as const, label: t("calendarAgenda", { defaultValue: "Повестка" }) },
							] as const
						).map((tab) => (
							<button
								key={tab.key}
								className={`inline-flex items-center justify-center rounded-md px-3 py-1 text-sm font-medium transition-colors ${
									view === tab.key
										? "bg-background text-foreground shadow-sm"
										: "text-muted-foreground hover:text-foreground"
								}`}
								onClick={() => setView(tab.key)}
							>
								{tab.label}
							</button>
						))}
					</div>

					<Button
						size="sm"
						onClick={() => {
							setDialogDefaults({});
							setSelectedEvent(null);
							setDialogOpen(true);
						}}
					>
						<Plus className="mr-1 h-4 w-4" />
						{t("calendarNewCall", { defaultValue: "Запланировать" })}
					</Button>
				</div>
			</div>

			{/* Calendar views */}
			{view === "month" && (
				<MonthView
					year={displayMonth.year}
					month={displayMonth.month}
					today={today}
					events={events}
					onDayClick={handleDayClick}
					onEventClick={handleEventClick}
				/>
			)}
			{view === "week" && (
				<WeekView
					weekDays={weekDays}
					today={today}
					events={events}
					isCurrentWeek={isCurrentWeek}
					onSlotClick={handleSlotClick}
					onEventClick={handleEventClick}
				/>
			)}
			{view === "day" && (
				<DayView
					day={displayDay}
					today={today}
					events={events}
					onSlotClick={handleSlotClick}
					onEventClick={handleEventClick}
				/>
			)}
			{view === "agenda" && (
				<AgendaView
					events={events}
					onEventClick={handleEventClick}
				/>
			)}

			{/* Event detail popover */}
			{selectedEvent && (
				<div
					className="fixed inset-0 z-40"
					onClick={() => setSelectedEvent(null)}
				>
					<div
						className="absolute left-1/2 top-1/2 z-50 w-80 -translate-x-1/2 -translate-y-1/2 rounded-xl border bg-card p-4 shadow-lg"
						onClick={(e) => e.stopPropagation()}
					>
						<div className="space-y-2">
							<div className="flex items-center justify-between">
								<h3 className="font-semibold">{selectedEvent.title}</h3>
								<button
									onClick={() => setSelectedEvent(null)}
									className="rounded-sm p-0.5 text-muted-foreground hover:text-foreground"
								>
									&times;
								</button>
							</div>
							{selectedEvent.description && (
								<p className="text-sm text-muted-foreground">
									{selectedEvent.description}
								</p>
							)}
							<div className="flex flex-col gap-1 text-sm">
								<span>
									{selectedEvent.date.toLocaleDateString("ru-RU", {
										weekday: "long",
										day: "numeric",
										month: "long",
									})}
								</span>
								<span>
									{selectedEvent.date.toLocaleTimeString("ru-RU", {
										hour: "2-digit",
										minute: "2-digit",
									})}
									{" \u2014 "}
									{new Date(
										selectedEvent.date.getTime() +
											selectedEvent.durationMinutes * 60_000,
									).toLocaleTimeString("ru-RU", {
										hour: "2-digit",
										minute: "2-digit",
									})}
								</span>
								{selectedEvent.rrule && (
									<span className="text-xs text-muted-foreground">
										{selectedEvent.rrule.includes("BYDAY=MO,TU,WE,TH,FR")
											? "По будням"
											: selectedEvent.rrule.includes("FREQ=DAILY")
												? "Ежедневно"
												: selectedEvent.rrule.includes("FREQ=WEEKLY")
													? "Еженедельно"
													: "Повторяющийся"}
									</span>
								)}
							</div>
						</div>
					</div>
				</div>
			)}

			{/* Create call dialog */}
			<CreateCallDialog
				open={dialogOpen}
				onOpenChange={setDialogOpen}
				defaultDate={dialogDefaults.date}
				defaultTime={dialogDefaults.time}
			/>
		</div>
	);
}
