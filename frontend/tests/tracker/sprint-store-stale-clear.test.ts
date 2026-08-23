import { beforeEach, describe, expect, it, vi } from "vitest";
import { useSprintStore } from "@/stores/sprint-store";

const listSprints = vi.fn();
const getVelocity = vi.fn();

vi.mock("@/api/sprint", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/sprint")>();
	return {
		...actual,
		listSprints: (...args: unknown[]) => listSprints(...args),
		getVelocity: (...args: unknown[]) => getVelocity(...args),
	};
});

const staleSprint = {
	id: "stale-sprint",
	name: "Other project sprint",
	status: "planning" as const,
	start_date: "2026-01-01",
	end_date: "2026-01-14",
	issue_count: 3,
	version: 1,
	created_at: "2026-01-01T00:00:00Z",
	updated_at: "2026-01-01T00:00:00Z",
};

describe("FE · sprint-store cross-project clear (A3W10-3-003)", () => {
	beforeEach(() => {
		listSprints.mockReset();
		getVelocity.mockReset();
		useSprintStore.setState({
			sprints: [staleSprint],
			activeSprint: staleSprint,
			velocity: [{ sprint_name: "Old", committed: 5, completed: 4 }],
			loading: false,
			velocityLoading: false,
			error: null,
		});
	});

	it("clears sprints immediately when fetchSprints starts", async () => {
		let resolveList: (value: typeof staleSprint[]) => void = () => {};
		listSprints.mockImplementation(
			() =>
				new Promise((resolve) => {
					resolveList = resolve;
				}),
		);

		const pending = useSprintStore.getState().fetchSprints("proj-new");

		expect(useSprintStore.getState().sprints).toEqual([]);
		expect(useSprintStore.getState().activeSprint).toBeNull();
		expect(useSprintStore.getState().loading).toBe(true);

		resolveList([]);
		await pending;
		expect(useSprintStore.getState().loading).toBe(false);

		// Mutation: не сбрасывать sprints при старте fetch → staleSprint остаётся → FAIL.
	});

	it("does not keep previous project sprints when fetchSprints fails", async () => {
		listSprints.mockRejectedValue(new Error("network"));

		await useSprintStore.getState().fetchSprints("proj-new");

		expect(useSprintStore.getState().sprints).toEqual([]);
		expect(useSprintStore.getState().activeSprint).toBeNull();
		expect(useSprintStore.getState().error).toBe(
			"Не удалось загрузить спринты",
		);
	});

	it("clears velocity immediately when fetchVelocity starts", async () => {
		let resolveVel: (value: unknown[]) => void = () => {};
		getVelocity.mockImplementation(
			() =>
				new Promise((resolve) => {
					resolveVel = resolve;
				}),
		);

		const pending = useSprintStore.getState().fetchVelocity("proj-new");

		expect(useSprintStore.getState().velocity).toEqual([]);
		expect(useSprintStore.getState().velocityLoading).toBe(true);

		resolveVel([{ sprint_name: "N1", committed: 1, completed: 1 }]);
		await pending;

		expect(useSprintStore.getState().velocity).toEqual([
			{ sprint_name: "N1", committed: 1, completed: 1 },
		]);
		expect(useSprintStore.getState().velocityLoading).toBe(false);
	});
});
