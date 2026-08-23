import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as accountingApi from "@/api/accounting";
import { AccountingPage } from "@/modules/accounting/accounting-page";

vi.mock("@/api/accounting");

/** A3W10-1-001: secondary accounting queries must show error≠empty. */
describe("AccountingPage secondary query error≠empty (A3W10-1-001)", () => {
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

	function renderPage() {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		return render(
			<QueryClientProvider client={qc}>
				<AccountingPage />
			</QueryClientProvider>,
		);
	}

	it("shows P&L alert when getIncomeStatement fails", async () => {
		vi.mocked(accountingApi.getIncomeStatement).mockRejectedValue(
			new Error("network"),
		);
		renderPage();

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить p&l/i,
			);
		});
		expect(screen.queryByText("Выручка")).not.toBeInTheDocument();
	});

	it("shows invoices alert when listInvoices fails", async () => {
		vi.mocked(accountingApi.listInvoices).mockRejectedValue(
			new Error("network"),
		);
		renderPage();

		await waitFor(() => {
			expect(
				screen.getByText(/не удалось загрузить счета/i),
			).toBeInTheDocument();
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});

	it("shows trial balance alert when getTrialBalance fails", async () => {
		vi.mocked(accountingApi.getTrialBalance).mockRejectedValue(
			new Error("network"),
		);
		renderPage();

		await waitFor(() => {
			expect(
				screen.getByText(/не удалось загрузить оборотно-сальдовую ведомость/i),
			).toBeInTheDocument();
		});
	});
});
