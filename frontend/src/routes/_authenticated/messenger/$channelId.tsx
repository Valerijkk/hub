import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";
import { ChatView } from "@/components/messenger/chat-view";
import { CallButton } from "@/components/calls/CallButton";
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

  return (
    <div className="flex h-full flex-col">
      {/* Кнопка звонка в верхней панели канала */}
      <div className="flex items-center justify-end border-b px-4 py-2">
        <CallButton channelId={channelId} />
      </div>
      <div className="flex-1 overflow-hidden">
        <ChatView channelId={channelId} />
      </div>
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/messenger/$channelId",
)({
  component: ChannelPage,
});
