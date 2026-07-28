import { beforeEach, describe, expect, it, vi } from "vitest";
import { useSprintStore } from "@/stores/sprint-store";

const listSprints = vi.fn();

vi.mock("@/api/sprint", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/sprint")>();
	return {
		...actual,
		listSprints: (...args: unknown[]) => listSprints(...args),
	};
});

describe("FE-P1-SPRINT-UI · fetchSprints error (W11-A3-005)", () => {
	beforeEach(() => {
		listSprints.mockReset();
		useSprintStore.setState({
			sprints: [],
			loading: false,
			error: null,
			activeSprint: null,
		});
	});

	it("sets error on catch instead of silent empty list", async () => {
		listSprints.mockRejectedValue(new Error("network"));
		await useSprintStore.getState().fetchSprints("proj-1");
		expect(useSprintStore.getState().error).toBe(
			"Не удалось загрузить спринты",
		);
		expect(useSprintStore.getState().loading).toBe(false);
		expect(useSprintStore.getState().sprints).toEqual([]);
	});

	it("mutation: without error field empty[] looks like «нет спринтов»", async () => {
		listSprints.mockRejectedValue(new Error("network"));
		await useSprintStore.getState().fetchSprints("proj-1");
		const { sprints, error } = useSprintStore.getState();
		expect(sprints).toEqual([]);
		expect(error).toBeTruthy();
	});

	it("clears error on successful fetch", async () => {
		listSprints.mockResolvedValue([]);
		useSprintStore.setState({ error: "stale" });
		await useSprintStore.getState().fetchSprints("proj-1");
		expect(useSprintStore.getState().error).toBeNull();
	});
});
