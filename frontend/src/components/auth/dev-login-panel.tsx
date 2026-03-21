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
  color: string;
}

const DEMO_USERS: DemoUser[] = [
  {
    email: "admin@hub.local",
    name: "Админ Хабов",
    role: "Admin",
    description: "Полный доступ",
    color: "bg-red-500/15 text-red-700 border-red-200",
  },
  {
    email: "pm@hub.local",
    name: "Петр Менеджеров",
    role: "Owner",
    projectRole: "owner",
    description: "Project Manager",
    color: "bg-purple-500/15 text-purple-700 border-purple-200",
  },
  {
    email: "dev1@hub.local",
    name: "Дмитрий Кодов",
    role: "Developer",
    projectRole: "developer",
    description: "Frontend разработчик",
    color: "bg-blue-500/15 text-blue-700 border-blue-200",
  },
  {
    email: "dev2@hub.local",
    name: "Анна Бэкендова",
    role: "Developer",
    projectRole: "developer",
    description: "Backend разработчик",
    color: "bg-blue-500/15 text-blue-700 border-blue-200",
  },
  {
    email: "tester1@hub.local",
    name: "Тестер Тестеров",
    role: "Tester",
    projectRole: "tester",
    description: "QA инженер",
    color: "bg-green-500/15 text-green-700 border-green-200",
  },
  {
    email: "tester2@hub.local",
    name: "Мария Автотестова",
    role: "Tester",
    projectRole: "tester",
    description: "Автоматизатор",
    color: "bg-green-500/15 text-green-700 border-green-200",
  },
  {
    email: "viewer@hub.local",
    name: "Наблюдатель Смотров",
    role: "Viewer",
    projectRole: "viewer",
    description: "Только просмотр",
    color: "bg-gray-500/15 text-gray-700 border-gray-200",
  },
  {
    email: "newbie@hub.local",
    name: "Новичок Первовходов",
    role: "Member",
    description: "Не прошел онбординг",
    color: "bg-yellow-500/15 text-yellow-700 border-yellow-200",
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

const AVATAR_COLORS = [
  "bg-red-500",
  "bg-purple-500",
  "bg-blue-500",
  "bg-blue-400",
  "bg-green-500",
  "bg-green-400",
  "bg-gray-500",
  "bg-yellow-500",
];

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
      <div className="flex items-center gap-2">
        <div className="h-px flex-1 bg-border" />
        <span className="text-xs font-medium uppercase text-orange-600">
          Dev Mode
        </span>
        <div className="h-px flex-1 bg-border" />
      </div>

      <div className="flex justify-center">
        <Button
          variant="outline"
          size="sm"
          onClick={handleSeed}
          disabled={loading === "seed" || seeded}
        >
          {seeded
            ? "Данные созданы"
            : loading === "seed"
              ? "Создание..."
              : "Создать демо-данные"}
        </Button>
      </div>

      {error && (
        <p className="text-center text-sm text-destructive">{error}</p>
      )}

      <div className="grid grid-cols-2 gap-2">
        {DEMO_USERS.map((user, i) => (
          <button
            key={user.email}
            onClick={() => handleLogin(user.email)}
            disabled={loading !== null}
            className="flex items-start gap-3 rounded-lg border p-3 text-left transition-colors hover:bg-accent disabled:opacity-50"
          >
            <div
              className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-xs font-bold text-white ${AVATAR_COLORS[i]}`}
            >
              {getInitials(user.name)}
            </div>
            <div className="min-w-0 space-y-1">
              <p className="text-sm font-medium leading-tight truncate">
                {user.name}
              </p>
              <Badge
                className={`${user.color} text-[10px] px-1.5 py-0`}
              >
                {user.role}
              </Badge>
              <p className="text-[11px] text-muted-foreground leading-tight">
                {user.description}
              </p>
            </div>
          </button>
        ))}
      </div>

      {loading && loading !== "seed" && (
        <p className="text-center text-sm text-muted-foreground">
          Вход как {DEMO_USERS.find((u) => u.email === loading)?.name}...
        </p>
      )}
    </div>
  );
}
