/**
 * Виджет «System Health» — статусы PG, Redis, NATS, MongoDB, SeaweedFS.
 */
import { Activity, Server } from "lucide-react";
import {
	type DashboardWidgetData,
	isAdminStats,
} from "@/components/dashboard/widget-data";

interface Props {
	data?: DashboardWidgetData;
}

/** Конфиг статусов — semantic tokens, без raw *-500 / inline rgba (A3W9-2-005). */
const STATUS_CONFIG: Record<
	string,
	{ color: string; glow: string; pulse: boolean; label: string; badge: string }
> = {
	healthy: {
		color: "bg-[var(--green)]",
		glow: "shadow-[0_0_8px_var(--gG)]",
		pulse: true,
		label: "Работает",
		badge: "bg-[var(--gG)] text-[var(--green)]",
	},
	unhealthy: {
		color: "bg-[var(--red)]",
		glow: "shadow-[0_0_8px_var(--gR)]",
		pulse: false,
		label: "Ошибка",
		badge: "bg-[var(--gR)] text-[var(--red)]",
	},
	degraded: {
		color: "bg-[var(--amber)]",
		glow: "shadow-[0_0_8px_var(--amber)]",
		pulse: true,
		label: "Деградация",
		badge: "bg-[var(--amber)]/10 text-[var(--amber)]",
	},
};

const DEFAULT_STATUS = {
	color: "bg-[var(--t3)]",
	glow: "",
	pulse: false,
	label: "Неизвестно",
	badge: "bg-[var(--t3)]/20 text-[var(--t3)]",
};

/** Виджет здоровья сервисов. */
export function SystemHealthWidget({ data }: Props) {
	const adminData = isAdminStats(data) ? data : undefined;
	const health = adminData?.service_health ?? {};

	const services = Object.entries(health);

	if (services.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-t2">
				<Server className="h-8 w-8 opacity-40" />
				<p className="text-sm">Нет данных о сервисах</p>
			</div>
		);
	}

	const healthyCount = services.filter(([, s]) => s === "healthy").length;

	return (
		<div className="space-y-2">
			{/* Сводная полоска */}
			<div className="flex items-center gap-2 pb-2 border-b border-white/[0.06]">
				<Activity className="h-4 w-4 text-t2/60" />
				<span className="text-[11px] uppercase font-semibold tracking-[0.04em] text-t2">
					{healthyCount}/{services.length} сервисов в норме
				</span>
			</div>

			{services.map(([name, status]) => {
				const config = STATUS_CONFIG[status as string] ?? DEFAULT_STATUS;
				return (
					<div
						key={name}
						className="flex items-center gap-3 rounded-lg px-2.5 py-2 transition-colors hover:bg-white/[0.04]"
					>
						<div className="relative flex items-center justify-center">
							<span
								className={`h-2.5 w-2.5 rounded-full ${config.color} ${config.glow}`}
							/>
							{config.pulse && (
								<span
									className={`absolute h-2.5 w-2.5 rounded-full ${config.color} animate-ping opacity-30`}
								/>
							)}
						</div>
						<span className="text-[13px] font-medium capitalize flex-1">
							{name}
						</span>
						<span
							className={`text-[10px] uppercase tracking-wider px-2 py-0.5 rounded-full ${config.badge}`}
						>
							{config.label}
						</span>
					</div>
				);
			})}
		</div>
	);
}
