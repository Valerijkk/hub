import { createFileRoute } from "@tanstack/react-router";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { CrmPage } from "@/modules/crm/crm-page";

function CrmRoutePage() {
	return (
		<RouteErrorBoundary module="CRM">
			<CrmPage />
		</RouteErrorBoundary>
	);
}

export const Route = createFileRoute("/_authenticated/crm")({
	component: CrmRoutePage,
});
