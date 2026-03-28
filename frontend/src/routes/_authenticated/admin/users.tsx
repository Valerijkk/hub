import { createFileRoute, redirect } from "@tanstack/react-router";
import { UserManagement } from "@/components/admin/user-management";
import { useAuthStore } from "@/stores/auth-store";

/** Страница управления пользователями (только для администратора). */
function UsersPage() {
	return (
		<div className="flex-1 overflow-y-auto p-6 animate-[eu_450ms_var(--spring)]">
			<UserManagement />
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/admin/users")({
	beforeLoad: () => {
		const { user } = useAuthStore.getState();
		if (!user || user.system_role !== "admin") {
			throw redirect({ to: "/" });
		}
	},
	component: UsersPage,
});
