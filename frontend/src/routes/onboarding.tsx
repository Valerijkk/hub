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
 * Три шага: профиль, обзор платформы, завершение.
 */

/** Иконки и цвета модулей платформы */
const MODULES = [
  {
    title: "Трекер задач",
    desc: "Kanban и Scrum доски, кастомные workflow, спринты и бэклог.",
    color: "bg-blue-500/10 text-blue-600 dark:text-blue-400",
  },
  {
    title: "Мессенджер",
    desc: "Каналы, треды, реакции, файлы. Интегрирован с задачами и тестами.",
    color: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400",
  },
  {
    title: "TMS",
    desc: "Тест-кейсы, прогоны, интеграция с Allure. Связь задач и тестов.",
    color: "bg-purple-500/10 text-purple-600 dark:text-purple-400",
  },
  {
    title: "Созвоны",
    desc: "Голосовые звонки, демонстрация экрана, запись. До 50 участников.",
    color: "bg-orange-500/10 text-orange-600 dark:text-orange-400",
  },
];

function OnboardingPage() {
  const navigate = useNavigate();
  const { user, setUser } = useAuthStore();
  const [step, setStep] = useState(0);
  const [loading, setLoading] = useState(false);

  const [profile, setProfile] = useState({
    name: user?.name ?? "",
    position: user?.position ?? "",
    department: user?.department ?? "",
  });

  const steps = [
    { title: "Заполните профиль", description: "Расскажите о себе" },
    { title: "Обзор платформы", description: "Познакомьтесь с возможностями Hub" },
    { title: "Готово!", description: "Начните работу" },
  ];

  const currentStep = steps[step];

  async function handleComplete() {
    setLoading(true);
    try {
      const updated = await completeOnboarding({ profile });
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

            {/* Шаг 2: Обзор модулей */}
            {step === 1 && (
              <div className="space-y-3">
                {MODULES.map((module) => (
                  <div
                    key={module.title}
                    className="flex items-start gap-3 rounded-lg border bg-card p-4"
                  >
                    <div className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg ${module.color}`}>
                      <span className="text-sm font-bold">{module.title.charAt(0)}</span>
                    </div>
                    <div>
                      <h3 className="text-sm font-medium">{module.title}</h3>
                      <p className="text-xs text-muted-foreground mt-0.5 leading-relaxed">{module.desc}</p>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Шаг 3: Завершение */}
            {step === 2 && (
              <div className="flex flex-col items-center py-6">
                <div className="flex h-14 w-14 items-center justify-center rounded-full bg-emerald-500/10 mb-4">
                  <svg className="h-7 w-7 text-emerald-500" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M3.5 8.5l3 3 6-6" /></svg>
                </div>
                <p className="text-center text-muted-foreground leading-relaxed max-w-xs">
                  Все готово! Создайте свой первый проект или присоединитесь к существующему.
                </p>
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
