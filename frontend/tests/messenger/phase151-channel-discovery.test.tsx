import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, fireEvent, renderHook, screen, waitFor, within } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
	discoverChannels,
	type Channel,
	type ChannelCapabilities,
} from "@/api/messenger";
import { api } from "@/api/client";
import { ChannelList } from "@/components/messenger/channel-list";
import {
	discoverableChannelsQueryKey,
	useChannels,
	useDiscoverableChannels,
} from "@/hooks/use-channels";
import { useAuthStore } from "@/stores/auth-store";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/client", () => ({
	api: {
		get: vi.fn(),
	},
}));

vi.mock("@/hooks/use-channels", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/hooks/use-channels")>();
	return {
		...actual,
		useChannels: vi.fn(),
	};
});

vi.mock("@/hooks/use-unread", () => ({
	useUnread: vi.fn(),
}));

vi.mock("@/hooks/use-presence", () => ({
	usePresence: () => ({
		getStatus: () => "offline",
		isError: false,
		isLoading: false,
		refetch: vi.fn(),
	}),
}));

vi.mock("@tanstack/react-router", () => ({
	useMatchRoute: () => () => false,
	useNavigate: () => vi.fn(),
	Link: ({ children, ...props }: { children: ReactNode }) => (
		<a {...props}>{children}</a>
	),
}));

const user = {
	id: "user-1",
	email: "user-1@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};

const readOnlyCapabilities: ChannelCapabilities = {
	is_member: false,
	has_active_project_membership: false,
	can_read: true,
	can_post: false,
	can_reply: false,
	can_join: false,
	can_follow: false,
	can_unfollow: false,
	can_view_unread: false,
	can_type: false,
	can_edit: false,
	can_forward: false,
	can_react: false,
	can_delete_message: false,
	can_pin: false,
	can_list_pins: false,
	can_bookmark: false,
	can_promote: false,
};

function makeChannel(
	partial: Partial<Channel> & Pick<Channel, "id" | "name">,
): Channel {
	return {
		channel_type: "public",
		is_readonly: false,
		created_by: "owner-1",
		is_starred: false,
		created_at: "2026-08-26T00:00:00Z",
		updated_at: "2026-08-26T00:00:00Z",
		capabilities: readOnlyCapabilities,
		...partial,
	};
}

const globalChannel = makeChannel({
	id: "global",
	name: "Global announcements",
});

const projectChannel = makeChannel({
	id: "project",
	name: "Project alpha",
	project_id: "project-1",
});

function createQueryClient() {
	return new QueryClient({
		defaultOptions: {
			queries: { retry: false },
			mutations: { retry: false },
		},
	});
}

function createWrapper(queryClient: QueryClient) {
	return function QueryWrapper({ children }: { children: ReactNode }) {
		return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
	};
}

describe("Phase 151 read-only public channel discovery", () => {
	const queryClients: QueryClient[] = [];

	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(api.get).mockReset();
		useAuthStore.setState({ user });
		vi.mocked(useChannels).mockReturnValue({
			data: [],
			isLoading: false,
			isError: false,
			refetch: vi.fn(),
		} as ReturnType<typeof useChannels>);
	});

	afterEach(() => {
		for (const queryClient of queryClients.splice(0)) {
			queryClient.clear();
		}
	});

	it("maps the discovery API to the exact read-only endpoint", async () => {
		vi.mocked(api.get).mockResolvedValue([globalChannel]);

		await expect(discoverChannels()).resolves.toEqual([globalChannel]);
		expect(api.get).toHaveBeenCalledTimes(1);
		expect(api.get).toHaveBeenCalledWith("/messenger/channels/discover");
	});

	it("does not start discovery until Browse opens and scopes its cache to the current identity", async () => {
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		vi.mocked(api.get).mockResolvedValue([globalChannel]);

		const discovery = renderHook(
			({ browseOpen }: { browseOpen: boolean }) =>
				useDiscoverableChannels(browseOpen),
			{
				initialProps: { browseOpen: false },
				wrapper: createWrapper(queryClient),
			},
		);

		expect(api.get).not.toHaveBeenCalled();
		expect(discoverableChannelsQueryKey(user.id)).toEqual([
			"messenger",
			"channels",
			"discover",
			"user-1",
		]);

		discovery.rerender({ browseOpen: true });

		await waitFor(() => {
			expect(discovery.result.current.data).toEqual([globalChannel]);
		});
		expect(api.get).toHaveBeenCalledWith("/messenger/channels/discover");
		expect(
			queryClient.getQueryData(discoverableChannelsQueryKey(user.id)),
		).toEqual([globalChannel]);

		vi.mocked(api.get).mockResolvedValueOnce([projectChannel]);
		act(() => {
			useAuthStore.setState({ user: { ...user, id: "user-2" } });
		});
		await waitFor(() => {
			expect(discovery.result.current.data).toEqual([projectChannel]);
		});
		expect(
			queryClient.getQueryData(discoverableChannelsQueryKey("user-2")),
		).toEqual([projectChannel]);
		expect(
			queryClient.getQueryData(discoverableChannelsQueryKey(user.id)),
		).toEqual([globalChannel]);
	});

	it.each([401, 403])(
		"does not retry authorization status %s and removes the exact identity cache",
		async (code) => {
			const queryClient = createQueryClient();
			queryClients.push(queryClient);
			vi.mocked(api.get).mockRejectedValue({ code, message: "denied" });

			const discovery = renderHook(() => useDiscoverableChannels(true), {
				wrapper: createWrapper(queryClient),
			});

			await waitFor(() => {
				expect(api.get).toHaveBeenCalledTimes(1);
			});
			await waitFor(() => {
				expect(discovery.result.current.data).toBeUndefined();
				expect(discovery.result.current.fetchStatus).toBe("idle");
				expect(
					queryClient.getQueryData(discoverableChannelsQueryKey(user.id)),
				).toBeUndefined();
			});
			await act(async () => {
				await Promise.resolve();
				await Promise.resolve();
			});
			expect(api.get).toHaveBeenCalledTimes(1);
		},
	);

	it("renders only readable public server rows in order as a non-navigating semantic browse list", async () => {
		const privateChannel = makeChannel({
			id: "private",
			name: "Must not appear",
			channel_type: "private",
		});
		const deniedChannel = makeChannel({
			id: "denied",
			name: "Also must not appear",
			capabilities: { ...readOnlyCapabilities, can_read: false },
		});
		vi.mocked(api.get).mockResolvedValue([
			projectChannel,
			globalChannel,
			privateChannel,
			deniedChannel,
		]);

		renderWithProviders(<ChannelList />);
		expect(api.get).not.toHaveBeenCalled();
		fireEvent.click(
			screen.getByRole("button", { name: /\u041e\u0431\u0437\u043e\u0440 \u043a\u0430\u043d\u0430\u043b\u043e\u0432/i }),
		);

		const browseList = await screen.findByRole("list", {
			name: /\u0414\u043e\u0441\u0442\u0443\u043f\u043d\u044b\u0435 \u043a\u0430\u043d\u0430\u043b\u044b/i,
		});
		const rows = within(browseList).getAllByRole("listitem");
		expect(rows.map((row) => row.textContent)).toEqual([
			expect.stringContaining("Project alpha"),
			expect.stringContaining("Global announcements"),
		]);
		expect(within(browseList).getByText(/\u043f\u0440\u043e\u0435\u043a\u0442\u043d/i)).toBeInTheDocument();
		expect(within(browseList).getByText(/\u0433\u043b\u043e\u0431\u0430\u043b\u044c\u043d/i)).toBeInTheDocument();
		expect(within(browseList).queryByText("Must not appear")).not.toBeInTheDocument();
		expect(within(browseList).queryByText("Also must not appear")).not.toBeInTheDocument();
		for (const row of rows) {
			expect(within(row).queryByRole("link")).not.toBeInTheDocument();
		}
		const pathBeforeRowClick = window.location.pathname;
		fireEvent.click(rows[0]!);
		expect(window.location.pathname).toBe(pathBeforeRowClick);
		expect(api.get).toHaveBeenCalledTimes(1);
		expect(
			screen.queryByRole("button", {
				name: /\u043f\u0440\u0438\u0441\u043e\u0435\u0434\u0438\u043d|\u043f\u043e\u043a\u0438\u043d\u0443\u0442\u044c|\u0437\u0430\u043f\u0440\u043e\u0441\u0438\u0442\u044c \u0434\u043e\u0441\u0442\u0443\u043f|join|leave|request access/i,
			}),
		).not.toBeInTheDocument();
	});

	it("hides stale browse rows and removes the exact discovery cache when a refetch is forbidden", async () => {
		vi.mocked(api.get).mockResolvedValueOnce([globalChannel]);
		const view = renderWithProviders(<ChannelList />);

		fireEvent.click(
			screen.getByRole("button", { name: /\u041e\u0431\u0437\u043e\u0440 \u043a\u0430\u043d\u0430\u043b\u043e\u0432/i }),
		);
		const browseList = await screen.findByRole("list", {
			name: /\u0414\u043e\u0441\u0442\u0443\u043f\u043d\u044b\u0435 \u043a\u0430\u043d\u0430\u043b\u044b/i,
		});
		expect(within(browseList).getByText("Global announcements")).toBeInTheDocument();

		vi.mocked(api.get).mockRejectedValueOnce(
			{ code: 403, message: "forbidden" },
		);
		await act(async () => {
			await view.queryClient.refetchQueries({
				queryKey: discoverableChannelsQueryKey(user.id),
				exact: true,
			});
		});

		await waitFor(() => {
			expect(
				view.queryClient.getQueryData(discoverableChannelsQueryKey(user.id)),
			).toBeUndefined();
		});
		expect(api.get).toHaveBeenCalledTimes(2);
		expect(screen.queryByText("Global announcements")).not.toBeInTheDocument();
		expect(screen.getByRole("status")).toBeInTheDocument();
	});

	it("uses a neutral status for an empty browse result", async () => {
		vi.mocked(api.get).mockResolvedValue([]);

		renderWithProviders(<ChannelList />);
		fireEvent.click(
			screen.getByRole("button", { name: /\u041e\u0431\u0437\u043e\u0440 \u043a\u0430\u043d\u0430\u043b\u043e\u0432/i }),
		);

		const status = await screen.findByRole("status");
		expect(status).toHaveTextContent(/\u041d\u0435\u0442 \u0434\u043e\u0441\u0442\u0443\u043f\u043d\u044b\u0445 \u043a\u0430\u043d\u0430\u043b\u043e\u0432/i);
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});
});
