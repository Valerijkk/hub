import { createFileRoute, redirect } from "@tanstack/react-router";
import { useAuthStore } from "@/stores/auth-store";
import { ResourceLimits } from "@/components/admin/resource-limits";

/** Страница настройки лимитов ресурсов (только для администратора). */
function LimitsPage() {
  return (
    <div className="mx-auto max-w-4xl p-6">
      <ResourceLimits />
    </div>
  );
}

export const Route = createFileRoute("/_authenticated/admin/limits")({
  beforeLoad: () => {
    const { user } = useAuthStore.getState();
    if (!user || user.system_role !== "admin") {
      throw redirect({ to: "/" });
    }
  },
  component: LimitsPage,
});
