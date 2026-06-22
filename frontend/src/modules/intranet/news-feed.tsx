import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Pin, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { createNews, deleteNews, listNews } from "@/api/intranet";
import { Button } from "@/components/ui/button";
import { isSystemAdmin } from "@/lib/is-system-admin";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

/** MOD-01: корпоративная лента новостей (Hub One v4). */
export function NewsFeed({ embedded = false }: { embedded?: boolean }) {
	const user = useAuthStore((s) => s.user);
	const isAdmin = isSystemAdmin(user?.system_role);
	const queryClient = useQueryClient();
	const [title, setTitle] = useState("");
	const [body, setBody] = useState("");
	const [pinned, setPinned] = useState(false);

	const { data, isLoading, isError } = useQuery({
		queryKey: ["intranet", "news"],
		queryFn: () => listNews(1, 50),
	});

	const createMutation = useMutation({
		mutationFn: createNews,
		onSuccess: () => {
			setTitle("");
			setBody("");
			setPinned(false);
			void queryClient.invalidateQueries({ queryKey: ["intranet", "news"] });
			toast.success("Новость опубликована");
		},
		onError: () => toast.error("Не удалось создать новость"),
	});

	const deleteMutation = useMutation({
		mutationFn: deleteNews,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["intranet", "news"] });
			toast.success("Новость удалена");
		},
		onError: () => toast.error("Не удалось удалить"),
	});

	const handleCreate = (e: React.FormEvent) => {
		e.preventDefault();
		if (!title.trim() || !body.trim()) return;
		createMutation.mutate({
			title: title.trim(),
			body: body.trim(),
			pinned,
			published: true,
		});
	};

	return (
		<div
			className={
				embedded ? "space-y-6" : "mx-auto max-w-3xl space-y-6 px-4 py-6"
			}
		>
			{!embedded && (
				<header className="space-y-1">
					<h1 className="text-2xl font-bold tracking-tight">
						Новости компании
					</h1>
					<p className="text-sm text-t2">
						Корпоративная лента — объявления HR и администрации
					</p>
				</header>
			)}

			{isAdmin && (
				<form
					onSubmit={handleCreate}
					className="space-y-3 rounded-xl border border-b1 bg-[var(--hover)]/20 p-4"
				>
					<p className="text-sm font-medium">Новая публикация</p>
					<label htmlFor="news-title" className="sr-only">
						Заголовок
					</label>
					<input
						id="news-title"
						type="text"
						value={title}
						onChange={(e) => setTitle(e.target.value)}
						placeholder="Заголовок"
						maxLength={200}
						className="w-full h-9 px-3 bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[13px]"
					/>
					<label htmlFor="news-body" className="sr-only">
						Текст
					</label>
					<textarea
						id="news-body"
						value={body}
						onChange={(e) => setBody(e.target.value)}
						placeholder="Текст новости"
						rows={4}
						className="w-full px-3 py-2 bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[13px] resize-y"
					/>
					<div className="flex flex-wrap items-center gap-3">
						<label className="flex items-center gap-2 text-[13px] text-t2 cursor-pointer">
							<input
								type="checkbox"
								checked={pinned}
								onChange={(e) => setPinned(e.target.checked)}
								className="rounded border-b1"
							/>
							Закрепить
						</label>
						<div className="flex-1" />
						<Button
							type="submit"
							disabled={
								createMutation.isPending || !title.trim() || !body.trim()
							}
						>
							Опубликовать
						</Button>
					</div>
				</form>
			)}

			{isLoading && <p className="text-sm text-t3">Загрузка…</p>}
			{isError && (
				<p className="text-sm text-red">Не удалось загрузить новости</p>
			)}

			<div className="space-y-4">
				{data?.items.map((post) => (
					<article
						key={post.id}
						className={cn(
							"rounded-xl border border-b1 p-4 space-y-2",
							post.pinned && "border-amber-500/40 bg-amber-500/[0.04]",
						)}
					>
						<div className="flex items-start gap-2">
							{post.pinned && (
								<Pin
									className="h-4 w-4 shrink-0 text-amber-400 mt-0.5"
									aria-label="Закреплено"
								/>
							)}
							<div className="flex-1 min-w-0">
								<h2 className="text-lg font-semibold leading-snug">
									{post.title}
								</h2>
								<p className="text-[12px] text-t3 mt-1">
									{post.author_name ?? "Admin"} ·{" "}
									{new Date(post.created_at).toLocaleString("ru-RU")}
								</p>
							</div>
							{isAdmin && (
								<button
									type="button"
									onClick={() => deleteMutation.mutate(post.id)}
									className="p-1.5 text-t3 hover:text-red transition-colors rounded"
									aria-label="Удалить новость"
								>
									<Trash2 className="h-4 w-4" />
								</button>
							)}
						</div>
						<p className="text-sm text-t1 whitespace-pre-wrap leading-relaxed">
							{post.body}
						</p>
					</article>
				))}
				{data && data.items.length === 0 && (
					<p className="text-sm text-t3 text-center py-12">Пока нет новостей</p>
				)}
			</div>
		</div>
	);
}
