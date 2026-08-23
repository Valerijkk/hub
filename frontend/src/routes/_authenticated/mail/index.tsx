import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

function MailIndex() {
	const { t } = useTranslation("mail");

	return <p className="text-sm text-t2">{t("active.workspace")}</p>;
}

export const Route = createFileRoute("/_authenticated/mail/")({
	component: MailIndex,
});
