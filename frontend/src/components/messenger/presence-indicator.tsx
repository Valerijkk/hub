import type { PresenceStatus } from "@/api/messenger";
import { cn } from "@/lib/utils";

/** Пропсы компонента PresenceIndicator. */
interface PresenceIndicatorProps {
	/** Статус присутствия пользователя. */
	status: PresenceStatus;
	/** Размер индикатора. */
	size?: "sm" | "md";
	/** Дополнительные CSS классы. */
	className?: string;
}

/** Цвета для каждого статуса присутствия. */
const statusColors: Record<PresenceStatus, string> = {
	online: "bg-green-500",
	offline: "bg-gray-400",
	away: "bg-yellow-500",
};

/** Подписи для каждого статуса (для aria-label). */
const statusLabels: Record<PresenceStatus, string> = {
	online: "В сети",
	offline: "Не в сети",
	away: "Отошёл",
};

/**
 * PresenceIndicator отображает цветной кружок, показывающий
 * статус присутствия пользователя: зелёный (online), серый (offline), жёлтый (away).
 */
export function PresenceIndicator({
	status,
	size = "sm",
	className,
}: PresenceIndicatorProps) {
	return (
		<span
			role="status"
			aria-label={statusLabels[status]}
			className={cn(
				"inline-block shrink-0 rounded-full",
				statusColors[status],
				size === "sm" ? "h-2.5 w-2.5" : "h-3.5 w-3.5",
				className,
			)}
		/>
	);
}
