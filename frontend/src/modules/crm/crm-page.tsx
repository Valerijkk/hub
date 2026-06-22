import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";
import { createContact, createDeal, listContacts } from "@/api/crm";
import { listProjects } from "@/api/projects";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { DealsPipeline } from "@/modules/crm/deals-pipeline";

/** MOD-05: CRM-lite — контакты и kanban сделок. */
export function CrmPage() {
	const queryClient = useQueryClient();
	const [dealOpen, setDealOpen] = useState(false);
	const [contactName, setContactName] = useState("");
	const [contactEmail, setContactEmail] = useState("");
	const [dealTitle, setDealTitle] = useState("");
	const [dealAmount, setDealAmount] = useState("");
	const [dealContactId, setDealContactId] = useState("");
	const [dealProjectId, setDealProjectId] = useState("");

	const { data: contactsData } = useQuery({
		queryKey: ["crm", "contacts"],
		queryFn: listContacts,
	});

	const { data: projects } = useQuery({
		queryKey: ["projects"],
		queryFn: listProjects,
	});

	const createContactMutation = useMutation({
		mutationFn: createContact,
		onSuccess: () => {
			setContactName("");
			setContactEmail("");
			void queryClient.invalidateQueries({ queryKey: ["crm", "contacts"] });
			toast.success("Контакт добавлен");
		},
		onError: () => toast.error("Не удалось создать контакт"),
	});

	const createDealMutation = useMutation({
		mutationFn: createDeal,
		onSuccess: () => {
			setDealOpen(false);
			setDealTitle("");
			setDealAmount("");
			setDealContactId("");
			setDealProjectId("");
			void queryClient.invalidateQueries({ queryKey: ["crm", "pipeline"] });
			toast.success("Сделка создана");
		},
		onError: () => toast.error("Не удалось создать сделку"),
	});

	const handleAddContact = (e: React.FormEvent) => {
		e.preventDefault();
		if (!contactName.trim()) return;
		createContactMutation.mutate({
			name: contactName.trim(),
			email: contactEmail.trim() || undefined,
		});
	};

	const handleCreateDeal = (e: React.FormEvent) => {
		e.preventDefault();
		if (!dealTitle.trim()) return;
		const amount = dealAmount.trim()
			? Math.round(Number.parseFloat(dealAmount) * 100)
			: undefined;
		createDealMutation.mutate({
			title: dealTitle.trim(),
			amount_cents: Number.isFinite(amount) ? amount : undefined,
			contact_id: dealContactId || undefined,
			project_id: dealProjectId || undefined,
			stage: "lead",
		});
	};

	const contacts = contactsData?.contacts ?? [];

	return (
		<div className="mx-auto max-w-[1400px] space-y-6 px-4 py-6">
			<header className="space-y-1">
				<h1 className="text-2xl font-bold tracking-tight">CRM</h1>
				<p className="text-sm text-t2">
					Контакты и сделки · связь с проектами Hub
				</p>
			</header>

			<div className="grid gap-6 lg:grid-cols-[280px_1fr]">
				<aside className="space-y-4 rounded-xl border border-b1 p-4">
					<h2 className="text-sm font-semibold">Контакты</h2>
					<form onSubmit={handleAddContact} className="space-y-2">
						<label htmlFor="crm-contact-name" className="sr-only">
							Имя контакта
						</label>
						<input
							id="crm-contact-name"
							type="text"
							value={contactName}
							onChange={(e) => setContactName(e.target.value)}
							placeholder="Имя / компания"
							className="w-full rounded-lg border border-b1 bg-[var(--hover)]/20 px-3 py-2 text-sm"
						/>
						<label htmlFor="crm-contact-email" className="sr-only">
							Email
						</label>
						<input
							id="crm-contact-email"
							type="email"
							value={contactEmail}
							onChange={(e) => setContactEmail(e.target.value)}
							placeholder="email@company.com"
							className="w-full rounded-lg border border-b1 bg-[var(--hover)]/20 px-3 py-2 text-sm"
						/>
						<Button
							type="submit"
							size="sm"
							className="w-full"
							disabled={createContactMutation.isPending}
						>
							Добавить
						</Button>
					</form>
					<ul className="max-h-64 space-y-1 overflow-y-auto text-sm">
						{contacts.map((c) => (
							<li
								key={c.id}
								className="rounded-md px-2 py-1.5 hover:bg-[var(--hover)]/40"
							>
								<p className="font-medium">{c.name}</p>
								{c.email && <p className="text-xs text-t3">{c.email}</p>}
							</li>
						))}
						{contacts.length === 0 && (
							<li className="text-xs text-t3">Нет контактов</li>
						)}
					</ul>
				</aside>

				<DealsPipeline onCreateDeal={() => setDealOpen(true)} />
			</div>

			<Dialog open={dealOpen} onOpenChange={setDealOpen}>
				<DialogContent className="sm:max-w-md">
					<DialogHeader>
						<DialogTitle>Новая сделка</DialogTitle>
					</DialogHeader>
					<form onSubmit={handleCreateDeal} className="space-y-3">
						<label htmlFor="deal-title" className="sr-only">
							Название
						</label>
						<input
							id="deal-title"
							type="text"
							value={dealTitle}
							onChange={(e) => setDealTitle(e.target.value)}
							placeholder="Название сделки"
							className="w-full rounded-lg border border-b1 px-3 py-2 text-sm"
						/>
						<label htmlFor="deal-amount" className="sr-only">
							Сумма
						</label>
						<input
							id="deal-amount"
							type="number"
							min="0"
							step="0.01"
							value={dealAmount}
							onChange={(e) => setDealAmount(e.target.value)}
							placeholder="Сумма (RUB)"
							className="w-full rounded-lg border border-b1 px-3 py-2 text-sm"
						/>
						<label htmlFor="deal-contact" className="sr-only">
							Контакт
						</label>
						<select
							id="deal-contact"
							value={dealContactId}
							onChange={(e) => setDealContactId(e.target.value)}
							className="w-full rounded-lg border border-b1 px-3 py-2 text-sm"
						>
							<option value="">Без контакта</option>
							{contacts.map((c) => (
								<option key={c.id} value={c.id}>
									{c.name}
								</option>
							))}
						</select>
						<label htmlFor="deal-project" className="sr-only">
							Проект
						</label>
						<select
							id="deal-project"
							value={dealProjectId}
							onChange={(e) => setDealProjectId(e.target.value)}
							className="w-full rounded-lg border border-b1 px-3 py-2 text-sm"
						>
							<option value="">Без проекта</option>
							{(projects ?? []).map((p) => (
								<option key={p.id} value={p.id}>
									{p.key} — {p.name}
								</option>
							))}
						</select>
						<div className="flex justify-end gap-2 pt-2">
							<Button
								type="button"
								variant="outline"
								onClick={() => setDealOpen(false)}
							>
								Отмена
							</Button>
							<Button type="submit" disabled={createDealMutation.isPending}>
								Создать
							</Button>
						</div>
					</form>
				</DialogContent>
			</Dialog>
		</div>
	);
}
