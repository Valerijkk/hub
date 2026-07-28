import { toast } from "sonner";
import type { ApiError } from "@/api/client";
import i18n from "@/i18n";

const RATE_LIMIT_TOAST_ID = "hub-api-rate-limit";

let rateLimitTimer: ReturnType<typeof setInterval> | null = null;

function clearRateLimitTimer(): void {
	if (rateLimitTimer) {
		clearInterval(rateLimitTimer);
		rateLimitTimer = null;
	}
}

function extractApiError(error: unknown): ApiError | null {
	if (!error || typeof error !== "object") return null;
	const candidate = error as Record<string, unknown>;
	if (typeof candidate.message !== "string") return null;
	// Обычный Error/DOMException тоже имеют message — требуем API-маркер.
	const hasCode = typeof candidate.code === "number";
	const hasRequestId = typeof candidate.request_id === "string";
	if (!hasCode && !hasRequestId) return null;
	return error as ApiError;
}

function requestIdDescription(requestId?: string): string | undefined {
	if (!requestId?.trim()) {
		return undefined;
	}
	return i18n.t("common:apiErrors.requestIdHint", { id: requestId.trim() });
}

/** Сообщение для 429 с секундами до повтора — FB-03 / UX-P1-05. */
export function formatRateLimitMessage(seconds: number): string {
	return i18n.t("common:apiErrors.rateLimitRetry", { seconds });
}

/** Извлекает человекочитаемое сообщение из ошибки API — UX-08. */
export function formatApiError(error: unknown): string {
	const apiErr = extractApiError(error);
	if (apiErr) {
		const code = apiErr.code ?? 0;
		if (code === 429 && apiErr.retry_after && apiErr.retry_after > 0) {
			return formatRateLimitMessage(apiErr.retry_after);
		}
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

function showRateLimitCountdown(retryAfter: number, requestId?: string): void {
	clearRateLimitTimer();

	let remaining = retryAfter;
	const description = requestIdDescription(requestId);
	const baseOpts = {
		id: RATE_LIMIT_TOAST_ID,
		duration: Number.POSITIVE_INFINITY,
		...(description ? { description } : {}),
	};

	toast.error(formatRateLimitMessage(remaining), baseOpts);

	rateLimitTimer = setInterval(() => {
		remaining -= 1;
		if (remaining <= 0) {
			clearRateLimitTimer();
			toast.dismiss(RATE_LIMIT_TOAST_ID);
			return;
		}
		toast.error(formatRateLimitMessage(remaining), baseOpts);
	}, 1000);
}

/** Toast с formatApiError и опциональным fallback — UX-08 rollout. */
export function toastApiError(error: unknown, fallback?: string): void {
	const apiErr = extractApiError(error);
	if (apiErr?.code === 429 && apiErr.retry_after && apiErr.retry_after > 0) {
		showRateLimitCountdown(apiErr.retry_after, apiErr.request_id);
		return;
	}

	const message =
		formatApiError(error) || fallback || i18n.t("common:apiErrors.unknown");
	const description = requestIdDescription(apiErr?.request_id);
	if (description) {
		toast.error(message, { description });
	} else {
		toast.error(message);
	}
}

/** Тестовый хук — сброс interval между Vitest cases. */
export function resetRateLimitToastForTests(): void {
	clearRateLimitTimer();
}
