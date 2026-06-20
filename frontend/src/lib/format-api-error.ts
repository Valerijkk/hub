import { toast } from "sonner";
import type { ApiError } from "@/api/client";
import i18n from "@/i18n";

/** Извлекает человекочитаемое сообщение из ошибки API — UX-08. */
export function formatApiError(error: unknown): string {
	if (error && typeof error === "object" && "message" in error) {
		const apiErr = error as ApiError;
		const code = apiErr.code ?? 0;
		if (apiErr.message && apiErr.message.trim().length > 0) {
			return apiErr.message;
		}
		const key = `common:apiErrors.${code}`;
		if (i18n.exists(key)) {
			return i18n.t(key);
		}
	}
	if (error instanceof Error && error.message) {
		return error.message;
	}
	return i18n.t("common:apiErrors.unknown");
}

/** Toast с formatApiError и опциональным fallback — UX-08 rollout. */
export function toastApiError(error: unknown, fallback?: string): void {
	toast.error(
		formatApiError(error) || fallback || i18n.t("common:apiErrors.unknown"),
	);
}
