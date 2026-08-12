import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as callsApi from "@/api/calls";
import * as messengerApi from "@/api/messenger";
import * as projectsApi from "@/api/projects";
import { CreateCallDialog } from "@/modules/calls/create-call-dialog";
import { renderWithProviders } from "../../render";

vi.mock("@/api/messenger");
vi.mock("@/api/projects");
vi.mock("@/api/calls");
vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

/** FE-MC-002 / TASK-186: listProjects/getChannels fail ≠ «нет проектов/каналов». */
describe("CreateCallDialog error≠empty (FE-MC-002 / TASK-186)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(callsApi.checkCallConflicts).mockResolvedValue({
			has_conflict: false,
			conflicts: [],
		});
		vi.mocked(projectsApi.listMembers).mockResolvedValue([]);
	});

	it("shows projects alert+retry on listProjects fail, not «Нет доступных проектов»", async () => {
		vi.mocked(projectsApi.listProjects).mockRejectedValue(new Error("network"));
		vi.mocked(messengerApi.getChannels).mockResolvedValue([]);

		renderWithProviders(
			<CreateCallDialog open onOpenChange={vi.fn()} defaultDate="2026-07-29" />,
		);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить проекты/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(
			screen.queryByText(/нет доступных проектов/i),
		).not.toBeInTheDocument();
	});

	it("shows channels alert+retry on getChannels fail, not empty «Выберите канал»", async () => {
		vi.mocked(projectsApi.listProjects).mockResolvedValue([
			{
				id: "p1",
				name: "Alpha",
				key: "ALP",
				status: "active",
				member_count: 1,
				version: 1,
				created_at: "2026-01-01T00:00:00Z",
				updated_at: "2026-01-01T00:00:00Z",
			},
		]);
		vi.mocked(messengerApi.getChannels).mockRejectedValue(new Error("network"));

		renderWithProviders(
			<CreateCallDialog open onOpenChange={vi.fn()} defaultDate="2026-07-29" />,
		);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить каналы/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(screen.queryByText(/выберите канал/i)).not.toBeInTheDocument();
		expect(
			screen.queryByText(/нет доступных каналов/i),
		).not.toBeInTheDocument();
	});

	it("shows «Нет доступных проектов» only on success+empty", async () => {
		vi.mocked(projectsApi.listProjects).mockResolvedValue([]);
		vi.mocked(messengerApi.getChannels).mockResolvedValue([]);

		renderWithProviders(
			<CreateCallDialog open onOpenChange={vi.fn()} defaultDate="2026-07-29" />,
		);

		const projectTrigger = await screen.findByRole("combobox", {
			name: /выбор проекта/i,
		});
		fireEvent.click(projectTrigger);

		await waitFor(() => {
			expect(
				screen.getAllByText(/нет доступных проектов/i).length,
			).toBeGreaterThan(0);
		});
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});

	it("channel SelectTrigger has accessible name (FE-MC-004)", async () => {
		vi.mocked(projectsApi.listProjects).mockResolvedValue([]);
		vi.mocked(messengerApi.getChannels).mockResolvedValue([]);

		renderWithProviders(
			<CreateCallDialog open onOpenChange={vi.fn()} defaultDate="2026-07-29" />,
		);

		await waitFor(() => {
			expect(
				screen.getByRole("combobox", { name: /выбор канала/i }),
			).toBeInTheDocument();
		});
		// project combobox still present (success path)
		expect(
			within(document.body).getByRole("combobox", { name: /выбор проекта/i }),
		).toBeInTheDocument();
	});
});
