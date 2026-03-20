import { api } from "./client";

// --- Request types ---

interface MagicLinkRequest {
  email: string;
}

interface VerifyMagicLinkRequest {
  email: string;
  code: string;
}

interface TelegramAuthRequest {
  init_data: string;
}

interface UpdateProfileRequest {
  name?: string;
  display_name?: string;
  avatar_url?: string;
  phone?: string;
  position?: string;
  department?: string;
  timezone?: string;
  language?: string;
  birthday?: string;
}

// --- Response types ---

export interface TokenPairResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

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

export interface SessionResponse {
  id: string;
  device: string;
  ip_address?: string;
  last_active_at: string;
  created_at: string;
  is_current: boolean;
}

// --- API functions ---

export function sendMagicLink(data: MagicLinkRequest) {
  return api.post<{ message: string }>("/auth/magic-link", data);
}

export function verifyMagicLink(data: VerifyMagicLinkRequest) {
  return api.post<TokenPairResponse>("/auth/magic-link/verify", data);
}

export function loginTelegram(data: TelegramAuthRequest) {
  return api.post<TokenPairResponse>("/auth/telegram", data);
}

export function getMe() {
  return api.get<UserResponse>("/auth/me");
}

export function updateProfile(data: UpdateProfileRequest) {
  return api.put<UserResponse>("/auth/me", data);
}

export function completeOnboarding(data: { profile: UpdateProfileRequest }) {
  return api.post<UserResponse>("/auth/me/onboarding", data);
}

export function listSessions() {
  return api.get<SessionResponse[]>("/auth/sessions");
}

export function revokeSession(sessionId: string) {
  return api.delete<void>(`/auth/sessions/${sessionId}`);
}

export function logout() {
  return api.post<void>("/auth/logout");
}
