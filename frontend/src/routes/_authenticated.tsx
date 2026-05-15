import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { AppShell } from "@/components/layout/app-shell";
import { PushPermissionBanner } from "@/components/notifications/push-permission-banner";
import { useWebSocket } from "@/hooks/use-websocket";
import { registerServiceWorker } from "@/lib/push-notifications";
import { subscribeAuthSync, useAuthStore } from "@/stores/auth-store";

function AuthenticatedLayout() {
	// Подключаем WebSocket для real-time обновлений.
	useWebSocket();

	const queryClient = useQueryClient();
	const navigate = useNavigate();
	const storeLogout = useAuthStore((s) => s.logout);

	// Регистрируем Service Worker для push-уведомлений при входе.
	useEffect(() => {
		void registerServiceWorker();
	}, []);

	// Multi-tab sync: при logout/login в другой вкладке очищаем cache
	// и редиректим на login (или перезагружаем, чтобы переподписаться на
	// новый user_id). См. FE-H2.
	useEffect(() => {
		const unsubscribe = subscribeAuthSync((event) => {
			if (event.type === "logout") {
				queryClient.clear();
				// Если стор ещё считает нас залогиненными — синхронизируем локально
				// (без повторного broadcast, чтобы не зациклить).
				const { isAuthenticated } = useAuthStore.getState();
				if (isAuthenticated) {
					storeLogout();
				}
				void navigate({ to: "/login" });
			} else if (event.type === "login") {
				// Другая вкладка залогинилась под, возможно, другим user_id:
				// чистим cache и перезагружаем приложение, чтобы переподписаться
				// на корректный WebSocket-канал.
				queryClient.clear();
				window.location.reload();
			}
		});
		return unsubscribe;
	}, [queryClient, navigate, storeLogout]);

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
