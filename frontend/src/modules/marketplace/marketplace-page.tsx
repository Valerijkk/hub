import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";
import {
	approveListing,
	CATEGORY_LABELS,
	createListing,
	deleteListing,
	formatListingPrice,
	type ListingCategory,
	listMarketplaceListings,
	listMyListings,
	listPendingListings,
	type MarketplaceListing,
	markListingSold,
	rejectListing,
	STATUS_LABELS,
	updateListing,
	withdrawListing,
} from "@/api/marketplace";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { isSystemAdmin } from "@/lib/is-system-admin";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

/** MOD-08: внутренняя доска объявлений с mod queue. */
export function MarketplacePage() {
	const user = useAuthStore((s) => s.user);
	const isAdmin = isSystemAdmin(user?.system_role);
	const queryClient = useQueryClient();

	const [tab, setTab] = useState<"browse" | "mine">("browse");
	const [category, setCategory] = useState<ListingCategory | "">("");
	const [title, setTitle] = useState("");
	const [description, setDescription] = useState("");
	const [listingCategory, setListingCategory] =
		useState<ListingCategory>("other");
	const [priceRub, setPriceRub] = useState("");
	const [editingId, setEditingId] = useState<string | null>(null);
	const [editTitle, setEditTitle] = useState("");
	const [deleteTarget, setDeleteTarget] = useState<MarketplaceListing | null>(
		null,
	);

	const {
		data: browseData,
		isLoading: browseLoading,
		isError: browseError,
		refetch: refetchBrowse,
	} = useQuery({
		queryKey: ["marketplace", "browse", category],
		queryFn: () => listMarketplaceListings(category || undefined, 1, 50),
		enabled: tab === "browse",
	});

	const {
		data: mineData,
		isError: mineError,
		refetch: refetchMine,
	} = useQuery({
		queryKey: ["marketplace", "mine"],
		queryFn: listMyListings,
		enabled: tab === "mine",
	});

	const { data: pendingData } = useQuery({
		queryKey: ["marketplace", "pending"],
		queryFn: listPendingListings,
		enabled: isAdmin,
	});

	const createMutation = useMutation({
		mutationFn: createListing,
		onSuccess: () => {
			setTitle("");
			setDescription("");
			setPriceRub("");
			void queryClient.invalidateQueries({ queryKey: ["marketplace"] });
			toast.success("Объявление отправлено на модерацию");
		},
		onError: () => toast.error("Не удалось создать объявление"),
	});

	const approveMutation = useMutation({
		mutationFn: ({ id, version }: { id: string; version: number }) =>
			approveListing(id, version),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["marketplace"] });
			toast.success("Объявление опубликовано");
		},
		onError: () => toast.error("Не удалось опубликовать объявление"),
	});

	const rejectMutation = useMutation({
		mutationFn: ({ id, version }: { id: string; version: number }) =>
			rejectListing(id, version, "Не прошло модерацию"),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["marketplace"] });
			toast.success("Объявление отклонено");
		},
		onError: () => toast.error("Не удалось отклонить объявление"),
	});

	const soldMutation = useMutation({
		mutationFn: ({ id, version }: { id: string; version: number }) =>
			markListingSold(id, version),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["marketplace"] });
			toast.success("Отмечено как проданное");
		},
		onError: () => toast.error("Не удалось отметить как проданное"),
	});

	const withdrawMutation = useMutation({
		mutationFn: ({ id, version }: { id: string; version: number }) =>
			withdrawListing(id, version),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["marketplace"] });
			toast.success("Объявление снято");
		},
		onError: () => toast.error("Не удалось снять объявление"),
	});

	const updateMutation = useMutation({
		mutationFn: ({
			id,
			title,
			version,
		}: {
			id: string;
			title: string;
			version: number;
		}) => updateListing(id, { title, version }),
		onSuccess: () => {
			setEditingId(null);
			setEditTitle("");
			void queryClient.invalidateQueries({ queryKey: ["marketplace"] });
			toast.success("Объявление обновлено");
		},
		onError: () => toast.error("Не удалось обновить объявление"),
	});

	const deleteMutation = useMutation({
		mutationFn: (id: string) => deleteListing(id),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["marketplace"] });
			toast.success("Объявление удалено");
		},
		onError: () => toast.error("Не удалось удалить объявление"),
	});

	const handleCreate = (e: React.FormEvent) => {
		e.preventDefault();
		if (!title.trim() || !description.trim()) return;
		let price_cents: number | undefined;
		if (priceRub.trim()) {
			const parsed = Number.parseFloat(priceRub.replace(",", "."));
			if (!Number.isNaN(parsed)) price_cents = Math.round(parsed * 100);
		}
		createMutation.mutate({
			title: title.trim(),
			description: description.trim(),
			category: listingCategory,
			price_cents,
		});
	};

	const browseItems = browseData?.items ?? [];
	const mineItems = mineData?.items ?? [];
	const pendingItems = pendingData?.items ?? [];
	const displayItems = tab === "browse" ? browseItems : mineItems;
	const listError = tab === "browse" ? browseError : mineError;
	const refetchList = tab === "browse" ? refetchBrowse : refetchMine;

	return (
		<div className="mx-auto max-w-4xl space-y-6 px-4 py-6">
			<header className="space-y-1">
				<h1 className="text-2xl font-bold tracking-tight">Барахолка</h1>
				<p className="text-sm text-t2">
					Внутренняя доска объявлений · модерация admin · канал #marketplace
				</p>
			</header>

			<form
				onSubmit={handleCreate}
				className="space-y-3 rounded-xl border border-b1 bg-[var(--hover)]/10 p-4"
			>
				<p className="text-sm font-medium">Новое объявление</p>
				<input
					type="text"
					value={title}
					onChange={(e) => setTitle(e.target.value)}
					placeholder="Заголовок"
					maxLength={200}
					className="w-full rounded-lg border border-b1 px-3 py-2 text-sm"
					required
				/>
				<textarea
					value={description}
					onChange={(e) => setDescription(e.target.value)}
					placeholder="Описание"
					rows={3}
					maxLength={2000}
					className="w-full rounded-lg border border-b1 px-3 py-2 text-sm"
					required
				/>
				<div className="grid gap-3 sm:grid-cols-2">
					<select
						value={listingCategory}
						onChange={(e) =>
							setListingCategory(e.target.value as ListingCategory)
						}
						className="rounded-lg border border-b1 px-3 py-2 text-sm"
					>
						{(Object.keys(CATEGORY_LABELS) as ListingCategory[]).map((c) => (
							<option key={c} value={c}>
								{CATEGORY_LABELS[c]}
							</option>
						))}
					</select>
					<input
						type="text"
						inputMode="decimal"
						value={priceRub}
						onChange={(e) => setPriceRub(e.target.value)}
						placeholder="Цена, RUB (пусто = договорная)"
						className="rounded-lg border border-b1 px-3 py-2 text-sm"
					/>
				</div>
				<Button type="submit" disabled={createMutation.isPending}>
					Опубликовать (модерация)
				</Button>
			</form>

			{isAdmin && pendingItems.length > 0 && (
				<section className="space-y-3 rounded-xl border border-amber-500/30 bg-amber-500/[0.04] p-4">
					<h2 className="text-sm font-semibold">
						Очередь модерации ({pendingItems.length})
					</h2>
					<ul className="space-y-2">
						{pendingItems.map((item) => (
							<li
								key={item.id}
								className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-b1 p-3 bg-[var(--bg)]"
							>
								<div>
									<p className="font-medium">{item.title}</p>
									<p className="text-sm text-t2">
										{item.author_name} · {CATEGORY_LABELS[item.category]}
									</p>
								</div>
								<div className="flex gap-2">
									<Button
										type="button"
										size="sm"
										onClick={() =>
											approveMutation.mutate({
												id: item.id,
												version: item.version,
											})
										}
									>
										Одобрить
									</Button>
									<Button
										type="button"
										size="sm"
										variant="outline"
										onClick={() =>
											rejectMutation.mutate({
												id: item.id,
												version: item.version,
											})
										}
									>
										Отклонить
									</Button>
								</div>
							</li>
						))}
					</ul>
				</section>
			)}

			<div className="flex flex-wrap items-center gap-2 border-b border-b1">
				<button
					type="button"
					onClick={() => setTab("browse")}
					className={cn(
						"px-3 py-2 text-sm font-medium border-b-2 -mb-px",
						tab === "browse"
							? "border-teal-400 text-t1"
							: "border-transparent text-t3",
					)}
				>
					Все объявления
				</button>
				<button
					type="button"
					onClick={() => setTab("mine")}
					className={cn(
						"px-3 py-2 text-sm font-medium border-b-2 -mb-px",
						tab === "mine"
							? "border-teal-400 text-t1"
							: "border-transparent text-t3",
					)}
				>
					Мои
				</button>
				{tab === "browse" && (
					<select
						value={category}
						onChange={(e) =>
							setCategory(e.target.value as ListingCategory | "")
						}
						className="ml-auto text-sm rounded-lg border border-b1 px-2 py-1"
					>
						<option value="">Все категории</option>
						{(Object.keys(CATEGORY_LABELS) as ListingCategory[]).map((c) => (
							<option key={c} value={c}>
								{CATEGORY_LABELS[c]}
							</option>
						))}
					</select>
				)}
			</div>

			{browseLoading && tab === "browse" && (
				<p className="text-sm text-t3">Загрузка…</p>
			)}

			<ul className="grid gap-3 sm:grid-cols-2">
				{displayItems.map((item) => (
					<li
						key={item.id}
						className="rounded-xl border border-b1 p-4 space-y-2 flex flex-col"
					>
						<div className="flex items-start justify-between gap-2">
							<h3 className="font-semibold leading-snug">{item.title}</h3>
							<span className="text-xs text-t3 shrink-0">
								{CATEGORY_LABELS[item.category]}
							</span>
						</div>
						<p className="text-sm text-t2 line-clamp-3 flex-1">
							{item.description}
						</p>
						<div className="flex items-center justify-between text-sm">
							<span className="font-medium text-teal-400">
								{formatListingPrice(item.price_cents, item.currency)}
							</span>
							<span className="text-t3">{item.author_name}</span>
						</div>
						{tab === "mine" && (
							<div className="flex flex-wrap items-center gap-2 pt-1">
								<span className="text-xs text-t3">
									{STATUS_LABELS[item.status]}
								</span>
								{editingId === item.id ? (
									<>
										<input
											type="text"
											value={editTitle}
											onChange={(e) => setEditTitle(e.target.value)}
											aria-label="Новый заголовок объявления"
											className="min-w-[140px] flex-1 rounded-lg border border-b1 bg-[var(--card)] px-2 py-1 text-sm text-[var(--t1)]"
										/>
										<Button
											type="button"
											size="sm"
											disabled={!editTitle.trim() || updateMutation.isPending}
											onClick={() =>
												updateMutation.mutate({
													id: item.id,
													title: editTitle.trim(),
													version: item.version,
												})
											}
										>
											Сохранить
										</Button>
										<Button
											type="button"
											size="sm"
											variant="outline"
											onClick={() => {
												setEditingId(null);
												setEditTitle("");
											}}
										>
											Отмена
										</Button>
									</>
								) : (
									<>
										{(item.status === "pending" ||
											item.status === "approved" ||
											item.status === "rejected") && (
											<Button
												type="button"
												size="sm"
												variant="outline"
												onClick={() => {
													setEditingId(item.id);
													setEditTitle(item.title);
												}}
											>
												Изменить
											</Button>
										)}
										{item.status === "approved" && (
											<Button
												type="button"
												size="sm"
												variant="outline"
												onClick={() =>
													soldMutation.mutate({
														id: item.id,
														version: item.version,
													})
												}
											>
												Продано
											</Button>
										)}
										{(item.status === "pending" ||
											item.status === "approved") && (
											<Button
												type="button"
												size="sm"
												variant="outline"
												onClick={() =>
													withdrawMutation.mutate({
														id: item.id,
														version: item.version,
													})
												}
											>
												Снять
											</Button>
										)}
										<Button
											type="button"
											size="sm"
											variant="outline"
											className="text-[var(--red)]"
											aria-label={`Удалить объявление ${item.title}`}
											onClick={() => setDeleteTarget(item)}
											disabled={deleteMutation.isPending}
										>
											Удалить
										</Button>
									</>
								)}
							</div>
						)}
					</li>
				))}
			</ul>
			{displayItems.length === 0 && !browseLoading && listError && (
				<div role="alert" className="space-y-2 py-8 text-center">
					<p className="text-sm text-[var(--red)]">
						Не удалось загрузить объявления
					</p>
					<button
						type="button"
						className="text-sm font-semibold text-t2 hover:text-t1 hover:underline"
						onClick={() => void refetchList()}
					>
						Попробовать снова
					</button>
				</div>
			)}
			{displayItems.length === 0 && !browseLoading && !listError && (
				<p className="text-sm text-t3 text-center py-8">
					{tab === "browse" ? "Объявлений пока нет" : "У вас нет объявлений"}
				</p>
			)}

			<ConfirmDialog
				open={deleteTarget !== null}
				onOpenChange={(open) => {
					if (!open) setDeleteTarget(null);
				}}
				title="Удалить объявление?"
				description={
					deleteTarget
						? `Объявление «${deleteTarget.title}» будет удалено безвозвратно.`
						: undefined
				}
				confirmLabel="Удалить"
				disabled={deleteMutation.isPending}
				onConfirm={() => {
					if (!deleteTarget) return;
					deleteMutation.mutate(deleteTarget.id);
					setDeleteTarget(null);
				}}
			/>
		</div>
	);
}
