/**
 * Виджет "Статусы задач" -- кольцевая диаграмма распределения задач по статусам.
 */
import {
	Cell,
	Legend,
	Pie,
	PieChart,
	ResponsiveContainer,
	Tooltip,
} from "recharts";
import type { ProjectStats } from "@/api/dashboard";

/** Цвета для статусов. */
const STATUS_COLORS = [
	"#3b82f6",
	"#10b981",
	"#f59e0b",
	"#ef4444",
	"#8b5cf6",
	"#6366f1",
	"#ec4899",
	"#14b8a6",
];

interface Props {
	data?: unknown;
}

/** Виджет диаграммы статусов задач. */
export function TaskStatusChartWidget({ data }: Props) {
	const projectData = data as ProjectStats | undefined;
	const statusMap = projectData?.tasks_by_status ?? {};

	const chartData = Object.entries(statusMap).map(([name, value]) => ({
		name,
		value,
	}));

	if (chartData.length === 0) {
		return (
			<p className="text-sm text-muted-foreground">Нет данных о задачах</p>
		);
	}

	return (
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
					paddingAngle={2}
				>
					{chartData.map((_, index) => (
						<Cell
							key={`cell-${index}`}
							fill={STATUS_COLORS[index % STATUS_COLORS.length]}
						/>
					))}
				</Pie>
				<Tooltip />
				<Legend />
			</PieChart>
		</ResponsiveContainer>
	);
}
