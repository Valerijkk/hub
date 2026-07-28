import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, render, renderHook, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as authApi from "@/api/auth";
import { api } from "@/api/client";
import * as dashboardApi from "@/api/dashboard";
import { ActivityFeed } from "@/components/tracker/activity-feed";
import { ErrorBoundary } from "@/components/ui/error-boundary";
import { useAuth } from "@/hooks/use-auth";
import { useAuthStore } from "@/stores/auth-store";
import { useDashboardStore } from "@/stores/dashboard-store";

function resetAuthStore(): void {
	sessionStorage.clear();
	useAuthStore.setState({
		accessToken: null,
		refreshToken: null,
		user: null,
		isAuthenticated: false,
	});
}

describe("FE-P1-ERROR-AS-EMPTY (TASK-033)", () => {
	beforeEach(() => {
		resetAuthStore();
		vi.restoreAllMocks();
		useDashboardStore.setState({
			personalData: null,
			projectData: null,
			adminData: null,
			layouts: {},
			loading: false,
			error: null,
		});
	});

	it("ignores late getMe setUser after logout (auth race)", async () => {
		let resolveMe!: (v: authApi.UserResponse) => void;
		const mePromise = new Promise<authApi.UserResponse>((resolve) => {
			resolveMe = resolve;
		});
		vi.spyOn(authApi, "getMe").mockReturnValue(mePromise);
		vi.spyOn(authApi, "logout").mockResolvedValue(undefined);

		useAuthStore.getState().setTokens("token-a", "refresh-a");

		const { result } = renderHook(() => useAuth());

		await act(async () => {
			await result.current.logout();
		});

		expect(useAuthStore.getState().user).toBeNull();
		expect(useAuthStore.getState().accessToken).toBeNull();

		await act(async () => {
			resolveMe({
				id: "u-old",
				email: "old@hub.local",
				name: "Old User",
				system_role: "member",
				timezone: "UTC",
				language: "ru",
				onboarding_completed: true,
			});
			await Promise.resolve();
		});

		await waitFor(() => {
			expect(useAuthStore.getState().user).toBeNull();
		});
	});

	it("ActivityFeed shows error — not «Нет активности» — on query fail", async () => {
		vi.spyOn(api, "get").mockRejectedValue(new Error("network"));

		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<ActivityFeed
					projectId="11111111-1111-4111-8111-111111111111"
					issueKey="HUB-1"
				/>
			</QueryClientProvider>,
		);

		expect(await screen.findByRole("alert")).toHaveTextContent(
			/не удалось загрузить активность/i,
		);
		expect(screen.queryByText(/нет активности/i)).not.toBeInTheDocument();
	});

	it("ErrorBoundary resets when resetKey changes (sidebar nav)", () => {
		function Boom({ fail }: { fail: boolean }) {
			if (fail) throw new Error("boom");
			return <p>ok content</p>;
		}

		const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});

		const { rerender } = render(
			<ErrorBoundary resetKey="/a">
				<Boom fail />
			</ErrorBoundary>,
		);

		expect(screen.getByRole("alert")).toHaveTextContent(/что-то пошло не так/i);

		rerender(
			<ErrorBoundary resetKey="/b">
				<p>recovered</p>
			</ErrorBoundary>,
		);

		expect(screen.getByText("recovered")).toBeInTheDocument();
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
		errSpy.mockRestore();
	});

	it("fetchProject sets error on failure (not silent empty)", async () => {
		vi.spyOn(dashboardApi, "getProjectDashboard").mockRejectedValue(
			new Error("fail"),
		);

		await useDashboardStore.getState().fetchProject("p1");

		expect(useDashboardStore.getState().error).toMatch(/не удалось/i);
		expect(useDashboardStore.getState().loading).toBe(false);
		expect(useDashboardStore.getState().projectData).toBeNull();
	});

	it("fetchAdmin sets error on failure", async () => {
		vi.spyOn(dashboardApi, "getAdminDashboard").mockRejectedValue(
			new Error("fail"),
		);

		await useDashboardStore.getState().fetchAdmin();

		expect(useDashboardStore.getState().error).toMatch(/не удалось/i);
		expect(useDashboardStore.getState().adminData).toBeNull();
	});
});
