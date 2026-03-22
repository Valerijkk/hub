import { ChevronLeft, ChevronRight, Plus } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
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
const DEFAULT_DURATION = 60; // minutes

/** Get the Monday of the week containing `date`. */
function getWeekStart(date: Date): Date {
	const d = new Date(date);
	const day = d.getDay();
	// Sunday = 0, so shift to make Monday = 0
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

function formatDayHeader(
	date: Date,
	locale: string,
): { weekday: string; day: string } {
	const weekday = date.toLocaleDateString(locale, { weekday: "short" });
	const day = date.getDate().toString();
	return { weekday, day };
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
		// Unknown frequency; just use the base date if it falls in range
		if (baseDate >= rangeStart && baseDate <= rangeEnd) dates.push(baseDate);
		return dates;
	}

	// Parse BYDAY if present
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
// CallCalendar component
// ---------------------------------------------------------------------------

/**
 * CallCalendar -- weekly calendar view for scheduled calls.
 * Shows scheduled and recurring calls as colored blocks in a CSS grid.
 */
export function CallCalendar() {
	const { t } = useTranslation("calls");

	const [weekOffset, setWeekOffset] = useState(0);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [dialogDefaults, setDialogDefaults] = useState<{
		date?: string;
		time?: string;
	}>({});
	const [selectedEvent, setSelectedEvent] = useState<CalendarEvent | null>(
		null,
	);
	const currentTimeRef = useRef<HTMLDivElement>(null);

	const today = useMemo(() => new Date(), []);

	const weekStart = useMemo(() => {
		const base = getWeekStart(today);
		base.setDate(base.getDate() + weekOffset * 7);
		return base;
	}, [today, weekOffset]);

	const weekDays = useMemo(() => getWeekDays(weekStart), [weekStart]);

	const rangeEnd = useMemo(() => {
		const end = new Date(weekDays[6]!);
		end.setHours(23, 59, 59, 999);
		return end;
	}, [weekDays]);

	// Scheduled calls — пока нет глобального endpoint, используем пустой массив.
	// TODO: подключить к реальному API когда будет endpoint /api/calls/scheduled
	const scheduledCalls: ScheduledCallRecord[] = [];

	// Expand calls into calendar events for the current week
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
					weekStart,
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
				if (baseDate >= weekStart && baseDate <= rangeEnd) {
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
	}, [scheduledCalls, weekStart, rangeEnd]);

	// Scroll current time into view on mount
	useEffect(() => {
		currentTimeRef.current?.scrollIntoView({
			block: "center",
			behavior: "smooth",
		});
	}, []);

	const handleSlotClick = useCallback((day: Date, hour: number) => {
		const d = new Date(day);
		d.setHours(hour, 0, 0, 0);
		setDialogDefaults({
			date: d.toISOString().split("T")[0],
			time: `${String(hour).padStart(2, "0")}:00`,
		});
		setSelectedEvent(null);
		setDialogOpen(true);
	}, []);

	const handleEventClick = useCallback((ev: CalendarEvent) => {
		setSelectedEvent(ev);
	}, []);

	const now = new Date();
	const currentHour = now.getHours();
	const currentMinute = now.getMinutes();
	const isCurrentWeek = weekOffset === 0;

	return (
		<div className="space-y-4">
			{/* Week navigation */}
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-2">
					<Button
						variant="outline"
						size="icon"
						onClick={() => setWeekOffset((w) => w - 1)}
					>
						<ChevronLeft className="h-4 w-4" />
					</Button>
					<Button variant="outline" size="sm" onClick={() => setWeekOffset(0)}>
						{t("calendarToday", { defaultValue: "Сегодня" })}
					</Button>
					<Button
						variant="outline"
						size="icon"
						onClick={() => setWeekOffset((w) => w + 1)}
					>
						<ChevronRight className="h-4 w-4" />
					</Button>
				</div>

				<span className="text-sm font-medium text-muted-foreground">
					{weekDays[0]!.toLocaleDateString("ru-RU", {
						day: "numeric",
						month: "long",
					})}
					{" \u2014 "}
					{weekDays[6]!.toLocaleDateString("ru-RU", {
						day: "numeric",
						month: "long",
						year: "numeric",
					})}
				</span>

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

			{/* Calendar grid */}
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
										onClick={() => handleSlotClick(day, hour)}
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
														handleEventClick(ev);
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
