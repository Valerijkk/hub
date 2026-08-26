import { QueryClient } from "@tanstack/react-query";
import { act, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
	getBookmarks,
	removeBookmark,
	type Bookmark,
} from "@/api/messenger";
import { BookmarksPanel } from "@/components/messenger/bookmarks-panel";
import { useAuthStore } from "@/stores/auth-store";
import { renderWithProviders } from "../support/render";

const navigate = vi.fn();

vi.mock("@/api/messenger", () => ({
	getBookmarks: vi.fn(),
	removeBookmark: vi.fn(),
}));

vi.mock("@tanstack/react-router", () => ({
	useNavigate: () => navigate,
}));

const userA = {
	id: "user-a",
	email: "user-a@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};

const userB = {
	...userA,
	id: "user-b",
	email: "user-b@example.test",
};

const bookmarkA: Bookmark = {
	id: "bookmark-a",
	message_id: "message-a",
	channel_id: "channel-a",
	channel_name: "Private channel A",
	content_preview: "Only user A may see this bookmark",
	created_at: "2026-08-26T00:00:00Z",
};

const bookmarkB: Bookmark = {
	id: "bookmark-b",
	message_id: "message-b",
	channel_id: "channel-b",
	channel_name: "Private channel B",
	content_preview: "Only user B may see this bookmark",
	created_at: "2026-08-26T00:00:00Z",
};

function deferred<T>() {
	let resolve: (value: T) => void = () => undefined;
	const promise = new Promise<T>((done) => {
		resolve = done;
	});
	return { promise, resolve };
}

function createQueryClient() {
	return new QueryClient({
		defaultOptions: {
			queries: { retry: false },
			mutations: { retry: false },
		},
	});
}

describe("Phase 151 bookmark identity boundary", () => {
	const queryClients: QueryClient[] = [];

	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(getBookmarks).mockReset();
		vi.mocked(removeBookmark).mockReset();
		useAuthStore.setState({ user: userA });
	});

	afterEach(() => {
		for (const queryClient of queryClients.splice(0)) {
			queryClient.clear();
		}
		useAuthStore.setState({ user: null });
	});

	it("uses an identity-scoped cache and never renders a delayed A response after switching to B", async () => {
		const fromA = deferred<Bookmark[]>();
		const fromB = deferred<Bookmark[]>();
		vi.mocked(getBookmarks)
			.mockImplementationOnce(() => fromA.promise)
			.mockImplementationOnce(() => fromB.promise);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);

		renderWithProviders(<BookmarksPanel />, { queryClient });

		await waitFor(() => {
			expect(getBookmarks).toHaveBeenCalledTimes(1);
		});

		act(() => {
			useAuthStore.setState({ user: userB });
		});

		await waitFor(() => {
			expect(getBookmarks).toHaveBeenCalledTimes(2);
		});

		await act(async () => {
			fromA.resolve([bookmarkA]);
			await fromA.promise;
		});

		expect(screen.queryByText(bookmarkA.channel_name!)).not.toBeInTheDocument();

		await act(async () => {
			fromB.resolve([bookmarkB]);
			await fromB.promise;
		});

		expect(await screen.findByText(bookmarkB.channel_name!)).toBeInTheDocument();
		expect(screen.queryByText(bookmarkA.channel_name!)).not.toBeInTheDocument();
		expect(
			queryClient.getQueryData(["messenger", "bookmarks", userA.id]),
		).toEqual([bookmarkA]);
		expect(
			queryClient.getQueryData(["messenger", "bookmarks", userB.id]),
		).toEqual([bookmarkB]);
	});

	it("fails closed without an authenticated identity, even when the old generic cache has private rows", async () => {
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		queryClient.setQueryData(["messenger", "bookmarks"], [bookmarkA]);
		useAuthStore.setState({ user: null });

		renderWithProviders(<BookmarksPanel />, { queryClient });

		await act(async () => {
			await Promise.resolve();
		});

		expect(getBookmarks).not.toHaveBeenCalled();
		expect(screen.queryByText(bookmarkA.channel_name!)).not.toBeInTheDocument();
	});


	it("never renders private rows from the disabled no-user cache", async () => {
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		queryClient.setQueryData(
			["messenger", "bookmarks", "__no-user__"],
			[bookmarkA],
		);
		useAuthStore.setState({ user: null });

		renderWithProviders(<BookmarksPanel />, { queryClient });

		await act(async () => {
			await Promise.resolve();
		});

		expect(getBookmarks).not.toHaveBeenCalled();
		expect(screen.queryByText(bookmarkA.channel_name!)).not.toBeInTheDocument();
		expect(screen.queryByText(bookmarkA.content_preview!)).not.toBeInTheDocument();
	});

	it("does not refetch or replace B bookmark state when a deferred removal started by A completes", async () => {
		const removeFromA = deferred<void>();
		vi.mocked(getBookmarks)
			.mockResolvedValueOnce([bookmarkA])
			.mockResolvedValueOnce([bookmarkB]);
		vi.mocked(removeBookmark).mockImplementationOnce(() => removeFromA.promise);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);

		renderWithProviders(<BookmarksPanel />, { queryClient });

		expect(await screen.findByText(bookmarkA.channel_name!)).toBeInTheDocument();
		fireEvent.click(
			screen.getByRole("button", {
				name: /\u0423\u0434\u0430\u043b\u0438\u0442\u044c \u0438\u0437 \u0437\u0430\u043a\u043b\u0430\u0434\u043e\u043a/i,
			}),
		);
		await waitFor(() => {
			expect(removeBookmark).toHaveBeenCalledWith(bookmarkA.id);
		});

		act(() => {
			useAuthStore.setState({ user: userB });
		});

		expect(await screen.findByText(bookmarkB.channel_name!)).toBeInTheDocument();
		expect(getBookmarks).toHaveBeenCalledTimes(2);

		await act(async () => {
			removeFromA.resolve(undefined);
			await removeFromA.promise;
			await Promise.resolve();
		});

		expect(getBookmarks).toHaveBeenCalledTimes(2);
		expect(screen.getByText(bookmarkB.channel_name!)).toBeInTheDocument();
		expect(screen.queryByText(bookmarkA.channel_name!)).not.toBeInTheDocument();
	});
});
