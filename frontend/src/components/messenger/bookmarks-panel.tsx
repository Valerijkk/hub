/**
 * BookmarksPanel отображает список сохранённых сообщений пользователя.
 * Доступен из sidebar мессенджера. Позволяет удалять закладки и переходить к сообщениям.
 */
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useCallback, useLayoutEffect, useRef, useState } from "react";
import { type Bookmark, getBookmarks, removeBookmark } from "@/api/messenger";
import {
	QueryStateView,
	type QueryLike,
} from "@/components/query-state-view";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";

export const BOOKMARKS_QUERY_KEY = ["messenger", "bookmarks"] as const;

export function bookmarksQueryKey(userId: string) {
	return [...BOOKMARKS_QUERY_KEY, userId] as const;
}

type RemovalScope = { actorId: string | null };
type RemovalRequest = { scope: RemovalScope; bookmarkId: string };

/** Primary accessible label: channel name, never raw UUID (A3W9-2-001). */
function bookmarkTitle(bm: Bookmark): string {
	const name = bm.channel_name?.trim();
	return name || "Канал";
}

/** Subtitle preview: message text, never raw message_id UUID. */
function bookmarkSubtitle(bm: Bookmark): string {
	const preview = bm.content_preview?.trim();
	return preview || "Сообщение";
}

export function BookmarksPanel() {
	const navigate = useNavigate();
	const queryClient = useQueryClient();
	const userId = useAuthStore((state) => state.user?.id ?? null);
	const [pendingIds, setPendingIds] = useState<Set<string>>(() => new Set());
	const scopeRef = useRef<RemovalScope | null>(null);
	const requestsRef = useRef(new Map<string, RemovalRequest>());
	const queryKey = bookmarksQueryKey(userId ?? "__no-user__");

	useLayoutEffect(() => {
		const scope: RemovalScope = { actorId: userId };
		scopeRef.current = scope;
		requestsRef.current = new Map();
		setPendingIds(new Set());

		return () => {
			if (scopeRef.current === scope) {
				scopeRef.current = null;
				requestsRef.current = new Map();
			}
		};
	}, [userId]);

	const bookmarksQuery = useQuery({
		queryKey,
		queryFn: getBookmarks,
		enabled: Boolean(userId),
	});
	const bookmarksViewQuery: typeof bookmarksQuery | QueryLike<Bookmark[]> =
		userId ? bookmarksQuery : { isLoading: false, data: undefined };

	const isCurrent = useCallback((request: RemovalRequest) => {
		return (
			scopeRef.current === request.scope &&
			Boolean(request.scope.actorId) &&
			useAuthStore.getState().user?.id === request.scope.actorId &&
			requestsRef.current.get(request.bookmarkId) === request
		);
	}, []);

	const handleRemove = useCallback(async (bookmark: Bookmark): Promise<void> => {
		const scope = scopeRef.current;
		const capturedActorId = scope?.actorId;
		const bookmarkId = bookmark.id;
		if (
			!scope ||
			!capturedActorId ||
			!bookmarkId ||
			useAuthStore.getState().user?.id !== capturedActorId ||
			requestsRef.current.has(bookmarkId)
		) return;

		const request: RemovalRequest = { scope, bookmarkId };
		requestsRef.current.set(bookmarkId, request);
		setPendingIds((prior) => new Set(prior).add(bookmarkId));

		try {
			await removeBookmark(bookmarkId);
			if (isCurrent(request)) {
				await queryClient.invalidateQueries({
					queryKey: bookmarksQueryKey(capturedActorId),
					exact: true,
				});
			}
		} catch (err) {
			if (isCurrent(request)) {
				toastApiError(err, "Не удалось удалить закладку");
			}
		} finally {
			if (
				scopeRef.current === scope &&
				requestsRef.current.get(bookmarkId) === request
			) {
				requestsRef.current.delete(bookmarkId);
				setPendingIds((prior) => {
					const next = new Set(prior);
					next.delete(bookmarkId);
					return next;
				});
			}
		}
	}, [isCurrent, queryClient]);

	const handleNavigate = useCallback(
		(bookmark: Bookmark) => {
			if (!userId || useAuthStore.getState().user?.id !== userId) return;
			void navigate({
				to: "/messenger/$channelId",
				params: { channelId: bookmark.channel_id },
			});
		},
		[navigate, userId],
	);

	return (
		<div className="flex h-full flex-col">
			<div className="flex h-12 items-center border-b px-4">
				<span className="text-sm font-semibold">Закладки</span>
			</div>
			<ScrollArea className="flex-1">
				{/* W9-A2-009: isError + retry, не путать с «Нет закладок» */}
				<QueryStateView<Bookmark[]>
					query={bookmarksViewQuery}
					loadingFallback={
						<div className="flex items-center justify-center py-8">
							<span className="text-sm text-t2">Загрузка...</span>
						</div>
					}
					isEmpty={(data) => data.length === 0}
					emptyFallback={
						<p className="px-4 py-8 text-center text-sm text-t2">
							Нет закладок
						</p>
					}
				>
					{(bookmarks) => (
						<div className="space-y-2 p-4">
							{bookmarks.map((bm) => {
								const title = bookmarkTitle(bm);
								const subtitle = bookmarkSubtitle(bm);
								return (
									<div
										key={bm.id}
										className="flex items-start gap-3 rounded-md border p-3"
									>
										<button
											type="button"
											className="min-w-0 flex-1 text-left"
											aria-label={`${title}: ${subtitle}`}
											onClick={() => handleNavigate(bm)}
										>
											<p className="text-sm font-medium">{title}</p>
											<p className="mt-0.5 line-clamp-2 text-xs text-t2">
												{subtitle}
											</p>
											<p className="mt-1 text-xs text-t2">
												{new Date(bm.created_at).toLocaleDateString("ru-RU")}
											</p>
										</button>
										<Button
											variant="ghost"
											size="icon"
											className="h-7 w-7 shrink-0"
											disabled={pendingIds.has(bm.id)}
											title="Удалить из закладок"
											aria-label="Удалить из закладок"
											onClick={() => void handleRemove(bm)}
										>
											<svg
												aria-hidden="true"
												className="h-3.5 w-3.5"
												fill="none"
												viewBox="0 0 24 24"
												stroke="currentColor"
												strokeWidth={2}
											>
												<path
													strokeLinecap="round"
													strokeLinejoin="round"
													d="M6 18L18 6M6 6l12 12"
												/>
											</svg>
										</Button>
									</div>
								);
							})}
						</div>
					)}
				</QueryStateView>
			</ScrollArea>
		</div>
	);
}
