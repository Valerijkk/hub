import { useEffect } from "react";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { useSprintStore } from "@/stores/sprint-store";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

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
  const { burndown, fetchBurndown } = useSprintStore();

  useEffect(() => {
    fetchBurndown(projectId, sprintId);
  }, [projectId, sprintId, fetchBurndown]);

  if (burndown.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Burndown</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">Нет данных для отображения</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Burndown</CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart
            data={burndown}
            margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
          >
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis
              dataKey="date"
              tick={{ fontSize: 12 }}
              tickFormatter={(v: string) => {
                const d = new Date(v);
                return `${d.getDate()}.${d.getMonth() + 1}`;
              }}
            />
            <YAxis />
            <Tooltip
              labelFormatter={(label) =>
                new Date(String(label)).toLocaleDateString()
              }
            />
            <Legend />
            <Area
              type="monotone"
              dataKey="ideal"
              stroke="#94a3b8"
              fill="#e2e8f0"
              strokeDasharray="5 5"
              name="Идеальный"
            />
            <Area
              type="monotone"
              dataKey="remaining"
              stroke="#3b82f6"
              fill="#93c5fd"
              name="Фактический"
            />
          </AreaChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
