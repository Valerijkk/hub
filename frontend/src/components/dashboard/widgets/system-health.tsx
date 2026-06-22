/**
 * Виджет «System Health» — статусы PG, Redis, NATS, MongoDB, SeaweedFS.
 */
import { Activity, Server } from "lucide-react";
import type { AdminStats } from "@/api/dashboard";

interface Props {
	data?: unknown;
}

/** Конфиг статусов с glow-цветами. */
const STATUS_CONFIG: Record<
	string,
	{ color: string; glow: string; pulse: boolean; label: string }
> = {
	healthy: {
		color: "bg-green-500",
		glow: "0 0 8px rgba(34,197,94,0.6)",
		pulse: true,
		label: "Работает",
	},
	unhealthy: {
		color: "bg-red-500",
		glow: "0 0 8px rgba(239,68,68,0.6)",
		pulse: false,
		label: "Ошибка",
	},
	degraded: {
		color: "bg-yellow-500",
		glow: "0 0 8px rgba(234,179,8,0.6)",
		pulse: true,
		label: "Деградация",
	},
};

const DEFAULT_STATUS = {
	color: "bg-gray-400",
	glow: "none",
	pulse: false,
	label: "Неизвестно",
};

/** Виджет здоровья сервисов. */
export function SystemHealthWidget({ data }: Props) {
	const adminData = data as AdminStats | undefined;
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
								className={`h-2.5 w-2.5 rounded-full ${config.color}`}
								style={{ boxShadow: config.glow }}
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
							className={`text-[10px] uppercase tracking-wider px-2 py-0.5 rounded-full ${
								status === "healthy"
									? "bg-green-500/10 text-green-400"
									: status === "unhealthy"
										? "bg-red-500/10 text-red-400"
										: "bg-yellow-500/10 text-yellow-400"
							}`}
						>
							{config.label}
						</span>
					</div>
				);
			})}
		</div>
	);
}
