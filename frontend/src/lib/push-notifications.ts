/**
 * Модуль работы с браузерными Push-уведомлениями.
 *
 * Паттерн "soft ask" — никогда не показываем браузерный промпт напрямую.
 * Сначала показываем кастомный баннер, и только при согласии пользователя
 * вызываем Notification.requestPermission().
 */

import { getVAPIDKey, subscribeToPush } from "@/api/notifications";

/** Ключ в localStorage для отслеживания отложенного запроса. */
const PUSH_DISMISSED_KEY = "hub_push_dismissed";
/** Ключ для отслеживания подписки. */
const PUSH_SUBSCRIBED_KEY = "hub_push_subscribed";
/** TTL «Позже» — баннер снова через 7 дней (P2-FE-15 / W11-A4-008). */
export const PUSH_DISMISS_TTL_MS = 7 * 24 * 60 * 60 * 1000;

/**
 * Регистрирует Service Worker для обработки push-событий.
 * Вызывается при загрузке приложения (после авторизации).
 */
export async function registerServiceWorker(): Promise<ServiceWorkerRegistration | null> {
	if (!("serviceWorker" in navigator)) {
		return null;
	}

	try {
		const registration = await navigator.serviceWorker.register("/sw.js", {
			scope: "/",
		});
		return registration;
	} catch (err) {
		console.error("Ошибка регистрации Service Worker:", err);
		return null;
	}
}

/**
 * Проверяет, поддерживаются ли push-уведомления в текущем браузере.
 */
export function isPushSupported(): boolean {
	return (
		"serviceWorker" in navigator &&
		"PushManager" in window &&
		"Notification" in window
	);
}

/**
 * Проверяет, нужно ли показывать баннер с предложением включить push.
 * Не показываем если:
 * - Push не поддерживается
 * - Уже подписан
 * - Пользователь нажал "Позже" и TTL ещё не истёк
 * - Разрешение уже отклонено в браузере
 */
export function shouldShowPushBanner(now = Date.now()): boolean {
	if (!isPushSupported()) return false;
	if (Notification.permission === "denied") return false;
	if (
		Notification.permission === "granted" &&
		localStorage.getItem(PUSH_SUBSCRIBED_KEY)
	)
		return false;
	const dismissed = localStorage.getItem(PUSH_DISMISSED_KEY);
	if (dismissed) {
		const ts = Number(dismissed);
		if (Number.isFinite(ts) && now - ts < PUSH_DISMISS_TTL_MS) {
			return false;
		}
		// TTL истёк — разрешаем баннер снова.
		localStorage.removeItem(PUSH_DISMISSED_KEY);
	}
	return true;
}

/**
 * Помечает, что пользователь нажал "Позже" — баннер скрыт на PUSH_DISMISS_TTL_MS.
 */
export function dismissPushBanner(): void {
	localStorage.setItem(PUSH_DISMISSED_KEY, Date.now().toString());
}

/**
 * Запрашивает разрешение на push-уведомления у браузера.
 * Возвращает true если разрешение получено.
 */
export async function requestPushPermission(): Promise<boolean> {
	if (!isPushSupported()) return false;

	const permission = await Notification.requestPermission();
	return permission === "granted";
}

/**
 * Подписывается на push-уведомления:
 * 1. Получает VAPID-ключ с бэкенда
 * 2. Создаёт PushSubscription через PushManager
 * 3. Отправляет ключи подписки на бэкенд
 *
 * @returns true если подписка успешна
 */
export async function subscribeToPushNotifications(): Promise<boolean> {
	try {
		// Получаем VAPID-ключ с бэкенда.
		const { vapid_public_key } = await getVAPIDKey();
		if (!vapid_public_key) {
			console.warn("VAPID ключ не настроен на сервере");
			return false;
		}

		// Получаем регистрацию Service Worker.
		const registration = await navigator.serviceWorker.ready;

		// Конвертируем VAPID-ключ из base64url в Uint8Array.
		const applicationServerKey = urlBase64ToUint8Array(vapid_public_key);

		// Создаём подписку через PushManager.
		const subscription = await registration.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey,
		});

		// Извлекаем ключи из подписки.
		const rawKey = subscription.getKey("p256dh");
		const rawAuth = subscription.getKey("auth");
		if (!rawKey || !rawAuth) {
			console.error("Не удалось получить ключи подписки");
			return false;
		}

		const p256dhKey = arrayBufferToBase64url(rawKey);
		const authKey = arrayBufferToBase64url(rawAuth);

		// Отправляем подписку на бэкенд.
		await subscribeToPush({
			endpoint: subscription.endpoint,
			p256dh_key: p256dhKey,
			auth_key: authKey,
		});

		localStorage.setItem(PUSH_SUBSCRIBED_KEY, "true");
		return true;
	} catch (err) {
		console.error("Ошибка подписки на push:", err);
		return false;
	}
}

/**
 * Конвертирует base64url строку в Uint8Array (для applicationServerKey).
 */
function urlBase64ToUint8Array(base64String: string): Uint8Array {
	const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
	const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
	const rawData = atob(base64);
	const outputArray = new Uint8Array(rawData.length);
	for (let i = 0; i < rawData.length; i++) {
		outputArray[i] = rawData.charCodeAt(i);
	}
	return outputArray;
}

/**
 * Конвертирует ArrayBuffer в base64url строку.
 */
function arrayBufferToBase64url(buffer: ArrayBuffer): string {
	const bytes = new Uint8Array(buffer);
	let binary = "";
	for (const byte of bytes) {
		binary += String.fromCharCode(byte);
	}
	return btoa(binary)
		.replace(/\+/g, "-")
		.replace(/\//g, "_")
		.replace(/=+$/, "");
}
