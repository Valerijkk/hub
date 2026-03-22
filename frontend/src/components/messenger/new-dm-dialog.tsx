import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useMemo, useState } from "react";
import { api } from "@/api/client";
import { createOrGetDM } from "@/api/messenger";
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

/** Пользователь для выбора в диалоге создания личного сообщения. */
interface PickerUser {
	id: string;
	email: string;
	name?: string;
	display_name?: string;
}

/** Ответ от эндпоинта /admin/users для получения списка пользователей. */
interface UsersListResponse {
	items?: PickerUser[];
	users?: PickerUser[];
}

/**
 * Диалог "Написать" -- выбор пользователя и открытие/создание ЛС-канала.
 */
export function NewDmDialog() {
	const [open, setOpen] = useState(false);
	const [users, setUsers] = useState<PickerUser[]>([]);
	const [loading, setLoading] = useState(false);
	const [search, setSearch] = useState("");
	const [creating, setCreating] = useState<string | null>(null);
	const navigate = useNavigate();
	const currentUserId = useAuthStore((s) => s.user?.id);
	const queryClient = useQueryClient();

	useEffect(() => {
		if (!open) return;
		setLoading(true);
		api
			.get<UsersListResponse | PickerUser[]>("/admin/users")
			.then((data) => {
				let list: PickerUser[] = [];
				if (Array.isArray(data)) {
					list = data;
				} else if (data.items) {
					list = data.items;
				} else if (data.users) {
					list = data.users;
				}
				setUsers(list.filter((u) => u.id !== currentUserId));
			})
			.catch(() => setUsers([]))
			.finally(() => setLoading(false));
	}, [open, currentUserId]);

	const filtered = useMemo(() => {
		if (!search.trim()) return users;
		const q = search.toLowerCase();
		return users.filter(
			(u) =>
				(u.name && u.name.toLowerCase().includes(q)) ||
				(u.display_name && u.display_name.toLowerCase().includes(q)) ||
				u.email.toLowerCase().includes(q),
		);
	}, [users, search]);

	const displayName = (u: PickerUser) => u.display_name || u.name || u.email;

	const handleSelect = async (userId: string) => {
		setCreating(userId);
		try {
			const channel = await createOrGetDM(userId);
			void queryClient.invalidateQueries({
				queryKey: ["messenger", "channels"],
			});
			setOpen(false);
			setSearch("");
			void navigate({
				to: "/messenger/$channelId",
				params: { channelId: channel.id },
			});
		} finally {
			setCreating(null);
		}
	};

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<button
					type="button"
					className="inline-flex h-7 w-7 items-center justify-center rounded-md text-sidebar-foreground/60 hover:bg-white/10 hover:text-sidebar-foreground transition-colors"
					title="Написать"
				>
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
							d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
						/>
					</svg>
				</button>
			</DialogTrigger>
			<DialogContent className="max-w-sm">
				<DialogHeader>
					<DialogTitle>Написать</DialogTitle>
				</DialogHeader>
				<div className="space-y-3 pt-1">
					<Input
						placeholder="Поиск по имени или email..."
						value={search}
						onChange={(e) => setSearch(e.target.value)}
						autoFocus
					/>
					<ScrollArea className="max-h-64">
						{loading ? (
							<div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
								Загрузка...
							</div>
						) : filtered.length === 0 ? (
							<div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
								{search ? "Никого не найдено" : "Нет пользователей"}
							</div>
						) : (
							<div className="space-y-0.5">
								{filtered.map((u) => (
									<button
										key={u.id}
										type="button"
										disabled={creating === u.id}
										onClick={() => handleSelect(u.id)}
										className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm hover:bg-accent transition-colors disabled:opacity-50"
									>
										<div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-medium text-primary">
											{displayName(u).charAt(0).toUpperCase()}
										</div>
										<div className="flex flex-col items-start overflow-hidden">
											<span className="truncate font-medium">
												{displayName(u)}
											</span>
											{u.email && (
												<span className="truncate text-xs text-muted-foreground">
													{u.email}
												</span>
											)}
										</div>
										{creating === u.id && (
											<span className="ml-auto text-xs text-muted-foreground">
												...
											</span>
										)}
									</button>
								))}
							</div>
						)}
					</ScrollArea>
				</div>
			</DialogContent>
		</Dialog>
	);
}
