import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as messengerApi from "@/api/messenger";
import { unreadQueryKey, useMarkRead } from "@/hooks/use-unread";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

vi.mock("@/api/messenger", () => ({
	markChannelRead: vi.fn(),
}));

const snapshot = {
	channels: { "ch-1": 4, "ch-2": 1 },
	dms: {},
	mentions: {},
	threads: {},
	any_unread: true,
};

describe("Phase 151 channel read receipt", () => {
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
		useMessengerStore.getState().setUnreadSnapshot(snapshot);
	});

	function setup() {
		const queryClient = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		const invalidateQueries = vi.spyOn(queryClient, "invalidateQueries");
		const wrapper = ({ children }: { children: ReactNode }) => (
			<QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
		);
		return { queryClient, invalidateQueries, wrapper };
	}

	it("keeps the authoritative snapshot and surfaces a failed receipt", async () => {
		vi.mocked(messengerApi.markChannelRead).mockRejectedValue(
			new Error("network"),
		);
		const { wrapper } = setup();
		const { result } = renderHook(() => useMarkRead(), { wrapper });

		await act(async () => {
			await expect(result.current("ch-1")).rejects.toThrow("network");
		});

		expect(useMessengerStore.getState().unreadSnapshot).toEqual(snapshot);
	});

	it("keeps the snapshot until exact authoritative refetch after success", async () => {
		vi.mocked(messengerApi.markChannelRead).mockResolvedValue(undefined);
		const { invalidateQueries, wrapper } = setup();
		const { result } = renderHook(() => useMarkRead(), { wrapper });

		await act(async () => {
			await result.current("ch-1");
		});

		expect(messengerApi.markChannelRead).toHaveBeenCalledWith("ch-1");
		expect(useMessengerStore.getState().unreadSnapshot).toEqual(snapshot);
		expect(invalidateQueries).toHaveBeenCalledWith({
			queryKey: unreadQueryKey("user-1"),
			exact: true,
		});
	});
});