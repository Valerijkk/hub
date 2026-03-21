/**
 * Виджет "Velocity" -- столбчатая диаграмма завершённых задач по неделям.
 */
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
} from "recharts";
import type { ProjectStats } from "@/api/dashboard";

interface Props {
  data?: unknown;
}

/** Виджет velocity-графика. */
export function VelocityChartWidget({ data }: Props) {
  const projectData = data as ProjectStats | undefined;
  const velocityData = projectData?.velocity_data ?? [];

  if (velocityData.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">Нет данных velocity</p>
    );
  }

  const chartData = velocityData.map((point) => ({
    week: point.label,
    completed: point.value,
  }));

  return (
    <ResponsiveContainer width="100%" height="100%">
      <BarChart data={chartData}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="week" tick={{ fontSize: 10 }} />
        <YAxis tick={{ fontSize: 10 }} />
        <Tooltip />
        <Bar dataKey="completed" fill="#10b981" name="Завершено" radius={[4, 4, 0, 0]} />
      </BarChart>
    </ResponsiveContainer>
  );
}
