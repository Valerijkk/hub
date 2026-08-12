import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as trackerApi from "@/api/tracker";
import { TimeTracker } from "@/components/tracker/time-tracker";
import { renderWithProviders } from "../../render";

vi.mock("@/api/tracker", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/tracker")>();
	return {
		...actual,
		listTimeLogs: vi.fn(),
		startTimer: vi.fn(),
		stopTimer: vi.fn(),
		logTime: vi.fn(),
	};
});

vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

vi.mock("@/stores/timer-store", () => ({
	useTimerStore: (
		selector: (s: {
			isRunning: boolean;
			activeIssueKey: string | null;
			activeProjectId: string | null;
			elapsedSeconds: number;
			startTimer: () => void;
			stopTimer: () => void;
			tick: () => void;
		}) => unknown,
	) =>
		selector({
			isRunning: false,
			activeIssueKey: null,
			activeProjectId: null,
			elapsedSeconds: 0,
			startTimer: vi.fn(),
			stopTimer: vi.fn(),
			tick: vi.fn(),
		}),
}));

describe("TimeTracker error≠empty (A3W9-4-006 / TASK-175)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows История + alert+retry when listTimeLogs fails", async () => {
		vi.mocked(trackerApi.listTimeLogs).mockRejectedValue(new Error("network"));

		renderWithProviders(
			<TimeTracker projectId="proj-1" issueKey="HUB-1" />,
		);

		await waitFor(() => {
			expect(screen.getByText("История")).toBeInTheDocument();
		});
		expect(screen.getByRole("alert")).toHaveTextContent(
			/не удалось загрузить историю времени/i,
		);
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});
});
