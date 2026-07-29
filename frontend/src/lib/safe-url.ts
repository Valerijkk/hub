/**
 * Whitelist безопасных схем для <a href> / window.open.
 * Блокирует `javascript:`, `data:`, `vbscript:` и другие опасные схемы.
 */
export const SAFE_URL_RE = /^(https?:|mailto:|\/|#)/i;

/** True, если строка проходит SAFE_URL_RE (можно ставить в href / рендерить preview). */
export function isSafeUrl(url: unknown): url is string {
	return typeof url === "string" && SAFE_URL_RE.test(url);
}

/**
 * Санитизация URL для href / window.open. Небезопасное → `"#"` (no-op).
 */
export function sanitizeUrl(url: unknown): string {
	return isSafeUrl(url) ? url : "#";
}
