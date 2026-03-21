import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useAuthStore } from "@/stores/auth-store";
import { getMe } from "@/api/auth";

import { MagicLinkForm } from "@/components/auth/magic-link-form";
import { TelegramButton } from "@/components/auth/telegram-button";
import { DevLoginPanel } from "@/components/auth/dev-login-panel";

/**
 * Страница входа — split-layout с hero-областью и формой.
 *
 * Левая часть: градиентный фон с брендингом Hub.
 * Правая часть: форма входа на чистом фоне.
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
        {/* Анимированный mesh-фон */}
        <div
          className="absolute inset-0 animate-mesh"
          style={{
            background: "linear-gradient(-45deg, #0C1222, #1E293B, #1E1B4B, #0C1222, #164E63, #0C1222)",
            backgroundSize: "400% 400%",
          }}
        />

        {/* Тонкие градиентные пятна */}
        <div className="absolute top-1/4 -left-20 h-[500px] w-[500px] rounded-full bg-blue-500/15 blur-[120px]" />
        <div className="absolute bottom-1/4 right-0 h-[400px] w-[400px] rounded-full bg-purple-500/15 blur-[120px]" />
        <div className="absolute bottom-0 left-1/3 h-[300px] w-[300px] rounded-full bg-cyan-500/10 blur-[100px]" />

        {/* Контент hero */}
        <div className="relative z-10 flex flex-1 flex-col justify-center px-12 xl:px-16">
          <div className="max-w-md">
            <div className="mb-8 flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-blue-500 to-purple-600">
                <span className="text-lg font-bold text-white">H</span>
              </div>
              <h1 className="text-3xl font-bold text-white tracking-tight">Hub</h1>
            </div>
            <p className="text-xl font-medium text-white/90 leading-relaxed mb-4">
              Единая платформа для управления проектами и командой
            </p>
            <p className="text-sm text-slate-400 leading-relaxed">
              Трекер задач, мессенджер, тест-менеджмент и видеозвонки — все инструменты в одном месте.
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
      <div className="flex flex-1 flex-col items-center justify-center px-6 py-12 lg:px-12 bg-background">
        <div className="w-full max-w-sm">
          {/* Логотип — только на мобильных */}
          <div className="mb-8 text-center lg:text-left">
            <div className="mb-6 flex items-center justify-center gap-3 lg:hidden">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-blue-500 to-purple-600">
                <span className="text-lg font-bold text-white">H</span>
              </div>
              <h1 className="text-2xl font-bold tracking-tight gradient-text">Hub</h1>
            </div>
            <h2 className="text-xl font-semibold tracking-tight text-foreground">
              Войти в аккаунт
            </h2>
            <p className="mt-1.5 text-sm text-muted-foreground">
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
                <span className="px-3 text-muted-foreground bg-background">
                  или
                </span>
              </div>
            </div>

            <TelegramButton />

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
