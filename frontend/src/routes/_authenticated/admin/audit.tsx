import { createFileRoute, redirect } from "@tanstack/react-router";
import { AuditLogViewer } from "@/components/admin/audit-log";
import { useAuthStore } from "@/stores/auth-store";

/** Страница журнала аудита (только для администратора). */
function AuditPage() {
	return (
		<div className="mx-auto max-w-6xl p-6">
			<AuditLogViewer />
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/admin/audit")({
	beforeLoad: () => {
		const { user } = useAuthStore.getState();
		if (!user || user.system_role !== "admin") {
			throw redirect({ to: "/" });
		}
	},
	component: AuditPage,
});
