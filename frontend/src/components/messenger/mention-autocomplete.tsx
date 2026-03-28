import { useQuery } from "@tanstack/react-query";
import {
	type KeyboardEvent,
	useCallback,
	useEffect,
	useMemo,
	useRef,
	useState,
} from "react";
import { getChannelMembers, type ChannelMember } from "@/api/messenger";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Типы
// ---------------------------------------------------------------------------

/** Пропсы компонента MentionAutocomplete. */
interface MentionAutocompleteProps {
	/** ID канала для загрузки списка участников. */
	channelId: string;
	/** Текущий текст в textarea. */
	text: string;
	/** Позиция курсора в textarea. */
	cursorPosition: number;
	/** Ref на textarea для позиционирования. */
	textareaRef: React.RefObject<HTMLTextAreaElement | null>;
	/** Вызывается при выборе упоминания. Возвращает новый текст, позицию курсора и user_id. */
	onSelect: (newText: string, newCursorPosition: number, userId: string) => void;
	/** Ref для регистрации обработчика клавиш (вызывается из parent onKeyDown). */
	keyHandlerRef?: React.MutableRefObject<
		((e: KeyboardEvent<HTMLTextAreaElement>) => boolean) | null
	>;
}

// ---------------------------------------------------------------------------
// Утилиты
// ---------------------------------------------------------------------------

/**
 * Извлекает текущий @-запрос из текста на позиции курсора.
 * Возвращает null если курсор не внутри @-упоминания.
 */
function extractMentionQuery(
	text: string,
	cursorPosition: number,
): { query: string; startIndex: number } | null {
	// Ищем ближайший @ перед курсором.
	const textBeforeCursor = text.slice(0, cursorPosition);
	const atIndex = textBeforeCursor.lastIndexOf("@");
	if (atIndex === -1) return null;

	// @ должен быть в начале строки или после пробела/переноса.
	if (atIndex > 0) {
		const charBefore = textBeforeCursor[atIndex - 1];
		if (charBefore !== " " && charBefore !== "\n") return null;
	}

	// Текст после @ до курсора — это запрос фильтрации.
	const query = textBeforeCursor.slice(atIndex + 1);

	// Если в запросе есть пробел — это не активный @mention.
	if (query.includes(" ")) return null;

	return { query, startIndex: atIndex };
}

// ---------------------------------------------------------------------------
// Компонент
// ---------------------------------------------------------------------------

/**
 * MentionAutocomplete — выпадающий список участников канала
 * для @-упоминаний в композере сообщений.
 *
 * Появляется когда пользователь набирает @ в textarea.
 * Поддерживает навигацию стрелками и выбор по Enter/клику.
 */
export function MentionAutocomplete({
	channelId,
	text,
	cursorPosition,
	textareaRef: _textareaRef,
	onSelect,
	keyHandlerRef,
}: MentionAutocompleteProps) {
	const [selectedIndex, setSelectedIndex] = useState(0);
	const listRef = useRef<HTMLDivElement>(null);

	// Загружаем участников канала.
	const { data: members } = useQuery({
		queryKey: ["messenger", "channel-members", channelId],
		queryFn: () => getChannelMembers(channelId),
		staleTime: 30_000, // Кэш на 30 сек.
	});

	// Определяем, активен ли автокомплит (есть @-запрос).
	const mentionContext = useMemo(
		() => extractMentionQuery(text, cursorPosition),
		[text, cursorPosition],
	);

	// Фильтруем участников по введённому тексту после @.
	const filteredMembers = useMemo(() => {
		if (!mentionContext || !members) return [];
		const q = mentionContext.query.toLowerCase();
		return members.filter(
			(m) =>
				(m.user_name ?? "").toLowerCase().includes(q) ||
				m.user_id.toLowerCase().includes(q),
		);
	}, [mentionContext, members]);

	// Сбрасываем выделение при изменении списка.
	useEffect(() => {
		setSelectedIndex(0);
	}, [filteredMembers.length]);

	// Вставка выбранного участника в текст.
	const insertMention = useCallback(
		(member: ChannelMember) => {
			if (!mentionContext) return;
			const displayName = member.user_name || member.user_id;
			const before = text.slice(0, mentionContext.startIndex);
			const after = text.slice(cursorPosition);
			const mention = `@${displayName} `;
			const newText = before + mention + after;
			const newPos = mentionContext.startIndex + mention.length;
			onSelect(newText, newPos, member.user_id);
		},
		[mentionContext, text, cursorPosition, onSelect],
	);

	// Обработка клавиш (вызывается из родительского компонента).
	const handleKeyDown = useCallback(
		(e: KeyboardEvent<HTMLTextAreaElement>): boolean => {
			if (!mentionContext || filteredMembers.length === 0) return false;

			switch (e.key) {
				case "ArrowDown":
					e.preventDefault();
					setSelectedIndex((prev) =>
						prev < filteredMembers.length - 1 ? prev + 1 : 0,
					);
					return true;
				case "ArrowUp":
					e.preventDefault();
					setSelectedIndex((prev) =>
						prev > 0 ? prev - 1 : filteredMembers.length - 1,
					);
					return true;
				case "Enter":
				case "Tab": {
					e.preventDefault();
					const member = filteredMembers[selectedIndex];
					if (member) insertMention(member);
					return true;
				}
				case "Escape":
					// Позволяем закрыть автокомплит нажатием Escape.
					return true;
				default:
					return false;
			}
		},
		[mentionContext, filteredMembers, selectedIndex, insertMention],
	);

	// Регистрируем обработчик клавиш в parent через ref.
	useEffect(() => {
		if (keyHandlerRef) {
			keyHandlerRef.current = handleKeyDown;
		}
		return () => {
			if (keyHandlerRef) {
				keyHandlerRef.current = null;
			}
		};
	}, [keyHandlerRef, handleKeyDown]);

	// Прокрутка к выделенному элементу.
	useEffect(() => {
		const list = listRef.current;
		if (!list) return;
		const item = list.children[selectedIndex] as HTMLElement | undefined;
		item?.scrollIntoView({ block: "nearest" });
	}, [selectedIndex]);

	// Не рендерим если автокомплит не активен.
	const isVisible =
		mentionContext !== null && filteredMembers.length > 0;

	if (!isVisible) return null;

	return (
		<div
			ref={listRef}
			className="absolute bottom-full left-0 mb-1 w-64 max-h-48 overflow-y-auto rounded-lg border bg-popover shadow-lg z-50"
			// Предотвращаем потерю фокуса textarea при клике по списку.
			onMouseDown={(e) => e.preventDefault()}
		>
			{filteredMembers.map((member, index) => (
				<button
					key={member.id}
					type="button"
					className={cn(
						"flex w-full items-center gap-2 px-3 py-2 text-sm transition-colors",
						index === selectedIndex
							? "bg-accent text-accent-foreground"
							: "hover:bg-accent/50",
					)}
					onClick={() => insertMention(member)}
				>
					{/* Аватар-заглушка с первой буквой имени */}
					<div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/20 text-[10px] font-semibold text-primary">
						{(member.user_name ?? member.user_id)
							.charAt(0)
							.toUpperCase()}
					</div>
					<span className="truncate font-medium">
						{member.user_name || member.user_id}
					</span>
					<span className="ml-auto text-[10px] text-t2">
						{member.role}
					</span>
				</button>
			))}
		</div>
	);
}

