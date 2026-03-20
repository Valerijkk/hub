import { useState, type FormEvent } from "react";
import { toast } from "sonner";
import { sendMagicLink, verifyMagicLink } from "@/api/auth";
import { useAuthStore } from "@/stores/auth-store";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface MagicLinkFormProps {
  onSuccess: () => void;
}

export function MagicLinkForm({ onSuccess }: MagicLinkFormProps) {
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [step, setStep] = useState<"email" | "code">("email");
  const [loading, setLoading] = useState(false);
  const { setTokens } = useAuthStore();

  const isEmailValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);

  async function handleSendCode(e: FormEvent) {
    e.preventDefault();
    if (!isEmailValid) {
      toast.error("Введите корректный email");
      return;
    }
    setLoading(true);
    try {
      await sendMagicLink({ email });
      setStep("code");
      toast.success("Код отправлен на " + email);
    } catch (err: unknown) {
      const message =
        err && typeof err === "object" && "message" in err
          ? (err as { message: string }).message
          : "Не удалось отправить код";
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function handleVerifyCode(e: FormEvent) {
    e.preventDefault();
    if (!code.trim()) {
      toast.error("Введите код");
      return;
    }
    setLoading(true);
    try {
      const tokens = await verifyMagicLink({ email, code: code.trim() });
      setTokens(tokens.access_token, tokens.refresh_token);
      onSuccess();
    } catch (err: unknown) {
      const message =
        err && typeof err === "object" && "message" in err
          ? (err as { message: string }).message
          : "Неверный код";
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  if (step === "email") {
    return (
      <form onSubmit={handleSendCode} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            type="email"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoComplete="email"
            autoFocus
          />
        </div>
        <Button type="submit" className="w-full" disabled={loading || !isEmailValid}>
          {loading ? "Отправка..." : "Отправить код"}
        </Button>
      </form>
    );
  }

  return (
    <form onSubmit={handleVerifyCode} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="code">Код подтверждения</Label>
        <Input
          id="code"
          type="text"
          inputMode="numeric"
          placeholder="000000"
          maxLength={6}
          value={code}
          onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
          autoFocus
        />
        <p className="text-sm text-muted-foreground">
          Код отправлен на {email}
        </p>
      </div>
      <Button type="submit" className="w-full" disabled={loading || !code.trim()}>
        {loading ? "Проверка..." : "Войти"}
      </Button>
      <Button
        type="button"
        variant="ghost"
        className="w-full"
        onClick={() => {
          setStep("email");
          setCode("");
        }}
      >
        Изменить email
      </Button>
    </form>
  );
}
