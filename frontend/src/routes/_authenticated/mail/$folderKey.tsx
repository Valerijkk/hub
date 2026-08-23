import {
	createFileRoute,
	Link,
	Outlet,
	useRouterState,
} from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import {
	isValidMailFolderKey,
	MailThreadList,
} from "@/modules/mail/mail-thread-list";

function InvalidFolder() {
	const { t } = useTranslation("mail");
	return (
		<section
			aria-labelledby="mail-invalid-folder-heading"
			className="flex min-h-[20rem] items-center justify-center rounded-[var(--r3)] border border-b1 bg-card p-6 text-center"
		>
			<div className="max-w-sm">
				<h2 id="mail-invalid-folder-heading" className="text-lg font-semibold text-t1">
					{t("folder.invalidTitle")}
				</h2>
				<p className="mt-2 text-sm leading-6 text-t2">
					{t("folder.invalidDescription")}
				</p>
				<Button asChild variant="outline" className="mt-4">
					<Link to="/mail">{t("folder.backToAll")}</Link>
				</Button>
			</div>
		</section>
	);
}

function FolderRoute() {
	const { folderKey } = Route.useParams();
	const { t } = useTranslation("mail");
	const pathname = useRouterState({ select: (state) => state.location.pathname });
	const segments = pathname.split("/").filter(Boolean);
	const threadChildOpen = segments[2] === "thread" && Boolean(segments[3]);

	if (!isValidMailFolderKey(folderKey)) return <InvalidFolder />;
	if (threadChildOpen) return <Outlet />;

	const normalizedFolderKey =
		folderKey === "all" ? "all" : folderKey.toLowerCase();
	return (
		<div className="flex min-h-[32rem] min-w-0 overflow-hidden rounded-[var(--r3)] border border-b1 bg-card">
			<MailThreadList folderKey={normalizedFolderKey} />
			<section
				aria-labelledby="mail-folder-reader-empty-heading"
				className="hidden min-w-0 flex-1 items-center justify-center bg-base/30 p-8 md:flex"
			>
				<div className="max-w-sm text-center">
					<h2
						id="mail-folder-reader-empty-heading"
						className="text-lg font-semibold tracking-tight text-t1"
					>
						{t("reader.emptyTitle")}
					</h2>
					<p className="mt-2 text-sm leading-6 text-t2">
						{t("reader.emptyDescription")}
					</p>
				</div>
			</section>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/mail/$folderKey")({
	component: FolderRoute,
});
