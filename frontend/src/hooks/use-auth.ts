import { useCallback, useEffect } from "react";
import { getMe, logout as logoutApi, type UserResponse } from "@/api/auth";
import { useAuthStore } from "@/stores/auth-store";

/**
 * useAuth предоставляет состояние аутентификации и действия (login/logout).
 *
 * Автоматически загружает профиль текущего пользователя при наличии токена.
 * Если токен невалиден — выполняет logout.
 *
 * @returns user — текущий пользователь, isAuthenticated — авторизован ли,
 *          isLoading — идёт загрузка профиля, login/logout — действия
 */
export function useAuth() {
	const {
		user,
		isAuthenticated,
		accessToken,
		setUser,
		setTokens,
		logout: storeLogout,
	} = useAuthStore();

	// Загрузка профиля при наличии токена, но отсутствии данных пользователя.
	useEffect(() => {
		if (isAuthenticated && accessToken && !user) {
			getMe()
				.then((data: UserResponse) => setUser(data))
				.catch(() => {
					// Токен невалиден — выходим.
					storeLogout();
				});
		}
	}, [isAuthenticated, accessToken, user, setUser, storeLogout]);

	const login = useCallback(
		(accessToken: string, refreshToken: string) => {
			setTokens(accessToken, refreshToken);
		},
		[setTokens],
	);

	/**
	 * Выход из системы (FE-H3).
	 *
	 * Сначала отправляем POST /auth/logout, чтобы сервер инвалидировал
	 * refresh-токен в БД. Только потом чистим локальные токены, иначе
	 * запрос уйдёт без авторизации.
	 *
	 * Если серверный logout упал (сеть, 5xx) — всё равно чистим локально,
	 * чтобы пользователь не остался в подвешенном состоянии. Refresh-токен
	 * протухнет естественным образом по TTL.
	 */
	const logout = useCallback(async () => {
		try {
			await logoutApi();
		} catch {
			// ignore: всё равно очищаем локальные токены
		}
		storeLogout();
	}, [storeLogout]);

	return {
		user,
		isAuthenticated,
		isLoading: isAuthenticated && !user,
		login,
		logout,
	};
}
