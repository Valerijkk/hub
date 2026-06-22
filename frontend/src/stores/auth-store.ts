import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

import type { UserResponse } from "@/api/auth";

/**
 * Состояние стора аутентификации.
 * Персистится в sessionStorage под ключом "hub-auth".
 * Сессия завершается при закрытии вкладки/браузера.
 */
interface AuthState {
	/** JWT access-токен для авторизации API-запросов. */
	accessToken: string | null;
	/** Refresh-токен для обновления access-токена. */
	refreshToken: string | null;
	/** Данные текущего пользователя (загружаются отдельно через /auth/me). */
	user: UserResponse | null;
	/** Флаг аутентификации (true после получения токенов). */
	isAuthenticated: boolean;
	/** Сохранить пару токенов после логина. */
	setTokens: (access: string, refresh: string) => void;
	/** Установить данные пользователя. */
	setUser: (user: UserResponse) => void;
	/** Очистить все данные аутентификации (выход). */
	logout: () => void;
}

/**
 * Multi-tab синхронизация login/logout через BroadcastChannel (FE-H2).
 *
 * Когда пользователь делает login/logout в одной вкладке, событие
 * транслируется в другие вкладки того же origin. Это позволяет:
 *  - очистить TanStack Query cache при logout (см. useAuthSync в _authenticated.tsx)
 *  - переключить вкладку на страницу логина / главную при login
 *
 * Реализовано через BroadcastChannel API (есть в Chromium/Firefox/Safari 15.4+).
 * Если API недоступен — sync молча отключается (single-tab fallback).
 */
const AUTH_CHANNEL_NAME = "hub-auth";

/** Типы событий, которые транслируются между вкладками. */
export type AuthSyncEvent =
	| { type: "logout" }
	| { type: "login"; userId: string | null };

let authChannel: BroadcastChannel | null = null;
if (typeof BroadcastChannel !== "undefined") {
	try {
		authChannel = new BroadcastChannel(AUTH_CHANNEL_NAME);
	} catch {
		authChannel = null;
	}
}

/**
 * Транслирует событие auth-sync в другие вкладки.
 * Используется внутри setTokens/logout стора.
 */
function broadcastAuth(event: AuthSyncEvent): void {
	if (!authChannel) return;
	try {
		authChannel.postMessage(event);
	} catch {
		// игнорируем: канал мог закрыться
	}
}

/**
 * Подписаться на auth-sync события из других вкладок.
 * Возвращает функцию отписки.
 *
 * @param handler - обработчик события (получает event payload)
 */
export function subscribeAuthSync(
	handler: (event: AuthSyncEvent) => void,
): () => void {
	if (typeof BroadcastChannel !== "undefined" && authChannel) {
		const listener = (ev: MessageEvent<AuthSyncEvent>) => {
			handler(ev.data);
		};
		authChannel.addEventListener("message", listener);
		return () => {
			authChannel?.removeEventListener("message", listener);
		};
	}

	// Запасной вариант: storage event (для очень старых браузеров и кейсов, когда
	// BroadcastChannel недоступен). sessionStorage НЕ триггерит storage event
	// между вкладками, но persist-обновления mirror'ятся через localStorage
	// в некоторых конфигурациях — оставлено как defence-in-depth.
	const storageListener = (ev: StorageEvent) => {
		if (ev.key === `${AUTH_CHANNEL_NAME}-event` && ev.newValue) {
			try {
				handler(JSON.parse(ev.newValue) as AuthSyncEvent);
			} catch {
				// игнорируем: операция не критична для UX
			}
		}
	};
	window.addEventListener("storage", storageListener);
	return () => window.removeEventListener("storage", storageListener);
}

/** Стор аутентификации с персистом в sessionStorage. */
export const useAuthStore = create<AuthState>()(
	persist(
		(set) => ({
			accessToken: null,
			refreshToken: null,
			user: null,
			isAuthenticated: false,
			setTokens: (access, refresh) => {
				set({
					accessToken: access,
					refreshToken: refresh,
					isAuthenticated: true,
				});
				broadcastAuth({ type: "login", userId: null });
			},
			setUser: (user) => set({ user }),
			logout: () => {
				set({
					accessToken: null,
					refreshToken: null,
					user: null,
					isAuthenticated: false,
				});
				broadcastAuth({ type: "logout" });
			},
		}),
		{ name: "hub-auth", storage: createJSONStorage(() => sessionStorage) },
	),
);
