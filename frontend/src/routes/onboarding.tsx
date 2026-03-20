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
      <Card className="w-full max-w-lg">
        <CardHeader className="text-center">
          <div className="mb-2 flex justify-center gap-2">
            {steps.map((_, i) => (
              <div
                key={i}
                className={`h-2 w-8 rounded-full transition-colors ${
                  i <= step ? "bg-primary" : "bg-muted"
                }`}
              />
            ))}
          </div>
          <CardTitle>{currentStep?.title}</CardTitle>
          <CardDescription>{currentStep?.description}</CardDescription>
        </CardHeader>
        <CardContent>
          {step === 0 && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="onb-name">Имя</Label>
                <Input
                  id="onb-name"
                  placeholder="Ваше имя"
                  value={profile.name}
                  onChange={(e) =>
                    setProfile((p) => ({ ...p, name: e.target.value }))
                  }
                  autoFocus
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="onb-position">Должность</Label>
                <Input
                  id="onb-position"
                  placeholder="Например: Frontend Developer"
                  value={profile.position}
                  onChange={(e) =>
                    setProfile((p) => ({ ...p, position: e.target.value }))
                  }
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="onb-department">Отдел</Label>
                <Input
                  id="onb-department"
                  placeholder="Например: Engineering"
                  value={profile.department}
                  onChange={(e) =>
                    setProfile((p) => ({ ...p, department: e.target.value }))
                  }
                />
              </div>
            </div>
          )}

          {step === 1 && (
            <div className="space-y-4">
              {[
                {
                  title: "Трекер задач",
                  desc: "Kanban и Scrum доски, кастомные workflow, спринты и бэклог.",
                },
                {
                  title: "Мессенджер",
                  desc: "Каналы, треды, реакции, файлы. Интегрирован с задачами и тестами.",
                },
                {
                  title: "TMS",
                  desc: "Тест-кейсы, прогоны, интеграция с Allure. Связь задач и тестов.",
                },
                {
                  title: "Созвоны",
                  desc: "Голосовые звонки, демонстрация экрана, запись. До 50 участников.",
                },
              ].map((module) => (
                <div
                  key={module.title}
                  className="rounded-lg border bg-card p-4"
                >
                  <h3 className="font-medium">{module.title}</h3>
                  <p className="text-sm text-muted-foreground">{module.desc}</p>
                </div>
              ))}
            </div>
          )}

          {step === 2 && (
            <div className="py-4 text-center">
              <p className="text-lg">
                Все готово! Создайте свой первый проект или присоединитесь к существующему.
              </p>
            </div>
          )}

          <div className="mt-6 flex justify-between">
            <Button
              variant="ghost"
              onClick={() => setStep((s) => s - 1)}
              disabled={step === 0}
            >
              Назад
            </Button>
            {step < steps.length - 1 ? (
              <Button onClick={() => setStep((s) => s + 1)}>Далее</Button>
            ) : (
              <Button onClick={handleComplete} disabled={loading}>
                {loading ? "Сохранение..." : "Завершить"}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export const Route = createFileRoute("/onboarding")({
  component: OnboardingPage,
});
