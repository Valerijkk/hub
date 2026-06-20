const DEFAULT_API_URL = "http://127.0.0.1:3000";

export function getApiUrl(): string {
	return process.env.PLAYWRIGHT_API_URL ?? DEFAULT_API_URL;
}

export interface HealthResponse {
	status: string;
	services: Record<string, string>;
}

export async function isBackendReachable(
	apiUrl = getApiUrl(),
	timeoutMs = 3_000,
): Promise<boolean> {
	try {
		const res = await fetch(`${apiUrl}/api/health`, {
			signal: AbortSignal.timeout(timeoutMs),
		});
		return res.ok || res.status === 503;
	} catch {
		return false;
	}
}

export async function isDevModeAvailable(
	apiUrl = getApiUrl(),
	timeoutMs = 3_000,
): Promise<boolean> {
	try {
		const res = await fetch(`${apiUrl}/api/dev/status`, {
			signal: AbortSignal.timeout(timeoutMs),
		});
		if (!res.ok) return false;
		const body = (await res.json()) as { dev_mode?: boolean };
		return body.dev_mode === true;
	} catch {
		return false;
	}
}

export async function fetchHealth(
	apiUrl = getApiUrl(),
): Promise<HealthResponse | null> {
	try {
		const res = await fetch(`${apiUrl}/api/health`, {
			signal: AbortSignal.timeout(5_000),
		});
		if (!res.ok && res.status !== 503) return null;
		return (await res.json()) as HealthResponse;
	} catch {
		return null;
	}
}
