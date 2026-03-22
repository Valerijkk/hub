import { createFileRoute, redirect } from "@tanstack/react-router";
import { TrashManager } from "@/components/admin/trash-manager";
import { useAuthStore } from "@/stores/auth-store";

/** Страница управления корзиной (только для администратора). */
function TrashPage() {
	return (
		<div className="mx-auto max-w-6xl p-6">
			<TrashManager />
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/admin/trash")({
	beforeLoad: () => {
		const { user } = useAuthStore.getState();
		if (!user || user.system_role !== "admin") {
			throw redirect({ to: "/" });
		}
	},
	component: TrashPage,
});
