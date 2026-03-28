/**
 * Панель быстрого входа для разработки.
 *
 * Сетка карточек пользователей с цветными акцентами по ролям.
 * Чистый дизайн — тонкие бордеры, подъём при наведении, аватары с инициалами.
 */
import { useEffect, useState } from "react";
import type { TokenPairResponse, UserResponse } from "@/api/auth";
import { api } from "@/api/client";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/stores/auth-store";

/* --- Типы --- */

interface DevLoginResponse extends TokenPairResponse {
	user: UserResponse;
}

interface DevSeedResponse {
	users: Array<{
		id: string;
		email: string;
		name: string;
		system_role: string;
		project_role?: string;
		description: string;
	}>;
	project: { id: string; name: string; key: string };
}

/* --- Демо-пользователи --- */

interface DemoUser {
	email: string;
	name: string;
	role: string;
	projectRole?: string;
	description: string;
	/** Цвет аватара */
	color: string;
	/** Цвет бейджа */
	badgeBg: string;
	badgeText: string;
}

const DEMO_USERS: DemoUser[] = [
	{
		email: "admin@hub.local",
		name: "Админ Хабов",
		role: "Admin",
		description: "Полный доступ",
		color: "bg-red-500",
		badgeBg: "bg-red-500/10",
		badgeText: "text-red-600 dark:text-red-400",
	},
	{
		email: "pm@hub.local",
		name: "Петр Менеджеров",
		role: "Owner",
		projectRole: "owner",
		description: "Project Manager",
		color: "bg-purple-500",
		badgeBg: "bg-purple-500/10",
		badgeText: "text-purple-600 dark:text-purple-400",
	},
	{
		email: "dev1@hub.local",
		name: "Дмитрий Кодов",
		role: "Developer",
		projectRole: "developer",
		description: "Frontend разработчик",
		color: "bg-blue-500",
		badgeBg: "bg-blue-500/10",
		badgeText: "text-blue-600 dark:text-blue-400",
	},
	{
		email: "dev2@hub.local",
		name: "Анна Бэкендова",
		role: "Developer",
		projectRole: "developer",
		description: "Backend разработчик",
		color: "bg-cyan-500",
		badgeBg: "bg-blue-500/10",
		badgeText: "text-blue-600 dark:text-blue-400",
	},
	{
		email: "tester1@hub.local",
		name: "Тестер Тестеров",
		role: "Tester",
		projectRole: "tester",
		description: "QA инженер",
		color: "bg-emerald-500",
		badgeBg: "bg-emerald-500/10",
		badgeText: "text-emerald-600 dark:text-emerald-400",
	},
	{
		email: "tester2@hub.local",
		name: "Мария Автотестова",
		role: "Tester",
		projectRole: "tester",
		description: "Автоматизатор",
		color: "bg-green-500",
		badgeBg: "bg-emerald-500/10",
		badgeText: "text-emerald-600 dark:text-emerald-400",
	},
	{
		email: "viewer@hub.local",
		name: "Наблюдатель Смотров",
		role: "Viewer",
		projectRole: "viewer",
		description: "Только просмотр",
		color: "bg-slate-400",
		badgeBg: "bg-slate-500/10",
		badgeText: "text-slate-600 dark:text-slate-400",
	},
	{
		email: "newbie@hub.local",
		name: "Новичок Первовходов",
		role: "Member",
		description: "Не прошел онбординг",
		color: "bg-amber-500",
		badgeBg: "bg-amber-500/10",
		badgeText: "text-amber-600 dark:text-amber-400",
	},
];

/* --- Утилиты --- */

function getInitials(name: string): string {
	return name
		.split(" ")
		.map((w) => w[0])
		.join("")
		.toUpperCase()
		.slice(0, 2);
}

/* --- Компонент --- */

interface DevLoginPanelProps {
	onSuccess: () => void;
}

export function DevLoginPanel({ onSuccess }: DevLoginPanelProps) {
	const [devMode, setDevMode] = useState<boolean | null>(null);
	const [seeded, setSeeded] = useState(false);
	const [loading, setLoading] = useState<string | null>(null);
	const [error, setError] = useState<string | null>(null);
	const setTokens = useAuthStore((s) => s.setTokens);
	const setUser = useAuthStore((s) => s.setUser);

	/* Проверяем доступность dev-режима */
	useEffect(() => {
		api
			.get<{ dev_mode: boolean }>("/dev/status")
			.then(() => setDevMode(true))
			.catch(() => setDevMode(false));
	}, []);

	if (devMode === null || devMode === false) {
		return null;
	}

	async function handleSeed() {
		setLoading("seed");
		setError(null);
		try {
			await api.post<DevSeedResponse>("/dev/seed");
			setSeeded(true);
		} catch (e: unknown) {
			const msg = e instanceof Error ? e.message : "Ошибка при создании данных";
			setError(msg);
		} finally {
			setLoading(null);
		}
	}

	async function handleLogin(email: string) {
		setLoading(email);
		setError(null);
		try {
			if (!seeded) {
				await api.post<DevSeedResponse>("/dev/seed");
				setSeeded(true);
			}

			const result = await api.post<DevLoginResponse>("/dev/login", { email });
			setTokens(result.access_token, result.refresh_token);
			setUser(result.user);
			onSuccess();
		} catch (e: unknown) {
			const msg = e instanceof Error ? e.message : "Ошибка входа";
			setError(msg);
		} finally {
			setLoading(null);
		}
	}

	return (
		<div className="space-y-4 rounded-2xl border border-dashed border-amber-500/30 bg-amber-500/[0.03] p-5">
			{/* Заголовок Демо-доступ */}
			<div className="flex items-center justify-center gap-2">
				<div className="h-px flex-1 bg-amber-500/20" />
				<span className="text-[11px] font-semibold uppercase tracking-wider text-amber-600 dark:text-amber-400">
					Демо-доступ
				</span>
				<div className="h-px flex-1 bg-amber-500/20" />
			</div>

			{/* Кнопка создания демо-данных */}
			<div className="flex justify-center">
				<Button
					variant="outline"
					size="sm"
					onClick={handleSeed}
					disabled={loading === "seed" || seeded}
					className="text-xs"
				>
					{seeded
						? "Данные созданы"
						: loading === "seed"
							? "Создание..."
							: "Создать демо-данные"}
				</Button>
			</div>

			{error && <p className="text-center text-xs text-destructive">{error}</p>}

			{/* Сетка карточек пользователей — с градиентными акцентами */}
			<div className="grid grid-cols-2 gap-2.5">
				{DEMO_USERS.map((user) => (
					<button
						key={user.email}
						onClick={() => handleLogin(user.email)}
						disabled={loading !== null}
						className="group relative flex items-center gap-3 rounded-xl border border-border bg-card p-3 text-left
              transition-all duration-200
              hover:border-primary/30 hover:shadow-md hover:-translate-y-0.5
              active:scale-[0.97] active:shadow-none
              disabled:opacity-50 disabled:pointer-events-none
              overflow-hidden"
					>
						{/* Тонкая градиентная полоска слева */}
						<div
							className={`absolute left-0 top-0 bottom-0 w-[3px] ${user.color} opacity-0 group-hover:opacity-100 transition-opacity duration-200 rounded-l-xl`}
						/>

						{/* Аватар с инициалами — увеличенный */}
						<div
							className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full ${user.color} text-[11px] font-bold text-white shadow-sm`}
						>
							{getInitials(user.name)}
						</div>

						<div className="min-w-0 flex-1">
							<p className="text-[13px] font-semibold leading-tight truncate text-t1">
								{user.name}
							</p>
							<div className="mt-1 flex items-center gap-1.5">
								<Badge
									variant="secondary"
									className={`${user.badgeBg} ${user.badgeText} text-[10px] leading-none px-1.5 py-0.5 border-0 font-medium`}
								>
									{user.role}
								</Badge>
							</div>
						</div>
					</button>
				))}
			</div>

			{loading && loading !== "seed" && (
				<p className="text-center text-xs text-t2">
					Вход как {DEMO_USERS.find((u) => u.email === loading)?.name}...
				</p>
			)}
		</div>
	);
}
