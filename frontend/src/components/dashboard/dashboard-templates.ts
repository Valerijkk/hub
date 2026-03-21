/**
 * Шаблоны раскладок дашбордов: предустановленные конфигурации виджетов.
 * Применяются при первом визите, когда нет сохранённой раскладки.
 */
import type { LayoutItem } from "@/api/dashboard";

/** Персональный дашборд по умолчанию. */
export const personalDefault: LayoutItem[] = [
  { i: "my-tasks-1", x: 0, y: 0, w: 8, h: 4, widget_type: "my-tasks" },
  {
    i: "unread-messages-1",
    x: 8,
    y: 0,
    w: 4,
    h: 2,
    widget_type: "unread-messages",
  },
  {
    i: "upcoming-calls-1",
    x: 8,
    y: 2,
    w: 4,
    h: 2,
    widget_type: "upcoming-calls",
  },
  { i: "deadlines-1", x: 0, y: 4, w: 6, h: 3, widget_type: "deadlines" },
  {
    i: "time-tracking-1",
    x: 6,
    y: 4,
    w: 6,
    h: 3,
    widget_type: "time-tracking",
  },
];

/** Шаблон проектного дашборда для разработчика. */
export const projectDev: LayoutItem[] = [
  {
    i: "task-status-chart-1",
    x: 0,
    y: 0,
    w: 4,
    h: 4,
    widget_type: "task-status-chart",
  },
  {
    i: "burndown-chart-1",
    x: 4,
    y: 0,
    w: 8,
    h: 4,
    widget_type: "burndown-chart",
  },
  {
    i: "velocity-chart-1",
    x: 0,
    y: 4,
    w: 6,
    h: 4,
    widget_type: "velocity-chart",
  },
  {
    i: "team-activity-1",
    x: 6,
    y: 4,
    w: 6,
    h: 4,
    widget_type: "team-activity",
  },
  { i: "deadlines-1", x: 0, y: 8, w: 12, h: 3, widget_type: "deadlines" },
];

/** Шаблон проектного дашборда для QA. */
export const projectQA: LayoutItem[] = [
  {
    i: "test-trends-1",
    x: 0,
    y: 0,
    w: 8,
    h: 4,
    widget_type: "test-trends",
  },
  {
    i: "task-status-chart-1",
    x: 8,
    y: 0,
    w: 4,
    h: 4,
    widget_type: "task-status-chart",
  },
  {
    i: "team-activity-1",
    x: 0,
    y: 4,
    w: 12,
    h: 4,
    widget_type: "team-activity",
  },
];

/** Шаблон проектного дашборда для менеджера. */
export const projectManager: LayoutItem[] = [
  {
    i: "burndown-chart-1",
    x: 0,
    y: 0,
    w: 6,
    h: 4,
    widget_type: "burndown-chart",
  },
  {
    i: "velocity-chart-1",
    x: 6,
    y: 0,
    w: 6,
    h: 4,
    widget_type: "velocity-chart",
  },
  {
    i: "task-status-chart-1",
    x: 0,
    y: 4,
    w: 4,
    h: 4,
    widget_type: "task-status-chart",
  },
  { i: "deadlines-1", x: 4, y: 4, w: 4, h: 4, widget_type: "deadlines" },
  {
    i: "team-activity-1",
    x: 8,
    y: 4,
    w: 4,
    h: 4,
    widget_type: "team-activity",
  },
];

/** Админский дашборд по умолчанию. */
export const adminDefault: LayoutItem[] = [
  {
    i: "system-health-1",
    x: 0,
    y: 0,
    w: 4,
    h: 3,
    widget_type: "system-health",
  },
  {
    i: "storage-usage-1",
    x: 4,
    y: 0,
    w: 4,
    h: 3,
    widget_type: "storage-usage",
  },
  {
    i: "team-activity-1",
    x: 8,
    y: 0,
    w: 4,
    h: 3,
    widget_type: "team-activity",
  },
];

/** Карта шаблонов по имени. */
export const DASHBOARD_TEMPLATES: Record<string, LayoutItem[]> = {
  personalDefault,
  projectDev,
  projectQA,
  projectManager,
  adminDefault,
};
