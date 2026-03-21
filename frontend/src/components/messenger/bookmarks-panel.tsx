import { useCallback } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getBookmarks, removeBookmark, type Bookmark } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";

/**
 * BookmarksPanel отображает список сохранённых сообщений пользователя.
 * Доступен из sidebar мессенджера. Позволяет удалять закладки и переходить к сообщениям.
 */
export function BookmarksPanel() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: bookmarks, isLoading } = useQuery({
    queryKey: ["messenger", "bookmarks"],
    queryFn: getBookmarks,
  });

  const handleRemove = useCallback(
    async (bookmark: Bookmark) => {
      await removeBookmark(bookmark.id);
      await queryClient.invalidateQueries({
        queryKey: ["messenger", "bookmarks"],
      });
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

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <span className="text-sm text-muted-foreground">Загрузка...</span>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex h-12 items-center border-b px-4">
        <span className="text-sm font-semibold">Закладки</span>
      </div>
      <ScrollArea className="flex-1">
        {bookmarks && bookmarks.length > 0 ? (
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
                  <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
                    {bm.message_id}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {new Date(bm.created_at).toLocaleDateString("ru-RU")}
                  </p>
                </button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 shrink-0"
                  title="Удалить из закладок"
                  onClick={() => void handleRemove(bm)}
                >
                  <svg
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
        ) : (
          <p className="px-4 py-8 text-center text-sm text-muted-foreground">
            Нет закладок
          </p>
        )}
      </ScrollArea>
    </div>
  );
}
