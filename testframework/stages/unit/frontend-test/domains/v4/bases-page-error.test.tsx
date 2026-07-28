import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
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
});
