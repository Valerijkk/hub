import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, expectTypeOf, it, vi } from "vitest";
import { api } from "@/api/client";
import {
	updateWebhook,
	type UpdateWebhookRequest,
	type WebhookEndpoint,
} from "@/api/webhooks";
import {
	AVAILABLE_EVENTS,
	WebhooksPage,
} from "@/modules/settings/webhooks";
import { renderWithProviders } from "../../render";

/**
 * FE-SMA-001 / TASK-192: toggle webhook шлёт version (BE VersionEQ).
 * Mutation: убрать version из UpdateWebhookRequest / handleToggleActive → RED.
 * TASK-179: AVAILABLE_EVENTS без sprint.completed — не регрессировать.
 */
const endpoint: WebhookEndpoint = {
	id: "wh-1",
	url: "https://example.com/hook",
	events: ["issue.created"],
	is_active: true,
	version: 3,
	created_at: "2026-07-01T00:00:00Z",
};

describe("WebhooksPage version OL (FE-SMA-001 / TASK-192)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("UpdateWebhookRequest requires version (mutation: omit → TS fail)", () => {
		expectTypeOf<UpdateWebhookRequest>().toHaveProperty("version");
		expectTypeOf<UpdateWebhookRequest["version"]>().toEqualTypeOf<number>();
		expectTypeOf<WebhookEndpoint>().toHaveProperty("version");
	});

	it("AVAILABLE_EVENTS omits sprint.completed (TASK-179 / C-W19-FE-001)", () => {
		expect(AVAILABLE_EVENTS).not.toContain("sprint.completed");
	});

	it("updateWebhook PUT body includes version from caller", async () => {
		const putSpy = vi.spyOn(api, "put").mockResolvedValue({
			...endpoint,
			is_active: false,
			version: 4,
		});

		await updateWebhook("wh-1", { is_active: false, version: 3 });

		expect(putSpy).toHaveBeenCalledWith("/webhooks/wh-1", {
			is_active: false,
			version: 3,
		});
	});

	it("toggle active sends is_active + version", async () => {
		vi.spyOn(api, "get").mockImplementation(async (path: string) => {
			if (typeof path === "string" && path.endsWith("/webhooks")) {
				return [endpoint];
			}
			return {};
		});
		const putSpy = vi.spyOn(api, "put").mockResolvedValue({
			...endpoint,
			is_active: false,
			version: 4,
		});

		renderWithProviders(
			<WebhooksPage
				projectId="11111111-1111-4111-8111-111111111111"
				canMutate={true}
			/>,
		);

		fireEvent.click(await screen.findByRole("button", { name: "Отключить" }));

		await waitFor(() => {
			expect(putSpy).toHaveBeenCalledWith("/webhooks/wh-1", {
				is_active: false,
				version: 3,
			});
		});
	});
});
