import { useAuthStore } from "@/stores/auth-store";

/** Базовый путь API. Все запросы идут через /api префикс. */
const API_BASE = "/api";

/**
 * Структура ошибки API.
 * Возвращается сервером при неуспешных запросах (4xx, 5xx).
 */
export interface ApiError {
	/** HTTP-код ошибки. */
	code: number;
	/** Человекочитаемое описание ошибки. */
	message: string;
	/** ID запроса для отладки (опционально). */
	request_id?: string;
}

/**
 * Single-flight mutex для refresh-токена (FE-H1).
 *
 * При параллельных 401-ответах несколько запросов могут одновременно
 * попытаться обновить токен. Это приводит к гонке: первый успешный refresh
 * инвалидирует старый refresh_token, и последующие refresh-запросы получают
 * 401, что выкидывает пользователя из системы.
 *
 * Решение: модульный promise-lock. Если refresh уже в работе — все
 * последующие вызовы дожидаются того же промиса. По завершении промис
 * сбрасывается, и при следующем 401 запускается новый refresh.
 */
let refreshing: Promise<boolean> | null = null;

/**
 * HTTP-клиент для взаимодействия с backend API.
 *
 * Автоматически добавляет JWT-токен из auth-store в заголовок Authorization.
 * При получении 401 пытается обновить токен через refresh endpoint.
 * Если refresh не удался — выполняет logout.
 */
class ApiClient {
	/** Формирует заголовки запроса, включая Content-Type и Authorization. */
	private getHeaders(): Record<string, string> {
		const headers: Record<string, string> = {
			"Content-Type": "application/json",
		};
		const token = useAuthStore.getState().accessToken;
		if (token) {
			headers["Authorization"] = `Bearer ${token}`;
		}
		return headers;
	}

	/**
	 * Выполняет HTTP-запрос к API.
	 * При 401 пытается обновить токены и повторить запрос.
	 * При 204 или пустом теле возвращает undefined.
	 * @param method - HTTP-метод (GET, POST, PUT, PATCH, DELETE)
	 * @param path - путь относительно API_BASE (например, "/auth/me")
	 * @param body - тело запроса (будет сериализовано в JSON)
	 */
	async request<T>(method: string, path: string, body?: unknown): Promise<T> {
		const response = await fetch(`${API_BASE}${path}`, {
			method,
			headers: this.getHeaders(),
			body: body ? JSON.stringify(body) : undefined,
		});

		if (response.status === 401) {
			const refreshed = await this.tryRefresh();
			if (refreshed) {
				return this.request(method, path, body);
			}
			useAuthStore.getState().logout();
			throw new Error("Session expired");
		}

		if (!response.ok) {
			let error: ApiError;
			try {
				error = await response.json();
			} catch {
				error = { code: response.status, message: response.statusText };
			}
			throw error;
		}

		if (response.status === 204) return undefined as T;

		const text = await response.text();
		if (!text) return undefined as T;
		return JSON.parse(text) as T;
	}

	/**
	 * Пытается обновить access-токен через refresh-токен.
	 *
	 * Single-flight: если refresh уже в работе — дожидается того же
	 * промиса вместо запуска параллельного refresh (FE-H1).
	 *
	 * @returns true если обновление успешно, false если нужен повторный логин
	 */
	private tryRefresh(): Promise<boolean> {
		if (refreshing) return refreshing;
		refreshing = this.doRefresh().finally(() => {
			refreshing = null;
		});
		return refreshing;
	}

	/** Фактический HTTP-запрос на обновление токена (без single-flight обвязки). */
	private async doRefresh(): Promise<boolean> {
		const refreshToken = useAuthStore.getState().refreshToken;
		if (!refreshToken) return false;
		try {
			const res = await fetch(`${API_BASE}/auth/refresh`, {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ refresh_token: refreshToken }),
			});
			if (!res.ok) return false;
			const data = (await res.json()) as {
				access_token: string;
				refresh_token: string;
			};
			useAuthStore.getState().setTokens(data.access_token, data.refresh_token);
			return true;
		} catch {
			return false;
		}
	}

	/** GET-запрос. */
	get<T>(path: string) {
		return this.request<T>("GET", path);
	}
	/** POST-запрос. */
	post<T>(path: string, body?: unknown) {
		return this.request<T>("POST", path, body);
	}
	/** PUT-запрос. */
	put<T>(path: string, body?: unknown) {
		return this.request<T>("PUT", path, body);
	}
	/** PATCH-запрос (частичное обновление). */
	patch<T>(path: string, body?: unknown) {
		return this.request<T>("PATCH", path, body);
	}
	/** DELETE-запрос. */
	delete<T>(path: string) {
		return this.request<T>("DELETE", path);
	}
}

/** Глобальный экземпляр API-клиента. Используется всеми API-модулями. */
export const api = new ApiClient();
