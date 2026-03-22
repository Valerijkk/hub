/**
 * Виджет "Тренды тестов" -- линейный график pass rate по последним прогонам.
 */
import {
	Area,
	AreaChart,
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

/** Виджет трендов тестирования. */
export function TestTrendsWidget({ data }: Props) {
	const projectData = data as ProjectStats | undefined;
	const trends = projectData?.test_trends ?? [];

	if (trends.length === 0) {
		return <p className="text-sm text-muted-foreground">Нет данных о тестах</p>;
	}

	const chartData = trends.map((point) => ({
		run: point.label,
		passRate: Math.round(point.value * 10) / 10,
	}));

	return (
		<ResponsiveContainer width="100%" height="100%">
			<AreaChart data={chartData}>
				<CartesianGrid strokeDasharray="3 3" />
				<XAxis dataKey="run" tick={{ fontSize: 10 }} />
				<YAxis domain={[0, 100]} tick={{ fontSize: 10 }} unit="%" />
				<Tooltip formatter={(value) => `${value}%`} />
				<Area
					type="monotone"
					dataKey="passRate"
					stroke="#8b5cf6"
					fill="#8b5cf680"
					name="Pass Rate"
				/>
			</AreaChart>
		</ResponsiveContainer>
	);
}
