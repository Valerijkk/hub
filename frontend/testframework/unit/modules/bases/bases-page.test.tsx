import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as basesApi from "@/api/bases";
import { BasesPage } from "@/modules/bases/bases-page";

vi.mock("@/api/bases");

describe("BasesPage", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(basesApi.listBases).mockResolvedValue([
			{
				id: "b1",
				name: "Inventory",
				version: 1,
				owner_id: "u1",
				table_count: 1,
				created_at: "2026-06-25T00:00:00Z",
				updated_at: "2026-06-25T00:00:00Z",
			},
		]);
	});

	it("renders Hub Base heading and loads bases", async () => {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<BasesPage />
			</QueryClientProvider>,
		);

		expect(
			screen.getByRole("heading", { name: /Hub Base/i }),
		).toBeInTheDocument();
		await waitFor(() => {
			expect(basesApi.listBases).toHaveBeenCalled();
		});
		await waitFor(() => {
			expect(screen.getByText("Inventory")).toBeInTheDocument();
		});
	});
});
