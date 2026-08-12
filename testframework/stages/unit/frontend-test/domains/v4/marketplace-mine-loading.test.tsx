import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as mpApi from "@/api/marketplace";
import { MarketplacePage } from "@/modules/marketplace/marketplace-page";
import { useAuthStore } from "@/stores/auth-store";

vi.mock("@/api/marketplace");

/**
 * FE-SMA-003 / TASK-192: вкладка «Мои» — loading ≠ empty «нет объявлений».
 * Mutation: гейтить empty через !browseLoading вместо listLoading → RED.
 */
describe("MarketplacePage mine loading≠empty (FE-SMA-003 / TASK-192)", () => {
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
	});

	it("shows loading on mine tab while listMyListings pending, not empty copy", async () => {
		let resolveMine: ((v: { items: never[] }) => void) | undefined;
		vi.mocked(mpApi.listMyListings).mockImplementation(
			() =>
				new Promise((resolve) => {
					resolveMine = resolve;
				}),
		);

		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<MarketplacePage />
			</QueryClientProvider>,
		);

		await waitFor(() => {
			expect(mpApi.listMarketplaceListings).toHaveBeenCalled();
		});

		fireEvent.click(screen.getByRole("button", { name: "Мои" }));

		await waitFor(() => {
			expect(screen.getByText("Загрузка…")).toBeInTheDocument();
		});
		expect(screen.queryByText("У вас нет объявлений")).toBeNull();

		resolveMine?.({ items: [] });

		await waitFor(() => {
			expect(screen.getByText("У вас нет объявлений")).toBeInTheDocument();
		});
		expect(screen.queryByText("Загрузка…")).toBeNull();
	});
});
