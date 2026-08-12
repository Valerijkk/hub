import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useMessages } from "@/hooks/use-messages";
import { useTypingIndicator } from "@/hooks/use-typing";
import { useMessengerStore } from "@/stores/messenger-store";
import { useWsStore } from "@/stores/ws-store";

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getMessages: vi.fn().mockResolvedValue([]),
	};
});

/**
 * P2-FE-10 / W11-A3-010,011: null/non-object WS payload не должен
 * бросать TypeError в listener (соседние handlers остаются живы).
 */
describe("WS payload guards (P2-FE-10)", () => {
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
		useMessengerStore.setState({ typingUsers: {} });
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

	it("useMessages: payload null → no throw; valid event still handled", () => {
		const { result } = renderHook(() => useMessages("ch-1"), { wrapper });
		expect(result.current.messages).toEqual([]);

		const listeners = [...useWsStore.getState().listeners];
		expect(listeners.length).toBeGreaterThan(0);

		expect(() => {
			act(() => {
				for (const listener of listeners) {
					listener("message_sent", null);
					listener("message_edited", undefined);
					listener("message_deleted", "not-an-object");
					listener("reaction_toggled", []);
				}
			});
		}).not.toThrow();

		// Соседний валидный handler-путь жив: setQueryData не падает.
		expect(() => {
			act(() => {
				for (const listener of listeners) {
					listener("message_sent", {
						id: "m-1",
						channel_id: "ch-1",
						sender_id: "u-1",
						content_text: "hi",
						content_json: {},
						created_at: "2026-07-28T12:00:00Z",
					});
				}
			});
		}).not.toThrow();
	});

	it("useTypingIndicator: null payload → no throw", () => {
		renderHook(() => useTypingIndicator("ch-1"), { wrapper });
		const listeners = [...useWsStore.getState().listeners];
		expect(listeners.length).toBeGreaterThan(0);

		expect(() => {
			act(() => {
				for (const listener of listeners) {
					listener("typing", null);
					listener("typing", undefined);
					listener("typing", []);
				}
			});
		}).not.toThrow();

		expect(useMessengerStore.getState().typingUsers["ch-1"] ?? []).toEqual([]);
	});
});
