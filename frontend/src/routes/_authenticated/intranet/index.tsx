import { createFileRoute } from "@tanstack/react-router";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { IntranetHub } from "@/modules/intranet/intranet-hub";

function IntranetPage() {
	return (
		<RouteErrorBoundary module="Intranet">
			<IntranetHub />
		</RouteErrorBoundary>
	);
}

export const Route = createFileRoute("/_authenticated/intranet/")({
	component: IntranetPage,
	validateSearch: (search: Record<string, unknown>) => ({
		tab: typeof search.tab === "string" ? search.tab : undefined,
	}),
});
