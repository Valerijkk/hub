import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CallChatMessage } from "@/api/calls";
import { CallChat } from "@/components/calls/CallChat";
import { useWsStore } from "@/stores/ws-store";
import {
	createTestQueryClient,
	renderWithProviders,
} from "@testframework/render";

const CALL_ID = "call-aaa-111";

vi.mock("@/api/calls", () => ({
	listChatMessages: vi.fn(async () => [] as CallChatMessage[]),
	sendChatMessage: vi.fn(),
}));

describe("CallChat WS room + chat.message filter", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
		useWsStore.setState({
			connected: true,
			socket: null,
			listeners: new Set(),
		});
	});

	it("subscribes to call:{callId} room when WS connected", async () => {
		const subscribe = vi.fn();
		const unsubscribe = vi.fn();
		useWsStore.setState({
			connected: true,
			subscribe,
			unsubscribe,
		});

		renderWithProviders(<CallChat callId={CALL_ID} />);

		await waitFor(() => {
			expect(subscribe).toHaveBeenCalledWith(`call:${CALL_ID}`);
		});
	});

	it("appends chat.message with text and ignores call.joined", async () => {
		const queryClient = createTestQueryClient();
		queryClient.setQueryData<CallChatMessage[]>(["call-chat", CALL_ID], []);

		renderWithProviders(<CallChat callId={CALL_ID} />, { queryClient });

		await waitFor(() => {
			expect(useWsStore.getState().listeners.size).toBeGreaterThan(0);
		});

		const listeners = [...useWsStore.getState().listeners];

		for (const handler of listeners) {
			handler("call.joined", {
				call_id: CALL_ID,
				user_id: "u-1",
				user_name: "Alice",
			});
		}

		expect(screen.queryByText("Alice")).not.toBeInTheDocument();
		expect(queryClient.getQueryData(["call-chat", CALL_ID])).toEqual([]);

		const chatMsg: CallChatMessage = {
			id: "msg-1",
			call_id: CALL_ID,
			user_id: "u-2",
			user_name: "Bob",
			text: "Привет из звонка",
			created_at: "2026-07-28T12:00:00Z",
		};

		for (const handler of listeners) {
			handler("chat.message", chatMsg);
		}

		await waitFor(() => {
			expect(screen.getByText("Привет из звонка")).toBeInTheDocument();
			expect(screen.getByText("Bob")).toBeInTheDocument();
		});
	});

	it("does not append chat.message for another call_id", async () => {
		const queryClient = createTestQueryClient();
		queryClient.setQueryData<CallChatMessage[]>(["call-chat", CALL_ID], []);

		renderWithProviders(<CallChat callId={CALL_ID} />, { queryClient });

		await waitFor(() => {
			expect(useWsStore.getState().listeners.size).toBeGreaterThan(0);
		});

		for (const handler of [...useWsStore.getState().listeners]) {
			handler("chat.message", {
				id: "msg-other",
				call_id: "other-call",
				user_id: "u-9",
				user_name: "Eve",
				text: "чужое",
				created_at: "2026-07-28T12:00:00Z",
			});
		}

		expect(queryClient.getQueryData(["call-chat", CALL_ID])).toEqual([]);
		expect(screen.queryByText("чужое")).not.toBeInTheDocument();
	});
});
