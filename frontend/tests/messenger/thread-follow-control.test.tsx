import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
	act,
	cleanup,
	fireEvent,
	render,
	screen,
	waitFor,
} from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import i18n from "@/i18n";
import type { Channel, Message, ThreadResponse } from "@/api/messenger";
import { ThreadPanel } from "@/components/messenger/thread-panel";
import { channelsQueryKey } from "@/hooks/use-channels";
import { threadDetailQueryKey } from "@/hooks/use-messages";
import { unreadQueryKey } from "@/hooks/use-unread";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

const api = vi.hoisted(() => ({
	getChannels: vi.fn(),
	getThread: vi.fn(),
	getThreadReplies: vi.fn(),
	followThread: vi.fn(),
	unfollowThread: vi.fn(),
}));

const toasts = vi.hoisted(() => ({
	apiError: vi.fn(),
}));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getChannels: api.getChannels,
		getThread: api.getThread,
		getThreadReplies: api.getThreadReplies,
		followThread: api.followThread,
		unfollowThread: api.unfollowThread,
	};
});

vi.mock("@/lib/format-api-error", () => ({
	toastApiError: toasts.apiError,
}));

vi.mock("@/components/messenger/message-item", () => ({
	MessageItem: ({ message }: { message: Message }) => (
		<div data-testid={"thread-message-" + message.id}>{message.content_text}</div>
	),
}));

vi.mock("@/components/messenger/message-composer", () => ({
	MessageComposer: () => <div data-testid="thread-composer" />,
}));

vi.mock("@/components/messenger/thread-issue-links", () => ({
	ThreadIssueLinks: () => null,
}));

const CHANNEL_ID = "channel-1";
const ROOT_ID = "root-1";
const OTHER_ROOT_ID = "root-2";

const userA = {
	id: "thread-follow-user-a",
	email: "thread-follow-user-a@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};

const userB = {
	...userA,
	id: "thread-follow-user-b",
	email: "thread-follow-user-b@example.test",
};

function capabilities(
	overrides: Partial<NonNullable<Channel["capabilities"]>> = {},
): NonNullable<Channel["capabilities"]> {
	return {
		is_member: true,
		has_active_project_membership: true,
		can_read: true,
		can_post: true,
		can_reply: true,
		can_join: false,
		can_follow: true,
		can_unfollow: false,
		can_view_unread: true,
		can_type: true,
		can_edit: false,
		can_forward: false,
		can_react: true,
		can_delete_message: false,
		can_pin: false,
		can_list_pins: true,
		can_bookmark: true,
		can_promote: false,
		...overrides,
	};
}

function channel(overrides: Partial<Channel> = {}): Channel {
	return {
		id: CHANNEL_ID,
		channel_type: "public",
		name: "General",
		is_readonly: false,
		capabilities: capabilities(),
		created_by: userA.id,
		is_starred: false,
		created_at: "2026-09-05T00:00:00Z",
		updated_at: "2026-09-05T00:00:00Z",
		...overrides,
	};
}

function message(id = ROOT_ID): Message {
	return {
		id,
		channel_id: CHANNEL_ID,
		sender_id: "sender-1",
		content_text: "Root message",
		reply_count: 0,
		created_at: "2026-09-05T00:00:00Z",
	};
}

function detail(followMode?: string, rootId = ROOT_ID): ThreadResponse {
	return {
		parent: message(rootId),
		replies: [],
		...(followMode === undefined ? {} : { follow_mode: followMode }),
	};
}

function deferred<T>() {
	let resolve: (value: T) => void = () => undefined;
	let reject: (reason?: unknown) => void = () => undefined;
	const promise = new Promise<T>((done, fail) => {
		resolve = done;
		reject = fail;
	});
	return { promise, resolve, reject };
}

function threadTestId(rootId = ROOT_ID): string {
	return "thread-message-" + rootId;
}

describe("ThreadPanel follow control", () => {
	let queryClient: QueryClient;

	beforeEach(async () => {
		queryClient = new QueryClient({
			defaultOptions: {
				queries: { retry: false },
				mutations: { retry: false },
			},
		});
		api.getChannels.mockReset();
		api.getThread.mockReset();
		api.getThreadReplies.mockReset();
		api.followThread.mockReset();
		api.unfollowThread.mockReset();
		toasts.apiError.mockReset();
		await i18n.loadNamespaces("messenger");
		await i18n.changeLanguage("en");
		useAuthStore.setState({ user: userA });
		useMessengerStore.getState().reset();
		api.getChannels.mockResolvedValue([channel()]);
		api.getThread.mockResolvedValue(detail("manual_unfollow"));
		api.getThreadReplies.mockResolvedValue([]);
		api.followThread.mockResolvedValue({
			changed: true,
			mode: "manual_follow",
		});
		api.unfollowThread.mockResolvedValue({
			changed: true,
			mode: "manual_unfollow",
		});
	});

	afterEach(() => {
		cleanup();
		queryClient.clear();
		useAuthStore.setState({ user: null });
		useMessengerStore.getState().reset();
		vi.clearAllMocks();
	});

	function TestProvider({ children }: { children: ReactNode }) {
		return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
	}

	function renderPanel({
		channelId = CHANNEL_ID,
		messageId = ROOT_ID,
		onClose = vi.fn(),
	}: {
		channelId?: string;
		messageId?: string;
		onClose?: () => void;
	} = {}) {
		return render(
			<ThreadPanel
				channelId={channelId}
				messageId={messageId}
				onClose={onClose}
			/>,
			{ wrapper: TestProvider },
		);
	}

	async function waitForChannelQuery(
		status: "success" | "error",
		userId = useAuthStore.getState().user?.id,
	): Promise<void> {
		if (!userId) throw new Error("expected an authenticated user");
		await waitFor(() => {
			expect(queryClient.getQueryState(channelsQueryKey(userId))?.status).toBe(
				status,
			);
		});
	}

	async function waitForThreadQuery(
		status: "success" | "error",
		rootId = ROOT_ID,
		userId = useAuthStore.getState().user?.id,
	): Promise<void> {
		if (!userId) throw new Error("expected an authenticated user");
		await waitFor(() => {
			expect(
				queryClient.getQueryState(
					threadDetailQueryKey(CHANNEL_ID, rootId, userId),
				)?.status,
			).toBe(status);
		});
	}

	async function actionButton(
		name: string,
		rootId = ROOT_ID,
	): Promise<HTMLButtonElement> {
		const userId = useAuthStore.getState().user?.id;
		if (!userId) throw new Error("expected an authenticated user");
		await waitForChannelQuery("success", userId);
		await waitForThreadQuery("success", rootId, userId);
		await waitFor(() => {
			const channelQuery = queryClient.getQueryState(channelsQueryKey(userId));
			const threadQuery = queryClient.getQueryState(
				threadDetailQueryKey(CHANNEL_ID, rootId, userId),
			);
			expect(channelQuery?.fetchStatus).toBe("idle");
			expect(threadQuery?.fetchStatus).toBe("idle");
		});
		return (await screen.findByRole("button", { name })) as HTMLButtonElement;
	}

	function expectNoAction(): void {
		expect(
			screen.queryByRole("button", { name: /^(Follow|Unfollow)$/ }),
		).not.toBeInTheDocument();
	}

	it.each([
		["an absent mode", undefined, "Follow", false],
		["manual_unfollow", "manual_unfollow", "Follow", false],
		["auto", "auto", "Unfollow", true],
		["manual_follow", "manual_follow", "Unfollow", true],
	] as const)(
		"renders %s as the only enabled action for %s",
		async (_description, followMode, label, isUnfollow) => {
			api.getChannels.mockResolvedValue([
				channel({
					capabilities: capabilities(
						isUnfollow
							? { can_follow: false, can_unfollow: true }
							: { can_follow: true, can_unfollow: false },
					),
				}),
			]);
			api.getThread.mockResolvedValue(detail(followMode));
			renderPanel();

			const action = await actionButton(label);
			expect(action).toBeEnabled();
			expect(api.getChannels).toHaveBeenCalledTimes(1);
			expect(
				screen.queryByRole("button", {
					name: label === "Follow" ? "Unfollow" : "Follow",
				}),
			).not.toBeInTheDocument();
		},
	);

	it("renders the Russian follow label from the Messenger locale", async () => {
		await i18n.changeLanguage("ru");
		renderPanel();
		expect(await actionButton("Подписаться")).toBeEnabled();
	});

	it.each(["auto_follow", "", " manual_follow "])(
		"fails closed for the non-contract follow mode %j",
		async (followMode) => {
			api.getChannels.mockResolvedValue([
				channel({
					capabilities: capabilities({
						can_follow: true,
						can_unfollow: true,
					}),
				}),
			]);
			api.getThread.mockResolvedValue(detail(followMode));
			renderPanel();

			await waitForChannelQuery("success");
			await waitForThreadQuery("success");
			expectNoAction();
			expect(api.followThread).not.toHaveBeenCalled();
			expect(api.unfollowThread).not.toHaveBeenCalled();
		},
	);

	it.each([
		[
			"missing capabilities",
			"manual_unfollow",
			undefined as Channel["capabilities"],
		],
		[
			"missing follow permission without falling through to unfollow",
			"manual_unfollow",
			capabilities({ can_follow: false, can_unfollow: true }),
		],
		[
			"missing unfollow permission without falling through to follow",
			"manual_follow",
			capabilities({ can_follow: true, can_unfollow: false }),
		],
	] as const)("fails closed for %s", async (_description, followMode, channelCapabilities) => {
		api.getChannels.mockResolvedValue([channel({ capabilities: channelCapabilities })]);
		api.getThread.mockResolvedValue(detail(followMode));
		renderPanel();

		await waitForChannelQuery("success");
		await waitForThreadQuery("success");
		expectNoAction();
	});

	it("denies actions for an absent channel and an absent authenticated user", async () => {
		api.getChannels.mockResolvedValue([]);
		renderPanel();
		await waitForChannelQuery("success");
		await waitForThreadQuery("success");
		expectNoAction();

		cleanup();
		queryClient.clear();
		api.getChannels.mockClear();
		api.getThread.mockClear();
		useAuthStore.setState({ user: null });
		renderPanel();
		expectNoAction();
		expect(api.getChannels).not.toHaveBeenCalled();
		expect(api.getThread).not.toHaveBeenCalled();
	});

	it("removes the action immediately after channel revocation", async () => {
		renderPanel();
		await actionButton("Follow");
		act(() => {
			useMessengerStore.getState().revokeChannel(CHANNEL_ID);
		});
		expectNoAction();
		expect(api.followThread).not.toHaveBeenCalled();
		expect(api.unfollowThread).not.toHaveBeenCalled();
	});

	it("denies actions while the channel list is loading or has failed", async () => {
		const loadingChannels = deferred<Channel[]>();
		api.getChannels.mockReturnValueOnce(loadingChannels.promise);
		renderPanel();
		await waitFor(() => {
			expect(
				queryClient.getQueryState(channelsQueryKey(userA.id))?.fetchStatus,
			).toBe("fetching");
		});
		expectNoAction();

		cleanup();
		queryClient.clear();
		api.getChannels.mockReset();
		api.getThread.mockClear();
		api.getChannels.mockRejectedValueOnce(new Error("channels unavailable"));
		api.getThread.mockResolvedValue(detail("manual_unfollow"));
		renderPanel();
		await waitForChannelQuery("error");
		expectNoAction();
	});

	it("denies actions when the successful thread detail is unavailable", async () => {
		api.getThread.mockRejectedValueOnce(new Error("thread unavailable"));
		renderPanel();
		await waitForThreadQuery("error");
		expectNoAction();
	});

	it("keeps the same action node disabled during authorized channel or detail refetches", async () => {
		renderPanel();
		const follow = await actionButton("Follow");
		const detailRefetch = deferred<ThreadResponse>();
		api.getThread.mockImplementationOnce(() => detailRefetch.promise);
		act(() => {
			void queryClient.invalidateQueries({
				queryKey: threadDetailQueryKey(CHANNEL_ID, ROOT_ID, userA.id),
				exact: true,
			});
		});
		await waitFor(() => expect(api.getThread).toHaveBeenCalledTimes(2));
		expect(screen.getByRole("button", { name: "Follow" })).toBe(follow);
		expect(follow).toBeDisabled();
		fireEvent.click(follow);
		expect(api.followThread).not.toHaveBeenCalled();
		await act(async () => {
			detailRefetch.resolve(detail("manual_unfollow"));
			await detailRefetch.promise;
		});

		const channelRefetch = deferred<Channel[]>();
		api.getChannels.mockImplementationOnce(() => channelRefetch.promise);
		act(() => {
			void queryClient.invalidateQueries({
				queryKey: channelsQueryKey(userA.id),
				exact: true,
			});
		});
		await waitFor(() => expect(api.getChannels).toHaveBeenCalledTimes(2));
		expect(screen.getByRole("button", { name: "Follow" })).toBe(follow);
		expect(follow).toBeDisabled();
		fireEvent.click(follow);
		expect(api.followThread).not.toHaveBeenCalled();
		await act(async () => {
			channelRefetch.resolve([channel()]);
			await channelRefetch.promise;
		});
	});

	it("invalidates exact detail and unread after a dispatched follow races a channel refetch", async () => {
		const pendingFollow = deferred<{ changed: boolean; mode: string }>();
		const pendingChannels = deferred<Channel[]>();
		const pendingDetail = deferred<ThreadResponse>();
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		api.followThread.mockReturnValueOnce(pendingFollow.promise);
		renderPanel();
		const follow = await actionButton("Follow");

		fireEvent.click(follow);
		await waitFor(() => expect(api.followThread).toHaveBeenCalledTimes(1));

		api.getChannels.mockImplementationOnce(() => pendingChannels.promise);
		act(() => {
			void queryClient.invalidateQueries({
				queryKey: channelsQueryKey(userA.id),
				exact: true,
			});
		});
		await waitFor(() => {
			expect(
				queryClient.getQueryState(channelsQueryKey(userA.id))?.fetchStatus,
			).toBe("fetching");
		});
		expect(screen.getByRole("button", { name: "Follow" })).toBe(follow);
		expect(follow).toHaveAttribute("aria-busy", "true");
		invalidate.mockClear();
		api.getThread.mockImplementationOnce(() => pendingDetail.promise);

		await act(async () => {
			pendingFollow.resolve({ changed: true, mode: "manual_follow" });
			await pendingFollow.promise;
		});

		await waitFor(() =>
			expect(invalidate).toHaveBeenCalledWith({
				queryKey: threadDetailQueryKey(CHANNEL_ID, ROOT_ID, userA.id),
				exact: true,
			}),
		);
		expect(screen.getByRole("button", { name: "Follow" })).toBe(follow);
		expect(follow).toHaveAttribute("aria-busy", "true");

		await act(async () => {
			pendingChannels.resolve([channel()]);
			await pendingChannels.promise;
		});
		await waitFor(() => expect(api.getThread).toHaveBeenCalledTimes(2));
		expect(screen.getByRole("button", { name: "Follow" })).toBe(follow);
		expect(follow).toHaveAttribute("aria-busy", "true");

		await act(async () => {
			pendingDetail.resolve(detail("manual_unfollow"));
			await pendingDetail.promise;
		});
		await waitFor(() =>
			expect(invalidate.mock.calls).toEqual([
				[
					{
						queryKey: threadDetailQueryKey(CHANNEL_ID, ROOT_ID, userA.id),
						exact: true,
					},
				],
				[
					{
						queryKey: unreadQueryKey(userA.id),
						exact: true,
					},
				],
			]),
		);
		expect(follow).toBeEnabled();
	});

	it("uses a native keyboard-activatable Follow button, exact API ids, and only authoritative invalidations", async () => {
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		const setQueryData = vi.spyOn(queryClient, "setQueryData");
		const unreadSnapshot = {
			channels: { [CHANNEL_ID]: 2 },
			dms: {},
			mentions: {},
			threads: { [ROOT_ID]: 7 },
			any_unread: true,
		};
		useMessengerStore.setState({ unreadSnapshot });
		renderPanel();
		const follow = await actionButton("Follow");
		invalidate.mockClear();
		setQueryData.mockClear();

		expect(follow.tagName).toBe("BUTTON");
		expect(follow).toHaveAttribute("type", "button");
		fireEvent.click(follow, { detail: 0 });
		await waitFor(() =>
			expect(api.followThread).toHaveBeenCalledWith(CHANNEL_ID, ROOT_ID),
		);
		await waitFor(() => expect(invalidate).toHaveBeenCalledTimes(2));
		expect(invalidate.mock.calls).toEqual(
			expect.arrayContaining([
				[
					{
						queryKey: threadDetailQueryKey(CHANNEL_ID, ROOT_ID, userA.id),
						exact: true,
					},
				],
				[
					{
						queryKey: unreadQueryKey(userA.id),
						exact: true,
					},
				],
			]),
		);
		expect(setQueryData).not.toHaveBeenCalled();
		expect(useMessengerStore.getState().unreadSnapshot).toEqual(unreadSnapshot);
	});

	it("single-flights duplicate clicks and stays busy through its owned detail refetch", async () => {
		const pendingFollow = deferred<{ changed: boolean; mode: string }>();
		const pendingDetail = deferred<ThreadResponse>();
		api.getChannels.mockResolvedValue([
			channel({
				capabilities: capabilities({ can_follow: true, can_unfollow: true }),
			}),
		]);
		api.followThread.mockReturnValueOnce(pendingFollow.promise);
		renderPanel();
		const follow = await actionButton("Follow");
		api.getThread.mockImplementationOnce(() => pendingDetail.promise);

		fireEvent.click(follow);
		fireEvent.click(follow);
		await waitFor(() => expect(api.followThread).toHaveBeenCalledTimes(1));
		expect(screen.getByRole("button", { name: "Follow" })).toBe(follow);
		expect(follow).toBeDisabled();
		expect(follow).toHaveAttribute("aria-busy", "true");
		expect(screen.getByRole("status")).toBeInTheDocument();

		await act(async () => {
			pendingFollow.resolve({ changed: true, mode: "manual_follow" });
			await pendingFollow.promise;
		});
		await waitFor(() => expect(api.getThread).toHaveBeenCalledTimes(2));
		expect(screen.getByRole("button", { name: "Follow" })).toBe(follow);
		expect(follow).toBeDisabled();
		expect(follow).toHaveAttribute("aria-busy", "true");

		await act(async () => {
			pendingDetail.resolve(detail("manual_follow"));
			await pendingDetail.promise;
		});
		expect(await actionButton("Unfollow")).toBeEnabled();
	});

	it("uses the exact Unfollow API with the active channel and root ids", async () => {
		api.getChannels.mockResolvedValue([
			channel({
				capabilities: capabilities({ can_follow: false, can_unfollow: true }),
			}),
		]);
		api.getThread.mockResolvedValue(detail("manual_follow"));
		renderPanel();
		const unfollow = await actionButton("Unfollow");
		fireEvent.click(unfollow);

		await waitFor(() =>
			expect(api.unfollowThread).toHaveBeenCalledWith(CHANNEL_ID, ROOT_ID),
		);
	});

	it("refetches both changed and changed=false responses without optimistic local mode state", async () => {
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		api.followThread.mockResolvedValueOnce({
			changed: false,
			mode: "manual_unfollow",
		});
		renderPanel();
		const follow = await actionButton("Follow");
		invalidate.mockClear();
		fireEvent.click(follow);

		await waitFor(() => expect(api.followThread).toHaveBeenCalledTimes(1));
		await waitFor(() => expect(invalidate).toHaveBeenCalledTimes(2));
		await waitFor(() => expect(api.getThread).toHaveBeenCalledTimes(2));
		expect(screen.getByRole("button", { name: "Follow" })).toBeEnabled();
	});

	it("uses the server mode only after the authoritative refetch", async () => {
		const pendingFollow = deferred<{ changed: boolean; mode: string }>();
		api.getChannels.mockResolvedValue([
			channel({
				capabilities: capabilities({ can_follow: true, can_unfollow: true }),
			}),
		]);
		api.getThread.mockReset();
		api.getThread
			.mockResolvedValueOnce(detail("manual_unfollow"))
			.mockResolvedValueOnce(detail("manual_follow"));
		api.followThread.mockReturnValueOnce(pendingFollow.promise);
		renderPanel();
		const follow = await actionButton("Follow");
		fireEvent.click(follow);

		expect(screen.getByRole("button", { name: "Follow" })).toBe(follow);
		expect(follow).toHaveAttribute("aria-busy", "true");
		expect(screen.queryByRole("button", { name: "Unfollow" })).not.toBeInTheDocument();
		await act(async () => {
			pendingFollow.resolve({ changed: true, mode: "manual_follow" });
			await pendingFollow.promise;
		});
		await waitFor(() => expect(api.getThread).toHaveBeenCalledTimes(2));
		expect(await actionButton("Unfollow")).toBeEnabled();
	});

	it.each([
		["en", "Follow", "Unable to update thread subscription"],
		["ru", "Подписаться", "Не удалось изменить подписку на тред"],
	] as const)(
		"reports a localized API error in %s without changing state and permits an explicit retry",
		async (language, label, fallback) => {
			await i18n.changeLanguage(language);
			const failure = new Error("follow unavailable");
			const invalidate = vi.spyOn(queryClient, "invalidateQueries");
			api.followThread
				.mockRejectedValueOnce(failure)
				.mockResolvedValueOnce({ changed: true, mode: "manual_follow" });
			renderPanel();
			const follow = await actionButton(label);
			invalidate.mockClear();
			fireEvent.click(follow);

			await waitFor(() =>
				expect(toasts.apiError).toHaveBeenCalledWith(failure, fallback),
			);
			expect(i18n.t("messenger:threadFollow.error")).toBe(fallback);
			expect(invalidate).not.toHaveBeenCalled();
			expect(screen.getByRole("button", { name: label })).toBeEnabled();
			expect(api.followThread).toHaveBeenCalledTimes(1);
			fireEvent.click(screen.getByRole("button", { name: label }));
			await waitFor(() => expect(api.followThread).toHaveBeenCalledTimes(2));
			await waitFor(() => expect(invalidate).toHaveBeenCalledTimes(2));
		},
	);

	it("drops a successful completion after an identity change without cache or toast effects", async () => {
		const pendingFollow = deferred<{ changed: boolean; mode: string }>();
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		api.followThread.mockReturnValueOnce(pendingFollow.promise);
		renderPanel();
		const follow = await actionButton("Follow");
		invalidate.mockClear();
		fireEvent.click(follow);
		await waitFor(() => expect(api.followThread).toHaveBeenCalledTimes(1));
		act(() => {
			useAuthStore.setState({ user: userB });
		});
		await waitForChannelQuery("success", userB.id);

		await act(async () => {
			pendingFollow.resolve({ changed: true, mode: "manual_follow" });
			await pendingFollow.promise;
		});
		expect(invalidate).not.toHaveBeenCalled();
		expect(toasts.apiError).not.toHaveBeenCalled();
	});

	it("drops an error completion after revocation without a cache update or toast", async () => {
		const pendingFollow = deferred<{ changed: boolean; mode: string }>();
		const failure = new Error("revoked while following");
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		api.followThread.mockReturnValueOnce(pendingFollow.promise);
		renderPanel();
		const follow = await actionButton("Follow");
		invalidate.mockClear();
		fireEvent.click(follow);
		await waitFor(() => expect(api.followThread).toHaveBeenCalledTimes(1));
		act(() => {
			useMessengerStore.getState().revokeChannel(CHANNEL_ID);
		});

		await act(async () => {
			pendingFollow.reject(failure);
			await pendingFollow.promise.catch(() => undefined);
		});
		expect(invalidate).not.toHaveBeenCalled();
		expect(toasts.apiError).not.toHaveBeenCalled();
	});

	it("drops a completion for a root that was replaced while the request was pending", async () => {
		const pendingFollow = deferred<{ changed: boolean; mode: string }>();
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		api.getThread.mockReset();
		api.getThread
			.mockResolvedValueOnce(detail("manual_unfollow", ROOT_ID))
			.mockResolvedValueOnce(detail("manual_unfollow", OTHER_ROOT_ID));
		api.followThread.mockReturnValueOnce(pendingFollow.promise);
		const view = renderPanel();
		const follow = await actionButton("Follow");
		invalidate.mockClear();
		fireEvent.click(follow);
		await waitFor(() => expect(api.followThread).toHaveBeenCalledTimes(1));
		view.rerender(
			<ThreadPanel
				channelId={CHANNEL_ID}
				messageId={OTHER_ROOT_ID}
				onClose={vi.fn()}
			/>,
		);
		await waitForThreadQuery("success", OTHER_ROOT_ID);

		await act(async () => {
			pendingFollow.resolve({ changed: true, mode: "manual_follow" });
			await pendingFollow.promise;
		});
		expect(invalidate).not.toHaveBeenCalled();
		expect(toasts.apiError).not.toHaveBeenCalled();
	});

	it("treats an A-to-B-to-A identity transition as stale for the original completion", async () => {
		const pendingFollow = deferred<{ changed: boolean; mode: string }>();
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		api.followThread.mockReturnValueOnce(pendingFollow.promise);
		renderPanel();
		const follow = await actionButton("Follow");
		invalidate.mockClear();
		fireEvent.click(follow);
		await waitFor(() => expect(api.followThread).toHaveBeenCalledTimes(1));
		act(() => {
			useAuthStore.setState({ user: userB });
		});
		await waitForChannelQuery("success", userB.id);
		act(() => {
			useAuthStore.setState({ user: userA });
		});
		await act(async () => {
			await Promise.resolve();
			pendingFollow.resolve({ changed: true, mode: "manual_follow" });
			await pendingFollow.promise;
		});
		expect(invalidate).not.toHaveBeenCalled();
		expect(toasts.apiError).not.toHaveBeenCalled();
	});

	it("preserves the close button and returns focus to the opener", async () => {
		const opener = document.createElement("button");
		document.body.append(opener);
		opener.focus();
		const onClose = vi.fn();
		renderPanel({ onClose });
		const close = screen.getByRole("button", { name: "Закрыть тред" });
		await waitFor(() => expect(close).toHaveFocus());
		fireEvent.click(close);

		expect(onClose).toHaveBeenCalledTimes(1);
		expect(opener).toHaveFocus();
		opener.remove();
	});
});
