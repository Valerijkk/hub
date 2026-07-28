import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as messengerApi from "@/api/messenger";
import { useMarkRead } from "@/hooks/use-unread";
import { useMessengerStore } from "@/stores/messenger-store";

vi.mock("@/api/messenger", () => ({
	getUnreadCounts: vi.fn(),
	markChannelRead: vi.fn(),
}));

describe("P2-FE-02 · useMarkRead rollback (W11-A4-004)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		useMessengerStore.setState({ unreadCounts: { "ch-1": 4, "ch-2": 1 } });
	});

	it("rolls back badge when markChannelRead fails", async () => {
		vi.mocked(messengerApi.markChannelRead).mockRejectedValue(
			new Error("network"),
		);

		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		const wrapper = ({ children }: { children: ReactNode }) => (
			<QueryClientProvider client={qc}>{children}</QueryClientProvider>
		);

		const { result } = renderHook(() => useMarkRead(), { wrapper });

		await act(async () => {
			await result.current("ch-1");
		});

		await waitFor(() => {
			expect(useMessengerStore.getState().unreadCounts["ch-1"]).toBe(4);
		});
		expect(useMessengerStore.getState().unreadCounts["ch-2"]).toBe(1);
	});

	it("keeps zero after successful markRead", async () => {
		vi.mocked(messengerApi.markChannelRead).mockResolvedValue(undefined);

		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		const wrapper = ({ children }: { children: ReactNode }) => (
			<QueryClientProvider client={qc}>{children}</QueryClientProvider>
		);

		const { result } = renderHook(() => useMarkRead(), { wrapper });

		await act(async () => {
			await result.current("ch-1");
		});

		expect(messengerApi.markChannelRead).toHaveBeenCalledWith("ch-1");
		expect(useMessengerStore.getState().unreadCounts["ch-1"]).toBe(0);
	});
});
