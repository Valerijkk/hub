import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Heart, Send } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import {
	createKudo,
	type DirectoryUser,
	listDirectory,
	listKudosFeed,
	listReceivedKudos,
} from "@/api/intranet";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

/** MOD-03: kudos / thanks — публичная лента + уведомление получателю. */
export function KudosFeed() {
	const user = useAuthStore((s) => s.user);
	const queryClient = useQueryClient();
	const [recipientId, setRecipientId] = useState("");
	const [message, setMessage] = useState("");
	const [isPublic, setIsPublic] = useState(true);
	const [tab, setTab] = useState<"feed" | "received">("feed");

	const { data: directory } = useQuery({
		queryKey: ["intranet", "directory"],
		queryFn: listDirectory,
	});

	const { data: feed, isLoading: feedLoading } = useQuery({
		queryKey: ["intranet", "kudos", "feed"],
		queryFn: () => listKudosFeed(1, 50),
		enabled: tab === "feed",
	});

	const { data: received, isLoading: receivedLoading } = useQuery({
		queryKey: ["intranet", "kudos", "received"],
		queryFn: () => listReceivedKudos(1, 50),
		enabled: tab === "received",
	});

	const createMutation = useMutation({
		mutationFn: createKudo,
		onSuccess: () => {
			setMessage("");
			void queryClient.invalidateQueries({ queryKey: ["intranet", "kudos"] });
			toast.success("Спасибо отправлено!");
		},
		onError: () => toast.error("Не удалось отправить"),
	});

	const colleagues: DirectoryUser[] =
		directory?.departments.flatMap((d) => d.users) ?? [];

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (!recipientId || !message.trim()) return;
		createMutation.mutate({
			recipient_id: recipientId,
			message: message.trim(),
			is_public: isPublic,
		});
	};

	const items = tab === "feed" ? feed?.items : received?.items;
	const isLoading = tab === "feed" ? feedLoading : receivedLoading;

	return (
		<div className="space-y-6">
			<form
				onSubmit={handleSubmit}
				className="space-y-3 rounded-xl border border-b1 bg-[var(--hover)]/10 p-4"
			>
				<p className="text-sm font-medium flex items-center gap-2">
					<Heart className="h-4 w-4 text-pink-400" aria-hidden="true" />
					Сказать спасибо коллеге
				</p>
				<label className="block space-y-1 text-sm">
					<span className="text-t2">Кому</span>
					<select
						value={recipientId}
						onChange={(e) => setRecipientId(e.target.value)}
						className="w-full rounded-lg border border-b1 px-3 py-2"
						required
					>
						<option value="">Выберите коллегу</option>
						{colleagues
							.filter((c) => c.id !== user?.id)
							.map((c) => (
								<option key={c.id} value={c.id}>
									{c.display_name}
									{c.department ? ` · ${c.department}` : ""}
								</option>
							))}
					</select>
				</label>
				<label className="block space-y-1 text-sm">
					<span className="text-t2">Сообщение</span>
					<textarea
						value={message}
						onChange={(e) => setMessage(e.target.value)}
						rows={3}
						maxLength={500}
						placeholder="За что благодарите?"
						className="w-full rounded-lg border border-b1 px-3 py-2"
						required
					/>
				</label>
				<label className="flex items-center gap-2 text-sm text-t2">
					<input
						type="checkbox"
						checked={isPublic}
						onChange={(e) => setIsPublic(e.target.checked)}
					/>
					Показать в публичной ленте
				</label>
				<Button type="submit" disabled={createMutation.isPending}>
					<Send className="h-4 w-4 mr-1.5" aria-hidden="true" />
					Отправить
				</Button>
			</form>

			<div className="flex gap-2 border-b border-b1">
				<button
					type="button"
					onClick={() => setTab("feed")}
					className={cn(
						"px-3 py-2 text-sm font-medium border-b-2 -mb-px transition-colors",
						tab === "feed"
							? "border-pink-400 text-t1"
							: "border-transparent text-t3 hover:text-t2",
					)}
				>
					Лента команды
				</button>
				<button
					type="button"
					onClick={() => setTab("received")}
					className={cn(
						"px-3 py-2 text-sm font-medium border-b-2 -mb-px transition-colors",
						tab === "received"
							? "border-pink-400 text-t1"
							: "border-transparent text-t3 hover:text-t2",
					)}
				>
					Мне сказали спасибо
				</button>
			</div>

			{isLoading && <p className="text-sm text-t3">Загрузка…</p>}
			<ul className="space-y-3">
				{items?.map((k) => (
					<li
						key={k.id}
						className="rounded-xl border border-b1 p-4 space-y-1 bg-pink-500/[0.03]"
					>
						<p className="text-sm">
							<span className="font-medium">{k.sender_name}</span>
							<span className="text-t3"> → </span>
							<span className="font-medium">{k.recipient_name}</span>
						</p>
						<p className="text-sm text-t1 whitespace-pre-wrap">{k.message}</p>
						<p className="text-[12px] text-t3">
							{new Date(k.created_at).toLocaleString("ru-RU")}
						</p>
					</li>
				))}
				{!isLoading && items && items.length === 0 && (
					<li className="text-sm text-t3 text-center py-8">
						{tab === "feed" ? "Лента пока пуста" : "Пока никто не поблагодарил"}
					</li>
				)}
			</ul>
		</div>
	);
}
