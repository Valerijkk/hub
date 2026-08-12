import { beforeEach, describe, expect, it, vi } from "vitest";
import * as callsApi from "@/api/calls";
import { useCallStore } from "@/stores/call-store";

vi.mock("@/api/calls", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/calls")>();
	return {
		...actual,
		raiseHand: vi.fn(),
	};
});

describe("P2-FE-02 · toggleHand feedback (W11-A3-008)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		useCallStore.setState({
			activeCall: {
				id: "call-1",
				invite_code: "abc",
				status: "active",
			} as never,
			handRaised: false,
			error: null,
		});
	});

	it("sets store error when raiseHand fails and does not flip hand", async () => {
		vi.mocked(callsApi.raiseHand).mockRejectedValue(new Error("denied"));

		await useCallStore.getState().toggleHand();

		expect(useCallStore.getState().handRaised).toBe(false);
		expect(useCallStore.getState().error).toBe("denied");
	});

	it("flips handRaised on success", async () => {
		vi.mocked(callsApi.raiseHand).mockResolvedValue(undefined as never);

		await useCallStore.getState().toggleHand();

		expect(useCallStore.getState().handRaised).toBe(true);
		expect(useCallStore.getState().error).toBeNull();
	});
});
