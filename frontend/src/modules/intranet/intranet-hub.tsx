import { useNavigate, useSearch } from "@tanstack/react-router";
import { cn } from "@/lib/utils";
import { KudosFeed } from "@/modules/intranet/kudos-feed";
import { NewsFeed } from "@/modules/intranet/news-feed";
import { RandomCoffeePanel } from "@/modules/intranet/random-coffee";

type IntranetTab = "news" | "kudos" | "coffee";

function tabFromSearch(tab?: string): IntranetTab {
	if (tab === "kudos") return "kudos";
	if (tab === "coffee") return "coffee";
	return "news";
}

/** Хаб intranet: новости + kudos + Random Coffee (MOD-01 + MOD-03 + V4-05). */
export function IntranetHub() {
	const navigate = useNavigate();
	const search = useSearch({ strict: false }) as { tab?: string };
	const tab = tabFromSearch(search.tab);

	const setTab = (next: IntranetTab) => {
		void navigate({
			to: "/intranet",
			search: {
				tab: next === "news" ? undefined : next,
			} satisfies { tab: string | undefined },
			replace: true,
		});
	};

	return (
		<div className="mx-auto max-w-3xl space-y-4 px-4 py-6">
			<header className="space-y-1">
				<h1 className="text-2xl font-bold tracking-tight">Intranet</h1>
				<p className="text-sm text-t2">
					Новости, благодарности и Random Coffee
				</p>
			</header>
			<div className="flex gap-2 border-b border-b1">
				{(
					[
						["news", "Новости", "border-amber-400"],
						["kudos", "Спасибо", "border-pink-400"],
						["coffee", "Coffee", "border-amber-500"],
					] as const
				).map(([id, label, color]) => (
					<button
						key={id}
						type="button"
						onClick={() => setTab(id)}
						className={cn(
							"px-3 py-2 text-sm font-medium border-b-2 -mb-px transition-colors",
							tab === id
								? cn(color, "text-t1")
								: "border-transparent text-t3 hover:text-t2",
						)}
					>
						{label}
					</button>
				))}
			</div>
			{tab === "news" && <NewsFeed embedded />}
			{tab === "kudos" && <KudosFeed />}
			{tab === "coffee" && <RandomCoffeePanel />}
		</div>
	);
}
