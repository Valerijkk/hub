/**
 * WidgetGrid -- обёртка над react-grid-layout v2 для drag-and-drop виджетов.
 * Поддерживает адаптивные breakpoints, изменение размеров и режим редактирования.
 */
import { useMemo, useCallback } from "react";
import {
  GridLayout,
  useContainerWidth,
  verticalCompactor,
  type LayoutItem as RGLLayoutItem,
} from "react-grid-layout";
import "react-grid-layout/css/styles.css";
import type { LayoutItem } from "@/api/dashboard";
import { useDashboardStore } from "@/stores/dashboard-store";
import { Button } from "@/components/ui/button";

import { MyTasksWidget } from "./widgets/my-tasks";
import { UnreadMessagesWidget } from "./widgets/unread-messages";
import { UpcomingCallsWidget } from "./widgets/upcoming-calls";
import { TaskStatusChartWidget } from "./widgets/task-status-chart";
import { BurndownChartWidget } from "./widgets/burndown-chart";
import { VelocityChartWidget } from "./widgets/velocity-chart";
import { TestTrendsWidget } from "./widgets/test-trends";
import { TeamActivityWidget } from "./widgets/team-activity";
import { SystemHealthWidget } from "./widgets/system-health";
import { StorageUsageWidget } from "./widgets/storage-usage";
import { TimeTrackingWidget } from "./widgets/time-tracking";
import { DeadlinesWidget } from "./widgets/deadlines";

/** Реестр виджетов: тип -> компонент. */
const WIDGET_REGISTRY: Record<
  string,
  React.ComponentType<{ data?: unknown }>
> = {
  "my-tasks": MyTasksWidget,
  "unread-messages": UnreadMessagesWidget,
  "upcoming-calls": UpcomingCallsWidget,
  "task-status-chart": TaskStatusChartWidget,
  "burndown-chart": BurndownChartWidget,
  "velocity-chart": VelocityChartWidget,
  "test-trends": TestTrendsWidget,
  "team-activity": TeamActivityWidget,
  "system-health": SystemHealthWidget,
  "storage-usage": StorageUsageWidget,
  "time-tracking": TimeTrackingWidget,
  deadlines: DeadlinesWidget,
};

/** Названия виджетов на русском. */
const WIDGET_NAMES: Record<string, string> = {
  "my-tasks": "Мои задачи",
  "unread-messages": "Непрочитанные",
  "upcoming-calls": "Предстоящие звонки",
  "task-status-chart": "Статусы задач",
  "burndown-chart": "Burndown",
  "velocity-chart": "Velocity",
  "test-trends": "Тренды тестов",
  "team-activity": "Активность команды",
  "system-health": "Здоровье системы",
  "storage-usage": "Хранилище",
  "time-tracking": "Тайм-трекинг",
  deadlines: "Дедлайны",
};

interface WidgetGridProps {
  /** Элементы раскладки. */
  items: LayoutItem[];
  /** Тип дашборда. */
  dashboardType: string;
  /** ID проекта (для проектных дашбордов). */
  projectId?: string;
  /** Данные для виджетов. */
  data?: unknown;
}

/**
 * Сетка виджетов с поддержкой drag-and-drop и изменения размеров.
 */
export function WidgetGrid({
  items,
  dashboardType,
  projectId,
  data,
}: WidgetGridProps) {
  const editing = useDashboardStore((s) => s.editing);
  const toggleEditing = useDashboardStore((s) => s.toggleEditing);
  const saveLayoutDebounced = useDashboardStore((s) => s.saveLayoutDebounced);
  const setLayout = useDashboardStore((s) => s.setLayout);

  const { width, containerRef } = useContainerWidth({ initialWidth: 1200 });

  const layout = useMemo(
    () =>
      items.map((item) => ({
        i: item.i,
        x: item.x,
        y: item.y,
        w: item.w,
        h: item.h,
      })),
    [items]
  );

  const handleLayoutChange = useCallback(
    (newLayout: readonly RGLLayoutItem[]) => {
      const updatedItems: LayoutItem[] = newLayout.map((l) => {
        const original = items.find((item) => item.i === l.i);
        return {
          i: l.i,
          x: l.x,
          y: l.y,
          w: l.w,
          h: l.h,
          widget_type: original?.widget_type ?? "",
          config: original?.config,
        };
      });

      if (editing) {
        saveLayoutDebounced(dashboardType, updatedItems, projectId);
      } else {
        setLayout(dashboardType, updatedItems);
      }
    },
    [items, editing, dashboardType, projectId, saveLayoutDebounced, setLayout]
  );

  const handleAddWidget = useCallback(
    (widgetType: string) => {
      const newItem: LayoutItem = {
        i: `${widgetType}-${Date.now()}`,
        x: 0,
        y: Infinity,
        w: 6,
        h: 4,
        widget_type: widgetType,
      };
      const updatedItems = [...items, newItem];
      saveLayoutDebounced(dashboardType, updatedItems, projectId);
    },
    [items, dashboardType, projectId, saveLayoutDebounced]
  );

  const existingTypes = useMemo(
    () => new Set(items.map((item) => item.widget_type)),
    [items]
  );

  return (
    <div ref={containerRef}>
      {/* Панель управления */}
      <div className="mb-4 flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={toggleEditing}>
          {editing ? "Готово" : "Редактировать"}
        </Button>

        {editing && (
          <div className="flex flex-wrap gap-1">
            {Object.keys(WIDGET_REGISTRY)
              .filter((type) => !existingTypes.has(type))
              .map((type) => (
                <Button
                  key={type}
                  variant="outline"
                  size="sm"
                  onClick={() => handleAddWidget(type)}
                >
                  + {WIDGET_NAMES[type] ?? type}
                </Button>
              ))}
          </div>
        )}
      </div>

      {/* Сетка виджетов */}
      {width > 0 && (
        <GridLayout
          layout={layout}
          width={width}
          gridConfig={{
            cols: 12,
            rowHeight: 80,
            margin: [16, 16] as const,
            containerPadding: null,
            maxRows: Infinity,
          }}
          dragConfig={{
            enabled: editing,
            handle: ".widget-drag-handle",
            bounded: false,
          }}
          resizeConfig={{
            enabled: editing,
            handles: ["se"] as const,
          }}
          compactor={verticalCompactor}
          onLayoutChange={handleLayoutChange}
        >
          {items.map((item) => {
            const WidgetComponent = WIDGET_REGISTRY[item.widget_type];
            if (!WidgetComponent) return <div key={item.i} />;

            return (
              <div
                key={item.i}
                className="rounded-lg border bg-card text-card-foreground shadow-sm overflow-hidden"
              >
                {/* Заголовок виджета */}
                <div className="widget-drag-handle flex items-center justify-between border-b px-4 py-2 cursor-move">
                  <span className="text-sm font-medium">
                    {WIDGET_NAMES[item.widget_type] ?? item.widget_type}
                  </span>
                  {editing && (
                    <span className="text-xs text-muted-foreground">
                      :::
                    </span>
                  )}
                </div>
                {/* Содержимое виджета */}
                <div className="p-4 overflow-auto h-[calc(100%-40px)]">
                  <WidgetComponent data={data} />
                </div>
              </div>
            );
          })}
        </GridLayout>
      )}
    </div>
  );
}
