import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useAuthStore } from "@/stores/auth-store";
import { getMe } from "@/api/auth";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { MagicLinkForm } from "@/components/auth/magic-link-form";
import { TelegramButton } from "@/components/auth/telegram-button";

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

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-3xl font-bold">Hub</CardTitle>
          <CardDescription>
            Войдите в свой аккаунт
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <MagicLinkForm onSuccess={handleLoginSuccess} />
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <Separator />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-card px-2 text-muted-foreground">или</span>
            </div>
          </div>
          <TelegramButton />
        </CardContent>
      </Card>
    </div>
  );
}

export const Route = createFileRoute("/login")({
  component: LoginPage,
});
