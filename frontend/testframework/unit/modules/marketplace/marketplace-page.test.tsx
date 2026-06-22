import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as mpApi from "@/api/marketplace";
import { MarketplacePage } from "@/modules/marketplace/marketplace-page";
import { useAuthStore } from "@/stores/auth-store";

vi.mock("@/api/marketplace");

describe("MarketplacePage", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		useAuthStore.setState({
			user: {
				id: "u1",
				email: "dev@hub.local",
				system_role: "member",
			} as never,
		});
		vi.mocked(mpApi.listMarketplaceListings).mockResolvedValue({
			items: [],
			total: 0,
		});
		vi.mocked(mpApi.listMyListings).mockResolvedValue({ items: [] });
	});

	it("renders marketplace heading and form", async () => {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<MarketplacePage />
			</QueryClientProvider>,
		);

		expect(
			screen.getByRole("heading", { name: /Барахолка/i }),
		).toBeInTheDocument();
		await waitFor(() => {
			expect(mpApi.listMarketplaceListings).toHaveBeenCalled();
		});
	});
});
