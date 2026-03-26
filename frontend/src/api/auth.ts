import { api } from "./client";

// ---------------------------------------------------------------------------
// Типы запросов (Request DTOs)
// ---------------------------------------------------------------------------

/** Запрос на отправку magic-link на email. */
interface MagicLinkRequest {
	email: string;
}

/** Запрос на верификацию magic-link кода. */
interface VerifyMagicLinkRequest {
	email: string;
	code: string;
}

/** Запрос на авторизацию через Telegram. */
interface TelegramAuthRequest {
	init_data: string;
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

/** Endpoint: POST /auth/telegram — авторизоваться через Telegram. */
export function loginTelegram(data: TelegramAuthRequest) {
	return api.post<TokenPairResponse>("/auth/telegram", data);
}

/** Endpoint: GET /auth/me — получить данные текущего пользователя. */
export function getMe() {
	return api.get<UserResponse>("/auth/me");
}

/** Endpoint: PUT /auth/me — обновить профиль текущего пользователя. */
export function updateProfile(data: UpdateProfileRequest) {
	return api.put<UserResponse>("/auth/me", data);
}

/** Endpoint: POST /auth/me/onboarding — завершить онбординг с данными профиля. */
export function completeOnboarding(data: { profile: UpdateProfileRequest }) {
	return api.post<UserResponse>("/auth/me/onboarding", data);
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
