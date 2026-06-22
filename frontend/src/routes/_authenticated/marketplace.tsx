import { createFileRoute } from "@tanstack/react-router";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { MarketplacePage } from "@/modules/marketplace/marketplace-page";

function MarketplaceRoutePage() {
	return (
		<RouteErrorBoundary module="Marketplace">
			<MarketplacePage />
		</RouteErrorBoundary>
	);
}

export const Route = createFileRoute("/_authenticated/marketplace")({
	component: MarketplaceRoutePage,
});
