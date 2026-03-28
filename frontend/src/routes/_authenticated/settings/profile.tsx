import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { type FormEvent, useState } from "react";
import { toast } from "sonner";
import type { SessionResponse } from "@/api/auth";
import {
	getMe,
	listSessions,
	logout,
	revokeSession,
	updateProfile,
} from "@/api/auth";
import { setCustomStatus } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/stores/auth-store";

/**
 * Страница настроек профиля — полная форма данных пользователя,
 * кастомный статус, часы работы и список активных сессий.
 */
function ProfilePage() {
	const queryClient = useQueryClient();
	const navigate = useNavigate();
	const setUser = useAuthStore((s) => s.setUser);
	const logoutStore = useAuthStore((s) => s.logout);

	const { data: user } = useQuery({
		queryKey: ["me"],
		queryFn: getMe,
	});

	const { data: sessions } = useQuery({
		queryKey: ["sessions"],
		queryFn: listSessions,
	});

	const [form, setForm] = useState({
		name: user?.name ?? "",
		display_name: user?.display_name ?? "",
		phone: user?.phone ?? "",
		position: user?.position ?? "",
		department: user?.department ?? "",
		timezone: user?.timezone ?? "",
		language: user?.language ?? "",
		birthday: user?.birthday ?? "",
	});

	/** Кастомный статус пользователя (например, "На обеде", "В отпуске"). */
	const [customStatusText, setCustomStatusText] = useState("");

	/** Длительность кастомного статуса. */
	const [statusDuration] = useState("1d");

	/* Синхронизация формы при загрузке данных пользователя */
	const [initialized, setInitialized] = useState(false);
	if (user && !initialized) {
		setForm({
			name: user.name ?? "",
			display_name: user.display_name ?? "",
			phone: user.phone ?? "",
			position: user.position ?? "",
			department: user.department ?? "",
			timezone: user.timezone ?? "",
			language: user.language ?? "",
			birthday: user.birthday ?? "",
		});
		setInitialized(true);
	}

	const profileMutation = useMutation({
		mutationFn: updateProfile,
		onSuccess: (updated) => {
			setUser(updated);
			void queryClient.invalidateQueries({ queryKey: ["me"] });
			toast.success("Профиль обновлен");
		},
		onError: () => toast.error("Не удалось обновить профиль"),
	});

	/** Мутация для сохранения кастомного статуса через Messenger API. */
	const statusMutation = useMutation({
		mutationFn: (params: { text: string; duration: string }) =>
			setCustomStatus(params.text, params.duration),
		onSuccess: () => toast.success("Статус обновлён"),
		onError: () => toast.error("Не удалось обновить статус"),
	});

	const revokeMutation = useMutation({
		mutationFn: revokeSession,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["sessions"] });
			toast.success("Сессия завершена");
		},
		onError: () => toast.error("Не удалось завершить сессию"),
	});

	function handleSubmit(e: FormEvent) {
		e.preventDefault();
		profileMutation.mutate({
			name: form.name || undefined,
			display_name: form.display_name || undefined,
			phone: form.phone || undefined,
			position: form.position || undefined,
			department: form.department || undefined,
			timezone: form.timezone || undefined,
			language: form.language || undefined,
			birthday: form.birthday || undefined,
		});
	}

	/** Сохранить кастомный статус. */
	function handleStatusSave() {
		statusMutation.mutate({
			text: customStatusText,
			duration: statusDuration,
		});
	}

	return (
		<div className="space-y-0">
			{/* Profile section */}
			<div className="mb-9">
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em] mb-1">Профиль</h2>
				<p className="text-[13px] text-t2 mb-5">
					Управление личной информацией
				</p>

			{/* Аватар профиля */}
			<div className="flex items-center gap-5 mb-7">
				<div className="group w-24 h-24 rounded-full bg-gradient-to-br from-[var(--blueS)] to-[var(--purpleS)] flex items-center justify-center text-[32px] font-extrabold text-white relative cursor-pointer transition-all duration-250 shadow-[0_4px_20px_var(--gB)] hover:scale-105 hover:shadow-[0_6px_30px_var(--gB)]">
						{user?.display_name?.[0] ?? user?.name?.[0] ?? "?"}
					<div className="absolute inset-0 rounded-full bg-black/50 flex items-center justify-center opacity-0 transition-opacity duration-200 group-hover:opacity-100">
						<svg className="h-6 w-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
							<path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z" />
							<circle cx="12" cy="13" r="4" />
						</svg>
					</div>
				</div>
				<div>
					<button type="button" className="settings-btn-save inline-flex items-center gap-[7px] px-4 py-2 rounded-[var(--r2)] text-[13px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_20px_var(--gB)] transition-all duration-220 hover:shadow-[0_0_32px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96]">
						Загрузить фото
					</button>
					<p className="text-[12px] text-t3 mt-1.5">JPG, PNG. Макс. 2MB</p>
				</div>
			</div>

			{/* Status section */}
			<div className="mb-9">
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em] mb-1">Статус</h2>
				<div className="flex items-center gap-[10px] px-4 py-3 bg-[var(--input)] border border-b1 rounded-[var(--r2)] mb-3">
					<span className="text-[20px] cursor-pointer">{"\uD83D\uDCBB"}</span>
					<input
						placeholder="Установить статус..."
						value={customStatusText}
						onChange={(e) => setCustomStatusText(e.target.value)}
						className="flex-1 bg-transparent border-none text-[14px] text-t1 outline-none placeholder:text-t3"
					/>
							<Button
								onClick={handleStatusSave}
								disabled={statusMutation.isPending}
								variant="outline"
								className="shrink-0 text-[12px]"
							>
								{statusMutation.isPending ? "..." : "OK"}
							</Button>
				</div>
						{/* Быстрый выбор статуса с эмодзи */}
						<div className="flex gap-2 flex-wrap">
							{[
								{ emoji: "\uD83D\uDCF1", label: "На связи" },
								{ emoji: "\uD83C\uDFE0", label: "Из дома" },
								{ emoji: "\uD83C\uDF54", label: "На обеде" },
								{ emoji: "\uD83C\uDFA7", label: "В наушниках" },
								{ emoji: "\uD83D\uDEAB", label: "Не беспокоить" },
								{ emoji: "\uD83C\uDFA6", label: "В отпуске" },
							].map((preset) => (
								<button
									key={preset.label}
									type="button"
									className="px-3 py-[5px] bg-[var(--card)] border border-b1 rounded-full text-[12px] font-semibold text-t2 cursor-pointer transition-all duration-180 hover:border-b2 hover:text-t1 hover:-translate-y-px"
									onClick={() =>
										setCustomStatusText(`${preset.emoji} ${preset.label}`)
									}
								>
									{preset.emoji} {preset.label}
								</button>
							))}
						</div>
			</div>

			{/* Profile form */}
					<form onSubmit={handleSubmit}>
						<div className="grid grid-cols-2 gap-4 mb-4">
							<div>
								<label htmlFor="prof-name" className="block text-[12px] font-semibold text-t2 mb-[6px]">
									Имя
								</label>
								<Input
									id="prof-name"
									value={form.name}
									onChange={(e) =>
										setForm((f) => ({ ...f, name: e.target.value }))
									}
									className="w-full h-[42px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-all duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
							</div>
							<div>
								<label htmlFor="prof-display" className="block text-[12px] font-semibold text-t2 mb-[6px]">
									Отображаемое имя
								</label>
								<Input
									id="prof-display"
									value={form.display_name}
									onChange={(e) =>
										setForm((f) => ({ ...f, display_name: e.target.value }))
									}
									className="w-full h-[42px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-all duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
							</div>
						</div>
						<div className="grid grid-cols-2 gap-4 mb-4">
							<div>
								<label htmlFor="prof-position" className="block text-[12px] font-semibold text-t2 mb-[6px]">
									Должность
								</label>
								<Input
									id="prof-position"
									value={form.position}
									onChange={(e) =>
										setForm((f) => ({ ...f, position: e.target.value }))
									}
									className="w-full h-[42px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-all duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
							</div>
							<div>
								<label htmlFor="prof-phone" className="block text-[12px] font-semibold text-t2 mb-[6px]">
									Телефон
								</label>
								<Input
									id="prof-phone"
									value={form.phone}
									onChange={(e) =>
										setForm((f) => ({ ...f, phone: e.target.value }))
									}
									className="w-full h-[42px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-all duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
							</div>
						</div>
						<div className="grid grid-cols-2 gap-4 mb-4">
							<div>
								<label className="block text-[12px] font-semibold text-t2 mb-[6px]">
									Email
								</label>
								<input
									value={user?.email ?? ""}
									disabled
									className="w-full h-[42px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 opacity-50 outline-none"
								/>
							</div>
							<div>
								<label htmlFor="prof-tz" className="block text-[12px] font-semibold text-t2 mb-[6px]">
									Часовой пояс
								</label>
								<Input
									id="prof-tz"
									value={form.timezone}
									onChange={(e) =>
										setForm((f) => ({ ...f, timezone: e.target.value }))
									}
									placeholder="Europe/Moscow"
									className="w-full h-[42px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-all duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
							</div>
						</div>
						<div className="grid grid-cols-2 gap-4 mb-4">
							<div>
								<label htmlFor="prof-lang" className="block text-[12px] font-semibold text-t2 mb-[6px]">
									Язык
								</label>
								<Input
									id="prof-lang"
									value={form.language}
									onChange={(e) =>
										setForm((f) => ({ ...f, language: e.target.value }))
									}
									placeholder="ru"
									className="w-full h-[42px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-all duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
							</div>
							<div>
								<label htmlFor="prof-birthday" className="block text-[12px] font-semibold text-t2 mb-[6px]">
									День рождения
								</label>
								<Input
									id="prof-birthday"
									type="date"
									value={form.birthday}
									onChange={(e) =>
										setForm((f) => ({ ...f, birthday: e.target.value }))
									}
									className="w-full h-[42px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-all duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
							</div>
						</div>
					</form>
			</div>

			{/* Sessions section */}
			<div className="mb-9">
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em] mb-1">Сессии</h2>

						{sessions?.map((session: SessionResponse) => (
						<div
							key={session.id}
							className="flex items-center gap-[14px] py-[14px] border-b border-b1"
						>
								<svg
									className="w-[22px] h-[22px] text-t3 shrink-0"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									strokeWidth="1.8"
								>
									<rect x="2" y="3" width="20" height="14" rx="2" />
									<line x1="8" y1="21" x2="16" y2="21" />
									<line x1="12" y1="17" x2="12" y2="21" />
								</svg>
							<div className="flex-1">
									<div className="flex items-center">
										<span className="text-[14px] font-semibold">
												{session.device || "Неизвестное устройство"}
										</span>
											{session.is_current && (
												<span className="bg-green/[0.08] text-green px-[10px] py-[2px] rounded-full text-[11px] font-bold ml-2">
													Текущая
												</span>
											)}
									</div>
										<p className="text-[12px] text-t3">
											{session.ip_address ?? "-"},{" "}
											{new Date(session.last_active_at).toLocaleString("ru-RU")}
										</p>
							</div>
								{!session.is_current && (
									<button
										type="button"
										onClick={() => revokeMutation.mutate(session.id)}
										disabled={revokeMutation.isPending}
										className="px-[14px] py-[5px] rounded-[var(--r2)] text-[12px] font-bold text-red border border-red/[0.15] transition-all duration-180 hover:bg-red/[0.06]"
									>
										Завершить
									</button>
								)}
						</div>
					))}
					{(!sessions || sessions.length === 0) && (
						<p className="py-4 text-center text-[13px] text-t3">
							Нет активных сессий
						</p>
					)}
			</div>

			{/* Save + Logout */}
			<div className="flex items-center gap-4 mt-6">
				<button
					type="button"
					onClick={(e) => { e.preventDefault(); handleSubmit(e as unknown as FormEvent); }}
					disabled={profileMutation.isPending}
					className="settings-btn-save inline-flex items-center gap-[7px] px-6 py-[10px] rounded-[var(--r2)] text-[14px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_20px_var(--gB)] transition-all duration-220 hover:shadow-[0_0_32px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96]"
				>
					{profileMutation.isPending ? "Сохранение..." : "Сохранить изменения"}
				</button>
				<button
					type="button"
					className="px-[14px] py-[5px] rounded-[var(--r2)] text-[12px] font-bold text-red border border-red/[0.15] transition-all duration-180 hover:bg-red/[0.06]"
					onClick={async () => {
						try {
							await logout();
						} catch {
							// ignore
						}
						logoutStore();
						void navigate({ to: "/login" });
					}}
				>
					Выйти из аккаунта
				</button>
			</div>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/settings/profile")({
	component: ProfilePage,
});
