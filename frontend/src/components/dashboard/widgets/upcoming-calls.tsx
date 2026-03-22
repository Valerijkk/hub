/**
 * Widget "Upcoming Calls" -- list of scheduled calls.
 */
import { Phone, PhoneOff } from "lucide-react";
import type { CallSummary, PersonalData } from "@/api/dashboard";
import { Button } from "@/components/ui/button";

interface Props {
	data?: unknown;
}

/** Upcoming calls widget. */
export function UpcomingCallsWidget({ data }: Props) {
	const personalData = data as PersonalData | undefined;
	const calls = personalData?.upcoming_calls ?? [];

	if (calls.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-muted-foreground">
				<PhoneOff className="h-8 w-8 opacity-40" />
				<p className="text-sm">Нет предстоящих звонков</p>
				<p className="text-xs opacity-60">Запланированные звонки появятся здесь</p>
			</div>
		);
	}

	return (
		<div className="space-y-1.5">
			{calls.map((call: CallSummary) => (
				<div
					key={call.id}
					className="group flex items-center justify-between gap-2 rounded-lg border border-border/50 p-2.5 text-sm transition-colors duration-150 hover:bg-accent/50"
				>
					<div className="flex items-center gap-2.5 min-w-0">
						<div className="h-7 w-7 rounded-full bg-violet-500/10 flex items-center justify-center shrink-0">
							<Phone className="h-3.5 w-3.5 text-violet-500" />
						</div>
						<div className="min-w-0">
							<p className="truncate font-medium">{call.title}</p>
							<p className="text-xs text-muted-foreground">
								{call.channel_name}
								<span className="mx-1 text-border">|</span>
								{new Date(call.scheduled_at).toLocaleString("ru-RU", {
									hour: "2-digit",
									minute: "2-digit",
									day: "numeric",
									month: "short",
								})}
							</p>
						</div>
					</div>
					<Button
						variant="outline"
						size="sm"
						className="shrink-0 gap-1.5 transition-all duration-200 hover:bg-violet-500/10 hover:text-violet-600 hover:border-violet-500/30"
					>
						<Phone className="h-3 w-3" />
						Войти
					</Button>
				</div>
			))}
		</div>
	);
}
