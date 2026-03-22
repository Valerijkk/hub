/**
 * Widget "Time Tracking" -- displays tracked time for today.
 */
import { Clock } from "lucide-react";
import type { PersonalData } from "@/api/dashboard";

interface Props {
	data?: unknown;
}

/** Format minutes to hours:minutes. */
function formatMinutes(minutes: number): string {
	const h = Math.floor(minutes / 60);
	const m = minutes % 60;
	return `${h}ч ${m}м`;
}

/** Time tracking widget. */
export function TimeTrackingWidget({ data }: Props) {
	const personalData = data as PersonalData | undefined;
	const timeToday = personalData?.time_today ?? 0;

	// Progress: 8 hours = 100%
	const targetMinutes = 8 * 60;
	const progress = Math.min((timeToday / targetMinutes) * 100, 100);
	const isComplete = progress >= 100;

	return (
		<div className="flex flex-col items-center justify-center h-full gap-3">
			<Clock className={`h-8 w-8 ${isComplete ? "text-green-500" : "text-muted-foreground opacity-60"}`} />
			<div className="text-center">
				<span className="text-3xl font-bold tabular-nums">
					{formatMinutes(timeToday)}
				</span>
				<p className="text-sm text-muted-foreground mt-0.5">сегодня</p>
			</div>
			<div className="w-full max-w-48 space-y-1.5">
				<div className="h-2 bg-muted rounded-full overflow-hidden">
					<div
						className={`h-full rounded-full transition-all duration-500 ease-out ${
							isComplete
								? "bg-green-500"
								: progress > 60
									? "bg-primary"
									: "bg-amber-500"
						}`}
						style={{ width: `${progress}%` }}
					/>
				</div>
				<div className="flex justify-between text-xs text-muted-foreground tabular-nums">
					<span>{formatMinutes(timeToday)}</span>
					<span>{formatMinutes(targetMinutes)}</span>
				</div>
			</div>
		</div>
	);
}
