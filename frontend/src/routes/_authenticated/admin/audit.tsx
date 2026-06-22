import { createFileRoute, redirect } from "@tanstack/react-router";
import { AuditLogViewer } from "@/components/admin/audit-log";
import { isSystemAdmin } from "@/lib/is-system-admin";
import { useAuthStore } from "@/stores/auth-store";

/** Страница журнала аудита (только для администратора). */
function AuditPage() {
	return (
		<div className="flex-1 overflow-y-auto p-6 animate-[eu_450ms_var(--spring)]">
			<AuditLogViewer />
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/admin/audit")({
	beforeLoad: () => {
		const { user } = useAuthStore.getState();
		if (!user || !isSystemAdmin(user.system_role)) {
			throw redirect({ to: "/" });
		}
	},
	component: AuditPage,
});
