/**
 * Widget "Team Activity" -- list of recent project actions.
 */
import { Users } from "lucide-react";
import type { ActivityEntry, ProjectStats } from "@/api/dashboard";

interface Props {
	data?: unknown;
}

/** Team activity widget. */
export function TeamActivityWidget({ data }: Props) {
	const projectData = data as ProjectStats | undefined;
	const activities = projectData?.team_activity ?? [];

	if (activities.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-t2">
				<Users className="h-8 w-8 opacity-40" />
				<p className="text-sm">Нет активности</p>
				<p className="text-xs opacity-60">Действия команды появятся здесь</p>
			</div>
		);
	}

	return (
		<div className="space-y-1">
			{activities.slice(0, 10).map((entry: ActivityEntry, idx: number) => (
				<div
					key={idx}
					className="group flex items-start gap-3 rounded-lg px-2 py-2 text-sm transition-colors duration-150 hover:bg-accent/50"
				>
					{/* Avatar */}
					<div className="h-7 w-7 rounded-full bg-gradient-to-br from-primary/20 to-primary/5 ring-1 ring-border/50 flex items-center justify-center text-xs font-semibold text-primary shrink-0">
						{entry.user_name?.charAt(0)?.toUpperCase() ?? "?"}
					</div>
					<div className="min-w-0 flex-1">
						<p className="leading-snug">
							<span className="font-medium">{entry.user_name}</span>{" "}
							<span className="text-t2">{entry.action}</span>
						</p>
						{entry.target && (
							<p className="text-xs text-t2 truncate mt-0.5">
								{entry.target}
							</p>
						)}
						<p className="text-[11px] text-t2/60 mt-0.5">
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
