import { createFileRoute } from "@tanstack/react-router";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { LeaveRequestsPage } from "@/modules/hr/leave-requests";

function HrPage() {
	return (
		<RouteErrorBoundary module="HR">
			<LeaveRequestsPage />
		</RouteErrorBoundary>
	);
}

export const Route = createFileRoute("/_authenticated/hr")({
	component: HrPage,
});
