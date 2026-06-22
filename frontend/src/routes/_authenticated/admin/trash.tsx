import { createFileRoute, redirect } from "@tanstack/react-router";
import { TrashManager } from "@/components/admin/trash-manager";
import { isSystemAdmin } from "@/lib/is-system-admin";
import { useAuthStore } from "@/stores/auth-store";

/** Страница управления корзиной (только для администратора). */
function TrashPage() {
	return (
		<div className="flex-1 overflow-y-auto p-6 animate-[eu_450ms_var(--spring)]">
			<TrashManager />
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/admin/trash")({
	beforeLoad: () => {
		const { user } = useAuthStore.getState();
		if (!user || !isSystemAdmin(user.system_role)) {
			throw redirect({ to: "/" });
		}
	},
	component: TrashPage,
});
