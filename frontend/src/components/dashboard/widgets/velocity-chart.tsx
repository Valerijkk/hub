/**
 * Виджет «Velocity» — завершённые задачи по неделям.
 */
import { ChartBar } from "lucide-react";
import {
	Bar,
	BarChart,
	CartesianGrid,
	ResponsiveContainer,
	Tooltip,
	XAxis,
	YAxis,
} from "recharts";
import type { ProjectStats } from "@/api/dashboard";

interface Props {
	data?: unknown;
}

/** Кастомный tooltip. */
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

/** Виджет velocity-графика. */
export function VelocityChartWidget({ data }: Props) {
	const projectData = data as ProjectStats | undefined;
	const velocityData = projectData?.velocity_data ?? [];

	if (velocityData.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-t2">
				<ChartBar className="h-8 w-8 opacity-40" />
				<p className="text-sm">Нет данных velocity</p>
			</div>
		);
	}

	const chartData = velocityData.map((point) => ({
		week: point.label,
		completed: point.value,
	}));

	return (
		<div
			role="img"
			aria-label={`Velocity: ${chartData.map((p) => `${p.week} — ${p.completed}`).join("; ")}`}
			className="h-full w-full"
		>
			<ResponsiveContainer width="100%" height="100%">
				<BarChart
					data={chartData}
					margin={{ top: 4, right: 8, bottom: 0, left: -12 }}
				>
					<defs>
						<linearGradient id="velocityGradient" x1="0" y1="0" x2="0" y2="1">
							<stop offset="5%" stopColor="var(--green)" stopOpacity={0.8} />
							<stop offset="95%" stopColor="var(--green)" stopOpacity={0.4} />
						</linearGradient>
					</defs>
					<CartesianGrid
						strokeDasharray="3 3"
						strokeOpacity={0.15}
						vertical={false}
					/>
					<XAxis
						dataKey="week"
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
					<Bar
						dataKey="completed"
						fill="url(#velocityGradient)"
						name="Завершено"
						radius={[6, 6, 0, 0]}
					/>
				</BarChart>
			</ResponsiveContainer>
		</div>
	);
}
