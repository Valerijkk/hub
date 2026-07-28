/**
 * Общие типы AI-провайдеров (user BYOK + admin).
 * Раздельные DTO наследуют базу, чтобы не дублировать `AIProvider` с разной формой.
 */

/** Общие поля провайдера (user + admin). */
export interface AIProviderBase {
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

/** Провайдер в админке (расширенный). */
export interface AdminAIProvider extends AIProviderBase {
	priority: number;
	max_rpm: number;
	created_at: string;
	updated_at: string;
}

/** Провайдер пользователя (BYOK). */
export type MyAIProvider = AIProviderBase;
