import { createFileRoute, redirect } from "@tanstack/react-router";
import { ResourceLimits } from "@/components/admin/resource-limits";
import { isSystemAdmin } from "@/lib/is-system-admin";
import { useAuthStore } from "@/stores/auth-store";

/** Страница настройки лимитов ресурсов (только для администратора). */
function LimitsPage() {
	return (
		<div className="flex-1 overflow-y-auto p-6 animate-[eu_450ms_var(--spring)]">
			<ResourceLimits />
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/admin/limits")({
	beforeLoad: () => {
		const { user } = useAuthStore.getState();
		if (!user || !isSystemAdmin(user.system_role)) {
			throw redirect({ to: "/" });
		}
	},
	component: LimitsPage,
});
