import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { AppShell } from "@/components/layout/app-shell";
import { PushPermissionBanner } from "@/components/notifications/push-permission-banner";
import { useWebSocket } from "@/hooks/use-websocket";
import { registerServiceWorker } from "@/lib/push-notifications";
import { resetAuthenticatedClientState } from "@/lib/reset-authenticated-client-state";
import { subscribeAuthSync, useAuthStore } from "@/stores/auth-store";

function AuthenticatedLayout() {
	// Подключаем WebSocket для real-time обновлений.
	useWebSocket();

	const navigate = useNavigate();
	const clearSessionLocal = useAuthStore((s) => s.clearSessionLocal);

	// Регистрируем Service Worker для push-уведомлений при входе.
	useEffect(() => {
		void registerServiceWorker();
	}, []);

	// Multi-tab sync: logout/login в другой вкладке — очищаем cache Query
	// и редиректим на login (или перезагружаем, чтобы переподписаться на
	// новый user_id). См. FE-H2.
	useEffect(() => {
		const unsubscribe = subscribeAuthSync((event) => {
			if (event.type === "logout") {
				clearSessionLocal();
				void navigate({ to: "/login" });
			} else if (event.type === "login") {
				const currentId = useAuthStore.getState().user?.id ?? null;
				// Same-user re-login (token refresh /me) — не перезагружаем вкладку.
				if (event.userId && currentId && event.userId === currentId) {
					return;
				}
				resetAuthenticatedClientState();
				window.location.reload();
			}
		});
		return unsubscribe;
	}, [navigate, clearSessionLocal]);

	return (
		<>
			<AppShell />
			<PushPermissionBanner />
		</>
	);
}

export const Route = createFileRoute("/_authenticated")({
	beforeLoad: () => {
		const { isAuthenticated } = useAuthStore.getState();
		if (!isAuthenticated) {
			throw redirect({ to: "/login" });
		}
	},
	component: AuthenticatedLayout,
});
