import { useCallback, useEffect, useLayoutEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import { useMessengerStore } from "@/stores/messenger-store";

/** Набор популярных эмодзи для быстрого выбора. */
const POPULAR_EMOJIS = [
	"\u{1F44D}",
	"\u{2764}\u{FE0F}",
	"\u{1F602}",
	"\u{1F389}",
	"\u{1F525}",
	"\u{1F440}",
	"\u{1F64C}",
	"\u{1F914}",
	"\u{2705}",
	"\u{274C}",
	"\u{1F680}",
	"\u{1F4AF}",
	"\u{1F44F}",
	"\u{1F622}",
	"\u{1F621}",
	"\u{1F60D}",
	"\u{1F973}",
	"\u{1F4A1}",
	"\u{26A0}\u{FE0F}",
	"\u{1F3AF}",
];

interface ReactionPickerProps {
	channelId: string;
	messageId: string;
	canReact: boolean;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onToggleReaction: (emoji: string) => Promise<boolean>;
	pending: boolean;
}

/** Emoji picker с прямым capability/revoke barrier перед сетевым вызовом. */
export function ReactionPicker({
	channelId,
	messageId,
	canReact,
	open,
	onOpenChange,
	onToggleReaction,
	pending,
}: ReactionPickerProps) {
	const channelRevoked = useMessengerStore((state) =>
		state.revokedChannelIds.has(channelId),
	);
	const disabled = !canReact || channelRevoked;
	const canReactNow = useCallback(() => {
		const state = useMessengerStore.getState();
		if (!canReact || state.isChannelRevoked(channelId)) return false;
		const channel = state.channels.find((item) => item.id === channelId);
		return channel?.capabilities?.can_react === true;
	}, [canReact, channelId]);
	const openSessionRef = useRef<object | null>(null);

	useLayoutEffect(() => {
		if (!open || disabled) {
			openSessionRef.current = null;
			return;
		}
		const session = {};
		openSessionRef.current = session;
		return () => {
			if (openSessionRef.current === session) openSessionRef.current = null;
		};
	}, [channelId, disabled, messageId, open]);

	useEffect(() => {
		if (disabled && open) onOpenChange(false);
	}, [disabled, onOpenChange, open]);

	const handleSelect = async (emoji: string) => {
		const session = openSessionRef.current;
		if (
			!session ||
			openSessionRef.current !== session ||
			disabled ||
			pending ||
			!canReactNow()
		) {
			return;
		}
		const succeeded = await onToggleReaction(emoji);
		if (
			succeeded &&
			openSessionRef.current === session &&
			canReactNow()
		) {
			onOpenChange(false);
		}
	};

	return (
		<Popover
			open={open && !disabled}
			onOpenChange={(nextOpen) => {
				if (nextOpen && disabled) return;
				onOpenChange(nextOpen);
			}}
		>
			<PopoverTrigger asChild>
				<Button
					variant="ghost"
					size="icon"
					className="h-7 w-7"
					title="Реакция"
					aria-label="Реакция"
					disabled={disabled}
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
							d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
				</Button>
			</PopoverTrigger>
			<PopoverContent className="w-auto p-2" align="center" side="top">
				<div className="grid grid-cols-5 gap-1">
					{POPULAR_EMOJIS.map((emoji) => (
						<button
							key={emoji}
							type="button"
							aria-label={emoji}
							className="flex h-8 w-8 items-center justify-center rounded text-lg transition-colors hover:bg-accent"
							disabled={disabled || pending}
							onClick={() => void handleSelect(emoji)}
						>
							{emoji}
						</button>
					))}
				</div>
			</PopoverContent>
		</Popover>
	);
}
