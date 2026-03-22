import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { getMe } from "@/api/auth";
import { DevLoginPanel } from "@/components/auth/dev-login-panel";

import { MagicLinkForm } from "@/components/auth/magic-link-form";
import { TelegramButton } from "@/components/auth/telegram-button";
import { useAuthStore } from "@/stores/auth-store";

/* --- Floating feature cards data --- */

interface FeatureCard {
	label: string;
	icon: string;
	x: string;
	y: string;
	delay: string;
	duration: string;
}

const FLOATING_FEATURES: FeatureCard[] = [
	{
		label: "Трекер задач",
		icon: "M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2",
		x: "8%",
		y: "18%",
		delay: "0s",
		duration: "7s",
	},
	{
		label: "Мессенджер",
		icon: "M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z",
		x: "72%",
		y: "12%",
		delay: "1.5s",
		duration: "8s",
	},
	{
		label: "Видеозвонки",
		icon: "M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z",
		x: "65%",
		y: "68%",
		delay: "3s",
		duration: "9s",
	},
	{
		label: "Тест-менеджмент",
		icon: "M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z",
		x: "12%",
		y: "72%",
		delay: "4.5s",
		duration: "7.5s",
	},
];

/**
 * Страница входа — split-layout с hero-областью и формой.
 *
 * Левая часть: градиентный фон с брендингом Hub и плавающими карточками.
 * Правая часть: форма входа на чистом фоне с карточным контейнером.
 */
function LoginPage() {
	const navigate = useNavigate();
	const setUser = useAuthStore((s) => s.setUser);

	async function handleLoginSuccess() {
		try {
			const user = await getMe();
			setUser(user);
			if (!user.onboarding_completed) {
				void navigate({ to: "/onboarding" });
			} else {
				void navigate({ to: "/projects" });
			}
		} catch {
			void navigate({ to: "/projects" });
		}
	}

	function handleDevLoginSuccess() {
		const user = useAuthStore.getState().user;
		if (user && !user.onboarding_completed) {
			void navigate({ to: "/onboarding" });
		} else {
			void navigate({ to: "/projects" });
		}
	}

	return (
		<div className="flex min-h-screen">
			{/* Левая hero-область — скрыта на мобильных */}
			<div className="relative hidden lg:flex lg:w-[55%] flex-col justify-between overflow-hidden">
				{/* Анимированный mesh-фон — замедленный */}
				<div
					className="absolute inset-0 login-mesh-bg"
					style={{
						background:
							"linear-gradient(-45deg, #0C1222, #1E293B, #1E1B4B, #0C1222, #164E63, #0C1222)",
						backgroundSize: "400% 400%",
					}}
				/>

				{/* Сетка-паттерн */}
				<div
					className="absolute inset-0 opacity-[0.04]"
					style={{
						backgroundImage:
							"linear-gradient(rgba(255,255,255,0.5) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.5) 1px, transparent 1px)",
						backgroundSize: "48px 48px",
					}}
				/>

				{/* Тонкие градиентные пятна */}
				<div className="absolute top-1/4 -left-20 h-[500px] w-[500px] rounded-full bg-blue-500/15 blur-[120px]" />
				<div className="absolute bottom-1/4 right-0 h-[400px] w-[400px] rounded-full bg-purple-500/15 blur-[120px]" />
				<div className="absolute bottom-0 left-1/3 h-[300px] w-[300px] rounded-full bg-cyan-500/10 blur-[100px]" />

				{/* Плавающие карточки фич */}
				{FLOATING_FEATURES.map((f) => (
					<div
						key={f.label}
						className="login-float-card absolute z-[5]"
						style={{
							left: f.x,
							top: f.y,
							animationDelay: f.delay,
							animationDuration: f.duration,
						}}
					>
						<div className="flex items-center gap-2 rounded-xl border border-white/[0.08] bg-white/[0.06] px-3.5 py-2 backdrop-blur-md shadow-lg shadow-black/10">
							<svg
								className="h-4 w-4 text-blue-400/80"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								strokeWidth={1.5}
								aria-hidden="true"
							>
								<path
									strokeLinecap="round"
									strokeLinejoin="round"
									d={f.icon}
								/>
							</svg>
							<span className="text-xs font-medium text-white/70 whitespace-nowrap">
								{f.label}
							</span>
						</div>
					</div>
				))}

				{/* Контент hero */}
				<div className="relative z-10 flex flex-1 flex-col justify-center px-12 xl:px-16">
					<div className="max-w-lg">
						<div className="mb-10 flex items-center gap-4">
							<div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-500 via-purple-500 to-pink-500 shadow-xl shadow-blue-500/30">
								<span className="text-2xl font-bold text-white">
									H
								</span>
							</div>
							<h1 className="text-4xl font-bold text-white tracking-tight">
								Hub
							</h1>
						</div>
						<p className="text-4xl xl:text-5xl font-bold text-white leading-[1.15] mb-5 tracking-tight">
							Единая платформа для управления проектами
						</p>
						<p className="text-base text-slate-400/90 leading-relaxed max-w-md">
							Трекер задач, мессенджер, тест-менеджмент и
							видеозвонки — все инструменты в одном месте.
						</p>
					</div>
				</div>

				{/* Декоративные точки внизу */}
				<div className="relative z-10 px-12 pb-8">
					<div className="flex items-center gap-1.5">
						<div className="h-1.5 w-1.5 rounded-full bg-blue-400/60" />
						<div className="h-1.5 w-1.5 rounded-full bg-purple-400/60" />
						<div className="h-1.5 w-1.5 rounded-full bg-pink-400/60" />
						<div className="h-1.5 w-1.5 rounded-full bg-cyan-400/60" />
					</div>
				</div>
			</div>

			{/* Правая часть — форма входа */}
			<div className="flex flex-1 flex-col items-center justify-center px-6 py-12 lg:px-16 bg-background max-lg:login-mobile-bg">
				<div className="w-full max-w-[420px] login-form-enter">
					{/* Логотип — только на мобильных */}
					<div className="mb-8 flex items-center justify-center gap-3 lg:hidden">
						<div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-500 via-purple-500 to-pink-500 shadow-lg shadow-blue-500/25">
							<span className="text-xl font-bold text-white">
								H
							</span>
						</div>
						<h1 className="text-3xl font-bold tracking-tight gradient-text">
							Hub
						</h1>
					</div>

					{/* Карточный контейнер формы */}
					<div className="rounded-2xl border border-border/60 bg-card p-8 shadow-lg shadow-black/[0.03] dark:shadow-black/20">
						<div className="mb-8 text-center lg:text-left">
							<h2 className="text-3xl font-bold tracking-tight text-foreground">
								Войти в аккаунт
							</h2>
							<p className="mt-2.5 text-sm text-muted-foreground leading-relaxed">
								Введите email для получения кода доступа
							</p>
						</div>

						{/* Форма */}
						<div className="space-y-6">
							<MagicLinkForm onSuccess={handleLoginSuccess} />

							{/* Разделитель */}
							<div className="relative">
								<div className="absolute inset-0 flex items-center">
									<div className="w-full border-t border-border" />
								</div>
								<div className="relative flex justify-center text-xs uppercase">
									<span className="px-4 text-muted-foreground/70 bg-card font-medium tracking-wider">
										или
									</span>
								</div>
							</div>

							<TelegramButton />
						</div>
					</div>

					{/* Dev login panel — вынесен за карточку */}
					<div className="mt-6">
						<DevLoginPanel onSuccess={handleDevLoginSuccess} />
					</div>
				</div>
			</div>
		</div>
	);
}

export const Route = createFileRoute("/login")({
	component: LoginPage,
});
