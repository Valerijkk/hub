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
import {
	threadDetailQueryKey,
	threadMessagesQueryKey,
} from "@/hooks/use-messages";
import { unreadQueryKey } from "@/hooks/use-unread";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

const api = vi.hoisted(() => ({
	getChannels: vi.fn(),
	getThread: vi.fn(),
	getThreadReplies: vi.fn(),
	markThreadRead: vi.fn(),
}));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getChannels: api.getChannels,
		getThread: api.getThread,
		getThreadReplies: api.getThreadReplies,
		markThreadRead: api.markThreadRead,
	};
});

vi.mock("@/components/messenger/message-item", () => ({
	MessageItem: ({ message }: { message: Message }) => (
		<div data-testid={`thread-message-${message.id}`}>{message.content_text}</div>
	),
}));

vi.mock("@/components/messenger/message-composer", () => ({
	MessageComposer: () => <div data-testid="thread-composer" />,
}));

vi.mock("@/components/messenger/thread-issue-links", () => ({
	ThreadIssueLinks: () => null,
}));

const CHANNEL_ID = "thread-read-channel";
const ROOT_ID = "thread-read-root";
const USER_ID = "thread-read-user";

const userA = {
	id: USER_ID,
	email: "thread-read@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};

const userB = {
	...userA,
	id: "thread-read-user-b",
	email: "thread-read-b@example.test",
};

type ObserverTarget = {
	observer: IntersectionObserverHarness;
	target: Element;
};

const observers: IntersectionObserverHarness[] = [];

class IntersectionObserverHarness {
	readonly root = null;
	readonly rootMargin = "0px";
	readonly thresholds = [1];
	readonly targets = new Set<Element>();

	constructor(
		readonly callback: IntersectionObserverCallback,
		readonly options: IntersectionObserverInit = {},
	) {
		observers.push(this);
	}

	observe(target: Element): void {
		this.targets.add(target);
	}

	unobserve(target: Element): void {
		this.targets.delete(target);
	}

	disconnect(): void {
		this.targets.clear();
	}

	takeRecords(): IntersectionObserverEntry[] {
		return [];
	}
}

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
		can_follow: false,
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
		name: "Thread receipts",
		is_readonly: false,
		capabilities: capabilities(),
		created_by: USER_ID,
		is_starred: false,
		created_at: "2026-09-05T00:00:00.000Z",
		updated_at: "2026-09-05T00:00:00.000Z",
		...overrides,
	};
}

function message(id: string, overrides: Partial<Message> = {}): Message {
	return {
		id,
		channel_id: CHANNEL_ID,
		sender_id: "thread-read-sender",
		content_text: `Reply ${id}`,
		reply_count: 0,
		created_at: "2026-09-05T00:00:00.000Z",
		...overrides,
	};
}

function reply(id: string, overrides: Partial<Message> = {}): Message {
	return message(id, { thread_id: ROOT_ID, ...overrides });
}

function thread(
	followMode?: string,
	overrides: Partial<ThreadResponse> = {},
): ThreadResponse {
	return {
		parent: message(ROOT_ID, { content_text: "Thread root" }),
		replies: [],
		...(followMode === undefined ? {} : { follow_mode: followMode }),
		...overrides,
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

function visibleEntry(
	target: Element,
	{ isIntersecting = true, intersectionRatio = 1 } = {},
): IntersectionObserverEntry {
	return {
		target,
		isIntersecting,
		intersectionRatio,
	} as unknown as IntersectionObserverEntry;
}

describe("ThreadPanel visible read-through receipts", () => {
	let queryClient: QueryClient;
	let focused = true;
	let visible = true;
	let initialVisibilityDescriptor: PropertyDescriptor | undefined;

	beforeEach(async () => {
		queryClient = new QueryClient({
			defaultOptions: {
				queries: { retry: false },
				mutations: { retry: false },
			},
		});
		observers.length = 0;
		focused = true;
		visible = true;
		initialVisibilityDescriptor = Object.getOwnPropertyDescriptor(
			document,
			"visibilityState",
		);
		Object.defineProperty(document, "visibilityState", {
			configurable: true,
			get: () => (visible ? "visible" : "hidden"),
		});
		vi.stubGlobal("IntersectionObserver", IntersectionObserverHarness);
		vi.spyOn(document, "hasFocus").mockImplementation(() => focused);
		api.getChannels.mockReset();
		api.getThread.mockReset();
		api.getThreadReplies.mockReset();
		api.markThreadRead.mockReset();
		api.getChannels.mockResolvedValue([channel()]);
		api.getThread.mockResolvedValue(thread("manual_follow"));
		api.getThreadReplies.mockResolvedValue([reply("reply-1")]);
		api.markThreadRead.mockResolvedValue({ advanced: true });
		await i18n.loadNamespaces("messenger");
		await i18n.changeLanguage("en");
		useAuthStore.setState({ user: userA });
		useMessengerStore.getState().reset();
	});

	afterEach(() => {
		cleanup();
		queryClient.clear();
		useAuthStore.setState({ user: null });
		useMessengerStore.getState().reset();
		vi.unstubAllGlobals();
		vi.restoreAllMocks();
		vi.clearAllMocks();
		if (initialVisibilityDescriptor) {
			Object.defineProperty(document, "visibilityState", initialVisibilityDescriptor);
		} else {
			delete (document as { visibilityState?: DocumentVisibilityState }).visibilityState;
		}
	});

	function TestProvider({ children }: { children: ReactNode }) {
		return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
	}

	function renderPanel({
		channelId = CHANNEL_ID,
		messageId = ROOT_ID,
	}: {
		channelId?: string;
		messageId?: string;
	} = {}) {
		return render(
			<ThreadPanel channelId={channelId} messageId={messageId} onClose={vi.fn()} />,
			{ wrapper: TestProvider },
		);
	}

	async function waitForReply(id = "reply-1"): Promise<void> {
		await screen.findByText(`Reply ${id}`);
		await waitFor(() => expect(api.getChannels).toHaveBeenCalledTimes(1));
		await waitFor(() => expect(api.getThread).toHaveBeenCalledTimes(1));
		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(1));
	}

	function observedTargetForReply(id: string): ObserverTarget | null {
		for (const observer of observers) {
			for (const target of observer.targets) {
				if (
					target instanceof HTMLElement &&
					target.dataset.threadReadMarker === "" &&
					target.dataset.messageId === id
				) {
					return { observer, target };
				}
			}
		}
		return null;
	}

	async function emitReply(
		id: string,
		options: { isIntersecting?: boolean; intersectionRatio?: number } = {},
	): Promise<void> {
		const observed = observedTargetForReply(id);
		if (!observed) return;
		await act(async () => {
			observed.observer.callback(
				[visibleEntry(observed.target, options)],
				observed.observer as unknown as IntersectionObserver,
			);
		});
	}

	async function emitReplies(ids: string[]): Promise<void> {
		const entries: Array<{ observer: IntersectionObserverHarness; entry: IntersectionObserverEntry }> = [];
		for (const id of ids) {
			const observed = observedTargetForReply(id);
			if (observed) {
				entries.push({
					observer: observed.observer,
					entry: visibleEntry(observed.target),
				});
			}
		}
		for (const observer of observers) {
			const observerEntries = entries
				.filter((entry) => entry.observer === observer)
				.map((entry) => entry.entry);
			if (observerEntries.length === 0) continue;
			await act(async () => {
				observer.callback(observerEntries, observer as unknown as IntersectionObserver);
			});
		}
	}

	function setDocumentVisibility(nextVisible: boolean): void {
		visible = nextVisible;
		document.dispatchEvent(new Event("visibilitychange"));
	}

	function setWindowFocus(nextFocused: boolean): void {
		focused = nextFocused;
		window.dispatchEvent(new Event(nextFocused ? "focus" : "blur"));
	}

	it("sends the exactly visible direct reply only after its end is observed", async () => {
		renderPanel();
		await waitForReply();

		expect(api.markThreadRead).not.toHaveBeenCalled();
		await emitReply("reply-1");

		await waitFor(() =>
			expect(api.markThreadRead).toHaveBeenCalledWith(
				CHANNEL_ID,
				ROOT_ID,
				"reply-1",
			),
		);
	});

	it("does not send eagerly, without an observer, while hidden, or while blurred", async () => {
		vi.stubGlobal("IntersectionObserver", undefined);
		renderPanel();
		await waitForReply();
		await emitReply("reply-1");
		expect(api.markThreadRead).not.toHaveBeenCalled();

		cleanup();
		queryClient.clear();
		api.getChannels.mockClear();
		api.getThread.mockClear();
		api.getThreadReplies.mockClear();
		vi.stubGlobal("IntersectionObserver", IntersectionObserverHarness);
		setDocumentVisibility(false);
		renderPanel();
		await waitForReply();
		await emitReply("reply-1");
		expect(api.markThreadRead).not.toHaveBeenCalled();

		cleanup();
		queryClient.clear();
		api.getChannels.mockClear();
		api.getThread.mockClear();
		api.getThreadReplies.mockClear();
		setDocumentVisibility(true);
		setWindowFocus(false);
		renderPanel();
		await waitForReply();
		await emitReply("reply-1");
		expect(api.markThreadRead).not.toHaveBeenCalled();
	});

	it("ignores partial and nonintersecting geometry before accepting a full witness", async () => {
		renderPanel();
		await waitForReply();

		await emitReply("reply-1", { intersectionRatio: 0.99 });
		await emitReply("reply-1", { isIntersecting: false, intersectionRatio: 1 });
		expect(api.markThreadRead).not.toHaveBeenCalled();

		await emitReply("reply-1");
		await waitFor(() => expect(api.markThreadRead).toHaveBeenCalledTimes(1));
	});

	it("fails closed when the direct-reply, follow, capability, or current-resource gate is invalid", async () => {
		const invalidReplies = [
			reply("wrong-channel", { channel_id: "other-channel" }),
			reply("wrong-thread", { thread_id: "other-root" }),
			reply("deleted", { deleted_at: "2026-09-05T00:00:01.000Z" }),
			reply("malformed-date", { created_at: "not-a-date" }),
			reply("eligible"),
		];
		api.getThreadReplies.mockResolvedValue(invalidReplies);
		renderPanel();
		await waitForReply("eligible");

		await emitReplies([
			"wrong-channel",
			"wrong-thread",
			"deleted",
			"malformed-date",
		]);
		expect(api.markThreadRead).not.toHaveBeenCalled();
		await emitReply("eligible");
		await waitFor(() =>
			expect(api.markThreadRead).toHaveBeenCalledWith(CHANNEL_ID, ROOT_ID, "eligible"),
		);

		cleanup();
		queryClient.clear();
		api.markThreadRead.mockClear();
		api.getChannels.mockClear();
		api.getThread.mockClear();
		api.getThreadReplies.mockClear();
		api.getChannels.mockResolvedValue([channel({ capabilities: capabilities({ can_read: false }) })]);
		renderPanel();
		await waitForReply("eligible");
		await emitReply("eligible");
		expect(api.markThreadRead).not.toHaveBeenCalled();

		cleanup();
		queryClient.clear();
		api.getChannels.mockClear();
		api.getThread.mockClear();
		api.getThreadReplies.mockClear();
		api.getChannels.mockResolvedValue([channel()]);
		api.getThread.mockResolvedValue(thread("manual_unfollow"));
		renderPanel();
		await waitForReply("eligible");
		await emitReply("eligible");
		expect(api.markThreadRead).not.toHaveBeenCalled();
	});

	it("does not receipt when the real public thread hook masks a page containing an empty reply ID", async () => {
		api.getThreadReplies.mockResolvedValue([reply("", { id: "" }), reply("eligible")]);
		renderPanel();
		await waitFor(() =>
			expect(
				queryClient.getQueryState(
					threadMessagesQueryKey(CHANNEL_ID, ROOT_ID, USER_ID),
				)?.status,
			).toBe("success"),
		);

		expect(screen.queryByText("Reply eligible")).not.toBeInTheDocument();
		await emitReply("eligible");
		expect(api.markThreadRead).not.toHaveBeenCalled();
	});

	it.each([
		["unknown mode", "auto_follow"],
		["absent mode", undefined],
		["missing membership", "manual_follow", capabilities({ is_member: false })],
		["missing read capability", "manual_follow", capabilities({ can_read: false })],
		[
			"missing unread capability",
			"manual_follow",
			capabilities({ can_view_unread: false }),
		],
	] as const)(
		"does not receipt with %s",
		async (_description, followMode, channelCapabilities = capabilities()) => {
			api.getChannels.mockResolvedValue([channel({ capabilities: channelCapabilities })]);
			api.getThread.mockResolvedValue(thread(followMode));
			renderPanel();
			await waitForReply();
			await emitReply("reply-1");
			expect(api.markThreadRead).not.toHaveBeenCalled();
		},
	);

	it("does not receipt when a truthy parent thread ID names another root", async () => {
		api.getThread.mockResolvedValue(
			thread("manual_follow", {
				parent: message(ROOT_ID, { thread_id: "other-root" }),
			}),
		);
		renderPanel();
		await waitForReply();
		await emitReply("reply-1");
		expect(api.markThreadRead).not.toHaveBeenCalled();
	});

	it("does not receipt after the current identity channel list no longer contains the thread channel", async () => {
		renderPanel();
		await waitForReply();
		api.getChannels.mockResolvedValueOnce([]);
		act(() => {
			void queryClient.invalidateQueries({
				queryKey: channelsQueryKey(USER_ID),
				exact: true,
			});
		});
		await waitFor(() =>
			expect(queryClient.getQueryState(channelsQueryKey(USER_ID))?.data).toEqual([]),
		);

		await emitReply("reply-1");
		expect(api.markThreadRead).not.toHaveBeenCalled();
	});

	it.each(["channels", "detail", "replies"] as const)(
		"does not receipt while the current %s resource is refetching",
		async (resource) => {
			const pending = deferred<Channel[] | ThreadResponse | Message[]>();
			let queryKey: readonly unknown[] = channelsQueryKey(USER_ID);
			let recovery: Channel[] | ThreadResponse | Message[] = [channel()];
			renderPanel();
			await waitForReply();

			switch (resource) {
				case "channels":
					api.getChannels.mockReturnValueOnce(pending.promise);
					break;
				case "detail":
					queryKey = threadDetailQueryKey(CHANNEL_ID, ROOT_ID, USER_ID);
					recovery = thread("manual_follow");
					api.getThread.mockReturnValueOnce(pending.promise);
					break;
				case "replies":
					queryKey = threadMessagesQueryKey(CHANNEL_ID, ROOT_ID, USER_ID);
					recovery = [reply("reply-1")];
					api.getThreadReplies.mockReturnValueOnce(pending.promise);
					break;
			}
			act(() => {
				void queryClient.invalidateQueries({ queryKey, exact: true });
			});
			await waitFor(() =>
				expect(queryClient.getQueryState(queryKey)?.fetchStatus).toBe("fetching"),
			);

			await emitReply("reply-1");
			expect(api.markThreadRead).not.toHaveBeenCalled();
			await act(async () => {
				pending.resolve(recovery);
				await pending.promise;
			});
		},
	);

	it.each(["channels", "detail", "replies"] as const)(
		"does not receipt after the current %s resource errors",
		async (resource) => {
			let queryKey: readonly unknown[] = channelsQueryKey(USER_ID);
			renderPanel();
			await waitForReply();

			switch (resource) {
				case "channels":
					api.getChannels.mockRejectedValueOnce(new Error("channels unavailable"));
					break;
				case "detail":
					queryKey = threadDetailQueryKey(CHANNEL_ID, ROOT_ID, USER_ID);
					api.getThread.mockRejectedValueOnce(new Error("thread unavailable"));
					break;
				case "replies":
					queryKey = threadMessagesQueryKey(CHANNEL_ID, ROOT_ID, USER_ID);
					api.getThreadReplies.mockRejectedValueOnce(new Error("replies unavailable"));
					break;
			}
			act(() => {
				void queryClient.invalidateQueries({ queryKey, exact: true });
			});
			await waitFor(() =>
				expect(queryClient.getQueryState(queryKey)?.status).toBe("error"),
			);

			await emitReply("reply-1");
			expect(api.markThreadRead).not.toHaveBeenCalled();
		},
	);

	it("uses the frozen viewport-clipped observer geometry when a live marker is observed", async () => {
		renderPanel();
		await waitForReply();
		const observer = observers[observers.length - 1];

		expect(observer).toBeDefined();
		expect(observer?.options.root ?? null).toBeNull();
		expect(observer?.options.threshold).toBe(1);
		expect([undefined, "0px"]).toContain(observer?.options.rootMargin);
	});

	it("clears unsent work through a reply refetch and requires a fresh full witness after recovery", async () => {
		const delayedReplies = deferred<Message[]>();
		renderPanel();
		await waitForReply();
		const previous = observedTargetForReply("reply-1");
		expect(previous).not.toBeNull();
		api.getThreadReplies.mockReturnValueOnce(delayedReplies.promise);
		act(() => {
			void queryClient.invalidateQueries({
				queryKey: threadMessagesQueryKey(CHANNEL_ID, ROOT_ID, USER_ID),
				exact: true,
			});
		});
		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(2));
		await emitReply("reply-1");
		expect(api.markThreadRead).not.toHaveBeenCalled();

		await act(async () => {
			delayedReplies.resolve([reply("reply-1")]);
			await delayedReplies.promise;
		});
		await waitFor(() => {
			const active = observedTargetForReply("reply-1");
			expect(active).not.toBeNull();
			expect(active?.observer).not.toBe(previous?.observer);
		});
		await emitReply("reply-1");
		await waitFor(() => expect(api.markThreadRead).toHaveBeenCalledTimes(1));
	});

	it("selects the greatest observed tuple, never a newer unobserved or DOM-neighbor reply", async () => {
		api.getThreadReplies.mockResolvedValue([
			reply("newer-unseen", { created_at: "2026-09-05T00:00:03.000Z" }),
			reply("z-tiebreak", { created_at: "2026-09-05T00:00:02.123Z" }),
			reply("a-tiebreak", { created_at: "2026-09-05T00:00:02.123Z" }),
			reply("older-observed", { created_at: "2026-09-05T00:00:01.999Z" }),
		]);
		renderPanel();
		await waitForReply("newer-unseen");

		await emitReplies(["older-observed", "a-tiebreak", "z-tiebreak"]);
		await waitFor(() =>
			expect(api.markThreadRead).toHaveBeenCalledWith(
				CHANNEL_ID,
				ROOT_ID,
				"z-tiebreak",
			),
		);
		expect(api.markThreadRead).not.toHaveBeenCalledWith(
			CHANNEL_ID,
			ROOT_ID,
			"newer-unseen",
		);
	});

	it("single-flights a pending receipt, coalesces one newer witness, and suppresses successful duplicates", async () => {
		const first = deferred<{ advanced: boolean }>();
		api.getThreadReplies.mockResolvedValue([
			reply("older", { created_at: "2026-09-05T00:00:01.000Z" }),
			reply("newer", { created_at: "2026-09-05T00:00:02.000Z" }),
		]);
		api.markThreadRead.mockReturnValueOnce(first.promise).mockResolvedValue({ advanced: true });
		renderPanel();
		await waitForReply("older");

		await emitReply("older");
		await waitFor(() =>
			expect(api.markThreadRead).toHaveBeenCalledWith(CHANNEL_ID, ROOT_ID, "older"),
		);
		await emitReply("newer");
		await emitReply("newer");
		expect(api.markThreadRead).toHaveBeenCalledTimes(1);

		await act(async () => {
			first.resolve({ advanced: true });
			await first.promise;
		});
		await waitFor(() =>
			expect(api.markThreadRead).toHaveBeenLastCalledWith(CHANNEL_ID, ROOT_ID, "newer"),
		);
		await emitReply("newer");
		await emitReply("older");
		expect(api.markThreadRead).toHaveBeenCalledTimes(2);
	});

	it("does not resend equal-or-older witnesses queued during a successful receipt", async () => {
		const receipt = deferred<{ advanced: boolean }>();
		const unreadRefresh = deferred<void>();
		const invalidate = vi
			.spyOn(queryClient, "invalidateQueries")
			.mockReturnValueOnce(unreadRefresh.promise);
		api.getThreadReplies.mockResolvedValue([
			reply("older", { created_at: "2026-09-05T00:00:01.000Z" }),
			reply("active", { created_at: "2026-09-05T00:00:02.000Z" }),
		]);
		api.markThreadRead.mockReturnValueOnce(receipt.promise);
		renderPanel();
		await waitForReply("older");

		await emitReply("active");
		await waitFor(() =>
			expect(api.markThreadRead).toHaveBeenCalledWith(
				CHANNEL_ID,
				ROOT_ID,
				"active",
			),
		);
		await emitReply("active");
		await emitReply("older");
		expect(api.markThreadRead).toHaveBeenCalledTimes(1);
		await act(async () => {
			receipt.resolve({ advanced: true });
			await receipt.promise;
		});
		await waitFor(() =>
			expect(invalidate).toHaveBeenCalledWith({
				queryKey: unreadQueryKey(USER_ID),
				exact: true,
			}),
		);
		await act(async () => {
			unreadRefresh.resolve();
			await unreadRefresh.promise;
			await Promise.resolve();
		});
		expect(api.markThreadRead).toHaveBeenCalledTimes(1);
	});

	it.each(["manual-unfollow detail", "unread-capability"] as const)(
		"requires a fresh current witness after %s loss while a receipt is in flight",
		async (gate) => {
			const receipt = deferred<{ advanced: boolean }>();
			const unreadRefresh = deferred<void>();
			const invalidate = vi
				.spyOn(queryClient, "invalidateQueries")
				.mockReturnValueOnce(unreadRefresh.promise);
			api.getThreadReplies.mockResolvedValue([
				reply("older", { created_at: "2026-09-05T00:00:01.000Z" }),
				reply("newer", { created_at: "2026-09-05T00:00:02.000Z" }),
			]);
			api.markThreadRead
				.mockReturnValueOnce(receipt.promise)
				.mockResolvedValue({ advanced: true });
			const view = renderPanel();
			await waitForReply("older");

			await emitReply("older");
			await waitFor(() =>
				expect(api.markThreadRead).toHaveBeenCalledWith(
					CHANNEL_ID,
					ROOT_ID,
					"older",
				),
			);
			await emitReply("newer");
			act(() => {
				if (gate === "manual-unfollow detail") {
					queryClient.setQueryData(
						threadDetailQueryKey(CHANNEL_ID, ROOT_ID, USER_ID),
						thread("manual_unfollow"),
					);
				} else {
					queryClient.setQueryData(channelsQueryKey(USER_ID), [
						channel({ capabilities: capabilities({ can_view_unread: false }) }),
					]);
				}
			});
			view.rerender(
				<ThreadPanel channelId={CHANNEL_ID} messageId={ROOT_ID} onClose={vi.fn()} />,
			);
			act(() => {
				queryClient.setQueryData(
					threadDetailQueryKey(CHANNEL_ID, ROOT_ID, USER_ID),
					thread("manual_follow"),
				);
				queryClient.setQueryData(channelsQueryKey(USER_ID), [channel()]);
			});
			view.rerender(
				<ThreadPanel channelId={CHANNEL_ID} messageId={ROOT_ID} onClose={vi.fn()} />,
			);
			await act(async () => {
				receipt.resolve({ advanced: true });
				await receipt.promise;
			});
			await waitFor(() =>
				expect(invalidate).toHaveBeenCalledWith({
					queryKey: unreadQueryKey(USER_ID),
					exact: true,
				}),
			);
			await act(async () => {
				unreadRefresh.resolve();
				await unreadRefresh.promise;
				await Promise.resolve();
			});
			expect(api.markThreadRead).toHaveBeenCalledTimes(1);
			await emitReply("newer");
			await waitFor(() =>
				expect(api.markThreadRead).toHaveBeenCalledWith(
					CHANNEL_ID,
					ROOT_ID,
					"newer",
				),
			);
			expect(api.markThreadRead).toHaveBeenCalledTimes(2);
		},
	);

	it("refreshes only this identity's unread truth even when the server says it was already advanced", async () => {
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		const unreadSnapshot = {
			channels: { [CHANNEL_ID]: 1 },
			dms: {},
			mentions: {},
			threads: { [ROOT_ID]: 2 },
			any_unread: true,
		};
		useMessengerStore.setState({ unreadSnapshot });
		api.markThreadRead.mockResolvedValue({ advanced: false });
		renderPanel();
		await waitForReply();
		invalidate.mockClear();

		await emitReply("reply-1");
		await waitFor(() =>
			expect(invalidate).toHaveBeenCalledWith({
				queryKey: unreadQueryKey(USER_ID),
				exact: true,
			}),
		);
		expect(invalidate.mock.calls).toEqual([
			[
				{
					queryKey: unreadQueryKey(USER_ID),
					exact: true,
				},
			],
		]);
		expect(useMessengerStore.getState().unreadSnapshot).toEqual(unreadSnapshot);
	});

	it("holds a queued newer witness until the successful receipt's unread refresh settles", async () => {
		const firstReceipt = deferred<{ advanced: boolean }>();
		const unreadRefresh = deferred<void>();
		const invalidate = vi
			.spyOn(queryClient, "invalidateQueries")
			.mockReturnValueOnce(unreadRefresh.promise);
		api.getThreadReplies.mockResolvedValue([
			reply("older", { created_at: "2026-09-05T00:00:01.000Z" }),
			reply("newer", { created_at: "2026-09-05T00:00:02.000Z" }),
		]);
		api.markThreadRead
			.mockReturnValueOnce(firstReceipt.promise)
			.mockResolvedValueOnce({ advanced: true });
		renderPanel();
		await waitForReply("older");

		await emitReply("older");
		await waitFor(() =>
			expect(api.markThreadRead).toHaveBeenCalledWith(CHANNEL_ID, ROOT_ID, "older"),
		);
		await act(async () => {
			firstReceipt.resolve({ advanced: true });
			await firstReceipt.promise;
		});
		await waitFor(() =>
			expect(invalidate).toHaveBeenCalledWith({
				queryKey: unreadQueryKey(USER_ID),
				exact: true,
			}),
		);

		await emitReply("newer");
		expect(api.markThreadRead).toHaveBeenCalledTimes(1);
		await act(async () => {
			unreadRefresh.resolve();
			await unreadRefresh.promise;
		});
		await waitFor(() =>
			expect(api.markThreadRead).toHaveBeenLastCalledWith(CHANNEL_ID, ROOT_ID, "newer"),
		);
	});

	it("keeps a fulfilled receipt deduplicated when the unread refresh rejects", async () => {
		const invalidate = vi
			.spyOn(queryClient, "invalidateQueries")
			.mockRejectedValueOnce(new Error("unread refresh unavailable"));
		renderPanel();
		await waitForReply();

		await emitReply("reply-1");
		await waitFor(() => expect(api.markThreadRead).toHaveBeenCalledTimes(1));
		await waitFor(() =>
			expect(invalidate).toHaveBeenCalledWith({
				queryKey: unreadQueryKey(USER_ID),
				exact: true,
			}),
		);
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: "Retry marking as read" }),
		).not.toBeInTheDocument();

		await emitReply("reply-1");
		expect(api.markThreadRead).toHaveBeenCalledTimes(1);
	});

	it("keeps the successful unread refresh when a background resource refetch begins after dispatch", async () => {
		const receipt = deferred<{ advanced: boolean }>();
		const channelsRefetch = deferred<Channel[]>();
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		api.markThreadRead.mockReturnValueOnce(receipt.promise);
		renderPanel();
		await waitForReply();
		invalidate.mockClear();

		await emitReply("reply-1");
		await waitFor(() => expect(api.markThreadRead).toHaveBeenCalledTimes(1));
		api.getChannels.mockReturnValueOnce(channelsRefetch.promise);
		act(() => {
			void queryClient.invalidateQueries({
				queryKey: channelsQueryKey(USER_ID),
				exact: true,
			});
		});
		await waitFor(() =>
			expect(queryClient.getQueryState(channelsQueryKey(USER_ID))?.fetchStatus).toBe(
				"fetching",
			),
		);

		await act(async () => {
			receipt.resolve({ advanced: true });
			await receipt.promise;
		});
		await waitFor(() =>
			expect(invalidate).toHaveBeenCalledWith({
				queryKey: unreadQueryKey(USER_ID),
				exact: true,
			}),
		);
		await act(async () => {
			channelsRefetch.resolve([channel()]);
			await channelsRefetch.promise;
		});
	});

	it("requires a fresh post-focus witness and ignores a late revoked-scope completion", async () => {
		const receipt = deferred<{ advanced: boolean }>();
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		api.markThreadRead.mockReturnValueOnce(receipt.promise);
		renderPanel();
		await waitForReply();
		invalidate.mockClear();
		const previous = observedTargetForReply("reply-1");
		expect(previous).not.toBeNull();
		act(() => {
			setWindowFocus(false);
			setWindowFocus(true);
		});
		expect(api.markThreadRead).not.toHaveBeenCalled();
		await waitFor(() => {
			const active = observedTargetForReply("reply-1");
			expect(active).not.toBeNull();
			expect(active?.observer).not.toBe(previous?.observer);
		});
		await emitReply("reply-1");
		await waitFor(() => expect(api.markThreadRead).toHaveBeenCalledTimes(1));
		act(() => {
			useMessengerStore.getState().revokeChannel(CHANNEL_ID);
		});
		await act(async () => {
			receipt.resolve({ advanced: true });
			await receipt.promise;
		});
		expect(invalidate).not.toHaveBeenCalled();
	});

	it("pauses on a localized failure and retries only after an explicit fresh visible witness", async () => {
		api.markThreadRead
			.mockRejectedValueOnce(new Error("receipt unavailable"))
			.mockResolvedValueOnce({ advanced: true });
		renderPanel();
		await waitForReply();
		await emitReply("reply-1");
		await waitFor(() => expect(api.markThreadRead).toHaveBeenCalledTimes(1));
		expect(await screen.findByRole("alert")).toHaveTextContent(
			"Unable to mark thread replies as read",
		);
		const retry = screen.getByRole("button", { name: "Retry marking as read" });
		expect(retry).toHaveAttribute("type", "button");

		await emitReply("reply-1");
		expect(api.markThreadRead).toHaveBeenCalledTimes(1);
		fireEvent.click(retry, { detail: 0 });
		await emitReply("reply-1");
		await waitFor(() => expect(api.markThreadRead).toHaveBeenCalledTimes(2));
	});

	it("renders the frozen Russian error and retry copy", async () => {
		await i18n.changeLanguage("ru");
		api.markThreadRead.mockRejectedValueOnce(new Error("receipt unavailable"));
		renderPanel();
		await waitForReply();
		await emitReply("reply-1");
		await waitFor(() => expect(api.markThreadRead).toHaveBeenCalledTimes(1));
		expect(await screen.findByRole("alert")).toHaveTextContent(
			"Не удалось отметить ответы прочитанными",
		);
		expect(
			screen.getByRole("button", { name: "Повторить отметку прочитанного" }),
		).toHaveAttribute("type", "button");
	});

	it("fences an A-B-A identity change so a completion from A cannot refresh A's unread query", async () => {
		const receipt = deferred<{ advanced: boolean }>();
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		api.markThreadRead.mockReturnValueOnce(receipt.promise);
		renderPanel();
		await waitForReply();
		invalidate.mockClear();
		await emitReply("reply-1");
		await waitFor(() => expect(api.markThreadRead).toHaveBeenCalledTimes(1));

		act(() => {
			useAuthStore.setState({ user: userB });
		});
		await waitFor(() =>
			expect(queryClient.getQueryState(channelsQueryKey(userB.id))?.status).toBe(
				"success",
			),
		);
		act(() => {
			useAuthStore.setState({ user: userA });
		});
		await act(async () => {
			receipt.resolve({ advanced: true });
			await receipt.promise;
		});
		expect(invalidate).not.toHaveBeenCalledWith({
			queryKey: unreadQueryKey(USER_ID),
			exact: true,
		});
	});

	it("fences a root A-B-A transition so an old root receipt cannot refresh unread truth", async () => {
		const receipt = deferred<{ advanced: boolean }>();
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		api.markThreadRead.mockReturnValueOnce(receipt.promise);
		const rendered = renderPanel();
		await waitForReply();
		invalidate.mockClear();
		await emitReply("reply-1");
		await waitFor(() => expect(api.markThreadRead).toHaveBeenCalledTimes(1));

		const otherRoot = "thread-read-root-b";
		api.getThread.mockResolvedValue(
			thread("manual_follow", {
				parent: message(otherRoot, { content_text: "Thread root B" }),
			}),
		);
		api.getThreadReplies.mockResolvedValue([
			message("reply-b", {
				thread_id: otherRoot,
				content_text: "Reply reply-b",
			}),
		]);
		rendered.rerender(
			<ThreadPanel channelId={CHANNEL_ID} messageId={otherRoot} onClose={vi.fn()} />,
		);
		expect(await screen.findByText("Reply reply-b")).toBeInTheDocument();

		api.getThread.mockResolvedValue(thread("manual_follow"));
		api.getThreadReplies.mockResolvedValue([reply("reply-1")]);
		rendered.rerender(
			<ThreadPanel channelId={CHANNEL_ID} messageId={ROOT_ID} onClose={vi.fn()} />,
		);
		expect(await screen.findByText("Reply reply-1")).toBeInTheDocument();
		await act(async () => {
			receipt.resolve({ advanced: true });
			await receipt.promise;
		});
		expect(invalidate).not.toHaveBeenCalledWith({
			queryKey: unreadQueryKey(USER_ID),
			exact: true,
		});
	});

	it.each(["success", "error"] as const)(
		"suppresses a late %s receipt completion after unmount",
		async (outcome) => {
			const receipt = deferred<{ advanced: boolean }>();
			const invalidate = vi.spyOn(queryClient, "invalidateQueries");
			api.markThreadRead.mockReturnValueOnce(receipt.promise);
			const rendered = renderPanel();
			await waitForReply();
			invalidate.mockClear();
			await emitReply("reply-1");
			await waitFor(() => expect(api.markThreadRead).toHaveBeenCalledTimes(1));
			rendered.unmount();

			await act(async () => {
				if (outcome === "success") {
					receipt.resolve({ advanced: true });
				} else {
					receipt.reject(new Error("receipt unavailable"));
				}
				await Promise.resolve();
			});
			expect(invalidate).not.toHaveBeenCalled();
		},
	);
});
