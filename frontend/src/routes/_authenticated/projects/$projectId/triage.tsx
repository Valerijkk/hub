import { createFileRoute } from "@tanstack/react-router";
import { TriageInbox } from "@/modules/tracker/triage-inbox";

function TriagePage() {
	const { projectId } = Route.useParams();
	return <TriageInbox projectId={projectId} />;
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/triage",
)({
	component: TriagePage,
});
