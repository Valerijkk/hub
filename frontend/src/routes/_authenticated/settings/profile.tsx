import { useState, type FormEvent } from "react";
import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { getMe, updateProfile, listSessions, revokeSession } from "@/api/auth";
import type { SessionResponse } from "@/api/auth";
import { useAuthStore } from "@/stores/auth-store";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

function ProfilePage() {
  const queryClient = useQueryClient();
  const setUser = useAuthStore((s) => s.setUser);

  const { data: user } = useQuery({
    queryKey: ["me"],
    queryFn: getMe,
  });

  const { data: sessions } = useQuery({
    queryKey: ["sessions"],
    queryFn: listSessions,
  });

  const [form, setForm] = useState({
    name: user?.name ?? "",
    display_name: user?.display_name ?? "",
    phone: user?.phone ?? "",
    position: user?.position ?? "",
    department: user?.department ?? "",
    timezone: user?.timezone ?? "",
    language: user?.language ?? "",
  });

  // Sync form when user data loads
  const [initialized, setInitialized] = useState(false);
  if (user && !initialized) {
    setForm({
      name: user.name ?? "",
      display_name: user.display_name ?? "",
      phone: user.phone ?? "",
      position: user.position ?? "",
      department: user.department ?? "",
      timezone: user.timezone ?? "",
      language: user.language ?? "",
    });
    setInitialized(true);
  }

  const profileMutation = useMutation({
    mutationFn: updateProfile,
    onSuccess: (updated) => {
      setUser(updated);
      void queryClient.invalidateQueries({ queryKey: ["me"] });
      toast.success("Профиль обновлен");
    },
    onError: () => toast.error("Не удалось обновить профиль"),
  });

  const revokeMutation = useMutation({
    mutationFn: revokeSession,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["sessions"] });
      toast.success("Сессия завершена");
    },
    onError: () => toast.error("Не удалось завершить сессию"),
  });

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    profileMutation.mutate({
      name: form.name || undefined,
      display_name: form.display_name || undefined,
      phone: form.phone || undefined,
      position: form.position || undefined,
      department: form.department || undefined,
      timezone: form.timezone || undefined,
      language: form.language || undefined,
    });
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <h1 className="text-2xl font-bold">Настройки профиля</h1>

      <Card>
        <CardHeader>
          <CardTitle>Профиль</CardTitle>
          <CardDescription>Ваши персональные данные</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="prof-name">Имя</Label>
                <Input
                  id="prof-name"
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prof-display">Отображаемое имя</Label>
                <Input
                  id="prof-display"
                  value={form.display_name}
                  onChange={(e) => setForm((f) => ({ ...f, display_name: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prof-phone">Телефон</Label>
                <Input
                  id="prof-phone"
                  value={form.phone}
                  onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prof-position">Должность</Label>
                <Input
                  id="prof-position"
                  value={form.position}
                  onChange={(e) => setForm((f) => ({ ...f, position: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prof-department">Отдел</Label>
                <Input
                  id="prof-department"
                  value={form.department}
                  onChange={(e) => setForm((f) => ({ ...f, department: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prof-tz">Часовой пояс</Label>
                <Input
                  id="prof-tz"
                  value={form.timezone}
                  onChange={(e) => setForm((f) => ({ ...f, timezone: e.target.value }))}
                  placeholder="Europe/Moscow"
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Email</Label>
              <p className="text-sm text-muted-foreground">{user?.email ?? "-"}</p>
            </div>
            <Button type="submit" disabled={profileMutation.isPending}>
              {profileMutation.isPending ? "Сохранение..." : "Сохранить"}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Активные сессии</CardTitle>
          <CardDescription>
            Управляйте сессиями на своих устройствах
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {sessions?.map((session: SessionResponse) => (
              <div
                key={session.id}
                className="flex items-center justify-between rounded-md border p-3"
              >
                <div>
                  <div className="flex items-center gap-2">
                    <p className="font-medium">{session.device || "Неизвестное устройство"}</p>
                    {session.is_current && (
                      <Badge variant="default">Текущая</Badge>
                    )}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    IP: {session.ip_address ?? "-"} | Последняя активность:{" "}
                    {new Date(session.last_active_at).toLocaleString("ru-RU")}
                  </p>
                </div>
                {!session.is_current && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => revokeMutation.mutate(session.id)}
                    disabled={revokeMutation.isPending}
                  >
                    Завершить
                  </Button>
                )}
              </div>
            ))}
            {(!sessions || sessions.length === 0) && (
              <p className="text-sm text-muted-foreground">Нет активных сессий</p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export const Route = createFileRoute("/_authenticated/settings/profile")({
  component: ProfilePage,
});
