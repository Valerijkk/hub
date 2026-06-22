import { createFileRoute } from "@tanstack/react-router";
import { FlakyTestsReport } from "@/modules/tms/components/flaky-tests-report";

function FlakyTestsPage() {
	const { projectId } = Route.useParams();
	return <FlakyTestsReport projectId={projectId} />;
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/tms/flaky",
)({
	component: FlakyTestsPage,
});
