import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as mpApi from "@/api/marketplace";
import { MarketplacePage } from "@/modules/marketplace/marketplace-page";
import { useAuthStore } from "@/stores/auth-store";

vi.mock("@/api/marketplace");

/** A3W10-1-004: admin pending queue fail must keep section + alert. */
describe("MarketplacePage pending error≠empty (A3W10-1-004)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		useAuthStore.setState({
			user: {
				id: "u1",
				email: "admin@hub.local",
				system_role: "admin",
			} as never,
		});
		vi.mocked(mpApi.listMarketplaceListings).mockResolvedValue({
			items: [],
			total: 0,
		});
		vi.mocked(mpApi.listMyListings).mockResolvedValue({ items: [] });
		vi.mocked(mpApi.listPendingListings).mockRejectedValue(new Error("network"));
	});

	it("shows moderation queue alert when listPendingListings fails", async () => {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<MarketplacePage />
			</QueryClientProvider>,
		);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить очередь модерации/i,
			);
		});
		expect(
			screen.getByRole("heading", { name: /очередь модерации/i }),
		).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});
});
