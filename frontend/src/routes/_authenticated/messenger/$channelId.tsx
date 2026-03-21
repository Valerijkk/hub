import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";
import { ChatView } from "@/components/messenger/chat-view";
import { CallButton } from "@/components/calls/CallButton";
import { useMessengerStore } from "@/stores/messenger-store";

/**
 * Маршрут для отображения конкретного канала мессенджера.
 * CallButton интегрирован в заголовок ChatView для единообразия.
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

  return (
    <div className="flex h-full flex-col overflow-hidden">
      <ChatView channelId={channelId} />
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/messenger/$channelId",
)({
  component: ChannelPage,
});
