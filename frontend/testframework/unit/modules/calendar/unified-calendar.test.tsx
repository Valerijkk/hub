import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as calendarApi from "@/api/calendar";
import { UnifiedCalendar } from "@/modules/calendar/unified-calendar";

vi.mock("@/api/calendar");
vi.mock("@tanstack/react-router", () => ({
	useNavigate: () => vi.fn(),
}));

describe("UnifiedCalendar", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(calendarApi.listCalendarEvents).mockResolvedValue({
			from: "2026-06-01T00:00:00.000Z",
			to: "2026-06-30T23:59:59.999Z",
			events: [
				{
					id: "sprint-1",
					type: "sprint",
					title: "Спринт: Sprint 1",
					starts_at: "2026-06-01T00:00:00.000Z",
					ends_at: "2026-06-03T00:00:00.000Z",
					all_day: true,
					project_key: "HUB",
				},
				{
					id: "deadline-1",
					type: "deadline",
					title: "HUB-1: Fix bug",
					starts_at: "2026-06-10T00:00:00.000Z",
					all_day: true,
				},
			],
		});
	});

	it("renders calendar header and filter chips", async () => {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<UnifiedCalendar />
			</QueryClientProvider>,
		);

		expect(
			screen.getByRole("heading", { name: /Календарь/i }),
		).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /Спринты/i }),
		).toBeInTheDocument();
		await waitFor(() => {
			expect(calendarApi.listCalendarEvents).toHaveBeenCalled();
		});
	});
});
