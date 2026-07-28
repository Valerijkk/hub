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
	/** Секунды до повтора (429, из заголовка Retry-After). */
	retry_after?: number;
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
	/**
	 * Формирует заголовки запроса, включая Authorization.
	 * Для FormData не ставим Content-Type — браузер сам добавит multipart boundary.
	 */
	private getHeaders(body?: unknown): Record<string, string> {
		const headers: Record<string, string> = {};
		if (!(body instanceof FormData)) {
			headers["Content-Type"] = "application/json";
		}
		const token = useAuthStore.getState().accessToken;
		if (token) {
			headers.Authorization = `Bearer ${token}`;
		}
		return headers;
	}

	/**
	 * Best-effort POST /auth/logout до локальной очистки (W11-A1-007 / P2-FE-09).
	 *
	 * Идёт через сырой fetch, а не api.post — иначе 401 на logout снова
	 * запустит refresh и возможный цикл. Токен захватываем до clear.
	 */
	private async revokeSessionBestEffort(): Promise<void> {
		const token = useAuthStore.getState().accessToken;
		if (!token) return;
		try {
			await fetch(`${API_BASE}/auth/logout`, {
				method: "POST",
				headers: {
					Authorization: `Bearer ${token}`,
					"Content-Type": "application/json",
				},
			});
		} catch {
			// сеть/CORS — локальный logout всё равно выполним ниже
		}
	}

	/** Инвалидировать серверную сессию (best-effort) и очистить локальный стор. */
	private async expireSession(): Promise<void> {
		await this.revokeSessionBestEffort();
		useAuthStore.getState().logout();
	}

	/**
	 * Выполняет HTTP-запрос к API.
	 * При 401 пытается обновить токены и повторить запрос **один раз**
	 * (`_retried`): повторный 401 после refresh → logout, без рекурсии.
	 * При 204 или пустом теле возвращает undefined.
	 * @param method - HTTP-метод (GET, POST, PUT, PATCH, DELETE)
	 * @param path - путь относительно API_BASE (например, "/auth/me")
	 * @param body - тело запроса (JSON-объект или FormData без stringify)
	 * @param _retried - внутренний флаг: уже был один retry после refresh
	 */
	async request<T>(
		method: string,
		path: string,
		body?: unknown,
		_retried = false,
	): Promise<T> {
		const isFormData = body instanceof FormData;
		const response = await fetch(`${API_BASE}${path}`, {
			method,
			headers: this.getHeaders(body),
			body:
				body === undefined || body === null
					? undefined
					: isFormData
						? (body as FormData)
						: JSON.stringify(body),
		});

		if (response.status === 401) {
			if (_retried) {
				await this.expireSession();
				throw new Error("Session expired");
			}
			const refreshed = await this.tryRefresh();
			if (refreshed) {
				return this.request(method, path, body, true);
			}
			await this.expireSession();
			throw new Error("Session expired");
		}

		if (!response.ok) {
			let error: ApiError;
			try {
				error = await response.json();
				if (!error.code) error.code = response.status;
			} catch {
				error = { code: response.status, message: response.statusText };
			}
			if (response.status === 429) {
				const retryHeader = response.headers.get("Retry-After");
				if (retryHeader) {
					const seconds = Number.parseInt(retryHeader, 10);
					if (!Number.isNaN(seconds) && seconds > 0) {
						error.retry_after = seconds;
					}
				}
			}
			throw error;
		}

		if (response.status === 204) return undefined as T;

		const text = await response.text();
		if (!text) return undefined as T;
		try {
			return JSON.parse(text) as T;
		} catch {
			const parseError: ApiError = {
				code: 502,
				message: "invalid json",
			};
			throw parseError;
		}
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

	/**
	 * GET binary body as Blob (CSV/XLSX/ICS exports).
	 * Не использует JSON.parse — в отличие от request() (A11-007).
	 */
	async getBlob(path: string): Promise<Blob> {
		const doFetch = async (retried = false): Promise<Response> => {
			const response = await fetch(`${API_BASE}${path}`, {
				method: "GET",
				headers: this.getHeaders(),
			});
			if (response.status === 401) {
				if (retried) {
					await this.expireSession();
					throw new Error("Session expired");
				}
				const refreshed = await this.tryRefresh();
				if (refreshed) return doFetch(true);
				await this.expireSession();
				throw new Error("Session expired");
			}
			return response;
		};

		const response = await doFetch();
		if (!response.ok) {
			let error: ApiError;
			try {
				error = await response.json();
				if (!error.code) error.code = response.status;
			} catch {
				error = { code: response.status, message: response.statusText };
			}
			throw error;
		}
		return response.blob();
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
