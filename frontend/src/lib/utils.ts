import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

/** Объединяет CSS-классы с помощью clsx и tailwind-merge для корректного слияния Tailwind-утилит. */
export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

/**
 * Построение query string из объекта параметров.
 * Пропускает null/undefined значения. Массивы соединяются через запятую.
 * Возвращает строку вида "?key=val&..." или пустую строку если параметров нет.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function buildQueryString(params?: Record<string, any>): string {
	if (!params) return "";
	const entries = Object.entries(params)
		.filter(([, v]) => v != null)
		.map(([k, v]) => [k, Array.isArray(v) ? v.join(",") : String(v)]);
	if (entries.length === 0) return "";
	return "?" + new URLSearchParams(entries).toString();
}
