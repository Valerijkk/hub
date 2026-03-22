/**
 * Виджет "Активность команды" -- список последних действий в проекте.
 */
import type { ActivityEntry, ProjectStats } from "@/api/dashboard";

interface Props {
	data?: unknown;
}

/** Виджет активности команды. */
export function TeamActivityWidget({ data }: Props) {
	const projectData = data as ProjectStats | undefined;
	const activities = projectData?.team_activity ?? [];

	if (activities.length === 0) {
		return <p className="text-sm text-muted-foreground">Нет активности</p>;
	}

	return (
		<div className="space-y-2">
			{activities.slice(0, 10).map((entry: ActivityEntry, idx: number) => (
				<div
					key={idx}
					className="flex items-start gap-2 text-sm border-b pb-2 last:border-0"
				>
					<div className="h-6 w-6 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium shrink-0">
						{entry.user_name?.charAt(0)?.toUpperCase() ?? "?"}
					</div>
					<div className="min-w-0">
						<p>
							<span className="font-medium">{entry.user_name}</span>{" "}
							<span className="text-muted-foreground">{entry.action}</span>
						</p>
						{entry.target && (
							<p className="text-xs text-muted-foreground truncate">
								{entry.target}
							</p>
						)}
						<p className="text-xs text-muted-foreground">
							{new Date(entry.created_at).toLocaleString("ru-RU", {
								hour: "2-digit",
								minute: "2-digit",
								day: "numeric",
								month: "short",
							})}
						</p>
					</div>
				</div>
			))}
		</div>
	);
}
