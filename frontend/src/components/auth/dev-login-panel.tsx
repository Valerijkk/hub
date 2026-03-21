/**
 * Панель быстрого входа для разработки.
 *
 * Карточки пользователей с градиентными акцентами по ролям.
 * Эффект подъёма при наведении, анимация нажатия.
 */
import { useState, useEffect } from "react";
import { api } from "@/api/client";
import { useAuthStore } from "@/stores/auth-store";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { UserResponse, TokenPairResponse } from "@/api/auth";

// --- Типы ---

interface DevLoginResponse extends TokenPairResponse {
  user: UserResponse;
}

interface DevSeedResponse {
  users: Array<{
    id: string;
    email: string;
    name: string;
    system_role: string;
    project_role?: string;
    description: string;
  }>;
  project: { id: string; name: string; key: string };
}

// --- Демо-пользователи ---

interface DemoUser {
  email: string;
  name: string;
  role: string;
  projectRole?: string;
  description: string;
  /** Градиент для аватара */
  avatarGradient: string;
  /** Градиент для левой рамки карточки */
  borderGradient: string;
  /** Цвет бейджа роли */
  badgeClass: string;
}

const DEMO_USERS: DemoUser[] = [
  {
    email: "admin@hub.local",
    name: "Админ Хабов",
    role: "Admin",
    description: "Полный доступ",
    avatarGradient: "from-red-500 to-rose-600",
    borderGradient: "from-red-500 to-rose-600",
    badgeClass: "bg-red-500/20 text-red-300 border-red-500/30",
  },
  {
    email: "pm@hub.local",
    name: "Петр Менеджеров",
    role: "Owner",
    projectRole: "owner",
    description: "Project Manager",
    avatarGradient: "from-purple-500 to-violet-600",
    borderGradient: "from-purple-500 to-violet-600",
    badgeClass: "bg-purple-500/20 text-purple-300 border-purple-500/30",
  },
  {
    email: "dev1@hub.local",
    name: "Дмитрий Кодов",
    role: "Developer",
    projectRole: "developer",
    description: "Frontend разработчик",
    avatarGradient: "from-blue-500 to-blue-600",
    borderGradient: "from-blue-500 to-blue-600",
    badgeClass: "bg-blue-500/20 text-blue-300 border-blue-500/30",
  },
  {
    email: "dev2@hub.local",
    name: "Анна Бэкендова",
    role: "Developer",
    projectRole: "developer",
    description: "Backend разработчик",
    avatarGradient: "from-blue-400 to-cyan-500",
    borderGradient: "from-blue-400 to-cyan-500",
    badgeClass: "bg-blue-500/20 text-blue-300 border-blue-500/30",
  },
  {
    email: "tester1@hub.local",
    name: "Тестер Тестеров",
    role: "Tester",
    projectRole: "tester",
    description: "QA инженер",
    avatarGradient: "from-emerald-500 to-green-600",
    borderGradient: "from-emerald-500 to-green-600",
    badgeClass: "bg-emerald-500/20 text-emerald-300 border-emerald-500/30",
  },
  {
    email: "tester2@hub.local",
    name: "Мария Автотестова",
    role: "Tester",
    projectRole: "tester",
    description: "Автоматизатор",
    avatarGradient: "from-green-400 to-emerald-500",
    borderGradient: "from-green-400 to-emerald-500",
    badgeClass: "bg-emerald-500/20 text-emerald-300 border-emerald-500/30",
  },
  {
    email: "viewer@hub.local",
    name: "Наблюдатель Смотров",
    role: "Viewer",
    projectRole: "viewer",
    description: "Только просмотр",
    avatarGradient: "from-slate-400 to-slate-500",
    borderGradient: "from-slate-400 to-slate-500",
    badgeClass: "bg-slate-500/20 text-slate-300 border-slate-500/30",
  },
  {
    email: "newbie@hub.local",
    name: "Новичок Первовходов",
    role: "Member",
    description: "Не прошел онбординг",
    avatarGradient: "from-amber-400 to-yellow-500",
    borderGradient: "from-amber-400 to-yellow-500",
    badgeClass: "bg-amber-500/20 text-amber-300 border-amber-500/30",
  },
];

// --- Утилиты ---

function getInitials(name: string): string {
  return name
    .split(" ")
    .map((w) => w[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
}

// --- Компонент ---

interface DevLoginPanelProps {
  onSuccess: () => void;
}

export function DevLoginPanel({ onSuccess }: DevLoginPanelProps) {
  const [devMode, setDevMode] = useState<boolean | null>(null);
  const [seeded, setSeeded] = useState(false);
  const [loading, setLoading] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const setTokens = useAuthStore((s) => s.setTokens);
  const setUser = useAuthStore((s) => s.setUser);

  // Проверяем доступность dev-режима через GET /api/dev/status
  useEffect(() => {
    api
      .get<{ dev_mode: boolean }>("/dev/status")
      .then(() => setDevMode(true))
      .catch(() => setDevMode(false));
  }, []);

  if (devMode === null || devMode === false) {
    return null;
  }

  async function handleSeed() {
    setLoading("seed");
    setError(null);
    try {
      await api.post<DevSeedResponse>("/dev/seed");
      setSeeded(true);
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Ошибка при создании данных";
      setError(msg);
    } finally {
      setLoading(null);
    }
  }

  async function handleLogin(email: string) {
    setLoading(email);
    setError(null);
    try {
      // Сначала seed, если ещё не делали
      if (!seeded) {
        await api.post<DevSeedResponse>("/dev/seed");
        setSeeded(true);
      }

      const result = await api.post<DevLoginResponse>("/dev/login", { email });
      setTokens(result.access_token, result.refresh_token);
      setUser(result.user);
      onSuccess();
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Ошибка входа";
      setError(msg);
    } finally {
      setLoading(null);
    }
  }

  return (
    <div className="space-y-4">
      {/* Разделитель Dev Mode */}
      <div className="flex items-center gap-2">
        <div className="h-px flex-1 bg-gradient-to-r from-transparent via-orange-500/50 to-transparent" />
        <span className="text-xs font-bold uppercase tracking-wider text-orange-400">
          Dev Mode
        </span>
        <div className="h-px flex-1 bg-gradient-to-r from-transparent via-orange-500/50 to-transparent" />
      </div>

      {/* Кнопка создания данных */}
      <div className="flex justify-center">
        <Button
          variant="outline"
          size="sm"
          onClick={handleSeed}
          disabled={loading === "seed" || seeded}
          className="border-white/20 text-slate-300 hover:bg-white/10 hover:text-white rounded-xl"
        >
          {seeded
            ? "Данные созданы"
            : loading === "seed"
              ? "Создание..."
              : "Создать демо-данные"}
        </Button>
      </div>

      {error && (
        <p className="text-center text-sm text-red-400">{error}</p>
      )}

      {/* Сетка карточек пользователей */}
      <div className="grid grid-cols-2 gap-2">
        {DEMO_USERS.map((user) => (
          <button
            key={user.email}
            onClick={() => handleLogin(user.email)}
            disabled={loading !== null}
            className="group relative flex items-start gap-3 rounded-xl border border-white/10 p-3 text-left transition-all duration-200 hover:bg-white/10 hover:border-white/20 hover:shadow-lg hover:shadow-black/20 hover:-translate-y-0.5 active:scale-[0.97] disabled:opacity-50 disabled:hover:translate-y-0 overflow-hidden"
          >
            {/* Градиентная левая полоска */}
            <div className={`absolute left-0 top-0 bottom-0 w-[3px] bg-gradient-to-b ${user.borderGradient} opacity-60 group-hover:opacity-100 transition-opacity`} />

            {/* Аватар с градиентным фоном */}
            <div
              className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-gradient-to-br ${user.avatarGradient} text-xs font-bold text-white shadow-lg`}
            >
              {getInitials(user.name)}
            </div>

            <div className="min-w-0 space-y-1">
              <p className="text-sm font-medium leading-tight truncate text-slate-200 group-hover:text-white">
                {user.name}
              </p>
              <Badge
                className={`${user.badgeClass} text-[10px] px-1.5 py-0 border rounded-md`}
              >
                {user.role}
              </Badge>
              <p className="text-[11px] text-slate-500 leading-tight">
                {user.description}
              </p>
            </div>
          </button>
        ))}
      </div>

      {loading && loading !== "seed" && (
        <p className="text-center text-sm text-slate-400">
          Вход как {DEMO_USERS.find((u) => u.email === loading)?.name}...
        </p>
      )}
    </div>
  );
}
