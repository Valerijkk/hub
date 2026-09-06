import {
	focusManager,
	onlineManager,
	QueryClient,
} from "@tanstack/react-query";
import { act, cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import i18n from "@/i18n";
import { type Message, searchMessages } from "@/api/messenger";
import { SearchDialog } from "@/components/messenger/search-dialog";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

const api = vi.hoisted(() => ({ searchMessages: vi.fn() }));
const navigate = vi.hoisted(() => vi.fn());

vi.mock("@/api/messenger", () => ({ searchMessages: api.searchMessages }));
vi.mock("@tanstack/react-router", () => ({
	useNavigate: () => navigate,
}));

const NO_RESULTS = "Ничего не найдено";
const ERROR = "Ошибка поиска";
const ERROR_HINT = "Попробуйте ещё раз";
const RETRY = "Повторить";
const SEARCH = /^(Найти|Поиск)$/;
const userA = {
	id: "user-a",
	email: "user-a@example.test",
	timezone: "UTC",
	language: "ru",
	system_role: "member",
	onboarding_completed: true,
};
const userB = { ...userA, id: "user-b", email: "user-b@example.test" };
const clients: QueryClient[] = [];

type Deferred<T> = {
	promise: Promise<T>;
	resolve: (value: T) => void;
	reject: (reason?: unknown) => void;
};

function message(content: string, channelId = "channel-a"): Message {
	return {
		id: `message-${content}`,
		channel_id: channelId,
		sender_id: "sender-a",
		content_text: content,
		reply_count: 0,
		created_at: "2026-09-06T00:00:00Z",
	};
}

function deferred<T>(): Deferred<T> {
	let resolve: (value: T) => void = () => undefined;
	let reject: (reason?: unknown) => void = () => undefined;
	const promise = new Promise<T>((done, fail) => {
		resolve = done;
		reject = fail;
	});
	return { promise, resolve, reject };
}

function client(retry: boolean | number = false) {
	return new QueryClient({
		defaultOptions: {
			queries: {
				retry,
				retryDelay: 0,
				refetchOnWindowFocus: true,
				refetchOnReconnect: true,
			},
			mutations: { retry: false },
		},
	});
}

function renderDialog(channelId = "channel-a", queryClient = client()) {
	clients.push(queryClient);
	return {
		queryClient,
		...renderWithProviders(<SearchDialog channelId={channelId} />, {
			queryClient,
		}),
	};
}

function open() {
	fireEvent.click(screen.getByRole("button", { name: "Поиск сообщений" }));
}

function input() {
	return screen.getByRole("textbox");
}

function draft(value: string) {
	fireEvent.change(input(), { target: { value } });
}

function submit(value: string) {
	draft(value);
	fireEvent.keyDown(input(), { key: "Enter", code: "Enter" });
}

function searchButton() {
	return screen.getByRole("button", { name: SEARCH });
}

function expectNoSearchState(stale?: string) {
	if (stale) expect(screen.queryByText(stale)).not.toBeInTheDocument();
	expect(screen.queryByText(NO_RESULTS)).not.toBeInTheDocument();
	expect(screen.queryByRole("alert")).not.toBeInTheDocument();
}

const staleSettles = [
	{
		name: "success",
		settle: (request: Deferred<Message[]>) =>
			request.resolve([message("stale-result")]),
	},
	{
		name: "error",
		settle: (request: Deferred<Message[]>) =>
			request.reject(new Error("stale failure")),
	},
];

describe("SearchDialog scoped request lifecycle", () => {
	beforeEach(async () => {
		vi.clearAllMocks();
		vi.mocked(searchMessages).mockReset().mockResolvedValue([]);
		await i18n.changeLanguage("ru");
		useAuthStore.setState({
			accessToken: null,
			refreshToken: null,
			isAuthenticated: true,
			user: userA,
		});
		useMessengerStore.getState().reset();
	});

	afterEach(() => {
		cleanup();
		for (const queryClient of clients.splice(0)) queryClient.clear();
		focusManager.setFocused(true);
		onlineManager.setOnline(true);
		useAuthStore.setState({
			accessToken: null,
			refreshToken: null,
			isAuthenticated: false,
			user: null,
		});
		useMessengerStore.getState().reset();
	});

	it.each(["", "   "])(
		"does not fetch or show a false state for initial or whitespace draft",
		async (value) => {
			renderDialog();
			open();
			expect(searchMessages).not.toHaveBeenCalled();

			draft(value);
			fireEvent.keyDown(input(), { key: "Enter", code: "Enter" });

			expect(searchMessages).not.toHaveBeenCalled();
			expectNoSearchState();
		},
	);

	it.each([
		["channel", "channel-a"],
		["global", undefined],
	] as const)("submits trimmed query with exact %s scope", async (scope, id) => {
		renderDialog();
		open();
		if (scope === "global") {
			fireEvent.click(screen.getByRole("button", { name: "Во всех каналах" }));
		}

		submit("  needle  ");
		await waitFor(() =>
			expect(searchMessages).toHaveBeenCalledWith("needle", id),
		);
	});

	it.each([
		[[message("one-result")], "one-result"],
		[[], NO_RESULTS],
	] as Array<[Message[], string]>)(
		"renders fulfilled messages and fulfilled empty results as distinct success states",
		async (response, expected) => {
			vi.mocked(searchMessages).mockResolvedValueOnce(response);
			renderDialog();
			open();
			submit("needle");

			expect(await screen.findByText(expected)).toBeInTheDocument();
		},
	);

	it("renders localized rejection UI and retries only after an explicit retry", async () => {
		vi.mocked(searchMessages)
			.mockRejectedValueOnce(new Error("transport failed"))
			.mockResolvedValueOnce([message("recovered")]);
		renderDialog();
		open();
		submit("needle");

		const alert = await screen.findByRole("alert");
		expect(alert).toHaveTextContent(ERROR);
		expect(alert).toHaveTextContent(ERROR_HINT);
		expect(screen.queryByText(NO_RESULTS)).not.toBeInTheDocument();

		fireEvent.click(screen.getByRole("button", { name: RETRY }));
		await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(2));
		expect(await screen.findByText("recovered")).toBeInTheDocument();
	});

	it("deduplicates repeated Enter and click during one pending submission", async () => {
		const pending = deferred<Message[]>();
		vi.mocked(searchMessages).mockReturnValue(pending.promise);
		renderDialog();
		open();
		const submitControl = searchButton();

		submit("once");
		fireEvent.keyDown(input(), { key: "Enter", code: "Enter" });
		fireEvent.click(submitControl);
		await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(1));

		await act(async () => {
			pending.resolve([]);
			await pending.promise;
		});
	});

	it.each(staleSettles)(
		"masks old draft $name completion, including away-and-back edits",
		async ({ settle }) => {
			const stale = deferred<Message[]>();
			const fresh = deferred<Message[]>();
			vi.mocked(searchMessages)
				.mockImplementationOnce(() => stale.promise)
				.mockImplementationOnce(() => fresh.promise);
			renderDialog();
			open();
			submit("old");
			await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(1));
			draft("other");
			draft("old");

			await act(async () => {
				settle(stale);
				await Promise.resolve();
			});
			expectNoSearchState("stale-result");

			submit("fresh");
			fresh.resolve([message("fresh-result")]);
			await act(async () => await fresh.promise);
			expect(await screen.findByText("fresh-result")).toBeInTheDocument();
		},
	);

	it.each(staleSettles)(
		"masks old scope $name completion",
		async ({ settle }) => {
			const stale = deferred<Message[]>();
			vi.mocked(searchMessages).mockReturnValueOnce(stale.promise);
			renderDialog();
			open();
			submit("needle");
			await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(1));
			fireEvent.click(screen.getByRole("button", { name: "Во всех каналах" }));

			await act(async () => {
				settle(stale);
				await Promise.resolve();
			});
			expectNoSearchState("stale-result");
		},
	);

	it.each(staleSettles)(
		"does not restore a closed session after late $name completion",
		async ({ settle }) => {
			const stale = deferred<Message[]>();
			vi.mocked(searchMessages).mockReturnValueOnce(stale.promise);
			renderDialog();
			open();
			submit("close");
			await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(1));
			fireEvent.keyDown(document, { key: "Escape", code: "Escape" });
			open();
			expect(searchMessages).toHaveBeenCalledTimes(1);

			await act(async () => {
				settle(stale);
				await Promise.resolve();
			});
			expectNoSearchState("stale-result");
		},
	);

	it.each([
		{
			boundary: "channel",
			outcome: "success",
			settle: (request: Deferred<Message[]>) =>
				request.resolve([message("stale-result")]),
		},
		{
			boundary: "channel",
			outcome: "error",
			settle: (request: Deferred<Message[]>) =>
				request.reject(new Error("stale failure")),
		},
		{
			boundary: "identity",
			outcome: "success",
			settle: (request: Deferred<Message[]>) =>
				request.resolve([message("stale-result")]),
		},
		{
			boundary: "identity",
			outcome: "error",
			settle: (request: Deferred<Message[]>) =>
				request.reject(new Error("stale failure")),
		},
	] as const)(
		"masks late $outcome completion after committed $boundary A-B-A",
		async ({ boundary, settle }) => {
			const stale = deferred<Message[]>();
			vi.mocked(searchMessages).mockReturnValueOnce(stale.promise);
			const view = renderDialog();
			open();
			submit(`aba-${boundary}`);
			await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(1));

			if (boundary === "channel") {
				view.rerender(<SearchDialog channelId="channel-b" />);
				view.rerender(<SearchDialog channelId="channel-a" />);
			} else {
				act(() => useAuthStore.setState({ user: userB }));
				act(() => useAuthStore.setState({ user: userA }));
			}
			if (!screen.queryByRole("textbox")) open();
			expect(searchMessages).toHaveBeenCalledTimes(1);

			await act(async () => {
				settle(stale);
				await Promise.resolve();
			});
			expectNoSearchState("stale-result");
		},
	);

	it.each([
		[
			"unknown identity",
			() => useAuthStore.setState({ isAuthenticated: false, user: null }),
		],
		[
			"revoked current channel",
			() => useMessengerStore.getState().revokeChannel("channel-a"),
		],
	] as const)("fails closed for %s submission", async (_name, arrange) => {
		arrange();
		renderDialog();
		const trigger = screen.getByRole("button", { name: "Поиск сообщений" });
		if (trigger.hasAttribute("disabled")) {
			expect(trigger).toBeDisabled();
			expect(searchMessages).not.toHaveBeenCalled();
			return;
		}

		fireEvent.click(trigger);
		const field = screen.queryByRole("textbox");
		if (!field || field.hasAttribute("disabled")) {
			expect(searchMessages).not.toHaveBeenCalled();
			return;
		}

		draft("denied");
		const submitControl = screen.queryByRole("button", { name: SEARCH });
		if (!submitControl || submitControl.hasAttribute("disabled")) {
			expect(searchMessages).not.toHaveBeenCalled();
			return;
		}

		fireEvent.keyDown(input(), { key: "Enter", code: "Enter" });
		expect(searchMessages).not.toHaveBeenCalled();
	});

	it("removes a live-revoked result and preserves valid channel navigation", async () => {
		renderDialog();
		open();
		fireEvent.click(screen.getByRole("button", { name: "Во всех каналах" }));
		vi.mocked(searchMessages).mockResolvedValueOnce([
			message("revoked-result", "channel-b"),
		]);
		submit("revoked-result");
		expect(await screen.findByText("revoked-result")).toBeInTheDocument();

		act(() => useMessengerStore.getState().revokeChannel("channel-b"));
		expect(screen.queryByText("revoked-result")).not.toBeInTheDocument();
		expect(navigate).not.toHaveBeenCalled();

		act(() => useMessengerStore.getState().reset());
		if (!screen.queryByRole("textbox")) open();
		vi.mocked(searchMessages).mockResolvedValueOnce([
			message("navigable-result", "channel-c"),
		]);
		submit("allowed");
		fireEvent.click(await screen.findByText("navigable-result"));
		expect(navigate).toHaveBeenCalledWith({
			to: "/messenger/$channelId",
			params: { channelId: "channel-c" },
		});
	});

	it("cleans discarded transient queries without unhandled failure or unrelated cache loss", async () => {
		const failure = deferred<Message[]>();
		vi.mocked(searchMessages).mockReturnValueOnce(failure.promise);
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			const view = renderDialog();
			view.queryClient.setQueryData(["unrelated"], "preserved");
			open();
			submit("discard");
			await waitFor(() => {
				expect(searchMessages).toHaveBeenCalledTimes(1);
				expect(
					view.queryClient
						.getQueryCache()
						.findAll({ queryKey: ["messenger", "search-dialog"] }),
				).toHaveLength(1);
			});
			draft("changed");
			view.unmount();

			await waitFor(() =>
				expect(
					view.queryClient
						.getQueryCache()
						.findAll({ queryKey: ["messenger", "search-dialog"] }),
				).toHaveLength(0),
			);
			await act(async () => {
				failure.reject(new Error("discarded"));
				await Promise.resolve();
			});
			expect(unhandled).not.toHaveBeenCalled();
			expect(view.queryClient.getQueryData(["unrelated"])).toBe("preserved");
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
		}
	});

	it("disables automatic retry, focus refetch, and reconnect refetch", async () => {
		vi.mocked(searchMessages).mockRejectedValue(new Error("retry me"));
		renderDialog("channel-a", client(3));
		open();
		submit("no-auto");

		expect(await screen.findByRole("alert")).toHaveTextContent(ERROR);
		expect(searchMessages).toHaveBeenCalledTimes(1);
		act(() => {
			focusManager.setFocused(false);
			onlineManager.setOnline(false);
			focusManager.setFocused(true);
			onlineManager.setOnline(true);
		});
		await act(async () => await Promise.resolve());
		expect(searchMessages).toHaveBeenCalledTimes(1);
	});
});
