import { create } from "zustand";

import type { Channel, Message } from "@/api/messenger";

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
  unreadCounts: Record<string, number>;
  /** ID сообщения, для которого открыт тред. */
  activeThreadMessageId: string | null;
  /** Сообщение, на которое отвечает пользователь (reply). */
  replyToMessage: Message | null;
  /** Печатающие пользователи по каналам. */
  typingUsers: Record<string, TypingUser[]>;

  /** Установить активный канал. */
  setActiveChannel: (channelId: string | null) => void;
  /** Установить полный список каналов. */
  setChannels: (channels: Channel[]) => void;
  /** Добавить канал в список. */
  addChannel: (channel: Channel) => void;
  /** Обновить канал в списке. */
  updateChannel: (channelId: string, partial: Partial<Channel>) => void;
  /** Обновить количество непрочитанных для канала. */
  updateUnread: (channelId: string, count: number) => void;
  /** Увеличить счётчик непрочитанных на 1. */
  incrementUnread: (channelId: string) => void;
  /** Сбросить счётчик непрочитанных для канала. */
  resetUnread: (channelId: string) => void;
  /** Открыть тред по ID сообщения. */
  openThread: (messageId: string) => void;
  /** Закрыть тред. */
  closeThread: () => void;
  /** Установить сообщение для ответа (reply). */
  setReplyTo: (message: Message | null) => void;
  /** Добавить печатающего пользователя в канал. */
  addTypingUser: (channelId: string, userId: string, displayName: string) => void;
  /** Убрать печатающего пользователя из канала. */
  removeTypingUser: (channelId: string, userId: string) => void;
}

/** Стор управления состоянием мессенджера. */
export const useMessengerStore = create<MessengerState>()((set) => ({
  activeChannelId: null,
  channels: [],
  unreadCounts: {},
  activeThreadMessageId: null,
  replyToMessage: null,
  typingUsers: {},

  setActiveChannel: (channelId) => set({ activeChannelId: channelId }),

  setChannels: (channels) => set({ channels }),

  addChannel: (channel) =>
    set((state) => ({ channels: [channel, ...state.channels] })),

  updateChannel: (channelId, partial) =>
    set((state) => ({
      channels: state.channels.map((ch) =>
        ch.id === channelId ? { ...ch, ...partial } : ch,
      ),
    })),

  updateUnread: (channelId, count) =>
    set((state) => ({
      unreadCounts: { ...state.unreadCounts, [channelId]: count },
    })),

  incrementUnread: (channelId) =>
    set((state) => ({
      unreadCounts: {
        ...state.unreadCounts,
        [channelId]: (state.unreadCounts[channelId] ?? 0) + 1,
      },
    })),

  resetUnread: (channelId) =>
    set((state) => ({
      unreadCounts: { ...state.unreadCounts, [channelId]: 0 },
    })),

  openThread: (messageId) => set({ activeThreadMessageId: messageId }),

  closeThread: () => set({ activeThreadMessageId: null }),

  setReplyTo: (message) => set({ replyToMessage: message }),

  addTypingUser: (channelId, userId, displayName) =>
    set((state) => {
      const current = state.typingUsers[channelId] ?? [];
      const filtered = current.filter((u) => u.userId !== userId);
      return {
        typingUsers: {
          ...state.typingUsers,
          [channelId]: [...filtered, { userId, displayName, timestamp: Date.now() }],
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
}));
