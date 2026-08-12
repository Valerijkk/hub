/** Безопасный вывод значения config окружения (не `[object Object]`). */
export function formatEnvConfigValue(value: unknown): string {
	if (value == null) return "";
	if (
		typeof value === "string" ||
		typeof value === "number" ||
		typeof value === "boolean"
	) {
		return String(value);
	}
	try {
		return JSON.stringify(value);
	} catch {
		return String(value);
	}
}
