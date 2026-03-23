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

/** Стор аутентификации с персистом в sessionStorage. */
export const useAuthStore = create<AuthState>()(
	persist(
		(set) => ({
			accessToken: null,
			refreshToken: null,
			user: null,
			isAuthenticated: false,
			setTokens: (access, refresh) =>
				set({
					accessToken: access,
					refreshToken: refresh,
					isAuthenticated: true,
				}),
			setUser: (user) => set({ user }),
			logout: () =>
				set({
					accessToken: null,
					refreshToken: null,
					user: null,
					isAuthenticated: false,
				}),
		}),
		{ name: "hub-auth", storage: createJSONStorage(() => sessionStorage) },
	),
);
