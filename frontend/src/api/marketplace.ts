import { api } from "./client";

export type ListingCategory =
	| "equipment"
	| "books"
	| "services"
	| "furniture"
	| "other";

export type ListingStatus =
	| "pending"
	| "approved"
	| "rejected"
	| "sold"
	| "withdrawn";

export interface MarketplaceListing {
	id: string;
	title: string;
	description: string;
	category: ListingCategory;
	status: ListingStatus;
	price_cents?: number;
	currency: string;
	review_note?: string;
	author_id: string;
	author_name?: string;
	reviewer_id?: string;
	reviewer_name?: string;
	version: number;
	reviewed_at?: string;
	created_at: string;
	updated_at: string;
}

export const CATEGORY_LABELS: Record<ListingCategory, string> = {
	equipment: "Техника",
	books: "Книги",
	services: "Услуги",
	furniture: "Мебель",
	other: "Другое",
};

export const STATUS_LABELS: Record<ListingStatus, string> = {
	pending: "На модерации",
	approved: "Опубликовано",
	rejected: "Отклонено",
	sold: "Продано",
	withdrawn: "Снято",
};

export function listMarketplaceListings(
	category?: ListingCategory,
	page = 1,
	perPage = 25,
): Promise<{ items: MarketplaceListing[]; total: number }> {
	const params = new URLSearchParams({
		page: String(page),
		per_page: String(perPage),
	});
	if (category) params.set("category", category);
	return api.get(`/marketplace/listings?${params}`);
}

export function listMyListings(): Promise<{ items: MarketplaceListing[] }> {
	return api.get("/marketplace/listings/mine");
}

export function listPendingListings(): Promise<{
	items: MarketplaceListing[];
}> {
	return api.get("/marketplace/listings/pending");
}

export function createListing(data: {
	title: string;
	description: string;
	category: ListingCategory;
	price_cents?: number;
	currency?: string;
}): Promise<MarketplaceListing> {
	return api.post("/marketplace/listings", data);
}

export function approveListing(
	id: string,
	version: number,
): Promise<MarketplaceListing> {
	return api.post(`/marketplace/listings/${id}/approve`, { version });
}

export function rejectListing(
	id: string,
	version: number,
	review_note?: string,
): Promise<MarketplaceListing> {
	return api.post(`/marketplace/listings/${id}/reject`, {
		version,
		review_note,
	});
}

export function markListingSold(
	id: string,
	version: number,
): Promise<MarketplaceListing> {
	return api.post(`/marketplace/listings/${id}/sold`, { version });
}

export function withdrawListing(
	id: string,
	version: number,
): Promise<MarketplaceListing> {
	return api.post(`/marketplace/listings/${id}/withdraw`, { version });
}

export function formatListingPrice(
	priceCents?: number,
	currency = "RUB",
): string {
	if (priceCents === undefined || priceCents === null) return "Договорная";
	if (priceCents === 0) return "Бесплатно";
	return `${(priceCents / 100).toLocaleString("ru-RU")} ${currency}`;
}
