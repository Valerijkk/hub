/**
 * Виджет «Burndown» — ideal vs actual по дням спринта.
 */
import { TrendingDown } from "lucide-react";
import {
	Area,
	AreaChart,
	CartesianGrid,
	ResponsiveContainer,
	Tooltip,
	XAxis,
	YAxis,
} from "recharts";
import {
	type DashboardWidgetData,
	isProjectStats,
} from "@/components/dashboard/widget-data";

interface Props {
	data?: DashboardWidgetData;
}

/** Кастомный tooltip для Recharts. */
function ChartTooltip({
	active,
	payload,
	label,
}: {
	active?: boolean;
	payload?: Array<{ value: number; name: string; color: string }>;
	label?: string;
}) {
	if (!active || !payload?.length) return null;
	return (
		<div className="rounded-lg border border-border/50 bg-popover px-3 py-2 shadow-md">
			<p className="text-xs font-medium text-t2 mb-1">{label}</p>
			{payload.map((entry) => (
				<p
					key={entry.name}
					className="text-sm font-medium"
					style={{ color: entry.color }}
				>
					{entry.name}: {entry.value}
				</p>
			))}
		</div>
	);
}

/** Виджет burndown-графика. */
export function BurndownChartWidget({ data }: Props) {
	const projectData = isProjectStats(data) ? data : undefined;
	const burndownData = projectData?.burndown_data ?? [];

	if (burndownData.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-t2">
				<TrendingDown className="h-8 w-8 opacity-40" />
				<p className="text-sm">Нет данных для burndown</p>
			</div>
		);
	}

	const chartData = burndownData.map((point) => ({
		label: point.label,
		completed: point.value,
	}));

	return (
		<div
			role="img"
			aria-label={`Burndown: ${chartData.map((p) => `${p.label} — ${p.completed}`).join("; ")}`}
			className="h-full w-full"
		>
			<ResponsiveContainer width="100%" height="100%">
				<AreaChart
					data={chartData}
					margin={{ top: 4, right: 8, bottom: 0, left: -12 }}
				>
					<defs>
						<linearGradient id="burndownGradient" x1="0" y1="0" x2="0" y2="1">
							<stop offset="5%" stopColor="var(--blue)" stopOpacity={0.2} />
							<stop offset="95%" stopColor="var(--blue)" stopOpacity={0} />
						</linearGradient>
					</defs>
					<CartesianGrid
						strokeDasharray="3 3"
						strokeOpacity={0.15}
						vertical={false}
					/>
					<XAxis
						dataKey="label"
						tick={{ fontSize: 10, fill: "hsl(var(--muted-foreground))" }}
						axisLine={false}
						tickLine={false}
					/>
					<YAxis
						tick={{ fontSize: 10, fill: "hsl(var(--muted-foreground))" }}
						axisLine={false}
						tickLine={false}
					/>
					<Tooltip content={<ChartTooltip />} />
					<Area
						type="monotone"
						dataKey="completed"
						stroke="var(--blue)"
						strokeWidth={2}
						fill="url(#burndownGradient)"
						name="Завершено"
					/>
				</AreaChart>
			</ResponsiveContainer>
		</div>
	);
}
