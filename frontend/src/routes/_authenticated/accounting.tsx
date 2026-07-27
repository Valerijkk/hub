import { createFileRoute, redirect } from "@tanstack/react-router";
import { isSystemAdmin } from "@/lib/is-system-admin";
import { useAuthStore } from "@/stores/auth-store";

/**
 * A13-017: stub — page in `accounting.lazy.tsx`; beforeLoad stays eager (auth gate).
 */
export const Route = createFileRoute("/_authenticated/accounting")({
	beforeLoad: () => {
		const user = useAuthStore.getState().user;
		if (!user || !isSystemAdmin(user.system_role)) {
			throw redirect({ to: "/dashboard" });
		}
	},
});
