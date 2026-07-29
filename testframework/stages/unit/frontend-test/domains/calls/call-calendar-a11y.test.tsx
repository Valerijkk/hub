import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as callsApi from "@/api/calls";
import * as messengerApi from "@/api/messenger";
import * as projectsApi from "@/api/projects";
import { CallCalendar } from "@/modules/calls/call-calendar";
import { getWeekDays, getWeekStart } from "@/modules/calls/calendar-types";
import { WeekView } from "@/modules/calls/calendar-week-view";
import { renderWithProviders } from "../../render";

vi.mock("@/api/messenger");
vi.mock("@/api/calls");
vi.mock("@/api/projects");

/** FE-MC-003 / FE-MC-006 / TASK-186: a11y names for popover close + week slots. */
describe("Calls calendar a11y (FE-MC-003/006 / TASK-186)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(projectsApi.listProjects).mockResolvedValue([]);
		vi.mocked(projectsApi.listMembers).mockResolvedValue([]);
	});

	it("event popover close control has accessible name (FE-MC-003)", async () => {
		const scheduledAt = new Date();
		scheduledAt.setHours(10, 0, 0, 0);

		vi.mocked(messengerApi.getChannels).mockResolvedValue([
			{
				id: "ch-1",
				channel_type: "group",
				name: "general",
				is_readonly: false,
				created_by: "u1",
				is_starred: false,
				created_at: "2026-01-01T00:00:00Z",
				updated_at: "2026-01-01T00:00:00Z",
			},
		]);
		vi.mocked(callsApi.listScheduledCallsForChannels).mockResolvedValue([
			{
				id: "call-1",
				title: "Standup",
				description: "Daily",
				scheduled_at: scheduledAt.toISOString(),
				duration_minutes: 60,
				channel_id: "ch-1",
				project_id: "p1",
				created_by: "u1",
				is_active: true,
				created_at: "2026-01-01T00:00:00Z",
				updated_at: "2026-01-01T00:00:00Z",
			},
		]);

		renderWithProviders(<CallCalendar />);

		const eventBtn = await screen.findByRole("button", {
			name: /встреча: standup/i,
		});
		fireEvent.click(eventBtn);

		await waitFor(() => {
			expect(screen.getByRole("dialog")).toBeInTheDocument();
		});

		const closeButtons = screen.getAllByRole("button", { name: /^закрыть$/i });
		// Backdrop + × popover close (FE-MC-003)
		expect(closeButtons.length).toBeGreaterThanOrEqual(2);
	});

	it("week slot aria-label includes weekday/date, not only hour (FE-MC-006)", () => {
		const today = new Date();
		const weekDays = getWeekDays(getWeekStart(today));

		renderWithProviders(
			<WeekView
				weekDays={weekDays}
				today={today}
				events={[]}
				isCurrentWeek
				onSlotClick={vi.fn()}
				onEventClick={vi.fn()}
			/>,
		);

		const labels = screen
			.getAllByRole("button", { name: /создать встречу на/i })
			.map((el) => el.getAttribute("aria-label") ?? "");

		expect(labels.length).toBeGreaterThan(7);
		// Each label must mention a day/month fragment, not only HH:00
		for (const label of labels.slice(0, 7)) {
			expect(label).toMatch(/\d/);
			expect(label).not.toMatch(/^Создать встречу на \d{2}:00$/);
		}
		// Adjacent days in the same hour-row must differ
		const firstRow = labels.filter((l) => l.includes("08:00"));
		expect(new Set(firstRow).size).toBe(7);
	});
});
