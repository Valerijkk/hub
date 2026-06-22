import { api } from "@/api/client";

export type AccountType =
	| "asset"
	| "liability"
	| "equity"
	| "revenue"
	| "expense";

export interface LedgerAccount {
	id: string;
	code: string;
	name: string;
	account_type: AccountType;
	is_active: boolean;
	version: number;
}

export interface JournalLine {
	id: string;
	account_id: string;
	account_code?: string;
	account_name?: string;
	debit_cents: number;
	credit_cents: number;
	memo?: string;
}

export interface JournalEntry {
	id: string;
	entry_date: string;
	description: string;
	reference?: string;
	status: "draft" | "posted";
	version: number;
	posted_at?: string;
	created_by: string;
	lines: JournalLine[];
	created_at: string;
}

export interface TrialBalanceRow {
	account_id: string;
	code: string;
	name: string;
	account_type: AccountType;
	debit_cents: number;
	credit_cents: number;
}

export function listAccounts() {
	return api
		.get<{ accounts: LedgerAccount[] }>("/accounting/accounts")
		.then((r) => r.accounts);
}

export function listJournal() {
	return api
		.get<{ entries: JournalEntry[] }>("/accounting/journal")
		.then((r) => r.entries);
}

export function getTrialBalance() {
	return api
		.get<{ rows: TrialBalanceRow[] }>("/accounting/trial-balance")
		.then((r) => r.rows);
}

export function createJournal(data: {
	entry_date: string;
	description: string;
	reference?: string;
	lines: Array<{
		account_id: string;
		debit_cents: number;
		credit_cents: number;
	}>;
}) {
	return api.post<JournalEntry>("/accounting/journal", data);
}

export function postJournal(entryId: string) {
	return api.post<JournalEntry>(`/accounting/journal/${entryId}/post`, {});
}

export type InvoiceStatus = "draft" | "issued" | "paid" | "cancelled";

export interface Invoice {
	id: string;
	number: string;
	customer_name: string;
	description?: string;
	amount_cents: number;
	currency: string;
	status: InvoiceStatus;
	due_date?: string;
	issued_at?: string;
	paid_at?: string;
	crm_deal_id?: string;
	journal_entry_id?: string;
	version: number;
	created_at: string;
}

export interface IncomeStatement {
	revenue_cents: number;
	expense_cents: number;
	net_cents: number;
}

export function listInvoices() {
	return api
		.get<{ invoices: Invoice[] }>("/accounting/invoices")
		.then((r) => r.invoices);
}

export function createInvoice(data: {
	customer_name: string;
	description?: string;
	amount_cents: number;
	currency?: string;
	due_date?: string;
}) {
	return api.post<Invoice>("/accounting/invoices", data);
}

export function issueInvoice(invoiceId: string) {
	return api.post<Invoice>(`/accounting/invoices/${invoiceId}/issue`, {});
}

export function payInvoice(invoiceId: string) {
	return api.post<Invoice>(`/accounting/invoices/${invoiceId}/pay`, {});
}

export function getIncomeStatement() {
	return api.get<IncomeStatement>("/accounting/income-statement");
}

export function formatRub(cents: number): string {
	return (cents / 100).toLocaleString("ru-RU", {
		style: "currency",
		currency: "RUB",
		maximumFractionDigits: 2,
	});
}
