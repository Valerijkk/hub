import { createFileRoute, redirect } from "@tanstack/react-router";
import { useEffect } from "react";
import { AppShell } from "@/components/layout/app-shell";
import { PushPermissionBanner } from "@/components/notifications/push-permission-banner";
import { useWebSocket } from "@/hooks/use-websocket";
import { registerServiceWorker } from "@/lib/push-notifications";
import { useAuthStore } from "@/stores/auth-store";

function AuthenticatedLayout() {
	// Подключаем WebSocket для real-time обновлений.
	useWebSocket();

	// Регистрируем Service Worker для push-уведомлений при входе.
	useEffect(() => {
		void registerServiceWorker();
	}, []);

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
