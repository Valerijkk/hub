import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";
import { ChatView } from "@/components/messenger/chat-view";
import { useMessengerStore } from "@/stores/messenger-store";

/**
 * Маршрут для отображения конкретного канала мессенджера.
 */
function ChannelPage() {
  const { channelId } = Route.useParams();
  const setActiveChannel = useMessengerStore((s) => s.setActiveChannel);
  const resetUnread = useMessengerStore((s) => s.resetUnread);

  useEffect(() => {
    setActiveChannel(channelId);
    resetUnread(channelId);
    return () => {
      setActiveChannel(null);
    };
  }, [channelId, setActiveChannel, resetUnread]);

  return <ChatView channelId={channelId} />;
}

export const Route = createFileRoute(
  "/_authenticated/messenger/$channelId",
)({
  component: ChannelPage,
});
