/**
 * WidgetGrid -- wrapper over react-grid-layout v2 for drag-and-drop widgets.
 * Widgets are always draggable and resizable with per-widget size constraints.
 */
import {
	CalendarClock,
	ChartBar,
	Ellipsis,
	EyeOff,
	GripVertical,
	HardDrive,
	Layers,
	Mail,
	Phone,
	RefreshCw,
	Server,
	SquareCheck,
	TestTube,
	TrendingUp,
	Users,
	type LucideIcon,
} from "lucide-react";
import { useCallback, useMemo, useState } from "react";
import {
	GridLayout,
	type LayoutItem as RGLLayoutItem,
	useContainerWidth,
	verticalCompactor,
} from "react-grid-layout";
import "react-grid-layout/css/styles.css";
import type { LayoutItem } from "@/api/dashboard";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useDashboardStore } from "@/stores/dashboard-store";
import { BurndownChartWidget } from "./widgets/burndown-chart";
import { DeadlinesWidget } from "./widgets/deadlines";
import { MyTasksWidget } from "./widgets/my-tasks";
import { StorageUsageWidget } from "./widgets/storage-usage";
import { SystemHealthWidget } from "./widgets/system-health";
import { TaskStatusChartWidget } from "./widgets/task-status-chart";
import { TeamActivityWidget } from "./widgets/team-activity";
import { TestTrendsWidget } from "./widgets/test-trends";
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
	deadlines: "from-red-500 to-orange-500",
};

/** Per-widget min/max size constraints. */
const WIDGET_SIZE_CONSTRAINTS: Record<
	string,
	{ minW: number; maxW: number; minH: number; maxH: number }
> = {
	"my-tasks": { minW: 3, maxW: 12, minH: 3, maxH: 6 },
	"unread-messages": { minW: 2, maxW: 6, minH: 2, maxH: 4 },
	"upcoming-calls": { minW: 2, maxW: 6, minH: 2, maxH: 4 },
	"task-status-chart": { minW: 3, maxW: 6, minH: 3, maxH: 6 },
	"burndown-chart": { minW: 4, maxW: 12, minH: 3, maxH: 6 },
	"velocity-chart": { minW: 4, maxW: 12, minH: 3, maxH: 6 },
	"test-trends": { minW: 4, maxW: 12, minH: 3, maxH: 6 },
	"team-activity": { minW: 3, maxW: 12, minH: 3, maxH: 6 },
	"system-health": { minW: 2, maxW: 6, minH: 2, maxH: 4 },
	"storage-usage": { minW: 2, maxW: 6, minH: 2, maxH: 4 },
	deadlines: { minW: 3, maxW: 12, minH: 2, maxH: 6 },
};

/** Default size constraints for unknown widget types. */
const DEFAULT_CONSTRAINTS = { minW: 2, maxW: 6, minH: 2, maxH: 6 };

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
 * Widgets are always draggable -- no edit mode required.
 */
export function WidgetGrid({
	items,
	dashboardType,
	projectId,
	data,
}: WidgetGridProps) {
	const saveLayoutDebounced = useDashboardStore((s) => s.saveLayoutDebounced);

	const { width, containerRef } = useContainerWidth({ initialWidth: 1200 });

	/** Key that increments to force re-render of individual widget content. */
	const [refreshKeys, setRefreshKeys] = useState<Record<string, number>>({});

	const layout = useMemo(
		() =>
			items.map((item) => {
				const constraints =
					WIDGET_SIZE_CONSTRAINTS[item.widget_type] ?? DEFAULT_CONSTRAINTS;
				return {
					i: item.i,
					x: item.x,
					y: item.y,
					w: item.w,
					h: item.h,
					minW: constraints.minW,
					maxW: constraints.maxW,
					minH: constraints.minH,
					maxH: constraints.maxH,
				};
			}),
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

			saveLayoutDebounced(dashboardType, updatedItems, projectId);
		},
		[items, dashboardType, projectId, saveLayoutDebounced],
	);

	const handleHideWidget = useCallback(
		(itemId: string) => {
			const updatedItems = items.filter((item) => item.i !== itemId);
			saveLayoutDebounced(dashboardType, updatedItems, projectId);
		},
		[items, dashboardType, projectId, saveLayoutDebounced],
	);

	const handleRefreshWidget = useCallback((itemId: string) => {
		setRefreshKeys((prev) => ({
			...prev,
			[itemId]: (prev[itemId] ?? 0) + 1,
		}));
	}, []);

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

	/** Widget types not currently on the grid. */
	const hiddenWidgetTypes = useMemo(
		() => Object.keys(WIDGET_REGISTRY).filter((type) => !existingTypes.has(type)),
		[existingTypes],
	);

	return (
		<div ref={containerRef}>
			{/* Add hidden widgets back */}
			{hiddenWidgetTypes.length > 0 && (
				<div className="mb-4">
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<button
								type="button"
								className="inline-flex items-center gap-1.5 rounded-md border border-border/60 bg-card px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
							>
								<Layers className="h-3.5 w-3.5" />
								Добавить виджет
							</button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="start">
							{hiddenWidgetTypes.map((type) => {
								const Icon = WIDGET_ICONS[type];
								return (
									<DropdownMenuItem
										key={type}
										onClick={() => handleAddWidget(type)}
									>
										{Icon && <Icon className="h-4 w-4" />}
										{WIDGET_NAMES[type] ?? type}
									</DropdownMenuItem>
								);
							})}
						</DropdownMenuContent>
					</DropdownMenu>
				</div>
			)}

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
						enabled: true,
						handle: ".widget-drag-handle",
						bounded: false,
					}}
					resizeConfig={{
						enabled: true,
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
								className="group/widget widget-card relative rounded-2xl bg-card text-card-foreground border border-white/[0.06] shadow-sm overflow-hidden"
							>
								{/* Accent top border */}
								<div
									className={`absolute inset-x-0 top-0 h-0.5 bg-gradient-to-r ${accent} opacity-60 group-hover/widget:opacity-100 transition-opacity duration-200`}
								/>

								{/* Widget header */}
								<div className="widget-drag-handle flex items-center justify-between border-b border-white/[0.06] px-4 py-2.5 cursor-move">
									<div className="flex items-center gap-2">
										<GripVertical className="h-3.5 w-3.5 text-muted-foreground/40" />
										{Icon && (
											<Icon className="h-4 w-4 text-muted-foreground" />
										)}
										<span className="text-[11px] uppercase font-semibold tracking-[0.04em] text-muted-foreground">
											{WIDGET_NAMES[item.widget_type] ?? item.widget_type}
										</span>
									</div>
									<div className="flex items-center gap-1">
										<DropdownMenu>
											<DropdownMenuTrigger asChild>
												<button
													type="button"
													className="rounded-md p-0.5 text-muted-foreground/50 opacity-0 group-hover/widget:opacity-100 hover:text-foreground hover:bg-accent transition-all duration-150"
												>
													<Ellipsis className="h-4 w-4" />
												</button>
											</DropdownMenuTrigger>
											<DropdownMenuContent align="end">
												<DropdownMenuItem
													onClick={() => handleRefreshWidget(item.i)}
												>
													<RefreshCw className="h-4 w-4" />
													Обновить
												</DropdownMenuItem>
												<DropdownMenuItem
													onClick={() => handleHideWidget(item.i)}
												>
													<EyeOff className="h-4 w-4" />
													Скрыть виджет
												</DropdownMenuItem>
											</DropdownMenuContent>
										</DropdownMenu>
									</div>
								</div>

								{/* Widget content */}
								<div className="p-4 overflow-hidden h-[calc(100%-44px)]">
									<WidgetComponent
										key={refreshKeys[item.i] ?? 0}
										data={data}
									/>
								</div>
							</div>
						);
					})}
				</GridLayout>
			)}
		</div>
	);
}
