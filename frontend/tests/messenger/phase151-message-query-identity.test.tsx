import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Message } from "@/api/messenger";
import { useMessages } from "@/hooks/use-messages";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { useWsStore } from "@/stores/ws-store";

const api = vi.hoisted(() => ({
	getMessages: vi.fn(),
}));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getMessages: api.getMessages,
	};
});

const channelId = "shared-channel";
const userA = {
	id: "user-a",
	email: "user-a@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};
const userB = {
	...userA,
	id: "user-b",
	email: "user-b@example.test",
};

function message(id: string, text: string): Message {
	return {
		id,
		channel_id: channelId,
		sender_id: "sender-1",
		content_text: text,
		content_json: { type: "doc", text },
		reply_count: 0,
		created_at: "2026-08-26T06:00:00Z",
	};
}

function deferred<T>() {
	let resolve: (value: T) => void = () => undefined;
	const promise = new Promise<T>((done) => {
		resolve = done;
	});
	return { promise, resolve };
}

describe("Phase 151 root message query identity boundary", () => {
	let queryClient: QueryClient;

	beforeEach(() => {
		queryClient = new QueryClient({
			defaultOptions: {
				queries: { retry: false },
				mutations: { retry: false },
			},
		});
		api.getMessages.mockReset();
		useAuthStore.setState({ user: userA });
		useMessengerStore.getState().reset();
		useWsStore.setState({ connected: false, socket: null, listeners: new Set() });
	});

	afterEach(() => {
		cleanup();
		queryClient.clear();
		useAuthStore.setState({ user: null });
		useMessengerStore.getState().reset();
		useWsStore.setState({ connected: false, socket: null, listeners: new Set() });
		vi.clearAllMocks();
	});

	function TestProvider({ children }: { children: ReactNode }) {
		return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
	}

	it("never renders a delayed A response after switching the same channel to B", async () => {
		const fromA = deferred<Message[]>();
		const fromB = deferred<Message[]>();
		const privateA = message("message-a", "Private timeline A");
		const privateB = message("message-b", "Private timeline B");
		api.getMessages
			.mockImplementationOnce(() => fromA.promise)
			.mockImplementationOnce(() => fromB.promise);

		const hook = renderHook(() => useMessages(channelId), { wrapper: TestProvider });
		await waitFor(() => expect(api.getMessages).toHaveBeenCalledTimes(1));

		act(() => {
			useAuthStore.setState({ user: userB });
		});
		await waitFor(() => expect(api.getMessages).toHaveBeenCalledTimes(2));

		await act(async () => {
			fromA.resolve([privateA]);
			await fromA.promise;
		});
		expect(hook.result.current.messages).not.toContainEqual(privateA);

		await act(async () => {
			fromB.resolve([privateB]);
			await fromB.promise;
		});
		await waitFor(() => expect(hook.result.current.messages).toEqual([privateB]));

		expect(
			queryClient.getQueryData<{ pages: Message[][] }>([
				"messenger",
				"messages",
				channelId,
				userA.id,
			])?.pages.flat(),
		).toEqual([privateA]);
		expect(
			queryClient.getQueryData<{ pages: Message[][] }>([
				"messenger",
				"messages",
				channelId,
				userB.id,
			])?.pages.flat(),
		).toEqual([privateB]);
	});

	it("does not fetch or render legacy and sentinel caches without identity", async () => {
		const privateMessage = message("private-message", "Private cached timeline");
		queryClient.setQueryData(["messenger", "messages", channelId], {
			pages: [[privateMessage]],
			pageParams: [undefined],
		});
		queryClient.setQueryData(
			["messenger", "messages", channelId, "__no-user__"],
			{ pages: [[privateMessage]], pageParams: [undefined] },
		);
		useAuthStore.setState({ user: null });

		const hook = renderHook(() => useMessages(channelId), { wrapper: TestProvider });
		await act(async () => Promise.resolve());

		expect(api.getMessages).not.toHaveBeenCalled();
		expect(hook.result.current.data).toBeUndefined();
		expect(hook.result.current.messages).toEqual([]);
	});

	it("hides resident root messages immediately when channel access is revoked", () => {
		const privateMessage = message("private-message", "Private cached timeline");
		const queryKey = ["messenger", "messages", channelId, userA.id] as const;
		const cachedData = {
			pages: [[privateMessage]],
			pageParams: [undefined],
		};
		queryClient.setQueryDefaults(queryKey, { staleTime: Infinity });
		queryClient.setQueryData(queryKey, cachedData);

		const hook = renderHook(() => useMessages(channelId), { wrapper: TestProvider });

		expect(hook.result.current.data).toEqual(cachedData);
		expect(hook.result.current.messages).toEqual([privateMessage]);

		act(() => {
			useMessengerStore.getState().revokeChannel(channelId);
		});

		expect(hook.result.current.data).toBeUndefined();
		expect(hook.result.current.messages).toEqual([]);
	});
});
