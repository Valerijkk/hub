import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import {
	createInvoice,
	createJournal,
	formatRub,
	getIncomeStatement,
	getTrialBalance,
	issueInvoice,
	type LedgerAccount,
	listAccounts,
	listInvoices,
	listJournal,
	payInvoice,
	postJournal,
} from "@/api/accounting";
import { QueryErrorAlert } from "@/components/query-state-view";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

/** V4-03: встроенный учёт Hub — план счетов + журнал (admin). */
export function AccountingPage() {
	const queryClient = useQueryClient();
	const [description, setDescription] = useState("");
	const [amount, setAmount] = useState("");
	const [debitAccountId, setDebitAccountId] = useState("");
	const [creditAccountId, setCreditAccountId] = useState("");
	const [customerName, setCustomerName] = useState("");
	const [invoiceAmount, setInvoiceAmount] = useState("");
	const [invoiceDesc, setInvoiceDesc] = useState("");

	const {
		data: accounts = [],
		isLoading: accountsLoading,
		isError: accountsError,
		refetch: refetchAccounts,
	} = useQuery({
		queryKey: ["accounting", "accounts"],
		queryFn: listAccounts,
	});

	const { data: entries = [], isError: journalError } = useQuery({
		queryKey: ["accounting", "journal"],
		queryFn: listJournal,
	});

	const { data: trialBalance = [] } = useQuery({
		queryKey: ["accounting", "trial-balance"],
		queryFn: getTrialBalance,
	});

	const { data: invoices = [] } = useQuery({
		queryKey: ["accounting", "invoices"],
		queryFn: listInvoices,
	});

	const { data: incomeStatement } = useQuery({
		queryKey: ["accounting", "income-statement"],
		queryFn: getIncomeStatement,
	});

	const accountsLoadError = accountsError || journalError;

	const expenseAccount = useMemo(
		() => accounts.find((a) => a.code === "5010"),
		[accounts],
	);
	const bankAccount = useMemo(
		() => accounts.find((a) => a.code === "1010"),
		[accounts],
	);

	const createMutation = useMutation({
		mutationFn: createJournal,
		onSuccess: () => {
			setDescription("");
			setAmount("");
			void queryClient.invalidateQueries({ queryKey: ["accounting"] });
			toast.success("Проводка создана (черновик)");
		},
		onError: () => toast.error("Не удалось создать проводку"),
	});

	const postMutation = useMutation({
		mutationFn: postJournal,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["accounting"] });
			toast.success("Проводка проведена");
		},
		onError: () => toast.error("Не удалось провести"),
	});

	const invoiceMutation = useMutation({
		mutationFn: createInvoice,
		onSuccess: () => {
			setCustomerName("");
			setInvoiceAmount("");
			setInvoiceDesc("");
			void queryClient.invalidateQueries({ queryKey: ["accounting"] });
			toast.success("Счёт создан (черновик)");
		},
		onError: () => toast.error("Не удалось создать счёт"),
	});

	const issueMutation = useMutation({
		mutationFn: issueInvoice,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["accounting"] });
			toast.success("Счёт выставлен");
		},
		onError: () => toast.error("Не удалось выставить счёт"),
	});

	const payMutation = useMutation({
		mutationFn: payInvoice,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["accounting"] });
			toast.success("Оплата проведена в журнал");
		},
		onError: () => toast.error("Не удалось провести оплату"),
	});

	const handleCreate = (e: React.FormEvent) => {
		e.preventDefault();
		const debitId = debitAccountId || expenseAccount?.id;
		const creditId = creditAccountId || bankAccount?.id;
		const rub = Number.parseFloat(amount.replace(",", "."));
		if (!description.trim() || !debitId || !creditId || !Number.isFinite(rub)) {
			return;
		}
		const cents = Math.round(rub * 100);
		createMutation.mutate({
			entry_date: new Date().toISOString().slice(0, 10),
			description: description.trim(),
			lines: [
				{ account_id: debitId, debit_cents: cents, credit_cents: 0 },
				{ account_id: creditId, debit_cents: 0, credit_cents: cents },
			],
		});
	};

	const handleCreateInvoice = (e: React.FormEvent) => {
		e.preventDefault();
		const rub = Number.parseFloat(invoiceAmount.replace(",", "."));
		if (!customerName.trim() || !Number.isFinite(rub)) {
			return;
		}
		invoiceMutation.mutate({
			customer_name: customerName.trim(),
			description: invoiceDesc.trim() || undefined,
			amount_cents: Math.round(rub * 100),
		});
	};

	if (accountsLoading) {
		return (
			<div className="mx-auto max-w-[1200px] space-y-6 px-4 py-6">
				<header className="space-y-1">
					<h1 className="text-2xl font-bold tracking-tight">Бухгалтерия</h1>
					<p className="text-sm text-t2">
						Двойная запись · счета · P&L · Hub-only (V4-03)
					</p>
				</header>
				<p className="text-sm text-t2">Загрузка...</p>
			</div>
		);
	}

	if (accountsLoadError) {
		return (
			<div className="mx-auto max-w-[1200px] space-y-6 px-4 py-6">
				<header className="space-y-1">
					<h1 className="text-2xl font-bold tracking-tight">Бухгалтерия</h1>
					<p className="text-sm text-t2">
						Двойная запись · счета · P&L · Hub-only (V4-03)
					</p>
				</header>
				<QueryErrorAlert
					message="Не удалось загрузить данные учёта"
					retry={() => void refetchAccounts()}
				/>
			</div>
		);
	}

	return (
		<div className="mx-auto max-w-[1200px] space-y-6 px-4 py-6">
			<header className="space-y-1">
				<h1 className="text-2xl font-bold tracking-tight">Бухгалтерия</h1>
				<p className="text-sm text-t2">
					Двойная запись · счета · P&L · Hub-only (V4-03)
				</p>
			</header>

			{incomeStatement && (
				<div className="grid gap-3 sm:grid-cols-3">
					<div className="rounded-xl border border-b1 p-4">
						<p className="text-xs text-t3">Выручка</p>
						<p className="text-lg font-semibold">
							{formatRub(incomeStatement.revenue_cents)}
						</p>
					</div>
					<div className="rounded-xl border border-b1 p-4">
						<p className="text-xs text-t3">Расходы</p>
						<p className="text-lg font-semibold">
							{formatRub(incomeStatement.expense_cents)}
						</p>
					</div>
					<div className="rounded-xl border border-b1 p-4">
						<p className="text-xs text-t3">Чистая прибыль</p>
						<p className="text-lg font-semibold">
							{formatRub(incomeStatement.net_cents)}
						</p>
					</div>
				</div>
			)}

			<div className="grid gap-6 lg:grid-cols-2">
				<section className="rounded-xl border border-b1 p-4 space-y-3">
					<h2 className="text-sm font-semibold">План счетов</h2>
					<ul className="space-y-1 text-[13px]">
						{accounts.map((a: LedgerAccount) => (
							<li key={a.id} className="flex justify-between gap-2">
								<span>
									<span className="font-mono text-t3">{a.code}</span> {a.name}
								</span>
								<span className="text-t3">{a.account_type}</span>
							</li>
						))}
					</ul>
				</section>

				<section className="rounded-xl border border-b1 p-4 space-y-3">
					<h2 className="text-sm font-semibold">Новая проводка</h2>
					<form onSubmit={handleCreate} className="space-y-2">
						<Input
							value={description}
							onChange={(e) => setDescription(e.target.value)}
							placeholder="Описание"
						/>
						<Input
							value={amount}
							onChange={(e) => setAmount(e.target.value)}
							placeholder="Сумма, ₽"
							inputMode="decimal"
						/>
						<select
							value={debitAccountId || expenseAccount?.id || ""}
							onChange={(e) => setDebitAccountId(e.target.value)}
							className="w-full h-9 rounded-md border border-b1 bg-card px-2 text-[13px]"
						>
							{accounts.map((a) => (
								<option key={a.id} value={a.id}>
									Дт: {a.code} {a.name}
								</option>
							))}
						</select>
						<select
							value={creditAccountId || bankAccount?.id || ""}
							onChange={(e) => setCreditAccountId(e.target.value)}
							className="w-full h-9 rounded-md border border-b1 bg-card px-2 text-[13px]"
						>
							{accounts.map((a) => (
								<option key={a.id} value={a.id}>
									Кт: {a.code} {a.name}
								</option>
							))}
						</select>
						<Button type="submit" disabled={createMutation.isPending}>
							Создать черновик
						</Button>
					</form>
				</section>
			</div>

			<section className="rounded-xl border border-b1 p-4 space-y-3">
				<h2 className="text-sm font-semibold">Счета на оплату</h2>
				<form
					onSubmit={handleCreateInvoice}
					className="grid gap-2 sm:grid-cols-4 sm:items-end"
				>
					<Input
						value={customerName}
						onChange={(e) => setCustomerName(e.target.value)}
						placeholder="Клиент"
					/>
					<Input
						value={invoiceDesc}
						onChange={(e) => setInvoiceDesc(e.target.value)}
						placeholder="Описание"
					/>
					<Input
						value={invoiceAmount}
						onChange={(e) => setInvoiceAmount(e.target.value)}
						placeholder="Сумма, ₽"
						inputMode="decimal"
					/>
					<Button type="submit" disabled={invoiceMutation.isPending}>
						Новый счёт
					</Button>
				</form>
				<div className="overflow-x-auto">
					<table className="w-full text-[13px]">
						<thead>
							<tr className="border-b border-b1 text-left text-t2">
								<th className="py-2 pr-3">№</th>
								<th className="py-2 pr-3">Клиент</th>
								<th className="py-2 pr-3">Сумма</th>
								<th className="py-2 pr-3">Статус</th>
								<th className="py-2" />
							</tr>
						</thead>
						<tbody>
							{invoices.map((inv) => (
								<tr key={inv.id} className="border-b border-b1/60">
									<td className="py-2 pr-3 font-mono">{inv.number}</td>
									<td className="py-2 pr-3">{inv.customer_name}</td>
									<td className="py-2 pr-3">{formatRub(inv.amount_cents)}</td>
									<td className="py-2 pr-3">{inv.status}</td>
									<td className="py-2 text-right space-x-2">
										{inv.status === "draft" && (
											<Button
												type="button"
												size="sm"
												variant="outline"
												onClick={() => issueMutation.mutate(inv.id)}
												disabled={issueMutation.isPending}
											>
												Выставить
											</Button>
										)}
										{inv.status === "issued" && (
											<Button
												type="button"
												size="sm"
												onClick={() => payMutation.mutate(inv.id)}
												disabled={payMutation.isPending}
											>
												Оплачен
											</Button>
										)}
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</section>

			<section className="rounded-xl border border-b1 p-4 space-y-3">
				<h2 className="text-sm font-semibold">Журнал</h2>
				<div className="overflow-x-auto">
					<table className="w-full text-[13px]">
						<thead>
							<tr className="border-b border-b1 text-left text-t2">
								<th className="py-2 pr-3">Дата</th>
								<th className="py-2 pr-3">Описание</th>
								<th className="py-2 pr-3">Статус</th>
								<th className="py-2" />
							</tr>
						</thead>
						<tbody>
							{entries.map((e) => (
								<tr key={e.id} className="border-b border-b1/60">
									<td className="py-2 pr-3">{e.entry_date}</td>
									<td className="py-2 pr-3">{e.description}</td>
									<td className="py-2 pr-3">{e.status}</td>
									<td className="py-2 text-right">
										{e.status === "draft" && (
											<Button
												type="button"
												size="sm"
												variant="outline"
												onClick={() => postMutation.mutate(e.id)}
												disabled={postMutation.isPending}
											>
												Провести
											</Button>
										)}
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</section>

			<section className="rounded-xl border border-b1 p-4 space-y-3">
				<h2 className="text-sm font-semibold">
					Оборотно-сальдовая ведомость (posted)
				</h2>
				<div className="overflow-x-auto">
					<table className="w-full text-[13px]">
						<thead>
							<tr className="border-b border-b1 text-left text-t2">
								<th className="py-2 pr-3">Счёт</th>
								<th className="py-2 pr-3 text-right">Дебет</th>
								<th className="py-2 text-right">Кредит</th>
							</tr>
						</thead>
						<tbody>
							{trialBalance.map((r) => (
								<tr key={r.account_id} className="border-b border-b1/60">
									<td className="py-2 pr-3">
										<span className="font-mono text-t3">{r.code}</span> {r.name}
									</td>
									<td className="py-2 pr-3 text-right">
										{r.debit_cents ? formatRub(r.debit_cents) : "—"}
									</td>
									<td className="py-2 text-right">
										{r.credit_cents ? formatRub(r.credit_cents) : "—"}
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</section>
		</div>
	);
}
