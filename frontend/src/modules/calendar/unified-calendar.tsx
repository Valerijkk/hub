import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { ChevronLeft, ChevronRight, Download } from "lucide-react";
import { useCallback, useMemo, useState } from "react";
import { toast } from "sonner";
import {
	type CalendarEventDTO,
	downloadCalendarExport,
	listCalendarEvents,
} from "@/api/calendar";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { AgendaView } from "@/modules/calls/calendar-agenda-view";
import { DayView } from "@/modules/calls/calendar-day-view";
import { MonthView } from "@/modules/calls/calendar-month-view";
import {
	type CalendarEvent,
	type CalendarEventKind,
	type CalendarView,
	getColorForKind,
	getMonthGridDays,
	getWeekDays,
	getWeekStart,
} from "@/modules/calls/calendar-types";
import { WeekView } from "@/modules/calls/calendar-week-view";

const KIND_LABELS: Record<CalendarEventKind, string> = {
	call: "Созвоны",
	sprint: "Спринты",
	deadline: "Дедлайны",
	leave: "Отпуска",
};

function toRFC3339(d: Date): string {
	return d.toISOString();
}

function mapApiEvent(ev: CalendarEventDTO): CalendarEvent {
	const date = new Date(ev.starts_at);
	let durationMinutes = 60;
	if (ev.ends_at && !ev.all_day) {
		const end = new Date(ev.ends_at);
		durationMinutes = Math.max(
			15,
			Math.round((end.getTime() - date.getTime()) / 60_000),
		);
	}
	return {
		id: ev.id,
		title: ev.title,
		description: ev.description,
		date,
		durationMinutes,
		channelId: "",
		colorClass: getColorForKind(ev.type),
		kind: ev.type,
		url: ev.url,
		projectKey: ev.project_key,
	};
}

function expandEvents(items: CalendarEventDTO[]): CalendarEvent[] {
	const out: CalendarEvent[] = [];
	for (const ev of items) {
		if (ev.type === "sprint" && ev.ends_at) {
			const start = new Date(ev.starts_at);
			start.setHours(0, 0, 0, 0);
			const end = new Date(ev.ends_at);
			end.setHours(0, 0, 0, 0);
			const cursor = new Date(start);
			while (cursor <= end) {
				const base = mapApiEvent(ev);
				out.push({
					...base,
					id: `${ev.id}-${cursor.toISOString().slice(0, 10)}`,
					date: new Date(cursor),
				});
				cursor.setDate(cursor.getDate() + 1);
			}
			continue;
		}
		out.push(mapApiEvent(ev));
	}
	return out;
}

/** MOD-04: единый календарь — созвоны, спринты, дедлайны. */
export function UnifiedCalendar() {
	const navigate = useNavigate();
	const [view, setView] = useState<CalendarView>("month");
	const [weekOffset, setWeekOffset] = useState(0);
	const [monthOffset, setMonthOffset] = useState(0);
	const [dayOffset, setDayOffset] = useState(0);
	const [selectedEvent, setSelectedEvent] = useState<CalendarEvent | null>(
		null,
	);
	const [filters, setFilters] = useState<Record<CalendarEventKind, boolean>>({
		call: true,
		sprint: true,
		deadline: true,
		leave: true,
	});

	const today = useMemo(() => new Date(), []);

	const displayMonth = useMemo(() => {
		const d = new Date(today.getFullYear(), today.getMonth() + monthOffset, 1);
		return { year: d.getFullYear(), month: d.getMonth() };
	}, [today, monthOffset]);

	const monthRangeStart = useMemo(() => {
		const days = getMonthGridDays(displayMonth.year, displayMonth.month);
		return days[0] ?? new Date(displayMonth.year, displayMonth.month, 1);
	}, [displayMonth]);

	const monthRangeEnd = useMemo(() => {
		const days = getMonthGridDays(displayMonth.year, displayMonth.month);
		const last = days[days.length - 1] ?? new Date();
		const end = new Date(last);
		end.setHours(23, 59, 59, 999);
		return end;
	}, [displayMonth]);

	const weekStart = useMemo(() => {
		const base = getWeekStart(today);
		base.setDate(base.getDate() + weekOffset * 7);
		return base;
	}, [today, weekOffset]);

	const weekDays = useMemo(() => getWeekDays(weekStart), [weekStart]);

	const weekRangeEnd = useMemo(() => {
		const end = new Date(weekDays[6] ?? weekStart);
		end.setHours(23, 59, 59, 999);
		return end;
	}, [weekDays, weekStart]);

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

	const rangeStart = useMemo(() => {
		switch (view) {
			case "month":
				return monthRangeStart;
			case "week":
				return weekStart;
			case "day":
				return displayDay;
			case "agenda":
				return agendaRangeStart;
		}
	}, [view, monthRangeStart, weekStart, displayDay, agendaRangeStart]);

	const rangeEnd = useMemo(() => {
		switch (view) {
			case "month":
				return monthRangeEnd;
			case "week":
				return weekRangeEnd;
			case "day":
				return dayRangeEnd;
			case "agenda":
				return agendaRangeEnd;
		}
	}, [view, monthRangeEnd, weekRangeEnd, dayRangeEnd, agendaRangeEnd]);

	const fromISO = toRFC3339(rangeStart);
	const toISO = toRFC3339(rangeEnd);

	const { data, isLoading, isError } = useQuery({
		queryKey: ["calendar", "events", fromISO, toISO],
		queryFn: () => listCalendarEvents(fromISO, toISO),
	});

	const events = useMemo(() => {
		const mapped = expandEvents(data?.events ?? []);
		return mapped.filter((ev) => {
			if (!ev.kind) return true;
			return filters[ev.kind];
		});
	}, [data, filters]);

	const handleEventClick = useCallback((ev: CalendarEvent) => {
		setSelectedEvent(ev);
	}, []);

	const handleExport = async () => {
		try {
			await downloadCalendarExport(fromISO, toISO);
			toast.success("Календарь экспортирован");
		} catch {
			toast.error("Не удалось экспортировать календарь");
		}
	};

	const headerText = useMemo(() => {
		if (view === "month") {
			return new Date(
				displayMonth.year,
				displayMonth.month,
				1,
			).toLocaleDateString("ru-RU", { month: "long", year: "numeric" });
		}
		if (view === "day") {
			return displayDay.toLocaleDateString("ru-RU", {
				weekday: "long",
				day: "numeric",
				month: "long",
				year: "numeric",
			});
		}
		if (view === "agenda") return "Ближайшие 30 дней";
		return `${weekDays[0]?.toLocaleDateString("ru-RU", { day: "numeric", month: "long" }) ?? ""} — ${weekDays[6]?.toLocaleDateString("ru-RU", { day: "numeric", month: "long", year: "numeric" }) ?? ""}`;
	}, [view, displayMonth, weekDays, displayDay]);

	const toggleFilter = (kind: CalendarEventKind) => {
		setFilters((prev) => ({ ...prev, [kind]: !prev[kind] }));
	};

	return (
		<div className="mx-auto max-w-6xl space-y-4 px-4 py-6">
			<header className="space-y-1">
				<h1 className="text-2xl font-bold tracking-tight">Календарь</h1>
				<p className="text-sm text-t2">
					Созвоны, спринты и дедлайны задач в одном представлении
				</p>
			</header>

			<div className="flex flex-wrap items-center gap-2">
				{(Object.keys(KIND_LABELS) as CalendarEventKind[]).map((kind) => (
					<button
						key={kind}
						type="button"
						onClick={() => toggleFilter(kind)}
						className={cn(
							"rounded-full border px-3 py-1 text-xs font-medium transition-colors",
							filters[kind]
								? getColorForKind(kind)
								: "border-b1 text-t3 opacity-60",
						)}
					>
						{KIND_LABELS[kind]}
					</button>
				))}
			</div>

			<div className="flex flex-wrap items-center gap-[14px] pb-2">
				{view !== "agenda" && (
					<>
						<button
							type="button"
							aria-label="Предыдущий период"
							onClick={() => {
								if (view === "month") setMonthOffset((m) => m - 1);
								else if (view === "week") setWeekOffset((w) => w - 1);
								else setDayOffset((d) => d - 1);
							}}
							className="flex h-[34px] w-[34px] items-center justify-center rounded-[var(--r2)] border border-b1 text-t2 hover:bg-[var(--hover)]"
						>
							<ChevronLeft className="h-4 w-4" aria-hidden="true" />
						</button>
						<span className="text-[17px] font-bold capitalize tracking-[-0.02em]">
							{headerText}
						</span>
						<button
							type="button"
							aria-label="Следующий период"
							onClick={() => {
								if (view === "month") setMonthOffset((m) => m + 1);
								else if (view === "week") setWeekOffset((w) => w + 1);
								else setDayOffset((d) => d + 1);
							}}
							className="flex h-[34px] w-[34px] items-center justify-center rounded-[var(--r2)] border border-b1 text-t2 hover:bg-[var(--hover)]"
						>
							<ChevronRight className="h-4 w-4" aria-hidden="true" />
						</button>
						<button
							type="button"
							onClick={() => {
								setMonthOffset(0);
								setWeekOffset(0);
								setDayOffset(0);
							}}
							className="rounded-full border border-b1 px-[14px] py-[5px] text-[12px] font-semibold text-t2 hover:border-purple hover:text-purple"
						>
							Сегодня
						</button>
					</>
				)}
				{view === "agenda" && (
					<span className="text-[17px] font-bold capitalize">{headerText}</span>
				)}

				<div className="ml-auto flex flex-wrap items-center gap-2">
					<div className="flex gap-[2px] rounded-[var(--r2)] border border-b1 bg-input p-[3px]">
						{(
							[
								["month", "Месяц"],
								["week", "Неделя"],
								["day", "День"],
								["agenda", "Список"],
							] as const
						).map(([key, label]) => (
							<button
								key={key}
								type="button"
								className={cn(
									"rounded-[var(--r1)] px-3 py-[5px] text-[12px] font-semibold transition-colors",
									view === key
										? "bg-card text-t1 shadow-sm"
										: "text-t3 hover:text-t2",
								)}
								onClick={() => setView(key)}
							>
								{label}
							</button>
						))}
					</div>
					<Button
						type="button"
						variant="outline"
						size="sm"
						onClick={() => void handleExport()}
					>
						<Download className="mr-1 h-3.5 w-3.5" aria-hidden="true" />
						iCal
					</Button>
					<Button
						type="button"
						size="sm"
						onClick={() => void navigate({ to: "/calls" })}
					>
						Запланировать созвон
					</Button>
				</div>
			</div>

			{isLoading && (
				<p className="text-sm text-t3" role="status">
					Загрузка событий…
				</p>
			)}
			{isError && (
				<p className="text-sm text-red-400" role="alert">
					Не удалось загрузить календарь
				</p>
			)}

			{view === "month" && (
				<MonthView
					year={displayMonth.year}
					month={displayMonth.month}
					today={today}
					events={events}
					onDayClick={() => {}}
					onEventClick={handleEventClick}
				/>
			)}
			{view === "week" && (
				<WeekView
					weekDays={weekDays}
					today={today}
					events={events}
					isCurrentWeek={weekOffset === 0}
					onSlotClick={() => {}}
					onEventClick={handleEventClick}
				/>
			)}
			{view === "day" && (
				<DayView
					day={displayDay}
					today={today}
					events={events}
					onSlotClick={() => {}}
					onEventClick={handleEventClick}
				/>
			)}
			{view === "agenda" && (
				<AgendaView events={events} onEventClick={handleEventClick} />
			)}

			{selectedEvent && (
				<div className="fixed inset-0 z-40 flex items-center justify-center">
					<button
						type="button"
						className="absolute inset-0 cursor-default border-0 bg-transparent p-0"
						aria-label="Закрыть"
						onClick={() => setSelectedEvent(null)}
					/>
					<div
						role="dialog"
						aria-modal="true"
						className="relative z-50 w-80 rounded-xl border bg-card p-4 shadow-lg"
					>
						<div className="space-y-2">
							<div className="flex items-center justify-between gap-2">
								<h3 className="font-semibold">{selectedEvent.title}</h3>
								<button
									type="button"
									onClick={() => setSelectedEvent(null)}
									className="rounded-sm p-0.5 text-t2 hover:text-t1"
								>
									&times;
								</button>
							</div>
							{selectedEvent.kind && (
								<p className="text-xs uppercase text-t3">
									{KIND_LABELS[selectedEvent.kind]}
									{selectedEvent.projectKey
										? ` · ${selectedEvent.projectKey}`
										: ""}
								</p>
							)}
							{selectedEvent.description && (
								<p className="text-sm text-t2">{selectedEvent.description}</p>
							)}
							<p className="text-sm text-t2">
								{selectedEvent.date.toLocaleString("ru-RU")}
							</p>
							{selectedEvent.url && (
								<Button
									type="button"
									size="sm"
									variant="outline"
									onClick={() => {
										setSelectedEvent(null);
										void navigate({ to: selectedEvent.url as string });
									}}
								>
									Открыть
								</Button>
							)}
						</div>
					</div>
				</div>
			)}
		</div>
	);
}
