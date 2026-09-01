import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as calendarApi from "@/api/calendar";
import type { CalendarEvent } from "@/modules/calls/calendar-types";
import { UnifiedCalendar } from "@/modules/calendar/unified-calendar";

const { navigate } = vi.hoisted(() => ({ navigate: vi.fn() }));

vi.mock("@/api/calendar");
vi.mock("@tanstack/react-router", () => ({
	useNavigate: () => navigate,
}));
vi.mock("@/modules/calls/calendar-month-view", () => ({
	MonthView: ({
		events,
		onEventClick,
	}: {
		events: CalendarEvent[];
		onEventClick: (event: CalendarEvent) => void;
	}) => (
		<button
			type="button"
			disabled={events.length === 0}
			onClick={() => {
				const event = events[0];
				if (event) onEventClick(event);
			}}
		>
			Open canonical event
		</button>
	),
}));

describe("UnifiedCalendar canonical event navigation", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(calendarApi.listCalendarEvents).mockResolvedValue({
			from: "2026-09-01T00:00:00.000Z",
			to: "2026-09-30T23:59:59.999Z",
			events: [
				{
					id: "event-occurrence-opaque",
					type: "event",
					title: "Planning",
					starts_at: "2026-09-15T10:00:00.000Z",
					ends_at: "2026-09-15T11:00:00.000Z",
					all_day: false,
					url: "/calendar/events/9b75aa3c-3dc9-4ce1-9995-524271eeb3c4",
				},
			],
		});
	});

	it("shows the event filter and navigates the aggregate URL without a detail request", async () => {
		const queryClient = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={queryClient}>
				<UnifiedCalendar />
			</QueryClientProvider>,
		);

		expect(screen.getByRole("button", { name: "События" })).toBeInTheDocument();
		await waitFor(() => {
			expect(calendarApi.listCalendarEvents).toHaveBeenCalledTimes(1);
		});
		const openEventButton = screen.getByRole("button", {
			name: "Open canonical event",
		});
		await waitFor(() => {
			expect(openEventButton).toBeEnabled();
		});
		fireEvent.click(openEventButton);

		expect(navigate).toHaveBeenCalledWith({
			to: "/calendar/events/9b75aa3c-3dc9-4ce1-9995-524271eeb3c4",
		});
		expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
		expect(calendarApi.listCalendarEvents).toHaveBeenCalledTimes(1);
	});
});
