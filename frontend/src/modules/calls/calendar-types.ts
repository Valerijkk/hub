// ---------------------------------------------------------------------------
// Типы календаря
// ---------------------------------------------------------------------------

/** Запись запланированного звонка из API. */
export interface ScheduledCallRecord {
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

/** Событие с координатами в сетке календаря. */
export interface CalendarEvent {
	id: string;
	title: string;
	description?: string;
	date: Date;
	durationMinutes: number;
	channelId: string;
	rrule?: string;
	/** Pastel-класс цвета блока события. */
	colorClass: string;
	kind?: CalendarEventKind;
	url?: string;
	projectKey?: string;
}

export type CalendarView = "month" | "week" | "day" | "agenda";

export type CalendarEventKind = "call" | "sprint" | "deadline" | "leave";

const CALL_COLORS = [
	"bg-violet-400/20 border-violet-400/40 text-violet-700 dark:text-violet-300",
	"bg-blue-400/20 border-blue-400/40 text-blue-700 dark:text-blue-300",
	"bg-emerald-400/20 border-emerald-400/40 text-emerald-700 dark:text-emerald-300",
	"bg-amber-400/20 border-amber-400/40 text-amber-700 dark:text-amber-300",
	"bg-rose-400/20 border-rose-400/40 text-rose-700 dark:text-rose-300",
	"bg-cyan-400/20 border-cyan-400/40 text-cyan-700 dark:text-cyan-300",
] as const;

export function getColorForEvent(id: string): string {
	let hash = 0;
	for (let i = 0; i < id.length; i++) {
		hash = (hash * 31 + id.charCodeAt(i)) | 0;
	}
	return CALL_COLORS[Math.abs(hash) % CALL_COLORS.length] as string;
}

const KIND_COLORS: Record<CalendarEventKind, string> = {
	call: "bg-violet-400/20 border-violet-400/40 text-violet-700 dark:text-violet-300",
	sprint:
		"bg-indigo-400/20 border-indigo-400/40 text-indigo-700 dark:text-indigo-300",
	deadline:
		"bg-orange-400/20 border-orange-400/40 text-orange-700 dark:text-orange-300",
	leave: "bg-pink-400/20 border-pink-400/40 text-pink-700 dark:text-pink-300",
};

export function getColorForKind(kind: CalendarEventKind): string {
	return KIND_COLORS[kind];
}

// ---------------------------------------------------------------------------
// Вспомогательные функции календаря
// ---------------------------------------------------------------------------

export const HOUR_START = 8;
export const HOUR_END = 20;
export const HOURS = Array.from(
	{ length: HOUR_END - HOUR_START },
	(_, i) => HOUR_START + i,
);
export const HOURS_FULL = Array.from({ length: 24 }, (_, i) => i);
export const DEFAULT_DURATION = 60; // minutes

export const WEEKDAY_NAMES_SHORT = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];

/** Понедельник недели, в которую попадает `date`. */
export function getWeekStart(date: Date): Date {
	const d = new Date(date);
	const day = d.getDay();
	const diff = day === 0 ? -6 : 1 - day;
	d.setDate(d.getDate() + diff);
	d.setHours(0, 0, 0, 0);
	return d;
}

/** Семь дней начиная с `weekStart`. */
export function getWeekDays(weekStart: Date): Date[] {
	return Array.from({ length: 7 }, (_, i) => {
		const d = new Date(weekStart);
		d.setDate(d.getDate() + i);
		return d;
	});
}

export function isSameDay(a: Date, b: Date): boolean {
	return (
		a.getFullYear() === b.getFullYear() &&
		a.getMonth() === b.getMonth() &&
		a.getDate() === b.getDate()
	);
}

export function isSameMonth(a: Date, b: Date): boolean {
	return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth();
}

export function formatDayHeader(
	date: Date,
	locale: string,
): { weekday: string; day: string } {
	const weekday = date.toLocaleDateString(locale, { weekday: "short" });
	const day = date.getDate().toString();
	return { weekday, day };
}

/** Все ячейки месячной сетки, включая padding из соседних месяцев. */
export function getMonthGridDays(year: number, month: number): Date[] {
	const firstDay = new Date(year, month, 1);
	const lastDay = new Date(year, month + 1, 0);

	// Понедельник на или до 1-го числа — начало сетки
	let startDay = firstDay.getDay(); // 0=Sun
	// Приводим к формату Mon=0 для расчёта offset
	startDay = startDay === 0 ? 6 : startDay - 1;

	const days: Date[] = [];

	// Добавляем хвост предыдущего месяца для выравнивания сетки
	for (let i = startDay - 1; i >= 0; i--) {
		const d = new Date(year, month, 1);
		d.setDate(d.getDate() - i - 1);
		days.push(d);
	}

	// Все дни текущего месяца
	for (let d = 1; d <= lastDay.getDate(); d++) {
		days.push(new Date(year, month, d));
	}

	// Добиваем последнюю неделю до 35/42 ячеек (6 строк × 7 дней)
	while (days.length % 7 !== 0) {
		const last = days[days.length - 1];
		if (!last) break;
		const next = new Date(last);
		next.setDate(next.getDate() + 1);
		days.push(next);
	}

	return days;
}

/**
 * Simple RRULE expansion within a date range.
 * Поддерживаются только FREQ=DAILY и FREQ=WEEKLY (опционально BYDAY).
 */
export function expandRRule(
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
	const byDays = byDayMatch?.[1]
		? byDayMatch[1].split(",").map((d) => {
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
