import { createLazyFileRoute } from "@tanstack/react-router";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { AccountingPage } from "@/modules/accounting/accounting-page";

function AccountingRoutePage() {
	return (
		<RouteErrorBoundary module="Accounting">
			<AccountingPage />
		</RouteErrorBoundary>
	);
}

export const Route = createLazyFileRoute("/_authenticated/accounting")({
	component: AccountingRoutePage,
});
