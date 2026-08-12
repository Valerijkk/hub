import type { MyAIProvider } from "@/api/ai-types";
import { api } from "@/api/client";

/** @deprecated Используйте MyAIProvider — alias для совместимости импортов. */
export type AIProvider = MyAIProvider;

export type { MyAIProvider };

export interface CreateMyAIProviderRequest {
	name: string;
	provider_type: string;
	base_url: string;
	api_key?: string;
	models: string[];
	ru_resident?: boolean;
}

export function listMyAIProviders() {
	return api.get<MyAIProvider[]>("/ai/my/providers");
}

export function createMyAIProvider(data: CreateMyAIProviderRequest) {
	return api.post<MyAIProvider>("/ai/my/providers", data);
}

export function updateMyAIProvider(
	id: string,
	data: Partial<CreateMyAIProviderRequest & { enabled?: boolean }>,
) {
	return api.patch<MyAIProvider>(`/ai/my/providers/${id}`, data);
}
