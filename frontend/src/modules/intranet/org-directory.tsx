import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { Building2, Cake, Mail, MapPin, Search, User } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import {
	type DirectoryUser,
	getDirectoryUser,
	listDirectory,
} from "@/api/intranet";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";

function userInitials(name: string): string {
	const parts = name.trim().split(/\s+/).filter(Boolean);
	if (parts.length >= 2) {
		const a = parts[0]?.[0] ?? "";
		const b = parts[1]?.[0] ?? "";
		if (a && b) return (a + b).toUpperCase();
	}
	return name.slice(0, 2).toUpperCase();
}

function ProfileCard({ user }: { user: DirectoryUser }) {
	return (
		<div className="space-y-4">
			<div className="flex items-center gap-4">
				<Avatar className="h-16 w-16">
					<AvatarImage src={user.avatar_url} alt="" />
					<AvatarFallback>{userInitials(user.display_name)}</AvatarFallback>
				</Avatar>
				<div className="min-w-0">
					<p className="truncate text-lg font-semibold">{user.display_name}</p>
					{user.position && (
						<p className="truncate text-sm text-t2">{user.position}</p>
					)}
					{user.department && (
						<p className="truncate text-xs text-t3">{user.department}</p>
					)}
				</div>
			</div>
			<dl className="space-y-2 text-sm">
				<div className="flex items-center gap-2 text-t2">
					<Mail className="h-4 w-4 shrink-0 text-t3" aria-hidden="true" />
					<dt className="sr-only">Email</dt>
					<dd>
						<a href={`mailto:${user.email}`} className="hover:text-t1">
							{user.email}
						</a>
					</dd>
				</div>
				{user.timezone && (
					<div className="flex items-center gap-2 text-t2">
						<MapPin className="h-4 w-4 shrink-0 text-t3" aria-hidden="true" />
						<dt className="sr-only">Часовой пояс</dt>
						<dd>{user.timezone}</dd>
					</div>
				)}
				{user.birthday && (
					<div className="flex items-center gap-2 text-t2">
						<Cake className="h-4 w-4 shrink-0 text-t3" aria-hidden="true" />
						<dt className="sr-only">День рождения</dt>
						<dd>{user.birthday}</dd>
					</div>
				)}
			</dl>
		</div>
	);
}

type OrgDirectoryProps = {
	selectedUserId?: string;
};

/** MOD-02: справочник сотрудников по отделам. */
export function OrgDirectory({ selectedUserId }: OrgDirectoryProps = {}) {
	const navigate = useNavigate();
	const [filter, setFilter] = useState("");
	const [dialogOpen, setDialogOpen] = useState(false);

	const { data, isLoading, isError } = useQuery({
		queryKey: ["intranet", "directory"],
		queryFn: listDirectory,
	});

	const { data: profileUser } = useQuery({
		queryKey: ["intranet", "directory", "user", selectedUserId],
		queryFn: () => getDirectoryUser(selectedUserId as string),
		enabled: Boolean(selectedUserId),
	});

	const filteredDepartments = useMemo(() => {
		if (!data) return [];
		const q = filter.trim().toLowerCase();
		if (!q) return data.departments;

		return data.departments
			.map((group) => ({
				...group,
				users: group.users.filter(
					(u) =>
						u.display_name.toLowerCase().includes(q) ||
						u.email.toLowerCase().includes(q) ||
						(u.position?.toLowerCase().includes(q) ?? false) ||
						group.department.toLowerCase().includes(q),
				),
			}))
			.filter((g) => g.users.length > 0);
	}, [data, filter]);

	useEffect(() => {
		if (selectedUserId) {
			setDialogOpen(true);
		}
	}, [selectedUserId]);

	const openProfile = (userId: string) => {
		void navigate({ to: "/intranet/directory", search: { user: userId } });
	};

	const closeProfile = () => {
		setDialogOpen(false);
		void navigate({ to: "/intranet/directory", search: { user: undefined } });
	};

	return (
		<div className="mx-auto max-w-4xl space-y-6 px-4 py-6">
			<header className="space-y-1">
				<h1 className="text-2xl font-bold tracking-tight">
					Справочник сотрудников
				</h1>
				<p className="text-sm text-t2">
					Организационная структура по отделам · поиск через Ctrl+K
				</p>
			</header>

			<div className="relative">
				<label htmlFor="directory-filter" className="sr-only">
					Поиск сотрудников
				</label>
				<Search
					className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-t3"
					aria-hidden="true"
				/>
				<input
					id="directory-filter"
					type="search"
					value={filter}
					onChange={(e) => setFilter(e.target.value)}
					placeholder="Фильтр по имени, email или отделу"
					className="w-full rounded-lg border border-b1 bg-[var(--hover)]/20 py-2 pl-9 pr-3 text-sm outline-none focus:border-[var(--accent)]"
				/>
			</div>

			{isLoading && (
				<p className="text-sm text-t3" role="status">
					Загрузка справочника…
				</p>
			)}
			{isError && (
				<p className="text-sm text-red-400" role="alert">
					Не удалось загрузить справочник
				</p>
			)}

			{data && (
				<p className="text-xs text-t3">
					{data.total} активных сотрудников
					{filter.trim()
						? ` · показано ${filteredDepartments.reduce((n, g) => n + g.users.length, 0)}`
						: ""}
				</p>
			)}

			<div className="space-y-6">
				{filteredDepartments.map((group) => (
					<section
						key={group.department}
						aria-labelledby={`dept-${group.department}`}
					>
						<h2
							id={`dept-${group.department}`}
							className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-t2"
						>
							<Building2 className="h-4 w-4" aria-hidden="true" />
							{group.department}
							<span className="font-normal text-t3">
								({group.users.length})
							</span>
						</h2>
						<ul className="grid gap-2 sm:grid-cols-2">
							{group.users.map((u) => (
								<li key={u.id}>
									<button
										type="button"
										onClick={() => openProfile(u.id)}
										className={cn(
											"flex w-full items-center gap-3 rounded-xl border border-b1 bg-[var(--hover)]/10 p-3 text-left transition-colors",
											"hover:border-[var(--accent)]/40 hover:bg-[var(--hover)]/30",
											selectedUserId === u.id && "border-[var(--accent)]/60",
										)}
									>
										<Avatar className="h-10 w-10">
											<AvatarImage src={u.avatar_url} alt="" />
											<AvatarFallback className="text-xs">
												{userInitials(u.display_name)}
											</AvatarFallback>
										</Avatar>
										<div className="min-w-0 flex-1">
											<p className="truncate font-medium">{u.display_name}</p>
											{u.position ? (
												<p className="truncate text-xs text-t3">{u.position}</p>
											) : (
												<p className="truncate text-xs text-t3">{u.email}</p>
											)}
										</div>
										<User
											className="h-4 w-4 shrink-0 text-t4"
											aria-hidden="true"
										/>
									</button>
								</li>
							))}
						</ul>
					</section>
				))}
			</div>

			{filteredDepartments.length === 0 && !isLoading && !isError && (
				<p className="text-sm text-t3">Сотрудники не найдены</p>
			)}

			<Dialog
				open={dialogOpen}
				onOpenChange={(open) => !open && closeProfile()}
			>
				<DialogContent className="sm:max-w-md">
					<DialogHeader>
						<DialogTitle>Профиль сотрудника</DialogTitle>
						<DialogDescription className="sr-only">
							Контактная информация и должность
						</DialogDescription>
					</DialogHeader>
					{profileUser ? (
						<ProfileCard user={profileUser} />
					) : selectedUserId ? (
						<p className="text-sm text-t3" role="status">
							Загрузка…
						</p>
					) : null}
					<div className="flex justify-end pt-2">
						<Button type="button" variant="outline" onClick={closeProfile}>
							Закрыть
						</Button>
					</div>
				</DialogContent>
			</Dialog>
		</div>
	);
}
