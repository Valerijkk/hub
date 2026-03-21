import { useState } from "react";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import { completeOnboarding } from "@/api/auth";
import { useAuthStore } from "@/stores/auth-store";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

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
    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
    </svg>
  ),
  /** Иконка мессенджера (чат). */
  messenger: (
    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
    </svg>
  ),
  /** Иконка TMS (тестирование). */
  tms: (
    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  /** Иконка звонков (телефон). */
  calls: (
    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
    </svg>
  ),
  /** Иконка дашбордов (графики). */
  dashboards: (
    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
    </svg>
  ),
};

/** Описания модулей платформы для онбординга. */
const MODULES = [
  {
    key: "tracker",
    title: "Трекер задач",
    desc: "Kanban и Scrum доски, кастомные workflow, спринты и бэклог.",
    color: "bg-blue-500/10 text-blue-600 dark:text-blue-400",
  },
  {
    key: "messenger",
    title: "Мессенджер",
    desc: "Каналы, треды, реакции, файлы. Интегрирован с задачами и тестами.",
    color: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400",
  },
  {
    key: "tms",
    title: "TMS",
    desc: "Тест-кейсы, прогоны, интеграция с Allure. Связь задач и тестов.",
    color: "bg-purple-500/10 text-purple-600 dark:text-purple-400",
  },
  {
    key: "calls",
    title: "Звонки",
    desc: "Голосовые и видеозвонки, демонстрация экрана, запись. До 50 участников.",
    color: "bg-orange-500/10 text-orange-600 dark:text-orange-400",
  },
  {
    key: "dashboards",
    title: "Дашборды",
    desc: "Виджеты, метрики проекта, графики скорости и качества. Кастомные панели.",
    color: "bg-pink-500/10 text-pink-600 dark:text-pink-400",
  },
];

/** Варианты темы интерфейса. */
type ThemeOption = "light" | "dark" | "system";

/** Варианты языка интерфейса. */
type LangOption = "ru" | "en";

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
  const [theme, setTheme] = useState<ThemeOption>("system");

  /** Выбранный язык (шаг 3). */
  const [lang, setLang] = useState<LangOption>("ru");

  const steps = [
    { title: "Заполните профиль", description: "Расскажите о себе" },
    { title: "Обзор платформы", description: "Познакомьтесь с возможностями Hub" },
    { title: "Настройки", description: "Тема, язык и первый проект" },
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

  /** Конфигурация тем для выбора. */
  const themeOptions: { value: ThemeOption; label: string; icon: React.ReactNode }[] = [
    {
      value: "light",
      label: "Светлая",
      icon: (
        <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
      ),
    },
    {
      value: "dark",
      label: "Тёмная",
      icon: (
        <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
        </svg>
      ),
    },
    {
      value: "system",
      label: "Системная",
      icon: (
        <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
        </svg>
      ),
    },
  ];

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <div className="w-full max-w-lg animate-fade-in">
        {/* Логотип */}
        <div className="mb-8 flex items-center justify-center gap-2.5">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
            <span className="text-sm font-bold text-white">H</span>
          </div>
          <span className="text-xl font-bold tracking-tight">Hub</span>
        </div>

        <Card className="rounded-xl border shadow-sm">
          <CardHeader className="text-center pb-4">
            {/* Индикатор шагов */}
            <div className="mb-4 flex justify-center gap-2">
              {steps.map((_, i) => (
                <div
                  key={i}
                  className={`h-1.5 rounded-full transition-all duration-300 ${
                    i <= step
                      ? "w-8 bg-primary"
                      : "w-8 bg-muted"
                  }`}
                />
              ))}
            </div>
            <CardTitle className="text-xl">{currentStep?.title}</CardTitle>
            <CardDescription>{currentStep?.description}</CardDescription>
          </CardHeader>
          <CardContent>
            {/* Шаг 1: Профиль */}
            {step === 0 && (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="onb-name" className="text-sm font-medium">Имя</Label>
                  <Input
                    id="onb-name"
                    placeholder="Ваше имя"
                    value={profile.name}
                    onChange={(e) =>
                      setProfile((p) => ({ ...p, name: e.target.value }))
                    }
                    autoFocus
                    className="h-10"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="onb-position" className="text-sm font-medium">Должность</Label>
                  <Input
                    id="onb-position"
                    placeholder="Например: Frontend Developer"
                    value={profile.position}
                    onChange={(e) =>
                      setProfile((p) => ({ ...p, position: e.target.value }))
                    }
                    className="h-10"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="onb-department" className="text-sm font-medium">Отдел</Label>
                  <Input
                    id="onb-department"
                    placeholder="Например: Engineering"
                    value={profile.department}
                    onChange={(e) =>
                      setProfile((p) => ({ ...p, department: e.target.value }))
                    }
                    className="h-10"
                  />
                </div>
              </div>
            )}

            {/* Шаг 2: Обзор модулей платформы (5 модулей с иконками) */}
            {step === 1 && (
              <div className="space-y-3">
                {MODULES.map((module) => (
                  <div
                    key={module.key}
                    className="flex items-start gap-3 rounded-lg border bg-card p-4"
                  >
                    <div className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg ${module.color}`}>
                      {MODULE_ICONS[module.key]}
                    </div>
                    <div>
                      <h3 className="text-sm font-medium">{module.title}</h3>
                      <p className="text-xs text-muted-foreground mt-0.5 leading-relaxed">{module.desc}</p>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Шаг 3: Тема, Язык, Проект */}
            {step === 2 && (
              <div className="space-y-6">
                {/* Выбор темы */}
                <div className="space-y-2">
                  <Label className="text-sm font-medium">Тема оформления</Label>
                  <div className="grid grid-cols-3 gap-2">
                    {themeOptions.map((opt) => (
                      <button
                        key={opt.value}
                        type="button"
                        className={`flex flex-col items-center gap-1.5 rounded-lg border p-3 transition-all ${
                          theme === opt.value
                            ? "border-primary bg-primary/5 ring-1 ring-primary/30"
                            : "hover:bg-muted/50"
                        }`}
                        onClick={() => setTheme(opt.value)}
                      >
                        <span className={theme === opt.value ? "text-primary" : "text-muted-foreground"}>
                          {opt.icon}
                        </span>
                        <span className="text-xs font-medium">{opt.label}</span>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Выбор языка */}
                <div className="space-y-2">
                  <Label className="text-sm font-medium">Язык интерфейса</Label>
                  <div className="grid grid-cols-2 gap-2">
                    {(
                      [
                        { value: "ru" as LangOption, label: "Русский", flag: "RU" },
                        { value: "en" as LangOption, label: "English", flag: "EN" },
                      ] as const
                    ).map((opt) => (
                      <button
                        key={opt.value}
                        type="button"
                        className={`flex items-center gap-2 rounded-lg border p-3 transition-all ${
                          lang === opt.value
                            ? "border-primary bg-primary/5 ring-1 ring-primary/30"
                            : "hover:bg-muted/50"
                        }`}
                        onClick={() => setLang(opt.value)}
                      >
                        <span className="flex h-6 w-6 items-center justify-center rounded bg-muted text-[10px] font-bold">
                          {opt.flag}
                        </span>
                        <span className="text-sm font-medium">{opt.label}</span>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Создание / присоединение к проекту */}
                <div className="space-y-3 border-t pt-4">
                  <Label className="text-sm font-medium">Начало работы</Label>
                  <div className="flex flex-col gap-2">
                    <Button
                      variant="outline"
                      className="justify-start gap-2 h-11"
                      onClick={() => {
                        void handleComplete();
                      }}
                    >
                      <svg className="h-4 w-4 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
                      </svg>
                      Создать проект
                    </Button>
                    <button
                      type="button"
                      className="text-xs text-muted-foreground hover:text-primary text-left px-1 transition-colors"
                      onClick={() => {
                        void handleComplete();
                      }}
                    >
                      Присоединиться по инвайту
                    </button>
                  </div>
                </div>
              </div>
            )}

            {/* Навигация */}
            <div className="mt-6 flex justify-between border-t pt-4">
              <Button
                variant="ghost"
                onClick={() => setStep((s) => s - 1)}
                disabled={step === 0}
                className="text-muted-foreground"
              >
                <svg className="mr-1.5 h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M10 12L6 8l4-4" /></svg>
                Назад
              </Button>
              {step < steps.length - 1 ? (
                <Button
                  onClick={() => setStep((s) => s + 1)}
                  className="bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white shadow-sm shadow-blue-500/20"
                >
                  Далее
                  <svg className="ml-1.5 h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M6 12l4-4-4-4" /></svg>
                </Button>
              ) : (
                <Button
                  onClick={handleComplete}
                  disabled={loading}
                  className="bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white shadow-sm shadow-blue-500/20"
                >
                  {loading ? "Сохранение..." : "Завершить"}
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export const Route = createFileRoute("/onboarding")({
  component: OnboardingPage,
});
