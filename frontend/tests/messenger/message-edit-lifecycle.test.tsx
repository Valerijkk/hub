import { QueryClient } from "@tanstack/react-query";
import { act, cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import i18n from "@/i18n";
import { editMessage, type Channel, type ChannelCapabilities, type Message } from "@/api/messenger";
import { MessageItem } from "@/components/messenger/message-item";
import { messagesQueryKey } from "@/hooks/use-messages";
import { toastApiError } from "@/lib/format-api-error";
import { useAIStore } from "@/stores/ai-store";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

const api = vi.hoisted(() => ({ editMessage: vi.fn() }));
const errors = vi.hoisted(() => ({ toastApiError: vi.fn() }));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return { ...actual, editMessage: api.editMessage };
});
vi.mock("@/lib/format-api-error", () => ({ toastApiError: errors.toastApiError }));
vi.mock("@/components/messenger/forward-dialog", () => ({ ForwardDialog: () => null }));

const CHANNEL_A = "edit-channel-a";
const CHANNEL_B = "edit-channel-b";
const EDIT_LABEL = "Редактировать сообщение";
const SAVE_LABEL = "Сохранить сообщение";
const REPLY_LABEL = "Ответить";
const DELETE_LABEL = "Удалить сообщение";
const EDIT_FALLBACK_EN = "Unable to edit message";
const EDIT_FALLBACK_RU = "Не удалось изменить сообщение";

const userA = {
	id: "edit-user-a",
	email: "edit-user-a@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};
const userB = { ...userA, id: "edit-user-b", email: "edit-user-b@example.test" };
type Timeline = { pages: Message[][]; pageParams: unknown[] };
const queryClients: QueryClient[] = [];

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
		can_edit: true,
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
		id, name: id, channel_type: "private", is_readonly: false,
		capabilities: currentCapabilities, created_by: userA.id, is_starred: false,
		created_at: "2026-09-06T00:00:00Z", updated_at: "2026-09-06T00:00:00Z",
	};
}
function message(overrides: Partial<Message> = {}): Message {
	return {
		id: "edit-message-a", channel_id: CHANNEL_A, sender_id: userA.id,
		content_text: "Original edit text", reply_count: 0,
		created_at: "2026-09-06T00:00:00Z", ...overrides,
	};
}
function timeline(items: Message[]): Timeline {
	return { pages: [items], pageParams: [undefined] };
}
function deferred<T>() {
	let resolve: (value: T) => void = () => undefined;
	let reject: (reason?: unknown) => void = () => undefined;
	const promise = new Promise<T>((onResolve, onReject) => {
		resolve = onResolve;
		reject = onReject;
	});
	return { promise, reject, resolve };
}
function createQueryClient() {
	return new QueryClient({
		defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
	});
}
function installChannels(entries: Array<{ id: string; currentCapabilities: ChannelCapabilities }>) {
	useMessengerStore.getState().setChannels(
		entries.map(({ id, currentCapabilities }) => channel(id, currentCapabilities)),
	);
}
function renderMessageItem(
	item = message(),
	rowChannelId = CHANNEL_A,
	itemCapabilities = capabilities(),
) {
	const queryClient = createQueryClient();
	queryClients.push(queryClient);
	return renderWithProviders(
		<MessageItem message={item} channelId={rowChannelId} capabilities={itemCapabilities} />,
		{ queryClient },
	);
}
function currentEditButton() {
	return screen.getByRole("button", { name: EDIT_LABEL });
}
function beginEdit(nextText: string) {
	fireEvent.click(currentEditButton());
	const editor = screen.getByRole("textbox");
	fireEvent.change(editor, { target: { value: nextText } });
	return editor;
}
function saveEdit() {
	fireEvent.click(screen.getByRole("button", { name: SAVE_LABEL }));
}

describe("message edit save lifecycle", () => {
	beforeEach(async () => {
		vi.clearAllMocks();
		vi.mocked(editMessage).mockReset().mockResolvedValue(message());
		vi.mocked(toastApiError).mockReset();
		await i18n.changeLanguage("en");
		await i18n.loadNamespaces("messenger");
		useAIStore.setState({ loaded: true, botLoadError: false });
		useAuthStore.setState({
			accessToken: null, refreshToken: null, isAuthenticated: true, user: userA,
		});
		useMessengerStore.getState().reset();
		installChannels([
			{ id: CHANNEL_A, currentCapabilities: capabilities() },
			{ id: CHANNEL_B, currentCapabilities: capabilities() },
		]);
	});
	afterEach(async () => {
		cleanup();
		for (const queryClient of queryClients.splice(0)) queryClient.clear();
		useAuthStore.setState({
			accessToken: null, refreshToken: null, isAuthenticated: false, user: null,
		});
		useMessengerStore.getState().reset();
		useAIStore.setState({ loaded: false, botLoadError: false });
		await i18n.changeLanguage("ru");
	});

	it("writes the exact trimmed optimistic origin timeline without touching another identity cache", async () => {
		const request = deferred<Message>();
		const item = message({ id: "optimistic-edit-message" });
		const originKey = messagesQueryKey(CHANNEL_A, userA.id);
		const unrelatedKey = messagesQueryKey(CHANNEL_B, userB.id);
		const origin = timeline([item]);
		const unrelated = timeline([
			message({ id: "unrelated-edit-message", channel_id: CHANNEL_B, sender_id: userB.id }),
		]);
		const expectedDocument = {
			type: "doc", children: [{ type: "p", children: [{ text: "Trimmed update" }] }],
		};
		vi.mocked(editMessage).mockReturnValueOnce(request.promise);
		installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities() }]);
		const { queryClient } = renderMessageItem(item);
		queryClient.setQueryData(originKey, origin);
		queryClient.setQueryData(unrelatedKey, unrelated);

		beginEdit("  Trimmed update  ");
		saveEdit();

		await waitFor(() => {
			expect(editMessage).toHaveBeenCalledWith(CHANNEL_A, item.id, {
				content_json: expectedDocument, content_text: "Trimmed update",
			});
		});
		expect(editMessage).toHaveBeenCalledTimes(1);
		expect(queryClient.getQueryData<Timeline>(originKey)).toEqual({
			pages: [[expect.objectContaining({
				id: item.id, content_json: expectedDocument,
				content_text: "Trimmed update", edited_at: expect.any(String),
			})]],
			pageParams: [undefined],
		});
		expect(queryClient.getQueryData(unrelatedKey)).toBe(unrelated);
		await act(async () => {
			request.resolve(message({ id: item.id, content_text: "Trimmed update" }));
			await request.promise;
		});
	});

	it("keeps whitespace-only edits local with no cache write or transport", async () => {
		const item = message({ id: "whitespace-edit-message" });
		const originKey = messagesQueryKey(CHANNEL_A, userA.id);
		const origin = timeline([item]);
		const { queryClient } = renderMessageItem(item);
		queryClient.setQueryData(originKey, origin);
		const editor = beginEdit("   ");
		expect(screen.getByRole("button", { name: SAVE_LABEL })).toBeDisabled();
		fireEvent.keyDown(editor, { key: "Enter" });
		await act(async () => Promise.resolve());
		expect(editMessage).not.toHaveBeenCalled();
		expect(toastApiError).not.toHaveBeenCalled();
		expect(queryClient.getQueryData(originKey)).toBe(origin);
	});

	it("treats unchanged trimmed content as a no-op", async () => {
		const item = message({ id: "unchanged-edit-message" });
		const originKey = messagesQueryKey(CHANNEL_A, userA.id);
		const origin = timeline([item]);
		const { queryClient } = renderMessageItem(item);
		queryClient.setQueryData(originKey, origin);
		beginEdit("  Original edit text  ");
		saveEdit();
		await act(async () => Promise.resolve());
		expect(editMessage).not.toHaveBeenCalled();
		expect(toastApiError).not.toHaveBeenCalled();
		expect(queryClient.getQueryData(originKey)).toBe(origin);
	});

	it.each([
		{
			name: "the message belongs to another author",
			item: message({ id: "foreign-edit-message", sender_id: userB.id }),
			arrange: () => undefined,
		},
		{
			name: "the authenticated actor is absent",
			item: message({ id: "unknown-edit-message" }),
			arrange: () => { useAuthStore.setState({ isAuthenticated: false, user: null }); },
		},
		{
			name: "the live channel capability is absent",
			item: message({ id: "capability-edit-message" }),
			arrange: () => {
				installChannels([{
					id: CHANNEL_A, currentCapabilities: capabilities({ can_edit: false }),
				}]);
			},
		},
		{
			name: "the rendered message belongs to another channel",
			item: message({ id: "mismatched-edit-message", channel_id: CHANNEL_B }),
			arrange: () => undefined,
		},
	])("denies edit admission before cache mutation or transport when $name", async ({ item, arrange }) => {
		const originKey = messagesQueryKey(CHANNEL_A, userA.id);
		const origin = timeline([item]);
		arrange();
		const { queryClient } = renderMessageItem(item, CHANNEL_A, capabilities({ can_edit: true }));
		queryClient.setQueryData(originKey, origin);
		const edit = screen.queryByRole("button", { name: EDIT_LABEL });
		if (edit) fireEvent.click(edit);
		await act(async () => Promise.resolve());
		expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
		expect(editMessage).not.toHaveBeenCalled();
		expect(queryClient.getQueryData(originKey)).toBe(origin);
	});

	it("disables only Edit and prevents save or edit re-entry while its request is pending", async () => {
		const request = deferred<Message>();
		vi.mocked(editMessage).mockReturnValueOnce(request.promise);
		installChannels([{
			id: CHANNEL_A,
			currentCapabilities: capabilities({ can_edit: true, can_reply: true, is_member: true }),
		}]);
		renderMessageItem(
			message({ id: "pending-edit-message" }), CHANNEL_A,
			capabilities({ can_edit: true, can_reply: true, is_member: true }),
		);
		beginEdit("Pending edit");
		const save = screen.getByRole("button", { name: SAVE_LABEL });
		fireEvent.click(save);
		fireEvent.click(save);
		await waitFor(() => expect(editMessage).toHaveBeenCalledTimes(1));
		const pendingEdit = currentEditButton();
		expect(pendingEdit).toBeDisabled();
		expect(pendingEdit).toHaveAttribute("type", "button");
		expect(screen.getByRole("button", { name: REPLY_LABEL })).toBeEnabled();
		expect(screen.getByRole("button", { name: DELETE_LABEL })).toBeEnabled();
		fireEvent.click(pendingEdit);
		expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
		await act(async () => {
			request.resolve(message({ id: "pending-edit-message" }));
			await request.promise;
		});
		await waitFor(() => expect(currentEditButton()).toBeEnabled());
	});

	it.each([
		{ language: "en", fallback: EDIT_FALLBACK_EN },
		{ language: "ru", fallback: EDIT_FALLBACK_RU },
	])("invalidates only the captured origin and reports the $language fallback", async ({ language, fallback }) => {
		const failure = new Error(`edit failure ${language}`);
		const item = message({ id: `failed-edit-${language}` });
		const originKey = messagesQueryKey(CHANNEL_A, userA.id);
		const origin = timeline([item]);
		vi.mocked(editMessage).mockRejectedValueOnce(failure);
		await i18n.changeLanguage(language);
		await i18n.loadNamespaces("messenger");
		const { queryClient } = renderMessageItem(item);
		queryClient.setQueryData(originKey, origin);
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		beginEdit("Failed update");
		saveEdit();
		await waitFor(() => {
			expect(toastApiError).toHaveBeenCalledWith(failure, fallback);
		});
		expect(invalidate).toHaveBeenCalledWith({ queryKey: originKey, exact: true });
		expect(invalidate).toHaveBeenCalledTimes(1);
		expect(currentEditButton()).toBeEnabled();
		invalidate.mockRestore();
	});

	it("allows a manual edit and save retry only after the current failure settles", async () => {
		const failure = new Error("first edit failure");
		vi.mocked(editMessage)
			.mockRejectedValueOnce(failure)
			.mockResolvedValueOnce(message({ id: "manual-retry-edit-message" }));
		renderMessageItem(message({ id: "manual-retry-edit-message" }));
		beginEdit("Retry this edit");
		saveEdit();
		await waitFor(() =>
			expect(toastApiError).toHaveBeenCalledWith(failure, EDIT_FALLBACK_EN),
		);
		expect(currentEditButton()).toBeEnabled();
		beginEdit("Retry this edit");
		saveEdit();
		await waitFor(() => expect(editMessage).toHaveBeenCalledTimes(2));
		expect(toastApiError).toHaveBeenCalledTimes(1);
	});

	it.each([
		["user", "success"], ["user", "error"],
		["channel", "success"], ["channel", "error"],
		["message", "success"], ["message", "error"],
	] as const)(
		"does not let an old %s A-B-A %s completion unlock, invalidate, or report a newer edit",
		async (boundary, outcome) => {
			const stale = deferred<Message>();
			const current = deferred<Message>();
			const staleError = new Error(`stale ${boundary} edit`);
			const initial = message({ id: "aba-edit-message-a", content_text: "A message" });
			const currentCapabilities = capabilities({ can_edit: true });
			const unhandled = vi.fn();
			window.addEventListener("unhandledrejection", unhandled);
			vi.mocked(editMessage)
				.mockReturnValueOnce(stale.promise)
				.mockReturnValueOnce(current.promise);
			installChannels([
				{ id: CHANNEL_A, currentCapabilities },
				{ id: CHANNEL_B, currentCapabilities },
			]);
			const view = renderMessageItem(initial, CHANNEL_A, currentCapabilities);
			const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
			try {
				beginEdit("Stale edit");
				saveEdit();
				await waitFor(() => expect(editMessage).toHaveBeenNthCalledWith(
					1, CHANNEL_A, initial.id, expect.any(Object),
				));
				const staleSettled = stale.promise.then(() => undefined, () => undefined);

				if (boundary === "user") {
					act(() => useAuthStore.setState({ user: userB }));
					await waitFor(() => expect(
						screen.queryByRole("button", { name: EDIT_LABEL }),
					).not.toBeInTheDocument());
					act(() => useAuthStore.setState({ user: userA }));
				} else if (boundary === "channel") {
					view.rerender(
						<MessageItem
							message={message({ id: "aba-edit-message-b", channel_id: CHANNEL_B })}
							channelId={CHANNEL_B} capabilities={currentCapabilities}
						/>,
					);
					view.rerender(
						<MessageItem message={initial} channelId={CHANNEL_A} capabilities={currentCapabilities} />,
					);
				} else {
					view.rerender(
						<MessageItem
							message={message({ id: "aba-edit-message-b", channel_id: CHANNEL_A })}
							channelId={CHANNEL_A} capabilities={currentCapabilities}
						/>,
					);
					view.rerender(
						<MessageItem message={initial} channelId={CHANNEL_A} capabilities={currentCapabilities} />,
					);
				}
				expect(currentEditButton()).toBeEnabled();
				beginEdit("Current edit");
				saveEdit();
				await waitFor(() => expect(editMessage).toHaveBeenNthCalledWith(
					2, CHANNEL_A, initial.id, expect.any(Object),
				));
				expect(currentEditButton()).toBeDisabled();
				await act(async () => {
					if (outcome === "success") stale.resolve(message({ id: initial.id }));
					else stale.reject(staleError);
					await staleSettled;
				});
				expect(currentEditButton()).toBeDisabled();
				expect(invalidate).not.toHaveBeenCalled();
				expect(toastApiError).not.toHaveBeenCalled();
				expect(unhandled).not.toHaveBeenCalled();
				await act(async () => {
					current.resolve(message({ id: initial.id }));
					await current.promise;
				});
				await waitFor(() => expect(currentEditButton()).toBeEnabled());
			} finally {
				invalidate.mockRestore();
				window.removeEventListener("unhandledrejection", unhandled);
			}
		},
	);

	it.each(["revoke", "authority loss", "unmount"] as const)(
		"silently consumes a deferred rejection after %s",
		async (boundary) => {
			const request = deferred<Message>();
			const failure = new Error(`late edit ${boundary}`);
			const currentCapabilities = capabilities({ can_edit: true });
			const unhandled = vi.fn();
			window.addEventListener("unhandledrejection", unhandled);
			vi.mocked(editMessage).mockReturnValueOnce(request.promise);
			installChannels([{ id: CHANNEL_A, currentCapabilities }]);
			const view = renderMessageItem(
				message({ id: `late-edit-${boundary}` }), CHANNEL_A, currentCapabilities,
			);
			const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
			try {
				beginEdit("Deferred edit");
				saveEdit();
				await waitFor(() => expect(editMessage).toHaveBeenCalledTimes(1));
				const settled = request.promise.then(() => undefined, () => undefined);
				if (boundary === "revoke") {
					act(() => useMessengerStore.getState().revokeChannel(CHANNEL_A));
				} else if (boundary === "authority loss") {
					act(() => installChannels([{
						id: CHANNEL_A, currentCapabilities: capabilities({ can_edit: false }),
					}]));
				} else {
					view.unmount();
				}
				await act(async () => {
					request.reject(failure);
					await settled;
				});
				expect(invalidate).not.toHaveBeenCalled();
				expect(toastApiError).not.toHaveBeenCalled();
				expect(unhandled).not.toHaveBeenCalled();
			} finally {
				invalidate.mockRestore();
				window.removeEventListener("unhandledrejection", unhandled);
			}
		},
	);

	it("closes an unsaved edit and restores its original draft after live can_edit loss", async () => {
		const item = message();
		const currentCapabilities = capabilities({ can_edit: true });
		installChannels([{ id: CHANNEL_A, currentCapabilities }]);
		const { queryClient } = renderMessageItem(item, CHANNEL_A, currentCapabilities);
		const originKey = messagesQueryKey(CHANNEL_A, userA.id);
		const origin = timeline([item]);
		queryClient.setQueryData(originKey, origin);
		beginEdit("Unsaved authority draft");
		act(() => installChannels([{
			id: CHANNEL_A, currentCapabilities: capabilities({ can_edit: false }),
		}]));
		await waitFor(() => expect(screen.queryByRole("textbox")).not.toBeInTheDocument());
		act(() => installChannels([{ id: CHANNEL_A, currentCapabilities }]));
		fireEvent.click(currentEditButton());
		expect(screen.getByRole("textbox")).toHaveValue("Original edit text");
		expect(editMessage).not.toHaveBeenCalled();
		expect(toastApiError).not.toHaveBeenCalled();
		expect(queryClient.getQueryData(originKey)).toBe(origin);
	});

	it.each(["user", "channel", "message"] as const)(
		"resets reused row edit mode and draft across a committed %s scope change",
		async (boundary) => {
			const initial = message({ id: "reused-edit-message-a", content_text: "Original A" });
			const currentCapabilities = capabilities({ can_edit: true });
			installChannels([
				{ id: CHANNEL_A, currentCapabilities }, { id: CHANNEL_B, currentCapabilities },
			]);
			const view = renderMessageItem(initial, CHANNEL_A, currentCapabilities);
			beginEdit("Unsaved draft A");
			expect(screen.getByRole("textbox")).toHaveValue("Unsaved draft A");
			if (boundary === "user") {
				const next = message({
					id: "reused-edit-message-b", sender_id: userB.id, content_text: "Original B",
				});
				act(() => {
					useAuthStore.setState({ user: userB });
					view.rerender(
						<MessageItem message={next} channelId={CHANNEL_A} capabilities={currentCapabilities} />,
					);
				});
				act(() => {
					useAuthStore.setState({ user: userA });
					view.rerender(
						<MessageItem message={initial} channelId={CHANNEL_A} capabilities={currentCapabilities} />,
					);
				});
			} else if (boundary === "channel") {
				view.rerender(
					<MessageItem
						message={message({
							id: "reused-edit-message-b", channel_id: CHANNEL_B, content_text: "Original B",
						})}
						channelId={CHANNEL_B} capabilities={currentCapabilities}
					/>,
				);
				view.rerender(
					<MessageItem message={initial} channelId={CHANNEL_A} capabilities={currentCapabilities} />,
				);
			} else {
				view.rerender(
					<MessageItem
						message={message({ id: "reused-edit-message-b", content_text: "Original B" })}
						channelId={CHANNEL_A} capabilities={currentCapabilities}
					/>,
				);
				view.rerender(
					<MessageItem message={initial} channelId={CHANNEL_A} capabilities={currentCapabilities} />,
				);
			}
			await waitFor(() => expect(screen.queryByRole("textbox")).not.toBeInTheDocument());
			fireEvent.click(currentEditButton());
			expect(screen.getByRole("textbox")).toHaveValue("Original A");
			expect(editMessage).not.toHaveBeenCalled();
		},
	);
});
