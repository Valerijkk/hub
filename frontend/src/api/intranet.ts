import { api } from "./client";

export interface NewsPost {
	id: string;
	title: string;
	body: string;
	pinned: boolean;
	published: boolean;
	version: number;
	author_id: string;
	author_name?: string;
	created_at: string;
	updated_at: string;
}

export interface NewsListResponse {
	items: NewsPost[];
	total: number;
	page: number;
	per_page: number;
}

export interface CreateNewsRequest {
	title: string;
	body: string;
	pinned?: boolean;
	published?: boolean;
}

export function listNews(page = 1, perPage = 25): Promise<NewsListResponse> {
	return api.get<NewsListResponse>(
		`/intranet/news?page=${page}&per_page=${perPage}`,
	);
}

export function createNews(data: CreateNewsRequest): Promise<NewsPost> {
	return api.post<NewsPost>("/intranet/news", data);
}

export interface UpdateNewsRequest {
	title?: string;
	body?: string;
	pinned?: boolean;
	published?: boolean;
	version: number;
}

export function updateNews(
	id: string,
	data: UpdateNewsRequest,
): Promise<NewsPost> {
	return api.put<NewsPost>(`/intranet/news/${id}`, data);
}

export function deleteNews(id: string): Promise<void> {
	return api.delete(`/intranet/news/${id}`);
}

export interface DirectoryUser {
	id: string;
	email: string;
	name?: string;
	display_name: string;
	department?: string;
	position?: string;
	avatar_url?: string;
	timezone?: string;
	birthday?: string;
}

export interface DirectoryDepartmentGroup {
	department: string;
	users: DirectoryUser[];
}

export interface DirectoryResponse {
	departments: DirectoryDepartmentGroup[];
	total: number;
}

export function listDirectory(): Promise<DirectoryResponse> {
	return api.get<DirectoryResponse>("/intranet/directory");
}

export function getDirectoryUser(userId: string): Promise<DirectoryUser> {
	return api.get<DirectoryUser>(`/intranet/directory/users/${userId}`);
}

export interface CoffeeEnrollment {
	enrolled: boolean;
	is_active: boolean;
	enrolled_at?: string;
}

export interface CoffeePartner {
	user_id: string;
	display_name?: string;
	email?: string;
	department?: string;
}

export interface CoffeeMyMatch {
	round_id?: string;
	round_at?: string;
	partner?: CoffeePartner;
	unpaired: boolean;
}

export interface CoffeeRound {
	id: string;
	pair_count: number;
	unpaired_count: number;
	pairs: Array<{
		id: string;
		participant_a: CoffeePartner;
		participant_b: CoffeePartner;
	}>;
	created_at: string;
	triggered_by?: string;
}

export function getCoffeeEnrollment(): Promise<CoffeeEnrollment> {
	return api.get<CoffeeEnrollment>("/intranet/coffee/enrollment");
}

export function setCoffeeEnrollment(
	active: boolean,
): Promise<CoffeeEnrollment> {
	return api.put<CoffeeEnrollment>("/intranet/coffee/enrollment", { active });
}

export function getMyCoffeeMatch(): Promise<CoffeeMyMatch> {
	return api.get<CoffeeMyMatch>("/intranet/coffee/match");
}

export function runCoffeeMatching(): Promise<CoffeeRound> {
	return api.post<CoffeeRound>("/intranet/coffee/run", {});
}

export function listCoffeeRounds(): Promise<{ items: CoffeeRound[] }> {
	return api.get<{ items: CoffeeRound[] }>("/intranet/coffee/rounds");
}

export interface Kudo {
	id: string;
	message: string;
	is_public: boolean;
	sender_id: string;
	sender_name?: string;
	recipient_id: string;
	recipient_name?: string;
	created_at: string;
}

export interface KudoListResponse {
	items: Kudo[];
	total: number;
	page: number;
	per_page: number;
}

export function listKudosFeed(
	page = 1,
	perPage = 25,
): Promise<KudoListResponse> {
	return api.get<KudoListResponse>(
		`/intranet/kudos/feed?page=${page}&per_page=${perPage}`,
	);
}

export function listReceivedKudos(
	page = 1,
	perPage = 25,
): Promise<KudoListResponse> {
	return api.get<KudoListResponse>(
		`/intranet/kudos/received?page=${page}&per_page=${perPage}`,
	);
}

export function createKudo(data: {
	recipient_id: string;
	message: string;
	is_public?: boolean;
}): Promise<Kudo> {
	return api.post<Kudo>("/intranet/kudos", data);
}
