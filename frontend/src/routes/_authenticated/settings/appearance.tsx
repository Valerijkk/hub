import { createFileRoute } from "@tanstack/react-router";
import { AppearanceSettings } from "@/modules/settings/appearance";

function AppearancePage() {
	return (
		<div>
			<AppearanceSettings />
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/settings/appearance")({
	component: AppearancePage,
});
