import { BarChart3, Check, Lock } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import type { Poll } from "@/api/messenger";
import { closePoll, getPollResults, votePoll } from "@/api/messenger";
import { toastApiError } from "@/lib/format-api-error";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";
import { useWsStore } from "@/stores/ws-store";

/** Пропсы компонента PollMessage. */
interface PollMessageProps {
	/** ID опроса для загрузки и отображения. */
	pollId: string;
	/** ID канала (для подписки на WS-обновления). */
	channelId: string;
}

/**
 * PollMessage отображает опрос в виде карточки внутри ленты сообщений.
 * Показывает вопрос, варианты с прогресс-барами, счётчики голосов.
 * Поддерживает голосование кликом, закрытие создателем, real-time обновления.
 */
export function PollMessage({ pollId, channelId }: PollMessageProps) {
	const [poll, setPoll] = useState<Poll | null>(null);
	const [loading, setLoading] = useState(true);
	const [voting, setVoting] = useState(false);
	/** Выбранные варианты (для multi-choice до подтверждения). */
	const [selected, setSelected] = useState<Set<string>>(new Set());
	const currentUserId = useAuthStore((s) => s.user?.id);
	const { addListener, removeListener } = useWsStore();

	/** Загрузка данных опроса. */
	const fetchPoll = useCallback(async () => {
		try {
			const data = await getPollResults(pollId);
			setPoll(data);
			setLoading(false);
		} catch (err) {
			setLoading(false);
			toastApiError(err, "Не удалось загрузить опрос");
		}
	}, [pollId]);

	useEffect(() => {
		void fetchPoll();
	}, [fetchPoll]);

	/** Подписка на WS-обновления опроса (poll_voted, poll_closed). */
	useEffect(() => {
		const handler = (event: string, payload: unknown) => {
			if (!payload || typeof payload !== "object") return;
			const data = payload as Record<string, unknown>;
			const eventChannelId = data.channel_id;
			if (
				typeof eventChannelId === "string" &&
				eventChannelId.length > 0 &&
				eventChannelId !== channelId
			) {
				return;
			}
			const eventPollId = data.poll_id as string | undefined;
			if (eventPollId !== pollId) return;

			if (event === "poll_voted" || event === "poll_closed") {
				void fetchPoll();
			}
		};
		addListener(handler);
		return () => removeListener(handler);
	}, [pollId, channelId, addListener, removeListener, fetchPoll]);

	/** Проголосовал ли текущий пользователь. */
	const hasVoted = poll?.options.some((o) => o.voted) ?? false;

	/** Обработка клика по варианту ответа. */
	const handleOptionClick = useCallback(
		async (optionId: string) => {
			if (!poll || poll.is_closed || voting) return;

			if (poll.is_multi_choice) {
				// Для multi-choice: переключаем выбор.
				setSelected((prev) => {
					const next = new Set(prev);
					if (next.has(optionId)) {
						next.delete(optionId);
					} else {
						next.add(optionId);
					}
					return next;
				});
				return;
			}

			// Одиночный выбор: голосуем сразу.
			setVoting(true);
			try {
				const updated = await votePoll(pollId, [optionId]);
				setPoll(updated);
			} catch (err) {
				toastApiError(err, "Не удалось проголосовать");
			} finally {
				setVoting(false);
			}
		},
		[poll, pollId, voting],
	);

	/** Подтверждение голосования (для multi-choice). */
	const handleSubmitVote = useCallback(async () => {
		if (!poll || selected.size === 0 || voting) return;
		setVoting(true);
		try {
			const updated = await votePoll(pollId, Array.from(selected));
			setPoll(updated);
			setSelected(new Set());
		} catch (err) {
			toastApiError(err, "Не удалось проголосовать");
		} finally {
			setVoting(false);
		}
	}, [poll, pollId, selected, voting]);

	/** Закрытие опроса создателем. */
	const handleClose = useCallback(async () => {
		if (!poll || voting) return;
		setVoting(true);
		try {
			const updated = await closePoll(pollId);
			setPoll(updated);
		} catch (err) {
			toastApiError(err, "Не удалось закрыть опрос");
		} finally {
			setVoting(false);
		}
	}, [poll, pollId, voting]);

	if (loading) {
		return (
			<div className="my-1.5 rounded-xl border bg-card p-4 max-w-md animate-pulse">
				<div className="h-4 bg-raised rounded w-3/4 mb-3" />
				<div className="space-y-2">
					<div className="h-8 bg-raised rounded" />
					<div className="h-8 bg-raised rounded" />
				</div>
			</div>
		);
	}

	if (!poll) {
		return (
			<div className="my-1.5 rounded-xl border bg-card p-4 max-w-md text-sm text-t2">
				Не удалось загрузить опрос
			</div>
		);
	}

	const isCreator = currentUserId === poll.creator_id;
	/** Показываем результаты, если пользователь проголосовал или опрос закрыт. */
	const showResults = hasVoted || poll.is_closed;

	return (
		<div className="my-1.5 rounded-xl border bg-card shadow-sm max-w-md overflow-hidden">
			{/* Заголовок опроса */}
			<div className="px-4 pt-4 pb-2">
				<div className="flex items-start gap-2">
					<BarChart3 className="h-4 w-4 mt-0.5 text-primary shrink-0" />
					<div className="flex-1 min-w-0">
						<h4 className="text-sm font-semibold text-t1 leading-snug">
							{poll.question}
						</h4>
						<div className="flex items-center gap-2 mt-1 text-[11px] text-t2">
							{poll.is_anonymous && <span>Анонимное</span>}
							{poll.is_multi_choice && <span>Несколько вариантов</span>}
							{poll.is_closed && (
								<span className="flex items-center gap-0.5 text-[var(--orange)]">
									<Lock className="h-3 w-3" />
									Завершён
								</span>
							)}
						</div>
					</div>
				</div>
			</div>

			{/* Варианты ответов */}
			<div className="px-4 pb-3 space-y-1.5">
				{poll.options.map((option) => {
					const percentage =
						poll.total_votes > 0
							? Math.round((option.count / poll.total_votes) * 100)
							: 0;
					const isSelected = selected.has(option.id);

					return (
						<button
							key={option.id}
							type="button"
							disabled={poll.is_closed || voting}
							onClick={() => void handleOptionClick(option.id)}
							className={cn(
								"relative w-full text-left rounded-lg border px-3 py-2 text-sm transition-colors",
								"hover:border-primary/50 hover:bg-primary/5",
								"disabled:opacity-60 disabled:cursor-not-allowed",
								option.voted && "border-primary/50 bg-primary/5",
								isSelected &&
									"border-primary bg-primary/10 ring-1 ring-primary/30",
							)}
						>
							{/* Прогресс-бар (фон) */}
							{showResults && (
								<div
									className="absolute inset-0 rounded-lg bg-primary/10 transition-[width] duration-500"
									style={{ width: `${percentage}%` }}
								/>
							)}

							<div className="relative flex items-center justify-between gap-2">
								<div className="flex items-center gap-2 min-w-0">
									{/* Чекбокс для multi-choice */}
									{poll.is_multi_choice && !poll.is_closed && !hasVoted && (
										<div
											className={cn(
												"h-4 w-4 rounded border-2 shrink-0 flex items-center justify-center transition-colors",
												isSelected
													? "bg-primary border-primary text-primary-foreground"
													: "border-t2/30",
											)}
										>
											{isSelected && <Check className="h-3 w-3" />}
										</div>
									)}
									{/* Индикатор "мой голос" */}
									{option.voted && (
										<Check className="h-3.5 w-3.5 text-primary shrink-0" />
									)}
									<span className="truncate">{option.text}</span>
								</div>

								{showResults && (
									<span className="text-xs font-medium text-t2 shrink-0">
										{option.count} ({percentage}%)
									</span>
								)}
							</div>
						</button>
					);
				})}
			</div>

			{/* Кнопка подтверждения (multi-choice) */}
			{poll.is_multi_choice &&
				!poll.is_closed &&
				!hasVoted &&
				selected.size > 0 && (
					<div className="px-4 pb-3">
						<button
							type="button"
							onClick={() => void handleSubmitVote()}
							disabled={voting}
							className="w-full rounded-lg bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50"
						>
							{voting ? "Голосование..." : "Проголосовать"}
						</button>
					</div>
				)}

			{/* Подвал: всего голосов + кнопка закрытия */}
			<div className="flex items-center justify-between px-4 py-2 border-t bg-raised/30">
				<span className="text-[11px] text-t2">
					{poll.total_votes}{" "}
					{poll.total_votes === 1
						? "голос"
						: poll.total_votes >= 2 && poll.total_votes <= 4
							? "голоса"
							: "голосов"}
				</span>

				{isCreator && !poll.is_closed && (
					<button
						type="button"
						onClick={() => void handleClose()}
						disabled={voting}
						className="text-[11px] font-medium text-[var(--orange)] hover:text-[var(--orange)]/80 transition-colors disabled:opacity-50"
					>
						Закрыть голосование
					</button>
				)}
			</div>
		</div>
	);
}
