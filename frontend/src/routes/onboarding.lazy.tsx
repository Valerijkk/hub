import { zodResolver } from "@hookform/resolvers/zod";
import { createLazyFileRoute, useNavigate } from "@tanstack/react-router";
import { useMemo, useRef, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { completeOnboarding } from "@/api/auth";
import { uploadFile } from "@/api/files";
import { createProject, joinByInvite } from "@/api/projects";
import {
	ACCENT_COLORS,
	type AccentColor,
	useAccentColor,
} from "@/components/theme-provider";
import { FormField } from "@/components/ui/form-field";
import { toastApiError } from "@/lib/format-api-error";
import {
	buildProfileDisplayName,
	type OnboardingProfileValues,
	onboardingProfileSchema,
} from "@/lib/onboarding/profile-schema";
import { validateOnboardingProjectStep } from "@/lib/onboarding/project-schema";
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
			aria-hidden="true"
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
			aria-hidden="true"
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
			aria-hidden="true"
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
			aria-hidden="true"
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
			aria-hidden="true"
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
			aria-hidden="true"
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

/** Описания модулей — стили фиксированы, тексты из i18n. */
const MODULE_KEYS = [
	"tracker",
	"messenger",
	"calls",
	"tms",
	"dashboards",
	"integrations",
] as const;

const MODULE_STYLES: Record<
	(typeof MODULE_KEYS)[number],
	{ iconBg: string; iconColor: string }
> = {
	tracker: { iconBg: "rgba(91,156,245,.08)", iconColor: "var(--blue)" },
	messenger: { iconBg: "rgba(52,211,153,.08)", iconColor: "var(--green)" },
	calls: { iconBg: "rgba(167,139,250,.08)", iconColor: "var(--purple)" },
	tms: { iconBg: "rgba(251,191,36,.06)", iconColor: "var(--amber)" },
	dashboards: { iconBg: "rgba(251,146,60,.06)", iconColor: "var(--orange)" },
	integrations: { iconBg: "rgba(248,113,113,.06)", iconColor: "var(--red)" },
};

/** Варианты темы интерфейса. */
type ThemeOption = "light" | "dark" | "system";

/** Варианты языка интерфейса. */
type LangOption = "ru" | "en";

/** Подписи шагов индикатора прогресса онбординга. */
const STEP_LABEL_KEYS = [
	"common:onboarding.stepProfile",
	"common:onboarding.stepTour",
	"common:onboarding.stepTheme",
	"common:onboarding.stepProject",
] as const;

type ProjectMode = "create" | "join";

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
			<p className="text-[12px] text-t2 leading-[1.5]">{module.desc}</p>
		</div>
	);
}

function OnboardingPage() {
	const navigate = useNavigate();
	const { t } = useTranslation(["common", "onboarding"]);
	const { user, setUser } = useAuthStore();
	const { accent, setAccent } = useAccentColor();
	const [step, setStep] = useState(0);
	const [loading, setLoading] = useState(false);
	const [avatarUrl, setAvatarUrl] = useState<string | undefined>(
		user?.avatar_url,
	);
	const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
	const [avatarUploading, setAvatarUploading] = useState(false);
	const fileInputRef = useRef<HTMLInputElement>(null);

	const profileForm = useForm<OnboardingProfileValues>({
		resolver: zodResolver(onboardingProfileSchema),
		mode: "onChange",
		defaultValues: {
			name: user?.name ?? "",
			surname: "",
			position: user?.position ?? "",
			department: user?.department ?? "",
			timezone: user?.timezone || "Europe/Moscow",
			language: (user?.language === "en" ? "en" : "ru") as LangOption,
		},
	});

	/** Выбранная тема (шаг 3). */
	const [theme, setTheme] = useState<ThemeOption>("dark");

	/** Шаг 4: проект */
	const [projectMode, setProjectMode] = useState<ProjectMode>("create");
	const [projectName, setProjectName] = useState("");
	const [projectKey, setProjectKey] = useState("");
	const [inviteCode, setInviteCode] = useState("");
	const [projectErrors, setProjectErrors] = useState<Record<string, string>>(
		{},
	);

	const steps = useMemo(
		() => [
			{
				title: t("onboarding:steps.profile.title"),
				description: t("onboarding:steps.profile.description"),
			},
			{
				title: t("onboarding:steps.tour.title"),
				description: t("onboarding:steps.tour.description"),
			},
			{
				title: t("onboarding:steps.theme.title"),
				description: t("onboarding:steps.theme.description"),
			},
			{
				title: t("common:onboarding.projectTitle"),
				description: t("common:onboarding.projectDesc"),
			},
		],
		[t],
	);

	const modules = useMemo(
		() =>
			MODULE_KEYS.map((key) => ({
				key,
				title: t(`onboarding:modules.${key}.title`),
				desc: t(`onboarding:modules.${key}.desc`),
				...MODULE_STYLES[key],
			})),
		[t],
	);

	const currentStep = steps[step];

	const isProjectStepValid =
		validateOnboardingProjectStep(projectMode, {
			projectName,
			projectKey,
			inviteCode,
		}) === null;

	const projectFieldError = (key: string) => {
		const msg = projectErrors[key];
		return msg ? t(msg as never) : null;
	};

	const profileFieldError = (key: keyof OnboardingProfileValues) => {
		const msg = profileForm.formState.errors[key]?.message;
		return msg ? t(msg as never) : null;
	};

	const advanceFromProfileStep = profileForm.handleSubmit(() => {
		setStep((s) => s + 1);
	});

	/** Завершить онбординг — отправить профиль и применить настройки. */
	async function handleComplete(skipProject = false) {
		if (!skipProject) {
			const errors = validateOnboardingProjectStep(projectMode, {
				projectName,
				projectKey,
				inviteCode,
			});
			if (errors) {
				setProjectErrors(errors);
				return;
			}
		}
		setProjectErrors({});
		setLoading(true);
		try {
			const profileValid = await profileForm.trigger();
			if (!profileValid) {
				setStep(0);
				return;
			}

			if (theme === "dark") {
				document.documentElement.classList.add("dark");
			} else if (theme === "light") {
				document.documentElement.classList.remove("dark");
			}

			if (!skipProject) {
				if (
					projectMode === "create" &&
					projectName.trim() &&
					projectKey.trim()
				) {
					await createProject({
						name: projectName.trim(),
						key: projectKey.trim().toUpperCase(),
					});
				} else if (projectMode === "join" && inviteCode.trim()) {
					await joinByInvite(inviteCode.trim());
				}
			}

			const profileValues = profileForm.getValues();
			const displayName = buildProfileDisplayName(profileValues);

			const updated = await completeOnboarding({
				profile: {
					name: displayName,
					position: profileValues.position?.trim() || undefined,
					department: profileValues.department?.trim() || undefined,
					timezone: profileValues.timezone,
					language: profileValues.language,
					avatar_url: avatarUrl,
				},
			});
			setUser(updated);
			toast.success(t("common:onboarding.welcome"));
			void navigate({ to: "/projects" });
		} catch (err: unknown) {
			toastApiError(err, t("common:onboarding.completeError"));
		} finally {
			setLoading(false);
		}
	}

	const accentPreviews: Record<AccentColor, string> = {
		red: "var(--red)",
		blue: "var(--blue)",
		green: "var(--green)",
		purple: "var(--purple)",
		orange: "var(--orange)",
	};

	async function handleAvatarUpload(file: File) {
		if (file.size > 2 * 1024 * 1024) {
			toast.error(t("common:validation.avatarMaxSize"));
			return;
		}
		if (!file.type.startsWith("image/")) {
			toast.error(t("common:validation.avatarImagesOnly"));
			return;
		}
		setAvatarUploading(true);
		try {
			const preview = URL.createObjectURL(file);
			setAvatarPreview(preview);
			const fileId = await uploadFile(file, "", "avatar", "me");
			setAvatarUrl(`/files/${fileId}/url`);
		} catch (err: unknown) {
			toastApiError(err, t("common:onboarding.completeError"));
			setAvatarPreview(null);
		} finally {
			setAvatarUploading(false);
		}
	}

	return (
		<div className="min-h-screen flex items-center justify-center bg-void">
			<div className="w-full max-w-[680px] px-10 py-10 relative z-[1]">
				{/* Индикатор прогресса */}
				<div className="flex items-center justify-center mb-11 gap-0">
					{STEP_LABEL_KEYS.map((labelKey, i) => (
						<div key={labelKey} className="flex items-center">
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
											aria-hidden="true"
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
									{t(labelKey)}
								</span>
							</div>
							{i < STEP_LABEL_KEYS.length - 1 && (
								<div
									className={`w-[60px] h-[2px] mx-[10px] transition-[background] duration-300 rounded-[1px] ${
										i < step ? "bg-green shadow-[0_0_6px_var(--gG)]" : "bg-b2"
									}`}
								/>
							)}
						</div>
					))}
				</div>

				{/* Шаг 1: форма профиля */}
				{step === 0 && (
					<div className="animate-[eu_400ms_var(--spring)]">
						<div className="text-[26px] font-black tracking-[-0.04em] text-center mb-2">
							{currentStep?.title}
						</div>
						<div className="text-[14px] text-t2 text-center mb-9">
							{currentStep?.description}
						</div>

						{/* Загрузка аватара */}
						<div className="flex flex-col items-center mb-[30px]">
							<input
								ref={fileInputRef}
								type="file"
								accept="image/*"
								className="hidden"
								aria-label={t("onboarding:profile.uploadAvatar")}
								onChange={(e) => {
									const file = e.target.files?.[0];
									if (file) void handleAvatarUpload(file);
									e.target.value = "";
								}}
							/>
							<button
								type="button"
								disabled={avatarUploading}
								onClick={() => fileInputRef.current?.click()}
								className="w-24 h-24 rounded-full bg-input border-2 border-dashed border-b2 flex items-center justify-center cursor-pointer transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-250 mb-2 hover:border-blue hover:bg-blue/[0.04] hover:scale-105 overflow-hidden disabled:opacity-50"
							>
								{avatarPreview || avatarUrl ? (
									<img
										src={avatarPreview ?? avatarUrl}
										alt=""
										className="h-full w-full object-cover"
									/>
								) : (
									<svg
										aria-hidden="true"
										className="w-8 h-8 text-t3"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										strokeWidth="1.5"
									>
										<path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z" />
										<circle cx="12" cy="13" r="4" />
									</svg>
								)}
							</button>
							<span className="text-[12px] text-t3">
								{avatarUploading ? "…" : t("onboarding:profile.uploadAvatar")}
							</span>
						</div>

						{/* Сетка полей формы */}
						<div className="grid grid-cols-1 md:grid-cols-2 gap-[14px]">
							<FormField
								label={t("onboarding:profile.name")}
								htmlFor="onb-name"
							>
								<input
									id="onb-name"
									placeholder={t("onboarding:profile.namePlaceholder")}
									{...profileForm.register("name")}
									aria-invalid={
										profileForm.formState.errors.name ? true : undefined
									}
									className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-colors duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
								{profileForm.formState.errors.name ? (
									<p className="text-sm text-red-500 mt-1" role="alert">
										{profileFieldError("name")}
									</p>
								) : null}
							</FormField>
							<FormField
								label={t("onboarding:profile.surname")}
								htmlFor="onb-surname"
							>
								<input
									id="onb-surname"
									placeholder={t("onboarding:profile.surnamePlaceholder")}
									{...profileForm.register("surname")}
									aria-invalid={
										profileForm.formState.errors.surname ? true : undefined
									}
									className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-colors duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
								{profileForm.formState.errors.surname ? (
									<p className="text-sm text-red-500 mt-1" role="alert">
										{profileFieldError("surname")}
									</p>
								) : null}
							</FormField>
							<FormField
								label={t("onboarding:profile.position")}
								htmlFor="onb-position"
							>
								<input
									id="onb-position"
									placeholder={t("onboarding:profile.positionPlaceholder")}
									{...profileForm.register("position")}
									aria-invalid={
										profileForm.formState.errors.position ? true : undefined
									}
									className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-colors duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
								{profileForm.formState.errors.position ? (
									<p className="text-sm text-red-500 mt-1" role="alert">
										{profileFieldError("position")}
									</p>
								) : null}
							</FormField>
							<div>
								<label
									htmlFor="onb-department"
									className="block text-[12px] font-semibold text-t2 mb-[5px]"
								>
									{t("onboarding:profile.department")}
								</label>
								<input
									id="onb-department"
									placeholder={t("onboarding:profile.departmentPlaceholder")}
									{...profileForm.register("department")}
									aria-invalid={
										profileForm.formState.errors.department ? true : undefined
									}
									className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 transition-colors duration-250 outline-none focus:border-[rgba(91,156,245,0.3)] focus:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] placeholder:text-t3"
								/>
								{profileForm.formState.errors.department ? (
									<p className="text-sm text-red-500 mt-1" role="alert">
										{profileFieldError("department")}
									</p>
								) : null}
							</div>
							<div>
								<label
									htmlFor="onb-timezone"
									className="block text-[12px] font-semibold text-t2 mb-[5px]"
								>
									{t("onboarding:profile.timezone")}
								</label>
								<select
									id="onb-timezone"
									{...profileForm.register("timezone")}
									className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 appearance-none outline-none"
								>
									<option value="Europe/Moscow">Europe/Moscow (UTC+3)</option>
									<option value="Europe/Warsaw">Europe/Warsaw (UTC+1)</option>
									<option value="UTC">UTC</option>
								</select>
							</div>
							<div>
								<label
									htmlFor="onb-language"
									className="block text-[12px] font-semibold text-t2 mb-[5px]"
								>
									{t("onboarding:profile.language")}
								</label>
								<select
									id="onb-language"
									{...profileForm.register("language")}
									className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 appearance-none outline-none"
								>
									<option value="ru">Русский</option>
									<option value="en">English</option>
								</select>
							</div>
						</div>

						{/* Навигация по шагам */}
						<div className="flex justify-between items-center mt-10 pt-[22px] border-t border-b1">
							<div />
							<button
								type="button"
								onClick={() => void advanceFromProfileStep()}
								disabled={!profileForm.formState.isValid}
								className="onb-btn-next px-7 py-[10px] rounded-[var(--r2)] text-[14px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_20px_var(--gB)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 relative overflow-hidden hover:shadow-[0_0_32px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96] disabled:opacity-50 disabled:cursor-not-allowed"
							>
								{t("btn.next")} &rarr;
							</button>
						</div>
					</div>
				)}

				{/* Шаг 2: тур по возможностям */}
				{step === 1 && (
					<div className="animate-[eu_400ms_var(--spring)]">
						<div className="text-[26px] font-black tracking-[-0.04em] text-center mb-2">
							{currentStep?.title}
						</div>
						<div className="text-[14px] text-t2 text-center mb-9">
							{currentStep?.description}
						</div>

						<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-[14px]">
							{modules.map((module, idx) => (
								<ModuleCard key={module.key} module={module} index={idx} />
							))}
						</div>

						{/* Навигация по шагам */}
						<div className="flex justify-between items-center mt-10 pt-[22px] border-t border-b1">
							<button
								type="button"
								onClick={() => setStep((s) => s - 1)}
								className="px-5 py-[10px] rounded-[var(--r2)] text-[14px] font-semibold text-t2 border border-b2 transition-colors duration-200 hover:bg-[var(--hover)] hover:text-t1"
							>
								&larr; {t("btn.back")}
							</button>
							<button
								type="button"
								onClick={() => setStep((s) => s + 1)}
								className="onb-btn-next px-7 py-[10px] rounded-[var(--r2)] text-[14px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_20px_var(--gB)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 relative overflow-hidden hover:shadow-[0_0_32px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96]"
							>
								{t("btn.next")} &rarr;
							</button>
						</div>
					</div>
				)}

				{/* Шаг 3: тема + accent */}
				{step === 2 && (
					<div className="animate-[eu_400ms_var(--spring)]">
						<div className="text-[26px] font-black tracking-[-0.04em] text-center mb-2">
							{currentStep?.title}
						</div>
						<div className="text-[14px] text-t2 text-center mb-9">
							{currentStep?.description}
						</div>

						{/* Подпись темы */}
						<h3 className="text-[14px] font-bold text-center mb-[14px]">
							{t("onboarding:theme.label")}
						</h3>

						{/* Карточки тем */}
						<div className="flex gap-[14px] mb-[30px]">
							{(
								[
									{
										value: "light" as ThemeOption,
										label: t("onboarding:theme.light"),
										preview: "bg-white border border-[#ddd]",
									},
									{
										value: "dark" as ThemeOption,
										label: t("onboarding:theme.dark"),
										preview: "bg-void border border-b1",
									},
									{
										value: "system" as ThemeOption,
										label: t("onboarding:theme.system"),
										preview: "border border-[#ddd]",
									},
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
												? {
														background:
															"linear-gradient(135deg, #fff 50%, var(--void) 50%)",
													}
												: undefined
										}
									/>
									<span className="text-[13px] font-semibold">{opt.label}</span>
								</button>
							))}
						</div>

						{/* Подпись accent */}
						<h3 className="text-[14px] font-bold text-center mb-[14px]">
							{t("onboarding:theme.accent")}
						</h3>

						{/* Точки accent-цветов */}
						<div className="flex gap-4 justify-center mb-[30px]">
							{ACCENT_COLORS.map((color) => (
								<button
									key={color}
									type="button"
									aria-label={`Выбрать акцентный цвет ${color}`}
									aria-pressed={accent === color}
									className={`w-11 h-11 rounded-full cursor-pointer transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 border-[3px] hover:scale-[1.18] ${
										accent === color
											? "border-white/40 shadow-[0_0_0_3px_currentColor,_0_0_20px_currentColor]"
											: "border-transparent"
									}`}
									style={{
										background: accentPreviews[color],
										color: accentPreviews[color],
									}}
									onClick={() => setAccent(color)}
								/>
							))}
						</div>

						{/* Навигация по шагам */}
						<div className="flex justify-between items-center mt-10 pt-[22px] border-t border-b1">
							<button
								type="button"
								onClick={() => setStep((s) => s - 1)}
								className="px-5 py-[10px] rounded-[var(--r2)] text-[14px] font-semibold text-t2 border border-b2 transition-colors duration-200 hover:bg-[var(--hover)] hover:text-t1"
							>
								&larr; {t("btn.back")}
							</button>
							<button
								type="button"
								onClick={() => setStep((s) => s + 1)}
								className="onb-btn-next px-7 py-[10px] rounded-[var(--r2)] text-[14px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_20px_var(--gB)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 relative overflow-hidden hover:shadow-[0_0_32px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96]"
							>
								{t("btn.next")} &rarr;
							</button>
						</div>
					</div>
				)}

				{/* Шаг 4: создать/вступить в проект — US-M08 */}
				{step === 3 && (
					<div className="animate-[eu_400ms_var(--spring)]">
						<div className="text-[26px] font-black tracking-[-0.04em] text-center mb-2">
							{currentStep?.title}
						</div>
						<div className="text-[14px] text-t2 text-center mb-9">
							{currentStep?.description}
						</div>

						<div className="flex gap-2 mb-6">
							<button
								type="button"
								onClick={() => {
									setProjectMode("create");
									setProjectErrors({});
								}}
								className={`flex-1 rounded-[var(--r2)] border px-4 py-2 text-sm font-semibold ${
									projectMode === "create"
										? "border-blue bg-blue/10 text-blue"
										: "border-b1 text-t2"
								}`}
							>
								{t("common:onboarding.createProject")}
							</button>
							<button
								type="button"
								onClick={() => {
									setProjectMode("join");
									setProjectErrors({});
								}}
								className={`flex-1 rounded-[var(--r2)] border px-4 py-2 text-sm font-semibold ${
									projectMode === "join"
										? "border-blue bg-blue/10 text-blue"
										: "border-b1 text-t2"
								}`}
							>
								{t("common:onboarding.joinProject")}
							</button>
						</div>

						{projectMode === "create" ? (
							<div className="grid gap-4">
								<FormField
									label={t("common:onboarding.projectName")}
									htmlFor="onb-project-name"
								>
									<input
										id="onb-project-name"
										value={projectName}
										onChange={(e) => {
											setProjectName(e.target.value);
											setProjectErrors((prev) => {
												const next = { ...prev };
												delete next.projectName;
												return next;
											});
										}}
										aria-invalid={projectErrors.projectName ? true : undefined}
										className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 outline-none focus:border-[rgba(91,156,245,0.3)]"
									/>
									{projectErrors.projectName ? (
										<p className="text-sm text-red-500 mt-1" role="alert">
											{projectFieldError("projectName")}
										</p>
									) : null}
								</FormField>
								<FormField
									label={t("common:onboarding.projectKey")}
									htmlFor="onb-project-key"
								>
									<input
										id="onb-project-key"
										value={projectKey}
										onChange={(e) => {
											setProjectKey(
												e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, ""),
											);
											setProjectErrors((prev) => {
												const next = { ...prev };
												delete next.projectKey;
												return next;
											});
										}}
										maxLength={10}
										aria-invalid={projectErrors.projectKey ? true : undefined}
										className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 font-mono outline-none focus:border-[rgba(91,156,245,0.3)]"
									/>
									{projectErrors.projectKey ? (
										<p className="text-sm text-red-500 mt-1" role="alert">
											{projectFieldError("projectKey")}
										</p>
									) : null}
								</FormField>
							</div>
						) : (
							<FormField
								label={t("common:onboarding.inviteCode")}
								htmlFor="onb-invite"
							>
								<input
									id="onb-invite"
									value={inviteCode}
									onChange={(e) => {
										setInviteCode(e.target.value);
										setProjectErrors((prev) => {
											const next = { ...prev };
											delete next.inviteCode;
											return next;
										});
									}}
									aria-invalid={projectErrors.inviteCode ? true : undefined}
									className="w-full h-11 px-[14px] bg-input border border-b1 rounded-[var(--r2)] text-[14px] text-t1 outline-none focus:border-[rgba(91,156,245,0.3)]"
								/>
								{projectErrors.inviteCode ? (
									<p className="text-sm text-red-500 mt-1" role="alert">
										{projectFieldError("inviteCode")}
									</p>
								) : null}
							</FormField>
						)}

						<div className="flex justify-between items-center mt-10 pt-[22px] border-t border-b1">
							<button
								type="button"
								onClick={() => setStep((s) => s - 1)}
								className="px-5 py-[10px] rounded-[var(--r2)] text-[14px] font-semibold text-t2 border border-b2 transition-colors duration-200 hover:bg-[var(--hover)] hover:text-t1"
							>
								&larr; {t("btn.back")}
							</button>
							<div className="flex gap-2">
								<button
									type="button"
									onClick={() => void handleComplete(true)}
									disabled={loading}
									className="px-5 py-[10px] rounded-[var(--r2)] text-[14px] font-semibold text-t2 border border-b2"
								>
									{t("common:onboarding.skipProject")}
								</button>
								<button
									type="button"
									onClick={() => void handleComplete(false)}
									disabled={loading || !isProjectStepValid}
									className="onb-btn-go px-8 py-3 rounded-[var(--r2)] text-[15px] font-extrabold bg-gradient-to-br from-[var(--greenS)] to-[#047857] text-white shadow-[0_0_24px_var(--gG)] disabled:opacity-50 disabled:cursor-not-allowed"
								>
									{loading ? t("btn.saving") : `${t("btn.next")} →`}
								</button>
							</div>
						</div>
					</div>
				)}
			</div>
		</div>
	);
}

export const Route = createLazyFileRoute("/onboarding")({
	component: OnboardingPage,
});
