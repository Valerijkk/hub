import { api } from "@/api/client";

export interface AIProvider {
	id: string;
	name: string;
	provider_type: string;
	base_url: string;
	models: string[];
	enabled: boolean;
	ru_resident: boolean;
	scope: string;
	has_api_key: boolean;
}

export interface CreateMyAIProviderRequest {
	name: string;
	provider_type: string;
	base_url: string;
	api_key?: string;
	models: string[];
	ru_resident?: boolean;
}

export function listMyAIProviders() {
	return api.get<AIProvider[]>("/ai/my/providers");
}

export function createMyAIProvider(data: CreateMyAIProviderRequest) {
	return api.post<AIProvider>("/ai/my/providers", data);
}

export function updateMyAIProvider(
	id: string,
	data: Partial<CreateMyAIProviderRequest & { enabled?: boolean }>,
) {
	return api.patch<AIProvider>(`/ai/my/providers/${id}`, data);
}
