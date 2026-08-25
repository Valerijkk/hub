import { create } from "zustand";

import type { UserResponse } from "@/api/auth";
import { resetAuthenticatedClientState } from "@/lib/reset-authenticated-client-state";

/**
 * Состояние стора аутентификации.
 * JWT (access + refresh) — только в памяти (ADR-WS-JWT-HANDSHAKE-2026 /
 * A2W11-2-001). sessionStorage/localStorage не содержат токены.
 * Сессия пропадает при reload вкладки до ADR httpOnly refresh cookie.
 */
interface AuthState {
	/** JWT access-токен для авторизации API-запросов. */
	accessToken: string | null;
	/** Refresh-токен для обновления access-токена. */
	refreshToken: string | null;
	/** Данные текущего пользователя (загружаются отдельно через /auth/me). */
	user: UserResponse | null;
	/** Флаг аутентификации (true после получения токенов). */
	isAuthenticated: boolean;
	/** Сохранить пару токенов после логина. */
	setTokens: (access: string, refresh: string) => void;
	/** Установить данные пользователя. */
	setUser: (user: UserResponse) => void;
	/**
	 * Очистить локальную сессию и транслировать logout в другие вкладки.
	 * Для remote-sync используйте clearSessionLocal — без повторного broadcast.
	 */
	logout: () => void;
	/**
	 * Очистить локальную сессию без BroadcastChannel (W11-A2-004).
	 * Вызывается при получении remote logout, чтобы не зациклить sync.
	 */
	clearSessionLocal: () => void;
}

/**
 * Multi-tab синхронизация login/logout через BroadcastChannel (FE-H2).
 *
 * Когда пользователь делает login/logout в одной вкладке, событие
 * транслируется в другие вкладки того же origin. Это позволяет:
 *  - очистить TanStack Query cache при logout (см. useAuthSync в _authenticated.tsx)
 *  - переключить вкладку на страницу логина / главную при login
 *
 * Реализовано через BroadcastChannel API (есть в Chromium/Firefox/Safari 15.4+).
 * Если API недоступен — sync молча отключается (single-tab fallback).
 */
const AUTH_CHANNEL_NAME = "hub-auth";

/** Типы событий, которые транслируются между вкладками. */
export type AuthSyncEvent =
	| { type: "logout" }
	| { type: "login"; userId: string | null };

let authChannel: BroadcastChannel | null = null;
if (typeof BroadcastChannel !== "undefined") {
	try {
		authChannel = new BroadcastChannel(AUTH_CHANNEL_NAME);
	} catch {
		authChannel = null;
	}
}

/**
 * Транслирует событие auth-sync в другие вкладки.
 * Пишет в BroadcastChannel и зеркало localStorage (storage fallback).
 */
function broadcastAuth(event: AuthSyncEvent): void {
	if (authChannel) {
		try {
			authChannel.postMessage(event);
		} catch {
			// игнорируем: канал мог закрыться
		}
	}
	// Mirror для storage-event fallback (старые браузеры / без BroadcastChannel).
	try {
		localStorage.setItem(`${AUTH_CHANNEL_NAME}-event`, JSON.stringify(event));
		// Удаляем сразу — storage срабатывает на setItem в других вкладках.
		localStorage.removeItem(`${AUTH_CHANNEL_NAME}-event`);
	} catch {
		// private mode / quota — не критично
	}
}

/**
 * Подписаться на auth-sync события из других вкладок.
 * Возвращает функцию отписки.
 *
 * @param handler - обработчик события (получает event payload)
 */
export function subscribeAuthSync(
	handler: (event: AuthSyncEvent) => void,
): () => void {
	if (typeof BroadcastChannel !== "undefined" && authChannel) {
		const listener = (ev: MessageEvent<AuthSyncEvent>) => {
			handler(ev.data);
		};
		authChannel.addEventListener("message", listener);
		return () => {
			authChannel?.removeEventListener("message", listener);
		};
	}

	// Запасной вариант: storage event (для очень старых браузеров и кейсов, когда
	// BroadcastChannel недоступен).
	const storageListener = (ev: StorageEvent) => {
		if (ev.key === `${AUTH_CHANNEL_NAME}-event` && ev.newValue) {
			try {
				handler(JSON.parse(ev.newValue) as AuthSyncEvent);
			} catch {
				// игнорируем: операция не критична для UX
			}
		}
	};
	window.addEventListener("storage", storageListener);
	return () => window.removeEventListener("storage", storageListener);
}

/** Bootstrap для E2E/Playwright: инъекция сессии без sessionStorage JWT. */
export type AuthBootstrap = {
	accessToken: string;
	refreshToken: string;
	user: UserResponse | null;
	isAuthenticated: boolean;
};

declare global {
	interface Window {
		__HUB_AUTH_BOOTSTRAP__?: AuthBootstrap;
		__HUB_SET_AUTH__?: (state: Partial<AuthBootstrap>) => void;
	}
}

function readAuthBootstrap(): Partial<AuthState> {
	if (typeof window === "undefined") return {};
	const boot = window.__HUB_AUTH_BOOTSTRAP__;
	if (!boot) return {};
	delete window.__HUB_AUTH_BOOTSTRAP__;
	return {
		accessToken: boot.accessToken,
		refreshToken: boot.refreshToken,
		user: boot.user,
		isAuthenticated: boot.isAuthenticated,
	};
}

const boot = readAuthBootstrap();

/** Стор аутентификации — JWT только в памяти (без persist). */
export const useAuthStore = create<AuthState>()((set) => ({
	accessToken: boot.accessToken ?? null,
	refreshToken: boot.refreshToken ?? null,
	user: boot.user ?? null,
	isAuthenticated: boot.isAuthenticated ?? false,
	setTokens: (access, refresh) => {
		set({
			accessToken: access,
			refreshToken: refresh,
			isAuthenticated: true,
		});
		const userId = useAuthStore.getState().user?.id ?? null;
		broadcastAuth({ type: "login", userId });
	},
	setUser: (user) => {
		set({ user });
		// После /auth/me соседние вкладки узнают userId (same-user skip в _authenticated).
		broadcastAuth({ type: "login", userId: user.id });
	},
	clearSessionLocal: () => {
		resetAuthenticatedClientState();
		set({
			accessToken: null,
			refreshToken: null,
			user: null,
			isAuthenticated: false,
		});
	},
	logout: () => {
		resetAuthenticatedClientState();
		set({
			accessToken: null,
			refreshToken: null,
			user: null,
			isAuthenticated: false,
		});
		broadcastAuth({ type: "logout" });
	},
}));

if (typeof window !== "undefined") {
	window.__HUB_SET_AUTH__ = (state) => {
		useAuthStore.setState({
			accessToken: state.accessToken ?? null,
			refreshToken: state.refreshToken ?? null,
			user: state.user ?? null,
			isAuthenticated: state.isAuthenticated ?? Boolean(state.accessToken),
		});
	};
}
