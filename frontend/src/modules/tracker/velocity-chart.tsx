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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
			<Card>
				<CardHeader>
					<CardTitle>Velocity</CardTitle>
				</CardHeader>
				<CardContent>
					<p className="text-muted-foreground">
						Нет завершённых спринтов для отображения
					</p>
				</CardContent>
			</Card>
		);
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle>Velocity</CardTitle>
			</CardHeader>
			<CardContent>
				<ResponsiveContainer width="100%" height={300}>
					<BarChart
						data={velocity}
						margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
					>
						<CartesianGrid strokeDasharray="3 3" />
						<XAxis dataKey="sprint_name" tick={{ fontSize: 12 }} />
						<YAxis />
						<Tooltip />
						<Legend />
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
			</CardContent>
		</Card>
	);
}
