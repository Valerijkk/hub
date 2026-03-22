import { toggleReaction } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";

/** Набор популярных эмодзи для быстрого выбора. */
const POPULAR_EMOJIS = [
	"\u{1F44D}", // thumbs up
	"\u{1F44E}", // thumbs down
	"\u{2764}\u{FE0F}", // red heart
	"\u{1F602}", // laughing
	"\u{1F622}", // crying
	"\u{1F632}", // astonished
	"\u{1F525}", // fire
	"\u{1F4AF}", // 100
	"\u{1F440}", // eyes
	"\u{1F389}", // party popper
	"\u{1F680}", // rocket
	"\u{2705}", // check mark
	"\u{274C}", // cross mark
	"\u{1F914}", // thinking
	"\u{1F44F}", // clapping
	"\u{1F60D}", // heart eyes
	"\u{1F64F}", // folded hands
	"\u{1F4A1}", // light bulb
	"\u{26A0}\u{FE0F}", // warning
	"\u{1F3AF}", // bullseye
];

/** Пропсы компонента ReactionPicker. */
interface ReactionPickerProps {
	/** ID канала. */
	channelId: string;
	/** ID сообщения. */
	messageId: string;
	/** Управление открытием попавера. */
	open: boolean;
	/** Колбэк изменения состояния попавера. */
	onOpenChange: (open: boolean) => void;
}

/**
 * ReactionPicker -- компонент выбора эмодзи-реакции в Popover.
 * Показывает сетку популярных эмодзи, клик переключает реакцию.
 */
export function ReactionPicker({
	channelId,
	messageId,
	open,
	onOpenChange,
}: ReactionPickerProps) {
	const handleSelect = async (emoji: string) => {
		await toggleReaction(channelId, messageId, emoji);
		onOpenChange(false);
	};

	return (
		<Popover open={open} onOpenChange={onOpenChange}>
			<PopoverTrigger asChild>
				<Button variant="ghost" size="icon" className="h-7 w-7" title="Реакция">
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
							className="flex h-8 w-8 items-center justify-center rounded text-lg hover:bg-accent transition-colors"
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
