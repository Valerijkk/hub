/** Допустим только #RRGGBB для inline status colors — защита от CSS injection. */
export function sanitizeHexColor(
	color: string | undefined,
	fallback = "#808080",
): string {
	if (!color) return fallback;
	return /^#[0-9A-Fa-f]{6}$/.test(color) ? color : fallback;
}
