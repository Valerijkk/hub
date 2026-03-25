import { createFileRoute } from "@tanstack/react-router";
import { AutotestSetup } from "@/modules/tms/autotest-setup";

/**
 * Страница настройки интеграции автотестов.
 * Инструкции по подключению CI/CD пайплайнов к TMS через webhook.
 */
function AutotestPage() {
	const { projectId } = Route.useParams();

	return <AutotestSetup projectId={projectId} />;
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/tms/autotest",
)({
	component: AutotestPage,
});
