import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as accountingApi from "@/api/accounting";
import { AccountingPage } from "@/modules/accounting/accounting-page";

vi.mock("@/api/accounting");

describe("AccountingPage", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(accountingApi.listAccounts).mockResolvedValue([
			{
				id: "a1",
				code: "1010",
				name: "Расчётный счёт",
				account_type: "asset",
				is_active: true,
				version: 1,
			},
			{
				id: "a2",
				code: "5010",
				name: "Операционные расходы",
				account_type: "expense",
				is_active: true,
				version: 1,
			},
		]);
		vi.mocked(accountingApi.listJournal).mockResolvedValue([]);
		vi.mocked(accountingApi.getTrialBalance).mockResolvedValue([]);
		vi.mocked(accountingApi.listInvoices).mockResolvedValue([]);
		vi.mocked(accountingApi.getIncomeStatement).mockResolvedValue({
			revenue_cents: 0,
			expense_cents: 0,
			net_cents: 0,
		});
	});

	it("renders accounting heading and chart", async () => {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<AccountingPage />
			</QueryClientProvider>,
		);

		expect(
			screen.getByRole("heading", { name: /Бухгалтерия/i }),
		).toBeInTheDocument();
		await waitFor(() => {
			expect(accountingApi.listAccounts).toHaveBeenCalled();
		});
		await waitFor(() => {
			expect(
				screen.getByRole("heading", { name: /План счетов/i }),
			).toBeInTheDocument();
		});
	});
});
