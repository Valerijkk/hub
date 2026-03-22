/**
 * WidgetGrid -- wrapper over react-grid-layout v2 for drag-and-drop widgets.
 * Supports adaptive breakpoints, resizing and edit mode.
 */
import {
	CalendarClock,
	ChartBar,
	Clock,
	Ellipsis,
	GripVertical,
	HardDrive,
	Layers,
	Mail,
	Phone,
	Server,
	SquareCheck,
	TestTube,
	TrendingUp,
	Users,
	type LucideIcon,
} from "lucide-react";
import { useCallback, useMemo } from "react";
import {
	GridLayout,
	type LayoutItem as RGLLayoutItem,
	useContainerWidth,
	verticalCompactor,
} from "react-grid-layout";
import "react-grid-layout/css/styles.css";
import type { LayoutItem } from "@/api/dashboard";
import { Button } from "@/components/ui/button";
import { useDashboardStore } from "@/stores/dashboard-store";
import { BurndownChartWidget } from "./widgets/burndown-chart";
import { DeadlinesWidget } from "./widgets/deadlines";
import { MyTasksWidget } from "./widgets/my-tasks";
import { StorageUsageWidget } from "./widgets/storage-usage";
import { SystemHealthWidget } from "./widgets/system-health";
import { TaskStatusChartWidget } from "./widgets/task-status-chart";
import { TeamActivityWidget } from "./widgets/team-activity";
import { TestTrendsWidget } from "./widgets/test-trends";
import { TimeTrackingWidget } from "./widgets/time-tracking";
import { UnreadMessagesWidget } from "./widgets/unread-messages";
import { UpcomingCallsWidget } from "./widgets/upcoming-calls";
import { VelocityChartWidget } from "./widgets/velocity-chart";

/** Widget registry: type -> component. */
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

/** Widget names in Russian. */
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

/** Widget icons. */
const WIDGET_ICONS: Record<string, LucideIcon> = {
	"my-tasks": SquareCheck,
	"unread-messages": Mail,
	"upcoming-calls": Phone,
	"task-status-chart": Layers,
	"burndown-chart": TrendingUp,
	"velocity-chart": ChartBar,
	"test-trends": TestTube,
	"team-activity": Users,
	"system-health": Server,
	"storage-usage": HardDrive,
	"time-tracking": Clock,
	deadlines: CalendarClock,
};

/** Accent color per widget (top border gradient). */
const WIDGET_ACCENT: Record<string, string> = {
	"my-tasks": "from-blue-500 to-blue-600",
	"unread-messages": "from-amber-500 to-orange-500",
	"upcoming-calls": "from-violet-500 to-purple-600",
	"task-status-chart": "from-indigo-500 to-blue-500",
	"burndown-chart": "from-sky-500 to-cyan-500",
	"velocity-chart": "from-emerald-500 to-green-500",
	"test-trends": "from-purple-500 to-violet-600",
	"team-activity": "from-teal-500 to-emerald-500",
	"system-health": "from-green-500 to-emerald-500",
	"storage-usage": "from-slate-500 to-zinc-600",
	"time-tracking": "from-rose-500 to-pink-500",
	deadlines: "from-red-500 to-orange-500",
};

/** WidgetGrid component props. */
interface WidgetGridProps {
	/** Layout items. */
	items: LayoutItem[];
	/** Dashboard type. */
	dashboardType: string;
	/** Project ID (for project dashboards). */
	projectId?: string;
	/** Data for widgets. */
	data?: unknown;
}

/**
 * Widget grid with drag-and-drop and resize support.
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
		[items],
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
		[items, editing, dashboardType, projectId, saveLayoutDebounced, setLayout],
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
		[items, dashboardType, projectId, saveLayoutDebounced],
	);

	const existingTypes = useMemo(
		() => new Set(items.map((item) => item.widget_type)),
		[items],
	);

	return (
		<div ref={containerRef}>
			{/* Control panel */}
			<div className="mb-4 flex items-center gap-2">
				<Button
					variant={editing ? "default" : "outline"}
					size="sm"
					onClick={toggleEditing}
					className="transition-all duration-200"
				>
					{editing ? "Готово" : "Редактировать"}
				</Button>

				{editing && (
					<div className="flex flex-wrap gap-1.5 animate-fade-in">
						{Object.keys(WIDGET_REGISTRY)
							.filter((type) => !existingTypes.has(type))
							.map((type) => {
								const Icon = WIDGET_ICONS[type];
								return (
									<Button
										key={type}
										variant="outline"
										size="sm"
										onClick={() => handleAddWidget(type)}
										className="gap-1.5 transition-all duration-200 hover:bg-accent"
									>
										{Icon && <Icon className="h-3.5 w-3.5" />}
										{WIDGET_NAMES[type] ?? type}
									</Button>
								);
							})}
					</div>
				)}
			</div>

			{/* Widget grid */}
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

						const Icon = WIDGET_ICONS[item.widget_type];
						const accent =
							WIDGET_ACCENT[item.widget_type] ?? "from-gray-500 to-gray-600";

						return (
							<div
								key={item.i}
								className={`
									group/widget relative rounded-xl bg-card text-card-foreground
									ring-1 ring-border/50 shadow-sm overflow-hidden
									transition-all duration-200 ease-out
									hover:shadow-md hover:-translate-y-px
									${editing ? "ring-dashed ring-2 ring-primary/30" : ""}
								`}
							>
								{/* Accent top border */}
								<div
									className={`absolute inset-x-0 top-0 h-0.5 bg-gradient-to-r ${accent} opacity-60 group-hover/widget:opacity-100 transition-opacity duration-200`}
								/>

								{/* Widget header */}
								<div className="widget-drag-handle flex items-center justify-between border-b border-border/50 px-4 py-2.5 cursor-move">
									<div className="flex items-center gap-2">
										{Icon && (
											<Icon className="h-4 w-4 text-muted-foreground" />
										)}
										<span className="text-sm font-medium">
											{WIDGET_NAMES[item.widget_type] ?? item.widget_type}
										</span>
									</div>
									<div className="flex items-center gap-1">
										{editing && (
											<GripVertical className="h-4 w-4 text-muted-foreground/50" />
										)}
										<button
											type="button"
											className="rounded-md p-0.5 text-muted-foreground/50 opacity-0 group-hover/widget:opacity-100 hover:text-foreground hover:bg-accent transition-all duration-150"
										>
											<Ellipsis className="h-4 w-4" />
										</button>
									</div>
								</div>

								{/* Widget content */}
								<div className="p-4 overflow-auto h-[calc(100%-44px)]">
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
