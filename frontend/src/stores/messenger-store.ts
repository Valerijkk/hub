import { create } from "zustand";

import type { Channel, Message, UnreadSnapshot } from "@/api/messenger";

export interface ActiveThread {
	channelId: string;
	rootId: string;
}

export function emptyUnreadSnapshot(): UnreadSnapshot {
	return {
		channels: {},
		dms: {},
		mentions: {},
		threads: {},
		any_unread: false,
	};
}

/** Информация о печатающем пользователе. */
interface TypingUser {
	/** ID пользователя. */
	userId: string;
	/** Имя пользователя для отображения. */
	displayName: string;
	/** Время начала печати (для автоочистки). */
	timestamp: number;
}

/** Состояние стора мессенджера. */
interface MessengerState {
	/** ID активного канала. */
	activeChannelId: string | null;
	/** Список каналов пользователя. */
	channels: Channel[];
	/** Счётчики непрочитанных сообщений по каналам. */
	unreadSnapshot: UnreadSnapshot;
	/** ID сообщения, для которого открыт тред. */
	activeThread: ActiveThread | null;
	/** Сообщение, на которое отвечает пользователь (reply). */
	replyToMessage: Message | null;
	/** Печатающие пользователи по каналам. */
	typingUsers: Record<string, TypingUser[]>;
	/** Каналы, отозванные до подтверждённой серверной повторной авторизации. */
	revokedChannelIds: ReadonlySet<string>;

	/** Установить активный канал. */
	setActiveChannel: (channelId: string | null) => void;
	/** Установить полный список каналов. */
	setChannels: (channels: Channel[]) => void;
	/** Добавить канал в список. */
	addChannel: (channel: Channel) => void;
	/** Обновить канал в списке. */
	updateChannel: (channelId: string, partial: Partial<Channel>) => void;
	/** Полностью заменить четыре категории ответом сервера. */
	setUnreadSnapshot: (snapshot: UnreadSnapshot) => void;
	/** Открыть тред по ID сообщения. */
	openThread: (channelId: string, rootId: string) => void;
	/** Закрыть тред. */
	closeThread: () => void;
	/** Установить сообщение для ответа (reply). */
	setReplyTo: (message: Message | null) => void;
	/** Добавить печатающего пользователя в канал. */
	addTypingUser: (
		channelId: string,
		userId: string,
		displayName: string,
	) => void;
	/** Убрать печатающего пользователя из канала. */
	removeTypingUser: (channelId: string, userId: string) => void;
	/** Немедленно закрыть локальный доступ к каналу. */
	revokeChannel: (channelId: string) => void;
	/** Снять барьер после серверно-авторизованного channel_created. */
	restoreChannelAccess: (channelId: string) => void;
	isChannelRevoked: (channelId: string) => boolean;
	reset: () => void;
}

/** Стор управления состоянием мессенджера. */
export const useMessengerStore = create<MessengerState>()((set, get) => ({
	activeChannelId: null,
	channels: [],
	unreadSnapshot: emptyUnreadSnapshot(),
	activeThread: null,
	replyToMessage: null,
	typingUsers: {},
	revokedChannelIds: new Set<string>(),

	reset: () =>
		set({
			activeChannelId: null,
			channels: [],
			unreadSnapshot: emptyUnreadSnapshot(),
			activeThread: null,
			replyToMessage: null,
			typingUsers: {},
			revokedChannelIds: new Set<string>(),
		}),

	setActiveChannel: (channelId) => set({ activeChannelId: channelId }),

	setChannels: (channels) =>
		set((state) => ({
			channels: channels.filter(
				(channel) => !state.revokedChannelIds.has(channel.id),
			),
		})),

	addChannel: (channel) =>
		set((state) => ({ channels: [channel, ...state.channels] })),

	updateChannel: (channelId, partial) =>
		set((state) => ({
			channels: state.channels.map((ch) =>
				ch.id === channelId ? { ...ch, ...partial } : ch,
			),
		})),

	setUnreadSnapshot: (snapshot) =>
		set({
			unreadSnapshot: {
				channels: { ...snapshot.channels },
				dms: { ...snapshot.dms },
				mentions: { ...snapshot.mentions },
				threads: { ...snapshot.threads },
				any_unread: snapshot.any_unread,
			},
		}),

	openThread: (channelId, rootId) =>
		set({ activeThread: { channelId, rootId } }),

	closeThread: () => set({ activeThread: null }),

	setReplyTo: (message) => set({ replyToMessage: message }),

	addTypingUser: (channelId, userId, displayName) =>
		set((state) => {
			if (state.revokedChannelIds.has(channelId)) return state;
			const current = state.typingUsers[channelId] ?? [];
			const filtered = current.filter((u) => u.userId !== userId);
			return {
				typingUsers: {
					...state.typingUsers,
					[channelId]: [
						...filtered,
						{ userId, displayName, timestamp: Date.now() },
					],
				},
			};
		}),

	removeTypingUser: (channelId, userId) =>
		set((state) => {
			const current = state.typingUsers[channelId] ?? [];
			return {
				typingUsers: {
					...state.typingUsers,
					[channelId]: current.filter((u) => u.userId !== userId),
				},
			};
		}),
	revokeChannel: (channelId) =>
		set((state) => {
			const typingUsers = { ...state.typingUsers };
			delete typingUsers[channelId];
			return {
				activeChannelId:
					state.activeChannelId === channelId ? null : state.activeChannelId,
				channels: state.channels.filter((channel) => channel.id !== channelId),
				unreadSnapshot: emptyUnreadSnapshot(),
				activeThread:
					state.activeThread?.channelId === channelId ? null : state.activeThread,
				replyToMessage:
					state.replyToMessage?.channel_id === channelId
						? null
						: state.replyToMessage,
				typingUsers,
				revokedChannelIds: new Set(state.revokedChannelIds).add(channelId),
			};
		}),

	restoreChannelAccess: (channelId) =>
		set((state) => {
			if (!state.revokedChannelIds.has(channelId)) return state;
			const revokedChannelIds = new Set(state.revokedChannelIds);
			revokedChannelIds.delete(channelId);
			return { revokedChannelIds };
		}),

	isChannelRevoked: (channelId) => get().revokedChannelIds.has(channelId),
}));
