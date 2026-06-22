/**
 * WidgetGrid -- обёртка над react-grid-layout v2 для drag-and-drop виджетов.
 * Виджеты всегда перетаскиваемы и ресайзятся с индивидуальными ограничениями размеров.
 */
import {
	CalendarClock,
	ChartBar,
	Ellipsis,
	EyeOff,
	GripVertical,
	HardDrive,
	Layers,
	type LucideIcon,
	Mail,
	Phone,
	RefreshCw,
	Server,
	SquareCheck,
	TestTube,
	TrendingUp,
	Users,
} from "lucide-react";
import { lazy, Suspense, useCallback, useMemo, useState } from "react";
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
// Лёгкие виджеты (без recharts) импортируются статически.
import { DeadlinesWidget } from "./widgets/deadlines";
import { MyTasksWidget } from "./widgets/my-tasks";
import { StorageUsageWidget } from "./widgets/storage-usage";
import { SystemHealthWidget } from "./widgets/system-health";
import { TeamActivityWidget } from "./widgets/team-activity";
import { UnreadMessagesWidget } from "./widgets/unread-messages";
import { UpcomingCallsWidget } from "./widgets/upcoming-calls";

// Тяжёлые виджеты на recharts (~390 KB) выносим в отдельные chunks через React.lazy,
// чтобы исключить vendor-charts из main-бандла. Грузятся только когда виджет реально на сетке.
const BurndownChartWidget = lazy(() =>
	import("./widgets/burndown-chart").then((m) => ({
		default: m.BurndownChartWidget,
	})),
);
const TaskStatusChartWidget = lazy(() =>
	import("./widgets/task-status-chart").then((m) => ({
		default: m.TaskStatusChartWidget,
	})),
);
const TestTrendsWidget = lazy(() =>
	import("./widgets/test-trends").then((m) => ({
		default: m.TestTrendsWidget,
	})),
);
const VelocityChartWidget = lazy(() =>
	import("./widgets/velocity-chart").then((m) => ({
		default: m.VelocityChartWidget,
	})),
);

/** Скелет-заглушка для тяжёлого виджета во время загрузки чанка. */
function ChartWidgetSkeleton() {
	return (
		<div className="flex flex-col items-center justify-center h-full gap-2 text-t3">
			<div className="h-full w-full rounded animate-shimmer opacity-50" />
		</div>
	);
}

/** Определение виджета: компонент, название, иконка, акцентный градиент, ограничения размеров. */
interface WidgetDefinition {
	component: React.ComponentType<{ data?: unknown }>;
	name: string;
	icon: LucideIcon;
	accent: string;
	constraints: { minW: number; maxW: number; minH: number; maxH: number };
}

const WIDGET_CONFIG: Record<string, WidgetDefinition> = {
	"my-tasks": {
		component: MyTasksWidget,
		name: "Мои задачи",
		icon: SquareCheck,
		accent: "from-blue-500 to-blue-600",
		constraints: { minW: 3, maxW: 12, minH: 3, maxH: 6 },
	},
	"unread-messages": {
		component: UnreadMessagesWidget,
		name: "Непрочитанные",
		icon: Mail,
		accent: "from-amber-500 to-orange-500",
		constraints: { minW: 2, maxW: 6, minH: 2, maxH: 4 },
	},
	"upcoming-calls": {
		component: UpcomingCallsWidget,
		name: "Предстоящие звонки",
		icon: Phone,
		accent: "from-violet-500 to-purple-600",
		constraints: { minW: 2, maxW: 6, minH: 2, maxH: 4 },
	},
	"task-status-chart": {
		component: TaskStatusChartWidget,
		name: "Статусы задач",
		icon: Layers,
		accent: "from-indigo-500 to-blue-500",
		constraints: { minW: 3, maxW: 6, minH: 3, maxH: 6 },
	},
	"burndown-chart": {
		component: BurndownChartWidget,
		name: "Burndown",
		icon: TrendingUp,
		accent: "from-sky-500 to-cyan-500",
		constraints: { minW: 4, maxW: 12, minH: 3, maxH: 6 },
	},
	"velocity-chart": {
		component: VelocityChartWidget,
		name: "Velocity",
		icon: ChartBar,
		accent: "from-emerald-500 to-green-500",
		constraints: { minW: 4, maxW: 12, minH: 3, maxH: 6 },
	},
	"test-trends": {
		component: TestTrendsWidget,
		name: "Тренды тестов",
		icon: TestTube,
		accent: "from-purple-500 to-violet-600",
		constraints: { minW: 4, maxW: 12, minH: 3, maxH: 6 },
	},
	"team-activity": {
		component: TeamActivityWidget,
		name: "Активность команды",
		icon: Users,
		accent: "from-teal-500 to-emerald-500",
		constraints: { minW: 3, maxW: 12, minH: 3, maxH: 6 },
	},
	"system-health": {
		component: SystemHealthWidget,
		name: "Здоровье системы",
		icon: Server,
		accent: "from-green-500 to-emerald-500",
		constraints: { minW: 2, maxW: 6, minH: 2, maxH: 4 },
	},
	"storage-usage": {
		component: StorageUsageWidget,
		name: "Хранилище",
		icon: HardDrive,
		accent: "from-slate-500 to-zinc-600",
		constraints: { minW: 2, maxW: 6, minH: 2, maxH: 4 },
	},
	deadlines: {
		component: DeadlinesWidget,
		name: "Дедлайны",
		icon: CalendarClock,
		accent: "from-red-500 to-orange-500",
		constraints: { minW: 3, maxW: 12, minH: 2, maxH: 6 },
	},
};

/** Ограничения размеров по умолчанию для неизвестных типов виджетов. */
const DEFAULT_CONSTRAINTS = { minW: 2, maxW: 6, minH: 2, maxH: 6 };

/** Пропсы компонента WidgetGrid. */
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
 * Сетка виджетов с поддержкой drag-and-drop и ресайза.
 * Виджеты всегда перетаскиваемы -- режим редактирования не требуется.
 */
export function WidgetGrid({
	items,
	dashboardType,
	projectId,
	data,
}: WidgetGridProps) {
	const saveLayoutDebounced = useDashboardStore((s) => s.saveLayoutDebounced);

	const { width, containerRef } = useContainerWidth({ initialWidth: 1200 });

	/** Ключ для принудительного re-render содержимого виджета после refresh. */
	const [refreshKeys, setRefreshKeys] = useState<Record<string, number>>({});

	const layout = useMemo(
		() =>
			items.map((item) => {
				const constraints =
					WIDGET_CONFIG[item.widget_type]?.constraints ?? DEFAULT_CONSTRAINTS;
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

	/** Типы виджетов, отсутствующие на сетке. */
	const hiddenWidgetTypes = useMemo(
		() => Object.keys(WIDGET_CONFIG).filter((type) => !existingTypes.has(type)),
		[existingTypes],
	);

	return (
		<div ref={containerRef}>
			{/* Вернуть скрытые виджеты */}
			{hiddenWidgetTypes.length > 0 && (
				<div className="mb-4">
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<button
								type="button"
								className="inline-flex items-center gap-1.5 rounded-md border border-border/60 bg-card px-3 py-1.5 text-sm text-t2 hover:bg-accent hover:text-t1 transition-colors"
							>
								<Layers className="h-3.5 w-3.5" />
								Добавить виджет
							</button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="start">
							{hiddenWidgetTypes.map((type) => {
								const cfg = WIDGET_CONFIG[type];
								const Icon = cfg?.icon;
								return (
									<DropdownMenuItem
										key={type}
										onClick={() => handleAddWidget(type)}
									>
										{Icon && <Icon className="h-4 w-4" />}
										{cfg?.name ?? type}
									</DropdownMenuItem>
								);
							})}
						</DropdownMenuContent>
					</DropdownMenu>
				</div>
			)}

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
						const cfg = WIDGET_CONFIG[item.widget_type];
						if (!cfg) return <div key={item.i} />;

						const {
							component: WidgetComponent,
							icon: Icon,
							accent,
							name,
						} = cfg;

						return (
							<div
								key={item.i}
								className="group/widget dash-widget relative text-card-foreground"
							>
								{/* Акцентная верхняя граница */}
								<div
									className={`absolute inset-x-0 top-0 h-0.5 bg-gradient-to-r ${accent} opacity-60 group-hover/widget:opacity-100 transition-opacity duration-200`}
								/>

								{/* Заголовок виджета */}
								<div className="widget-drag-handle flex items-center justify-between px-5 pt-[18px] cursor-move">
									<div className="flex items-center gap-2">
										<GripVertical className="h-3.5 w-3.5 text-[var(--t4)]" />
										{Icon && <Icon className="h-4 w-4 text-[var(--t2)]" />}
										<span className="text-[13px] font-bold text-[var(--t2)] uppercase tracking-[0.04em]">
											{name}
										</span>
									</div>
									<div className="flex items-center gap-1">
										<DropdownMenu>
											<DropdownMenuTrigger asChild>
												<button
													type="button"
													className="w-7 h-7 flex items-center justify-center rounded-[var(--r1)] text-[var(--t4)] text-base hover:bg-[var(--hover)] hover:text-[var(--t2)] transition-colors opacity-0 group-hover/widget:opacity-100"
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

								{/* Содержимое виджета */}
								<div className="px-5 pb-5 pt-3.5 overflow-hidden h-[calc(100%-52px)]">
									<Suspense fallback={<ChartWidgetSkeleton />}>
										<WidgetComponent
											key={refreshKeys[item.i] ?? 0}
											data={data}
										/>
									</Suspense>
								</div>
							</div>
						);
					})}
				</GridLayout>
			)}
		</div>
	);
}
