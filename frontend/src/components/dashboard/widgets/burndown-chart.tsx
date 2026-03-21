/**
 * Виджет "Burndown" -- график завершённых задач по дням (ideal vs actual).
 */
import {
  AreaChart,
  Area,
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

/** Виджет burndown-графика. */
export function BurndownChartWidget({ data }: Props) {
  const projectData = data as ProjectStats | undefined;
  const burndownData = projectData?.burndown_data ?? [];

  if (burndownData.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">Нет данных для burndown</p>
    );
  }

  const chartData = burndownData.map((point) => ({
    label: point.label,
    completed: point.value,
  }));

  return (
    <ResponsiveContainer width="100%" height="100%">
      <AreaChart data={chartData}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="label" tick={{ fontSize: 10 }} />
        <YAxis tick={{ fontSize: 10 }} />
        <Tooltip />
        <Area
          type="monotone"
          dataKey="completed"
          stroke="#3b82f6"
          fill="#3b82f680"
          name="Завершено"
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}
