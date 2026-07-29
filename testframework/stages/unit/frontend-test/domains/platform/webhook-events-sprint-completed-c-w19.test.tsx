import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "@/api/client";
import {
	AVAILABLE_EVENTS,
	WebhooksPage,
} from "@/modules/settings/webhooks";
import { renderWithProviders } from "../../render";

/**
 * C-W19-FE-001 / TASK-179: FE не предлагает `sprint.completed` после BE allowlist trim.
 * Mutation: вернуть `"sprint.completed"` в AVAILABLE_EVENTS → оба it ниже краснеют.
 * BE при Create/Update с этим event отвечает 400 «неизвестное событие» (TASK-152).
 */
describe("TASK-179 C-W19-FE-001 webhook events allowlist", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
		vi.spyOn(api, "get").mockResolvedValue([]);
	});

	it("AVAILABLE_EVENTS does not include sprint.completed", () => {
		expect(AVAILABLE_EVENTS).not.toContain("sprint.completed");
		expect([...AVAILABLE_EVENTS]).toEqual([
			"issue.created",
			"issue.updated",
			"issue.status_changed",
			"test_run.completed",
		]);
	});

	it("create form checkboxes omit sprint.completed", async () => {
		renderWithProviders(
			<WebhooksPage
				projectId="11111111-1111-4111-8111-111111111111"
				canMutate={true}
			/>,
		);

		fireEvent.click(await screen.findByTestId("webhook-create-btn"));

		await waitFor(() => {
			expect(screen.getByText("Новый вебхук")).toBeInTheDocument();
		});

		expect(screen.getByText("issue.created")).toBeInTheDocument();
		expect(screen.getByText("test_run.completed")).toBeInTheDocument();
		expect(screen.queryByText("sprint.completed")).toBeNull();
	});
});
