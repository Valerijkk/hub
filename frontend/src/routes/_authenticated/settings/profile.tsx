import { useState, type FormEvent } from "react";
import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { getMe, updateProfile, listSessions, revokeSession } from "@/api/auth";
import { setCustomStatus } from "@/api/messenger";
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

/**
 * Страница настроек профиля — полная форма данных пользователя,
 * кастомный статус, часы работы и список активных сессий.
 */
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
    birthday: user?.birthday ?? "",
  });

  /** Кастомный статус пользователя (например, "На обеде", "В отпуске"). */
  const [customStatusText, setCustomStatusText] = useState("");

  /** Часы работы — начало и конец рабочего дня. */
  const [workHoursStart, setWorkHoursStart] = useState("09:00");
  const [workHoursEnd, setWorkHoursEnd] = useState("18:00");

  /* Синхронизация формы при загрузке данных пользователя */
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
      birthday: user.birthday ?? "",
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

  /** Мутация для сохранения кастомного статуса через Messenger API. */
  const statusMutation = useMutation({
    mutationFn: (text: string) => setCustomStatus(text),
    onSuccess: () => toast.success("Статус обновлён"),
    onError: () => toast.error("Не удалось обновить статус"),
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
      birthday: form.birthday || undefined,
    });
  }

  /** Сохранить кастомный статус. */
  function handleStatusSave() {
    statusMutation.mutate(customStatusText);
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6 animate-fade-in">
      {/* Заголовок страницы */}
      <div className="mb-2">
        <h1 className="text-2xl font-semibold tracking-tight">Настройки профиля</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Управляйте персональными данными и активными сессиями
        </p>
      </div>

      {/* Карточка кастомного статуса */}
      <Card className="rounded-xl border shadow-sm">
        <CardHeader className="pb-4">
          <CardTitle className="text-lg">Кастомный статус</CardTitle>
          <CardDescription>Установите текст статуса, видимый коллегам</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-3">
            <Input
              placeholder="Например: На обеде, В отпуске, Не беспокоить"
              value={customStatusText}
              onChange={(e) => setCustomStatusText(e.target.value)}
              className="h-10 flex-1"
            />
            <Button
              onClick={handleStatusSave}
              disabled={statusMutation.isPending}
              variant="outline"
              className="shrink-0"
            >
              {statusMutation.isPending ? "..." : "Сохранить"}
            </Button>
            {customStatusText && (
              <Button
                variant="ghost"
                className="shrink-0 text-muted-foreground"
                onClick={() => {
                  setCustomStatusText("");
                  statusMutation.mutate("");
                }}
              >
                Сбросить
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Карточка профиля */}
      <Card className="rounded-xl border shadow-sm">
        <CardHeader className="pb-4">
          <CardTitle className="text-lg">Профиль</CardTitle>
          <CardDescription>Ваши персональные данные</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-5">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="prof-name" className="text-sm font-medium">Имя</Label>
                <Input
                  id="prof-name"
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  className="h-10"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prof-display" className="text-sm font-medium">Отображаемое имя</Label>
                <Input
                  id="prof-display"
                  value={form.display_name}
                  onChange={(e) => setForm((f) => ({ ...f, display_name: e.target.value }))}
                  className="h-10"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prof-phone" className="text-sm font-medium">Телефон</Label>
                <Input
                  id="prof-phone"
                  value={form.phone}
                  onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))}
                  className="h-10"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prof-position" className="text-sm font-medium">Должность</Label>
                <Input
                  id="prof-position"
                  value={form.position}
                  onChange={(e) => setForm((f) => ({ ...f, position: e.target.value }))}
                  className="h-10"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prof-department" className="text-sm font-medium">Отдел</Label>
                <Input
                  id="prof-department"
                  value={form.department}
                  onChange={(e) => setForm((f) => ({ ...f, department: e.target.value }))}
                  className="h-10"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prof-tz" className="text-sm font-medium">Часовой пояс</Label>
                <Input
                  id="prof-tz"
                  value={form.timezone}
                  onChange={(e) => setForm((f) => ({ ...f, timezone: e.target.value }))}
                  placeholder="Europe/Moscow"
                  className="h-10"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prof-lang" className="text-sm font-medium">Язык</Label>
                <Input
                  id="prof-lang"
                  value={form.language}
                  onChange={(e) => setForm((f) => ({ ...f, language: e.target.value }))}
                  placeholder="ru"
                  className="h-10"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prof-birthday" className="text-sm font-medium">День рождения</Label>
                <Input
                  id="prof-birthday"
                  type="date"
                  value={form.birthday}
                  onChange={(e) => setForm((f) => ({ ...f, birthday: e.target.value }))}
                  className="h-10"
                />
              </div>
            </div>

            {/* Email и Telegram — только чтение */}
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label className="text-sm font-medium">Email</Label>
                <p className="flex items-center gap-2 text-sm text-muted-foreground">
                  <svg className="h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"><rect x="1.5" y="3" width="13" height="10" rx="2" /><path d="M1.5 5l6.5 4 6.5-4" /></svg>
                  {user?.email ?? "-"}
                </p>
              </div>
              <div className="space-y-2">
                <Label className="text-sm font-medium">Telegram</Label>
                <p className="flex items-center gap-2 text-sm text-muted-foreground">
                  <svg className="h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"><path d="M14 2L2 8l4 2m8-8l-8 10-2-4m10-6l-10 6" /></svg>
                  {user?.telegram_id ? `ID: ${user.telegram_id}` : "Не привязан"}
                </p>
              </div>
            </div>

            <div className="border-t pt-4">
              <Button
                type="submit"
                disabled={profileMutation.isPending}
                className="bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white shadow-sm shadow-blue-500/20"
              >
                {profileMutation.isPending ? "Сохранение..." : "Сохранить изменения"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {/* Карточка часов работы */}
      <Card className="rounded-xl border shadow-sm">
        <CardHeader className="pb-4">
          <CardTitle className="text-lg">Часы работы</CardTitle>
          <CardDescription>
            Уведомления вне рабочих часов будут приглушены
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-3">
            <div className="space-y-1.5">
              <Label className="text-xs text-muted-foreground">Начало</Label>
              <Input
                type="time"
                value={workHoursStart}
                onChange={(e) => setWorkHoursStart(e.target.value)}
                className="h-9 w-32"
              />
            </div>
            <span className="mt-5 text-muted-foreground">--</span>
            <div className="space-y-1.5">
              <Label className="text-xs text-muted-foreground">Конец</Label>
              <Input
                type="time"
                value={workHoursEnd}
                onChange={(e) => setWorkHoursEnd(e.target.value)}
                className="h-9 w-32"
              />
            </div>
            <div className="mt-5">
              <Badge variant="outline" className="text-xs">
                {(() => {
                  const [sh, sm] = workHoursStart.split(":").map(Number);
                  const [eh, em] = workHoursEnd.split(":").map(Number);
                  const diff = (eh! * 60 + em!) - (sh! * 60 + sm!);
                  const hours = Math.floor(diff / 60);
                  const mins = diff % 60;
                  return `${hours}ч${mins > 0 ? ` ${mins}м` : ""}`;
                })()}
              </Badge>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Карточка сессий */}
      <Card className="rounded-xl border shadow-sm">
        <CardHeader className="pb-4">
          <CardTitle className="text-lg">Активные сессии</CardTitle>
          <CardDescription>
            Управляйте сессиями на своих устройствах
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {sessions?.map((session: SessionResponse) => (
              <div
                key={session.id}
                className="flex items-center justify-between rounded-lg border bg-card p-3.5"
              >
                <div className="flex items-start gap-3 min-w-0">
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-muted">
                    <svg className="h-4 w-4 text-muted-foreground" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"><rect x="2" y="3" width="12" height="8" rx="1.5" /><path d="M5 14h6" /><path d="M8 11v3" /></svg>
                  </div>
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium truncate">{session.device || "Неизвестное устройство"}</p>
                      {session.is_current && (
                        <Badge
                          variant="default"
                          className="bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-0 text-[10px]"
                        >
                          Текущая
                        </Badge>
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      IP: {session.ip_address ?? "-"} &middot; Активность:{" "}
                      {new Date(session.last_active_at).toLocaleString("ru-RU")}
                    </p>
                  </div>
                </div>
                {!session.is_current && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => revokeMutation.mutate(session.id)}
                    disabled={revokeMutation.isPending}
                    className="shrink-0 text-destructive hover:text-destructive hover:bg-destructive/10"
                  >
                    Завершить
                  </Button>
                )}
              </div>
            ))}
            {(!sessions || sessions.length === 0) && (
              <p className="py-4 text-center text-sm text-muted-foreground">
                Нет активных сессий
              </p>
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
