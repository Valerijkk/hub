import { api } from "@/api/client";

export interface AIProvider {
	id: string;
	name: string;
	provider_type: string;
	base_url: string;
	models: string[];
	enabled: boolean;
	priority: number;
	max_rpm: number;
	ru_resident: boolean;
	scope: string;
	has_api_key: boolean;
	created_at: string;
	updated_at: string;
}

export interface CreateAIProviderRequest {
	name: string;
	provider_type: string;
	base_url: string;
	api_key?: string;
	models: string[];
	enabled?: boolean;
	priority?: number;
	max_rpm?: number;
	ru_resident?: boolean;
}

export interface UpdateAIProviderRequest {
	name?: string;
	base_url?: string;
	api_key?: string;
	models?: string[];
	enabled?: boolean;
	priority?: number;
	max_rpm?: number;
	ru_resident?: boolean;
}

export interface TestAIProviderResponse {
	ok: boolean;
	latency_ms: number;
	model?: string;
	error?: string;
	provider: string;
}

export function listAIProviders() {
	return api.get<AIProvider[]>("/admin/ai/providers");
}

export function createAIProvider(data: CreateAIProviderRequest) {
	return api.post<AIProvider>("/admin/ai/providers", data);
}

export function updateAIProvider(id: string, data: UpdateAIProviderRequest) {
	return api.patch<AIProvider>(`/admin/ai/providers/${id}`, data);
}

export function testAIProvider(id: string) {
	return api.post<TestAIProviderResponse>(`/admin/ai/providers/${id}/test`, {});
}

export interface AIUsageSummary {
	usage_date: string;
	total_requests: number;
	tokens_in: number;
	tokens_out: number;
	active_users: number;
}

export function getAIUsageSummary() {
	return api.get<AIUsageSummary>("/admin/ai/usage/summary");
}

export interface AIUsageHistoryDay {
	date: string;
	requests: number;
	tokens_in: number;
	tokens_out: number;
	active_users: number;
}

export interface AIUsageHistory {
	days: AIUsageHistoryDay[];
}

export function getAIUsageHistory(days = 7) {
	return api.get<AIUsageHistory>(`/admin/ai/usage/history?days=${days}`);
}

export interface AIAdminSettings {
	system_prompt: string;
	offline_only: boolean;
}

export function getAIAdminSettings() {
	return api.get<AIAdminSettings>("/admin/ai/settings");
}

export function updateAIAdminSettings(data: { system_prompt?: string }) {
	return api.patch<AIAdminSettings>("/admin/ai/settings", data);
}

/** Пресеты провайдеров из AI-PROVIDER-HUB-SPEC.md — ускоряют настройку admin UI. */
export const AI_PROVIDER_PRESETS: Array<
	Pick<
		CreateAIProviderRequest,
		"name" | "provider_type" | "base_url" | "models"
	> & {
		ru_resident?: boolean;
	}
> = [
	{
		name: "DeepSeek",
		provider_type: "openai_compatible",
		base_url: "https://api.deepseek.com/v1",
		models: ["deepseek-chat", "deepseek-reasoner"],
	},
	{
		name: "OpenRouter",
		provider_type: "openai_compatible",
		base_url: "https://openrouter.ai/api/v1",
		models: [
			"openai/gpt-4o-mini",
			"anthropic/claude-3.5-sonnet",
			"deepseek/deepseek-chat",
		],
	},
	{
		name: "Ollama (local)",
		provider_type: "ollama",
		base_url: "http://127.0.0.1:11434/v1",
		models: ["llama3", "mistral"],
	},
	{
		name: "GigaChat (RU)",
		provider_type: "gigachat",
		base_url: "https://gigachat.devices.sberbank.ru/api/v1",
		models: ["GigaChat"],
		ru_resident: true,
	},
	{
		name: "Anthropic",
		provider_type: "anthropic",
		base_url: "https://api.anthropic.com",
		models: ["claude-3-5-haiku-20241022", "claude-3-5-sonnet-20241022"],
	},
	{
		name: "Mistral (EU)",
		provider_type: "openai_compatible",
		base_url: "https://api.mistral.ai/v1",
		models: ["mistral-small-latest", "codestral-latest"],
	},
	{
		name: "YandexGPT (RU)",
		provider_type: "openai_compatible",
		base_url: "https://llm.api.cloud.yandex.net",
		models: ["yandexgpt/latest", "yandexgpt-lite/latest"],
		ru_resident: true,
	},
];
