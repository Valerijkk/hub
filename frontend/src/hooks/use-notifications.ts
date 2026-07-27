import {
	type InfiniteData,
	useInfiniteQuery,
	useMutation,
	useQuery,
	useQueryClient,
} from "@tanstack/react-query";
import { useCallback, useEffect, useRef } from "react";
import { toast } from "sonner";
import {
	deleteNotification as apiDeleteNotification,
	markAllRead as apiMarkAllRead,
	markRead as apiMarkRead,
	getNotifications,
	getUnreadCount,
	type NotificationDTO,
	type NotificationListResponse,
} from "@/api/notifications";
import { toastApiError } from "@/lib/format-api-error";
import { useWsStore } from "@/stores/ws-store";

/** Ключ списка уведомлений (InfiniteQuery). */
export const NOTIFICATIONS_LIST_KEY = ["notifications", "list"] as const;
/** Ключ счётчика непрочитанных. */
export const NOTIFICATIONS_UNREAD_KEY = ["notifications", "unread"] as const;

const PER_PAGE = 20;

type ListData = InfiniteData<NotificationListResponse, number>;

/** Сплющить страницы InfiniteQuery в один список. */
export function flattenNotificationPages(
	data: InfiniteData<NotificationListResponse> | undefined,
): NotificationDTO[] {
	if (!data?.pages) return [];
	return data.pages.flatMap((p) => p.notifications ?? []);
}

/**
 * Вставляет WS-уведомление в начало первой страницы кэша (без дубликатов).
 * Экспортировано для unit-тестов кэш-хелпера.
 */
export function prependNotification(
	data: ListData | undefined,
	notif: NotificationDTO,
): ListData {
	if (!data?.pages.length) {
		return {
			pages: [
				{
					notifications: [notif],
					total: 1,
					page: 1,
					per_page: PER_PAGE,
				},
			],
			pageParams: [1],
		};
	}
	const exists = data.pages.some((p) =>
		(p.notifications ?? []).some((n) => n.id === notif.id),
	);
	if (exists) return data;

	const first = data.pages[0];
	if (!first) {
		return {
			pages: [
				{
					notifications: [notif],
					total: 1,
					page: 1,
					per_page: PER_PAGE,
				},
			],
			pageParams: [1],
		};
	}
	return {
		...data,
		pages: [
			{
				...first,
				notifications: [notif, ...(first.notifications ?? [])],
				total: (first.total ?? 0) + 1,
				page: first.page ?? 1,
				per_page: first.per_page ?? PER_PAGE,
			},
			...data.pages.slice(1),
		],
	};
}

function patchNotificationRead(
	data: InfiniteData<NotificationListResponse> | undefined,
	id: string,
): InfiniteData<NotificationListResponse> | undefined {
	if (!data) return data;
	return {
		...data,
		pages: data.pages.map((page) => ({
			...page,
			notifications: (page.notifications ?? []).map((n) =>
				n.id === id ? { ...n, is_read: true } : n,
			),
		})),
	};
}

function removeNotification(
	data: InfiniteData<NotificationListResponse> | undefined,
	id: string,
): InfiniteData<NotificationListResponse> | undefined {
	if (!data) return data;
	return {
		...data,
		pages: data.pages.map((page) => ({
			...page,
			notifications: (page.notifications ?? []).filter((n) => n.id !== id),
			total: Math.max(0, (page.total ?? 0) - 1),
		})),
	};
}

function markAllPagesRead(
	data: InfiniteData<NotificationListResponse> | undefined,
): InfiniteData<NotificationListResponse> | undefined {
	if (!data) return data;
	return {
		...data,
		pages: data.pages.map((page) => ({
			...page,
			notifications: (page.notifications ?? []).map((n) => ({
				...n,
				is_read: true,
			})),
		})),
	};
}

/**
 * Список уведомлений через useInfiniteQuery (offset/page).
 * Ошибки — в query.isError (UI), без silent console warn.
 */
export function useNotificationList(enabled = true) {
	return useInfiniteQuery({
		queryKey: NOTIFICATIONS_LIST_KEY,
		queryFn: ({ pageParam }) => getNotifications(pageParam, PER_PAGE),
		initialPageParam: 1,
		getNextPageParam: (lastPage, allPages) => {
			const loaded = allPages.reduce(
				(sum, p) => sum + (p.notifications?.length ?? 0),
				0,
			);
			if (loaded >= lastPage.total) return undefined;
			return (lastPage.page ?? allPages.length) + 1;
		},
		enabled,
	});
}

/** Счётчик непрочитанных через useQuery. */
export function useUnreadNotificationCount() {
	return useQuery({
		queryKey: NOTIFICATIONS_UNREAD_KEY,
		queryFn: async () => {
			const res = await getUnreadCount();
			return res.count;
		},
	});
}

/**
 * Мутации уведомлений с патчем Query-кэша и toastApiError (не silent warn).
 */
export function useNotificationMutations() {
	const queryClient = useQueryClient();

	const markRead = useMutation({
		mutationFn: (id: string) => apiMarkRead(id),
		onSuccess: (_void, id) => {
			const list = queryClient.getQueryData<
				InfiniteData<NotificationListResponse>
			>(NOTIFICATIONS_LIST_KEY);
			const wasUnread = flattenNotificationPages(list).some(
				(n) => n.id === id && !n.is_read,
			);
			queryClient.setQueryData(
				NOTIFICATIONS_LIST_KEY,
				(old: InfiniteData<NotificationListResponse> | undefined) =>
					patchNotificationRead(old, id),
			);
			if (wasUnread) {
				queryClient.setQueryData<number>(NOTIFICATIONS_UNREAD_KEY, (c) =>
					Math.max(0, (c ?? 0) - 1),
				);
			}
		},
		onError: (err) => toastApiError(err, "Не удалось отметить уведомление"),
	});

	const deleteOne = useMutation({
		mutationFn: (id: string) => apiDeleteNotification(id),
		onSuccess: (_void, id) => {
			const list = queryClient.getQueryData<
				InfiniteData<NotificationListResponse>
			>(NOTIFICATIONS_LIST_KEY);
			const target = flattenNotificationPages(list).find((n) => n.id === id);
			queryClient.setQueryData(
				NOTIFICATIONS_LIST_KEY,
				(old: InfiniteData<NotificationListResponse> | undefined) =>
					removeNotification(old, id),
			);
			if (target && !target.is_read) {
				queryClient.setQueryData<number>(NOTIFICATIONS_UNREAD_KEY, (c) =>
					Math.max(0, (c ?? 0) - 1),
				);
			}
		},
		onError: (err) => toastApiError(err, "Не удалось удалить уведомление"),
	});

	const markAll = useMutation({
		mutationFn: () => apiMarkAllRead(),
		onSuccess: () => {
			queryClient.setQueryData(
				NOTIFICATIONS_LIST_KEY,
				(old: InfiniteData<NotificationListResponse> | undefined) =>
					markAllPagesRead(old),
			);
			queryClient.setQueryData(NOTIFICATIONS_UNREAD_KEY, 0);
		},
		onError: (err) => toastApiError(err, "Не удалось отметить все уведомления"),
	});

	/** Локальный clear после mark-all — прежнее UX «Удалить все». */
	const clearAll = useMutation({
		mutationFn: () => apiMarkAllRead(),
		onSuccess: () => {
			queryClient.setQueryData<ListData>(NOTIFICATIONS_LIST_KEY, {
				pages: [
					{
						notifications: [],
						total: 0,
						page: 1,
						per_page: PER_PAGE,
					},
				],
				pageParams: [1],
			});
			queryClient.setQueryData(NOTIFICATIONS_UNREAD_KEY, 0);
		},
		onError: (err) => toastApiError(err, "Не удалось очистить уведомления"),
	});

	return {
		markRead: (id: string) => markRead.mutateAsync(id),
		deleteNotification: (id: string) => deleteOne.mutateAsync(id),
		markAllRead: () => markAll.mutateAsync(),
		clearAll: () => clearAll.mutateAsync(),
	};
}

/**
 * WS listener: notification.new → toast + prepend в Query + bump unread.
 * Без zustand server-state.
 */
export function useNotificationRealtime() {
	const queryClient = useQueryClient();
	const addListener = useWsStore((s) => s.addListener);
	const removeListener = useWsStore((s) => s.removeListener);

	const handleRef = useRef<(notif: NotificationDTO) => void>(() => {});
	handleRef.current = (notif: NotificationDTO) => {
		queryClient.setQueryData(
			NOTIFICATIONS_LIST_KEY,
			(old: ListData | undefined) => prependNotification(old, notif),
		);
		queryClient.setQueryData<number>(
			NOTIFICATIONS_UNREAD_KEY,
			(c) => (c ?? 0) + 1,
		);
		toast(notif.title, { description: notif.body });
	};

	const handler = useCallback((event: string, payload: unknown) => {
		if (event === "notification.new" && payload) {
			handleRef.current(payload as NotificationDTO);
		}
	}, []);

	useEffect(() => {
		addListener(handler);
		return () => {
			removeListener(handler);
		};
	}, [addListener, removeListener, handler]);
}
