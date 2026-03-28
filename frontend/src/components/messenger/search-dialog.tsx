import { useNavigate } from "@tanstack/react-router";
import { useCallback, useState } from "react";
import { type Message, searchMessages } from "@/api/messenger";
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

/** Пропсы компонента SearchDialog. */
interface SearchDialogProps {
	/** ID текущего канала для фильтрации (опционально). */
	channelId?: string;
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

/**
 * SearchDialog -- диалог полнотекстового поиска по сообщениям.
 * Поддерживает поиск в текущем канале и глобально.
 */
export function SearchDialog({ channelId }: SearchDialogProps) {
	const [open, setOpen] = useState(false);
	const [query, setQuery] = useState("");
	const [scope, setScope] = useState<"channel" | "global">("channel");
	const [results, setResults] = useState<Message[]>([]);
	const [isSearching, setIsSearching] = useState(false);
	const navigate = useNavigate();

	const handleSearch = useCallback(async () => {
		if (!query.trim()) return;
		setIsSearching(true);
		try {
			const scopeChannelId =
				scope === "channel" && channelId ? channelId : undefined;
			const data = await searchMessages(query.trim(), scopeChannelId);
			setResults(data);
		} catch {
			setResults([]);
		} finally {
			setIsSearching(false);
		}
	}, [query, scope, channelId]);

	const handleNavigate = useCallback(
		(msg: Message) => {
			setOpen(false);
			void navigate({
				to: "/messenger/$channelId",
				params: { channelId: msg.channel_id },
			});
		},
		[navigate],
	);

	const handleKeyDown = useCallback(
		(e: React.KeyboardEvent) => {
			if (e.key === "Enter") {
				void handleSearch();
			}
		},
		[handleSearch],
	);

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<Button variant="ghost" size="icon" className="h-8 w-8" title="Поиск">
					<svg
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
			<DialogContent className="max-w-lg">
				<DialogHeader>
					<DialogTitle>Поиск сообщений</DialogTitle>
				</DialogHeader>
				<div className="space-y-3 pt-2">
					<div className="flex gap-2">
						<Input
							placeholder="Введите запрос..."
							value={query}
							onChange={(e) => setQuery(e.target.value)}
							onKeyDown={handleKeyDown}
						/>
						<Button onClick={() => void handleSearch()} disabled={isSearching}>
							{isSearching ? "..." : "Найти"}
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
						{results.length > 0 ? (
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
						) : query && !isSearching ? (
							<p className="py-4 text-center text-sm text-t2">
								Ничего не найдено
							</p>
						) : null}
					</ScrollArea>
				</div>
			</DialogContent>
		</Dialog>
	);
}
