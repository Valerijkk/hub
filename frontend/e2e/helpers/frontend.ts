const DEFAULT_BASE_URL = "http://localhost:5173";

export function getBaseUrl(): string {
	return process.env.PLAYWRIGHT_BASE_URL ?? DEFAULT_BASE_URL;
}

export async function isFrontendReachable(
	baseURL = getBaseUrl(),
	timeoutMs = 3_000,
): Promise<boolean> {
	try {
		const res = await fetch(baseURL, {
			signal: AbortSignal.timeout(timeoutMs),
		});
		return res.ok;
	} catch {
		return false;
	}
}
