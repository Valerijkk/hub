import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Message } from "@/api/messenger";
import { useMessages } from "@/hooks/use-messages";
import { useWsStore } from "@/stores/ws-store";

const { baseMessage } = vi.hoisted(() => {
	const baseMessage: Message = {
		id: "m-1",
		channel_id: "ch-1",
		sender_id: "u-1",
		content_text: "hi",
		content_json: {},
		created_at: "2026-07-28T12:00:00Z",
		reply_count: 0,
		reactions: [{ emoji: "👍", user_ids: ["u-1"] }],
	};
	return { baseMessage };
});

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getMessages: vi.fn().mockResolvedValue([baseMessage]),
	};
});

/**
 * A3W11-3-002 / TASK-131: reaction_toggled обновляет кэш только при
 * Array.isArray(reactions); иначе skip (не затирает кэш).
 */
describe("A3W11-3-002 · reaction_toggled reactions guard", () => {
	let queryClient: QueryClient;

	beforeEach(() => {
		queryClient = new QueryClient({
			defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
		});
		useWsStore.setState({
			connected: true,
			socket: null,
			listeners: new Set(),
		});
	});

	afterEach(() => {
		queryClient.clear();
		useWsStore.setState({ listeners: new Set(), connected: false });
	});

	function wrapper({ children }: { children: ReactNode }) {
		return (
			<QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
		);
	}

	function seedCache() {
		queryClient.setQueryData(["messenger", "messages", "ch-1"], {
			pages: [[{ ...baseMessage }]],
			pageParams: [undefined],
		});
	}

	function cachedReactions() {
		const data = queryClient.getQueryData<{
			pages: Message[][];
		}>(["messenger", "messages", "ch-1"]);
		return data?.pages[0]?.[0]?.reactions;
	}

	it("skips update when reactions is undefined or non-array", async () => {
		const { result } = renderHook(() => useMessages("ch-1"), { wrapper });
		await waitFor(() => expect(result.current.isSuccess).toBe(true));
		seedCache();

		const listeners = [...useWsStore.getState().listeners];
		expect(listeners.length).toBeGreaterThan(0);

		act(() => {
			for (const listener of listeners) {
				listener("reaction_toggled", {
					channel_id: "ch-1",
					message_id: "m-1",
					reactions: undefined,
				});
				listener("reaction_toggled", {
					channel_id: "ch-1",
					message_id: "m-1",
					reactions: { emoji: "🔥" },
				});
			}
		});

		expect(cachedReactions()).toEqual([
			{ emoji: "👍", user_ids: ["u-1"] },
		]);
	});

	it("updates cache when reactions is a valid array", async () => {
		const { result } = renderHook(() => useMessages("ch-1"), { wrapper });
		await waitFor(() => expect(result.current.isSuccess).toBe(true));
		seedCache();

		const listeners = [...useWsStore.getState().listeners];
		const next = [{ emoji: "🎉", user_ids: ["u-2"] }];

		act(() => {
			for (const listener of listeners) {
				listener("reaction_toggled", {
					channel_id: "ch-1",
					message_id: "m-1",
					reactions: next,
				});
			}
		});

		expect(cachedReactions()).toEqual(next);
	});

	it("updates cache when reactions is empty array", async () => {
		const { result } = renderHook(() => useMessages("ch-1"), { wrapper });
		await waitFor(() => expect(result.current.isSuccess).toBe(true));
		seedCache();

		const listeners = [...useWsStore.getState().listeners];

		act(() => {
			for (const listener of listeners) {
				listener("reaction_toggled", {
					channel_id: "ch-1",
					message_id: "m-1",
					reactions: [],
				});
			}
		});

		expect(cachedReactions()).toEqual([]);
	});
});
