import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Message } from "@/api/messenger";
import { useMessages } from "@/hooks/use-messages";
import { useWsStore } from "@/stores/ws-store";

type MessageCache = {
	pages: Message[][];
	pageParams: unknown[];
};

const queryKey = ["messenger", "messages", "ch-1"];

function message(id: string, contentText: string): Message {
	return {
		id,
		channel_id: "ch-1",
		sender_id: "user-1",
		content_text: contentText,
		content_json: { type: "paragraph", text: contentText },
		created_at: "2026-08-14T10:00:00Z",
		reply_count: 0,
	};
}

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getMessages: vi.fn().mockResolvedValue([]),
	};
});

describe("useMessages cache events", () => {
	let queryClient: QueryClient;

	beforeEach(() => {
		queryClient = new QueryClient({
			defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
		});
		useWsStore.setState({
			connected: false,
			socket: null,
			listeners: new Set(),
		});
	});

	afterEach(() => {
		cleanup();
		queryClient.clear();
		useWsStore.setState({ connected: false, socket: null, listeners: new Set() });
	});

	function wrapper({ children }: { children: ReactNode }) {
		return (
			<QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
		);
	}

	function seed(pages: Message[][]): void {
		queryClient.setQueryData<MessageCache>(queryKey, {
			pages,
			pageParams: pages.map(() => undefined),
		});
	}

	function cached(): MessageCache | undefined {
		return queryClient.getQueryData<MessageCache>(queryKey);
	}

	function publish(event: string, payload: unknown): void {
		act(() => {
			for (const listener of useWsStore.getState().listeners) {
				listener(event, payload);
			}
		});
	}

	async function renderMessages() {
		const hook = renderHook(() => useMessages("ch-1"), { wrapper });
		await waitFor(() => expect(hook.result.current.isSuccess).toBe(true));
		expect(useWsStore.getState().listeners.size).toBeGreaterThan(0);
		return hook;
	}

	it("appends a target event once, dedupes across pages, and ignores a foreign channel", async () => {
		await renderMessages();
		seed([[message("m-duplicate", "older")], [message("m-1", "latest")]]);

		publish("message_sent", {
			id: "m-duplicate",
			channel_id: "ch-1",
			sender_id: "user-1",
			content_text: "older duplicate",
		});
		publish("message_sent", {
			id: "m-2",
			channel_id: "ch-1",
			sender_id: "user-2",
			content_text: "new target message",
		});
		publish("message_sent", {
			id: "m-2",
			channel_id: "ch-1",
			sender_id: "user-2",
			content_text: "duplicate target message",
		});
		publish("message_sent", {
			id: "foreign-message",
			channel_id: "ch-foreign",
			sender_id: "user-3",
			content_text: "must not leak into ch-1",
		});

		expect(cached()?.pages.map((page) => page.map((item) => item.id))).toEqual([
			["m-duplicate"],
			["m-1", "m-2"],
		]);
	});

	it("edits only the target message and leaves a foreign event untouched", async () => {
		await renderMessages();
		seed([[message("m-1", "first"), message("m-2", "before edit")]]);

		publish("message_edited", {
			channel_id: "ch-1",
			message_id: "m-2",
			content_text: "after edit",
			content_json: { type: "paragraph", text: "after edit" },
			edited_at: "2026-08-14T11:00:00Z",
		});
		publish("message_edited", {
			channel_id: "ch-foreign",
			message_id: "m-1",
			content_text: "foreign overwrite",
		});

		expect(cached()?.pages[0]).toEqual([
			message("m-1", "first"),
			{
				...message("m-2", "before edit"),
				content_text: "after edit",
				content_json: { type: "paragraph", text: "after edit" },
				edited_at: "2026-08-14T11:00:00Z",
			},
		]);
	});

	it("deletes the target message and ignores a foreign delete", async () => {
		await renderMessages();
		seed([[message("m-1", "first"), message("m-2", "remove me")]]);

		publish("message_deleted", {
			channel_id: "ch-1",
			message_id: "m-2",
		});
		publish("message_deleted", {
			channel_id: "ch-foreign",
			message_id: "m-1",
		});

		expect(cached()?.pages[0]).toEqual([message("m-1", "first")]);
	});
});
