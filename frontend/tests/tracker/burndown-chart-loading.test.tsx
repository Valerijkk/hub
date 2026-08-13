import { act, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { BurndownChart } from "@/modules/tracker/burndown-chart";
import { useSprintStore } from "@/stores/sprint-store";
import { renderWithProviders } from "../support/render";

const getBurndown = vi.fn();

vi.mock("@/api/sprint", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/sprint")>();
	return {
		...actual,
		getBurndown: (...args: unknown[]) => getBurndown(...args),
	};
});

vi.mock("recharts", () => ({
	ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
		<div data-testid="chart">{children}</div>
	),
	AreaChart: ({ children }: { children: React.ReactNode }) => (
		<div>{children}</div>
	),
	CartesianGrid: () => null,
	XAxis: () => null,
	YAxis: () => null,
	Tooltip: () => null,
	Legend: () => null,
	Area: () => null,
}));

describe("FE · BurndownChart loading≠empty (A3W10-2-005)", () => {
	beforeEach(() => {
		getBurndown.mockReset();
		useSprintStore.setState({
			burndown: [],
			loading: false,
			error: null,
		});
	});

	it("shows loading UI while fetch is in flight, not empty copy", async () => {
		let resolveBurndown: (value: unknown[]) => void = () => {};
		getBurndown.mockImplementation(
			() =>
				new Promise((resolve) => {
					resolveBurndown = resolve;
				}),
		);

		renderWithProviders(
			<BurndownChart projectId="proj-1" sprintId="sprint-1" />,
		);

		await waitFor(() => {
			expect(screen.getByText("Загрузка…")).toBeInTheDocument();
		});
		expect(
			screen.queryByText("Нет данных для отображения"),
		).not.toBeInTheDocument();

		await act(async () => {
			resolveBurndown([]);
		});

		await waitFor(() => {
			expect(
				screen.getByText("Нет данных для отображения"),
			).toBeInTheDocument();
		});
		expect(screen.queryByText("Загрузка…")).not.toBeInTheDocument();

		// Mutation: empty при loading===true → assert на «Загрузка…» FAIL.
	});

	it("sets loading true then false in store around fetchBurndown", async () => {
		let resolveBurndown: (value: unknown[]) => void = () => {};
		getBurndown.mockImplementation(
			() =>
				new Promise((resolve) => {
					resolveBurndown = resolve;
				}),
		);

		const pending = useSprintStore
			.getState()
			.fetchBurndown("proj-1", "sprint-1");
		expect(useSprintStore.getState().loading).toBe(true);

		await act(async () => {
			resolveBurndown([{ date: "2026-07-01", ideal: 10, remaining: 8 }]);
			await pending;
		});

		expect(useSprintStore.getState().loading).toBe(false);
		expect(useSprintStore.getState().burndown).toHaveLength(1);
		expect(useSprintStore.getState().error).toBeNull();
	});
});
