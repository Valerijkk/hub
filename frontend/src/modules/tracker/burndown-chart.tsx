import { TrendingDown } from "lucide-react";
import { useEffect } from "react";
import {
	Area,
	AreaChart,
	CartesianGrid,
	Legend,
	ResponsiveContainer,
	Tooltip,
	XAxis,
	YAxis,
} from "recharts";
import { useSprintStore } from "@/stores/sprint-store";

/** Пропсы компонента BurndownChart. */
interface BurndownChartProps {
	/** ID проекта. */
	projectId: string;
	/** ID спринта для отображения burndown. */
	sprintId: string;
}

/**
 * BurndownChart отображает график burndown спринта (SCRM-02).
 * Две линии: Ideal (прямая от total к 0) и Remaining (фактическая).
 */
export function BurndownChart({ projectId, sprintId }: BurndownChartProps) {
	const { burndown, error, fetchBurndown } = useSprintStore();

	useEffect(() => {
		void fetchBurndown(projectId, sprintId);
	}, [projectId, sprintId, fetchBurndown]);

	if (error && burndown.length === 0) {
		return (
			<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5">
				<div className="flex items-center gap-2.5 mb-4">
					<div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500/10 to-blue-500/5">
						<TrendingDown className="h-4 w-4 text-blue-400" />
					</div>
					<h2 className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3">
						Burndown
					</h2>
				</div>
				<div role="alert" className="space-y-2">
					<p className="text-[13px] text-[var(--red)]">{error}</p>
					<button
						type="button"
						onClick={() => void fetchBurndown(projectId, sprintId)}
						className="text-[13px] font-semibold text-t2 hover:text-t1 hover:underline"
					>
						Попробовать снова
					</button>
				</div>
			</div>
		);
	}

	if (burndown.length === 0) {
		return (
			<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5">
				<div className="flex items-center gap-2.5 mb-4">
					<div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500/10 to-blue-500/5">
						<TrendingDown className="h-4 w-4 text-blue-400" />
					</div>
					<h2 className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3">
						Burndown
					</h2>
				</div>
				<p className="text-[13px] text-t2">Нет данных для отображения</p>
			</div>
		);
	}

	return (
		<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5">
			<div className="flex items-center gap-2.5 mb-4">
				<div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500/10 to-blue-500/5">
					<TrendingDown className="h-4 w-4 text-blue-400" />
				</div>
				<h2 className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3">
					Burndown
				</h2>
			</div>
			<ResponsiveContainer width="100%" height={300}>
				<AreaChart
					data={burndown}
					margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
				>
					<defs>
						<linearGradient id="burndown-ideal" x1="0" y1="0" x2="0" y2="1">
							<stop offset="5%" stopColor="var(--gray)" stopOpacity={0.15} />
							<stop offset="95%" stopColor="var(--gray)" stopOpacity={0} />
						</linearGradient>
						<linearGradient id="burndown-remaining" x1="0" y1="0" x2="0" y2="1">
							<stop offset="5%" stopColor="var(--blue)" stopOpacity={0.2} />
							<stop offset="95%" stopColor="var(--blue)" stopOpacity={0} />
						</linearGradient>
					</defs>
					<CartesianGrid
						strokeDasharray="3 3"
						stroke="rgba(255,255,255,0.04)"
					/>
					<XAxis
						dataKey="date"
						tick={{ fontSize: 12, fill: "var(--t3)" }}
						tickFormatter={(v: string) => {
							const d = new Date(v);
							return `${d.getDate()}.${d.getMonth() + 1}`;
						}}
						axisLine={{ stroke: "rgba(255,255,255,0.06)" }}
						tickLine={false}
					/>
					<YAxis
						tick={{ fontSize: 12, fill: "var(--t3)" }}
						axisLine={{ stroke: "rgba(255,255,255,0.06)" }}
						tickLine={false}
					/>
					<Tooltip
						labelFormatter={(label) =>
							new Date(String(label)).toLocaleDateString()
						}
						contentStyle={{
							background: "var(--raised)",
							border: "1px solid var(--b2)",
							borderRadius: "12px",
							boxShadow: "0 8px 24px rgba(0,0,0,0.4)",
							fontSize: "13px",
						}}
						itemStyle={{ color: "var(--t2)" }}
						labelStyle={{ color: "var(--t3)", marginBottom: "4px" }}
					/>
					<Legend wrapperStyle={{ fontSize: "12px", color: "var(--t3)" }} />
					<Area
						type="monotone"
						dataKey="ideal"
						stroke="var(--gray)"
						fill="url(#burndown-ideal)"
						strokeDasharray="5 5"
						strokeWidth={1.5}
						name="Идеальный"
					/>
					<Area
						type="monotone"
						dataKey="remaining"
						stroke="var(--blue)"
						fill="url(#burndown-remaining)"
						strokeWidth={2}
						name="Фактический"
					/>
				</AreaChart>
			</ResponsiveContainer>
		</div>
	);
}
