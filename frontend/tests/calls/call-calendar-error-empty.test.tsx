import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as callsApi from "@/api/calls";
import * as messengerApi from "@/api/messenger";
import * as projectsApi from "@/api/projects";
import { CallCalendar } from "@/modules/calls/call-calendar";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/messenger");
vi.mock("@/api/calls");
vi.mock("@/api/projects");

/** FE-MC-001 / TASK-186: getChannels fail ≠ пустая сетка календаря. */
describe("CallCalendar channels error≠empty (FE-MC-001 / TASK-186)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(callsApi.listScheduledCallsForChannels).mockResolvedValue([]);
		vi.mocked(projectsApi.listProjects).mockResolvedValue([]);
		vi.mocked(projectsApi.listMembers).mockResolvedValue([]);
	});

	it("shows alert+retry when getChannels fails, not month grid", async () => {
		vi.mocked(messengerApi.getChannels).mockRejectedValue(
			new Error("network"),
		);

		renderWithProviders(<CallCalendar />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить каналы/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		// Month grid table must not appear on channels fail (empty-on-fail RED if isError removed)
		expect(screen.queryByRole("table")).not.toBeInTheDocument();
	});

	it("renders calendar grid when channels load successfully (even if empty)", async () => {
		vi.mocked(messengerApi.getChannels).mockResolvedValue([]);

		renderWithProviders(<CallCalendar />);

		await waitFor(() => {
			expect(screen.getByRole("table")).toBeInTheDocument();
		});
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});
});
