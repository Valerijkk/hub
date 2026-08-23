import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { PlanningPoker } from "@/modules/tracker/planning-poker";
import { useSprintStore } from "@/stores/sprint-store";
import { renderWithProviders } from "../support/render";

const getPokerSession = vi.fn();

vi.mock("@/api/sprint", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/sprint")>();
	return {
		...actual,
		getPokerSession: (...args: unknown[]) => getPokerSession(...args),
		castVote: vi.fn(),
		revealVotes: vi.fn(),
	};
});

describe("PlanningPoker error≠empty (FE-TA-003 / TASK-187)", () => {
	beforeEach(() => {
		getPokerSession.mockReset();
		useSprintStore.setState({ pokerSession: null });
	});

	it("network fail → alert+retry, not «сессия не найдена»", async () => {
		getPokerSession.mockRejectedValue(new Error("network"));

		renderWithProviders(
			<PlanningPoker sessionId="sess-1" currentUserId="u-1" />,
		);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить сессию/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(
			screen.queryByText(/Сессия не найдена/),
		).not.toBeInTheDocument();
	});

	it("success null session → «не найдена», not alert", async () => {
		getPokerSession.mockResolvedValue(null);

		renderWithProviders(
			<PlanningPoker sessionId="sess-1" currentUserId="u-1" />,
		);

		await waitFor(() => {
			expect(screen.getByText(/Сессия не найдена/)).toBeInTheDocument();
		});
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});
});
