import type { PresenceStatus } from "@/api/messenger";
import { cn } from "@/lib/utils";

/** Пропсы компонента PresenceIndicator. */
interface PresenceIndicatorProps {
	/** Статус присутствия пользователя. */
	status: PresenceStatus;
	/**
	 * Poll presence упал (A2W11-3-001): не показывать «Не в сети» —
	 * статус недоступен.
	 */
	unavailable?: boolean;
	/** Размер индикатора. */
	size?: "sm" | "md";
	/** Дополнительные CSS классы. */
	className?: string;
}

/** Цвета для каждого статуса присутствия (дизайн-токены `--green`/`--gray`/`--amber`). */
const statusColors: Record<PresenceStatus, string> = {
	online: "bg-[var(--green)]",
	offline: "bg-[var(--gray)]",
	away: "bg-[var(--amber)]",
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
	unavailable = false,
	size = "sm",
	className,
}: PresenceIndicatorProps) {
	const label = unavailable ? "Статус недоступен" : statusLabels[status];
	const color = unavailable
		? "bg-[var(--amber)] opacity-60"
		: statusColors[status];

	return (
		<span
			role="status"
			aria-label={label}
			className={cn(
				"inline-block shrink-0 rounded-full",
				color,
				size === "sm" ? "h-2.5 w-2.5" : "h-3.5 w-3.5",
				className,
			)}
		/>
	);
}
