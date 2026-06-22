import { createFileRoute, Link } from "@tanstack/react-router";
import { BookOpen, Keyboard, Search, TestTube2, Webhook } from "lucide-react";
import { useTranslation } from "react-i18next";
import { isSystemAdmin } from "@/lib/is-system-admin";
import { useAuthStore } from "@/stores/auth-store";

function HelpPage() {
	const { t } = useTranslation("help");
	const isAdmin = useAuthStore((s) => isSystemAdmin(s.user?.system_role));

	const sectionKeys = [
		"quickStart",
		"hotkeys",
		"allure",
		"integrations",
	] as const;

	return (
		<div className="mx-auto max-w-3xl space-y-8">
			<header className="space-y-2">
				<div className="flex items-center gap-3">
					<BookOpen className="h-8 w-8 text-amber-400" aria-hidden />
					<h1
						className="text-2xl font-bold tracking-tight"
						data-testid="help-page-heading"
					>
						{t("title")}
					</h1>
				</div>
				<p className="text-t2">{t("subtitle")}</p>
			</header>

			<div className="grid gap-4 sm:grid-cols-2">
				<Link
					to="/settings/hotkeys"
					className="flex items-start gap-3 rounded-xl border border-b1 bg-[var(--hover)]/30 p-4 transition-colors hover:bg-[var(--hover)]"
				>
					<Keyboard className="mt-0.5 h-5 w-5 text-amber-400" aria-hidden />
					<div>
						<p className="font-medium">{t("cards.hotkeys.title")}</p>
						<p className="text-sm text-t2">{t("cards.hotkeys.desc")}</p>
					</div>
				</Link>
				<div className="flex items-start gap-3 rounded-xl border border-b1 bg-[var(--hover)]/30 p-4">
					<Search className="mt-0.5 h-5 w-5 text-blue-400" aria-hidden />
					<div>
						<p className="font-medium">{t("cards.search.title")}</p>
						<p className="text-sm text-t2">{t("cards.search.desc")}</p>
					</div>
				</div>
				<div className="flex items-start gap-3 rounded-xl border border-b1 bg-[var(--hover)]/30 p-4">
					<TestTube2 className="mt-0.5 h-5 w-5 text-green-400" aria-hidden />
					<div>
						<p className="font-medium">{t("cards.tms.title")}</p>
						<p className="text-sm text-t2">{t("cards.tms.desc")}</p>
					</div>
				</div>
				<div className="flex items-start gap-3 rounded-xl border border-b1 bg-[var(--hover)]/30 p-4">
					<Webhook className="mt-0.5 h-5 w-5 text-purple-400" aria-hidden />
					<div>
						<p className="font-medium">{t("cards.integrations.title")}</p>
						<p className="text-sm text-t2">{t("cards.integrations.desc")}</p>
					</div>
				</div>
			</div>

			{sectionKeys.map((key) => (
				<section key={key} className="space-y-3">
					<h2 className="text-lg font-semibold">
						{t(`sections.${key}.title`)}
					</h2>
					<ul className="list-disc space-y-1.5 pl-5 text-sm text-t2">
						{(
							t(`sections.${key}.items`, { returnObjects: true }) as string[]
						).map((item) => (
							<li key={item}>{item}</li>
						))}
					</ul>
				</section>
			))}

			{isAdmin ? (
				<p className="border-t border-b1 pt-4 text-sm text-t2">
					{t("adminLink")}{" "}
					<Link to="/admin/users" className="text-amber-400 hover:underline">
						{t("adminPanel")}
					</Link>
				</p>
			) : null}
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/help")({
	component: HelpPage,
});
