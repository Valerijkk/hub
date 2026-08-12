import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CallResponse } from "@/api/calls";
import * as callsApi from "@/api/calls";
import { useCallStore } from "@/stores/call-store";

vi.mock("@/api/calls", () => ({
	createCall: vi.fn(),
	createInstantCall: vi.fn(),
	joinByInviteCode: vi.fn(),
	joinCall: vi.fn(),
	leaveCall: vi.fn(),
	endCall: vi.fn(),
	getActiveCall: vi.fn(),
	raiseHand: vi.fn(),
}));

const activeCall: CallResponse = {
	id: "call-1",
	room_name: "room-1",
	channel_id: "ch-1",
	status: "active",
	created_by: "user-1",
	participant_count: 1,
	token: "lk-token",
	created_at: "2026-07-28T10:00:00Z",
	updated_at: "2026-07-28T10:00:00Z",
};

describe("call-store lifecycle (FE-P1-CALLS-LIFECYCLE)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		useCallStore.getState().reset();
	});

	it("checkActiveCall returns null on 404 without setting error", async () => {
		vi.mocked(callsApi.getActiveCall).mockResolvedValue(null);

		const call = await useCallStore.getState().checkActiveCall("ch-1");

		expect(call).toBeNull();
		expect(useCallStore.getState().error).toBeNull();
	});

	it("checkActiveCall propagates non-404 errors and does not return null as empty", async () => {
		vi.mocked(callsApi.getActiveCall).mockRejectedValue({
			code: 503,
			message: "service unavailable",
		});

		await expect(
			useCallStore.getState().checkActiveCall("ch-1"),
		).rejects.toMatchObject({ code: 503 });

		expect(useCallStore.getState().error).toBe("service unavailable");
	});

	it("endCall does not reset local state when API fails", async () => {
		useCallStore.setState({
			activeCall,
			token: "lk-token",
			isInCall: true,
			isConnecting: false,
			error: null,
		});
		vi.mocked(callsApi.endCall).mockRejectedValue({
			code: 500,
			message: "end failed",
		});

		await useCallStore.getState().endCall();

		const state = useCallStore.getState();
		expect(state.isInCall).toBe(true);
		expect(state.activeCall?.id).toBe("call-1");
		expect(state.error).toBe("end failed");
	});

	it("leaveCall resets only after successful API leave", async () => {
		useCallStore.setState({
			activeCall,
			token: "lk-token",
			isInCall: true,
			isConnecting: false,
			error: null,
		});
		vi.mocked(callsApi.leaveCall).mockResolvedValue(undefined);

		await useCallStore.getState().leaveCall();

		expect(useCallStore.getState().isInCall).toBe(false);
		expect(useCallStore.getState().activeCall).toBeNull();
	});

	it("leaveCall keeps in-call state when API fails", async () => {
		useCallStore.setState({
			activeCall,
			token: "lk-token",
			isInCall: true,
			isConnecting: false,
			error: null,
		});
		vi.mocked(callsApi.leaveCall).mockRejectedValue({
			code: 500,
			message: "leave failed",
		});

		await useCallStore.getState().leaveCall();

		expect(useCallStore.getState().isInCall).toBe(true);
		expect(useCallStore.getState().activeCall?.id).toBe("call-1");
		expect(useCallStore.getState().error).toBe("leave failed");
	});

	it("joinCall sets active call from API", async () => {
		vi.mocked(callsApi.joinCall).mockResolvedValue(activeCall);

		await useCallStore.getState().joinCall("call-1");

		expect(callsApi.joinCall).toHaveBeenCalledWith("call-1");
		expect(useCallStore.getState().isInCall).toBe(true);
		expect(useCallStore.getState().activeCall?.id).toBe("call-1");
	});
});
