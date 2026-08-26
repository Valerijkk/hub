import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as messengerApi from "@/api/messenger";
import { unreadQueryKey, useMarkRead, useUnread } from "@/hooks/use-unread";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getUnreadCounts: vi.fn(),
		markChannelRead: vi.fn(),
	};
});

const snapshot = {
	channels: { "ch-1": 2 },
	dms: { "dm-1": 3 },
	mentions: { "ch-1": 1 },
	threads: { "root-1": 4 },
	any_unread: true,
};

describe("Phase 151 server-authoritative unread", () => {
	let queryClient: QueryClient;
	let invalidateQueries: ReturnType<typeof vi.spyOn>;

	beforeEach(() => {
		vi.clearAllMocks();
		useMessengerStore.getState().reset();
		useAuthStore.setState({
			user: {
				id: "user-1",
				email: "u@hub.local",
				timezone: "UTC",
				language: "en",
				system_role: "member",
				onboarding_completed: true,
			},
		});
		queryClient = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		invalidateQueries = vi.spyOn(queryClient, "invalidateQueries");
		vi.mocked(messengerApi.getUnreadCounts).mockResolvedValue(snapshot);
	});

	function wrapper({ children }: { children: ReactNode }) {
		return (
			<QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
		);
	}

	it("stores all four maps from one authoritative response", async () => {
		renderHook(() => useUnread(), { wrapper });

		await waitFor(() => {
			expect(useMessengerStore.getState().unreadSnapshot).toEqual(snapshot);
		});
	});

	it("does not mutate local unread state while a channel receipt is pending", async () => {
		useMessengerStore.getState().setUnreadSnapshot(snapshot);
		let resolveMark: () => void = () => undefined;
		vi.mocked(messengerApi.markChannelRead).mockImplementation(
			() => new Promise<void>((resolve) => (resolveMark = resolve)),
		);

		const { result } = renderHook(() => useMarkRead(), { wrapper });
		let pending: Promise<void> | undefined;
		act(() => {
			pending = result.current("ch-1");
		});

		expect(useMessengerStore.getState().unreadSnapshot).toEqual(snapshot);

		await act(async () => {
			resolveMark();
			await pending;
		});

		expect(useMessengerStore.getState().unreadSnapshot).toEqual(snapshot);
		expect(invalidateQueries).toHaveBeenCalledWith({
			queryKey: unreadQueryKey("user-1"),
			exact: true,
		});
	});

	it("keeps the server snapshot unchanged when the receipt fails", async () => {
		useMessengerStore.getState().setUnreadSnapshot(snapshot);
		vi.mocked(messengerApi.markChannelRead).mockRejectedValue(
			new Error("receipt failed"),
		);

		const { result } = renderHook(() => useMarkRead(), { wrapper });
		await act(async () => {
			await expect(result.current("ch-1")).rejects.toThrow("receipt failed");
		});

		expect(useMessengerStore.getState().unreadSnapshot).toEqual(snapshot);
		expect(invalidateQueries).not.toHaveBeenCalled();
	});
});
