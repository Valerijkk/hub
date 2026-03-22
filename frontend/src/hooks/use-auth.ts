import { useCallback, useEffect } from "react";
import { getMe, type UserResponse } from "@/api/auth";
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

	const logout = useCallback(() => {
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
