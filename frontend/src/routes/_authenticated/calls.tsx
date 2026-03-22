import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CallCalendar } from "@/modules/calls/call-calendar";

/**
 * Страница "Созвоны" — календарь и информация о звонках.
 * Звонки начинаются из каналов мессенджера (кнопка в шапке чата).
 */
function CallsPage() {
	const { t } = useTranslation("calls");

	return (
		<div className="mx-auto max-w-5xl p-6 space-y-8 animate-fade-in">
			<div className="flex items-center gap-3">
				<div className="flex h-10 w-10 items-center justify-center rounded-xl gradient-calls text-white shadow-sm">
					<svg
						className="h-5 w-5"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2}
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
						/>
					</svg>
				</div>
				<div>
					<h1 className="text-2xl font-bold tracking-tight">
						{t("title", { defaultValue: "Созвоны" })}
					</h1>
					<p className="text-sm text-muted-foreground">
						{t("subtitle", { defaultValue: "Календарь звонков и встреч" })}
					</p>
				</div>
			</div>

			<Tabs defaultValue="calendar">
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
											<p className="text-xs text-muted-foreground mt-0.5">
												{s.desc}
											</p>
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
									className="flex gap-3 rounded-xl border p-3 hover:bg-muted/50 transition-colors"
								>
									<span className="text-xl">{f.icon}</span>
									<div>
										<p className="font-medium text-sm">{f.title}</p>
										<p className="text-xs text-muted-foreground">{f.desc}</p>
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

export const Route = createFileRoute("/_authenticated/calls")({
	component: CallsPage,
});
