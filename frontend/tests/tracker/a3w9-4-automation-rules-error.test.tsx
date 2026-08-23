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

describe("AutomationRulesPanel error≠empty (A3W9-4-002 / TASK-175)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(trackerApi.listWorkflowStatuses).mockResolvedValue([]);
		vi.mocked(messengerApi.getChannels).mockResolvedValue([]);
	});

	it("shows alert+retry on list fail, not empty «Нет правил»", async () => {
		vi.mocked(automationApi.listAutomationRules).mockRejectedValue(
			new Error("network"),
		);

		renderWithProviders(<AutomationRulesPanel projectId="proj-1" />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить правила автоматизации/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(
			screen.queryByText(/нет правил автоматизации/i),
		).not.toBeInTheDocument();
	});
});
