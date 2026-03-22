/**
 * Форматирует дату как относительное время на русском языке.
 * Возвращает "Только что", "5 мин", "2 ч", "Вчера" или дату (15 мар).
 *
 * @param dateStr - ISO-строка даты
 * @returns Отформатированная строка относительного времени
 */
export function formatRelativeTime(dateStr: string): string {
	const date = new Date(dateStr);
	const now = new Date();
	const diffMs = now.getTime() - date.getTime();
	const diffMin = Math.floor(diffMs / 60000);
	const diffHours = Math.floor(diffMs / 3600000);
	const diffDays = Math.floor(diffMs / 86400000);

	if (diffMin < 1) return "Только что";
	if (diffMin < 60) return `${diffMin} мин`;
	if (diffHours < 24) return `${diffHours} ч`;
	if (diffDays === 1) return "Вчера";
	if (diffDays < 7) return `${diffDays} дн`;

	return date.toLocaleDateString("ru-RU", {
		day: "numeric",
		month: "short",
	});
}
