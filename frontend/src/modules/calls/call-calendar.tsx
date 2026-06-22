import { useQuery } from "@tanstack/react-query";
import { ChevronLeft, ChevronRight, Plus } from "lucide-react";
import { useCallback, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { listScheduledCallsForChannels } from "@/api/calls";
import { getChannels } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import { AgendaView } from "./calendar-agenda-view";
import { DayView } from "./calendar-day-view";
import { MonthView } from "./calendar-month-view";
import {
	type CalendarEvent,
	type CalendarView,
	DEFAULT_DURATION,
	expandRRule,
	getColorForEvent,
	getMonthGridDays,
	getWeekDays,
	getWeekStart,
	type ScheduledCallRecord,
} from "./calendar-types";
import { WeekView } from "./calendar-week-view";
import { CreateCallDialog } from "./create-call-dialog";

// ---------------------------------------------------------------------------
// Компонент календаря звонков
// ---------------------------------------------------------------------------

/**
 * CallCalendar — календарь запланированных звонков.
 * Режимы месяц (по умолчанию) и неделя + переключатель видов.
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

	// --- Состояние вида «Месяц» ---
	const displayMonth = useMemo(() => {
		const d = new Date(today.getFullYear(), today.getMonth() + monthOffset, 1);
		return { year: d.getFullYear(), month: d.getMonth() };
	}, [today, monthOffset]);

	const monthRangeStart = useMemo(() => {
		const days = getMonthGridDays(displayMonth.year, displayMonth.month);
		// biome-ignore lint/style/noNonNullAssertion: getMonthGridDays always returns >=35 days
		return days[0]!;
	}, [displayMonth]);

	const monthRangeEnd = useMemo(() => {
		const days = getMonthGridDays(displayMonth.year, displayMonth.month);
		// biome-ignore lint/style/noNonNullAssertion: getMonthGridDays always returns >=35 days
		const last = days[days.length - 1]!;
		const end = new Date(last);
		end.setHours(23, 59, 59, 999);
		return end;
	}, [displayMonth]);

	// --- Состояние вида «Неделя» ---
	const weekStart = useMemo(() => {
		const base = getWeekStart(today);
		base.setDate(base.getDate() + weekOffset * 7);
		return base;
	}, [today, weekOffset]);

	const weekDays = useMemo(() => getWeekDays(weekStart), [weekStart]);

	const weekRangeEnd = useMemo(() => {
		// biome-ignore lint/style/noNonNullAssertion: getWeekDays always returns 7 elements
		const end = new Date(weekDays[6]!);
		end.setHours(23, 59, 59, 999);
		return end;
	}, [weekDays]);

	// --- Состояние вида «День» ---
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

	// --- Состояние «Повестки» (30 дней от сегодня) ---
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

	// --- Диапазон дат для загрузки звонков по активному виду ---
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

	// Разворачиваем RRULE в события только для видимого диапазона
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

	// --- Обработчики навигации и действий ---

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

	// --- Текст заголовка периода ---
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
		// biome-ignore lint/style/noNonNullAssertion: getWeekDays always returns 7 elements
		return `${weekDays[0]!.toLocaleDateString("ru-RU", {
			day: "numeric",
			month: "long",
			// biome-ignore lint/style/noNonNullAssertion: getWeekDays always returns 7 elements
		})} \u2014 ${weekDays[6]!.toLocaleDateString("ru-RU", {
			day: "numeric",
			month: "long",
			year: "numeric",
		})}`;
	}, [view, displayMonth, weekDays, displayDay]);

	const isCurrentWeek = weekOffset === 0;

	return (
		<div className="space-y-4">
			{/* Панель навигации */}
			<div className="flex items-center gap-[14px] pt-[18px] pb-[14px]">
				{view !== "agenda" && (
					<>
						<button
							type="button"
							onClick={handleNavigatePrev}
							className="w-[34px] h-[34px] flex items-center justify-center rounded-[var(--r2)] border border-b1 text-t2 transition-colors duration-180 hover:bg-[var(--hover)] hover:border-b2 hover:text-t1 text-[16px]"
						>
							<ChevronLeft className="h-4 w-4" />
						</button>
						<span className="text-[17px] font-bold tracking-[-0.02em] capitalize">
							{headerText}
						</span>
						<button
							type="button"
							onClick={handleNavigateNext}
							className="w-[34px] h-[34px] flex items-center justify-center rounded-[var(--r2)] border border-b1 text-t2 transition-colors duration-180 hover:bg-[var(--hover)] hover:border-b2 hover:text-t1 text-[16px]"
						>
							<ChevronRight className="h-4 w-4" />
						</button>
						<button
							type="button"
							onClick={handleNavigateToday}
							className="px-[14px] py-[5px] rounded-full border border-b1 text-[12px] font-semibold text-t2 transition-colors duration-180 hover:border-purple hover:text-purple"
						>
							{t("calendarToday", { defaultValue: "Сегодня" })}
						</button>
					</>
				)}
				{view === "agenda" && (
					<span className="text-[17px] font-bold tracking-[-0.02em] capitalize">
						{headerText}
					</span>
				)}

				{/* Переключатель видов */}
				<div className="flex gap-[2px] bg-input rounded-[var(--r2)] p-[3px] ml-auto border border-b1">
					{(
						[
							{
								key: "month" as const,
								label: t("calendarMonth", { defaultValue: "Месяц" }),
							},
							{
								key: "week" as const,
								label: t("calendarWeek", { defaultValue: "Неделя" }),
							},
							{
								key: "day" as const,
								label: t("calendarDay", { defaultValue: "День" }),
							},
							{
								key: "agenda" as const,
								label: t("calendarAgenda", { defaultValue: "Список" }),
							},
						] as const
					).map((tab) => (
						<button
							key={tab.key}
							type="button"
							className={`px-3 py-[5px] text-[12px] font-semibold rounded-[var(--r1)] transition-colors duration-180 ${
								view === tab.key
									? "bg-card text-t1 shadow-[0_1px_3px_rgba(0,0,0,0.5)]"
									: "text-t3 hover:text-t2"
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
					className="calls-btn-p inline-flex items-center gap-[7px] px-[14px] py-[5px] rounded-[var(--r2)] text-[12px] font-semibold bg-gradient-to-br from-[var(--purpleS)] to-[#6d28d9] text-white shadow-[0_0_12px_var(--gP)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-180 hover:shadow-[0_0_20px_rgba(167,139,250,0.3)] hover:-translate-y-px relative overflow-hidden"
				>
					<Plus className="mr-1 h-3.5 w-3.5" />
					{t("calendarNewCall", { defaultValue: "Запланировать" })}
				</Button>
			</div>

			{/* Область календарных видов */}
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
				<AgendaView events={events} onEventClick={handleEventClick} />
			)}

			{/* Popover деталей события */}
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
							<div className="flex items-center justify-between">
								<h3 className="font-semibold">{selectedEvent.title}</h3>
								<button
									type="button"
									onClick={() => setSelectedEvent(null)}
									className="rounded-sm p-0.5 text-t2 hover:text-t1"
								>
									&times;
								</button>
							</div>
							{selectedEvent.description && (
								<p className="text-sm text-t2">{selectedEvent.description}</p>
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
									<span className="text-xs text-t2">
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

			{/* Диалог создания звонка */}
			<CreateCallDialog
				open={dialogOpen}
				onOpenChange={setDialogOpen}
				defaultDate={dialogDefaults.date}
				defaultTime={dialogDefaults.time}
			/>
		</div>
	);
}
