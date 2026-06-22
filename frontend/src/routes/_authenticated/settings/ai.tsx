import { createFileRoute } from "@tanstack/react-router";
import { UserAISettings } from "@/modules/settings/ai-privacy";

function AISettingsPage() {
	return <UserAISettings />;
}

export const Route = createFileRoute("/_authenticated/settings/ai")({
	component: AISettingsPage,
});
