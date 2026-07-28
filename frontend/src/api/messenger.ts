import { buildQueryString } from "@/lib/utils";
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
	is_starred: boolean;
	muted_until?: string;
	created_at: string;
	updated_at: string;
}

/** Участник канала (соответствует backend ChannelMemberResponse). */
export interface ChannelMember {
	id: string;
	user_id: string;
	user_name?: string;
	role: string;
	joined_at: string;
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
	/** Тип сообщения: обычное (пусто) или "poll". */
	message_type?: string;
	/** ID опроса (только для message_type="poll"). */
	poll_id?: string;
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

/**
 * Конструирует объект Message из WS-события (MessageSentEvent).
 * WS присылает частичные данные — заполняем недостающие поля безопасными дефолтами.
 */
export function buildMessageFromWsEvent(
	evt: Record<string, unknown>,
	fallbackChannelId: string,
): Message | null {
	const id = (evt.message_id ?? evt.id ?? "") as string;
	if (!id) return null;
	return {
		id,
		channel_id: (evt.channel_id ?? fallbackChannelId) as string,
		sender_id: (evt.sender_id ?? "") as string,
		message_type: (evt.message_type as string) ?? undefined,
		poll_id: (evt.poll_id as string) ?? undefined,
		content_text: (evt.content_text as string) ?? undefined,
		content_json: (evt.content_json as Record<string, unknown>) ?? undefined,
		reply_count: (evt.reply_count as number) ?? 0,
		created_at: (evt.created_at as string) ?? new Date().toISOString(),
		thread_id: (evt.thread_id as string) ?? undefined,
		reactions: (evt.reactions as Message["reactions"]) ?? undefined,
		attachments: (evt.attachments as Message["attachments"]) ?? undefined,
		forwarded_from_id: (evt.forwarded_from_id as string) ?? undefined,
		edited_at: (evt.edited_at as string) ?? undefined,
		deleted_at: (evt.deleted_at as string) ?? undefined,
		voice_url: (evt.voice_url as string) ?? undefined,
		voice_duration_ms: (evt.voice_duration_ms as number) ?? undefined,
		mentions: (evt.mentions as string[]) ?? undefined,
		mention_groups: (evt.mention_groups as string[]) ?? undefined,
	};
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
	/** ID упомянутых пользователей (@mentions). */
	mentions?: string[];
	/** Групповые упоминания (@all, @online, @role). */
	mention_groups?: string[];
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

/** Переключить закрепление (star) канала. */
export function starChannel(channelId: string) {
	return api.post<{ starred: boolean }>(
		`/messenger/channels/${channelId}/star`,
	);
}

/** Замьютить канал на указанную длительность. */
export function muteChannel(
	channelId: string,
	duration: "1h" | "2h" | "8h" | "1d" | "forever",
) {
	return api.post<{ muted_until: string }>(
		`/messenger/channels/${channelId}/mute`,
		{ duration },
	);
}

/** Снять мьют с канала. */
export function unmuteChannel(channelId: string) {
	return api.post<void>(`/messenger/channels/${channelId}/unmute`);
}

/** Удалить канал. */
export function deleteChannel(channelId: string) {
	return api.delete<void>(`/messenger/channels/${channelId}`);
}

/** Получить список участников канала. */
export function getChannelMembers(channelId: string) {
	return api.get<ChannelMember[]>(`/messenger/channels/${channelId}/members`);
}

// ---------------------------------------------------------------------------
// API-функции: Сообщения (Messages)
// ---------------------------------------------------------------------------

/** Получить список сообщений канала с курсорной пагинацией. */
export function getMessages(
	channelId: string,
	params?: { before?: string; limit?: number },
) {
	const qs = buildQueryString(params);
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
	const qs = buildQueryString(params);
	return api.get<Message[]>(
		`/messenger/channels/${channelId}/messages/${messageId}/replies${qs}`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Индикатор набора текста (Typing)
// ---------------------------------------------------------------------------

/** Отправить индикатор набора текста в канал. */
export function sendTypingIndicator(channelId: string) {
	return api.post<void>(`/messenger/channels/${channelId}/typing`);
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

/** Установить кастомный текст статуса с длительностью. */
export function setCustomStatus(text: string, duration?: string) {
	return api.put<void>("/messenger/presence/status", { text, duration });
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

/** Данные превью ссылки — контракт совпадает с backend / github.LinkPreviewResponse. */
export interface LinkPreviewData {
	url?: string;
	type?: string;
	/** Заголовок страницы (og:title). */
	title?: string;
	/** Описание страницы (og:description). */
	description?: string;
	/** URL изображения превью (og:image) — канон API. */
	image_url?: string;
	/** Домен сайта. */
	domain?: string;
	/** @deprecated используйте image_url */
	image?: string;
	/** @deprecated используйте domain */
	site_name?: string;
}

/** Получить превью ссылки по URL (OG-теги). */
export function fetchLinkPreview(url: string) {
	const params = new URLSearchParams({ url });
	return api.get<LinkPreviewData>(`/link-preview?${params.toString()}`);
}

// ---------------------------------------------------------------------------
// API-функции: Опросы (Polls)
// ---------------------------------------------------------------------------

/** Вариант ответа в опросе. */
export interface PollOption {
	id: string;
	text: string;
	position: number;
	count: number;
	/** ID проголосовавших (пусто при анонимном голосовании). */
	voter_ids?: string[];
	/** Проголосовал ли текущий пользователь за этот вариант. */
	voted: boolean;
}

/** Опрос (соответствует backend PollResponse). */
export interface Poll {
	id: string;
	channel_id: string;
	creator_id: string;
	question: string;
	is_anonymous: boolean;
	is_multi_choice: boolean;
	is_closed: boolean;
	total_votes: number;
	options: PollOption[];
	created_at: string;
}

/** Запрос на создание опроса. */
export interface CreatePollRequest {
	question: string;
	options: string[];
	is_anonymous?: boolean;
	is_multi_choice?: boolean;
}

/** Создать опрос в канале. */
export function createPoll(channelId: string, data: CreatePollRequest) {
	return api.post<Poll>(`/messenger/channels/${channelId}/polls`, data);
}

/** Проголосовать в опросе. */
export function votePoll(pollId: string, optionIds: string[]) {
	return api.post<Poll>(`/messenger/polls/${pollId}/vote`, {
		option_ids: optionIds,
	});
}

/** Закрыть опрос (только создатель). */
export function closePoll(pollId: string) {
	return api.post<Poll>(`/messenger/polls/${pollId}/close`);
}

/** Получить результаты опроса. */
export function getPollResults(pollId: string) {
	return api.get<Poll>(`/messenger/polls/${pollId}`);
}

// ---------------------------------------------------------------------------
// API-функции: Пригласительные ссылки (Channel Invites)
// ---------------------------------------------------------------------------

/** Пригласительная ссылка канала. */
export interface ChannelInvite {
	id: string;
	channel_id: string;
	creator_id: string;
	token: string;
	invite_url: string;
	max_uses: number;
	use_count: number;
	expires_at?: string;
	is_revoked: boolean;
	created_at: string;
}

/** Результат валидации пригласительной ссылки. */
export interface InviteValidation {
	valid: boolean;
	channel_name?: string;
	channel_type?: string;
	channel_id?: string;
	member_count?: number;
	message?: string;
}

/** Запрос на создание пригласительной ссылки. */
export interface CreateInviteLinkRequest {
	max_uses?: number;
	expires_in?: string;
}

/** Создать пригласительную ссылку для канала. */
export function createInviteLink(
	channelId: string,
	data: CreateInviteLinkRequest,
) {
	return api.post<ChannelInvite>(
		`/messenger/channels/${channelId}/invites`,
		data,
	);
}

/** Получить список пригласительных ссылок канала. */
export function getChannelInvites(channelId: string) {
	return api.get<ChannelInvite[]>(`/messenger/channels/${channelId}/invites`);
}

/** Валидировать пригласительную ссылку по токену. */
export function validateInvite(token: string) {
	return api.get<InviteValidation>(`/messenger/invites/${token}`);
}

/** Принять пригласительную ссылку и вступить в канал. */
export function acceptInvite(token: string) {
	return api.post<Channel>(`/messenger/invites/${token}/accept`);
}

/** Отозвать пригласительную ссылку. */
export function revokeInvite(channelId: string, inviteId: string) {
	return api.delete<void>(
		`/messenger/channels/${channelId}/invites/${inviteId}`,
	);
}
