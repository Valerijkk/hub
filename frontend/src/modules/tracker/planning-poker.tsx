import { useCallback, useEffect, useState } from "react";
import * as pokerApi from "@/api/sprint";
import { Button } from "@/components/ui/button";
import { useSprintStore } from "@/stores/sprint-store";

/** Значения Фибоначчи для голосования. */
const FIBONACCI_VALUES = [
	"0",
	"1",
	"2",
	"3",
	"5",
	"8",
	"13",
	"21",
	"34",
	"55",
	"89",
	"?",
];

/** Пропсы компонента PlanningPoker. */
interface PlanningPokerProps {
	/** ID сессии Planning Poker. */
	sessionId: string;
	/** ID текущего пользователя. */
	currentUserId: string;
	/** Может ли пользователь раскрыть голоса (создатель сессии). */
	canReveal?: boolean;
}

/**
 * PlanningPoker отображает интерфейс голосования (SCRM-05).
 * Карточки Фибоначчи, скрытые голоса до раскрытия, средний/медиана после раскрытия.
 */
export function PlanningPoker({
	sessionId,
	currentUserId,
	canReveal = false,
}: PlanningPokerProps) {
	const { pokerSession, setPokerSession } = useSprintStore();
	const [selectedValue, setSelectedValue] = useState<string | null>(null);
	const [voting, setVoting] = useState(false);

	const loadSession = useCallback(async () => {
		try {
			const session = await pokerApi.getPokerSession(sessionId);
			setPokerSession(session);
		} catch {
			// сессия истекла
		}
	}, [sessionId, setPokerSession]);

	useEffect(() => {
		loadSession();
		const interval = setInterval(loadSession, 3000);
		return () => clearInterval(interval);
	}, [loadSession]);

	const handleVote = async (value: string) => {
		setVoting(true);
		try {
			await pokerApi.castVote(sessionId, { value });
			setSelectedValue(value);
			loadSession();
		} finally {
			setVoting(false);
		}
	};

	const handleReveal = async () => {
		try {
			const revealed = await pokerApi.revealVotes(sessionId);
			setPokerSession(revealed);
		} catch {
			// ошибка раскрытия
		}
	};

	const computeStats = (votes: Record<string, string>) => {
		const numericVotes = Object.values(votes)
			.filter((v) => v !== "?")
			.map(Number)
			.filter((n) => !Number.isNaN(n));

		if (numericVotes.length === 0) return null;

		const sum = numericVotes.reduce((a, b) => a + b, 0);
		const avg = sum / numericVotes.length;

		const sorted = [...numericVotes].sort((a, b) => a - b);
		const mid = Math.floor(sorted.length / 2);
		const median =
			sorted.length % 2 !== 0
				? (sorted[mid] ?? 0)
				: ((sorted[mid - 1] ?? 0) + (sorted[mid] ?? 0)) / 2;

		return { avg: Math.round(avg * 10) / 10, median };
	};

	if (!pokerSession) {
		return (
			<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5">
				<p className="text-[13px] text-t2 text-center py-8">
					Сессия не найдена или истекла
				</p>
			</div>
		);
	}

	const stats =
		pokerSession?.revealed && pokerSession?.votes
			? computeStats(pokerSession.votes)
			: null;

	return (
		<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5">
			<h3 className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3 mb-4">
				Planning Poker
			</h3>
			<div className="space-y-6">
				{/* Карточки для голосования */}
				{!pokerSession.revealed && (
					<div className="flex flex-wrap gap-2">
						{FIBONACCI_VALUES.map((value) => (
							<button
								type="button"
								key={value}
								onClick={() => handleVote(value)}
								disabled={voting}
								className={`flex h-16 w-12 items-center justify-center rounded-[var(--r3)] border transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-250 hover:border-b2 hover:-translate-y-1 ${
									selectedValue === value
										? "border-blue shadow-[0_0_0_3px_rgba(91,156,245,0.1),_0_0_16px_rgba(91,156,245,0.06)] bg-blue/10"
										: "bg-gradient-to-b from-card to-raised border-b1 hover:shadow-[var(--sh2)]"
								}`}
							>
								<span className="text-[18px] font-black text-t1">{value}</span>
							</button>
						))}
					</div>
				)}

				{/* Статус голосов */}
				<div className="space-y-2">
					{!pokerSession.revealed && pokerSession.voters && (
						<div className="flex flex-wrap gap-2">
							{pokerSession.voters.map((voterId) => (
								<span
									key={voterId}
									className="bg-emerald-500/[0.08] text-emerald-400 border border-emerald-500/[0.15] px-[10px] py-[3px] rounded-full text-[11px] font-[650]"
								>
									{voterId === currentUserId ? "Вы" : voterId.slice(0, 8)}{" "}
									Проголосовал
								</span>
							))}
						</div>
					)}

					{/* Раскрытые голоса */}
					{pokerSession.revealed && pokerSession.votes && (
						<div className="space-y-3">
							<div className="flex flex-wrap gap-3">
								{Object.entries(pokerSession.votes).map(([userId, value]) => (
									<div
										key={userId}
										className="flex flex-col items-center gap-1"
									>
										<span className="flex h-14 w-10 items-center justify-center rounded-[var(--r3)] border border-blue bg-blue/10 text-[18px] font-black text-t1">
											{value}
										</span>
										<span className="text-[11px] text-t3">
											{userId === currentUserId ? "Вы" : userId.slice(0, 8)}
										</span>
									</div>
								))}
							</div>

							{stats && (
								<div className="flex gap-4 rounded-[var(--r3)] bg-input border border-b1 p-3">
									<div>
										<span className="text-[13px] text-t2">Среднее: </span>
										<span className="text-[14px] font-bold text-t1">
											{stats.avg}
										</span>
									</div>
									<div>
										<span className="text-[13px] text-t2">Медиана: </span>
										<span className="text-[14px] font-bold text-t1">
											{stats.median}
										</span>
									</div>
								</div>
							)}
						</div>
					)}
				</div>

				{/* Кнопка раскрытия */}
				{canReveal && !pokerSession.revealed && (
					<Button
						onClick={handleReveal}
						className="bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-md shadow-blue/20"
					>
						Раскрыть голоса
					</Button>
				)}
			</div>
		</div>
	);
}
