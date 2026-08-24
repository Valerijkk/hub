import { fireEvent, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Message, SendMessageRequest } from "@/api/messenger";
import { MessageComposer } from "@/components/messenger/message-composer";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

const UUIDS = [

	"00000000-0000-4000-8000-000000000001",
	"00000000-0000-4000-8000-000000000002",
	"00000000-0000-4000-8000-000000000003",
	"00000000-0000-4000-8000-000000000004",
	"00000000-0000-4000-8000-000000000005",
] as const;

const sendMessage = vi.fn();
const sendTypingIndicator = vi.fn();
const uploadFile = vi.fn();
const invalidateMessages = vi.fn();
const toastError = vi.fn();
const toastDismiss = vi.fn();
const randomUUID = vi.fn<() => string>();

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		sendMessage: (...args: unknown[]) => sendMessage(...args),
		sendTypingIndicator: (...args: unknown[]) => sendTypingIndicator(...args),
	};
});

vi.mock("@/api/files", () => ({
	uploadFile: (...args: unknown[]) => uploadFile(...args),
}));

vi.mock("@/hooks/use-invalidate", () => ({
	useInvalidateMessages: () => invalidateMessages,
}));

vi.mock("sonner", () => ({
	toast: {
		error: (...args: unknown[]) => toastError(...args),
		dismiss: (...args: unknown[]) => toastDismiss(...args),
	},
}));

vi.mock("@/components/ui/popover", () => ({
	Popover: ({ children }: { children: ReactNode }) => <>{children}</>,
	PopoverContent: ({ children }: { children: ReactNode }) => <>{children}</>,
	PopoverTrigger: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

vi.mock("@/components/messenger/mention-autocomplete", () => ({
	MentionAutocomplete: () => null,
}));

vi.mock("@/components/messenger/create-poll-dialog", () => ({
	CreatePollDialog: () => null,
}));

vi.mock("@/components/messenger/gif-picker", () => ({
	GifPicker: ({ onSelect }: { onSelect: (url: string) => void }) => (
		<button
			type="button"
			onClick={() => onSelect("https://cdn.example.test/retry.gif")}
		>
			Выбрать GIF для retry
		</button>
	),
}));

vi.mock("@/components/messenger/sticker-picker", () => ({
	buildStickerContentJson: (sticker: {
		pack_id: string;
		sticker_id: string;
		emoji: string;
		label: string;
	}) => ({
		type: "doc",
		children: [
			{
				type: "sticker",
				pack_id: sticker.pack_id,
				sticker_id: sticker.sticker_id,
				emoji: sticker.emoji,
				label: sticker.label,
				children: [{ text: "" }],
			},
		],
	}),
	StickerPicker: ({
		onSelect,
	}: {
		onSelect: (sticker: {
			pack_id: string;
			sticker_id: string;
			emoji: string;
			label: string;
		}) => void;
	}) => (
		<button
			type="button"
			onClick={() =>
				onSelect({
					pack_id: "retry-pack",
					sticker_id: "retry-sticker",
					emoji: "🔁",
					label: "Retry sticker",
				})
			}
		>
			Выбрать стикер для retry
		</button>
	),
}));

vi.mock("@/components/messenger/voice-recorder", () => ({
	VoiceRecorder: ({
		onSend,
	}: {
		onSend: (voiceUrl: string, durationMs: number) => void;
	}) => (
		<button
			type="button"
			onClick={() => onSend("https://cdn.example.test/retry.webm", 1234)}
		>
			Отправить голосовое для retry
		</button>
	),
}));

interface Deferred<T> {
	promise: Promise<T>;
	resolve: (value: T) => void;
	reject: (reason?: unknown) => void;
}

interface RetryToastOptions {
	duration: number;
	dismissible: boolean;
	action: { label: string; onClick: () => void };
	cancel: { label: string; onClick: () => void };
}

interface OwnedRetryToast {
	id: unknown;
	options: RetryToastOptions;
}

function deferred<T>(): Deferred<T> {
	let resolve!: (value: T) => void;
	let reject!: (reason?: unknown) => void;
	const promise = new Promise<T>((resolvePromise, rejectPromise) => {
		resolve = resolvePromise;
		reject = rejectPromise;
	});
	return { promise, resolve, reject };
}

function retryToastAt(callIndex: number): OwnedRetryToast {
	expect(toastError).toHaveBeenCalledTimes(callIndex + 1);
	const [, options] = toastError.mock.calls[callIndex] as [
		unknown,
		RetryToastOptions,
	];
	expect(options).toMatchObject({
		duration: Number.POSITIVE_INFINITY,
		dismissible: false,
		action: { label: "Повторить" },
		cancel: { label: "Отменить отправку" },
	});
	return { id: toastError.mock.results[callIndex]?.value, options };
}

function sendButton() {
	return screen.getByRole("button", { name: "Отправить сообщение" });
}

function composerInput() {
	return screen.getByRole("textbox");
}

async function addAttachment(container: HTMLElement): Promise<void> {
	const fileInput = container.querySelector("input[type=file]");
	if (!(fileInput instanceof HTMLInputElement)) {
		throw new Error("MessageComposer file input was not rendered");
	}

	fireEvent.change(fileInput, {
		target: {
			files: [
				new File(["retry attachment"], "retry.txt", {
					type: "text/plain",
				}),
			],
		},
	});

	await waitFor(() => {
		expect(uploadFile).toHaveBeenCalledTimes(1);
		expect(screen.getByText("retry.txt")).toBeInTheDocument();
	});
}

describe("FE-P149-MESSENGER · retry nonce and optimistic recovery", () => {
	let uuidIndex = 0;
	let toastIndex = 0;

	beforeEach(() => {
		sendMessage.mockReset();
		sendTypingIndicator.mockReset();
		uploadFile.mockReset();
		invalidateMessages.mockReset();
		toastError.mockReset();
		toastDismiss.mockReset();
		randomUUID.mockReset();
		uuidIndex = 0;
		toastIndex = 0;
		randomUUID.mockImplementation(() => UUIDS[uuidIndex++] ?? UUIDS.at(-1)!);
		toastError.mockImplementation(
			(_message, options?: { id?: string | number }) =>
				options?.id ?? `retry-toast-${++toastIndex}`,
		);
		uploadFile.mockResolvedValue("retry-attachment-id");
		useMessengerStore.setState({ replyToMessage: null });
		vi.stubGlobal("crypto", { randomUUID });
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it("requires client_nonce in outgoing and received messenger contracts", () => {
		const outgoing: SendMessageRequest = {
			content_json: { type: "doc", children: [] },
			client_nonce: UUIDS[0],
		};
		const received: Message = {
			id: "message-with-nonce",
			channel_id: "channel-retry",
			sender_id: "user-retry",
			content_text: "nonce acknowledgement",
			client_nonce: UUIDS[0],
			reply_count: 0,
			created_at: "2026-08-24T00:00:00Z",
		};

		expect(outgoing.client_nonce).toBe(UUIDS[0]);
		expect(received.client_nonce).toBe(UUIDS[0]);
	});

	it("retries one immutable text-and-attachment request without replacing a newer draft", async () => {
		const retry = deferred<unknown>();
		sendMessage
			.mockRejectedValueOnce(new Error("initial send failed"))
			.mockReturnValueOnce(retry.promise);
		const { container } = renderWithProviders(
			<MessageComposer channelId="channel-retry" threadId="thread-retry" />,
		);

		await addAttachment(container);
		fireEvent.change(composerInput(), { target: { value: "initial draft" } });
		fireEvent.click(sendButton());

		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(1));
		const [, firstPayload] = sendMessage.mock.calls[0] as [
			string,
			SendMessageRequest,
		];
		const immutableFirstPayload = structuredClone(firstPayload);
		expect(sendMessage.mock.calls[0]?.[0]).toBe("channel-retry");
		expect(firstPayload).toMatchObject({
			content_text: "initial draft",
			thread_id: "thread-retry",
			attachment_ids: ["retry-attachment-id"],
			client_nonce: UUIDS[0],
		});

		const retryToast = await waitFor(() => retryToastAt(0));
		expect(composerInput()).toHaveValue("");
		expect(screen.queryByText("retry.txt")).not.toBeInTheDocument();

		retryToast.options.action.onClick();
		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(2));
		const [, retriedPayload] = sendMessage.mock.calls[1] as [
			string,
			SendMessageRequest,
		];
		expect(sendMessage.mock.calls[1]?.[0]).toBe("channel-retry");
		expect(retriedPayload).toEqual(immutableFirstPayload);
		expect(retriedPayload.client_nonce).toBe(firstPayload.client_nonce);
		expect(randomUUID).toHaveBeenCalledTimes(1);

		expect(composerInput()).toBeEnabled();
		fireEvent.change(composerInput(), { target: { value: "newer draft" } });
		expect(composerInput()).toHaveValue("newer draft");

		retry.resolve({ id: "retry-success" });
		await waitFor(() => {
			expect(invalidateMessages).toHaveBeenCalledTimes(1);
			expect(toastDismiss).toHaveBeenCalledWith(retryToast.id);
			expect(composerInput()).toHaveValue("newer draft");
		});
	});

	it("cancels without another request and gives the next logical text send a new nonce", async () => {
		sendMessage
			.mockRejectedValueOnce(new Error("cancelled send failed"))
			.mockResolvedValueOnce({ id: "later-send" });
		renderWithProviders(<MessageComposer channelId="channel-retry" />);

		fireEvent.change(composerInput(), { target: { value: "cancelled draft" } });
		fireEvent.click(sendButton());
		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(1));
		const [, cancelledPayload] = sendMessage.mock.calls[0] as [
			string,
			SendMessageRequest,
		];

		const retryToast = await waitFor(() => retryToastAt(0));
		retryToast.options.cancel.onClick();
		expect(sendMessage).toHaveBeenCalledTimes(1);
		expect(toastDismiss).toHaveBeenCalledWith(retryToast.id);

		fireEvent.change(composerInput(), { target: { value: "later logical draft" } });
		fireEvent.click(sendButton());
		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(2));
		const [, laterPayload] = sendMessage.mock.calls[1] as [
			string,
			SendMessageRequest,
		];
		expect(laterPayload).toMatchObject({
			content_text: "later logical draft",
			client_nonce: UUIDS[1],
		});
		expect(laterPayload.client_nonce).not.toBe(cancelledPayload.client_nonce);
		expect(randomUUID).toHaveBeenCalledTimes(2);
	});

	it("retries GIF, sticker, and voice payloads with route-distinct stable nonces", async () => {
		const routes = [
			{
				name: "GIF",
				select: () =>
					fireEvent.click(
						screen.getByRole("button", { name: "Выбрать GIF для retry" }),
					),
			},
			{
				name: "sticker",
				select: () =>
					fireEvent.click(
						screen.getByRole("button", { name: "Выбрать стикер для retry" }),
					),
			},
			{
				name: "voice",
				select: () => {
					fireEvent.click(
						screen.getByRole("button", { name: "Голосовое сообщение" }),
					);
					fireEvent.click(
						screen.getByRole("button", {
							name: "Отправить голосовое для retry",
						}),
					);
				},
			},
		] as const;
		renderWithProviders(<MessageComposer channelId="channel-retry" />);

		const firstRouteNonces: string[] = [];
		for (const [routeIndex, route] of routes.entries()) {
			const retry = deferred<unknown>();
			sendMessage
				.mockRejectedValueOnce(new Error(`${route.name} send failed`))
				.mockReturnValueOnce(retry.promise);
			route.select();

			const firstCallIndex = routeIndex * 2;
			await waitFor(() =>
				expect(sendMessage).toHaveBeenCalledTimes(firstCallIndex + 1),
			);
			const [, firstPayload] = sendMessage.mock.calls[firstCallIndex] as [
				string,
				SendMessageRequest,
			];
			const immutableFirstPayload = structuredClone(firstPayload);
			expect(firstPayload.client_nonce).toBe(UUIDS[routeIndex]);
			firstRouteNonces.push(firstPayload.client_nonce!);

			const retryToast = await waitFor(() => retryToastAt(routeIndex));
			retryToast.options.action.onClick();
			await waitFor(() =>
				expect(sendMessage).toHaveBeenCalledTimes(firstCallIndex + 2),
			);
			const [, retryPayload] = sendMessage.mock.calls[firstCallIndex + 1] as [
				string,
				SendMessageRequest,
			];
			expect(retryPayload).toEqual(immutableFirstPayload);
			expect(retryPayload.client_nonce).toBe(firstPayload.client_nonce);

			retry.resolve({ id: `${route.name}-retry-success` });
		}

		expect(firstRouteNonces).toEqual(UUIDS.slice(0, 3));
		expect(new Set(firstRouteNonces).size).toBe(3);
		expect(randomUUID).toHaveBeenCalledTimes(3);
	});

	it("suppresses a late retry toast after the composer changes channels", async () => {
		const inFlight = deferred<unknown>();
		sendMessage.mockReturnValueOnce(inFlight.promise);
		const { rerender } = renderWithProviders(
			<MessageComposer channelId="channel-retry" />,
		);

		fireEvent.change(composerInput(), { target: { value: "route-bound draft" } });
		fireEvent.click(sendButton());
		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(1));

		rerender(<MessageComposer channelId="channel-next" />);
		inFlight.reject(new Error("late route failure"));
		await Promise.resolve();
		await Promise.resolve();
		expect(toastError).not.toHaveBeenCalled();
	});

	it("dismisses the owned retry toast when unmounted after one failed pending send", async () => {
		sendMessage.mockRejectedValueOnce(new Error("unmount send failed"));
		const { unmount } = renderWithProviders(
			<MessageComposer channelId="channel-retry" />,
		);

		fireEvent.change(composerInput(), { target: { value: "pending draft" } });
		fireEvent.click(sendButton());
		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(1));
		const retryToast = await waitFor(() => retryToastAt(0));

		unmount();
		expect(toastDismiss).toHaveBeenCalledWith(retryToast.id);
	});
});
