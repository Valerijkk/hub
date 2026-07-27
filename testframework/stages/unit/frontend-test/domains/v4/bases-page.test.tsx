import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as basesApi from "@/api/bases";
import { BasesPage } from "@/modules/bases/bases-page";

vi.mock("@/api/bases");

const baseSummary = {
	id: "b1",
	name: "Inventory",
	version: 1,
	owner_id: "u1",
	table_count: 1,
	created_at: "2026-06-25T00:00:00Z",
	updated_at: "2026-06-25T00:00:00Z",
};

describe("BasesPage update/delete (A11-010)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(basesApi.listBases).mockResolvedValue([baseSummary]);
		vi.mocked(basesApi.getBase).mockResolvedValue({
			...baseSummary,
			tables: [
				{
					id: "t1",
					name: "Sheet",
					sort_order: 0,
					version: 1,
				},
			],
		});
		vi.mocked(basesApi.updateBase).mockResolvedValue({
			...baseSummary,
			name: "Renamed",
			version: 2,
			tables: [],
		});
		vi.mocked(basesApi.deleteBase).mockResolvedValue(undefined);
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

	it("calls updateBase when renaming selected base", async () => {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<BasesPage />
			</QueryClientProvider>,
		);

		fireEvent.click(await screen.findByRole("button", { name: "Inventory" }));
		const rename = await screen.findByLabelText("Название базы");
		fireEvent.change(rename, { target: { value: "Renamed" } });
		fireEvent.click(screen.getByRole("button", { name: /^Сохранить$/i }));

		await waitFor(() => {
			expect(basesApi.updateBase).toHaveBeenCalledWith("b1", {
				name: "Renamed",
				version: 1,
			});
		});
	});
});
