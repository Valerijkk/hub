import { api } from "@/api/client";

/** DTO уведомления. */
export interface NotificationDTO {
	id: string;
	type: string;
	title: string;
	body?: string;
	entity_type?: string;
	entity_id?: string;
	project_id?: string;
	actor_id?: string;
	is_read: boolean;
	is_urgent: boolean;
	created_at: string;
}

/** Ответ со списком уведомлений с пагинацией. */
export interface NotificationListResponse {
	notifications: NotificationDTO[];
	total: number;
	page: number;
	per_page: number;
}

/** Ответ с количеством непрочитанных уведомлений. */
export interface UnreadCountResponse {
	count: number;
}

/** DTO настройки уведомлений. */
export interface PreferenceDTO {
	id: string;
	notification_type: string;
	channel: string;
	enabled: boolean;
	project_id?: string;
}

/** Запрос обновления настройки уведомлений. */
export interface PreferenceUpdateRequest {
	notification_type: string;
	channel: string;
	enabled: boolean;
	project_id?: string;
}

export interface TelegramBindResponse {
	bind_token: string;
	deep_link?: string;
	expires_in: number;
	bot_enabled: boolean;
}

export interface TelegramStatusResponse {
	linked: boolean;
	telegram_id?: string;
	username?: string;
}

export function createTelegramBind() {
	return api.post<TelegramBindResponse>("/notifications/telegram/bind");
}

export function confirmTelegramLink(code: string) {
	return api.post<{ message: string }>("/notifications/telegram/confirm", {
		code,
	});
}

export function unlinkTelegram() {
	return api.delete<void>("/notifications/telegram");
}

export function sendTelegramTest() {
	return api.post<{ message: string }>("/notifications/telegram/test");
}

export function getTelegramStatus() {
	return api.get<TelegramStatusResponse>("/notifications/telegram/status");
}

/**
 * Получить список уведомлений с пагинацией.
 * @param page - номер страницы
 * @param perPage - количество на страницу
 * @param isRead - фильтр по прочитанности (опционально)
 */
export function getNotifications(
	page: number = 1,
	perPage: number = 20,
	isRead?: boolean,
) {
	let path = `/notifications?page=${page}&per_page=${perPage}`;
	if (isRead !== undefined) {
		path += `&is_read=${isRead}`;
	}
	return api.get<NotificationListResponse>(path);
}

/** Получить количество непрочитанных уведомлений. */
export function getUnreadCount() {
	return api.get<UnreadCountResponse>("/notifications/unread-count");
}

/** Пометить уведомление как прочитанное. */
export function markRead(id: string) {
	return api.patch<void>(`/notifications/${id}/read`);
}

/** Пометить все уведомления как прочитанные. */
export function markAllRead() {
	return api.patch<void>("/notifications/read-all");
}

/** Удалить уведомление. */
export function deleteNotification(id: string) {
	return api.delete<void>(`/notifications/${id}`);
}

/** Получить настройки уведомлений. */
export function getPreferences() {
	return api.get<PreferenceDTO[]>("/notifications/preferences");
}

/** Обновить настройку уведомлений. */
export function updatePreference(data: PreferenceUpdateRequest) {
	return api.put<PreferenceDTO>("/notifications/preferences", data);
}

// --- Web Push API ---

/** Ответ с публичным VAPID-ключом. */
export interface VAPIDKeyResponse {
	vapid_public_key: string;
}

/** Запрос на подписку push-уведомлений. */
export interface PushSubscribeRequest {
	endpoint: string;
	p256dh_key: string;
	auth_key: string;
}

/** Получить публичный VAPID-ключ. */
export function getVAPIDKey() {
	return api.get<VAPIDKeyResponse>("/notifications/push/vapid-key");
}

/** Подписаться на push-уведомления. */
export function subscribeToPush(data: PushSubscribeRequest) {
	return api.post<void>("/notifications/push/subscribe", data);
}

/** Отписаться от push-уведомлений. */
export function unsubscribeFromPush(endpoint: string) {
	return api.request<void>("DELETE", "/notifications/push/unsubscribe", {
		endpoint,
	});
}
