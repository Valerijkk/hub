import { createFileRoute, redirect } from "@tanstack/react-router";
import { MailSettingsPanel } from "@/components/admin/mail-settings";
import { isSystemAdmin } from "@/lib/is-system-admin";
import { useAuthStore } from "@/stores/auth-store";

function AdminMailPage() {
	return (
		<div className="flex-1 overflow-y-auto p-6 animate-[eu_450ms_var(--spring)]">
			<MailSettingsPanel />
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/admin/mail")({
	beforeLoad: () => {
		const user = useAuthStore.getState().user;
		if (!user || !isSystemAdmin(user.system_role)) {
			throw redirect({ to: "/dashboard" });
		}
	},
	component: AdminMailPage,
});
