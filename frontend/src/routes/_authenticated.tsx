import { createFileRoute, redirect } from "@tanstack/react-router";
import { AppShell } from "@/components/layout/app-shell";
import { useWebSocket } from "@/hooks/use-websocket";
import { useAuthStore } from "@/stores/auth-store";

function AuthenticatedLayout() {
	// Connect WebSocket for real-time updates
	useWebSocket();
	return <AppShell />;
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
