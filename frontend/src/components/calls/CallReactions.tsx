import { useRoomContext } from "@livekit/components-react";
import { RoomEvent } from "livekit-client";
import { useCallback, useEffect, useRef, useState } from "react";

import { cn } from "@/lib/utils";

/** Набор доступных реакций. */
const REACTIONS = [
	{ emoji: "\u{1F44D}", label: "Like" },
	{ emoji: "\u{1F44F}", label: "Clap" },
	{ emoji: "\u{2764}\u{FE0F}", label: "Heart" },
	{ emoji: "\u{1F602}", label: "Laugh" },
	{ emoji: "\u{1F389}", label: "Party" },
] as const;

/** Структура плавающей реакции на экране. */
interface FloatingReaction {
	/** Уникальный ID для React key. */
	id: number;
	/** Эмодзи реакции. */
	emoji: string;
	/** Имя отправителя. */
	sender: string;
	/** Горизонтальная позиция (0-100%). */
	x: number;
}

/**
 * Формат данных реакции, передаваемый через LiveKit data messages.
 * Используется для сериализации/десериализации при отправке между участниками.
 */
interface ReactionPayload {
	type: "reaction";
	emoji: string;
	sender: string;
}

/** Счётчик для уникальных ID плавающих реакций. */
let reactionIdCounter = 0;

/**
 * Простой EventTarget для локальных реакций.
 * CallReactionBar диспатчит событие при отправке,
 * CallReactionsOverlay слушает его для немедленного показа.
 */
const localReactionBus = new EventTarget();

/** Пропсы панели реакций. */
interface CallReactionBarProps {
	/** Коллбэк при закрытии панели. */
	onClose: () => void;
}

/**
 * Панель выбора реакций -- отображается над кнопкой "Реакции" в CallControls.
 * При клике на эмодзи отправляет реакцию через LiveKit data messages
 * и через локальный event bus для немедленного показа, затем закрывает панель.
 */
export function CallReactionBar({ onClose }: CallReactionBarProps) {
	const room = useRoomContext();

	/** Отправить реакцию всем участникам через LiveKit data channel. */
	const sendReaction = useCallback(
		async (emoji: string) => {
			const senderName =
				room.localParticipant.name || room.localParticipant.identity;

			const payload: ReactionPayload = {
				type: "reaction",
				emoji,
				sender: senderName,
			};

			const data = new TextEncoder().encode(JSON.stringify(payload));

			// Показываем реакцию локально через event bus (мгновенно)
			localReactionBus.dispatchEvent(
				new CustomEvent("reaction", { detail: { emoji, sender: senderName } }),
			);

			// Отправляем reliable data message всем участникам
			try {
				await room.localParticipant.publishData(data, { reliable: true });
			} catch {
				// Ошибка отправки -- реакция уже показана локально
			}

			onClose();
		},
		[room, onClose],
	);

	return (
		<div className="absolute bottom-full mb-3 left-1/2 -translate-x-1/2 z-50">
			<div
				className={cn(
					"flex items-center gap-1 rounded-2xl px-3 py-2",
					"bg-[#1a1f2e]/95 backdrop-blur-xl border border-white/10",
					"shadow-2xl shadow-black/40",
					"animate-in fade-in slide-in-from-bottom-2 duration-200",
				)}
			>
				{REACTIONS.map(({ emoji, label }) => (
					<button
						key={label}
						type="button"
						onClick={() => void sendReaction(emoji)}
						className={cn(
							"flex h-10 w-10 items-center justify-center rounded-xl text-xl",
							"transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-150",
							"hover:bg-white/10 hover:scale-125",
							"active:scale-95",
						)}
						title={label}
						aria-label={label}
					>
						{emoji}
					</button>
				))}
			</div>
		</div>
	);
}

/**
 * Оверлей для плавающих реакций -- размещается поверх области участников.
 * Слушает входящие data messages от LiveKit (реакции других участников)
 * и локальный event bus (собственные реакции), отображает эмодзи
 * с CSS-анимацией всплытия и затухания.
 */
export function CallReactionsOverlay() {
	const room = useRoomContext();
	const [reactions, setReactions] = useState<FloatingReaction[]>([]);
	const cleanupTimers = useRef<Map<number, ReturnType<typeof setTimeout>>>(
		new Map(),
	);

	/**
	 * Добавить плавающую реакцию на экран.
	 * Реакция автоматически удаляется через 2.5 секунды (совпадает с CSS-анимацией).
	 */
	const addFloatingReaction = useCallback((emoji: string, sender: string) => {
		const id = ++reactionIdCounter;
		// Случайная горизонтальная позиция (10-90% от ширины)
		const x = 10 + Math.random() * 80;

		setReactions((prev) => [...prev, { id, emoji, sender, x }]);

		// Удаляем реакцию после завершения анимации
		const timer = setTimeout(() => {
			setReactions((prev) => prev.filter((r) => r.id !== id));
			cleanupTimers.current.delete(id);
		}, 2500);

		cleanupTimers.current.set(id, timer);
	}, []);

	/** Слушаем входящие data messages для получения реакций от других участников. */
	useEffect(() => {
		const handleDataReceived = (
			payload: Uint8Array,
			_participant?: unknown,
		) => {
			try {
				const text = new TextDecoder().decode(payload);
				const parsed = JSON.parse(text) as ReactionPayload;

				if (parsed.type === "reaction" && parsed.emoji) {
					addFloatingReaction(parsed.emoji, parsed.sender || "");
				}
			} catch {
				// Не JSON или не реакция -- игнорируем
			}
		};

		room.on(RoomEvent.DataReceived, handleDataReceived);

		return () => {
			room.off(RoomEvent.DataReceived, handleDataReceived);
		};
	}, [room, addFloatingReaction]);

	/** Слушаем локальный event bus для собственных реакций. */
	useEffect(() => {
		const handleLocal = (e: Event) => {
			const { emoji, sender } = (e as CustomEvent).detail as {
				emoji: string;
				sender: string;
			};
			addFloatingReaction(emoji, sender);
		};

		localReactionBus.addEventListener("reaction", handleLocal);

		return () => {
			localReactionBus.removeEventListener("reaction", handleLocal);
			// Очищаем все таймеры при размонтировании
			for (const timer of cleanupTimers.current.values()) {
				clearTimeout(timer);
			}
		};
	}, [addFloatingReaction]);

	if (reactions.length === 0) return null;

	return (
		<div className="pointer-events-none absolute inset-0 z-40 overflow-hidden">
			{reactions.map((r) => (
				<div
					key={r.id}
					className="absolute animate-float-up"
					style={{ left: `${r.x}%`, bottom: "80px" }}
				>
					<div className="flex flex-col items-center gap-1">
						<span className="text-4xl drop-shadow-lg">{r.emoji}</span>
						{r.sender && (
							<span className="text-[10px] text-white/50 font-medium bg-black/30 rounded-full px-2 py-0.5 whitespace-nowrap">
								{r.sender}
							</span>
						)}
					</div>
				</div>
			))}

			{/* CSS-анимация всплытия реакций */}
			<style>{`
				@keyframes float-up {
					0% {
						opacity: 1;
						transform: translateY(0) scale(1);
					}
					50% {
						opacity: 0.8;
						transform: translateY(-120px) scale(1.1);
					}
					100% {
						opacity: 0;
						transform: translateY(-240px) scale(0.8);
					}
				}
				.animate-float-up {
					animation: float-up 2.5s ease-out forwards;
				}
			`}</style>
		</div>
	);
}
