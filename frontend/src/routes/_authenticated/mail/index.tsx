import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { MailThreadList } from "@/modules/mail/mail-thread-list";

function MailIndex() {
	const { t } = useTranslation("mail");

	return (
		<div className="flex min-h-[32rem] min-w-0 overflow-hidden rounded-[var(--r3)] border border-b1 bg-card">
			<MailThreadList folderKey="all" />
			<section
				aria-labelledby="mail-reader-empty-heading"
				className="hidden min-w-0 flex-1 items-center justify-center bg-base/30 p-8 md:flex"
			>
				<div className="max-w-sm text-center">
					<h2 id="mail-reader-empty-heading" className="text-lg font-semibold tracking-tight text-t1">
						{t("reader.emptyTitle")}
					</h2>
					<p className="mt-2 text-sm leading-6 text-t2">{t("reader.emptyDescription")}</p>
				</div>
			</section>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/mail/")({
	component: MailIndex,
});
