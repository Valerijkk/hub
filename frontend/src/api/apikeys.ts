import { api } from "./client";

/** Метаданные API-ключа. */
export interface APIKeyInfo {
	id: string;
	name: string;
	prefix: string;
	last_used_at?: string;
	expires_at?: string;
	created_at: string;
}

/** Ответ при создании ключа (содержит полный ключ). */
export interface CreateAPIKeyResponse extends APIKeyInfo {
	key: string;
}

/** Запрос на создание API-ключа. */
export interface CreateAPIKeyRequest {
	name: string;
	expires_at?: string;
}

/** Получить список API-ключей пользователя. */
export function listAPIKeys() {
	return api.get<APIKeyInfo[]>("/user/api-keys");
}

/** Создать новый API-ключ. */
export function createAPIKey(req: CreateAPIKeyRequest) {
	return api.post<CreateAPIKeyResponse>("/user/api-keys", req);
}

/** Отозвать (удалить) API-ключ. */
export function revokeAPIKey(id: string) {
	return api.delete<void>(`/user/api-keys/${id}`);
}
