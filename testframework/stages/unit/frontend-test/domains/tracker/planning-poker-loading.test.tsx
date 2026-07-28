import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { PlanningPoker } from "@/modules/tracker/planning-poker";
import { useSprintStore } from "@/stores/sprint-store";
import { renderWithProviders } from "@testframework/render";

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

describe("FE-P1-SPRINT-UI · poker loading vs empty (W10-A3-002)", () => {
	beforeEach(() => {
		getPokerSession.mockReset();
		useSprintStore.setState({ pokerSession: null });
	});

	it("shows loading before response, not «не найдена»", async () => {
		let resolveSession!: (v: unknown) => void;
		getPokerSession.mockReturnValue(
			new Promise((resolve) => {
				resolveSession = resolve;
			}),
		);

		renderWithProviders(
			<PlanningPoker sessionId="sess-1" currentUserId="u-1" />,
		);

		expect(screen.getByText("Загрузка сессии...")).toBeInTheDocument();
		expect(
			screen.queryByText(/Сессия не найдена/),
		).not.toBeInTheDocument();

		resolveSession({
			id: "sess-1",
			revealed: false,
			voters: [],
			votes: {},
		});

		await waitFor(() => {
			expect(screen.getByText("Planning Poker")).toBeInTheDocument();
		});
	});

	it("shows not-found only after failed fetch", async () => {
		getPokerSession.mockRejectedValue(new Error("gone"));

		renderWithProviders(
			<PlanningPoker sessionId="sess-1" currentUserId="u-1" />,
		);

		await waitFor(() => {
			expect(screen.getByText(/Сессия не найдена/)).toBeInTheDocument();
		});
	});
});
