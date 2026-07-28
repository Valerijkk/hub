import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import * as dashboardApi from "@/api/dashboard";
import { useDashboardStore } from "@/stores/dashboard-store";

vi.mock("@/api/dashboard", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/dashboard")>();
	return {
		...actual,
		saveLayout: vi.fn(),
	};
});

describe("P2-FE-02 · dashboard layout save rollback (W11-A3-009)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.useFakeTimers();
		useDashboardStore.setState({
			layouts: {
				personal: [{ i: "a", x: 0, y: 0, w: 2, h: 2, widget_type: "tasks" }],
			},
			error: null,
		});
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it("sets error and restores previous layout when save fails", async () => {
		vi.mocked(dashboardApi.saveLayout).mockRejectedValue(new Error("fail"));

		const next = [{ i: "a", x: 1, y: 0, w: 2, h: 2, widget_type: "tasks" }];
		useDashboardStore.getState().saveLayoutDebounced("personal", next);

		expect(useDashboardStore.getState().layouts.personal).toEqual(next);

		await vi.advanceTimersByTimeAsync(500);
		await Promise.resolve();
		await Promise.resolve();

		expect(useDashboardStore.getState().error).toBe(
			"Не удалось сохранить раскладку дашборда",
		);
		expect(useDashboardStore.getState().layouts.personal).toEqual([
			{ i: "a", x: 0, y: 0, w: 2, h: 2, widget_type: "tasks" },
		]);
	});
});
