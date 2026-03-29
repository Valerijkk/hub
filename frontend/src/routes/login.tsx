import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { DevLoginPanel } from "@/components/auth/dev-login-panel";
import { TelegramButton } from "@/components/auth/telegram-button";
import { useAuthStore } from "@/stores/auth-store";

/* --- Floating feature cards data --- */

interface FeatureCard {
	label: string;
	icon: string;
	desc: string;
	iconBg: string;
	iconColor: string;
	delay: string;
	duration: string;
}

const FLOATING_FEATURES: FeatureCard[] = [
	{
		label: "Tracker",
		icon: "M4 5a1 1 0 011-1h6v16H5a1 1 0 01-1-1V5zM14 4h6a1 1 0 011 1v9a1 1 0 01-1 1h-6V4z",
		desc: "Kanban-доски, спринты, бэклог — полный agile-цикл",
		iconBg: "rgba(91,156,245,.1)",
		iconColor: "var(--blue)",
		delay: "0s",
		duration: "7s",
	},
	{
		label: "Messenger",
		icon: "M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z",
		desc: "Каналы, личные сообщения, реакции, файлы и медиа",
		iconBg: "rgba(52,211,153,.1)",
		iconColor: "var(--green)",
		delay: "0.5s",
		duration: "8s",
	},
	{
		label: "Calls",
		icon: "M22 16.92v3a2 2 0 01-2.18 2A19.79 19.79 0 014.11 2h3a2 2 0 012 1.72 12.84 12.84 0 00.7 2.81 2 2 0 01-.45 2.11L8.09 9.91a16 16 0 006 6l1.27-1.27a2 2 0 012.11-.45 12.84 12.84 0 002.81.7A2 2 0 0122 16.92z",
		desc: "Видеозвонки до 50 человек с демонстрацией экрана",
		iconBg: "rgba(167,139,250,.1)",
		iconColor: "var(--purple)",
		delay: "1s",
		duration: "9s",
	},
	{
		label: "TMS",
		icon: "M9 11l3 3L22 4M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11",
		desc: "Тест-кейсы, прогоны, отчёты — управление качеством",
		iconBg: "rgba(251,191,36,.08)",
		iconColor: "var(--amber)",
		delay: "1.5s",
		duration: "7.5s",
	},
];

/**
 * Страница входа — split-layout с animated mesh gradient и формой.
 *
 * Левая часть (55%): mesh gradient, grid lines, 2x2 feature cards, tagline.
 * Правая часть (45%): clean form panel with logo, email/code form, Telegram.
 */
function LoginPage() {
	const navigate = useNavigate();


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
			{/* Левая панель — animated mesh gradient + feature cards */}
			<div className="relative hidden lg:flex lg:w-[55%] flex-col items-center justify-center overflow-hidden">
				{/* Mesh gradient background */}
				<div className="absolute inset-0 login-mesh" />
				{/* Grid lines overlay */}
				<div className="absolute inset-0 login-grid-lines" />

				{/* 2x2 Feature cards grid */}
				<div className="relative z-[2] grid grid-cols-2 gap-5 max-w-[440px] p-10">
					{FLOATING_FEATURES.map((f) => (
						<div
							key={f.label}
							className="rounded-[var(--r4)] border border-white/[0.06] bg-white/[0.03] backdrop-blur-[20px] p-6 cursor-default transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-300 relative overflow-hidden hover:bg-white/[0.05] hover:border-white/[0.1] hover:-translate-y-1.5"
							style={{
								animation: `cardFloat ${f.duration} ease-in-out ${f.delay} infinite`,
							}}
						>
							{/* Top highlight line */}
							<div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.06] to-transparent" />
							{/* Icon */}
							<div
								className="w-[42px] h-[42px] rounded-[var(--r3)] flex items-center justify-center mb-3.5"
								style={{ background: f.iconBg }}
							>
								<svg
									className="w-[22px] h-[22px]"
									fill="none"
									viewBox="0 0 24 24"
									stroke={f.iconColor}
									strokeWidth={1.8}
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d={f.icon}
									/>
								</svg>
							</div>
							<h4 className="text-[15px] font-bold tracking-[-0.01em] mb-1.5">
								{f.label}
							</h4>
							<p className="text-xs text-[var(--t2)] leading-relaxed">
								{f.desc}
							</p>
						</div>
					))}
				</div>

				{/* Tagline — bottom-left */}
				<div className="absolute bottom-12 left-12 z-[2]">
					<h1 className="text-[30px] font-black tracking-[-0.04em] leading-tight bg-gradient-to-br from-[var(--blue)] via-[var(--purple)] to-pink-500 bg-clip-text text-transparent mb-2">
						Единая платформа
					</h1>
					<p className="text-[15px] text-[var(--t2)] max-w-[360px] leading-relaxed">
						Проекты, общение, звонки и тестирование — в одном месте
						для всей команды
					</p>
				</div>
			</div>

			{/* Правая панель — форма входа */}
			<div className="flex flex-1 items-center justify-center bg-[var(--base)] relative">
				{/* Left border separator */}
				<div className="absolute top-0 left-0 bottom-0 w-px bg-gradient-to-b from-transparent via-[rgba(91,156,245,.06)] to-transparent" />

				<div className="w-full max-w-[380px] px-4 sm:px-10 login-form-animate">
					{/* Logo area */}
					<div className="text-center mb-9">
						<div className="text-4xl font-black tracking-[-0.05em] bg-gradient-to-br from-[var(--blue)] via-[var(--purple)] to-pink-500 bg-clip-text text-transparent">
							Hub
						</div>
						<div className="text-sm text-[var(--t2)] mt-2">
							Единая платформа для команд
						</div>
					</div>

					{/* Telegram авторизация — основной способ входа */}
					<TelegramButton />

					{/* Описание */}
					<div className="text-center mt-7 text-sm text-[var(--t2)]">
						Войдите через Telegram — быстро и безопасно.
						<br />
						<span className="text-[var(--t3)] text-xs">Аккаунт создаётся автоматически при первом входе.</span>
					</div>

					{/* Dev login panel */}
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
