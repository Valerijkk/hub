import { TrendingUp } from "lucide-react";
import { useEffect } from "react";
import {
	Bar,
	BarChart,
	CartesianGrid,
	Legend,
	ResponsiveContainer,
	Tooltip,
	XAxis,
	YAxis,
} from "recharts";
import { useSprintStore } from "@/stores/sprint-store";

/** Пропсы компонента VelocityChart. */
interface VelocityChartProps {
	/** ID проекта. */
	projectId: string;
}

/**
 * VelocityChart отображает график velocity проекта (SCRM-02).
 * Для каждого завершённого спринта два бара: Committed и Completed.
 */
export function VelocityChart({ projectId }: VelocityChartProps) {
	const { velocity, fetchVelocity } = useSprintStore();

	useEffect(() => {
		fetchVelocity(projectId);
	}, [projectId, fetchVelocity]);

	if (velocity.length === 0) {
		return (
			<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5">
				<div className="flex items-center gap-2.5 mb-4">
					<div className="flex h-8 w-8 items-center justify-center rounded-[var(--r2)] bg-gradient-to-br from-emerald-500/10 to-emerald-500/5">
						<TrendingUp className="h-4 w-4 text-emerald-400" />
					</div>
					<h2 className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3">
						Velocity
					</h2>
				</div>
				<p className="text-[13px] text-t2">
					Нет завершённых спринтов для отображения
				</p>
			</div>
		);
	}

	return (
		<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5">
			<div className="flex items-center gap-2.5 mb-4">
				<div className="flex h-8 w-8 items-center justify-center rounded-[var(--r2)] bg-gradient-to-br from-emerald-500/10 to-emerald-500/5">
					<TrendingUp className="h-4 w-4 text-emerald-400" />
				</div>
				<h2 className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3">
					Velocity
				</h2>
			</div>
			<ResponsiveContainer width="100%" height={300}>
				<BarChart
					data={velocity}
					margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
				>
					<CartesianGrid
						strokeDasharray="3 3"
						stroke="rgba(255,255,255,0.04)"
					/>
					<XAxis
						dataKey="sprint_name"
						tick={{ fontSize: 12, fill: "#71717a" }}
						axisLine={{ stroke: "rgba(255,255,255,0.06)" }}
						tickLine={false}
					/>
					<YAxis
						tick={{ fontSize: 12, fill: "#71717a" }}
						axisLine={{ stroke: "rgba(255,255,255,0.06)" }}
						tickLine={false}
					/>
					<Tooltip
						contentStyle={{
							background: "rgba(24, 24, 27, 0.95)",
							border: "1px solid rgba(255,255,255,0.08)",
							borderRadius: "12px",
							boxShadow: "0 8px 24px rgba(0,0,0,0.4)",
							fontSize: "13px",
						}}
						itemStyle={{ color: "#d4d4d8" }}
						labelStyle={{ color: "#a1a1aa", marginBottom: "4px" }}
					/>
					<Legend
						wrapperStyle={{ fontSize: "12px", color: "#71717a" }}
					/>
					<Bar
						dataKey="committed"
						fill="#94a3b8"
						name="Запланировано"
						radius={[4, 4, 0, 0]}
					/>
					<Bar
						dataKey="completed"
						fill="#3b82f6"
						name="Завершено"
						radius={[4, 4, 0, 0]}
					/>
				</BarChart>
			</ResponsiveContainer>
		</div>
	);
}
