import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { Menu, Plus } from "lucide-react";
import { useEffect, useState } from "react";
import { useChannelDrawer } from "@/components/messenger/channel-drawer-context";
import { CreateChannelDialogContent } from "@/components/messenger/create-channel-dialog";
import { QueryErrorAlert } from "@/components/query-state-view";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { useChannels } from "@/hooks/use-channels";

/**
 * Дефолтная страница мессенджера.
 *
 * Если у пользователя есть каналы — автоматически переходим в первый,
 * чтобы не показывать пустой экран без способа выбрать канал. Если каналов
 * нет (новый пользователь) — показываем welcoming empty state с CTA на
 * создание канала. На мобилке также даём гамбургер открыть drawer.
 * A3W10-4-001: fail load ≠ welcoming empty.
 */
export function MessengerIndex() {
	const navigate = useNavigate();
	const { data: channels, isLoading, isError, refetch } = useChannels();
	const { setOpen: setDrawerOpen } = useChannelDrawer();
	const [createOpen, setCreateOpen] = useState(false);

	const firstChannelId = channels?.[0]?.id;
	const hasChannels = (channels?.length ?? 0) > 0;

	// Авто-переход в первый канал, чтобы пользователь не залипал на пустом экране.
	useEffect(() => {
		if (firstChannelId) {
			void navigate({
				to: "/messenger/$channelId",
				params: { channelId: firstChannelId },
				replace: true,
			});
		}
	}, [firstChannelId, navigate]);

	// Пока грузятся каналы — нейтральный skeleton, чтобы избежать вспышки empty state.
	if (isLoading || hasChannels) {
		return (
			<div className="flex h-full flex-1 items-center justify-center bg-base">
				<div className="text-center">
					<div className="mx-auto mb-4 h-10 w-10 rounded-full border-2 border-[var(--green)] border-t-transparent animate-spin opacity-60" />
					<p className="text-sm text-t3">Загрузка...</p>
				</div>
			</div>
		);
	}

	// A3W10-4-001: сеть/4xx/5xx ≠ «нет каналов».
	if (isError) {
		return (
			<div className="flex h-full flex-1 items-center justify-center bg-base px-4">
				<QueryErrorAlert
					message="Не удалось загрузить каналы. Попробовать снова?"
					retry={() => void refetch()}
				/>
			</div>
		);
	}

	// Каналов нет — приветственный экран с CTA (только success + empty).
	return (
		<div className="flex h-full flex-1 flex-col bg-base">
			{/* Mobile: мини-шапка с гамбургером — на случай если каналы появятся позже. */}
			<div className="md:hidden h-[58px] shrink-0 flex items-center px-4 border-b border-[var(--b1)]">
				<button
					type="button"
					className="w-[34px] h-[34px] flex items-center justify-center rounded-[var(--r2)] text-[var(--t2)] hover:bg-[var(--hover)] hover:text-[var(--t1)] transition-colors"
					onClick={() => setDrawerOpen(true)}
					aria-label="Открыть список каналов"
					title="Каналы"
				>
					<Menu className="h-5 w-5" />
				</button>
				<span className="ml-3 text-[15px] font-bold tracking-[-0.015em]">
					Мессенджер
				</span>
			</div>

			<div className="flex flex-1 items-center justify-center px-4">
				<div className="text-center max-w-sm animate-fade-in">
					<div className="mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500/10 to-teal-500/10">
						<svg
							className="h-10 w-10 text-emerald-500/60"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={1.5}
							aria-hidden="true"
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
							/>
						</svg>
					</div>
					<h2 className="text-xl font-semibold text-t1 mb-2">
						Добро пожаловать в мессенджер
					</h2>
					<p className="text-sm text-t2 leading-relaxed mb-6">
						У вас пока нет каналов. Создайте первый, чтобы начать общение с
						командой.
					</p>
					<Dialog open={createOpen} onOpenChange={setCreateOpen}>
						<DialogTrigger asChild>
							<button
								type="button"
								className="inline-flex items-center gap-2 rounded-[var(--r2)] bg-gradient-to-br from-[var(--green)] to-[var(--greenS)] px-4 py-2.5 text-sm font-semibold text-white shadow-[0_2px_12px_var(--gG)] hover:scale-[1.02] active:scale-[0.98] transition-transform"
							>
								<Plus className="h-4 w-4" />
								Создать канал
							</button>
						</DialogTrigger>
						<DialogContent>
							<DialogHeader>
								<DialogTitle>Создать канал</DialogTitle>
							</DialogHeader>
							<CreateChannelDialogContent />
						</DialogContent>
					</Dialog>
				</div>
			</div>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/messenger/")({
	component: MessengerIndex,
});
