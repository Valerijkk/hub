/**
 * Виджет "Здоровье системы" -- индикаторы состояния сервисов (PG, Redis, NATS, MongoDB, SeaweedFS).
 */
import type { AdminStats } from "@/api/dashboard";

interface Props {
	data?: unknown;
}

/** Виджет состояния сервисов. */
export function SystemHealthWidget({ data }: Props) {
	const adminData = data as AdminStats | undefined;
	const health = adminData?.service_health ?? {};

	const services = Object.entries(health);

	if (services.length === 0) {
		return <p className="text-sm text-muted-foreground">Нет данных</p>;
	}

	return (
		<div className="space-y-3">
			{services.map(([name, status]) => (
				<div key={name} className="flex items-center gap-3">
					<span
						className={`h-3 w-3 rounded-full shrink-0 ${
							status === "healthy"
								? "bg-green-500"
								: status === "unhealthy"
									? "bg-red-500"
									: "bg-yellow-500"
						}`}
					/>
					<span className="text-sm font-medium capitalize">{name}</span>
					<span className="ml-auto text-xs text-muted-foreground">
						{status === "healthy"
							? "OK"
							: status === "unhealthy"
								? "Ошибка"
								: "Неизвестно"}
					</span>
				</div>
			))}
		</div>
	);
}
