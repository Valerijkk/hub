import { createFileRoute } from "@tanstack/react-router";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { OrgDirectory } from "@/modules/intranet/org-directory";

function DirectoryPage() {
	const { user } = Route.useSearch();
	return (
		<RouteErrorBoundary module="Intranet">
			<OrgDirectory selectedUserId={user} />
		</RouteErrorBoundary>
	);
}

export const Route = createFileRoute("/_authenticated/intranet/directory")({
	validateSearch: (search: Record<string, unknown>) => ({
		user: typeof search.user === "string" ? search.user : undefined,
	}),
	component: DirectoryPage,
});
