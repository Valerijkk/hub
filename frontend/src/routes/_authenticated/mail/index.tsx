import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

function MailIndex() {
	const { t } = useTranslation("mail");

	return (
		<section aria-labelledby="mail-workspace-heading" className="flex min-h-48 items-center">
			<div className="max-w-md space-y-2">
				<h2 id="mail-workspace-heading" className="text-lg font-semibold tracking-tight text-t1">{t("active.workspaceTitle")}</h2>
				<p className="text-sm leading-6 text-t2">{t("active.workspace")}</p>
			</div>
		</section>
	);
}

export const Route = createFileRoute("/_authenticated/mail/")({
	component: MailIndex,
});
