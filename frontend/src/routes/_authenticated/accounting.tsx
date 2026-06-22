import { createFileRoute, redirect } from "@tanstack/react-router";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { isSystemAdmin } from "@/lib/is-system-admin";
import { AccountingPage } from "@/modules/accounting/accounting-page";
import { useAuthStore } from "@/stores/auth-store";

export const Route = createFileRoute("/_authenticated/accounting")({
	beforeLoad: () => {
		const user = useAuthStore.getState().user;
		if (!user || !isSystemAdmin(user.system_role)) {
			throw redirect({ to: "/dashboard" });
		}
	},
	component: AccountingRoutePage,
});

function AccountingRoutePage() {
	return (
		<RouteErrorBoundary module="Accounting">
			<AccountingPage />
		</RouteErrorBoundary>
	);
}
