import { createFileRoute, redirect } from "@tanstack/react-router";
import { useAuthStore } from "@/stores/auth-store";

/**
 * Стаб маршрута /onboarding. Реальный компонент (~543 LOC) вынесен в
 * `onboarding.lazy.tsx` — TanStack Router code-split: страница онбординга
 * грузится только при первом заходе нового пользователя.
 *
 * beforeLoad здесь (не в lazy), чтобы guard сработал до загрузки chunk.
 */
export const Route = createFileRoute("/onboarding")({
	beforeLoad: () => {
		const { isAuthenticated, user } = useAuthStore.getState();
		if (!isAuthenticated) {
			throw redirect({ to: "/login" });
		}
		if (user?.onboarding_completed) {
			throw redirect({ to: "/projects" });
		}
	},
});
