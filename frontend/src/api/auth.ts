import { api } from "./client";

// ---------------------------------------------------------------------------
// Типы запросов (Request DTOs)
// ---------------------------------------------------------------------------

/** Запрос на отправку magic-link на email. */
interface MagicLinkRequest {
	email: string;
	invite_token?: string;
}

/** Запрос на верификацию magic-link кода. */
interface VerifyMagicLinkRequest {
	email: string;
	code: string;
	invite_token?: string;
}

/** Запрос на обновление профиля пользователя. */
export interface UpdateProfileRequest {
	name?: string;
	display_name?: string;
	avatar_url?: string;
	phone?: string;
	position?: string;
	department?: string;
	timezone?: string;
	language?: string;
	birthday?: string;
	quiet_hours_start?: string;
	quiet_hours_end?: string;
}

// ---------------------------------------------------------------------------
// Типы ответов (Response DTOs)
// ---------------------------------------------------------------------------

/** Пара JWT-токенов (access + refresh). */
export interface TokenPairResponse {
	access_token: string;
	refresh_token: string;
	expires_in: number;
}

/** Полные данные пользователя. */
export interface UserResponse {
	id: string;
	email: string;
	name?: string;
	display_name?: string;
	avatar_url?: string;
	telegram_id?: string;
	phone?: string;
	position?: string;
	department?: string;
	timezone: string;
	language: string;
	birthday?: string;
	system_role: string;
	status?: string;
	onboarding_completed: boolean;
}

/** Данные сессии пользователя. */
export interface SessionResponse {
	id: string;
	device: string;
	ip_address?: string;
	last_active_at: string;
	created_at: string;
	is_current: boolean;
}

// ---------------------------------------------------------------------------
// API-функции: Аутентификация и профиль
// ---------------------------------------------------------------------------

/** Endpoint: POST /auth/magic-link — отправить magic-link на email. */
export function sendMagicLink(data: MagicLinkRequest) {
	return api.post<{ message: string; code?: string }>("/auth/magic-link", data);
}

/** Endpoint: POST /auth/magic-link/verify — верифицировать код из magic-link. */
export function verifyMagicLink(data: VerifyMagicLinkRequest) {
	return api.post<TokenPairResponse>("/auth/magic-link/verify", data);
}

/**
 * BE `*time.Time` accepts RFC3339 JSON; `<input type="date">` yields YYYY-MM-DD.
 * Normalize date-only to UTC midnight so encoding/json does not 400 (A11-006).
 */
export function toBirthdayRFC3339(
	birthday: string | undefined,
): string | undefined {
	if (!birthday) return undefined;
	const trimmed = birthday.trim();
	if (/^\d{4}-\d{2}-\d{2}$/.test(trimmed)) {
		return `${trimmed}T00:00:00Z`;
	}
	return trimmed;
}

/** Slice RFC3339 / ISO birthday to YYYY-MM-DD for date inputs. */
export function birthdayForDateInput(
	birthday: string | undefined | null,
): string {
	if (!birthday) return "";
	const m = /^(\d{4}-\d{2}-\d{2})/.exec(birthday.trim());
	return m?.[1] ?? "";
}

function withNormalizedBirthday(
	data: UpdateProfileRequest,
): UpdateProfileRequest {
	if (data.birthday === undefined) return data;
	return { ...data, birthday: toBirthdayRFC3339(data.birthday) };
}

/** Endpoint: GET /auth/me — получить данные текущего пользователя. */
export function getMe() {
	return api.get<UserResponse>("/auth/me");
}

/** Endpoint: PUT /auth/me — обновить профиль текущего пользователя. */
export function updateProfile(data: UpdateProfileRequest) {
	return api.put<UserResponse>("/auth/me", withNormalizedBirthday(data));
}

/** Endpoint: POST /auth/me/onboarding — завершить онбординг с данными профиля. */
export function completeOnboarding(data: { profile: UpdateProfileRequest }) {
	return api.post<UserResponse>("/auth/me/onboarding", {
		profile: withNormalizedBirthday(data.profile),
	});
}

/** Endpoint: GET /auth/sessions — получить список активных сессий. */
export function listSessions() {
	return api.get<SessionResponse[]>("/auth/sessions");
}

/** Endpoint: DELETE /auth/sessions/:id — отозвать (завершить) сессию. */
export function revokeSession(sessionId: string) {
	return api.delete<void>(`/auth/sessions/${sessionId}`);
}

/** Endpoint: POST /auth/logout — выйти из системы. */
export function logout() {
	return api.post<void>("/auth/logout");
}

/** Ответ POST /auth/ws-ticket (ADR-WS-JWT-HANDSHAKE-2026). */
export interface WSTicketResponse {
	ticket: string;
	expires_in: number;
}

/** Endpoint: POST /auth/ws-ticket — одноразовый ticket для WS handshake. */
export function createWSTicket() {
	return api.post<WSTicketResponse>("/auth/ws-ticket");
}
