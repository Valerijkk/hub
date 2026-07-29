import { beforeEach, describe, expect, it, vi } from "vitest";
import * as dashboardApi from "@/api/dashboard";
import { useDashboardStore } from "@/stores/dashboard-store";

vi.mock("@/api/dashboard", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/dashboard")>();
	return {
		...actual,
		getLayout: vi.fn(),
	};
});

/**
 * A3W11-3-001 / TASK-131: loadLayout fail → store.error + reject;
 * не silent catch (иначе caller подставляет default как «успех»).
 */
describe("A3W11-3-001 · dashboard loadLayout error≠empty", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		useDashboardStore.setState({
			layouts: {},
			error: null,
			personalData: null,
			projectData: null,
			adminData: null,
			loading: false,
		});
	});

	it("sets error and rejects when getLayout fails", async () => {
		vi.mocked(dashboardApi.getLayout).mockRejectedValue(new Error("network"));

		await expect(
			useDashboardStore.getState().loadLayout("personal"),
		).rejects.toThrow("network");

		expect(useDashboardStore.getState().error).toBe(
			"Не удалось загрузить раскладку",
		);
		expect(useDashboardStore.getState().layouts.personal).toBeUndefined();
	});

	it("writes layout items on success without error", async () => {
		const items = [
			{ i: "a", x: 0, y: 0, w: 2, h: 2, widget_type: "tasks" as const },
		];
		vi.mocked(dashboardApi.getLayout).mockResolvedValue({
			dashboard_type: "personal",
			items,
		});

		await useDashboardStore.getState().loadLayout("personal");

		expect(useDashboardStore.getState().layouts.personal).toEqual(items);
		expect(useDashboardStore.getState().error).toBeNull();
	});
});
