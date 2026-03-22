import { api } from "./client";

// ---------------------------------------------------------------------------
// Типы перечислений
// ---------------------------------------------------------------------------

/** Тип канала мессенджера. */
export type ChannelType = "dm" | "group" | "private" | "public";

/** Тип вложения сообщения. */
export type AttachmentType = "file" | "image" | "voice";

// ---------------------------------------------------------------------------
// Интерфейсы ответов (Response DTOs)
// ---------------------------------------------------------------------------

/** Канал мессенджера (соответствует backend ChannelResponse). */
export interface Channel {
	id: string;
	channel_type: ChannelType;
	name: string;
	description?: string;
	is_readonly: boolean;
	project_id?: string;
	created_by: string;
	created_at: string;
	updated_at: string;
}

/** Реакция на сообщение (raw от backend — содержит user_ids). */
export interface MessageReaction {
	emoji: string;
	user_ids: string[];
}

/** Вложение сообщения (raw от backend). */
export interface AttachmentRef {
	file_id: string;
	file_name: string;
	file_size: number;
	mime_type: string;
	url: string;
}

/** Сообщение мессенджера (соответствует backend MessageResponse). */
export interface Message {
	id: string;
	channel_id: string;
	sender_id: string;
	content_json?: Record<string, unknown>;
	content_text?: string;
	voice_url?: string;
	voice_duration_ms?: number;
	attachments?: AttachmentRef[];
	thread_id?: string;
	reply_count: number;
	reactions?: MessageReaction[];
	forwarded_from_id?: string;
	mentions?: string[];
	mention_groups?: string[];
	edited_at?: string;
	deleted_at?: string;
	created_at: string;
}

/** Закладка на сообщение (соответствует backend BookmarkResponse). */
export interface Bookmark {
	id: string;
	message_id: string;
	channel_id: string;
	created_at: string;
}

// ---------------------------------------------------------------------------
// Интерфейсы запросов (Request DTOs)
// ---------------------------------------------------------------------------

/** Запрос на создание канала. */
export interface CreateChannelRequest {
	name: string;
	channel_type: ChannelType;
	description?: string;
	member_ids?: string[];
	project_id?: string;
}

/** Запрос на отправку сообщения. */
export interface SendMessageRequest {
	content_json: Record<string, unknown>;
	content_text?: string;
	voice_url?: string;
	voice_duration_ms?: number;
	thread_id?: string;
	attachment_ids?: string[];
}

/** Запрос на редактирование сообщения. */
export interface EditMessageRequest {
	content_json: Record<string, unknown>;
	content_text: string;
}

// ---------------------------------------------------------------------------
// API-функции: Каналы (Channels)
// ---------------------------------------------------------------------------

/** Получить список каналов текущего пользователя. */
export function getChannels() {
	return api.get<Channel[]>("/messenger/channels");
}

/** Получить информацию о канале по ID. */
export function getChannel(id: string) {
	return api.get<Channel>(`/messenger/channels/${id}`);
}

/** Создать новый канал. */
export function createChannel(data: CreateChannelRequest) {
	return api.post<Channel>("/messenger/channels", data);
}

// ---------------------------------------------------------------------------
// API-функции: Сообщения (Messages)
// ---------------------------------------------------------------------------

/** Получить список сообщений канала с курсорной пагинацией. */
export function getMessages(
	channelId: string,
	params?: { before?: string; limit?: number },
) {
	const qs = params
		? "?" +
			new URLSearchParams(
				Object.entries(params)
					.filter(([, v]) => v != null)
					.map(([k, v]) => [k, String(v)]),
			).toString()
		: "";
	return api.get<Message[]>(`/messenger/channels/${channelId}/messages${qs}`);
}

/** Отправить сообщение в канал. */
export function sendMessage(channelId: string, data: SendMessageRequest) {
	return api.post<Message>(`/messenger/channels/${channelId}/messages`, data);
}

/** Редактировать сообщение. */
export function editMessage(
	channelId: string,
	messageId: string,
	data: EditMessageRequest,
) {
	return api.put<Message>(
		`/messenger/channels/${channelId}/messages/${messageId}`,
		data,
	);
}

/** Удалить сообщение. */
export function deleteMessage(channelId: string, messageId: string) {
	return api.delete<void>(
		`/messenger/channels/${channelId}/messages/${messageId}`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Реакции (Reactions)
// ---------------------------------------------------------------------------

/** Переключить реакцию на сообщение (добавить/убрать). */
export function toggleReaction(
	channelId: string,
	messageId: string,
	emoji: string,
) {
	return api.post<{ added: boolean }>(
		`/messenger/channels/${channelId}/messages/${messageId}/reactions`,
		{ emoji },
	);
}

// ---------------------------------------------------------------------------
// API-функции: Закрепление (Pins)
// ---------------------------------------------------------------------------

/** Переключить закрепление сообщения. */
export function togglePin(channelId: string, messageId: string) {
	return api.post<{ pinned: boolean }>(
		`/messenger/channels/${channelId}/messages/${messageId}/pin`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Пересылка (Forward)
// ---------------------------------------------------------------------------

/** Переслать сообщение в другой канал. */
export function forwardMessage(
	channelId: string,
	messageId: string,
	targetChannelId: string,
) {
	return api.post<Message>(
		`/messenger/channels/${channelId}/messages/${messageId}/forward`,
		{ target_channel_id: targetChannelId },
	);
}

// ---------------------------------------------------------------------------
// API-функции: Треды (Threads)
// ---------------------------------------------------------------------------

/** Получить ответы в треде (с курсорной пагинацией). */
export function getThreadReplies(
	channelId: string,
	messageId: string,
	params?: { before?: string; limit?: number },
) {
	const qs = params
		? "?" +
			new URLSearchParams(
				Object.entries(params)
					.filter(([, v]) => v != null)
					.map(([k, v]) => [k, String(v)]),
			).toString()
		: "";
	return api.get<Message[]>(
		`/messenger/channels/${channelId}/messages/${messageId}/replies${qs}`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Личные сообщения (DM)
// ---------------------------------------------------------------------------

/** Создать или получить существующий DM-канал с пользователем. */
export function createOrGetDM(targetUserId: string) {
	return api.post<Channel>("/messenger/dm", {
		target_user_id: targetUserId,
	});
}

// ---------------------------------------------------------------------------
// API-функции: Поиск (Search)
// ---------------------------------------------------------------------------

/** Поиск сообщений по тексту. */
export function searchMessages(query: string, channelId?: string) {
	const params = new URLSearchParams({ query });
	if (channelId) params.set("channel_id", channelId);
	return api.get<Message[]>(`/messenger/search?${params.toString()}`);
}

// ---------------------------------------------------------------------------
// API-функции: Присутствие (Presence)
// ---------------------------------------------------------------------------

/** Статус присутствия: online, offline или away. */
export type PresenceStatus = "online" | "offline" | "away";

/** Получить статусы присутствия для списка пользователей. */
export function getPresence(userIds: string[]) {
	const params = new URLSearchParams({ user_ids: userIds.join(",") });
	return api.get<Record<string, PresenceStatus>>(
		`/messenger/presence?${params.toString()}`,
	);
}

/** Отправить heartbeat присутствия. */
export function sendHeartbeat() {
	return api.post<void>("/messenger/presence/heartbeat");
}

/** Установить кастомный текст статуса. */
export function setCustomStatus(text: string) {
	return api.put<void>("/messenger/presence/status", { text });
}

// ---------------------------------------------------------------------------
// API-функции: Непрочитанные (Unread)
// ---------------------------------------------------------------------------

/** Получить счётчики непрочитанных для всех каналов. */
export function getUnreadCounts() {
	return api.get<Record<string, number>>("/messenger/unread");
}

/** Отметить канал как прочитанный. */
export function markChannelRead(channelId: string) {
	return api.post<void>(`/messenger/channels/${channelId}/read`);
}

// ---------------------------------------------------------------------------
// API-функции: Закреплённые сообщения (Pins)
// ---------------------------------------------------------------------------

/** Закреплённое сообщение. */
export interface PinnedMessage {
	id: string;
	message_id: string;
	channel_id: string;
	pinned_by: string;
	pinned_at: string;
}

/** Получить список закреплённых сообщений канала. */
export function getChannelPins(channelId: string) {
	return api.get<PinnedMessage[]>(`/messenger/channels/${channelId}/pins`);
}

// ---------------------------------------------------------------------------
// API-функции: Закладки (Bookmarks)
// ---------------------------------------------------------------------------

/** Получить список закладок текущего пользователя. */
export function getBookmarks() {
	return api.get<Bookmark[]>("/messenger/bookmarks");
}

/** Добавить сообщение в закладки. */
export function addBookmark(messageId: string, channelId: string) {
	return api.post<void>("/messenger/bookmarks", {
		message_id: messageId,
		channel_id: channelId,
	});
}

/** Удалить закладку. */
export function removeBookmark(bookmarkId: string) {
	return api.delete<void>(`/messenger/bookmarks/${bookmarkId}`);
}

// ---------------------------------------------------------------------------
// API-функции: Превью ссылок (Link Preview)
// ---------------------------------------------------------------------------

/** Данные превью ссылки, полученные из OG-тегов страницы. */
export interface LinkPreviewData {
	/** Заголовок страницы (og:title). */
	title?: string;
	/** Описание страницы (og:description). */
	description?: string;
	/** URL изображения превью (og:image). */
	image?: string;
	/** Имя сайта (og:site_name). */
	site_name?: string;
}

/** Получить превью ссылки по URL (OG-теги). */
export function fetchLinkPreview(url: string) {
	const params = new URLSearchParams({ url });
	return api.get<LinkPreviewData>(`/link-preview?${params.toString()}`);
}
