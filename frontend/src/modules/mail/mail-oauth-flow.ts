export type MailOAuthFlow = "setup" | "connection";

const STORAGE_KEY = "hub.mail.oauth-flow.v1";
const MARKER_VERSION = 1;
const STATE_BYTES = 32;
const MARKER_MAX_LENGTH = 256;
const MARKER_TTL_MS = 11 * 60_000;
const CONTROL_CHARACTERS = /[\u0000-\u001F\u007F-\u009F]/;
const BASE64URL_STATE = /^[A-Za-z0-9_-]{43}$/;
const LOCAL_HTTP_HOSTS = new Set(["localhost", "127.0.0.1", "[::1]"]);

interface MailOAuthFlowMarker {
	version: number;
	state: string;
	flow: MailOAuthFlow;
	created_at_ms: number;
}

function isMailOAuthFlow(value: unknown): value is MailOAuthFlow {
	return value === "setup" || value === "connection";
}

function hasControlCharacters(value: string): boolean {
	if (CONTROL_CHARACTERS.test(value)) return true;
	try {
		return CONTROL_CHARACTERS.test(decodeURIComponent(value));
	} catch {
		return true;
	}
}

function isSafeTimestamp(value: unknown): value is number {
	return typeof value === "number" && Number.isSafeInteger(value) && value >= 0;
}

function encodeBase64URL(bytes: Uint8Array): string {
	let binary = "";
	for (const byte of bytes) {
		binary += String.fromCharCode(byte);
	}
	return btoa(binary).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/, "");
}

function decodeBase64URL(value: string): Uint8Array | null {
	try {
		const binary = atob(value.replaceAll("-", "+").replaceAll("_", "/") + "=");
		const bytes = new Uint8Array(binary.length);
		for (let index = 0; index < binary.length; index += 1) {
			bytes[index] = binary.charCodeAt(index);
		}
		return bytes;
	} catch {
		return null;
	}
}

export function isCanonicalMailOAuthState(value: unknown): value is string {
	if (typeof value !== "string" || !BASE64URL_STATE.test(value)) return false;
	const decoded = decodeBase64URL(value);
	return decoded?.byteLength === STATE_BYTES && encodeBase64URL(decoded) === value;
}

function parseAuthorizationState(authorizationURL: string): string | null {
	if (hasControlCharacters(authorizationURL)) return null;
	try {
		const url = new URL(authorizationURL);
		const isLocalHTTP =
			url.protocol === "http:" && LOCAL_HTTP_HOSTS.has(url.hostname.toLowerCase());
		if (
			(url.protocol !== "https:" && !isLocalHTTP) ||
			url.username.length > 0 ||
			url.password.length > 0
		) {
			return null;
		}
		const states = url.searchParams.getAll("state");
		if (states.length !== 1) return null;
		return isCanonicalMailOAuthState(states[0]) ? states[0] : null;
	} catch {
		return null;
	}
}

function removeMarker(storage: Storage): boolean {
	try {
		storage.removeItem(STORAGE_KEY);
		return true;
	} catch {
		return false;
	}
}

function readValidMarker(value: unknown, nowMs: number): MailOAuthFlowMarker | null {
	if (!isSafeTimestamp(nowMs) || typeof value !== "object" || value === null || Array.isArray(value)) {
		return null;
	}
	const marker = value as Record<string, unknown>;
	const keys = Object.keys(marker);
	if (
		keys.length !== 4 ||
		!Object.prototype.hasOwnProperty.call(marker, "version") ||
		!Object.prototype.hasOwnProperty.call(marker, "state") ||
		!Object.prototype.hasOwnProperty.call(marker, "flow") ||
		!Object.prototype.hasOwnProperty.call(marker, "created_at_ms") ||
		marker.version !== MARKER_VERSION ||
		!isCanonicalMailOAuthState(marker.state) ||
		!isMailOAuthFlow(marker.flow) ||
		!isSafeTimestamp(marker.created_at_ms)
	) {
		return null;
	}
	const age = nowMs - marker.created_at_ms;
	if (age < 0 || age > MARKER_TTL_MS) return null;
	return {
		version: MARKER_VERSION,
		state: marker.state,
		flow: marker.flow,
		created_at_ms: marker.created_at_ms,
	};
}

export function rememberMailOAuthFlow(
	authorizationURL: string,
	flow: MailOAuthFlow,
	storage: Storage,
	nowMs: number,
): boolean {
	if (!isMailOAuthFlow(flow) || !isSafeTimestamp(nowMs)) return false;
	const state = parseAuthorizationState(authorizationURL);
	if (!state) return false;

	const serialized = JSON.stringify({
		version: MARKER_VERSION,
		state,
		flow,
		created_at_ms: nowMs,
	});
	if (serialized.length > MARKER_MAX_LENGTH) return false;
	try {
		storage.setItem(STORAGE_KEY, serialized);
		return true;
	} catch {
		return false;
	}
}

export function consumeMailOAuthFlow(
	callbackState: string,
	storage: Storage,
	nowMs: number,
): MailOAuthFlow | null {
	if (!isCanonicalMailOAuthState(callbackState) || !isSafeTimestamp(nowMs)) return null;

	let serialized: string | null;
	try {
		serialized = storage.getItem(STORAGE_KEY);
	} catch {
		return null;
	}
	if (serialized === null) return null;
	if (serialized.length === 0 || serialized.length > MARKER_MAX_LENGTH) {
		removeMarker(storage);
		return null;
	}

	let parsed: unknown;
	try {
		parsed = JSON.parse(serialized);
	} catch {
		removeMarker(storage);
		return null;
	}
	const marker = readValidMarker(parsed, nowMs);
	if (!marker) {
		removeMarker(storage);
		return null;
	}
	if (marker.state !== callbackState) return null;
	return removeMarker(storage) ? marker.flow : null;
}
