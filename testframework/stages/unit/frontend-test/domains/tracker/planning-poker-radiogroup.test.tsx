import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { PlanningPoker } from "@/modules/tracker/planning-poker";
import { useSprintStore } from "@/stores/sprint-store";
import { renderWithProviders } from "../../render";

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

describe("PlanningPoker radiogroup (A2W10-A3-006 / TASK-093)", () => {
	beforeEach(() => {
		getPokerSession.mockReset();
		useSprintStore.setState({ pokerSession: null });
	});

	it("vote cards expose radiogroup + aria-checked", async () => {
		getPokerSession.mockResolvedValue({
			id: "sess-1",
			revealed: false,
			voters: [],
			votes: {},
		});

		renderWithProviders(
			<PlanningPoker sessionId="sess-1" currentUserId="u-1" />,
		);

		await waitFor(() => {
			expect(
				screen.getByRole("radiogroup", { name: /planning poker/i }),
			).toBeInTheDocument();
		});

		const radios = screen.getAllByRole("radio");
		expect(radios.length).toBeGreaterThanOrEqual(5);
		for (const radio of radios) {
			expect(radio).toHaveAttribute("aria-checked");
		}
	});
});
