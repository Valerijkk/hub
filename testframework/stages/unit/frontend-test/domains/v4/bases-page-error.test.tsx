import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as basesApi from "@/api/bases";
import { BasesPage } from "@/modules/bases/bases-page";

vi.mock("@/api/bases");

describe("BasesPage error≠empty (W10-A1-006 / TASK-047)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows alert + retry instead of empty sidebar when listBases fails", async () => {
		vi.mocked(basesApi.listBases).mockRejectedValue(new Error("network"));
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<BasesPage />
			</QueryClientProvider>,
		);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить базы/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});

	it("shows alert in main when getBase fails for selected base (A3W10-1-005)", async () => {
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
		vi.mocked(basesApi.getBase).mockRejectedValue(new Error("network"));

		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<BasesPage />
			</QueryClientProvider>,
		);

		fireEvent.click(await screen.findByRole("button", { name: "Inventory" }));

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить базу/i,
			);
		});
		expect(
			screen.queryByText(/создайте или выберите базу слева/i),
		).not.toBeInTheDocument();
	});
});
