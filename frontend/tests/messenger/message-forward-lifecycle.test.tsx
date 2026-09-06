import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { flushSync } from "react-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import i18n from "@/i18n";
import {
	forwardMessage,
	type Channel,
	type ChannelCapabilities,
	type Message,
} from "@/api/messenger";
import { ForwardDialog } from "@/components/messenger/forward-dialog";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

const api = vi.hoisted(() => ({ forwardMessage: vi.fn() }));
const errors = vi.hoisted(() => ({ toastApiError: vi.fn() }));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return { ...actual, forwardMessage: api.forwardMessage };
});
vi.mock("@/lib/format-api-error", () => ({ toastApiError: errors.toastApiError }));

const SOURCE = "forward-source";
const SOURCE_B = "forward-source-b";
const TARGET = "forward-target";
const TARGET_B = "forward-target-b";
const OTHER = "forward-other";
const EN_FAILURE = "Unable to forward message";
const RU_FAILURE = "Не удалось переслать сообщение";
const actorA = {
	id: "forward-actor-a",
	email: "forward-actor-a@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};
const actorB = { ...actorA, id: "forward-actor-b", email: "forward-actor-b@example.test" };
type Deferred<T> = {
	promise: Promise<T>;
	reject: (reason?: unknown) => void;
	resolve: (value: T) => void;
};
type DialogOptions = {
	canForward?: boolean;
	channelId?: string;
	item?: Message;
	onOpenChange?: (open: boolean) => void;
	open?: boolean;
};

function capabilities(overrides: Partial<ChannelCapabilities> = {}): ChannelCapabilities {
	return {
		is_member: true,
		has_active_project_membership: false,
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
		can_react: false,
		can_delete_message: false,
		can_pin: false,
		can_list_pins: false,
		can_bookmark: false,
		can_promote: false,
		...overrides,
	};
}
function channel(id: string, currentCapabilities: ChannelCapabilities): Channel {
	return {
		id,
		name: id,
		channel_type: "private",
		is_readonly: false,
		capabilities: currentCapabilities,
		created_by: actorA.id,
		is_starred: false,
		created_at: "2026-09-06T00:00:00Z",
		updated_at: "2026-09-06T00:00:00Z",
	};
}
function source(overrides: Partial<ChannelCapabilities> = {}) {
	return channel(SOURCE, capabilities({ can_forward: true, ...overrides }));
}
function target(id = TARGET, overrides: Partial<ChannelCapabilities> = {}) {
	return channel(id, capabilities({ can_post: true, ...overrides }));
}
function message(overrides: Partial<Message> = {}): Message {
	return {
		id: "forward-message-a",
		channel_id: SOURCE,
		sender_id: actorA.id,
		content_text: "Forward lifecycle message",
		reply_count: 0,
		created_at: "2026-09-06T00:00:00Z",
		...overrides,
	};
}
function deferred<T>(): Deferred<T> {
	let resolve: (value: T) => void = () => undefined;
	let reject: (reason?: unknown) => void = () => undefined;
	const promise = new Promise<T>((onResolve, onReject) => {
		resolve = onResolve;
		reject = onReject;
	});
	return { promise, reject, resolve };
}
function observe<T>(request: Deferred<T>) {
	return request.promise.then(() => undefined, () => undefined);
}
async function settle(
	request: Deferred<Message>,
	observed: Promise<void>,
	outcome: "success" | "error",
	error: Error,
) {
	await act(async () => {
		if (outcome === "success") request.resolve(message());
		else request.reject(error);
		await observed;
	});
}
function install(...channels: Channel[]) {
	useMessengerStore.getState().setChannels(channels);
}
function dialog({
	item = message(),
	channelId = SOURCE,
	canForward = true,
	open = true,
	onOpenChange = vi.fn(),
}: DialogOptions = {}) {
	return (
		<ForwardDialog
			message={item}
			channelId={channelId}
			canForward={canForward}
			open={open}
			onOpenChange={onOpenChange}
		/>
	);
}
function renderDialog(options: DialogOptions = {}) {
	return render(dialog(options));
}
function targetButton(name = TARGET) {
	return screen.getByRole("button", { name });
}

describe("ForwardDialog forward command lifecycle", () => {
	beforeEach(async () => {
		vi.clearAllMocks();
		vi.mocked(forwardMessage).mockReset().mockResolvedValue(message());
		vi.mocked(toastApiError).mockReset();
		await i18n.changeLanguage("en");
		await i18n.loadNamespaces("messenger");
		useAuthStore.setState({
			accessToken: null,
			refreshToken: null,
			isAuthenticated: true,
			user: actorA,
		});
		useMessengerStore.getState().reset();
		install(source(), target());
	});
	afterEach(async () => {
		cleanup();
		useAuthStore.setState({
			accessToken: null,
			refreshToken: null,
			isAuthenticated: false,
			user: null,
		});
		useMessengerStore.getState().reset();
		await i18n.changeLanguage("ru");
	});

	it("posts exact source, message and target while unrelated channel updates retain the lock", async () => {
		const post = deferred<Message>();
		const onOpenChange = vi.fn();
		vi.mocked(forwardMessage).mockReturnValueOnce(post.promise);
		install(source(), target(), channel(OTHER, capabilities()));
		renderDialog({ onOpenChange });
		fireEvent.click(targetButton());
		await waitFor(() => expect(forwardMessage).toHaveBeenCalledWith(SOURCE, "forward-message-a", TARGET));
		expect(targetButton()).toBeDisabled();
		act(() => install(source(), target(), channel(OTHER, capabilities())));
		expect(targetButton()).toBeDisabled();
		await settle(post, observe(post), "success", new Error("unused"));
		await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
		expect(onOpenChange).toHaveBeenCalledTimes(1);
	});

	it.each([
		{ language: "en", fallback: EN_FAILURE },
		{ language: "ru", fallback: RU_FAILURE },
	] as const)("reports one current $language failure and admits an explicit retry", async ({ language, fallback }) => {
		const failure = new Error("forward " + language + " failure");
		const onOpenChange = vi.fn();
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			await i18n.changeLanguage(language);
			await i18n.loadNamespaces("messenger");
			vi.mocked(forwardMessage).mockRejectedValueOnce(failure);
			renderDialog({ onOpenChange });
			fireEvent.click(targetButton());
			await waitFor(() => expect(toastApiError).toHaveBeenCalledWith(failure, fallback));
			expect(toastApiError).toHaveBeenCalledTimes(1);
			expect(onOpenChange).not.toHaveBeenCalled();
			expect(targetButton()).toBeEnabled();
			fireEvent.click(targetButton());
			await waitFor(() => expect(forwardMessage).toHaveBeenNthCalledWith(2, SOURCE, "forward-message-a", TARGET));
			await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
			expect(forwardMessage).toHaveBeenCalledTimes(2);
			expect(onOpenChange).toHaveBeenCalledTimes(1);
			expect(toastApiError).toHaveBeenCalledTimes(1);
			expect(unhandled).not.toHaveBeenCalled();
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
		}
	});

	it("claims one request for two same-tick target clicks and disables every target", async () => {
		const post = deferred<Message>();
		const onOpenChange = vi.fn();
		vi.mocked(forwardMessage).mockReturnValueOnce(post.promise);
		install(source(), target(), target(TARGET_B));
		renderDialog({ onOpenChange });
		const first = targetButton(TARGET);
		const second = targetButton(TARGET_B);
		act(() => {
			first.click();
			second.click();
		});
		await waitFor(() => expect(forwardMessage).toHaveBeenCalledTimes(1));
		expect(forwardMessage).toHaveBeenCalledWith(SOURCE, "forward-message-a", TARGET);
		expect(first).toBeDisabled();
		expect(second).toBeDisabled();
		await settle(post, observe(post), "success", new Error("unused"));
	});

	it.each([
		"absent actor",
		"empty source id",
		"empty message id",
		"mismatched message channel",
		"empty target id",
	] as const)("does not admit a forward with %s", async (scenario) => {
		let item = message();
		let channelId = SOURCE;
		let buttonName = TARGET;
		let channels = [source(), target()];
		if (scenario === "absent actor") {
			useAuthStore.setState({ isAuthenticated: false, user: null });
		} else if (scenario === "empty source id") {
			channelId = "";
			item = message({ channel_id: "" });
			channels = [channel("", capabilities({ can_forward: true })), target()];
		} else if (scenario === "empty message id") {
			item = message({ id: "" });
		} else if (scenario === "mismatched message channel") {
			item = message({ channel_id: OTHER });
		} else {
			buttonName = "DM";
			channels = [source(), target("", { can_post: true })];
		}
		install(...channels);
		renderDialog({ item, channelId });
		fireEvent.click(targetButton(buttonName));
		await act(async () => Promise.resolve());
		expect(forwardMessage).not.toHaveBeenCalled();
	});

	it.each([
		"source absent",
		"source can_forward loss",
		"source revoke",
		"target absent",
		"target can_post loss",
		"target revoke",
	] as const)("rejects a captured handler after fresh %s", async (scenario) => {
		renderDialog();
		const control = targetButton();
		act(() => {
			if (scenario === "source absent") install(target());
			else if (scenario === "source can_forward loss") install(source({ can_forward: false }), target());
			else if (scenario === "source revoke") useMessengerStore.getState().revokeChannel(SOURCE);
			else if (scenario === "target absent") install(source());
			else if (scenario === "target can_post loss") install(source(), target(TARGET, { can_post: false }));
			else useMessengerStore.getState().revokeChannel(TARGET);
			control.click();
		});
		await act(async () => Promise.resolve());
		expect(forwardMessage).not.toHaveBeenCalled();
	});

	it.each([
		["prop canForward", "success"], ["prop canForward", "error"],
		["source can_forward", "success"], ["source can_forward", "error"],
		["target can_post", "success"], ["target can_post", "error"],
		["target revoke", "success"], ["target revoke", "error"],
	] as const)("silences an invalidated %s %s and then permits retry", async (boundary, outcome) => {
		const post = deferred<Message>();
		const failure = new Error("late " + boundary + " " + outcome);
		const onOpenChange = vi.fn();
		const item = message({ id: "loss-" + boundary + "-" + outcome });
		vi.mocked(forwardMessage).mockReturnValueOnce(post.promise);
		const view = renderDialog({ item, onOpenChange });
		fireEvent.click(targetButton());
		await waitFor(() => expect(forwardMessage).toHaveBeenCalledWith(SOURCE, item.id, TARGET));
		const observed = observe(post);
		if (boundary === "prop canForward") {
			act(() => view.rerender(dialog({ item, canForward: false, onOpenChange })));
			act(() => view.rerender(dialog({ item, onOpenChange })));
		} else {
			act(() => {
				if (boundary === "source can_forward") {
					install(source({ can_forward: false }), target());
					install(source(), target());
				} else if (boundary === "target can_post") {
					install(source(), target(TARGET, { can_post: false }));
					install(source(), target());
				} else {
					useMessengerStore.getState().revokeChannel(TARGET);
					useMessengerStore.getState().restoreChannelAccess(TARGET);
					install(source(), target());
				}
			});
		}
		await settle(post, observed, outcome, failure);
		expect(onOpenChange).not.toHaveBeenCalled();
		expect(toastApiError).not.toHaveBeenCalled();
		expect(targetButton()).toBeEnabled();
		fireEvent.click(targetButton());
		await waitFor(() => expect(forwardMessage).toHaveBeenNthCalledWith(2, SOURCE, item.id, TARGET));
		await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
		expect(forwardMessage).toHaveBeenCalledTimes(2);
		expect(onOpenChange).toHaveBeenCalledTimes(1);
		expect(toastApiError).not.toHaveBeenCalled();
	});

	it.each([
		["actor", "success"], ["actor", "error"],
		["source channel", "success"], ["source channel", "error"],
		["message id", "success"], ["message id", "error"],
		["dialog open", "success"], ["dialog open", "error"],
	] as const)("keeps a newer request pending when an old %s A-B-A %s settles", async (boundary, outcome) => {
		const stale = deferred<Message>();
		const current = deferred<Message>();
		const failure = new Error("stale " + boundary + " " + outcome);
		const onOpenChange = vi.fn();
		const initial = message({ id: "aba-forward-message-a" });
		vi.mocked(forwardMessage).mockReturnValueOnce(stale.promise).mockReturnValueOnce(current.promise);
		install(source(), channel(SOURCE_B, capabilities({ can_forward: true })), target());
		const view = renderDialog({ item: initial, onOpenChange });
		fireEvent.click(targetButton());
		await waitFor(() => expect(forwardMessage).toHaveBeenNthCalledWith(1, SOURCE, initial.id, TARGET));
		const observed = observe(stale);
		if (boundary === "actor") {
			act(() => useAuthStore.setState({ user: actorB }));
			act(() => useAuthStore.setState({ user: actorA }));
		} else if (boundary === "source channel") {
			act(() => view.rerender(dialog({
				item: message({ id: initial.id, channel_id: SOURCE_B }),
				channelId: SOURCE_B,
				onOpenChange,
			})));
			act(() => view.rerender(dialog({ item: initial, onOpenChange })));
		} else if (boundary === "message id") {
			act(() => view.rerender(dialog({ item: message({ id: "aba-forward-message-b" }), onOpenChange })));
			act(() => view.rerender(dialog({ item: initial, onOpenChange })));
		} else {
			act(() => view.rerender(dialog({ item: initial, open: false, onOpenChange })));
			act(() => view.rerender(dialog({ item: initial, onOpenChange })));
		}
		expect(targetButton()).toBeEnabled();
		fireEvent.click(targetButton());
		await waitFor(() => expect(forwardMessage).toHaveBeenNthCalledWith(2, SOURCE, initial.id, TARGET));
		expect(targetButton()).toBeDisabled();
		await settle(stale, observed, outcome, failure);
		expect(onOpenChange).not.toHaveBeenCalled();
		expect(toastApiError).not.toHaveBeenCalled();
		expect(targetButton()).toBeDisabled();
		await settle(current, observe(current), "success", new Error("unused"));
		await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
		expect(onOpenChange).toHaveBeenCalledTimes(1);
	});

	it.each(["success", "error"] as const)("silences an unmounted %s without an unhandled rejection", async (outcome) => {
		const post = deferred<Message>();
		const failure = new Error("unmounted " + outcome);
		const onOpenChange = vi.fn();
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			vi.mocked(forwardMessage).mockReturnValueOnce(post.promise);
			const view = renderDialog({ onOpenChange });
			fireEvent.click(targetButton());
			await waitFor(() => expect(forwardMessage).toHaveBeenCalledTimes(1));
			view.unmount();
			await settle(post, observe(post), outcome, failure);
			expect(onOpenChange).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			expect(unhandled).not.toHaveBeenCalled();
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
		}
	});

	it("releases before the latest callback and isolates an old completion from callback re-entry", async () => {
		const first = deferred<Message>();
		const second = deferred<Message>();
		const oldCallback = vi.fn();
		const latestCallback = vi.fn();
		const item = message({ id: "latest-callback-forward" });
		vi.mocked(forwardMessage)
			.mockReturnValueOnce(first.promise)
			.mockReturnValueOnce(second.promise);
		const view = renderDialog({ item, onOpenChange: oldCallback });
		latestCallback.mockImplementationOnce(() => {
			flushSync(() => view.rerender(dialog({ item, onOpenChange: latestCallback })));
			targetButton().click();
		});
		fireEvent.click(targetButton());
		await waitFor(() => expect(forwardMessage).toHaveBeenCalledTimes(1));
		act(() => view.rerender(dialog({ item, onOpenChange: latestCallback })));
		expect(targetButton()).toBeDisabled();
		expect(forwardMessage).toHaveBeenCalledTimes(1);
		await settle(first, observe(first), "success", new Error("unused"));
		await waitFor(() => expect(forwardMessage).toHaveBeenNthCalledWith(2, SOURCE, item.id, TARGET));
		expect(forwardMessage).toHaveBeenCalledTimes(2);
		expect(targetButton()).toBeDisabled();
		expect(latestCallback).toHaveBeenCalledTimes(1);
		expect(latestCallback).toHaveBeenNthCalledWith(1, false);
		expect(oldCallback).not.toHaveBeenCalled();
		await settle(second, observe(second), "success", new Error("unused"));
		await waitFor(() => expect(targetButton()).toBeEnabled());
		expect(latestCallback).toHaveBeenCalledTimes(2);
		expect(latestCallback).toHaveBeenNthCalledWith(2, false);
		expect(toastApiError).not.toHaveBeenCalled();
		expect(oldCallback).not.toHaveBeenCalled();
	});

	it.each(["success", "error"] as const)("never revives a batched actor A-B-A %s before explicit retry", async (outcome) => {
		const post = deferred<Message>();
		const failure = new Error("batched actor " + outcome);
		const onOpenChange = vi.fn();
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			vi.mocked(forwardMessage).mockReturnValueOnce(post.promise);
			renderDialog({ onOpenChange });
			fireEvent.click(targetButton());
			await waitFor(() => expect(forwardMessage).toHaveBeenCalledTimes(1));
			const observed = observe(post);
			act(() => {
				useAuthStore.setState({ user: actorB });
				useAuthStore.setState({ user: actorA });
			});
			await settle(post, observed, outcome, failure);
			expect(onOpenChange).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			expect(targetButton()).toBeEnabled();
			fireEvent.click(targetButton());
			await waitFor(() => expect(forwardMessage).toHaveBeenNthCalledWith(2, SOURCE, "forward-message-a", TARGET));
			await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
			expect(forwardMessage).toHaveBeenCalledTimes(2);
			expect(onOpenChange).toHaveBeenCalledTimes(1);
			expect(unhandled).not.toHaveBeenCalled();
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
		}
	});
});
