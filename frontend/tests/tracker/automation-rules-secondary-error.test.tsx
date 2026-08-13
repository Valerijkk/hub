import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as automationApi from "@/api/automation";
import * as messengerApi from "@/api/messenger";
import * as trackerApi from "@/api/tracker";
import { AutomationRulesPanel } from "@/components/tracker/automation-rules-panel";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/automation");
vi.mock("@/api/messenger");
vi.mock("@/api/tracker");
vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

describe("AutomationRulesPanel secondary error≠empty (FE-TA-004 / TASK-187)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(automationApi.listAutomationRules).mockResolvedValue([]);
	});

	it("statuses fail → alert, not empty status Select as «нет данных»", async () => {
		vi.mocked(trackerApi.listWorkflowStatuses).mockRejectedValue(
			new Error("statuses down"),
		);
		vi.mocked(messengerApi.getChannels).mockResolvedValue([]);

		renderWithProviders(<AutomationRulesPanel projectId="proj-1" />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить статусы workflow/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});

	it("channels fail → alert, not empty channel Select", async () => {
		vi.mocked(trackerApi.listWorkflowStatuses).mockResolvedValue([]);
		vi.mocked(messengerApi.getChannels).mockRejectedValue(
			new Error("channels down"),
		);

		renderWithProviders(<AutomationRulesPanel projectId="proj-1" />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить каналы/i,
			);
		});
	});
});
