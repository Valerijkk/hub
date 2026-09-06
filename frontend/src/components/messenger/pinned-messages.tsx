import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import {
	getChannelPins,
	type ChannelCapabilities,
	type Message,
	type PinnedMessage,
} from "@/api/messenger";
import { QueryErrorAlert, QueryStateView } from "@/components/query-state-view";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { messagesQueryKey } from "@/hooks/use-messages";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { usePanelUnpinCommand } from "./use-panel-unpin-command";

/** Пропсы компонента PinnedMessages. */
interface PinnedMessagesProps {
	/** ID канала. */
	channelId: string;
	/** Fresh server-authoritative pin capabilities; missing data fails closed. */
	capabilities?: Pick<ChannelCapabilities, "can_list_pins" | "can_pin">;
}

export const PINS_QUERY_KEY = ["messenger", "pins"] as const;

export function pinsQueryKey(channelId: string, userId: string) {
	return [...PINS_QUERY_KEY, channelId, userId] as const;
}

/**
 * Preview текста пина: content_text из кэша сообщений канала, иначе нейтральный label.
 * A3W9-3-002: не показывать UUID как единственный смысл строки.
 */
export function resolvePinPreview(
	pages: Message[][] | undefined,
	messageId: string,
): string {
	if (pages) {
		for (const page of pages) {
			const msg = page.find((m) => m.id === messageId);
			const text = msg?.content_text?.trim();
			if (text) {
				return text.length > 120 ? `${text.slice(0, 120)}…` : text;
			}
		}
	}
	return "Закреплённое сообщение";
}

/**
 * PinnedMessages -- панель закреплённых сообщений канала.
 * Открывается кнопкой pin в заголовке чата.
 */
export function PinnedMessages({
	channelId,
	capabilities,
}: PinnedMessagesProps) {
	const [open, setOpen] = useState(false);
	const queryClient = useQueryClient();
	const userId = useAuthStore((state) => state.user?.id ?? null);
	const channelRevoked = useMessengerStore((state) =>
		state.revokedChannelIds.has(channelId),
	);
	const canListPins =
		capabilities?.can_list_pins === true && !channelRevoked;
	const canPin = canListPins && capabilities?.can_pin === true;
	const { pending: unpinPending, submit: submitUnpin } = usePanelUnpinCommand({
		open,
		channelId,
		canListPins,
		canPin,
	});
	const queryKey = pinsQueryKey(channelId, userId ?? "__no-user__");

	const pinsQuery = useQuery({
		queryKey,
		queryFn: () => getChannelPins(channelId),
		enabled: open && Boolean(userId) && canListPins,
	});

	useEffect(() => {
		if (!canListPins && open) setOpen(false);
	}, [canListPins, open]);

	const messagesCache = userId
		? queryClient.getQueryData<{ pages: Message[][] }>(
				messagesQueryKey(channelId, userId),
			)
		: undefined;

	if (!canListPins) return null;

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<Button
					variant="ghost"
					size="icon"
					className="h-8 w-8"
					title="Закреплённые"
					aria-label="Закреплённые сообщения"
				>
					<svg
						aria-hidden="true"
						className="h-4 w-4"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2}
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
						/>
					</svg>
				</Button>
			</DialogTrigger>
			<DialogContent className="max-w-md">
				<DialogHeader>
					<DialogTitle>Закреплённые сообщения</DialogTitle>
				</DialogHeader>
				<ScrollArea className="max-h-96">
					{!userId ? (
						<p className="py-4 text-center text-sm text-t2">
							Нет закреплённых сообщений
						</p>
					) : (
						<QueryStateView<PinnedMessage[]>
							query={pinsQuery}
							loadingFallback={
								<p
									role="status"
									aria-live="polite"
									className="py-4 text-center text-sm text-t2"
								>
									Загрузка закреплённых сообщений...
								</p>
							}
							errorFallback={(_, retry) => (
								<QueryErrorAlert
									compact
									message="Не удалось загрузить закреплённые сообщения"
									retry={retry}
								/>
							)}
							isEmpty={(data) => data.length === 0}
							emptyFallback={
								<p className="py-4 text-center text-sm text-t2">
									Нет закреплённых сообщений
								</p>
							}
						>
							{(currentPins) => (
								<div className="space-y-2 pt-2">
									{currentPins.map((pin) => {
										const preview = resolvePinPreview(
											messagesCache?.pages,
											pin.message_id,
										);
										return (
											<div
												key={pin.id}
												className="flex items-start justify-between rounded-md border p-3"
											>
												<div className="min-w-0 flex-1">
													<p
														className="text-sm line-clamp-2"
														title={preview}
													>
														{preview}
													</p>
													<p className="text-xs text-t2">
														Закреплено:{" "}
														{new Date(pin.pinned_at).toLocaleDateString(
															"ru-RU",
														)}
													</p>
												</div>
												{canPin && (
													<Button
														variant="ghost"
														size="sm"
														className="shrink-0"
														disabled={unpinPending}
														onClick={() => void submitUnpin(pin)}
													>
														Открепить
													</Button>
												)}
											</div>
										);
									})}
								</div>
							)}
						</QueryStateView>
					)}
				</ScrollArea>
			</DialogContent>
		</Dialog>
	);
}
