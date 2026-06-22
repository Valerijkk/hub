import { useAuthStore } from "@/stores/auth-store";
import { api } from "./client";

export type CalendarEventKind = "call" | "sprint" | "deadline" | "leave";

export interface CalendarEventDTO {
	id: string;
	type: CalendarEventKind;
	title: string;
	description?: string;
	starts_at: string;
	ends_at?: string;
	all_day: boolean;
	project_id?: string;
	project_key?: string;
	url?: string;
}

export interface CalendarEventsResponse {
	events: CalendarEventDTO[];
	from: string;
	to: string;
}

export function listCalendarEvents(
	from: string,
	to: string,
): Promise<CalendarEventsResponse> {
	return api.get<CalendarEventsResponse>(
		`/calendar/events?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
	);
}

/** Скачивание .ics с JWT (не через JSON api client). */
export async function downloadCalendarExport(
	from: string,
	to: string,
): Promise<void> {
	const token = useAuthStore.getState().accessToken;
	const res = await fetch(
		`/api/calendar/export.ics?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
		{
			headers: token ? { Authorization: `Bearer ${token}` } : {},
		},
	);
	if (!res.ok) {
		throw new Error("export failed");
	}
	const blob = await res.blob();
	const url = URL.createObjectURL(blob);
	const a = document.createElement("a");
	a.href = url;
	a.download = "hub-calendar.ics";
	a.click();
	URL.revokeObjectURL(url);
}
