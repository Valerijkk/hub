/**
 * Виджет "Тайм-трекинг" -- отображение затреканного времени за сегодня.
 */
import type { PersonalData } from "@/api/dashboard";

interface Props {
	data?: unknown;
}

/** Форматирует минуты в часы:минуты. */
function formatMinutes(minutes: number): string {
	const h = Math.floor(minutes / 60);
	const m = minutes % 60;
	return `${h}ч ${m}м`;
}

/** Виджет тайм-трекинга. */
export function TimeTrackingWidget({ data }: Props) {
	const personalData = data as PersonalData | undefined;
	const timeToday = personalData?.time_today ?? 0;

	// Прогресс: 8 часов = 100%
	const targetMinutes = 8 * 60;
	const progress = Math.min((timeToday / targetMinutes) * 100, 100);

	return (
		<div className="flex flex-col items-center justify-center h-full gap-3">
			<span className="text-3xl font-bold">{formatMinutes(timeToday)}</span>
			<span className="text-sm text-muted-foreground">сегодня</span>
			<div className="w-full max-w-48 h-2 bg-muted rounded-full overflow-hidden">
				<div
					className="h-full bg-primary rounded-full transition-all"
					style={{ width: `${progress}%` }}
				/>
			</div>
			<span className="text-xs text-muted-foreground">
				{formatMinutes(timeToday)} / {formatMinutes(targetMinutes)}
			</span>
		</div>
	);
}
