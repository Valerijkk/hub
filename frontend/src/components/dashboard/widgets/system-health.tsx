/**
 * Widget "System Health" -- service status indicators (PG, Redis, NATS, MongoDB, SeaweedFS).
 */
import { Activity, Server } from "lucide-react";
import type { AdminStats } from "@/api/dashboard";

interface Props {
	data?: unknown;
}

/** Status configuration. */
const STATUS_CONFIG: Record<string, { color: string; pulse: boolean; label: string }> = {
	healthy: { color: "bg-green-500", pulse: true, label: "Работает" },
	unhealthy: { color: "bg-red-500", pulse: false, label: "Ошибка" },
	degraded: { color: "bg-yellow-500", pulse: true, label: "Деградация" },
};

const DEFAULT_STATUS = { color: "bg-gray-400", pulse: false, label: "Неизвестно" };

/** Service health widget. */
export function SystemHealthWidget({ data }: Props) {
	const adminData = data as AdminStats | undefined;
	const health = adminData?.service_health ?? {};

	const services = Object.entries(health);

	if (services.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-muted-foreground">
				<Server className="h-8 w-8 opacity-40" />
				<p className="text-sm">Нет данных о сервисах</p>
			</div>
		);
	}

	const healthyCount = services.filter(([, s]) => s === "healthy").length;

	return (
		<div className="space-y-3">
			{/* Summary bar */}
			<div className="flex items-center gap-2 pb-2 border-b border-border/50">
				<Activity className="h-4 w-4 text-muted-foreground" />
				<span className="text-xs text-muted-foreground">
					{healthyCount}/{services.length} сервисов в норме
				</span>
			</div>

			{services.map(([name, status]) => {
				const config = STATUS_CONFIG[status as string] ?? DEFAULT_STATUS;
				return (
					<div
						key={name}
						className="flex items-center gap-3 rounded-lg px-2 py-1.5 transition-colors hover:bg-accent/50"
					>
						<div className="relative flex items-center justify-center">
							<span
								className={`h-2.5 w-2.5 rounded-full ${config.color}`}
							/>
							{config.pulse && (
								<span
									className={`absolute h-2.5 w-2.5 rounded-full ${config.color} animate-ping opacity-30`}
								/>
							)}
						</div>
						<span className="text-sm font-medium capitalize flex-1">
							{name}
						</span>
						<span
							className={`text-xs px-2 py-0.5 rounded-full ${
								status === "healthy"
									? "bg-green-500/10 text-green-600 dark:text-green-400"
									: status === "unhealthy"
										? "bg-red-500/10 text-red-600 dark:text-red-400"
										: "bg-yellow-500/10 text-yellow-600 dark:text-yellow-400"
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
