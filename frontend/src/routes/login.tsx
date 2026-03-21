import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useAuthStore } from "@/stores/auth-store";
import { getMe } from "@/api/auth";

import { MagicLinkForm } from "@/components/auth/magic-link-form";
import { TelegramButton } from "@/components/auth/telegram-button";
import { DevLoginPanel } from "@/components/auth/dev-login-panel";

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
    <div className="relative flex min-h-screen items-center justify-center px-4 overflow-hidden">
      {/* Градиентный mesh-фон с анимацией */}
      <div
        className="absolute inset-0 animate-mesh"
        style={{
          background: "linear-gradient(-45deg, #0F172A, #1E293B, #1E1B4B, #0F172A, #164E63, #0F172A)",
          backgroundSize: "400% 400%",
        }}
      />

      {/* Декоративные градиентные пятна */}
      <div className="absolute top-1/4 -left-20 h-80 w-80 rounded-full bg-blue-500/20 blur-3xl" />
      <div className="absolute bottom-1/4 -right-20 h-80 w-80 rounded-full bg-purple-500/20 blur-3xl" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 h-96 w-96 rounded-full bg-pink-500/10 blur-3xl" />

      {/* Карточка входа — эффект матового стекла */}
      <div className="relative z-10 w-full max-w-md">
        <div className="rounded-2xl border border-white/10 bg-white/5 backdrop-blur-xl p-8 shadow-2xl shadow-black/20">
          {/* Логотип и заголовок */}
          <div className="text-center mb-8">
            <h1 className="text-4xl font-bold gradient-text mb-2">Hub</h1>
            <p className="text-slate-400 text-sm">
              Единая платформа для команд
            </p>
          </div>

          {/* Форма входа */}
          <div className="space-y-6">
            <div className="[&_label]:text-slate-300 [&_input]:bg-white/10 [&_input]:border-white/20 [&_input]:text-white [&_input]:placeholder:text-slate-500 [&_input]:focus:border-blue-500/50 [&_input]:focus:ring-blue-500/20 [&_button[type=submit]]:gradient-tracker [&_button[type=submit]]:text-white [&_button[type=submit]]:border-0 [&_button[type=submit]]:hover:opacity-90 [&_button[type=submit]]:shadow-lg [&_button[type=submit]]:shadow-blue-500/25 [&_p]:text-slate-400">
              <MagicLinkForm onSuccess={handleLoginSuccess} />
            </div>

            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full h-px bg-white/10" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="px-3 text-slate-500 bg-transparent backdrop-blur-sm">
                  или
                </span>
              </div>
            </div>

            <div className="[&_button]:border-white/20 [&_button]:text-slate-300 [&_button]:hover:bg-white/10">
              <TelegramButton />
            </div>

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
