import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { PlanningPoker } from "@/modules/tracker/planning-poker";
import { useSprintStore } from "@/stores/sprint-store";
import { renderWithProviders } from "../../render";

const getPokerSession = vi.fn();
const castVote = vi.fn();

vi.mock("@/api/sprint", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/sprint")>();
	return {
		...actual,
		getPokerSession: (...args: unknown[]) => getPokerSession(...args),
		castVote: (...args: unknown[]) => castVote(...args),
		revealVotes: vi.fn(),
	};
});

describe("PlanningPoker radiogroup (A2W10-A3-006 / TASK-093)", () => {
	beforeEach(() => {
		getPokerSession.mockReset();
		castVote.mockReset();
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

	it("shows a visible card focus indicator for keyboard focus", async () => {
		getPokerSession.mockResolvedValue({
			id: "sess-1",
			revealed: false,
			voters: [],
			votes: {},
		});
		castVote.mockResolvedValue(undefined);

		renderWithProviders(
			<PlanningPoker sessionId="sess-1" currentUserId="u-1" />,
		);

		await waitFor(() => {
			expect(screen.getAllByRole("radio").length).toBeGreaterThanOrEqual(5);
		});

		const firstRadio = screen.getAllByRole("radio")[0];
		const firstCard = firstRadio?.closest("label");
		expect(firstCard).not.toBeNull();
		expect(firstCard?.className).toMatch(/focus-within:ring-2/);
		expect(firstCard?.className).toMatch(/focus-within:ring-\[var\(--blue\)\]/);

		firstRadio?.focus();
		expect(firstRadio).toHaveFocus();
		fireEvent.click(firstRadio as HTMLInputElement);

		await waitFor(() => {
			expect(castVote).toHaveBeenCalledWith("sess-1", { value: "0" });
		});
	});
});
