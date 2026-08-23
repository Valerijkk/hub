import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as messengerApi from "@/api/messenger";
import { usePresence } from "@/hooks/use-presence";

vi.mock("@/api/messenger", async () => {
	const actual = await vi.importActual<typeof import("@/api/messenger")>(
		"@/api/messenger",
	);
	return {
		...actual,
		getPresence: vi.fn(),
		sendHeartbeat: vi.fn().mockResolvedValue(undefined),
	};
});

function wrapper({ children }: { children: ReactNode }) {
	const qc = new QueryClient({
		defaultOptions: { queries: { retry: false } },
	});
	return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

describe("usePresence isError (A2W11-3-001 / TASK-093)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("exposes isError when presence poll fails — not silent offline-only", async () => {
		vi.mocked(messengerApi.getPresence).mockRejectedValue(new Error("network"));

		const { result } = renderHook(() => usePresence(["user-1"]), { wrapper });

		await waitFor(() => {
			expect(result.current.isError).toBe(true);
		});
		expect(result.current.isLoading).toBe(false);
		expect(typeof result.current.refetch).toBe("function");
	});

	it("logs heartbeat failures instead of empty catch (A2W11-3-003)", async () => {
		vi.mocked(messengerApi.getPresence).mockResolvedValue({});
		vi.mocked(messengerApi.sendHeartbeat).mockRejectedValue(
			new Error("heartbeat down"),
		);
		const warn = vi.spyOn(console, "warn").mockImplementation(() => undefined);

		renderHook(() => usePresence(["user-1"]), { wrapper });

		await waitFor(() => {
			expect(warn).toHaveBeenCalledWith(
				"[presence] heartbeat failed",
				expect.any(Number),
				expect.any(Error),
			);
		});

		warn.mockRestore();
	});
});
