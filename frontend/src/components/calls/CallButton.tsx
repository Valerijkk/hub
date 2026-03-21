import { useCallback, useEffect, useState } from "react";
import { Loader2, Phone } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useCallStore } from "@/stores/call-store";
import type { CallResponse } from "@/api/calls";

/** Пропсы компонента CallButton. */
interface CallButtonProps {
  /** ID канала для звонка. */
  channelId: string;
}

/**
 * Кнопка для начала или присоединения к звонку из канала.
 * Проверяет наличие активного звонка и отображает
 * соответствующее действие: "Начать звонок" или "Присоединиться".
 */
export function CallButton({ channelId }: CallButtonProps) {
  const { startCall, joinCall, checkActiveCall, isConnecting, isInCall } =
    useCallStore();
  const [existingCall, setExistingCall] = useState<CallResponse | null>(null);
  const [checking, setChecking] = useState(false);

  /** Проверить наличие активного звонка при монтировании и смене канала. */
  useEffect(() => {
    let cancelled = false;
    const check = async () => {
      setChecking(true);
      const call = await checkActiveCall(channelId);
      if (!cancelled) {
        setExistingCall(call);
        setChecking(false);
      }
    };
    void check();
    return () => {
      cancelled = true;
    };
  }, [channelId, checkActiveCall]);

  /** Обработчик клика: начать или присоединиться. */
  const handleClick = useCallback(async () => {
    if (existingCall) {
      await joinCall(existingCall.id);
    } else {
      await startCall(channelId);
    }
  }, [existingCall, joinCall, startCall, channelId]);

  /** Не показывать кнопку если уже в звонке. */
  if (isInCall) return null;

  const loading = isConnecting || checking;
  const label = existingCall ? "Присоединиться к звонку" : "Начать звонок";

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            onClick={handleClick}
            disabled={loading}
            aria-label={label}
          >
            {loading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Phone className="h-4 w-4" />
            )}
          </Button>
        </TooltipTrigger>
        <TooltipContent>{label}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
