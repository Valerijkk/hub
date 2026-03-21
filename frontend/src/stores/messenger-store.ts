import { create } from "zustand";

import type { Channel } from "@/api/messenger";

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
}

/** Стор управления состоянием мессенджера. */
export const useMessengerStore = create<MessengerState>()((set) => ({
  activeChannelId: null,
  channels: [],
  unreadCounts: {},
  activeThreadMessageId: null,

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
}));
