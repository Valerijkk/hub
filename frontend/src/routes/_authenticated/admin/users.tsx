import { createFileRoute, redirect } from "@tanstack/react-router";
import { useAuthStore } from "@/stores/auth-store";
import { UserManagement } from "@/components/admin/user-management";

/** Страница управления пользователями (только для администратора). */
function UsersPage() {
  return (
    <div className="mx-auto max-w-6xl p-6">
      <UserManagement />
    </div>
  );
}

export const Route = createFileRoute("/_authenticated/admin/users")({
  beforeLoad: () => {
    const { user } = useAuthStore.getState();
    if (!user || user.system_role !== "admin") {
      throw redirect({ to: "/" });
    }
  },
  component: UsersPage,
});
