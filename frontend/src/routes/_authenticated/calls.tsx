import { createFileRoute } from "@tanstack/react-router";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CallCalendar } from "@/modules/calls/call-calendar";
import { useCallStore } from "@/stores/call-store";

/**
 * Страница "Созвоны" — календарь и информация о звонках.
 * Звонки начинаются из каналов мессенджера (кнопка в шапке чата).
 */
function CallsPage() {
	const { t } = useTranslation("calls");
	const startInstantCall = useCallStore((s) => s.startInstantCall);
	const [creatingCall, setCreatingCall] = useState(false);

	const handleInstantCall = useCallback(async () => {
		setCreatingCall(true);
		const inviteCode = await startInstantCall();
		setCreatingCall(false);
		if (inviteCode) {
			const url = `${window.location.origin}/call/${inviteCode}`;
			// Запасной вариант для HTTP — clipboard API доступен только по HTTPS
			try {
				await navigator.clipboard.writeText(url);
			} catch {
				// запасной prompt для ручного копирования
				window.prompt("Скопируйте ссылку:", url);
			}
			toast.success("Звонок создан!", {
				description: url,
				action: {
					label: "Копировать",
					onClick: () => {
						navigator.clipboard.writeText(url).catch(() => {
							window.prompt("Скопируйте ссылку:", url);
						});
					},
				},
				duration: 15000,
			});
		}
	}, [startInstantCall]);

	return (
		<div className="flex-1 overflow-y-auto flex flex-col animate-[eu_450ms_var(--spring)_both]">
			<div className="flex items-center justify-between px-6 pt-[18px] pb-[14px]">
				<h1 className="text-[18px] font-extrabold tracking-[-0.03em] bg-gradient-to-r from-purple to-blue bg-clip-text text-transparent">
					{t("title", { defaultValue: "Звонки" })}
				</h1>
				<button
					type="button"
					disabled={creatingCall}
					onClick={() => void handleInstantCall()}
					className="calls-btn-p inline-flex items-center gap-[7px] px-[18px] py-2 rounded-[var(--r2)] text-[13px] font-semibold bg-gradient-to-br from-[var(--purpleS)] to-[#6d28d9] text-white shadow-[0_0_20px_var(--gP)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 hover:shadow-[0_0_32px_rgba(167,139,250,0.3)] hover:-translate-y-px active:scale-[0.96] relative overflow-hidden disabled:opacity-50"
				>
					<svg
						aria-hidden="true"
						className="h-3.5 w-3.5"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2.5}
					>
						<line x1="12" y1="5" x2="12" y2="19" />
						<line x1="5" y1="12" x2="19" y2="12" />
					</svg>
					{creatingCall
						? "Создание..."
						: t("instantCall", { defaultValue: "Новый звонок" })}
				</button>
			</div>

			<Tabs defaultValue="calendar" className="px-6">
				<TabsList>
					<TabsTrigger value="calendar">
						{t("tabCalendar", { defaultValue: "Календарь" })}
					</TabsTrigger>
					<TabsTrigger value="info">
						{t("tabInfo", { defaultValue: "О звонках" })}
					</TabsTrigger>
				</TabsList>

				<TabsContent value="calendar">
					<CallCalendar />
				</TabsContent>

				<TabsContent value="info">
					<section className="rounded-xl border bg-card p-6 space-y-6">
						<div>
							<h2 className="text-lg font-semibold mb-4">Как начать звонок</h2>
							<div className="grid gap-4 md:grid-cols-3">
								{[
									{
										n: "1",
										color: "blue",
										title: "Откройте канал",
										desc: "Перейдите в мессенджер и выберите канал",
									},
									{
										n: "2",
										color: "green",
										title: "Нажмите кнопку звонка",
										desc: "Иконка телефона в шапке чата",
									},
									{
										n: "3",
										color: "violet",
										title: "Общайтесь!",
										desc: "Голос + screen share, до 50 участников",
									},
								].map((s) => (
									<div key={s.n} className="flex gap-3">
										<div
											className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-${s.color}-500/10 text-${s.color}-500 text-sm font-bold`}
										>
											{s.n}
										</div>
										<div>
											<p className="font-medium text-sm">{s.title}</p>
											<p className="text-xs text-t2 mt-0.5">{s.desc}</p>
										</div>
									</div>
								))}
							</div>
						</div>
						<div className="grid gap-3 md:grid-cols-2">
							{[
								{ icon: "🎙️", title: "Аудиозвонки", desc: "До 50 участников" },
								{
									icon: "🖥️",
									title: "Screen Share",
									desc: "Демонстрация экрана",
								},
								{
									icon: "🔴",
									title: "Запись",
									desc: "По кнопке с уведомлением",
								},
								{
									icon: "📅",
									title: "Планирование",
									desc: "Повторяющиеся звонки",
								},
								{
									icon: "✋",
									title: "Поднять руку",
									desc: "Сигнал без прерывания",
								},
								{
									icon: "💬",
									title: "Чат звонка",
									desc: "Исчезает после звонка",
								},
							].map((f) => (
								<div
									key={f.title}
									className="flex gap-3 rounded-xl border p-3 hover:bg-raised/50 transition-colors"
								>
									<span className="text-xl">{f.icon}</span>
									<div>
										<p className="font-medium text-sm">{f.title}</p>
										<p className="text-xs text-t2">{f.desc}</p>
									</div>
								</div>
							))}
						</div>
					</section>
				</TabsContent>
			</Tabs>
		</div>
	);
}

/** Обёртка для изоляции рантайм-ошибок раздела «Звонки» (FE-H4). */
function CallsPageWithBoundary() {
	return (
		<RouteErrorBoundary module="Звонки">
			<CallsPage />
		</RouteErrorBoundary>
	);
}

export const Route = createFileRoute("/_authenticated/calls")({
	component: CallsPageWithBoundary,
});
