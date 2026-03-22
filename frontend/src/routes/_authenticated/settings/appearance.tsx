import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { AppearanceSettings } from "@/modules/settings/appearance";

function AppearancePage() {
	const { t } = useTranslation("settings");

	return (
		<div className="mx-auto max-w-2xl space-y-6">
			<h1 className="text-2xl font-bold">{t("appearance")}</h1>
			<AppearanceSettings />
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/settings/appearance")({
	component: AppearancePage,
});
