/**
 * Виджет «Статусы задач» — donut по распределению статусов.
 */
import { Layers } from "lucide-react";
import {
	Cell,
	Legend,
	Pie,
	PieChart,
	ResponsiveContainer,
	Tooltip,
} from "recharts";
import {
	type DashboardWidgetData,
	isProjectStats,
} from "@/components/dashboard/widget-data";

/** Цвета статусов. */
const STATUS_COLORS = [
	"var(--blue)",
	"var(--green)",
	"var(--amber)",
	"var(--red)",
	"var(--purple)",
	"#6366f1",
	"#ec4899",
	"#14b8a6",
];

interface Props {
	data?: DashboardWidgetData;
}

/** Кастомный tooltip. */
function ChartTooltip({
	active,
	payload,
}: {
	active?: boolean;
	payload?: Array<{ name: string; value: number; payload: { fill: string } }>;
}) {
	if (!active || !payload?.length) return null;
	const entry = payload?.[0];
	if (!entry) return null;
	return (
		<div className="rounded-lg border border-border/50 bg-popover px-3 py-2 shadow-md">
			<div className="flex items-center gap-2">
				<span
					className="h-2.5 w-2.5 rounded-full"
					style={{ backgroundColor: entry.payload?.fill }}
				/>
				<span className="text-sm font-medium">{entry.name}</span>
			</div>
			<p className="text-xs text-t2 mt-0.5">{entry.value} задач</p>
		</div>
	);
}

/** Виджет диаграммы статусов задач. */
export function TaskStatusChartWidget({ data }: Props) {
	const projectData = isProjectStats(data) ? data : undefined;
	const statusMap = projectData?.tasks_by_status ?? {};

	const chartData = Object.entries(statusMap).map(([name, value]) => ({
		name,
		value,
	}));

	if (chartData.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-t2">
				<Layers className="h-8 w-8 opacity-40" />
				<p className="text-sm">Нет данных о задачах</p>
			</div>
		);
	}

	return (
		<div
			role="img"
			aria-label={`Статусы задач: ${chartData.map((p) => `${p.name} — ${p.value}`).join("; ")}`}
			className="h-full w-full"
		>
			<ResponsiveContainer width="100%" height="100%">
				<PieChart>
					<Pie
						data={chartData}
						cx="50%"
						cy="50%"
						innerRadius={40}
						outerRadius={70}
						dataKey="value"
						nameKey="name"
						paddingAngle={3}
						strokeWidth={0}
					>
						{chartData.map((entry, index) => (
							<Cell
								key={`cell-${entry.name}`}
								fill={STATUS_COLORS[index % STATUS_COLORS.length]}
							/>
						))}
					</Pie>
					<Tooltip content={<ChartTooltip />} />
					<Legend
						iconType="circle"
						iconSize={8}
						formatter={(value) => (
							<span className="text-xs text-t2">{value}</span>
						)}
					/>
				</PieChart>
			</ResponsiveContainer>
		</div>
	);
}
