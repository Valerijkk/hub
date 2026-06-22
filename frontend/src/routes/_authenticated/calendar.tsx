import { createFileRoute } from "@tanstack/react-router";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { UnifiedCalendar } from "@/modules/calendar/unified-calendar";

function CalendarPage() {
	return (
		<RouteErrorBoundary module="Calendar">
			<UnifiedCalendar />
		</RouteErrorBoundary>
	);
}

export const Route = createFileRoute("/_authenticated/calendar")({
	component: CalendarPage,
});
