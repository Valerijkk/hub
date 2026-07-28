import i18n from "@/i18n";

/**
 * Форматирует дату как относительное время с учётом locale пользователя (ru|en).
 * Будущие даты — «через N…» / "in N…", не «Только что» (P2-FE-15).
 *
 * @param dateStr - ISO-строка даты
 * @param options.now - якорь времени (для тестов)
 * @param options.locale - принудительная локаль
 */
export function formatRelativeTime(
	dateStr: string | undefined | null,
	options?: { now?: number; locale?: string },
): string {
	if (!dateStr || typeof dateStr !== "string") return "";
	const date = new Date(dateStr);
	if (Number.isNaN(date.getTime())) return "";

	const nowMs = options?.now ?? Date.now();
	const diffMs = nowMs - date.getTime();
	const locale =
		options?.locale ??
		(i18n.language?.startsWith("en") ? "en" : "ru");

	const rtf = new Intl.RelativeTimeFormat(locale, { numeric: "auto" });
	const absMs = Math.abs(diffMs);
	const past = diffMs >= 0;

	const fmt = (n: number, unit: Intl.RelativeTimeFormatUnit) =>
		rtf.format(past ? -n : n, unit);

	const absSec = Math.floor(absMs / 1000);
	if (absSec < 60) return fmt(0, "second");

	const absMin = Math.floor(absMs / 60000);
	if (absMin < 60) return fmt(absMin, "minute");

	const absHours = Math.floor(absMs / 3600000);
	if (absHours < 24) return fmt(absHours, "hour");

	const absDays = Math.floor(absMs / 86400000);
	if (absDays < 7) return fmt(absDays, "day");

	return date.toLocaleDateString(locale === "en" ? "en-US" : "ru-RU", {
		day: "numeric",
		month: "short",
	});
}
