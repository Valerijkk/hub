import { create } from "zustand";
import { toast } from "sonner";
import {
  getNotifications,
  getUnreadCount,
  markRead as apiMarkRead,
  markAllRead as apiMarkAllRead,
  deleteNotification as apiDeleteNotification,
  type NotificationDTO,
} from "@/api/notifications";

/** Состояние стора уведомлений. */
interface NotificationState {
  /** Список уведомлений. */
  notifications: NotificationDTO[];
  /** Количество непрочитанных. */
  unreadCount: number;
  /** Загрузка списка. */
  loading: boolean;
  /** Текущая страница. */
  page: number;
  /** Общее количество. */
  total: number;
  /** Есть ли ещё страницы. */
  hasMore: boolean;

  /** Загрузить уведомления (первая страница). */
  fetchNotifications: () => Promise<void>;
  /** Загрузить следующую страницу. */
  fetchMore: () => Promise<void>;
  /** Добавить уведомление из WS. */
  addNotification: (notif: NotificationDTO) => void;
  /** Пометить прочитанным. */
  markRead: (id: string) => Promise<void>;
  /** Удалить уведомление. */
  deleteNotification: (id: string) => Promise<void>;
  /** Пометить все прочитанными. */
  markAllRead: () => Promise<void>;
  /** Загрузить количество непрочитанных. */
  fetchUnreadCount: () => Promise<void>;
}

const PER_PAGE = 20;

export const useNotificationStore = create<NotificationState>((set, get) => ({
  notifications: [],
  unreadCount: 0,
  loading: false,
  page: 1,
  total: 0,
  hasMore: false,

  fetchNotifications: async () => {
    set({ loading: true });
    try {
      const res = await getNotifications(1, PER_PAGE);
      set({
        notifications: res.notifications ?? [],
        total: res.total,
        page: 1,
        hasMore: (res.notifications?.length ?? 0) < res.total,
      });
    } catch {
      // Ошибка загрузки уведомлений — игнорируем.
    } finally {
      set({ loading: false });
    }
  },

  fetchMore: async () => {
    const { page, hasMore, loading, notifications } = get();
    if (!hasMore || loading) return;

    set({ loading: true });
    try {
      const nextPage = page + 1;
      const res = await getNotifications(nextPage, PER_PAGE);
      const newItems = res.notifications ?? [];
      const allItems = [...notifications, ...newItems];
      set({
        notifications: allItems,
        page: nextPage,
        total: res.total,
        hasMore: allItems.length < res.total,
      });
    } catch {
      // Ошибка загрузки — игнорируем.
    } finally {
      set({ loading: false });
    }
  },

  addNotification: (notif: NotificationDTO) => {
    set((state) => ({
      notifications: [notif, ...state.notifications],
      unreadCount: state.unreadCount + 1,
      total: state.total + 1,
    }));
    toast(notif.title, {
      description: notif.body,
    });
  },

  markRead: async (id: string) => {
    try {
      await apiMarkRead(id);
      set((state) => ({
        notifications: state.notifications.map((n) =>
          n.id === id ? { ...n, is_read: true } : n,
        ),
        unreadCount: Math.max(0, state.unreadCount - 1),
      }));
    } catch {
      // Ошибка — игнорируем.
    }
  },

  deleteNotification: async (id: string) => {
    try {
      await apiDeleteNotification(id);
      set((state) => {
        const notif = state.notifications.find((n) => n.id === id);
        return {
          notifications: state.notifications.filter((n) => n.id !== id),
          total: Math.max(0, state.total - 1),
          unreadCount: notif && !notif.is_read ? Math.max(0, state.unreadCount - 1) : state.unreadCount,
        };
      });
    } catch {
      // Ошибка удаления — игнорируем.
    }
  },

  markAllRead: async () => {
    try {
      await apiMarkAllRead();
      set((state) => ({
        notifications: state.notifications.map((n) => ({
          ...n,
          is_read: true,
        })),
        unreadCount: 0,
      }));
    } catch {
      // Ошибка — игнорируем.
    }
  },

  fetchUnreadCount: async () => {
    try {
      const res = await getUnreadCount();
      set({ unreadCount: res.count });
    } catch {
      // Ошибка — игнорируем.
    }
  },
}));
