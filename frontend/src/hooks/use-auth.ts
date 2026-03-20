import { useCallback, useEffect } from "react";
import { useAuthStore } from "@/stores/auth-store";
import { getMe, type UserResponse } from "@/api/auth";

/**
 * useAuth provides authentication state and actions.
 *
 * Automatically fetches the current user profile when authenticated.
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

  // Fetch user profile when authenticated but user data is missing.
  useEffect(() => {
    if (isAuthenticated && accessToken && !user) {
      getMe()
        .then((data: UserResponse) => setUser(data))
        .catch(() => {
          // Token might be invalid; logout.
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
