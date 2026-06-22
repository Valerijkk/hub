import { createFileRoute } from "@tanstack/react-router";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { BasesPage } from "@/modules/bases/bases-page";

function BasesRoutePage() {
	return (
		<RouteErrorBoundary module="Hub Base">
			<BasesPage />
		</RouteErrorBoundary>
	);
}

export const Route = createFileRoute("/_authenticated/bases")({
	component: BasesRoutePage,
});
