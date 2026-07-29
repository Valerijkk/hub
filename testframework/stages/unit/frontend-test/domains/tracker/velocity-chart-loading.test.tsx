import { act, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { VelocityChart } from "@/modules/tracker/velocity-chart";
import { useSprintStore } from "@/stores/sprint-store";
import { renderWithProviders } from "@testframework/render";

const getVelocity = vi.fn();

vi.mock("@/api/sprint", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/sprint")>();
	return {
		...actual,
		getVelocity: (...args: unknown[]) => getVelocity(...args),
	};
});

vi.mock("recharts", () => ({
	ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
		<div data-testid="chart">{children}</div>
	),
	BarChart: ({ children }: { children: React.ReactNode }) => (
		<div>{children}</div>
	),
	CartesianGrid: () => null,
	XAxis: () => null,
	YAxis: () => null,
	Tooltip: () => null,
	Legend: () => null,
	Bar: () => null,
}));

describe("FE · VelocityChart loading≠empty (A3W10-3-002)", () => {
	beforeEach(() => {
		getVelocity.mockReset();
		useSprintStore.setState({
			velocity: [],
			velocityLoading: false,
			error: null,
		});
	});

	it("shows loading UI while fetch is in flight, not empty copy", async () => {
		let resolveVelocity: (value: unknown[]) => void = () => {};
		getVelocity.mockImplementation(
			() =>
				new Promise((resolve) => {
					resolveVelocity = resolve;
				}),
		);

		renderWithProviders(<VelocityChart projectId="proj-1" />);

		await waitFor(() => {
			expect(screen.getByText("Загрузка…")).toBeInTheDocument();
		});
		expect(
			screen.queryByText("Нет завершённых спринтов для отображения"),
		).not.toBeInTheDocument();

		await act(async () => {
			resolveVelocity([]);
		});

		await waitFor(() => {
			expect(
				screen.getByText("Нет завершённых спринтов для отображения"),
			).toBeInTheDocument();
		});
		expect(screen.queryByText("Загрузка…")).not.toBeInTheDocument();

		// Mutation: empty при velocityLoading===true → assert на «Загрузка…» FAIL.
	});

	it("renders chart after successful non-empty velocity", async () => {
		getVelocity.mockResolvedValue([
			{ sprint_name: "S1", committed: 10, completed: 8 },
		]);

		renderWithProviders(<VelocityChart projectId="proj-1" />);

		await waitFor(() => {
			expect(screen.getByTestId("chart")).toBeInTheDocument();
		});
		expect(
			screen.queryByText("Нет завершённых спринтов для отображения"),
		).not.toBeInTheDocument();
		expect(screen.queryByText("Загрузка…")).not.toBeInTheDocument();
	});
});
