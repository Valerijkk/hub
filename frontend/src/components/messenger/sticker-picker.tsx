import { useCallback, useState } from "react";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Стикер-паки — тематические наборы эмодзи, рендерятся крупно как стикеры
// ---------------------------------------------------------------------------

/** Один стикер. */
export interface Sticker {
	/** Уникальный ID стикера внутри пака. */
	id: string;
	/** Эмодзи-символ (или комбинация). */
	emoji: string;
	/** Подпись-тултип. */
	label: string;
}

/** Стикер-пак. */
export interface StickerPack {
	id: string;
	name: string;
	icon: string;
	stickers: Sticker[];
}

/** Встроенные стикер-паки. */
export const STICKER_PACKS: StickerPack[] = [
	{
		id: "emotions",
		name: "Эмоции",
		icon: "\u{1F60E}",
		stickers: [
			{ id: "cool", emoji: "\u{1F60E}", label: "Круто" },
			{ id: "love", emoji: "\u{1F970}", label: "Обожаю" },
			{ id: "fire", emoji: "\u{1F525}", label: "Огонь" },
			{ id: "laugh", emoji: "\u{1F923}", label: "Ржу" },
			{ id: "wow", emoji: "\u{1F92F}", label: "Взрыв мозга" },
			{ id: "sad", emoji: "\u{1F62D}", label: "Плачу" },
			{ id: "angry", emoji: "\u{1F621}", label: "Злой" },
			{ id: "party", emoji: "\u{1F973}", label: "Вечеринка" },
			{ id: "think", emoji: "\u{1F914}", label: "Хммм" },
			{ id: "shrug", emoji: "\u{1F937}", label: "Не знаю" },
			{ id: "facepalm", emoji: "\u{1F926}", label: "Фейспалм" },
			{ id: "clown", emoji: "\u{1F921}", label: "Клоун" },
			{ id: "skull", emoji: "\u{1F480}", label: "Мёртв" },
			{ id: "ghost", emoji: "\u{1F47B}", label: "Привет" },
			{ id: "alien", emoji: "\u{1F47D}", label: "Пришелец" },
			{ id: "zany", emoji: "\u{1F92A}", label: "Дурашка" },
		],
	},
	{
		id: "work",
		name: "Работа",
		icon: "\u{1F4BB}",
		stickers: [
			{ id: "laptop", emoji: "\u{1F4BB}", label: "За работой" },
			{ id: "rocket", emoji: "\u{1F680}", label: "Запуск!" },
			{ id: "check", emoji: "\u2705", label: "Готово" },
			{ id: "bug", emoji: "\u{1F41B}", label: "Баг" },
			{ id: "hammer", emoji: "\u{1F528}", label: "Фиксим" },
			{ id: "bulb", emoji: "\u{1F4A1}", label: "Идея!" },
			{ id: "chart", emoji: "\u{1F4C8}", label: "Рост" },
			{ id: "alarm", emoji: "\u23F0", label: "Дедлайн" },
			{ id: "coffee", emoji: "\u2615", label: "Кофе" },
			{ id: "trophy", emoji: "\u{1F3C6}", label: "Победа" },
			{ id: "target", emoji: "\u{1F3AF}", label: "В точку" },
			{ id: "construction", emoji: "\u{1F6A7}", label: "В разработке" },
			{ id: "ship", emoji: "\u{1F6A2}", label: "Шипим" },
			{ id: "boom", emoji: "\u{1F4A5}", label: "Бум" },
			{ id: "eyes", emoji: "\u{1F440}", label: "Ревью" },
			{ id: "muscle", emoji: "\u{1F4AA}", label: "Сила" },
		],
	},
	{
		id: "reactions",
		name: "Реакции",
		icon: "\u{1F44D}",
		stickers: [
			{ id: "thumbsup", emoji: "\u{1F44D}", label: "Класс" },
			{ id: "thumbsdown", emoji: "\u{1F44E}", label: "Не ок" },
			{ id: "clap", emoji: "\u{1F44F}", label: "Аплодисменты" },
			{ id: "wave", emoji: "\u{1F44B}", label: "Привет" },
			{ id: "pray", emoji: "\u{1F64F}", label: "Спасибо" },
			{ id: "heart", emoji: "\u2764\uFE0F", label: "Сердце" },
			{ id: "heartfire", emoji: "\u2764\uFE0F\u200D\u{1F525}", label: "Горячее сердце" },
			{ id: "100", emoji: "\u{1F4AF}", label: "Сотка" },
			{ id: "ok", emoji: "\u{1F44C}", label: "Ок" },
			{ id: "salute", emoji: "\u{1FAE1}", label: "Салют" },
			{ id: "hug", emoji: "\u{1F917}", label: "Обнимашки" },
			{ id: "star", emoji: "\u2B50", label: "Звезда" },
			{ id: "rainbow", emoji: "\u{1F308}", label: "Радуга" },
			{ id: "sparkles", emoji: "\u2728", label: "Блеск" },
			{ id: "peace", emoji: "\u270C\uFE0F", label: "Мир" },
			{ id: "handshake", emoji: "\u{1F91D}", label: "Договорились" },
		],
	},
	{
		id: "animals",
		name: "Зверьё",
		icon: "\u{1F431}",
		stickers: [
			{ id: "cat", emoji: "\u{1F431}", label: "Котик" },
			{ id: "dog", emoji: "\u{1F436}", label: "Пёсик" },
			{ id: "bear", emoji: "\u{1F43B}", label: "Мишка" },
			{ id: "fox", emoji: "\u{1F98A}", label: "Лиса" },
			{ id: "unicorn", emoji: "\u{1F984}", label: "Единорог" },
			{ id: "penguin", emoji: "\u{1F427}", label: "Пингвин" },
			{ id: "owl", emoji: "\u{1F989}", label: "Сова" },
			{ id: "octopus", emoji: "\u{1F419}", label: "Осьминог" },
			{ id: "monkey", emoji: "\u{1F435}", label: "Обезьянка" },
			{ id: "panda", emoji: "\u{1F43C}", label: "Панда" },
			{ id: "frog", emoji: "\u{1F438}", label: "Лягушка" },
			{ id: "butterfly", emoji: "\u{1F98B}", label: "Бабочка" },
		],
	},
];

/** Данные стикера для отправки как сообщение. */
export interface StickerData {
	pack_id: string;
	sticker_id: string;
	emoji: string;
	label: string;
}

// ---------------------------------------------------------------------------
// Хелпер для определения стикер-сообщения
// ---------------------------------------------------------------------------

/** Проверяет, является ли content_json стикером. */
export function isStickerMessage(contentJson?: Record<string, unknown> | null): boolean {
	if (!contentJson) return false;
	const children = contentJson.children as unknown[];
	if (!Array.isArray(children) || children.length !== 1) return false;
	const child = children[0] as Record<string, unknown> | undefined;
	return child?.type === "sticker";
}

/** Извлекает данные стикера из content_json. */
export function extractStickerData(contentJson: Record<string, unknown>): StickerData | null {
	const children = contentJson.children as unknown[];
	if (!Array.isArray(children) || children.length !== 1) return null;
	const child = children[0] as Record<string, unknown>;
	if (child?.type !== "sticker") return null;
	return {
		pack_id: String(child.pack_id ?? ""),
		sticker_id: String(child.sticker_id ?? ""),
		emoji: String(child.emoji ?? ""),
		label: String(child.label ?? ""),
	};
}

/** Создаёт content_json для стикер-сообщения. */
export function buildStickerContentJson(sticker: StickerData): Record<string, unknown> {
	return {
		type: "doc",
		children: [
			{
				type: "sticker",
				pack_id: sticker.pack_id,
				sticker_id: sticker.sticker_id,
				emoji: sticker.emoji,
				label: sticker.label,
				children: [{ text: "" }],
			},
		],
	};
}

// ---------------------------------------------------------------------------
// Компонент StickerPicker
// ---------------------------------------------------------------------------

interface StickerPickerProps {
	onSelect: (sticker: StickerData) => void;
}

/**
 * StickerPicker — поповер для выбора стикера из встроенных паков.
 * Табы для переключения между паками, сетка стикеров.
 */
export function StickerPicker({ onSelect }: StickerPickerProps) {
	const [open, setOpen] = useState(false);
	const [activePackId, setActivePackId] = useState(STICKER_PACKS[0]!.id);

	const activePack = STICKER_PACKS.find((p) => p.id === activePackId) ?? STICKER_PACKS[0]!;

	const handleSelect = useCallback(
		(pack: StickerPack, sticker: Sticker) => {
			onSelect({
				pack_id: pack.id,
				sticker_id: sticker.id,
				emoji: sticker.emoji,
				label: sticker.label,
			});
			setOpen(false);
		},
		[onSelect],
	);

	return (
		<Popover open={open} onOpenChange={setOpen}>
			<PopoverTrigger asChild>
				<button
					type="button"
					className="flex h-11 w-9 shrink-0 items-center justify-center text-t2/60 hover:text-t2 transition-colors"
					title="Стикеры"
				>
					<svg
						className="h-5 w-5"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2}
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
						/>
					</svg>
				</button>
			</PopoverTrigger>

			<PopoverContent
				className="w-80 p-0 overflow-hidden"
				align="end"
				side="top"
				sideOffset={8}
			>
				<div className="flex flex-col max-h-[380px]">
					{/* Табы паков */}
					<div className="flex gap-0.5 px-2 py-1.5 border-b border-border/50 bg-raised/20">
						{STICKER_PACKS.map((pack) => (
							<button
								key={pack.id}
								type="button"
								className={cn(
									"flex items-center justify-center h-8 w-8 rounded-md text-lg transition-colors",
									activePackId === pack.id
										? "bg-primary/10 ring-1 ring-primary/30"
										: "hover:bg-accent",
								)}
								onClick={() => setActivePackId(pack.id)}
								title={pack.name}
							>
								{pack.icon}
							</button>
						))}
					</div>

					{/* Название пака */}
					<div className="px-3 py-1.5">
						<span className="text-xs font-medium text-t2">
							{activePack.name}
						</span>
					</div>

					{/* Сетка стикеров */}
					<div className="flex-1 overflow-y-auto px-2 pb-2">
						<div className="grid grid-cols-4 gap-1">
							{activePack.stickers.map((sticker) => (
								<button
									key={sticker.id}
									type="button"
									className={cn(
										"flex items-center justify-center rounded-lg p-1.5 transition-[color,background-color,border-color,box-shadow,transform,filter,opacity]",
										"hover:bg-accent hover:scale-110",
										"active:scale-95",
									)}
									onClick={() => handleSelect(activePack, sticker)}
									title={sticker.label}
								>
									<span className="text-4xl leading-none select-none">
										{sticker.emoji}
									</span>
								</button>
							))}
						</div>
					</div>
				</div>
			</PopoverContent>
		</Popover>
	);
}
