import { api } from "./client";

export type DealStage =
	| "lead"
	| "qualified"
	| "proposal"
	| "negotiation"
	| "won"
	| "lost";

export interface CrmContact {
	id: string;
	name: string;
	email?: string;
	phone?: string;
	company?: string;
	notes?: string;
	linked_user_id?: string;
	version: number;
	created_at: string;
	updated_at: string;
}

export interface CrmDeal {
	id: string;
	title: string;
	description?: string;
	stage: DealStage;
	amount_cents?: number;
	currency: string;
	rank?: string;
	contact_id?: string;
	contact_name?: string;
	project_id?: string;
	project_key?: string;
	owner_id: string;
	owner_name?: string;
	version: number;
	created_at: string;
	updated_at: string;
}

export interface PipelineStage {
	stage: DealStage;
	deals: CrmDeal[];
}

export interface PipelineResponse {
	stages: PipelineStage[];
}

export function listContacts(): Promise<{ contacts: CrmContact[] }> {
	return api.get<{ contacts: CrmContact[] }>("/crm/contacts");
}

export function createContact(data: {
	name: string;
	email?: string;
	phone?: string;
	company?: string;
}): Promise<CrmContact> {
	return api.post<CrmContact>("/crm/contacts", data);
}

export function updateContact(
	id: string,
	data: {
		name?: string;
		email?: string;
		phone?: string;
		company?: string;
		notes?: string;
		version: number;
	},
): Promise<CrmContact> {
	return api.put<CrmContact>(`/crm/contacts/${id}`, data);
}

export function deleteContact(id: string): Promise<void> {
	return api.delete(`/crm/contacts/${id}`);
}

export function getPipeline(): Promise<PipelineResponse> {
	return api.get<PipelineResponse>("/crm/pipeline");
}

export function createDeal(data: {
	title: string;
	description?: string;
	stage?: DealStage;
	amount_cents?: number;
	contact_id?: string;
	project_id?: string;
}): Promise<CrmDeal> {
	return api.post<CrmDeal>("/crm/deals", data);
}

export function updateDeal(
	id: string,
	data: {
		title?: string;
		stage?: DealStage;
		amount_cents?: number;
		contact_id?: string;
		project_id?: string;
		version: number;
	},
): Promise<CrmDeal> {
	return api.put<CrmDeal>(`/crm/deals/${id}`, data);
}

export function deleteDeal(id: string): Promise<void> {
	return api.delete(`/crm/deals/${id}`);
}

export const STAGE_LABELS: Record<DealStage, string> = {
	lead: "Лид",
	qualified: "Квалификация",
	proposal: "Предложение",
	negotiation: "Переговоры",
	won: "Выиграна",
	lost: "Проиграна",
};

export function formatAmount(cents?: number, currency = "RUB"): string {
	if (cents == null) return "";
	const value = cents / 100;
	return new Intl.NumberFormat("ru-RU", {
		style: "currency",
		currency,
		maximumFractionDigits: 0,
	}).format(value);
}
