import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import type { Message } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useAuthStore } from "@/stores/auth-store";
import { useSearchDialogSession } from "./search-dialog-session";

/** Пропсы компонента SearchDialog. */
interface SearchDialogProps {
	/** ID текущего канала для фильтрации (опционально). */
	channelId?: string;
}

interface ScopedDialogProps extends SearchDialogProps {
	userId: string | null;
}

interface OpenSearchSessionProps extends ScopedDialogProps {
	onClose: () => void;
}

/** Форматирование даты для результатов поиска. */
function formatDate(iso: string): string {
	const date = new Date(iso);
	return date.toLocaleDateString("ru-RU", {
		day: "2-digit",
		month: "2-digit",
		year: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
	});
}

/** Состояние одной открытой сессии поиска, изолированное от следующих открытий. */
function OpenSearchSession({
	userId,
	channelId,
	onClose,
}: OpenSearchSessionProps) {
	const { t } = useTranslation("common");
	const {
		draft,
		scope,
		setDraft,
		setScope,
		submit,
		retry,
		pending,
		successEmpty,
		error,
		results,
		currentSelectionGuard,
	} = useSearchDialogSession({ userId, channelId });
	const navigate = useNavigate();

	const handleNavigate = (msg: Message) => {
		if (!currentSelectionGuard?.(msg)) return;
		onClose();
		void navigate({
			to: "/messenger/$channelId",
			params: { channelId: msg.channel_id },
		});
	};

	const handleKeyDown = (e: React.KeyboardEvent) => {
		if (e.key === "Enter") {
			submit();
		}
	};

	return (
		<div className="space-y-3 pt-2">
			<div className="flex gap-2">
				<Input
					placeholder="Введите запрос..."
					value={draft}
					onChange={(e) => setDraft(e.target.value)}
					onKeyDown={handleKeyDown}
				/>
				<Button onClick={submit} disabled={pending}>
					{pending ? "..." : "Найти"}
				</Button>
			</div>

			{channelId && (
				<div className="flex gap-2">
					<Button
						variant={scope === "channel" ? "default" : "outline"}
						size="sm"
						onClick={() => setScope("channel")}
					>
						В текущем канале
					</Button>
					<Button
						variant={scope === "global" ? "default" : "outline"}
						size="sm"
						onClick={() => setScope("global")}
					>
						Во всех каналах
					</Button>
				</div>
			)}

			<ScrollArea className="max-h-80">
				{error ? (
					<div role="alert" className="py-4 text-center text-sm text-t2">
						<p>{t("search.error")}</p>
						<p>{t("search.errorHint")}</p>
						<Button type="button" onClick={retry}>
							{t("btn.retry")}
						</Button>
					</div>
				) : results.length > 0 ? (
					<div className="space-y-2">
						{results.map((msg) => (
							<button
								key={msg.id}
								type="button"
								className="w-full rounded-md border p-3 text-left transition-colors hover:bg-accent"
								onClick={() => handleNavigate(msg)}
							>
								<div className="flex items-center justify-between">
									<span className="text-sm font-medium">
										{msg.sender_id}
									</span>
									<span className="text-xs text-t2">
										{formatDate(msg.created_at)}
									</span>
								</div>
								<p className="mt-1 line-clamp-2 text-sm text-t2">
									{msg.content_text || "---"}
								</p>
							</button>
						))}
					</div>
				) : successEmpty ? (
					<p className="py-4 text-center text-sm text-t2">
						Ничего не найдено
					</p>
				) : null}
			</ScrollArea>
		</div>
	);
}

function ScopedDialog({ userId, channelId }: ScopedDialogProps) {
	const [open, setOpen] = useState(false);

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<Button
					variant="ghost"
					size="icon"
					className="h-8 w-8"
					title="Поиск"
					aria-label="Поиск сообщений"
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
							d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
						/>
					</svg>
				</Button>
			</DialogTrigger>
			{open && (
				<DialogContent className="max-w-lg">
					<DialogHeader>
						<DialogTitle>Поиск сообщений</DialogTitle>
					</DialogHeader>
					<OpenSearchSession
						userId={userId}
						channelId={channelId}
						onClose={() => setOpen(false)}
					/>
				</DialogContent>
			)}
		</Dialog>
	);
}

/** Диалог полнотекстового поиска в текущем канале или во всех доступных каналах. */
export function SearchDialog({ channelId }: SearchDialogProps) {
	const userId = useAuthStore((state) =>
		state.isAuthenticated ? (state.user?.id ?? null) : null,
	);

	return (
		<ScopedDialog
			key={JSON.stringify([userId, channelId])}
			userId={userId}
			channelId={channelId}
		/>
	);
}
