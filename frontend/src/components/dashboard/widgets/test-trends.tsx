/**
 * Widget "Test Trends" -- line chart of pass rate across recent runs.
 */
import { TestTube } from "lucide-react";
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

/** Custom tooltip. */
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
				<p key={entry.name} className="text-sm font-medium" style={{ color: entry.color }}>
					{entry.name}: {entry.value}%
				</p>
			))}
		</div>
	);
}

/** Test trends widget. */
export function TestTrendsWidget({ data }: Props) {
	const projectData = data as ProjectStats | undefined;
	const trends = projectData?.test_trends ?? [];

	if (trends.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-t2">
				<TestTube className="h-8 w-8 opacity-40" />
				<p className="text-sm">Нет данных о тестах</p>
			</div>
		);
	}

	const chartData = trends.map((point) => ({
		run: point.label,
		passRate: Math.round(point.value * 10) / 10,
	}));

	return (
		<ResponsiveContainer width="100%" height="100%">
			<AreaChart data={chartData} margin={{ top: 4, right: 8, bottom: 0, left: -12 }}>
				<defs>
					<linearGradient id="testTrendsGradient" x1="0" y1="0" x2="0" y2="1">
						<stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.2} />
						<stop offset="95%" stopColor="#8b5cf6" stopOpacity={0} />
					</linearGradient>
				</defs>
				<CartesianGrid strokeDasharray="3 3" strokeOpacity={0.15} vertical={false} />
				<XAxis
					dataKey="run"
					tick={{ fontSize: 10, fill: "hsl(var(--muted-foreground))" }}
					axisLine={false}
					tickLine={false}
				/>
				<YAxis
					domain={[0, 100]}
					tick={{ fontSize: 10, fill: "hsl(var(--muted-foreground))" }}
					axisLine={false}
					tickLine={false}
					unit="%"
				/>
				<Tooltip content={<ChartTooltip />} />
				<Area
					type="monotone"
					dataKey="passRate"
					stroke="#8b5cf6"
					strokeWidth={2}
					fill="url(#testTrendsGradient)"
					name="Pass Rate"
				/>
			</AreaChart>
		</ResponsiveContainer>
	);
}
