import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { AppearanceSettings } from "@/modules/settings/appearance";

function AppearancePage() {
	const { t } = useTranslation("settings");

	return (
		<div className="mx-auto max-w-2xl space-y-6">
			<div className="mb-2">
				<h1 className="text-2xl font-semibold tracking-tight">{t("appearance")}</h1>
				<p className="mt-1 text-sm text-muted-foreground">
					Тема, шрифт и визуальные настройки интерфейса
				</p>
			</div>
			<AppearanceSettings />
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/settings/appearance")({
	component: AppearancePage,
});
