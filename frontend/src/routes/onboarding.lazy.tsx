import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import { completeOnboarding } from "@/api/auth";
import { FormField } from "@/components/ui/form-field";
import { useAuthStore } from "@/stores/auth-store";

/**
 * Страница онбординга — пошаговый мастер для новых пользователей.
 * Три шага: профиль (имя, должность, отдел, аватар),
 * обзор платформы (5 модулей с иконками),
 * настройки (тема, язык, создание/присоединение к проекту).
 */

/** Иконки SVG для каждого модуля платформы. */
const MODULE_ICONS: Record<string, React.ReactNode> = {
	/** Иконка трекера задач (Kanban-доска). */
	tracker: (
		<svg
			className="w-6 h-6"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			strokeWidth={1.8}
		>
			<rect x="4" y="4" width="6" height="16" rx="1" />
			<rect x="14" y="4" width="6" height="10" rx="1" />
		</svg>
	),
	/** Иконка мессенджера (чат). */
	messenger: (
		<svg
			className="w-6 h-6"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			strokeWidth={1.8}
		>
			<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
		</svg>
	),
	/** Иконка TMS (тестирование). */
	tms: (
		<svg
			className="w-6 h-6"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			strokeWidth={1.8}
		>
			<path d="M9 11l3 3L22 4" />
			<path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11" />
		</svg>
	),
	/** Иконка звонков (телефон). */
	calls: (
		<svg
			className="w-6 h-6"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			strokeWidth={1.8}
		>
			<path d="M22 16.92v3a2 2 0 0 1-2.18 2A19.79 19.79 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z" />
		</svg>
	),
	/** Иконка дашбордов (графики). */
	dashboards: (
		<svg
			className="w-6 h-6"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			strokeWidth={1.8}
		>
			<rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
			<line x1="16" y1="2" x2="16" y2="6" />
			<line x1="8" y1="2" x2="8" y2="6" />
			<line x1="3" y1="10" x2="21" y2="10" />
		</svg>
	),
	/** Иконка интеграций. */
	integrations: (
		<svg
			className="w-6 h-6"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			strokeWidth={1.8}
		>
			<path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
			<path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
		</svg>
	),
};

/** Тип описания модуля платформы. */
interface ModuleInfo {
	key: string;
	title: string;
	desc: string;
	iconBg: string;
	iconColor: string;
}

/** Описания модулей платформы для онбординга. */
const MODULES: ModuleInfo[] = [
	{
		key: "tracker",
		title: "Трекер",
		desc: "Kanban-доски, спринты, бэклог. Полный agile-цикл.",
		iconBg: "rgba(91,156,245,.08)",
		iconColor: "var(--blue)",
	},
	{
		key: "messenger",
		title: "Мессенджер",
		desc: "Каналы, ЛС, реакции, файлы, опросы.",
		iconBg: "rgba(52,211,153,.08)",
		iconColor: "var(--green)",
	},
	{
		key: "calls",
		title: "Звонки",
		desc: "Видеозвонки до 50 чел, шейринг, запись.",
		iconBg: "rgba(167,139,250,.08)",
		iconColor: "var(--purple)",
	},
	{
		key: "tms",
		title: "TMS",
		desc: "Тест-кейсы, прогоны, отчеты, автотесты.",
		iconBg: "rgba(251,191,36,.06)",
		iconColor: "var(--amber)",
	},
	{
		key: "dashboards",
		title: "Календарь",
		desc: "Планирование встреч и событий команды.",
		iconBg: "rgba(251,146,60,.06)",
		iconColor: "var(--orange)",
	},
	{
		key: "integrations",
		title: "Интеграции",
		desc: "GitHub, GitLab, CI/CD в одном месте.",
		iconBg: "rgba(248,113,113,.06)",
		iconColor: "var(--red)",
	},
];

/** Варианты темы интерфейса. */
type ThemeOption = "light" | "dark" | "system";

/** Варианты языка интерфейса. */
type LangOption = "ru" | "en";

/** Step labels for progress indicator */
const STEP_LABELS = ["Профиль", "Обзор", "Настройки"];

/** Карточка модуля для тура онбординга. */
function ModuleCard({ module, index }: { module: ModuleInfo; index: number }) {
	return (
		<div
			key={module.key}
			className="onb-tour-card bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-6 text-center transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-250 hover:-translate-y-1 hover:shadow-[0_12px_40px_rgba(0,0,0,0.4)] hover:border-b2 animate-[eu_400ms_var(--spring)_both]"
			style={{ animationDelay: `${index * 60}ms` }}
		>
			<div
				className="w-[52px] h-[52px] rounded-[var(--r3)] flex items-center justify-center mx-auto mb-[14px]"
				style={{ background: module.iconBg, color: module.iconColor }}
			>
				{MODULE_ICONS[module.key]}
			</div>
			<h4 className="text-[14px] font-bold mb-[5px] tracking-[-0.01em]">
				{module.title}
			</h4>
			<p className="text-[12px] text-t2 leading-[1.5]">
				{module.desc}
			</p>
		</div>
	);
}

function OnboardingPage() {
	const navigate = useNavigate();
	const { user, setUser } = useAuthStore();
	const [step, setStep] = useState(0);
	const [loading, setLoading] = useState(false);

	/** Данные профиля (шаг 1). */
	const [profile, setProfile] = useState({
		name: user?.name ?? "",
		position: user?.position ?? "",
		department: user?.department ?? "",
	});

	/** Выбранная тема (шаг 3). */
	const [theme, setTheme] = useState<ThemeOption>("dark");

	/** Выбранный язык (шаг 3). */
	const [lang] = useState<LangOption>("ru");

	/** Selected accent color index */
	const [accent, setAccent] = useState(0);

	const steps = [
		{ title: "Настройте профиль", description: "Расскажите о себе, так коллеги будут узнавать вас" },
		{
			title: "Знакомство с Hub",
			description: "Вот что вы сможете делать на платформе",
		},
		{ title: "Выберите оформление", description: "Можно изменить позже в настройках" },
	];

	const currentStep = steps[step];

	/** Завершить онбординг — отправить профиль и применить настройки. */
	async function handleComplete() {
		setLoading(true);
		try {
			// Применяем тему
			if (theme === "dark") {
				document.documentElement.classList.add("dark");
			} else if (theme === "light") {
				document.documentElement.classList.remove("dark");
			}

			const updated = await completeOnboarding({
				profile: { ...profile, language: lang },
			});
			setUser(updated);
			toast.success("Добро пожаловать в Hub!");
			void navigate({ to: "/projects" });
		} catch (err: unknown) {
			const message =
				err && typeof err === "object" && "message" in err
					? (err as { message: string }).message
					: "Не удалось завершить онбординг";
			toast.error(message);
		} finally {
			setLoading(false);
		}
	}

	const accentColors = [
		{ color: "var(--blue)", cssColor: "var(--blue)" },
		{ color: "var(--purple)", cssColor: "var(--purple)" },
		{ color: "var(--green)", cssColor: "var(--green)" },
		{ color: "var(--orange)", cssColor: "var(--orange)" },
		{ color: "var(--red)", cssColor: "var(--red)" },
	];

	return (
		<div className="min-h-screen flex items-center justify-center bg-void">
			<div className="w-full max-w-[680px] px-10 py-10 relative z-[1]">
				{/* Progress indicator */}
				<div className="flex items-center justify-center mb-11 gap-0">
					{STEP_LABELS.map((label, i) => (
						<div key={label} className="flex items-center">
							<div className="flex items-center gap-3">
								<div
									className={`w-10 h-10 rounded-full flex items-center justify-center text-[15px] font-extrabold border-2 transition-[box-shadow] duration-350 ${
										i < step
											? "border-green bg-green text-white shadow-[0_0_12px_var(--gG)]"
											: i === step
												? "border-blue text-blue bg-blue/[0.08] shadow-[0_0_16px_var(--gB)]"
												: "border-b2 text-t3"
									}`}
								>
									{i < step ? (
										<svg
											className="w-5 h-5"
											fill="none"
											viewBox="0 0 24 24"
											stroke="currentColor"
											strokeWidth={3}
										>
											<path
												strokeLinecap="round"
												strokeLinejoin="round"
												d="M5 13l4 4L19 7"
											/>
										</svg>
									) : (
										i + 1
									)}
								</div>
								<span
									className={`text-[13px] font-semibold transition-colors duration-300 ${
										i === step ? "text-t1" : "text-t4"
									}`}
								>
									{label}
								</span>
							</div>
							{i < STEP_LABELS.length - 1 && (
								<div
									className={`w-[60px] h-[2px] mx-[10px] transition-[background] duration-300 rounded-[1px] ${
										i < step
											? "bg-green shadow-[0_0_6px_var(--gG)]"
											: "bg-b2"
									}`}
								/>
							)}
						</div>
					))}
				</div>

				{/* Step 1: Profile form */}
				{step === 0 && (
					<div className="animate-[eu_400ms_var(--spring)]">
						<div className="text-[26px] font-black tracking-[-0.04em] text-center mb-2">
							{currentStep?.title}
						</div>
						<div className="text-[14px] text-t2 text-center mb-9">
							{currentStep?.description}
						</div>

						{/* Avatar upload */}
						<div className="flex flex-col items-center mb-[30px]">
							<div className="w-24 h-24 rounded-full bg-input border-2 border-dashed border-b2 flex items-center justify-center cursor-pointer transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-250 mb-2 hover:border-blue hover:bg-blue/[0.04] hover:scale-105">
								<svg
									className="w-8 h-8 text-t3"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									strokeWidth="1.5"
								>
									<path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z" />
									<circle cx="12" cy="13" r="4" />
								</svg>
							</div>
							<span className="text-[12px] text-t3">Нажмите чтобы загрузить</span>
						</div>

						{/* Form grid */}
						<div className="grid grid-cols-1 md:grid-cols-2 gap-[14px]">
							<FormField label="Имя">
								<input
									id="onb-name"
									placeholder="Алексей"
									value={profile.name}
									onChange={(e) =>
										setProfile((p) => ({ ...p, name: e.target.value }))
									}
									autoFocus
									className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-colors duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
							</FormField>
							<FormField label="Фамилия">
								<input
									placeholder="Иванов"
									className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-colors duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
							</FormField>
							<FormField label="Должность">
								<input
									id="onb-position"
									placeholder="Senior Developer"
									value={profile.position}
									onChange={(e) =>
										setProfile((p) => ({ ...p, position: e.target.value }))
									}
									className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-colors duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
							</FormField>
							<div>
								<label className="block text-[12px] font-semibold text-t2 mb-[5px]">
									Отдел
								</label>
								<input
									id="onb-department"
									placeholder="Backend Engineering"
									value={profile.department}
									onChange={(e) =>
										setProfile((p) => ({ ...p, department: e.target.value }))
									}
									className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-colors duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
							</div>
							<div>
								<label className="block text-[12px] font-semibold text-t2 mb-[5px]">
									Часовой пояс
								</label>
								<select className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 appearance-none outline-none">
									<option>Europe/Moscow (UTC+3)</option>
									<option>Europe/Warsaw (UTC+1)</option>
								</select>
							</div>
							<div>
								<label className="block text-[12px] font-semibold text-t2 mb-[5px]">
									Язык
								</label>
								<select className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 appearance-none outline-none">
									<option>Русский</option>
									<option>English</option>
								</select>
							</div>
						</div>

						{/* Navigation */}
						<div className="flex justify-between items-center mt-10 pt-[22px] border-t border-b1">
							<div />
							<button
								type="button"
								onClick={() => setStep((s) => s + 1)}
								className="onb-btn-next px-7 py-[10px] rounded-[var(--r2)] text-[14px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_20px_var(--gB)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 relative overflow-hidden hover:shadow-[0_0_32px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96]"
							>
								Далее &rarr;
							</button>
						</div>
					</div>
				)}

				{/* Step 2: Feature tour */}
				{step === 1 && (
					<div className="animate-[eu_400ms_var(--spring)]">
						<div className="text-[26px] font-black tracking-[-0.04em] text-center mb-2">
							{currentStep?.title}
						</div>
						<div className="text-[14px] text-t2 text-center mb-9">
							{currentStep?.description}
						</div>

						<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-[14px]">
							{MODULES.map((module, idx) => (
								<ModuleCard key={module.key} module={module} index={idx} />
							))}
						</div>

						{/* Navigation */}
						<div className="flex justify-between items-center mt-10 pt-[22px] border-t border-b1">
							<button
								type="button"
								onClick={() => setStep((s) => s - 1)}
								className="px-5 py-[10px] rounded-[var(--r2)] text-[14px] font-semibold text-t2 border border-b2 transition-colors duration-200 hover:bg-[var(--hover)] hover:text-t1"
							>
								&larr; Назад
							</button>
							<button
								type="button"
								onClick={() => setStep((s) => s + 1)}
								className="onb-btn-next px-7 py-[10px] rounded-[var(--r2)] text-[14px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_20px_var(--gB)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 relative overflow-hidden hover:shadow-[0_0_32px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96]"
							>
								Далее &rarr;
							</button>
						</div>
					</div>
				)}

				{/* Step 3: Theme + Accent */}
				{step === 2 && (
					<div className="animate-[eu_400ms_var(--spring)]">
						<div className="text-[26px] font-black tracking-[-0.04em] text-center mb-2">
							{currentStep?.title}
						</div>
						<div className="text-[14px] text-t2 text-center mb-9">
							{currentStep?.description}
						</div>

						{/* Theme label */}
						<h3 className="text-[14px] font-bold text-center mb-[14px]">Тема</h3>

						{/* Theme cards */}
						<div className="flex gap-[14px] mb-[30px]">
							{(
								[
									{ value: "light" as ThemeOption, label: "Светлая", preview: "bg-white border border-[#ddd]" },
									{ value: "dark" as ThemeOption, label: "Тёмная", preview: "bg-void border border-b1" },
									{ value: "system" as ThemeOption, label: "Системная", preview: "border border-[#ddd]" },
								] as const
							).map((opt) => (
								<button
									key={opt.value}
									type="button"
									className={`flex-1 p-[18px] bg-card border-2 rounded-[var(--r3)] text-center cursor-pointer transition-colors duration-250 hover:border-b2 ${
										theme === opt.value
											? "border-blue shadow-[0_0_0_3px_rgba(91,156,245,0.1),_0_0_20px_rgba(91,156,245,0.06)]"
											: "border-b1"
									}`}
									onClick={() => setTheme(opt.value)}
								>
									<div
										className={`w-full h-[52px] rounded-[var(--r2)] mb-[10px] ${opt.preview}`}
										style={
											opt.value === "system"
												? { background: "linear-gradient(135deg, #fff 50%, var(--void) 50%)" }
												: undefined
										}
									/>
									<span className="text-[13px] font-semibold">{opt.label}</span>
								</button>
							))}
						</div>

						{/* Accent label */}
						<h3 className="text-[14px] font-bold text-center mb-[14px]">Акцентный цвет</h3>

						{/* Accent dots */}
						<div className="flex gap-4 justify-center mb-[30px]">
							{accentColors.map((ac, idx) => (
								<div
									key={ac.color}
									className={`w-11 h-11 rounded-full cursor-pointer transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 border-[3px] hover:scale-[1.18] ${
										accent === idx
											? "border-white/40 shadow-[0_0_0_3px_currentColor,_0_0_20px_currentColor]"
											: "border-transparent"
									}`}
									style={{ background: ac.color, color: ac.cssColor }}
									onClick={() => setAccent(idx)}
								/>
							))}
						</div>

						{/* Navigation */}
						<div className="flex justify-between items-center mt-10 pt-[22px] border-t border-b1">
							<button
								type="button"
								onClick={() => setStep((s) => s - 1)}
								className="px-5 py-[10px] rounded-[var(--r2)] text-[14px] font-semibold text-t2 border border-b2 transition-colors duration-200 hover:bg-[var(--hover)] hover:text-t1"
							>
								&larr; Назад
							</button>
							<button
								type="button"
								onClick={handleComplete}
								disabled={loading}
								className="onb-btn-go px-8 py-3 rounded-[var(--r2)] text-[15px] font-extrabold bg-gradient-to-br from-[var(--greenS)] to-[#047857] text-white shadow-[0_0_24px_var(--gG)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 relative overflow-hidden hover:shadow-[0_0_36px_rgba(52,211,153,0.3)] hover:-translate-y-px active:scale-[0.96]"
							>
								{loading ? "Сохранение..." : "Начать работу →"}
							</button>
						</div>
					</div>
				)}
			</div>
		</div>
	);
}

export const Route = createFileRoute("/onboarding")({
	component: OnboardingPage,
});
