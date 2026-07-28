import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useCallback } from "react";
import { type Bookmark, getBookmarks, removeBookmark } from "@/api/messenger";
import { QueryStateView } from "@/components/query-state-view";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { toastApiError } from "@/lib/format-api-error";

/**
 * BookmarksPanel отображает список сохранённых сообщений пользователя.
 * Доступен из sidebar мессенджера. Позволяет удалять закладки и переходить к сообщениям.
 */
export function BookmarksPanel() {
	const navigate = useNavigate();
	const queryClient = useQueryClient();

	const bookmarksQuery = useQuery({
		queryKey: ["messenger", "bookmarks"],
		queryFn: getBookmarks,
	});

	const handleRemove = useCallback(
		async (bookmark: Bookmark) => {
			try {
				await removeBookmark(bookmark.id);
				await queryClient.invalidateQueries({
					queryKey: ["messenger", "bookmarks"],
				});
			} catch (err) {
				toastApiError(err, "Не удалось удалить закладку");
			}
		},
		[queryClient],
	);

	const handleNavigate = useCallback(
		(bookmark: Bookmark) => {
			void navigate({
				to: "/messenger/$channelId",
				params: { channelId: bookmark.channel_id },
			});
		},
		[navigate],
	);

	return (
		<div className="flex h-full flex-col">
			<div className="flex h-12 items-center border-b px-4">
				<span className="text-sm font-semibold">Закладки</span>
			</div>
			<ScrollArea className="flex-1">
				{/* W9-A2-009: isError + retry, не путать с «Нет закладок» */}
				<QueryStateView
					query={bookmarksQuery}
					loadingFallback={
						<div className="flex items-center justify-center py-8">
							<span className="text-sm text-t2">Загрузка...</span>
						</div>
					}
					isEmpty={(data) => data.length === 0}
					emptyFallback={
						<p className="px-4 py-8 text-center text-sm text-t2">Нет закладок</p>
					}
				>
					{(bookmarks) => (
						<div className="space-y-2 p-4">
							{bookmarks.map((bm) => (
								<div
									key={bm.id}
									className="flex items-start gap-3 rounded-md border p-3"
								>
									<button
										type="button"
										className="min-w-0 flex-1 text-left"
										onClick={() => handleNavigate(bm)}
									>
										<p className="text-sm font-medium">{bm.channel_id}</p>
										<p className="mt-0.5 line-clamp-2 text-xs text-t2">
											{bm.message_id}
										</p>
										<p className="mt-1 text-xs text-t2">
											{new Date(bm.created_at).toLocaleDateString("ru-RU")}
										</p>
									</button>
									<Button
										variant="ghost"
										size="icon"
										className="h-7 w-7 shrink-0"
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
							))}
						</div>
					)}
				</QueryStateView>
			</ScrollArea>
		</div>
	);
}
